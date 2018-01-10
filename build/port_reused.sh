#!/bin/sh
while true;do
	pid=`netstat -lntp |grep ":80"|grep "/haproxy"|awk '{print $7}'|sort -r|awk -F "/" '{ if (NR==2) print $1 }'`
	if [[ "$pid" != "" ]];then
			echo "kill -9 $pid"
			kill -9  $pid
	fi
	sleep 60
done
