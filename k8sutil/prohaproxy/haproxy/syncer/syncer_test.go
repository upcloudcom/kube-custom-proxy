/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package syncer

import (
	"encoding/gob"
	"io"
	"io/ioutil"
	"os"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/parser"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/render"
	"testing"

	"k8s.io/client-go/1.5/pkg/api/v1"
)

type renderConfigToLog struct {
	t *testing.T
}

func (r renderConfigToLog) Reload(reader io.Reader) {
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		r.t.Fatal(err)
	}
	r.t.Log(string(content))
}

type preDumpedSvcsAndPods struct {
	svcs []v1.Service
	pods []v1.Pod
}

func newK8SClient(file string) (client K8SClient, err error) {
	var f io.ReadCloser
	if f, err = os.Open(file); err != nil {
		return
	}
	dec := gob.NewDecoder(f)
	var svcs v1.ServiceList
	if err = dec.Decode(&svcs); err != nil {
		return
	}
	var pods v1.PodList
	if err = dec.Decode(&pods); err != nil {
		return
	}
	client = &preDumpedSvcsAndPods{
		svcs: svcs.Items,
		pods: pods.Items,
	}
	return
}

func (p preDumpedSvcsAndPods) GetAllServices() []v1.Service {
	return p.svcs
}

func (p preDumpedSvcsAndPods) GetAllRunningPods() []v1.Pod {
	return p.pods
}

type aProxyInstance struct {
}

func (aProxyInstance) PublicIP() string {
	return "192.168.88.88"
}

func (aProxyInstance) Domain() string {
	return "tenxcloud.com"
}

func (aProxyInstance) GroupID() string {
	return "group-tenx"
}

func (aProxyInstance) DefaultSSLCert() string {
	return "/etc/certs/default.pem"
}

func TestCFGRender(t *testing.T) {
	reloader := &renderConfigToLog{t: t}
	proxy := &aProxyInstance{}
	r := render.NewRender(proxy.PublicIP(), proxy.DefaultSSLCert())
	client, err := newK8SClient("/Users/lizhen/Documents/Go/src/tenx-proxy/k8sutil/prohaproxy/haproxy/syncer/syncer_test.dat")
	if err != nil {
		t.Fatal(err)
	}
	p := parser.NewParser(proxy)
	s := NewSyncer(r, reloader, p, client)
	go s.Run()
	s.Event() <- struct{}{}
	<-s.Stop()
	s.Close()
}
