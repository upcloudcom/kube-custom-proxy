/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type SSLCert struct {
	File string
}

func (sc SSLCert) Key() string {
	return sc.File
}

func (sc SSLCert) Set(cfg cfgfile.WholeConfig) {
	if flb, ok := cfg.FrontendLB().(cfgfile.FrontendLB); ok {
		flb.AddSSLCert(sc.File)
	}
}

func NewSSLCert(file string) *SSLCert {
	return &SSLCert{
		File: file,
	}
}
