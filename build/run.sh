#!/bin/sh

if [ $# -lt 1 ]; then
    echo usage: $0 "tenx-proxy options"
    exit 1
fi

 

echo tenx tc opt $@

template=$(cat <<EOF
[supervisord]
# run in foreground
nodaemon = true
pidfile = /tmp/supervisord.pid
logfile = /tmp/supervisord.log

[inet_http_server]
port = 0.0.0.0:60000

[program:tenx-proxy]
command=/opt/bin/tenx-proxy --hafolder=/etc/default/hafolder $@

startretries=99999

stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5

stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5

[program:haproxy]
command=/pidrunner

startretries=99999

stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5

stderr_logfile_maxbytes=10MB
stderr_logfile_backups=5
EOF
        )

echo "${template}" > supervisord.conf

supervisord -c supervisord.conf
