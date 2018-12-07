.PHONY: all build docker docker-build prepare test build-item clean

TAG=obitech/micro-obs:latest
DOCKERFILE=Dockerfile

all: build

build: prepare test build-item

docker: prepare test docker-build

docker-build:
	docker build -t $(TAG) -f $(DOCKERFILE) .

prepare:
	go mod tidy
	go fmt ./...
	go vet ./...

test:
	go test ./...

build-item:
	go build -o bin/item ./cmd/item

clean:
	rm bin/*