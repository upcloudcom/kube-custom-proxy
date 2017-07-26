/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package sections

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type Backend struct {
	*sectionBase
}

func (b *Backend) Parts() map[string]cfgfile.ConfigPart {
	return b.parts
}

func (b Backend) Equal(other cfgfile.ConfigSection) bool {
	_, ok := other.(*Backend)
	if !ok {
		return false
	}
	return b.sectionBase.Equal(other)
}

func (b *Backend) SyncWith(other cfgfile.ConfigSection) {
}

func NewBackend() *Backend {
	return &Backend{
		sectionBase: newSectionBase(),
	}
}
