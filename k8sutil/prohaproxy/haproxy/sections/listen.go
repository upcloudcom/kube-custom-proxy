/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package sections

import (
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type Listen struct {
	*sectionBase
}

func (l *Listen) Parts() map[string]cfgfile.ConfigPart {
	return l.parts
}

func (l Listen) Equal(other cfgfile.ConfigSection) bool {
	_, ok := other.(*Listen)
	if !ok {
		return false
	}
	return l.sectionBase.Equal(other)
}

func (l *Listen) SyncWith(other cfgfile.ConfigSection) {
}

func NewListen() *Listen {
	return &Listen{
		sectionBase: newSectionBase(),
	}
}
