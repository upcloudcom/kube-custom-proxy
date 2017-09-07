.PHONY: build

PREFIX=harbor.enncloud.cn/paas
VERSION=v2.2.2-enn-v3

build:  
	cp ../tenx-proxy  .
	docker build -t $(PREFIX)/proxy:$(VERSION) -f Dockerfile-proxy .  
	rm tenx-proxy
	echo "docker push $(PREFIX)/proxy:$(VERSION)"
