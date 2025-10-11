# PeerVault Makefile
# Cross-platform build and development tasks

.PHONY: help build build-all run test clean fmt lint docker-build docker-run build-cli

# Default target
help:
	@echo "PeerVault Development Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build        - Build main application"
	@echo "  build-all    - Build all binaries (main, node, demo, config, cli)"
	@echo "  build-node   - Build individual node binary"
	@echo "  build-demo   - Build demo client binary"
	@echo "  build-config - Build configuration management tool"
	@echo "  build-cli    - Build CLI tool"
	@echo ""
	@echo "Run Commands:"
	@echo "  run          - Run main application (all-in-one)"
	@echo "  run-node     - Run individual node"
	@echo "  run-demo     - Run demo client"
	@echo "  run-cli      - Run CLI tool"
	@echo ""
	@echo "Test Commands:"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-race    - Run tests with race detector"
	@echo "  test-fuzz    - Run fuzz tests"
	@echo ""
	@echo "Development Commands:"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format Go code"
	@echo "  lint         - Run linter"
	@echo "  mod-tidy     - Tidy Go modules"
	@echo ""
	@echo "Docker Commands:"
	@echo "  docker-build - Build all Docker images"
	@echo "  docker-run   - Run multi-container setup"
	@echo "  docker-dev   - Run development setup"
	@echo "  docker-stop  - Stop all containers"
	@echo "  docker-clean - Clean Docker resources"

# Build targets
build:
	@echo "Building main application..."
	@mkdir -p bin
	@go build -o bin/peervault ./cmd/peervault
	@echo "✓ Main application built successfully"

build-node:
	@echo "Building node binary..."
	@mkdir -p bin
	@go build -o bin/peervault-node ./cmd/peervault-node
	@echo "✓ Node binary built successfully"

build-demo:
	@echo "Building demo client..."
	@mkdir -p bin
	@go build -o bin/peervault-demo ./cmd/peervault-demo
	@echo "✓ Demo client built successfully"

build-config:
	@echo "Building configuration tool..."
	@mkdir -p bin
	@go build -o bin/peervault-config ./cmd/peervault-config
	@echo "✓ Configuration tool built successfully"

build-cli:
	@echo "Building CLI tool..."
	@mkdir -p bin
	@go build -o bin/peervault-cli ./cmd/peervault-cli
	@echo "✓ CLI tool built successfully"

build-all: build build-node build-demo build-config build-cli
	@echo "✓ All binaries built successfully"

# Run targets
run: build
	@echo "Running main application..."
	@./bin/peervault

run-node: build-node
	@echo "Running individual node..."
	@./bin/peervault-node --listen :3000

run-demo: build-demo
	@echo "Running demo client..."
	@./bin/peervault-demo --target localhost:5000

run-cli: build-cli
	@echo "Running CLI tool..."
	@./bin/peervault-cli

# Test targets
test:
	@echo "Running all tests..."
	@go test -v ./...

test-unit:
	@echo "Running unit tests..."
	@go test -v ./internal/...

test-race:
	@echo "Running tests with race detector..."
	@go test -race -v ./...

test-fuzz:
	@echo "Running fuzz tests..."
	@go test -run ^$ -fuzz=Fuzz -fuzztime=30s ./internal/transport/p2p

# Development targets
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf *_network/
	@rm -rf node*_network/
	@rm -rf demo-client-data/
	@rm -rf peervault-*_data/
	@go clean -cache -testcache
	@echo "✓ Cleaned build artifacts"

fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "✓ Code formatted"

lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

mod-tidy:
	@echo "Tidying Go modules..."
	@go mod tidy
	@go mod verify
	@echo "✓ Modules tidied"

# Docker targets
docker-build:
	@echo "Building Docker images..."
	@docker build -t peervault .
	@docker build -f Dockerfile.node -t peervault-node .
	@docker build -f Dockerfile.demo -t peervault-demo .
	@echo "✓ Docker images built"

docker-run:
	@echo "Running multi-container setup..."
	@docker-compose up --build

docker-dev:
	@echo "Running development setup..."
	@docker-compose -f docker-compose.dev.yml up --build

docker-stop:
	@echo "Stopping containers..."
	@docker-compose down

docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v
	@docker system prune -f
	@echo "✓ Docker resources cleaned"

# Quick start
quick-start: mod-tidy build-all test-unit
	@echo "✓ Quick start completed successfully!"