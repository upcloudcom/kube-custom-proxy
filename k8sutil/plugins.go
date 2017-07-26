/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package k8sutil

import (
	"tenx-proxy/k8sutil/plugin"
	"tenx-proxy/k8sutil/prohaproxy"
)

func ProbWatcherPlugins() []plugin.WatcherPlugin {
	allPlugins := []plugin.WatcherPlugin{}

	allPlugins = append(allPlugins, prohaproxy.ProbWatcherPlugin()...)

	return allPlugins
}
