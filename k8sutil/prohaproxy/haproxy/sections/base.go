/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package sections

import (
	"reflect"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
)

type sectionBase struct {
	parts map[string]cfgfile.ConfigPart
}

func (sb sectionBase) Equal(other cfgfile.ConfigSection) bool {
	return stringKeysEqual(sb.parts, other.Parts())
}

func (sb *sectionBase) Reset() {
	sb.parts = make(map[string]cfgfile.ConfigPart)
}

func newSectionBase() *sectionBase {
	return &sectionBase{
		parts: make(map[string]cfgfile.ConfigPart),
	}
}

func stringKeysEqual(a, b interface{}) bool {
	vA := reflect.ValueOf(a)
	vB := reflect.ValueOf(b)
	if vA.Kind() != reflect.Map || vB.Kind() != reflect.Map {
		return false
	}
	aKeys := vA.MapKeys()
	bKeys := vB.MapKeys()
	if len(aKeys) != len(bKeys) {
		return false
	}
	aMap := make(map[string]interface{})
	for _, key := range aKeys {
		ki := key.Interface()
		if ks, ok := ki.(string); !ok {
			return false
		} else {
			aMap[ks] = struct{}{}
		}
	}
	for _, key := range bKeys {
		ki := key.Interface()
		if ks, ok := ki.(string); !ok {
			return false
		} else {
			if _, exist := aMap[ks]; !exist {
				return false
			}
		}
	}
	return true
}
