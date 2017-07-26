/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package cfgfile

import "encoding/json"

type HAProxyInstance interface {
	PublicIP() string
	Domain() string
	GroupID() string
	DefaultSSLCert() string
}

type ConfigPart interface {
	Key() string
	Set(cfg WholeConfig)
}

type ConfigSection interface {
	Parts() map[string]ConfigPart
	Equal(other ConfigSection) bool
	SyncWith(other ConfigSection)
	Reset()
}

type WholeConfig interface {
	DefaultHTTP() ConfigSection
	FrontendLB() ConfigSection
	Listen() ConfigSection
	Backend() ConfigSection
	DataBag() *DataBag
}

type FrontendLB interface {
	ConfigSection
	SetDefaultSSLCert(file string)
	AddSSLCert(file string)
}

type PublicIP interface {
	ConfigSection
	SetPublicIP(ip string)
}

type DomainNames struct {
	DomainNames []string `json:"domainNames"`
}

func NewDomainNames() *DomainNames {
	return &DomainNames{
		DomainNames: make([]string, 0),
	}
}

type Domains struct {
	DomainNames []string `json:"domainNames"`
	BackendName string   `json:"backendName"`
}

func NewDomains() *Domains {
	return &Domains{
		DomainNames: make([]string, 0),
	}
}

type DefaultHTTP struct {
	Domains  []*Domains     `json:"domains"`
	Redirect []*DomainNames `json:"redirect"`
}

func NewDefaultHTTP() *DefaultHTTP {
	return &DefaultHTTP{
		Domains:  make([]*Domains, 0),
		Redirect: make([]*DomainNames, 0),
	}
}

type FrontendLBBag struct {
	DefaultSSLCert string     `json:"defaultSSLCert"`
	SSLCerts       []string   `json:"sslCerts"`
	Domains        []*Domains `json:"domains"`
}

func NewFrontendLB() *FrontendLBBag {
	return &FrontendLBBag{
		SSLCerts: make([]string, 0),
		Domains:  make([]*Domains, 0),
	}
}

type Pod struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type Listen struct {
	PublicPort string `json:"publicPort"`
	DomainName string `json:"domainName"`
	Port       int32  `json:"port"`
	Pods       []*Pod `json:"pods"`
}

func NewListen() *Listen {
	return &Listen{
		Pods: make([]*Pod, 0),
	}
}

type Backend struct {
	BackendName string `json:"backendName"`
	Port        int32  `json:"port"`
	Pods        []*Pod `json:"pods"`
}

func NewBackend() *Backend {
	return &Backend{
		Pods: make([]*Pod, 0),
	}
}

type DataBag struct {
	PublicIP    string         `json:"publicIP"`
	DefaultHTTP *DefaultHTTP   `json:"defaultHTTP"`
	FrontendLB  *FrontendLBBag `json:"frontendLB"`
	Listen      []*Listen      `json:"listen"`
	Backend     []*Backend     `json:"backend"`
}

func NewDataBag() *DataBag {
	return &DataBag{
		DefaultHTTP: NewDefaultHTTP(),
		FrontendLB:  NewFrontendLB(),
		Listen:      make([]*Listen, 0),
		Backend:     make([]*Backend, 0),
	}
}

func (db DataBag) String() string {
	if content, err := json.Marshal(db); err != nil {
		return "serialize failed"
	} else {
		return string(content)
	}
}
