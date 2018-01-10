/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"tenx-proxy/utils"
	"tenx-proxy/utils/watcher"
	"time"

	//"gopkg.in/yaml.v2"
	"encoding/json"
	"errors"
)

type ConfigMaps struct {
	configMap map[string]interface{}
	mutex     sync.Mutex
}

func NewConfigMaps() *ConfigMaps {
	return &ConfigMaps{
		configMap: map[string]interface{}{},
	}
}
func (c *ConfigMaps) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configMap[key] = value
}
func (c *ConfigMaps) SetMap(target map[string]string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configMap = map[string]interface{}{}
	for key, value := range target {
		c.configMap[key] = value
	}

}

func (c *ConfigMaps) Get(key string) interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if ret, ok := c.configMap[key]; ok {
		return ret
	}
	return ""

}

var DefaultConfigMap = NewConfigMaps()

func Get(key string) string {
	return DefaultConfigMap.Get(key).(string)
}
func SetMap(target map[string]string) {
	DefaultConfigMap.SetMap(target)
}

func Set(key, value string) {
	DefaultConfigMap.Set(key, value)
}

func loadConfig(configFile string) *map[string]string {
	conf := &map[string]string{}
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(configData, conf)
	if err != nil {
		return nil
	}
	return conf
}

func SyncConfig() chan interface{} {
	var ReloadFileChan chan interface{} = make(chan interface{})
	//default values
	Set(ConstDomanName, *BindingDomain)
	Set(ConstExternalIP, *ExternalIP)
	Set(ConstGroup, "")
	extentionConfigfile := *ExtConfigFile
	fmt.Printf("extention file:%s\n", extentionConfigfile)
	//if there's extention file existing, will read it firstly,and skip the left ones
	//watching extention files
	if utils.FileExist(extentionConfigfile) {
		ParseExtentionConfig(extentionConfigfile)
	}
	watcher.WatchFile(extentionConfigfile, 2*time.Second, func() {
		fmt.Printf("Configfile Updated, %s\n", extentionConfigfile)
		err := ParseExtentionConfig(extentionConfigfile)
		if err == nil || !utils.FileExist(extentionConfigfile) {
			ReloadFileChan <- struct{}{}
		}
		//if the file is removed, will restart the proxy

		// channel to restart
	})
	if ConfigFile != nil && *ConfigFile != "" {
		//confManager := utils.NewMutexConfigManager(loadConfig(*ConfigFile))
		if !utils.FileExist(extentionConfigfile) {
			if conf := loadConfig(*ConfigFile); conf != nil {
				SetMap(*conf)
			}
		}

		fmt.Printf("Configfile  start monitor\n")
		watcher.WatchFile(*ConfigFile, time.Second, func() {
			if utils.FileExist(extentionConfigfile) {
				return
			}
			fmt.Printf("Configfile Updated\n")
			conf := loadConfig(*ConfigFile)
			if conf != nil {
				SetMap(*conf)
				ReloadFileChan <- struct{}{}
			}
			// channel to restart

		})
	}
	return ReloadFileChan
}

/*
[
    {
        "host":"ubuntu02",
        "address":"127.0.0.1",
        "domain":"test.tenxcloud.com"
    }
]

extention.conf : [{"host":"ubuntu02", "address":"127.0.0.1",  "domain":"test.tenxcloud.com" }]
*/
type NodeItem struct {
	HostName      string `json:"host,omitempty"`
	ListenAddress string `json:"address,omitempty"`
	Domain        string `json:"domain,omitempty"`
	Group         string `json:"group,omitempty"`
}
type Group struct {
	Address   string `json:"address,omitempty"`
	Domain    string `json:"domain,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"`
	Id        string `json:"id,omitempty"`
}
type ExtionConfig struct {
	Hosts  []NodeItem `json:"nodes,omitempty"`
	Groups []Group    `json:"groups,omitempty"`
}

func ParseOldExtentionConfig(file string) error {
	confs := []NodeItem{}
	configData, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error to read file:%s\n", err)
		return err
	}

	err = json.Unmarshal(configData, &confs)
	if err != nil {
		fmt.Printf("Error to :%s\n", err)
		return err
	}
	hostName, err := os.Hostname()
	if err != nil {
		return err
	}
	for _, item := range confs {
		if hostName == item.HostName {
			fmt.Println("Setting config for ", item.HostName, " address:", item.ListenAddress, " Group: ", item.Group)
			Set(ConstDomanName, item.Domain)
			Set(ConstExternalIP, item.ListenAddress)
			Set(ConstGroup, item.Group)
			return nil
		}
	}
	return errors.New("no valid host found for the machine")
}
func ParseExtentionConfig(file string) error {
	confs := ExtionConfig{}
	configData, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error to read file:%s\n", err)
		return err
	}

	err = json.Unmarshal(configData, &confs)
	if err != nil {
		fmt.Printf("fail  to parse the format, will try with old format \n")
		return ParseOldExtentionConfig(file)
	}
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Printf("fail  to get hostname \n")
		return err
	}
	groupMap := map[string]Group{}
	publicGroupMap := map[string]Group{}
	for _, groupitem := range confs.Groups {
		if groupitem.Id != "" {
			groupMap[groupitem.Id] = groupitem
		}
		if groupitem.Type == "public" {
			publicGroupMap[groupitem.Id] = groupitem
		}
	}
	DefaultConfigMap.Set(ConstPublicGroup, publicGroupMap)
	//  fmt.Printf("Group Map : %#v\n", groupMap)
	//  fmt.Printf("Hosts  : %#v\n", confs.Hosts)
	for _, item := range confs.Hosts {

		if hostName == item.HostName {
			if found, ok := groupMap[item.Group]; ok {
				//fmt.Println("Setting config for ", item.HostName," address:", item.ListenAddress ," Group: ", item.Group)
				Set(ConstDomanName, found.Domain)
				Set(ConstExternalIP, item.ListenAddress)
				Set(ConstGroup, item.Group)
				//ConstIsDefault
				Set(ConstIsDefault, strconv.FormatBool(found.IsDefault))

			}

			return nil
		}
	}
	return errors.New("no valid host found for the machine")
}
