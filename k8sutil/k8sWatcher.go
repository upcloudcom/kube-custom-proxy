/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package k8sutil

import (
	"github.com/golang/glog"
	//api "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/1.5/pkg/api"
	apiv1 "k8s.io/client-go/1.5/pkg/api/v1"

	kcache "k8s.io/client-go/1.5/tools/cache"
	//k8sclient "k8s.io/kubernetes/pkg/client/unversioned"
	k8sclient "k8s.io/client-go/1.5/kubernetes"
	//	kframework "k8s.io/kubernetes/pkg/controller/framework"
	//kSelector "k8s.io/kubernetes/pkg/fields"
	"fmt"
	"os"
	"reflect"
	"strings"
	"tenx-proxy/config"
	"tenx-proxy/k8smod"
	"tenx-proxy/k8sutil/plugin"
	"time"

	"k8s.io/client-go/1.5/pkg/runtime"
	"k8s.io/client-go/1.5/pkg/watch"
	dockertools "k8s.io/kubernetes/pkg/kubelet/dockertools"
)

const (
	resyncPeriod = 30 * time.Minute
	PODS         = "pods"
	SERVICES     = "services"
	ALL          = "all"
)

var stopCh chan struct{} = make(chan struct{})

type K8sWatcher struct {
	// mu                sync.Mutex
	client  *k8sclient.Clientset
	plugins []plugin.WatcherPlugin
}

func GetEnabledPlugins(pluginlist string) map[string]bool {
	pluginMap := map[string]bool{}
	pluginarray := strings.Split(pluginlist, ",")
	for _, plugin := range pluginarray {
		pluginMap[plugin] = true
	}
	return pluginMap
}

// NewK8sWatcher creates a new watcher client
func NewK8sWatcher(apiServerURL, apiServerToken, apiVersion, username string) (*K8sWatcher, error) {
	method := "NewK8sWather"
	//client, _, err := k8smod.CreateApiserverClient("tenxcloud", apiServerURL, apiServerToken, apiVersion)

	client, err := k8smod.NewK8sClientSet("tenxcloud", "https", apiServerURL, apiServerToken, apiVersion, username)
	if err != nil {
		glog.Errorf("[%s] create k8sclient error:%#v\n", method, err)
		return nil, err
	}
	return &K8sWatcher{
		client:  client,
		plugins: []plugin.WatcherPlugin{},
	}, nil
}

// Init initialize plugins
func (w *K8sWatcher) Init(obj interface{}) {

	pluginMapper := GetEnabledPlugins(*config.Plugins)
	isall := false
	if len(pluginMapper) == 0 { //nothing in the map
		isall = true
	}
	if _, ok := pluginMapper[ALL]; ok {
		isall = true
	}
	pluginEnabled := []string{}
	for _, plugin := range ProbWatcherPlugins() {
		if _, ok := pluginMapper[plugin.Name()]; ok == true ||
			plugin.Name() == "" ||
			plugin.ForceEnabled() ||
			isall == true {
			pluginEnabled = append(pluginEnabled, plugin.Name())
			w.plugins = append(w.plugins, plugin)
		}
	}
	if len(pluginEnabled) == 0 {
		glog.Infof("There's no supported plugin enabled!")
	} else {
		glog.Infof("Plugin enabled: %s", pluginEnabled)
	}

	for _, plugin := range w.plugins {
		plugin.Init(obj)
	}
}
func (w *K8sWatcher) CreateEvent(obj interface{}) {
	for _, plugin := range w.plugins {
		go plugin.CreateEvent(obj)
	}
}

func (w *K8sWatcher) RemoveEvent(obj interface{}) {
	for _, plugin := range w.plugins {
		go plugin.RemoveEvent(obj)
	}
}

func (w *K8sWatcher) UpdateEvent(oldObj, newObj interface{}) {
	for _, plugin := range w.plugins {
		go plugin.UpdateEvent(oldObj, newObj)
	}
}

