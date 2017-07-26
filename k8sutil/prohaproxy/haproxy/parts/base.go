/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parts

import (
	"strings"
)

type PodPort struct {
	Name string
	IP   string
}

func distinctKeyPodPort(ports []PodPort) string {
	c := len(ports)
	s := make([]string, c)
	for i, p := range ports {
		s[i] = p.Name + p.IP
	}
	return distinctKeyStrings(s)
}

func distinctKeyStrings(s []string) string {
	return strings.Join(s, "")
}
