/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package prohaproxy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"tenx-proxy/config"
	"tenx-proxy/k8sutil/common"
	"tenx-proxy/k8sutil/plugin"
	"tenx-proxy/modules"
	"time"

	"github.com/golang/glog"
	api "k8s.io/client-go/1.5/pkg/api/v1"

	"tenx-proxy/k8sutil/prohaproxy/haproxy/parser"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/render"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/syncer"
)

type Haproxy struct {
	signal           chan int //use lock
	PublicIp         []string
	haproxyConfig    map[string]string
	Group            string
	IsDefault        bool
	LastErrEmailTime time.Time
	mutex            sync.Mutex
	syncer           *syncer.Syncer
}

func ProbWatcherPlugin() []plugin.WatcherPlugin {
	return []plugin.WatcherPlugin{&Haproxy{}}
}

func (w *Haproxy) Name() string {
	return "tenx-proxy"
}

func (w *Haproxy) ForceEnabled() bool {
	return false
}
func (w *Haproxy) syncHaproxyConfigmap() {
	var ExternalIP, DomainName, Group string
	if *config.ExternalIP == "" && *config.BindingDomain == "" {
		ExternalIP = config.Get(config.ConstExternalIP)
		DomainName = config.Get(config.ConstDomanName)
		Group = config.Get(config.ConstGroup)
		w.IsDefault, _ = strconv.ParseBool(config.Get(config.ConstIsDefault))
	} else {
		ExternalIP = *config.ExternalIP
		DomainName = *config.BindingDomain
		Group = ""
		w.IsDefault = false
	}

	glog.V(5).Infof("Using customized publicIP: %s and domain: %s", ExternalIP, DomainName)
	w.PublicIp = []string{ExternalIP}
	w.haproxyConfig = map[string]string{ExternalIP: DomainName}
	w.Group = Group
}

func (w *Haproxy) Init(obj interface{}) {
	glog.V(2).Infoln("Starting service haproxy plugin.")

	w.syncHaproxyConfigmap()
	w.signal = make(chan int, 3000)
	w.LastErrEmailTime = time.Now()

	if *config.HaFolder != "" {
		haFileFolder = strings.TrimSuffix(*config.HaFolder, "/")
	}
	glog.V(2).Infoln("Location of haproxy configuration:", haFileFolder)
	go w.initSyncer()
	w.InitServicePodList()
	go w.Refresh()
	go w.updateConfigFile()
	go w.HaproxyShedule()

	//	go PrintsvcPodlist()
}

func (w *Haproxy) PublicIP() string {
	return config.Get(config.ConstExternalIP)
}

func (w *Haproxy) Domain() string {
	return config.Get(config.ConstDomanName)
}

func (w *Haproxy) GroupID() string {
	return config.Get(config.ConstGroup)
}

func (w *Haproxy) DefaultSSLCert() string {
	return DefaultSSLKeyFolder + "default.pem"
}

func (w *Haproxy) GetAllServices() []api.Service {
	return GetServiceList()
}

func (w *Haproxy) GetAllRunningPods() []api.Pod {
	return GetPodList()
}

func (w *Haproxy) initSyncer() {
	glog.V(2).Infoln("Initializing sync process...")
	r := render.NewRender(w.PublicIP(), w.DefaultSSLCert())
	p := parser.NewParser(w)
	w.syncer = syncer.NewSyncer(r, w, p, w)
	w.syncer.Run()
}

// Reload action ...
func (w *Haproxy) Reload(reader io.Reader) {
	buffer := new(bytes.Buffer)
	cfgcontent := io.TeeReader(reader, buffer)
	cfgBytes, err := ioutil.ReadAll(cfgcontent)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(haproxy_cfg, cfgBytes, 0755); err != nil {
		glog.Errorf("write \"/etc/haproxy/haproxy.cfg\" failed: %v\n", err)
		return
	}
	glog.V(5).Info(buffer.String())
	errReload := reloadHAProxy()
	if errReload != nil {
		glog.Errorln("Haproxy.Reload", "Error reload haproxy config", errReload)
		//send email at least 2 minutes.
		if time.Now().Sub(w.LastErrEmailTime) > time.Minute*time.Duration(2) {
			errStr := errReload.Error()
			errStr += "\n" + common.GetErrProcess(errStr)
			w.sendReloadErrEmail(errStr)
		}
	}
}

// CreateEvent because the docker id don't contains in new create pod. This part should be empty.
// get a service and sync all pod belongs current service
// send signal, then reload all config,
func (w *Haproxy) CreateEvent(obj interface{}) {
	method := "CreateEvent"
	// glog.V(5).Infoln(method, "debug. create event not use reload.")
	newObject, ok2 := obj.(*api.Service)
	// return if not service
	if !ok2 {
		return
	}
	if !w.CheckServiceShouldProxy(newObject) {
		glog.V(2).Infoln(method, "no meet condition, skipping...")
		return
	}

	glog.V(2).Infoln(method, "Adding new service proxy: service name:", newObject.ObjectMeta.Name, "namespace:", newObject.ObjectMeta.Namespace)
	ServiceAddChan <- (*newObject)
}

// RemoveEvent ...
func (w *Haproxy) RemoveEvent(obj interface{}) {
	method := "RemoveEvent"
	object, ok1 := obj.(*api.Pod)
	if ok1 {
		glog.V(2).Infoln(method, "Removing pod, name:", object.ObjectMeta.Name, "namespace:", object.ObjectMeta.Namespace)
		PodRemoveChan <- (*object)
	}
	svcobject, ok2 := obj.(*api.Service)
	if ok2 {
		glog.V(2).Info("Removing service, name:", svcobject.ObjectMeta.Name, "namespace:", svcobject.ObjectMeta.Namespace)
		ServiceRemoveChan <- (*svcobject)
	}
}

