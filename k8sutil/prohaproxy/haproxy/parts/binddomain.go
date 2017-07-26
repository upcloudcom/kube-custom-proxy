/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"fmt"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type BindDomain struct {
	BackendName string
	Domains     []string
}

func (bd BindDomain) Key() string {
	return fmt.Sprintf("%s%s", bd.BackendName, distinctKeyStrings(bd.Domains))
}

func (bd BindDomain) Set(cfg cfgfile.WholeConfig) {
	cfg.DefaultHTTP().Parts()[bd.Key()] = bd
}

func NewBindDomain(backendName string, domains []string) *BindDomain {
	return &BindDomain{
		BackendName: backendName,
		Domains:     domains,
	}
}
