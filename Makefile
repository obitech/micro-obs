.PHONY: all
all: build

.PHONY: build
build: test build-item

.PHONY: test
test:
	go test ./...

.PHONY: build-item
build-item:
	go build -o bin/item ./cmd/item

.PHONE: clean
clean:
	rm bin/*