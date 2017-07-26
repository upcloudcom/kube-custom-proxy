/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"fmt"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type Listen struct {
	Domain     string
	PublicIP   string
	PublicPort string
	PodPorts   []PodPort
	Port       int32
}

func (l Listen) Key() string {
	return fmt.Sprintf("%s%s%s%d%s", l.Domain, l.PublicIP, l.PublicPort, l.Port, distinctKeyPodPort(l.PodPorts))
}

func (l Listen) Set(cfg cfgfile.WholeConfig) {
	cfg.Listen().Parts()[l.Key()] = l
}

func NewListen(domain, publicIP, publicPort string, port int32, ports []PodPort) *Listen {
	return &Listen{
		Domain:     domain,
		PublicIP:   publicIP,
		PublicPort: publicPort,
		PodPorts:   ports,
		Port:       port,
	}
}