// func (M *K8sWatcher) createListWatcher(resource string) *kcache.ListWatch {
// 	return kcache.NewListWatchFromClient(M.client, resource, api.NamespaceAll, kSelector.Everything())
// }

func (M *K8sWatcher) Watchpods() kcache.Store {
	// Returns a cache.ListWatch that gets all changes to endpoints.
	eStore, eController := kcache.NewInformer(
		&kcache.ListWatch{
			ListFunc: (kcache.ListFunc)(func(options api.ListOptions) (runtime.Object, error) {
				return M.client.Pods(apiv1.NamespaceAll).List(options)
			}),
			WatchFunc: (kcache.WatchFunc)(func(options api.ListOptions) (watch.Interface, error) {
				return M.client.Pods(apiv1.NamespaceAll).Watch(options)
			}),
		},
		&apiv1.Pod{},
		resyncPeriod,
		kcache.ResourceEventHandlerFuncs{
			AddFunc:    M.CreateEvent,
			DeleteFunc: M.RemoveEvent,
			UpdateFunc: M.UpdateEvent,
		},
	)
	go eController.Run(stopCh)
	return eStore
}

func (M *K8sWatcher) Watchsrvs() kcache.Store {
	eStore, eController := kcache.NewInformer(
		&kcache.ListWatch{
			ListFunc: (kcache.ListFunc)(func(options api.ListOptions) (runtime.Object, error) {
				return M.client.Services(apiv1.NamespaceAll).List(options)
			}),
			WatchFunc: (kcache.WatchFunc)(func(options api.ListOptions) (watch.Interface, error) {
				return M.client.Services(apiv1.NamespaceAll).Watch(options)
			}),
		},
		&apiv1.Service{},
		resyncPeriod,
		kcache.ResourceEventHandlerFuncs{
			AddFunc:    M.CreateEvent,
			DeleteFunc: M.RemoveEvent,
			UpdateFunc: M.UpdateEvent,
		},
	)

	go eController.Run(stopCh)
	return eStore
}

func (M *K8sWatcher) Stop() error {
	_, ok := <-stopCh
	if ok == false {
		close(stopCh)
	}
	return nil
}

func (M *K8sWatcher) Run(dockerClient dockertools.DockerInterface, restartChan chan interface{}) error {
	glog.Infof("Start Main process for the tenx-proxy")
	M.Init(dockerClient)
	// M.Watchpods()
	for {
		methods := M.GetEnabledWatchs()
		for _, method := range methods {
			method()
		}
		fmt.Println("Waiting for tenx-proxy to restart or stop")
		select {
		case <-restartChan:
			for _, plugin := range M.plugins {
				plugin.UpdateEvent(config.ReloadConfig, config.ReloadConfig)
			}
			//stopCh <- struct{}{}
			//fmt.Println("Restarting tenx-proxy")
		}
	}

}

//convert the input "watch method" to func
func (M *K8sWatcher) getMethod(args string) (func() kcache.Store, bool) {
	if args == "" {
		return nil, false
	}
	args = strings.ToUpper(args[:1]) + strings.ToLower(args[1:])

	method := reflect.ValueOf(M).MethodByName(args)

	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func() kcache.Store), true
}

//if add watch method. you should make sure the method name is same with input argument.
func (M *K8sWatcher) GetEnabledWatchs() []func() kcache.Store {
	watchs := []func() kcache.Store{M.Watchpods}
	methods := strings.Split(*config.Watch, ",")
	if *config.Watch == "" || len(methods) == 0 {
		return watchs
	}

	for _, method := range methods {
		if strings.ToLower(method) == "watchpods" {
			continue
		}
		me, has := M.getMethod(method)
		if !has {
			glog.Error("Watch method not found.")
			os.Exit(1)
		}
		watchs = append(watchs, me)
	}
	return watchs
}
