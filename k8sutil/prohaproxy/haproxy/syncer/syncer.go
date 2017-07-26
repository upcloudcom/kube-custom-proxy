/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package syncer

import (
	"io"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/parser"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/parts"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/render"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/1.5/pkg/api/v1"
)

type HAProxyReloader interface {
	Reload(reader io.Reader)
}

type K8SClient interface {
	GetAllServices() []v1.Service
	GetAllRunningPods() []v1.Pod
}

type Syncer struct {
	instance chan cfgfile.HAProxyInstance
	event    chan interface{}
	join     chan interface{}
	run      bool
	render   *render.Render
	reloader HAProxyReloader
	parser   *parser.Parser
	k8s      K8SClient
}

func NewSyncer(render *render.Render, reloader HAProxyReloader, parser *parser.Parser, client K8SClient) *Syncer {
	return &Syncer{
		instance: make(chan cfgfile.HAProxyInstance),
		event:    make(chan interface{}),
		join:     make(chan interface{}),
		run:      true,
		render:   render,
		reloader: reloader,
		parser:   parser,
		k8s:      client,
	}
}

func (s *Syncer) updateInstance(instance cfgfile.HAProxyInstance) {
	s.parser.Update(instance)
	s.updateConfigPart([]cfgfile.ConfigPart{
		parts.NewPublicIP(instance.PublicIP()),
		parts.NewDefaultSSLCert(instance.DefaultSSLCert()),
	})
}

func (s *Syncer) updateConfigPart(parts []cfgfile.ConfigPart) {
	s.render.Set(parts...)
}

func (s *Syncer) renderAndReload() {
	modified, buffer := s.render.Render()
	if !modified {
		glog.V(2).Infoln("No changed configuration, skipping reload...")
		return
	}
	s.reloader.Reload(buffer)
}

func (s *Syncer) getCurrentK8SState() {
	svcs := s.k8s.GetAllServices()
	pods := s.k8s.GetAllRunningPods()
	for _, svc := range svcs {
		if ps := s.parser.Parse(&svc, pods); ps != nil {
			s.render.Set(ps...)
		}
	}
}

func (s *Syncer) heartbeat() {
	// TODO
}

func (s *Syncer) Run() {
	for s.run {
		select {
		case instance := <-s.instance:
			s.updateInstance(instance)
			s.getCurrentK8SState()
			s.renderAndReload()
		case <-s.event:
			s.getCurrentK8SState()
			s.renderAndReload()
		case <-time.After(time.Second * 5):
			s.heartbeat()
		}
	}
	s.join <- struct{}{}
}

func (s *Syncer) Stop() <-chan interface{} {
	s.run = false
	return s.join
}

func (s *Syncer) Close() {
	if s.instance != nil {
		close(s.instance)
		s.instance = nil
	}
	if s.event != nil {
		close(s.event)
		s.event = nil
	}
	if s.join != nil {
		close(s.join)
		s.join = nil
	}
}

func (s Syncer) Instance() chan<- cfgfile.HAProxyInstance {
	return s.instance
}

func (s Syncer) Event() chan<- interface{} {
	return s.event
}
