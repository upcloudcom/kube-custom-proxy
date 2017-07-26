/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type DefaultSSLCert struct {
	File string
}

func (dsc DefaultSSLCert) Key() string {
	return dsc.File
}

func NewDefaultSSLCert(file string) *DefaultSSLCert {
	return &DefaultSSLCert{
		File: file,
	}
}

func (dsc DefaultSSLCert) Set(cfg cfgfile.WholeConfig) {
	if flb, ok := cfg.FrontendLB().(cfgfile.FrontendLB); ok {
		flb.SetDefaultSSLCert(dsc.File)
	}
}
