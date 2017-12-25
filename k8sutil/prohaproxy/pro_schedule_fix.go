//create: 2017/12/09 21:12:12 change: 2017/12/25 21:26:32 author:lijiao
package prohaproxy

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	//	"github.com/google/go-cmp/cmp"
	api "k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/labels"
	"sync"
	"tenx-proxy/config"
	"tenx-proxy/modules"
	"time"
)

type SvcPodInfo struct {
	Pods           map[string]api.Pod
	Services       map[string]api.Service
	Namespaces     map[string]string
	PodMutex       sync.RWMutex
	ServiceMutex   sync.RWMutex
	NamespaceMutex sync.RWMutex
}

var svcPodInfo SvcPodInfo

func init() {
	svcPodInfo = SvcPodInfo{
		Pods:       make(map[string]api.Pod),
		Services:   make(map[string]api.Service),
		Namespaces: make(map[string]string),
	}
}

/* TODO
func ClearCache() {
	//Danger!
	//first lock, last unlock!
	svcPodInfo.PodMutex.Lock()
	svcPodInfo.ServiceMutex.Lock()
	svcPodInfo.NamespaceMutex.Lock()

	svcPodInfo.Pods = make(map[string]api.Pod)
	svcPodInfo.Services = make(map[string]api.Service)
	svcPodInfo.Namespace = make(map[string]string)

	svcPodInfo.NamespaceMutex.Unlock()
	svcPodInfo.ServiceMutex.Unlock()
	svcPodInfo.PodMutex.Unlock()
}
*/

func GetServiceList() (ret []api.Service) {
	svcPodInfo.ServiceMutex.RLock()
	defer svcPodInfo.ServiceMutex.RUnlock()
	for _, service := range svcPodInfo.Services {
		ret = append(ret, service)
	}
	return ret
}
func GetPodList() (ret []api.Pod) {
	svcPodInfo.PodMutex.RLock()
	defer svcPodInfo.PodMutex.RUnlock()
	for _, pod := range svcPodInfo.Pods {
		ret = append(ret, pod)
	}
	return ret
}

func (w *Haproxy) GetKey(meta api.ObjectMeta) string {
	//return meta.Namespace + "." + meta.Name + "." + string(meta.UID)
	return meta.Namespace + "." + meta.Name
}

// CheckExistsAndUpdateService ...
// return true, means updates
func (w *Haproxy) CheckExistsAndUpdateService(svc api.Service) (error, bool) {
	svcPodInfo.ServiceMutex.Lock()
	defer svcPodInfo.ServiceMutex.Unlock()
	key := w.GetKey(svc.ObjectMeta)
	v, ok := svcPodInfo.Services[key]
	if !ok {
		svcPodInfo.Services[key] = svc
		w.RefreshPodList(svc)
		return nil, true
	}
	glog.V(2).Infof("%s %s new svc annotation: %v\n", svc.Name, svc.Namespace, svc.ObjectMeta.Annotations)
	glog.V(2).Infof("%s %s old svc annotation: %v\n", v.Name, v.Namespace, v.ObjectMeta.Annotations)
	/*
		if cmp.Equal(svc.ObjectMeta.Annotations, v.ObjectMeta.Annotations) {
			glog.V(2).Infof("svc annotation not change return: %v\n", v.ObjectMeta.Annotations)
			return nil, false
		}
	*/
	svcPodInfo.Services[key] = svc
	return w.RefreshPodList(svc)
}

// CheckExistsAndUpdatePod exist or not, if not, add in list
func (w *Haproxy) CheckExistsAndUpdatePod(pod api.Pod) bool {
	svcPodInfo.PodMutex.Lock()
	defer svcPodInfo.PodMutex.Unlock()

	key := w.GetKey(pod.ObjectMeta)
	_, ok := svcPodInfo.Pods[key]
	if !ok {
		svcPodInfo.Pods[key] = pod
		return true
	}
	return false
}

