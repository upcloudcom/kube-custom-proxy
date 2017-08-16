/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package cmd

import (
	"flag"
	"tenx-proxy/config"
	"tenx-proxy/consts"
	"tenx-proxy/daemon"

	_ "github.com/golang/glog"
	"github.com/spf13/cobra"
)

var (
	Pod                   = "pod"
	Rc                    = "rc"
	ReplicationController = "replicationcontroller"
)

func NewCommand() *cobra.Command {
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "tenx-proxy",
		Short: "tenx-proxy can run as a daemon or a Command line tool.",
		Long: `tenx-proxy can run as a daemon or a Command line tool,
please type tenx-proxy --help for more info.`,
		Run: cmdMain,
	}
	config.DockerDaemonUrl = cmds.PersistentFlags().String("dockerd",
		consts.DockerAccess, "Docker api url, default: unix:///var/run/docker.sock")
	config.KubernetesMasterUrl = cmds.PersistentFlags().String("master",
		"http://localhost:8080", "Kubernetes apiserver url, default: http://localhost:8080")
	config.BearerToken = cmds.PersistentFlags().String("token",
		"", "access token for kubernetes apiserver url")
	config.Username = cmds.PersistentFlags().String("username",
		"", "username for kubernetes apiserver url")
	config.Namespace = cmds.PersistentFlags().String("namespace", "default",
		"Kubernetes namespace used by command line tool")
	config.PubIface = cmds.PersistentFlags().String("Iface",
		"", "The public network interface by which host can reach the internet, if not set, it will detect automatically")
	config.HostName = cmds.PersistentFlags().String("hostname",
		"", "The hostname same the kubelet")
	config.Plugins = cmds.PersistentFlags().String("plugins",
		"proxy", "choose the plugins we need. The current plugins supported: proxy")
	config.ExternalIP = cmds.PersistentFlags().String("externalIP",
		"", "External IP address that we expose the service inside Kubernetes cluster")
	config.SkipPublicIpCheck = cmds.PersistentFlags().Bool("skipPublicIpCheck",
		true, "Whether check the public ip for the service")

	config.BindingDomain = cmds.PersistentFlags().String("bindingDomain",
		"", "The domain name used to bind the external IP address")

	config.ConfigFile = cmds.PersistentFlags().String("config",
		"", "config file that will be set for domain and externalip")

	config.ExtConfigFile = cmds.PersistentFlags().String("extconfig",
		"/etc/tenx/extention.conf", "extention config file that will be set high priority for domain and externalip")

	config.CertDir = cmds.PersistentFlags().String("cert-dir",
		"/etc/sslkeys/certs/", "folder that will contain the haproxy certificate")

	config.HaFolder = cmds.PersistentFlags().String("hafolder",
		"", "point the path of the haproxy config.")
	config.Region = cmds.PersistentFlags().String("region",
		"", "point the region of this task. The current regions supported: ali,qingyun,hangzhou.")
	config.Watch = cmds.PersistentFlags().String("watch",
		"", "choose the watch plugins we need. The current watch supported: Watchsrvs. default add run Watchpods.")
	//BoolVar
	// email reciever
	config.EmailReiver = cmds.PersistentFlags().String("emailReceiver",
		"", `Set recievers of alarm email (e.g., --emailReceiver="user1@mail.com" or --emailReceiver="user1@mail.com;user2@mail.com")`)
	config.V = cmds.PersistentFlags().String("v",
		"2", "log level")
	return cmds
}

func cmdMain(cmd *cobra.Command, args []string) {
	// Set log level of glog
	flag.Lookup("v").Value.Set(*config.V)
	daemon := daemon.NewDaemon()
	daemon.Run(config.SyncConfig())
}
