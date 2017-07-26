/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package sections

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type DefaultHTTP struct {
	*sectionBase
	PublicIP string
}

func (dh *DefaultHTTP) Parts() map[string]cfgfile.ConfigPart {
	return dh.parts
}

func (dh DefaultHTTP) Equal(other cfgfile.ConfigSection) bool {
	dho, ok := other.(*DefaultHTTP)
	if !ok {
		return false
	}
	if dh.PublicIP != dho.PublicIP {
		return false
	}
	return dh.sectionBase.Equal(other)
}

func (dh *DefaultHTTP) SyncWith(other cfgfile.ConfigSection) {
	odh, ok := other.(*DefaultHTTP)
	if !ok {
		return
	}
	dh.PublicIP = odh.PublicIP
}

func NewDefaultHTTP(publicIP string) *DefaultHTTP {
	return &DefaultHTTP{
		sectionBase: newSectionBase(),
		PublicIP:    publicIP,
	}
}
