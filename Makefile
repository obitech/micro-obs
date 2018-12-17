.PHONY: all build docker docker-build prepare test build-item clean

TAG=obitech/micro-obs:master
DOCKERFILE=Dockerfile

all: prepare test build

build: build-item build-order

prepare:
	go mod tidy
	go fmt ./...
	go vet ./...
	golint ./...

test:
	go test -v ./...

build-item:
	go build -o bin/item ./cmd/item

build-order:
	go build -o bin/order ./cmd/order

docker: prepare test docker-build

docker-build:
	docker build -t $(TAG) -f $(DOCKERFILE) .

clean:
	rm bin/*