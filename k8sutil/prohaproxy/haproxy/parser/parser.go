/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package parser

import (
	"fmt"
	"os"
	"strings"
	"tenx-proxy/config"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/cfgfile"
	"tenx-proxy/k8sutil/prohaproxy/haproxy/parts"

	"strconv"

	"github.com/golang/glog"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/labels"
)

const TenxCloudPortProtocol = "tenxcloud.com/schemaPortname"

type Parser struct {
	PublicIP       string
	Domain         string
	GroupID        string
	DefaultSSLCert string
}

func NewParser(instance cfgfile.HAProxyInstance) *Parser {
	p := new(Parser)
	p.Update(instance)
	return p
}

func (p *Parser) Update(instance cfgfile.HAProxyInstance) {
	p.PublicIP = instance.PublicIP()
	p.Domain = instance.Domain()
	p.GroupID = instance.GroupID()
	p.DefaultSSLCert = instance.DefaultSSLCert()
}

func (p Parser) Parse(service *v1.Service, pods []v1.Pod) []cfgfile.ConfigPart {
	if p.noneOfMyBusiness(service) {
		glog.V(4).Infoln("not responsible for this service", service)
		return nil
	}
	backendName := backendName(service)
	bindingDomains := strings.SplitN(service.ObjectMeta.Annotations["binding_domains"], ",", -1)
	sslEnable := service.ObjectMeta.Annotations["tenxcloud.com/https"]
	cfgParts := p.parsePods(backendName, service, pods)
	if cfgParts == nil || len(cfgParts) <= 0 {
		return nil
	}
	if len(bindingDomains) > 0 && bindingDomains[0] != "" {
		return append(cfgParts, p.parseBindingDomain(backendName, service, pods, bindingDomains, sslEnable)...)
	}
	subDomains := strings.SplitN(service.ObjectMeta.Annotations["sub_domain"], ",", -1)
	if len(subDomains) > 0 && subDomains[0] != "" {
		return append(cfgParts, p.parseSubDomain(backendName, service, pods, subDomains)...)
	}
	return cfgParts
}

func (p Parser) parseBindingDomain(backendName string, svc *v1.Service, pods []v1.Pod, bindingDomains []string, sslEnable string) []cfgfile.ConfigPart {
	pp := make([]cfgfile.ConfigPart, 0)
	if sslEnable == "true" {
		file := sslCertFile(svc.Namespace, svc.Name)
		if exist(file) {
			pp = append(pp, parts.NewSSLCert(file))
		} else {
			glog.V(2).Infof("ssl cert file not exist, file: %s", file)
		}
		pp = append(pp, parts.NewRedirectHTTPS(bindingDomains))
		pp = append(pp, parts.NewBindDomainHTTPS(backendName, bindingDomains))
	} else {
		pp = append(pp, parts.NewBindDomain(backendName, bindingDomains))
	}
	return pp
}

func (p Parser) parseSubDomain(backendName string, svc *v1.Service, pods []v1.Pod, subDomains []string) []cfgfile.ConfigPart {
	return []cfgfile.ConfigPart{parts.NewBindDomain(backendName, subDomains)}
}

const noBindingPort = -1

func bindingPort(svc *v1.Service) int32 {
	port := svc.Annotations["binding_port"]
	if port == "" {
		return noBindingPort
	}
	p64, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		glog.V(1).Infof("can not parse binding port, annotation: %s", port)
		return noBindingPort
	}
	return int32(p64)
}

func (p Parser) parsePods(backendName string, svc *v1.Service, pods []v1.Pod) []cfgfile.ConfigPart {
	portDetails := parsePortDetails(svc)
	myPods := filterPods(svc, pods)
	if len(myPods) <= 0 {
		return nil
	}
	podPorts := toPodPort(myPods)
	ppp := make([]cfgfile.ConfigPart, 0)
	bp := bindingPort(svc)
	if bp != noBindingPort {
		ppp = append(ppp, parts.NewBackend(backendName, podPorts, bp))
	}
	for _, port := range portDetails {
		backendName = fmt.Sprintf("%s-%s-%d", portName(port.ServicePort), svc.Namespace, port.ServicePort.TargetPort.IntVal)
		domainName := strings.TrimSuffix(fmt.Sprintf("%s-%s.%s", portName(port.ServicePort), svc.Namespace, p.Domain), ".")
		if strings.ToLower(port.Protocol) == "http" {
			ppp = append(ppp, parts.NewBindDomain(backendName, []string{domainName}), parts.NewBackend(backendName, podPorts, port.ServicePort.Port))
		} else if strings.ToLower(port.Protocol) == "tcp" {
			ppp = append(ppp, parts.NewListen(domainName, p.PublicIP, port.Port, port.ServicePort.Port, podPorts))
		}
	}
	return ppp
}

