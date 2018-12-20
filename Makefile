.PHONY: all build docker docker-build prepare test build-item build-order build-dummy clean

TAG=obitech/micro-obs:master
DOCKERFILE=Dockerfile

all: prepare test build

build: build-item build-order build-dummy

prepare:
	go mod tidy
	go fmt ./...
	go vet ./...
	golint ./...

test:
	go test ./...

build-item:
	go build -o bin/item ./cmd/item

build-order:
	go build -o bin/order ./cmd/order

build-dummy:
	go build -o bin/dummy ./cmd/dummy

docker: prepare test docker-build

docker-build:
	docker build -t $(TAG) -f $(DOCKERFILE) .

clean:
	rm bin/*