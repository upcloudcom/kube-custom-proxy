/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package prohaproxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"tenx-proxy/config"
	"tenx-proxy/modules"
	"time"

	api "k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/kubernetes/pkg/labels"

	"github.com/golang/glog"
	"github.com/google/go-cmp/cmp"
)

var SvcPodList = servicePodList{}

type servicePodList struct {
	PodInfo PodListInfo
	SvcInfo ServiceListInfo
}

type ServiceListInfo struct {
	ServiceList api.ServiceList
	SvcMutex    sync.RWMutex
}
type PodListInfo struct {
	PodList  api.PodList
	PodMutex sync.RWMutex
}

var (
	PIPELINE          int = 20
	PodAddChan            = make(chan api.Pod, PIPELINE)
	ServiceAddChan        = make(chan api.Service, PIPELINE)
	PodRemoveChan         = make(chan api.Pod, PIPELINE)
	ServiceRemoveChan     = make(chan api.Service, PIPELINE)
)

// PrintsvcPodlist just for debug
func PrintsvcPodlist() {
	for {
		time.Sleep(10 * time.Second)
		strsvcpod, err := json.Marshal(SvcPodList)
		if err != nil {
			continue
		}
		// glog.Info(string(strsvcpod))
		fmt.Println(string(strsvcpod))
	}
}

// UpdateService ...
func (w *Haproxy) UpdateService(obj api.Service) {
	SvcPodList.SvcInfo.SvcMutex.Lock()
	defer SvcPodList.SvcInfo.SvcMutex.Unlock()
	SvcPodList.SvcInfo.ServiceList.Items = append(SvcPodList.SvcInfo.ServiceList.Items, obj)
}

// UpdatePods ...
func (w *Haproxy) UpdatePods(objs api.PodList) {
	method := "UpdatePods"
	glog.V(4).Info(method, objs)

	SvcPodList.PodInfo.PodMutex.Lock()
	defer SvcPodList.PodInfo.PodMutex.Unlock()
	SvcPodList.PodInfo.PodList.Items = append(SvcPodList.PodInfo.PodList.Items, objs.Items...)
}

// UpdatePod ...
func (w *Haproxy) UpdatePod(obj api.Pod) {
	method := "UpdatePods"
	glog.V(4).Info(method, obj)

	SvcPodList.PodInfo.PodMutex.Lock()
	defer SvcPodList.PodInfo.PodMutex.Unlock()
	SvcPodList.PodInfo.PodList.Items = append(SvcPodList.PodInfo.PodList.Items, obj)
}

// RemoveService ...
func (w *Haproxy) RemoveService(svc api.Service) bool {
	SvcPodList.SvcInfo.SvcMutex.Lock()
	defer SvcPodList.SvcInfo.SvcMutex.Unlock()
	if len(SvcPodList.SvcInfo.ServiceList.Items) == 0 {
		return false
	}
	for i, v := range SvcPodList.SvcInfo.ServiceList.Items {
		if svc.ObjectMeta.Name == v.ObjectMeta.Name && svc.ObjectMeta.Namespace == v.ObjectMeta.Namespace {
			SvcPodList.SvcInfo.ServiceList.Items[i] = SvcPodList.SvcInfo.ServiceList.Items[0]
			SvcPodList.SvcInfo.ServiceList.Items = SvcPodList.SvcInfo.ServiceList.Items[1:]
		}
	}
	return true
}

// RemovePod ...
func (w *Haproxy) RemovePod(pod api.Pod) bool {
	SvcPodList.PodInfo.PodMutex.Lock()
	defer SvcPodList.PodInfo.PodMutex.Unlock()
	if len(SvcPodList.PodInfo.PodList.Items) == 0 {
		return false
	}
	for i, v := range SvcPodList.PodInfo.PodList.Items {
		if pod.ObjectMeta.Name == v.ObjectMeta.Name && pod.ObjectMeta.Namespace == v.ObjectMeta.Namespace {
			SvcPodList.PodInfo.PodList.Items[i] = SvcPodList.PodInfo.PodList.Items[0]
			SvcPodList.PodInfo.PodList.Items = SvcPodList.PodInfo.PodList.Items[1:]
		}
	}
	return true
}

