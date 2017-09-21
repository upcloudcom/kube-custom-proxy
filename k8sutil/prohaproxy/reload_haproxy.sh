#!/bin/sh

# If we want to kill the old haproxy process, enable this line
#oldpid=$(ps aux | grep "[/]etc/haproxy/haproxy.cfg" | awk '{print $2}')
kill  -s SIGUSR1 $(cat /var/run/pidrunner.pid) 2>/dev/null
