# Variables
SERVER_BINARY_NAME=marketplace-mcp-server
PROXY_BINARY_NAME=marketplace-mcp-proxy
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
build-server:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o ./bin/$(SERVER_BINARY_NAME) ./cmd/mcp-server

build-proxy: build-server
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o ./bin/$(PROXY_BINARY_NAME) ./cmd/mcp-proxy

# Build for current platform
build-server-local:
	$(GOBUILD) $(LDFLAGS) -o ./bin/$(SERVER_BINARY_NAME) ./cmd/mcp-server

build-proxy-local: build-server-local
	$(GOBUILD) $(LDFLAGS) -o ./bin/$(PROXY_BINARY_NAME) ./cmd/mcp-proxy

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f ./bin/$(SERVER_BINARY_NAME) ./bin/$(PROXY_BINARY_NAME)

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
run-server:
	$(GOCMD) run ./cmd/mcp-server/main.go

run-proxy:
	$(GOCMD) run ./cmd/mcp-proxy/main.go

# Build Docker image
docker-build-stdio:
	docker build -t $(DOCKER_IMAGE):$(VERSION) --target=stdio .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-build-http:
	docker build -t $(DOCKER_IMAGE):$(VERSION) --target=http .
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
docker-run-proxy:
	docker run --rm -p 8765:8765 $(DOCKER_IMAGE):$(VERSION)

docker-run-local:
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
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(SERVER_BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(SERVER_BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(SERVER_BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(SERVER_BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(SERVER_BINARY_NAME)-windows-amd64.exe .

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
	@echo "  all           		  	- Clean, download deps, test, and build"
	@echo "  build-server         	- Build the binary for Linux"
	@echo "  build-server-local   	- Build the binary for current platform"
	@echo "  build-all     			- Cross-compile for multiple platforms"
	@echo "  clean         			- Clean build artifacts"
	@echo "  test          			- Run tests"
	@echo "  test-coverage 			- Run tests with coverage report"
	@echo "  deps          			- Download dependencies"
	@echo "  deps-update   			- Update dependencies"
	@echo "  run-server    			- Run the server locally"
	@echo "  run-proxy    			- Run the server with a proxy locally"
	@echo "  docker-build-stdio  	- Build Docker image for standalone mode"
	@echo "  docker-build-http		- Build Docker image for HTTP proxy mode"
	@echo "  docker-push   			- Build and push Docker image to registry"
	@echo "  docker-pull   			- Pull Docker image from registry"
	@echo "  docker-run-proxy		- Run Docker container locally with HTTP proxy"
	@echo "  docker-run-local		- Run Docker container locally in interactive mode"	
	@echo "  docker-run-registry 	- Run Docker container from registry"
	@echo "  lint          			- Lint the code"
	@echo "  fmt           			- Format the code"
	@echo "  vet           			- Vet the code"
	@echo "  security      			- Run security scan"
	@echo "  docs          			- Generate documentation"
	@echo "  install-tools 			- Install development tools"
	@echo "  release       			- Build release artifacts"
	@echo "  deploy        			- Deploy to registry"
	@echo "  dev-setup     			- Setup development environment"
	@echo "  check         			- Run all checks (deps, vet, lint, test)"
	@echo "  help          			- Show this help message" 