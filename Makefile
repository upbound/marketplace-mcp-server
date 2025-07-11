# Pull in .envrc file details, if it exists. This exists in the event you do
# not have direnv installed.
ifneq (,$(wildcard ./.envrc))
    include .envrc
    export
endif

# ====================================================================================
# Setup Project

PROJECT_NAME := marketplace-mcp-server
PROJECT_REPO := github.com/upbound/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64

# -include will silently skip missing files, which allows us
# to load those files with a target in the Makefile. If only
# "include" was used, the make command would fail and refuse
# to run a target until the include commands succeeded.
-include build/makelib/common.mk

# Variables
STDIO_BINARY_NAME=mcp-server
HTTP_BINARY_NAME=mcp-http
DOCKER_IMAGE=marketplace-mcp-server
REGISTRY_ORG ?= xpkg.upbound.io/upbound
VERSION?=latest

.PHONY: all build clean test deps docker docker-build docker-run help

# ====================================================================================
# Setup Go

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

GO_REQUIRED_VERSION = 1.24
GOLANGCILINT_VERSION = 2.2.0
GO111MODULE = on
GO_NOCOV = true
GO_SUBDIRS = cmd
GO_LINT_DIFF_TARGET ?= HEAD~
GO_LINT_ARGS ?= --fix
-include build/makelib/golang.mk

# Default target
all: clean deps test build

# Build the binaries
build-binaries:
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o ./bin/$(STDIO_BINARY_NAME) ./cmd/mcp-server
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o ./bin/$(HTTP_BINARY_NAME) ./cmd/mcp-http

build: build-binaries

# Build for current platform
build-local:
	$(GOBUILD) $(LDFLAGS) -o ./bin/$(STDIO_BINARY_NAME) ./cmd/mcp-server
	$(GOBUILD) $(LDFLAGS) -o ./bin/$(HTTP_BINARY_NAME) ./cmd/mcp-http

# Clean build artifacts
clean-local:
	$(GOCLEAN)
	rm -f ./bin/$(STDIO_BINARY_NAME) ./bin/$(HTTP_BINARY_NAME)

clean: clean-local

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
	docker buildx build --target stdio -t $(DOCKER_IMAGE):$(VERSION) . --load
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-build-http:
	docker buildx build --target http -t $(DOCKER_IMAGE)-http:$(VERSION) . --load
	docker tag $(DOCKER_IMAGE)-http:$(VERSION) $(DOCKER_IMAGE)-http:latest

docker-build: docker-build-stdio docker-build-http

publish-docker-stdio: docker-build-stdio
	@docker tag $(DOCKER_IMAGE):latest $(REGISTRY_ORG)/$(DOCKER_IMAGE):$(VERSION)
	@docker push $(REGISTRY_ORG)/$(DOCKER_IMAGE):$(VERSION)

publish: publish-docker-stdio

# Run Docker containers
docker-run-stdio:
	docker run -i --rm -v ~/.up:/mcp/.up:ro $(DOCKER_IMAGE):$(VERSION)

docker-run-http:
	docker run --rm -p 8765:8765 -v ~/.up:/mcp/.up:ro $(DOCKER_IMAGE)-http:$(VERSION)

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
help-repo:
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

help: help-repo