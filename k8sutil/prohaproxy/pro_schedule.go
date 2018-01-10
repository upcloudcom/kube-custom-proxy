/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package prohaproxy

import (
	"strings"
	"tenx-proxy/config"
	//	"tenx-proxy/modules"
	//"time"

	api "k8s.io/client-go/1.5/pkg/api/v1"

	"github.com/golang/glog"
)

var (
	PIPELINE          int = 20
	PodAddChan            = make(chan api.Pod, PIPELINE)
	ServiceAddChan        = make(chan api.Service, PIPELINE)
	PodRemoveChan         = make(chan api.Pod, PIPELINE)
	ServiceRemoveChan     = make(chan api.Service, PIPELINE)
)

// CheckServiceShouldProxy validate if service should be fine for proxy
func (w *Haproxy) CheckServiceShouldProxy(service *api.Service) bool {
	method := "CheckServiceShouldProxy"
	if service.ObjectMeta.Namespace == strings.TrimSpace("kube-system") || service.ObjectMeta.Namespace == strings.TrimSpace("default") {
		glog.V(2).Info(method, "Skip Service:", service.ObjectMeta.Name, "because namespaces:", service.ObjectMeta.Namespace)
		return false
	}
	if len(service.Spec.Ports) < 0 {
		glog.V(2).Infoln(method, "Skip service ", service.ObjectMeta.Name, "Port is empty")
		return false
	}
	if service.Spec.ClusterIP == "None" || service.Spec.ClusterIP == "" {
		glog.V(2).Infoln(method, "Skip service ", service.ObjectMeta.Name, "ClusterIP is None or empty")
		return false
	}
	if service.ObjectMeta.Annotations == nil {
		glog.V(2).Infoln(method, "Skip service ", service.ObjectMeta.Name, "no annotation")
		return false
	}

	if group, ok := service.ObjectMeta.Annotations[config.GROUP_KEY_ANNOTAION]; ok {
		if publicGroupMap, ok := config.DefaultConfigMap.Get(config.ConstPublicGroup).(map[string]config.Group); ok {
			if _, ok := publicGroupMap[group]; ok {
				glog.V(2).Infof("public service is my job: %s %s\n", service.Name, service.Namespace)
				return true
			}
		}
	}

	if w.Group != "" {
		group, ok := service.ObjectMeta.Annotations[config.GROUP_KEY_ANNOTAION]
		if ok && group != w.Group {
			glog.V(2).Infoln("there's no group of service or the group is not match with  proxy group, will just skip.", w.Group)
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
func (w *Haproxy) CheckPodShouldProxy(pod api.Pod) (bool, del bool) {
	if pod.Status.Phase != "Running" {
		glog.V(2).Info("not running: ", pod.Namespace, " ", pod.Name)
		return false, true
	}
	if !w.CheckPodCondition(pod.Status.Conditions) {
		glog.V(2).Info("not ready: ", pod.Namespace, " ", pod.Name)
		return false, true
	}
	if pod.ObjectMeta.Namespace == "default" || pod.ObjectMeta.Namespace == "kube-system" {
		return false, false
	}
	return true, false
}

// HaproxyShedule ...
func (w *Haproxy) HaproxyShedule() {
	// listen relevant channel
	for {
		select {
		case addsvc := <-ServiceAddChan:
			glog.V(2).Infoln("receive service add: ", addsvc.ObjectMeta.Name)
			if _, update := w.CheckExistsAndUpdateService(addsvc); update {
				w.signal <- 1
			}
		case addpod := <-PodAddChan:
			glog.V(2).Infoln("receive pod add: ", addpod.ObjectMeta.Name)
			if w.CheckExistsAndUpdatePod(addpod) {
				w.signal <- 1
			}
		case delsvc := <-ServiceRemoveChan:
			glog.V(2).Infoln("receive service delete: ", delsvc.ObjectMeta.Name)
			if w.RemoveService(delsvc) {
				w.signal <- 1
			}
		case delpod := <-PodRemoveChan:
			glog.V(2).Infoln("receive pod delete: ", delpod.ObjectMeta.Name)
			if w.RemovePod(delpod) {
				w.signal <- 1
			}
		}
	}
}
