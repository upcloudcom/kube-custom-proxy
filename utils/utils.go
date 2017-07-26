/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package utils

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func FileExist(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func JoinURL(url1 string, url2 string) (url string) {
	if strings.HasSuffix(url1, "/") {
		if strings.HasPrefix(url2, "/") {
			url = url1 + url2[1:]
		} else {
			url = url1 + url2
		}
	} else {
		if strings.HasPrefix(url2, "/") {
			url = url1 + url2
		} else {
			url = url1 + "/" + url2
		}
	}
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return
}

// Convert net.IP to int64 ,  http://www.sharejs.com
func inet_aton(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}
func GetIpHostNumber(ipaddress string) (string, error) {
	ip := net.ParseIP(ipaddress)
	if ip == nil {
		return "", errors.New("Fail to parse the ipaddress")
	}
	ipmark := ip.Mask(net.IPv4Mask(0, 0, 0x1f, 0xff)) //tc class, classid cannot be larger then 9999
	maskint := inet_aton(ipmark)
	host := strconv.FormatInt(maskint, 10)
	return host, nil
}

func GetPublicInterface() (string, error) { //Get ip
	serveraddr := "8.8.8.8:53"
	conn, err := net.Dial("tcp", serveraddr)
	// if failed, try at most 4 times
	for i := 0; err != nil && i < 3; i++ {
		time.Sleep(time.Second)
		conn, err = net.Dial("tcp", serveraddr)
	}
	if err != nil {
		return "", err
	}
	defer conn.Close()
	//return strings.Split(conn.LocalAddr().String(), ":")[0], nil
	publicIp := strings.Split(conn.LocalAddr().String(), ":")[0]
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip := strings.Split(addr.String(), "/")[0]
			if ip == publicIp {
				return iface.Name, nil
			}
		}
	}
	return "", nil
}

func GetLocalIPAddresses() []string {
	ifaces, _ := net.InterfaceAddrs()
	ips := []string{}
	for _, iface := range ifaces {
		ips = append(ips, iface.String())
	}
	return ips
}
func IsLocalIPAddress(ipaddress string) bool {
	ifaces, _ := net.InterfaceAddrs()
	for _, iface := range ifaces {
		ip := strings.Split(iface.String(), "/")[0]
		if ip == ipaddress {
			return true
		}
	}
	return false

}

func Debug(obj interface{}) string {
	strobj, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(strobj)
}
