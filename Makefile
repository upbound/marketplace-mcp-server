# Variables
STDIO_BINARY_NAME=mcp-server
HTTP_BINARY_NAME=mcp-http
DOCKER_IMAGE=marketplace-mcp-server
VERSION?=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-a -installsuffix cgo

.PHONY: all build clean test deps docker docker-build docker-run help

# Default target
all: clean deps test build

# Build the binaries
build:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o ./bin/$(STDIO_BINARY_NAME) ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o ./bin/$(HTTP_BINARY_NAME) ./cmd/mcp-http

# Build for current platform
build-local:
	$(GOBUILD) $(LDFLAGS) -o ./bin/$(STDIO_BINARY_NAME) ./cmd/mcp-server
	$(GOBUILD) $(LDFLAGS) -o ./bin/$(HTTP_BINARY_NAME) ./cmd/mcp-http

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f ./bin/$(STDIO_BINARY_NAME) ./bin/$(HTTP_BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

# Run the application locally
run-stdio:
	$(GOCMD) run ./cmd/mcp-server/main.go

run-http:
	$(GOCMD) run ./cmd/mcp-http/main.go

# Build Docker images
docker-build-stdio:
	docker build --target stdio -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-build-http:
	docker build --target http -t $(DOCKER_IMAGE)-http:$(VERSION) .
	docker tag $(DOCKER_IMAGE)-http:$(VERSION) $(DOCKER_IMAGE)-http:latest

docker-build: docker-build-stdio docker-build-http

# Run Docker containers
docker-run-stdio:
	docker run -i --rm -v ~/.up:/mcp/.up:ro $(DOCKER_IMAGE):$(VERSION)

docker-run-http:
	docker run --rm -p 8765:8765 -v ~/.up:/mcp/.up:ro $(DOCKER_IMAGE)-http:$(VERSION)

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	$(GOCMD) fmt ./...

# Vet the code
vet:
	$(GOCMD) vet ./...

# Security scan
security:
	gosec ./...

# Generate documentation
docs:
	godoc -http=:6060

# Install development tools
install-tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec

# Cross-compile for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(STDIO_BINARY_NAME)-linux-amd64 ./cmd/mcp-server
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(STDIO_BINARY_NAME)-linux-arm64 ./cmd/mcp-server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(STDIO_BINARY_NAME)-darwin-amd64 ./cmd/mcp-server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(STDIO_BINARY_NAME)-darwin-arm64 ./cmd/mcp-server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(STDIO_BINARY_NAME)-windows-amd64.exe ./cmd/mcp-server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(HTTP_BINARY_NAME)-linux-amd64 ./cmd/mcp-http
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(HTTP_BINARY_NAME)-linux-arm64 ./cmd/mcp-http
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(HTTP_BINARY_NAME)-darwin-amd64 ./cmd/mcp-http
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(HTTP_BINARY_NAME)-darwin-arm64 ./cmd/mcp-http
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(HTTP_BINARY_NAME)-windows-amd64.exe ./cmd/mcp-http

# Release build
release: clean deps test build-all docker-build

# Development setup
dev-setup: install-tools deps

# Check if everything is ready for release
check: deps vet lint test

# Run a local mcp inspector
inspector:
	@DANGEROUSLY_OMIT_AUTH=true npx @modelcontextprotocol/inspector

# Help target
help:
	@echo "Available targets:"
	@echo "  all                - Clean, download deps, test, and build"
	@echo "  build              - Build both stdio and HTTP binaries for Linux"
	@echo "  build-local        - Build both binaries for current platform"
	@echo "  build-all          - Cross-compile for multiple platforms"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  deps               - Download dependencies"
	@echo "  deps-update        - Update dependencies"
	@echo "  run-stdio          - Run the stdio server locally"
	@echo "  run-http           - Run the HTTP server locally"
	@echo "  docker-build       - Build both Docker images"
	@echo "  docker-build-stdio - Build stdio Docker image"
	@echo "  docker-build-http  - Build HTTP Docker image"
	@echo "  docker-run-stdio   - Run stdio Docker container"
	@echo "  docker-run-http    - Run HTTP Docker container on port 8765"
	@echo "  lint               - Lint the code"
	@echo "  fmt                - Format the code"
	@echo "  vet                - Vet the code"
	@echo "  security           - Run security scan"
	@echo "  docs               - Generate documentation"
	@echo "  install-tools      - Install development tools"
	@echo "  release            - Build release artifacts"
	@echo "  dev-setup          - Setup development environment"
	@echo "  check              - Run all checks (deps, vet, lint, test)"
	@echo "  help               - Show this help message" 
	@echo "  inspector          - Run a local mcp inspector" 