.PHONY: all build docker docker-build prepare test build-item clean

TAG=obitech/micro-obs:master
DOCKERFILE=Dockerfile

all: prepare test build

build: build-item

prepare:
	go mod tidy
	go fmt ./...
	go vet ./...
	golint ./...

test:
	go test -v ./...

build-item:
	go build -o bin/item ./cmd/item

docker: prepare test docker-build

docker-build:
	docker build -t $(TAG) -f $(DOCKERFILE) .

clean:
	rm bin/*