# Variables
BINARY_NAME=marketplace-mcp-server
DOCKER_IMAGE=marketplace-mcp-server
DOCKER_REGISTRY=xpkg.upbound.io/upbound
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

.PHONY: all build clean test deps docker docker-build docker-push help

# Default target
all: clean deps test build

# Build the binary
build:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

# Build for current platform
build-local:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

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
run:
	$(GOCMD) run main.go

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

# Push Docker image to registry
docker-push: docker-build
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

# Pull Docker image from registry
docker-pull:
	docker pull $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)

# Run Docker container
docker-run:
	docker run -it --rm -p 8765:8765 $(DOCKER_IMAGE):$(VERSION)

# Run Docker container from registry
docker-run-registry:
	docker run -it --rm -p 8765:8765 $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(VERSION)

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
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Release build
release: clean deps test build-all docker-build

# Deploy to registry
deploy: docker-push

# Development setup
dev-setup: install-tools deps

# Check if everything is ready for release
check: deps vet lint test

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Clean, download deps, test, and build"
	@echo "  build         - Build the binary for Linux"
	@echo "  build-local   - Build the binary for current platform"
	@echo "  build-all     - Cross-compile for multiple platforms"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Download dependencies"
	@echo "  deps-update   - Update dependencies"
	@echo "  run           - Run the application locally"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-push   - Build and push Docker image to registry"
	@echo "  docker-pull   - Pull Docker image from registry"
	@echo "  docker-run    - Run Docker container locally"
	@echo "  docker-run-registry - Run Docker container from registry"
	@echo "  lint          - Lint the code"
	@echo "  fmt           - Format the code"
	@echo "  vet           - Vet the code"
	@echo "  security      - Run security scan"
	@echo "  docs          - Generate documentation"
	@echo "  install-tools - Install development tools"
	@echo "  release       - Build release artifacts"
	@echo "  deploy        - Deploy to registry"
	@echo "  dev-setup     - Setup development environment"
	@echo "  check         - Run all checks (deps, vet, lint, test)"
	@echo "  help          - Show this help message" 