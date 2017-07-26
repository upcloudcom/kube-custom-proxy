/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type RedirectHTTPS struct {
	Domains []string
}

func (rh RedirectHTTPS) Key() string {
	return distinctKeyStrings(rh.Domains)
}

func (rh RedirectHTTPS) Set(cfg cfgfile.WholeConfig) {
	cfg.DefaultHTTP().Parts()[rh.Key()] = rh
}

func NewRedirectHTTPS(domains []string) *RedirectHTTPS {
	return &RedirectHTTPS{
		Domains: domains,
	}
}