func (w *Haproxy) reloadConfig() {
	glog.V(2).Infoln("Syncing configuration from configmap...")
	w.syncHaproxyConfigmap()
	w.syncer.Instance() <- w
}

func isReloadConfigEvent(objs ...interface{}) bool {
	if len(objs) != 2 {
		return false
	}
	for _, obj := range objs {
		str, ok := obj.(string)
		if !ok {
			return false
		}
		if str != config.ReloadConfig {
			return false
		}
	}
	return true
}

// UpdateEvent ...
func (w *Haproxy) UpdateEvent(oldObj, newObj interface{}) {
	if isReloadConfigEvent(oldObj, newObj) {
		w.reloadConfig()
		return
	}
	method := "UpdateEvent"
	newObject := &api.Pod{}
	var ok1 bool
	var ok2 bool
	_, ok1 = oldObj.(*api.Pod)
	newObject, ok2 = newObj.(*api.Pod)

	if ok1 && ok2 {
		glog.V(2).Info("receive update for pod", newObject.ObjectMeta.Name)

		proxy, del := w.CheckPodShouldProxy((*newObject))
		if !proxy && !del {
			return
		}
		if del {
			glog.V(2).Info("delete not ready pod, name: ", newObject.ObjectMeta.Name, " namespace: ", newObject.ObjectMeta.Namespace)
			PodRemoveChan <- *newObject
		} else {
			glog.V(2).Info("update proxy for pod, name: ", newObject.ObjectMeta.Name, " namespace: ", newObject.ObjectMeta.Namespace)
			PodAddChan <- *newObject
		}
	} else {
		//update service
		newObject := &api.Service{}
		newObject, ok2 = newObj.(*api.Service)
		if !ok2 {
			return
		}
		glog.V(2).Infoln(method, "update proxy for service, name:", newObject.ObjectMeta.Name, "namespace:", newObject.ObjectMeta.Namespace)
		ServiceAddChan <- *newObject
	}
}

func (w *Haproxy) updateConfigFile() {
	for {
		<-w.signal
		l := len(w.signal)
		glog.V(2).Infoln("Refreshing - current event number: ", l)
		for l > 0 {
			<-w.signal
			l--
		}
		glog.V(2).Infoln("Refreshing - start refresh")
		w.syncer.Event() <- struct{}{}
		// Don't reload twice within 5 seconds
		time.Sleep(5 * time.Second)
	}
}

func reloadHAProxy() error {
	glog.V(2).Infoln("Reloading haproxy configuration...")

	reloadSh := haFileFolder + "/" + reload

	cmd := exec.Command("/bin/sh", reloadSh)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		glog.V(5).Infoln("StdoutPipe: " + err.Error())
	}

	//cmd error
	stderr, errPipe := cmd.StderrPipe()
	if errPipe != nil {
		glog.V(5).Infoln("StderrPipe: ", err.Error())
	}

	if errStart := cmd.Start(); err != nil {
		glog.V(5).Infoln("Start: ", err.Error())
		return errStart
	}

	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		glog.V(5).Infoln("ReadAll stderr: ", err.Error())
	}

	if len(bytesErr) != 0 {
		glog.V(5).Infof("stderr is not nil: %s", bytesErr)
		return fmt.Errorf(string(bytesErr))
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		glog.V(5).Infoln("ReadAll stdout: ", err.Error())

	}

	if err := cmd.Wait(); err != nil {
		glog.V(5).Infoln("Wait: ", err.Error())
		return err
	}
	if len(bytes) != 0 {
		glog.V(5).Infof("stdout: %s", bytes)
	}
	return err
}

// 发送报警邮件，如果运行时传入参数emailReceiver，则所有报警邮件都发送给emailReceiver
// 如果没有传入此参数，则过滤掉WARNING邮件发送给config.ServiceEmail
func (w *Haproxy) sendReloadErrEmail(errStr string) {
	method := "sendReloadErrEmail"

	reciever := config.ServiceEmail
	if *config.EmailReiver != "" {
		reciever = *config.EmailReiver
	}

	// msg "Redirecting to /bin/systemctl reload haproxy.service Capture port failed!" is not error, no need to send email
	if strings.Index(errStr, `Redirecting to /bin/systemctl reload  haproxy.service`) == 0 {
		return
	}

	hostName := "unknown"
	if name, err := os.Hostname(); err == nil {
		hostName = name
	}
	w.LastErrEmailTime = time.Now()
	body := `<html>
				<body>
				<div align="center">
				<img src="https://dn-tenxcloud.qbox.me/98eb51e1c356406567bb764d1bbff0c9.png"><br>
				<b>史上最紧急！！！</b>
				</div>
				<h3>
				  亲爱的时速云Service团队,你好！
				</h3>
				<p>
				  我很郑重的告诉你，` + *config.Region + `的 Haproxy 重载失败了，快去修复吧，后果你懂的。
				</p>
				<p>
				   以下是报告的错误: ` + errStr + `
				</p>
				<p>
					此封邮件由主机 ` + hostName + ` 发送。启动时可设置参数--emailReceiver="user1@mail.com" 或 --emailReceiver="user1@mail.com;user2@mail.com"来指定收件人
				</p>
				</body>
			</html>`
	subject := "代理服务出问题了，请尽快检查"

	glog.V(5).Infoln(method, "send email to:", reciever)
	err := modules.SendMail(config.Euser, config.EUserName, config.Epassword, config.Ehost, reciever, subject, body, "html")
	if err != nil {
		glog.V(5).Infoln(method, err)
	}
}
