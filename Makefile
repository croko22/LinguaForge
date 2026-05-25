.PHONY: lint sec test qa fmt build

lint:
	golangci-lint run ./...

sec:
	gosec ./...

test:
	go test -v ./...

qa: lint sec test

fmt:
	go fmt ./...

build:
	go build ./cmd/api/
