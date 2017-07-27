
CURRENTPATH=$(shell pwd)
Version=v1.0.0.rc
.PHONY: update-deps build docker-build docker-build-old

all: build

update-deps:
	go get github.com/tools/godep
	export GOPATH=$(GOPATH):$(shell cd ../../;pwd);go get -v
dep-save:
	export GOPATH=$(GOPATH):$(shell cd ../../;pwd);godep save ./...
build:
	export GOPATH=$(GOPATH):$(shell cd ../../;pwd);CGO_ENABLED=0 GOARCH=amd64 GOOS=linux godep go build --ldflags '-extldflags "-static"'

build-ppc64le:
	docker run --rm --workdir=/go/src/admin-console --volume=/root/code/src:/go/src ppc64le/golang:1.5 go build

docker-build: clean
	docker run --rm -v $(shell pwd):/go/src/$(shell basename $(shell pwd))  \
	-w /go/src/$(shell basename $(shell pwd)) index.tenxcloud.com/tenx_containers/godep:alpine  	\
	/bin/sh -c "CGO_ENABLED=0 godep go build --ldflags '-extldflags \"-static\"' "


container: 
	docker build -t index.tenxcloud.com/tenxcloud/tenx-proxy:${Version} -f Dockerfile .

push:
	docker push index.tenxcloud.com/tenx_containers/admin-console:${Version}

docker-build-ppc64le:
	#go get -v, need to manual get, ignore the errors at ppc64le platform (because gccgo-go is not different from go released by google)
	docker run --rm --workdir=/go/src/admin-console --volume=/root/code/src:/go/src ppc64le/golang:1.5 go build
	docker build -t index.tenxcloud.com/tenx_containers/admin-console-ppc64le:${Version} -f Dockerfile.ppc64le .
	
docker-push-ppc64le:
	docker push index.tenxcloud.com/tenx_containers/admin-console:${Version}
clean:
	rm -f ./$(shell basename $(shell pwd)) >/dev/null
