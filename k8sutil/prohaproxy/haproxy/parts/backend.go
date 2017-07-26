/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"fmt"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type Backend struct {
	Name     string
	PodPorts []PodPort
	Port     int32
}

func (b Backend) Key() string {
	return fmt.Sprintf("%s%d%s", b.Name, b.Port, distinctKeyPodPort(b.PodPorts))
}

func (b Backend) Set(cfg cfgfile.WholeConfig) {
	cfg.Backend().Parts()[b.Key()] = b
}

func NewBackend(name string, ports []PodPort, port int32) *Backend {
	return &Backend{
		Name:     name,
		PodPorts: ports,
		Port:     port,
	}
}