// CheckServiceShouldProxy validate if service should be fine for proxy
func (w *Haproxy) CheckServiceShouldProxy(service *api.Service) bool {
	method := "CheckServiceShouldProxy"
	if service.ObjectMeta.Namespace == strings.TrimSpace("kube-system") || service.ObjectMeta.Namespace == strings.TrimSpace("default") {
		glog.V(4).Info(method, "Skip Service:", service.ObjectMeta.Name, "because namespaces:", service.ObjectMeta.Namespace)
		return false
	}
	if len(service.Spec.Ports) < 0 {
		glog.V(4).Infoln(method, "Skip service ", service.ObjectMeta.Name, "Port is empty")
		return false
	}
	if service.Spec.ClusterIP == "None" || service.Spec.ClusterIP == "" {
		glog.V(4).Infoln(method, "Skip service ", service.ObjectMeta.Name, "ClusterIP is None or empty")
		return false
	}
	if service.ObjectMeta.Annotations == nil {
		glog.V(4).Infoln(method, "Skip service ", service.ObjectMeta.Name, "no annotation")
		return false
	}
	if w.Group != "" {
		group, ok := service.ObjectMeta.Annotations[config.GROUP_KEY_ANNOTAION]
		if ok && group != w.Group {
			glog.V(4).Infoln("there's no group of service or the group is not match with  proxy group, will just skip.", w.Group)
			return false
		}
		if group == config.GROUP_VAR_NONE {
			return false
		}
		if !ok {
			// For old services, there is no lbgroup info, so every proxy will do the the job
			return true
		}

	}

	return true
}

// CheckPodCondition ...
func (w *Haproxy) CheckPodCondition(cons []api.PodCondition) bool {
	for _, v := range cons {
		if v.Type == "Ready" && string(v.Status) == "True" {
			return true
		}
	}
	return false
}

// CheckPodShouldProxy validate if pod should be fine for proxy
func (w *Haproxy) CheckPodShouldProxy(pod api.Pod) bool {
	if pod.Status.Phase != "Running" {
		return false
	}
	if !w.CheckPodCondition(pod.Status.Conditions) {
		return false
	}
	if pod.ObjectMeta.Namespace == "default" || pod.ObjectMeta.Namespace == "kube-system" {
		return false
	}
	return true
}

// CheckPodBelongService check  pod belongs svc
func CheckPodBelongService(pod api.Pod, svc api.Service) bool {
	selector := labels.Set(svc.Spec.Selector).AsSelectorPreValidated()
	if pod.ObjectMeta.Namespace != svc.ObjectMeta.Namespace {
		return false
	}
	if !selector.Matches(labels.Set(pod.ObjectMeta.Labels)) {
		return false
	}
	return true
}

// CheckExistsAndUpdateService ...
// return true, means updates
func (w *Haproxy) CheckExistsAndUpdateService(svc api.Service) bool {
	SvcPodList.SvcInfo.SvcMutex.Lock()
	defer SvcPodList.SvcInfo.SvcMutex.Unlock()
	if len(SvcPodList.SvcInfo.ServiceList.Items) == 0 {
		SvcPodList.SvcInfo.ServiceList.Items = append(SvcPodList.SvcInfo.ServiceList.Items, svc)
		return true
	}

	for k, v := range SvcPodList.SvcInfo.ServiceList.Items {
		if v.ObjectMeta.Name == svc.ObjectMeta.Name && v.ObjectMeta.Namespace == svc.ObjectMeta.Namespace {
			if cmp.Equal(svc.ObjectMeta.Annotations, v.ObjectMeta.Annotations) {
				return false
			}
			SvcPodList.SvcInfo.ServiceList.Items[k] = svc
			return true
		}
	}
	SvcPodList.SvcInfo.ServiceList.Items = append(SvcPodList.SvcInfo.ServiceList.Items, svc)
	return true

}

// CheckExistsAndUpdatePod exist or not, if not, add in list
func (w *Haproxy) CheckExistsAndUpdatePod(pod api.Pod) bool {
	SvcPodList.PodInfo.PodMutex.Lock()
	defer SvcPodList.PodInfo.PodMutex.Unlock()
	if len(SvcPodList.PodInfo.PodList.Items) == 0 {
		SvcPodList.PodInfo.PodList.Items = append(SvcPodList.PodInfo.PodList.Items, pod)
		return true
	}
	for _, v := range SvcPodList.PodInfo.PodList.Items {
		if v.ObjectMeta.Name == pod.ObjectMeta.Name && v.ObjectMeta.Namespace == pod.ObjectMeta.Namespace {
			return false
		}
	}
	SvcPodList.PodInfo.PodList.Items = append(SvcPodList.PodInfo.PodList.Items, pod)
	return true
}

