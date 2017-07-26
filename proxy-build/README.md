# build and run supervisor-haproxy image

## build
```sh
go install -v tenx-proxy
cp $GOPATH/bin/tenx-proxy .
docker build -f Dockerfile-proxy -t proxy:0.1 .
```

## run
```sh
# run as service proxy
docker run --name proxy --net=host -v /var/run/docker.sock:/var/run/docker.sock proxy:0.1 /run.sh "--master=192.168.1.92:6443 --plugins=prohaproxy --watch=watchsrvs --externalIP=192.168.1.103 --bindingDomain=test.tenxcloud.com --emailReceiver=service@tenxcloud.com"

# run as webterminal proxy
docker run --name proxy --net=host -v /var/run/docker.sock:/var/run/docker.sock proxy:0.1 /run.sh "--master=192.168.1.92:6443 --plugins=webhaproxy --emailReceiver=service@tenxcloud.com"
```

# build and run supervisor-gotty image

## build
```sh
docker build -f Dockerfile-gotty -t gotty:0.1 .
```

## run
```sh
docker run --name gotty -v /var/run/docker.sock:/var/run/docker.sock gotty:0.1 /run.sh "--master=192.168.1.93:6443 --emailReceiver=service@tenxcloud.com"
```
# dist

supervisor version 3.2.0

haproxy version 1.6.3
