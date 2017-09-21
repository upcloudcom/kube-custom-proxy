PREFIX = harbor.enncloud.cn/paas
TAG = v2.2.0-v4-test

FLAGS = 
PROJECT_DIR=$(shell cd ../../;pwd)
SOURCE_DIR=$(shell pwd)
PROJECT=$(shell basename $(SOURCE_DIR))
IMAGE=$(PREFIX)/proxy:$(TAG)

.PHONY: build host-build image push

build:
	docker run --rm -v $(SOURCE_DIR):/go/src/$(PROJECT) -w /go/src/$(PROJECT) golang:1.8.1 /bin/sh -c "CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -gcflags \"-N -l\""

host-build:
	GOPATH=$(PROJECT_DIR);CGO_ENABLED=0 go build 

image:
	docker build -t $(IMAGE)  .
	echo "docker push $(IMAGE)"

push:
	docker push $(IMAGE)
