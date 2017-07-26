/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type PublicIP struct {
	IP string
}

func (pi PublicIP) Key() string {
	return pi.IP
}

func NewPublicIP(ip string) *PublicIP {
	return &PublicIP{
		IP: ip,
	}
}

func (pi PublicIP) Set(cfg cfgfile.WholeConfig) {
	if pip, ok := cfg.FrontendLB().(cfgfile.PublicIP); ok {
		pip.SetPublicIP(pi.IP)
	}
	if pip, ok := cfg.DefaultHTTP().(cfgfile.PublicIP); ok {
		pip.SetPublicIP(pi.IP)
	}
}
