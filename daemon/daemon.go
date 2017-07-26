/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package daemon

import (
	"fmt"

	"k8s.io/kubernetes/pkg/kubelet/dockertools"

	tcconfig "tenx-proxy/config"
	"tenx-proxy/k8sutil"
)

type tcdaemon struct {
}

func NewDaemon() *tcdaemon {
	return &tcdaemon{}
}

func (t *tcdaemon) Run(restartChan chan interface{}) {
	dockerclient := dockertools.ConnectToDockerOrDie(*tcconfig.DockerDaemonUrl, 0)
	if K8sClient, err := k8sutil.NewK8sWatcher(*tcconfig.KubernetesMasterUrl, *tcconfig.BearerToken, "v1"); err != nil {
		fmt.Printf("failed to create kubernetes client, %s\n", err)
	} else {
		K8sClient.Run(dockerclient, restartChan)
	}
}
