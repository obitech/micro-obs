build: test build-item

test:
	go test ./...

build-item:
	GOBIN=$(shell pwd)/bin/item go install ./cmd/item