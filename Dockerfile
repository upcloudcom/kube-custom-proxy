FROM alpine:3.6

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories
RUN apk update && apk upgrade
RUN apk add --no-cache tzdata supervisor haproxy
RUN ln -snf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN mkdir -p /etc/sslkeys
RUN mkdir -p /run/haproxy
ADD build/haproxy/haproxy.tpl /etc/default/hafolder/haproxy.tpl
ADD build/haproxy/reload_haproxy.sh /etc/default/hafolder/reload_haproxy.sh
ADD build/haproxy/haproxy.cfg /etc/haproxy/haproxy.cfg
ADD build/haproxy/default.pem /etc/sslkeys/default.pem
ADD build/haproxy/503.http /etc/haproxy/errors/503.http
ADD build/pidrunner /pidrunner
ADD build/run.sh /run.sh
ADD build/test   /root/
ADD tenx-proxy /opt/bin/tenx-proxy

CMD ["/run.sh"]
