/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package config

var (
	Help                *string
	KubernetesMasterUrl *string
	DockerDaemonUrl     *string
	BearerToken         *string

	ExtConfigFile     *string
	ConfigFile        *string
	Namespace         *string
	StandAlone        *bool
	SkipPublicIpCheck *bool
	PubIface          *string
	HostName          *string
	Plugins           *string

	// For service haproxy, we need external IP and binding domain
	ExternalIP    *string
	BindingDomain *string
	CertDir       *string
	HaFolder      *string
	Region        *string
	Watch         *string

	EmailReiver *string
	GottyBin    *string

	V *string
)
