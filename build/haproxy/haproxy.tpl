# Licensed Materials - Property of tenxcloud.com
# (C) Copyright 2017 TenxCloud. All Rights Reserved.
# 2017-07-20  @author lizhen

global
    log 127.0.0.1 local2
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin
    stats timeout 30s
    user haproxy
    group haproxy
    daemon
    tune.ssl.default-dh-param 2048

defaults
    mode                    http
    log                     global
    option                  dontlognull
    option http-server-close
    option                  redispatch
    retries                 3
    timeout check           3s
    timeout client          900s
    timeout client-fin      3s
    timeout connect         5s
    timeout http-keep-alive 900s
    timeout http-request    60s
    timeout queue           300s
    timeout server          900s
    timeout server-fin      3s
    timeout tarpit          900s
    timeout tunnel          24h
    maxconn                 50000

listen stats
    bind *:8889
    mode http
    stats uri /tenx-stats
    stats realm Haproxy\ Statistics
    stats auth tenxcloud:haproxy-agent

{{with .FrontendLB}}
frontend LB
    mode http
    option forwardfor       except 127.0.0.0/8
    errorfile 503 /etc/haproxy/errors/503.http
    bind {{$.PublicIP}}:443 ssl crt {{.DefaultSSLCert}}{{range .SSLCerts}} crt {{.}}{{end}}{{range .Domains}}
    acl {{.BackendName}} hdr(host) -i {{range .DomainNames}} {{.}}{{end}}
    use_backend {{.BackendName}} if {{.BackendName}} { ssl_fc_sni{{range .DomainNames}} {{.}}{{end}} }{{end}}{{end}}

{{with .DefaultHTTP}}
listen defaulthttp
    bind {{$.PublicIP}}:80
    mode http
    option forwardfor       except 127.0.0.0/8
    errorfile 503 /etc/haproxy/errors/503.http{{range .Redirect}}
    redirect scheme https code 301 if { hdr(Host) -i {{range .DomainNames}} {{.}}{{end}} } !{ ssl_fc }{{end}}{{range .Domains}}
    acl {{.BackendName}} hdr(host) -i {{range .DomainNames}} {{.}}{{end}}
    use_backend {{.BackendName}} if {{.BackendName}}{{end}}{{end}}

{{with .Backend}}{{range .}}
backend {{.BackendName}}{{$port := .Port}}{{range .Pods}}
    cookie cookie.enncloud.cn insert indirect postonly
    server {{.Name}} {{.IP}}:{{$port}} cookie {{.Name}} check maxconn 2000{{end}}{{end}}{{end}}

{{with .Listen}}{{range .}}
listen {{.DomainName}}
    bind {{$.PublicIP}}:{{.PublicPort}}
    mode tcp
    balance roundrobin{{$port := .Port}}{{range .Pods}}
    server {{.Name}} {{.IP}}:{{$port}} maxconn 2000{{end}}{{end}}{{end}}
