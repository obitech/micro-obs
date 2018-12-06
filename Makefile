TAG=obitech/micro-obs:latest
DOCKERFILE=Dockerfile

.PHONY: all
all: build

.PHONY: build
build: prepare test build-item

.PHONY: prepare
prepare:
	go fmt ./...
	go vet ./...

.PHONY:
test:
	go test ./...

.PHONY: build-item
build-item:
	go build -o bin/item ./cmd/item

.PHONY: docker
docker: 
	docker build -t $(TAG) -f $(DOCKERFILE) .

.PHONE: clean
clean:
	rm bin/*