/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package prohaproxy

const (
	haproxy_cfg         = "/etc/haproxy/haproxy.cfg"
	reload              = "reload_haproxy.sh"
	sslsuffix           = ".pem"
	defaultDomainSuffix = "tenxcloud.com"
)

var DefaultSSLKeyFolder = "/etc/sslkeys/"
var SslKeyFolder = "/etc/sslkeys/certs/"
var haFileFolder = "k8sutil/prohaproxy"
