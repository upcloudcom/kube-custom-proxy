/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2016 TenxCloud. All Rights Reserved.
 */

package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	// "github.com/golang/glog"
	glog "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.5/pkg/api"
)

//run gotty need container id. Get container id from Pod.
func GetPodContainerId(pod *api.Pod) (string, error) {
	containerLen := len(pod.Status.ContainerStatuses)
	if containerLen == 0 {
		return "", fmt.Errorf("Get container failed. No container is contained in pod %s.", pod.ObjectMeta.Name)
	} else if containerLen >= 2 {
		glog.Warningln("Get container WARNING. More than one container is contained in pod %s.", pod.ObjectMeta.Name)
	}
	podContainerId := pod.Status.ContainerStatuses[0].ContainerID
	return strings.TrimPrefix(podContainerId, "docker://"), nil
}

//assert interface to *api.Pod{}.
func ConvertObjToPod(obj interface{}) (*api.Pod, error) {
	pod, ok := obj.(*api.Pod)
	if !ok {
		return nil, fmt.Errorf("Assert error. Can't assert interface{} to *api.Pod")
	}
	return pod, nil
}

//Get process name from error message
func GetErrProcess(s string) string {
	port, err := getPort(s)
	if err != nil {
		return err.Error()
	}
	process, err := getProcess(port)
	if err != nil {
		return err.Error()
	}
	return process
}

func getPort(s string) (string, error) {
	reg, err := regexp.Compile(`(:\d+)]`)
	if err != nil {
		return "", err
	}
	port := reg.FindStringSubmatch(s)
	if len(port) < 2 {
		return "", errors.New("Capture port failed!")
	}

	return port[1], nil
}

func getProcess(port string) (string, error) {
	c1 := exec.Command("netstat", "-anp")
	c2 := exec.Command("grep", port)

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	err := c1.Start()
	if err != nil {
		return "", err
	}
	err = c2.Start()
	if err != nil {
		return "", err
	}
	err = c1.Wait()
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	err = c2.Wait()
	if err != nil {
		return "", err
	}

	return b2.String(), nil
}
