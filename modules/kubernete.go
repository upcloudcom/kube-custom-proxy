/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package modules

import (
	"errors"

	// glog "github.com/Sirupsen/logrus"
	"github.com/golang/glog"
	//"k8s.io/client-go/pkg/api"
	k8sclient "k8s.io/client-go/1.5/kubernetes"
	api "k8s.io/client-go/1.5/pkg/api"
	apiv1 "k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/fields"
	"k8s.io/client-go/1.5/pkg/labels"

	"tenx-proxy/k8smod"
)

type Kubernetes struct {
	c *k8sclient.Clientset
}

type K8sServerList struct {
	apiv1.ServiceList
}
type K8sPodList struct {
	apiv1.PodList
}
type K8sPod struct {
	apiv1.Pod
}
type K8sService struct {
	apiv1.Service
}

func NewKubernetes(master, bearerToken string) (ret *Kubernetes, err error) {
	client, err := k8smod.NewK8sClientSet("tenxcloud", "https", master, bearerToken, "v1")
	if err != nil {
		glog.Errorf("failed to create api server client, error: %#v", err)
		return nil, err
	}

	return &Kubernetes{client}, nil
}

// GetPodsByOneField gets pod by selector
func (this *Kubernetes) GetPodsByOneField(field, value string) (*apiv1.PodList, error) {
	if field == "" || value == "" {
		return nil, errors.New("field or value can't be nil")
	}
	selector := fields.Set{field: value}.AsSelector()
	return this.GetPodsByField(selector)
}

func FieldFactory(k, v string) fields.Selector {

	return fields.Set{k: v}.AsSelector()
}

func LabelFactory(label map[string]string) labels.Selector {
	return labels.Set(label).AsSelector()
}

//
func (this *Kubernetes) GetAllService() (*apiv1.ServiceList, error) {

	return this.c.Core().Services(apiv1.NamespaceAll).List(api.ListOptions{})
}
func (this *Kubernetes) GetPodsByField(selector fields.Selector) (*apiv1.PodList, error) {
	return this.c.Core().Pods(apiv1.NamespaceAll).List(api.ListOptions{
		FieldSelector: selector,
	})
}
func (this *Kubernetes) GetServices(selector labels.Selector) (*apiv1.ServiceList, error) {
	return this.c.Core().Services(apiv1.NamespaceAll).List(api.ListOptions{
		LabelSelector: selector,
	})
	// response := &api.ServiceList{}
	// err := this.c.Get().Resource("services").LabelsSelectorParam(selector).Do().Into(response)
	// return response, err
}

func (this *Kubernetes) GetPodsBySelector(namespace string, labelselector labels.Selector, fieldselector fields.Selector) (*apiv1.PodList, error) {
	return this.c.Core().Pods(namespace).List(api.ListOptions{
		LabelSelector: labelselector,
		FieldSelector: fieldselector,
	})
}
