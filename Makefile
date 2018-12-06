TAG=obitech/micro-obs:latest
DOCKERFILE=Dockerfile

.PHONY: all
all: build

.PHONY: build
build: prepare test build-item

.PHONY: docker
docker: prepare test
	docker build -t $(TAG) -f $(DOCKERFILE) .

.PHONY: prepare
prepare:
	go mod tidy
	go fmt ./...
	go vet ./...

.PHONY:
test:
	go test ./...

.PHONY: build-item
build-item:
	go build -o bin/item ./cmd/item



.PHONE: clean
clean:
	rm bin/*