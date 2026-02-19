BINARY_NAME=autofix
VERSION=1.0.0

GO=go
GOFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: all build clean test

all: build

build: build-amd64 build-arm64

build-amd64:
	@echo "Building amd64 binaries..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/autofix
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/autofix

build-arm64:
	@echo "Building arm64 binaries..."
	GOARCH=arm64 GOOS=linux $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/autofix
	GOARCH=arm64 GOOS=darwin $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/autofix

build-local:
	@echo "Building for local platform..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/autofix

clean:
	@echo "Cleaning..."
	rm -rf $(BINARY_NAME)*

test:
	@echo "Running tests..."
	$(GO) test ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

lint:
	@echo "Linting..."
	golangci-lint run