func toPodPort(pods []*v1.Pod) []parts.PodPort {
	length := len(pods)
	pp := make([]parts.PodPort, length)
	for i := 0; i < length; i++ {
		pp[i] = parts.PodPort{
			Name: pods[i].Name,
			IP:   pods[i].Status.PodIP,
		}
	}
	return pp
}

type port struct {
	Name        string
	Protocol    string
	Port        string
	ServicePort *v1.ServicePort
}

func parsePortDetails(svc *v1.Service) []*port {
	raw := strings.SplitN(svc.Annotations[TenxCloudPortProtocol], ",", -1)
	join := make(map[string]*port)
	for _, p := range raw {
		pp := strings.SplitN(p, "/", -1)
		length := len(pp)
		if length < 2 {
			glog.V(1).Infof("unknown port forward annotation, service name: %s, namespace: %s, annotation: %s",
				svc.Name, svc.Namespace, p)
		} else if length < 3 && strings.ToLower(pp[1]) == "tcp" {
			glog.V(1).Infof("no proxied port specified, service name: %s, namespace: %s, annotation: %s",
				svc.Name, svc.Namespace, p)
		} else {
			ppp := &port{
				Name:     pp[0],
				Protocol: pp[1],
			}
			if len(pp) > 2 {
				ppp.Port = pp[2]
			}
			join[ppp.Name] = ppp
		}
	}
	for _, sp := range svc.Spec.Ports {
		if ppp, exist := join[portName(&sp)]; exist {
			tsp := sp
			ppp.ServicePort = &tsp
		}
	}
	p := make([]*port, 0)
	for name, ppp := range join {
		if ppp.ServicePort != nil {
			p = append(p, ppp)
		} else {
			glog.V(1).Infof("no service port found meet the annotation, port name: %s, service name: %s, namespace: %s, annotation: %s",
				name, svc.Name, svc.Namespace, svc.Annotations[TenxCloudPortProtocol])
		}
	}
	return p
}

func filterPods(svc *v1.Service, pods []v1.Pod) []*v1.Pod {
	selector := labels.Set(svc.Spec.Selector).AsSelectorPreValidated()
	p := make([]*v1.Pod, 0)
	for _, pod := range pods {
		if svc.Namespace == pod.Namespace &&
			selector.Matches(labels.Set(pod.ObjectMeta.Labels)) &&
			pod.Status.Phase == "Running" &&
			len(pod.Spec.Containers) > 0 {
			pp := pod
			p = append(p, &pp)
		}
	}
	return p
}

func backendName(svc *v1.Service) string {
	first := svc.Spec.Ports[0]
	return fmt.Sprintf("%s-%s-%s", svc.Namespace, svc.Name, portName(&first))
}

func portName(port *v1.ServicePort) string {
	if port.Name == "" {
		return fmt.Sprintf("%d", port.Port)
	}
	return port.Name
}

type filter func(*v1.Service) bool

func (p Parser) noneOfMyBusiness(service *v1.Service) bool {
	f := []filter{
		p.inKubeSystemOrDefault,
		p.noPorts,
		p.noClusterIP,
		p.notMyJob,
		p.notTenxService,
	}
	for _, ff := range f {
		if ff(service) {
			return true
		}
	}
	return false
}

func (p Parser) inKubeSystemOrDefault(svc *v1.Service) bool {
	return svc.Namespace == "kube-system" || svc.Namespace == "default"
}

func (p Parser) noPorts(svc *v1.Service) bool {
	return len(svc.Spec.Ports) <= 0
}

func (p Parser) noClusterIP(svc *v1.Service) bool {
	return svc.Spec.ClusterIP == "None" || svc.Spec.ClusterIP == ""
}

func (p Parser) notMyJob(svc *v1.Service) bool {
	groupID, exist := svc.ObjectMeta.Annotations[config.GROUP_KEY_ANNOTAION]
	if !exist {
		// For old services, there is no lbgroup info, so every proxy will do the the job
		return false
	}
	if groupID == config.GROUP_VAR_NONE {
		return true
	}
	if groupID != p.GroupID {
		return true
	}
	return false
}

func (p Parser) notTenxService(svc *v1.Service) bool {
	_, exist := svc.Annotations[TenxCloudPortProtocol]
	return !exist
}

func sslCertFile(namespace, name string) string {
	return *config.CertDir + namespace + ".tenxsep." + name
}

func exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
