/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package k8smod

import (
	"fmt"

	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/1.5/tools/clientcmd/api"
)

func NewK8sClientSet(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion, username string) (*kubernetes.Clientset, error) {

	if apiserverHost != "" && apiserverToken != "" {
		config := clientcmdapi.NewConfig()
		config.Clusters[clusterName] = &clientcmdapi.Cluster{Server: fmt.Sprintf("%s://%s", apiserverProtocol, apiserverHost), InsecureSkipTLSVerify: true, APIVersion: apiVersion}
		config.AuthInfos[clusterName] = &clientcmdapi.AuthInfo{Token: apiserverToken}
		config.Contexts[clusterName] = &clientcmdapi.Context{
			Cluster:  clusterName,
			AuthInfo: clusterName,
		}
		config.CurrentContext = clusterName

		clientBuilder := clientcmd.NewNonInteractiveClientConfig(*config, clusterName, &clientcmd.ConfigOverrides{}, nil)

		cfg, err := clientBuilder.ClientConfig()

		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}

		return clientset, nil
	} else if apiserverHost != "" && username != "" {
		config := clientcmdapi.NewConfig()
		config.Clusters[clusterName] = &clientcmdapi.Cluster{Server: fmt.Sprintf("%s://%s", apiserverProtocol, apiserverHost), InsecureSkipTLSVerify: true, APIVersion: apiVersion}
		config.AuthInfos[clusterName] = &clientcmdapi.AuthInfo{Username: username}
		config.Contexts[clusterName] = &clientcmdapi.Context{
			Cluster:  clusterName,
			AuthInfo: clusterName,
		}
		config.CurrentContext = clusterName

		clientBuilder := clientcmd.NewNonInteractiveClientConfig(*config, clusterName, &clientcmd.ConfigOverrides{}, nil)

		cfg, err := clientBuilder.ClientConfig()

		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, err
		}

		return clientset, nil

	} else {
		overrides := &clientcmd.ConfigOverrides{}
		overrides.ClusterInfo.Server = ""                              // might be "", but that is OK
		rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: ""} // might be "", but that is OK
		config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		return clientset, nil
	}

}

type ClientSet struct {
	*kubernetes.Clientset
}

func NewClientSet(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion, username string) (c *ClientSet, err error) {
	cs, err := NewK8sClientSet(clusterName, apiserverProtocol, apiserverHost, apiserverToken, apiVersion, username)
	if nil != err {
		return
	}
	c = &ClientSet{Clientset: cs}
	return
}
