.PHONY: lint sec test qa fmt build frontend-dev frontend-build frontend-test

lint:
	command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || go vet ./...

sec:
	command -v gosec >/dev/null 2>&1 && gosec ./... || echo "gosec not installed, skipping"

test:
	go test -v ./...

qa: lint sec test

fmt:
	go fmt ./...

build:
	go build ./cmd/api/

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-test:
	cd frontend && npm run test
