.PHONY: build

PREFIX = harbor.enncloud.cn/paas
FLAGS = 
VERSION=dev-bigdata

build:  
	cp ../tenx-proxy  .
	docker build -t $(PREFIX)/proxy:$(VERSION) -f Dockerfile-proxy .  
	rm tenx-proxy
	echo "docker push $(PREFIX)/proxy:$(VERSION)"