// HaproxyShedule ...
func (w *Haproxy) HaproxyShedule() {
	// listen relevant channel
	for {
		select {
		case addsvc := <-ServiceAddChan:
			glog.V(4).Infoln("receive service", addsvc.ObjectMeta.Name)
			if w.CheckExistsAndUpdateService(addsvc) {
				w.signal <- 1
			}
		case addpod := <-PodAddChan:
			// Pod Add and update
			glog.V(4).Infoln("receive pod", addpod.ObjectMeta.Name)
			if w.CheckExistsAndUpdatePod(addpod) {
				w.signal <- 1
			}
		case delsvc := <-ServiceRemoveChan:
			// Pod delete service
			glog.V(4).Infoln("delete service", delsvc.ObjectMeta.Name)
			if w.RemoveService(delsvc) {
				w.signal <- 1
			}
		case delpod := <-PodRemoveChan:
			glog.V(4).Infoln("delete Pod", delpod.ObjectMeta.Name)
			if w.RemovePod(delpod) {
				w.signal <- 1
			}
		}
		// time.Sleep(2 * time.Second)

	}
}

// GetServiceList return all available service list
func GetServiceList() api.ServiceList {
	return SvcPodList.SvcInfo.ServiceList
}

// GetPodList return all available pod list
func GetPodList() api.PodList {
	return SvcPodList.PodInfo.PodList
}

// SyncPods ...
func (w *Haproxy) SyncPods() {
	method := "SyncPods"
	ticker := time.NewTicker(time.Minute * 30)
	go func() {
		for _ = range ticker.C {
			glog.V(2).Info("Syncing pod information...")
			clientApi, err := modules.NewKubernetes(*config.KubernetesMasterUrl, *config.BearerToken, *config.Username)
			if err != nil {
				time.Sleep(time.Second * 2)
				continue
			}
			svclist := GetServiceList()
			Pods, errPods := clientApi.GetPodsByOneField("status.phase", "Running")
			if errPods != nil {
				glog.V(2).Info(method, "Error get all running pods", errPods)
			}

			for _, svc := range svclist.Items {
				for _, pod := range Pods.Items {
					if !w.CheckPodShouldProxy(pod) {
						continue
					}
					if !CheckPodBelongService(pod, svc) {
						continue
					}
					if w.CheckExistsAndUpdatePod(pod) {
						w.signal <- 1
					}
				}
			}
		}
	}()
}

// SyncServices ...
func (w *Haproxy) SyncServices() {
	ticker := time.NewTicker(time.Minute * 30)
	go func() {
		for _ = range ticker.C {
			glog.V(2).Info("Syncing service information...")
			clientApi, err := modules.NewKubernetes(*config.KubernetesMasterUrl, *config.BearerToken, *config.Username)
			if err != nil {
				time.Sleep(time.Second * 2)
				continue
			}
			services, err := clientApi.GetAllService()
			if err != nil {
				time.Sleep(time.Second * 2)
				glog.Info("sync service error", err)
				continue
			}
			for _, svc := range services.Items {
				if !w.CheckServiceShouldProxy(&svc) {
					glog.V(4).Infof("%s service in namespace %s cannot proxy", svc.ObjectMeta.Name, svc.ObjectMeta.Namespace)
					continue
				}
				glog.V(4).Info(svc.ObjectMeta.Namespace, "\t", svc.ObjectMeta.Name)
				if w.CheckExistsAndUpdateService(svc) {
					w.signal <- 1
				}
			}
		}
	}()

}

// InitServicePodList ...
func (w *Haproxy) InitServicePodList() {
	method := "InitServicePodList"
	glog.V(2).Infoln("Initializing service/pod list...")
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
	svclist := GetServiceList()
	Pods, errPods := clientApi.GetPodsByOneField("status.phase", "Running")
	if errPods != nil {
		glog.Infoln(method, "Error get all running pods", errPods)
	}
	for _, svc := range svclist.Items {
		for _, pod := range Pods.Items {
			if !w.CheckPodShouldProxy(pod) {
				continue
			}
			if !CheckPodBelongService(pod, svc) {
				continue
			}
			w.CheckExistsAndUpdatePod(pod)
		}
	}
	w.signal <- 1
}
