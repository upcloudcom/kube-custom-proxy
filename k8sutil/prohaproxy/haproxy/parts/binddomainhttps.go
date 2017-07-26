/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type BindDomainHTTPS struct {
	*BindDomain
}

func (bdh BindDomainHTTPS) Set(cfg cfgfile.WholeConfig) {
	cfg.FrontendLB().Parts()[bdh.Key()] = bdh
}

func NewBindDomainHTTPS(backendName string, domains []string) *BindDomainHTTPS {
	return &BindDomainHTTPS{
		BindDomain: NewBindDomain(backendName, domains),
	}
}
