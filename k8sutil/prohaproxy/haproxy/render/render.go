/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package render

import (
	"bytes"
	"io"
	"path"
	"path/filepath"
	"tenx-proxy/config"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/parts"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/sections"
	"text/template"
)

const HAPROXY_TPL_NAME = "haproxy.tpl"

type Config struct {
	defaultHTTP *sections.DefaultHTTP
	frontendLB  *sections.FrontendLB
	listen      *sections.Listen
	backend     *sections.Backend
}

func (c *Config) DefaultHTTP() cfgfile.ConfigSection {
	return c.defaultHTTP
}

func (c *Config) FrontendLB() cfgfile.ConfigSection {
	return c.frontendLB
}

func (c *Config) Listen() cfgfile.ConfigSection {
	return c.listen
}

func (c *Config) Backend() cfgfile.ConfigSection {
	return c.backend
}

// data bag => haproxy.tpl
func (c Config) DataBag() *cfgfile.DataBag {
	bag := cfgfile.NewDataBag()
	bag.PublicIP = c.defaultHTTP.PublicIP
	for _, http := range c.defaultHTTP.Parts() {
		if bd, ok := http.(parts.BindDomain); ok {
			dm := cfgfile.NewDomains()
			dm.BackendName = bd.BackendName
			dm.DomainNames = append(dm.DomainNames, bd.Domains...)
			bag.DefaultHTTP.Domains = append(bag.DefaultHTTP.Domains, dm)
		} else if rd, ok := http.(parts.RedirectHTTPS); ok {
			r := cfgfile.NewDomainNames()
			r.DomainNames = append(r.DomainNames, rd.Domains...)
			bag.DefaultHTTP.Redirect = append(bag.DefaultHTTP.Redirect, r)
		}
	}
	bag.FrontendLB.DefaultSSLCert = c.frontendLB.DefaultSSLCert
	for cert := range c.frontendLB.SSLCerts {
		bag.FrontendLB.SSLCerts = append(bag.FrontendLB.SSLCerts, cert)
	}
	for _, lb := range c.frontendLB.Parts() {
		if l, ok := lb.(parts.BindDomainHTTPS); ok {
			dm := cfgfile.NewDomains()
			dm.BackendName = l.BackendName
			dm.DomainNames = append(dm.DomainNames, l.Domains...)
			bag.FrontendLB.Domains = append(bag.FrontendLB.Domains, dm)
		}
	}
	for _, l := range c.listen.Parts() {
		if ll, ok := l.(parts.Listen); ok {
			listen := cfgfile.NewListen()
			listen.DomainName = ll.Domain
			listen.PublicPort = ll.PublicPort
			listen.Port = ll.Port
			for _, p := range ll.PodPorts {
				pod := &cfgfile.Pod{
					Name: p.Name,
					IP:   p.IP,
				}
				listen.Pods = append(listen.Pods, pod)
			}
			bag.Listen = append(bag.Listen, listen)
		}
	}
	for _, b := range c.backend.Parts() {
		if bb, ok := b.(parts.Backend); ok {
			backend := cfgfile.NewBackend()
			backend.BackendName = bb.Name
			backend.Port = bb.Port
			for _, p := range bb.PodPorts {
				pod := &cfgfile.Pod{
					Name: p.Name,
					IP:   p.IP,
				}
				backend.Pods = append(backend.Pods, pod)
			}
			bag.Backend = append(bag.Backend, backend)
		}
	}
	return bag
}

type Render struct {
	rendered   *Config
	configured *Config
}

func NewRender(publicIP, defaultSSLCert string) *Render {
	return &Render{
		rendered: &Config{
			defaultHTTP: sections.NewDefaultHTTP(publicIP),
			frontendLB:  sections.NewFrontendLB(publicIP, defaultSSLCert),
			listen:      sections.NewListen(),
			backend:     sections.NewBackend(),
		},
		configured: &Config{
			defaultHTTP: sections.NewDefaultHTTP(publicIP),
			frontendLB:  sections.NewFrontendLB(publicIP, defaultSSLCert),
			listen:      sections.NewListen(),
			backend:     sections.NewBackend(),
		},
	}
}

func (r *Render) Set(ps ...cfgfile.ConfigPart) {
	for _, part := range ps {
		part.Set(r.configured)
	}
}

func (r *Render) Render() (modified bool, buffer io.Reader) {
	rendered := []cfgfile.ConfigSection{
		r.rendered.defaultHTTP,
		r.rendered.frontendLB,
		r.rendered.listen,
		r.rendered.backend,
	}
	configured := []cfgfile.ConfigSection{
		r.configured.defaultHTTP,
		r.configured.frontendLB,
		r.configured.listen,
		r.configured.backend,
	}
	defer r.reset(rendered, configured)
	for i, r := range rendered {
		c := configured[i]
		if !r.Equal(c) {
			modified = true
			break
		}
	}
	if !modified {
		return
	}
	bb := new(bytes.Buffer)
	haproxyTPLPath := HAPROXY_TPL_NAME
	if config.HaFolder != nil {
		// If args specified, use that one
		haproxyTPLPath = path.Join(*config.HaFolder, HAPROXY_TPL_NAME)
	} else {
		// For testing purpose by now
		haproxyTPLFolder, err := filepath.Abs("../")
		if err == nil {
			haproxyTPLPath = path.Join(haproxyTPLFolder, HAPROXY_TPL_NAME)
		}
	}
	t := template.Must(template.ParseFiles(haproxyTPLPath))
	if err := t.Execute(bb, r.configured.DataBag()); err != nil {
		panic(err)
	}
	buffer = bb
	return
}

func (r *Render) reset(rendered, configured []cfgfile.ConfigSection) {
	for i, section := range configured {
		rendered[i].SyncWith(section)
	}
	t := r.rendered
	r.rendered = r.configured
	for _, s := range rendered {
		s.Reset()
	}
	r.configured = t
}
