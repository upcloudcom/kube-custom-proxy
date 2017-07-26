#!/bin/bash
#haproxy -f /root/gopath/src/tenx-proxy/k8sutil/haproxy/haproxy.cfg
#service haproxy reload
kill  -s SIGUSR1 $(cat /var/run/pidrunner.pid) 2>/dev/null