func (w *Haproxy) RefreshPodList(svc api.Service) (err error, update bool) {
	method := "RefreshServicePodList"
	glog.V(2).Infoln("Refresh service/pod list...")

	clientApi, err := modules.NewKubernetes(*config.KubernetesMasterUrl, *config.BearerToken, *config.Username)
	if err != nil {
		return err, false
	}

	//fselector := modules.FieldFactory("status.phase", "Running")
	selector := labels.Set(svc.Spec.Selector).AsSelectorPreValidated()
	Pods, errPods := clientApi.GetPodsBySelector(svc.Namespace, selector, nil)
	if errPods != nil {
		glog.Infoln(method, "Error get all running pods", errPods)
	}

	for _, pod := range Pods.Items {
		proxy, del := w.CheckPodShouldProxy((pod))
		if !proxy && !del {
			continue
		}
		if del {
			w.RemovePod(pod)
		} else {
			w.CheckExistsAndUpdatePod(pod)
		}
	}
	return nil, true
}

// PrintsvcPodlist just for debug
func PrintsvcPodlist() {
	for {

		time.Sleep(10 * time.Second)
		strsvcpod, err := json.Marshal(svcPodInfo)
		if err != nil {
			continue
		}
		// glog.Info(string(strsvcpod))
		fmt.Println(string(strsvcpod))
	}
}

// UpdateService ...
func (w *Haproxy) UpdateService(svc api.Service) {
	svcPodInfo.ServiceMutex.Lock()
	defer svcPodInfo.ServiceMutex.Unlock()
	key := w.GetKey(svc.ObjectMeta)
	svcPodInfo.Services[key] = svc
}

// UpdatePods ...
func (w *Haproxy) UpdatePods(pods api.PodList) {
	method := "UpdatePods"
	glog.V(4).Info(method, pods)

	svcPodInfo.PodMutex.Lock()
	defer svcPodInfo.PodMutex.Unlock()

	for _, pod := range pods.Items {
		key := w.GetKey(pod.ObjectMeta)
		svcPodInfo.Pods[key] = pod
	}
}

// UpdatePod ...
func (w *Haproxy) UpdatePod(pod api.Pod) {
	method := "UpdatePods"
	glog.V(4).Info(method, pod)

	svcPodInfo.PodMutex.Lock()
	defer svcPodInfo.PodMutex.Unlock()
	key := w.GetKey(pod.ObjectMeta)
	svcPodInfo.Pods[key] = pod
}

// RemoveService ...
func (w *Haproxy) RemoveService(svc api.Service) bool {
	svcPodInfo.ServiceMutex.Lock()
	defer svcPodInfo.ServiceMutex.Unlock()
	key := w.GetKey(svc.ObjectMeta)
	if _, ok := svcPodInfo.Services[key]; ok {
		glog.V(2).Info("delete services from cache: ", key)
		delete(svcPodInfo.Services, key)
		return true
	}
	return false
}

// RemovePod ...
func (w *Haproxy) RemovePod(pod api.Pod) bool {
	svcPodInfo.PodMutex.Lock()
	defer svcPodInfo.PodMutex.Unlock()

	key := w.GetKey(pod.ObjectMeta)
	if _, ok := svcPodInfo.Pods[key]; ok {
		glog.V(2).Info("delete pod from cache: ", key)
		delete(svcPodInfo.Pods, key)
		return true
	}
	return false
}

// InitServicePodList ...
func (w *Haproxy) InitServicePodList() {
	method := "InitServicePodList"
	glog.V(2).Infoln("Initializing service/pod list...", method)
	clientApi, err := modules.NewKubernetes(*config.KubernetesMasterUrl, *config.BearerToken, *config.Username)
	if err != nil {
		panic("connect kubernetes cluster error")
	}
	services, err := clientApi.GetAllService()
	if err != nil {
		glog.Errorln("sync service error", err)
	}

	for _, svc := range services.Items {
		if !w.CheckServiceShouldProxy(&svc) {
			glog.V(4).Infof("%s service in namespace %s cannot proxy", svc.ObjectMeta.Name, svc.ObjectMeta.Namespace)
			continue
		}
		w.CheckExistsAndUpdateService(svc)
	}
	w.signal <- 1
}