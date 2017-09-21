#!/bin/sh

PidFile=/var/run/$(basename $0).pid
ChildPidFile=/var/run/haproxy.pid
echo "$$">${PidFile}
function reloadchild
{
	echo "Reloading"
	EXT_CMD=
	CHPID=
	if [ -f "$ChildPidFile" ]; then
		CHPID=$(cat ${ChildPidFile})
		if [ -n "$CHPID" ] && [ -n $(ps -o pid | grep  "$CHPID") ]; then
			CHPIDS=`ps aux|grep "haproxy -f"|grep -v grep|awk '{print $1}'`
			EXT_CMD="-sf $CHPIDS"  
		fi
	fi
	haproxy -f /etc/haproxy/haproxy.cfg -db ${EXT_CMD} &
	CHPID=$!
	echo "$CHPID" >${ChildPidFile}
}

while true
do
	reloadchild
	ps aux
	sleep 1
done
