/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package sections

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type FrontendLB struct {
	*sectionBase
	PublicIP       string
	DefaultSSLCert string
	SSLCerts       map[string]interface{}
}

func (flb *FrontendLB) Parts() map[string]cfgfile.ConfigPart {
	return flb.parts
}

func (flb FrontendLB) Equal(other cfgfile.ConfigSection) bool {
	flbo, ok := other.(*FrontendLB)
	if !ok {
		return false
	}
	if flb.PublicIP != flbo.PublicIP || !stringKeysEqual(flb.SSLCerts, flbo.SSLCerts) {
		return false
	}
	return flb.sectionBase.Equal(other)
}

func (flb *FrontendLB) SyncWith(other cfgfile.ConfigSection) {
	oflb, ok := other.(*FrontendLB)
	if !ok {
		return
	}
	flb.PublicIP = oflb.PublicIP
	flb.DefaultSSLCert = oflb.DefaultSSLCert
	sslCerts := make(map[string]interface{})
	for key := range oflb.SSLCerts {
		sslCerts[key] = struct{}{}
	}
	flb.SSLCerts = sslCerts
}

func (flb *FrontendLB) SetDefaultSSLCert(file string) {
	flb.DefaultSSLCert = file
}

func (flb *FrontendLB) AddSSLCert(file string) {
	flb.SSLCerts[file] = struct{}{}
}

func (flb *FrontendLB) SetPublicIP(ip string) {
	flb.PublicIP = ip
}

func NewFrontendLB(publicIP, defaultSSLCert string) *FrontendLB {
	return &FrontendLB{
		sectionBase:    newSectionBase(),
		PublicIP:       publicIP,
		DefaultSSLCert: defaultSSLCert,
		SSLCerts:       make(map[string]interface{}),
	}
}
