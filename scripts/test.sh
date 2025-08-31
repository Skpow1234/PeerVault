#!/bin/bash

# PeerVault Test Script for Unix-like systems
# Usage: ./scripts/test.sh [type]

set -e

TYPE=${1:-"all"}

echo "PeerVault Test Script for Unix-like systems"
echo "Test Type: $TYPE"

case $TYPE in
    "unit")
        echo "Running unit tests..."
        go test -v ./internal/...
        ;;
    "integration")
        echo "Running integration tests..."
        if [ -d "./tests/integration" ]; then
            go test -v ./tests/integration/...
        else
            echo "Integration tests directory not found"
        fi
        ;;
    "race")
        echo "Running tests with race detector..."
        go test -race -v ./...
        ;;
    "fuzz")
        echo "Running fuzz tests..."
        go test -run ^$ -fuzz=Fuzz -fuzztime=30s ./internal/transport/p2p
        ;;
    "all")
        echo "Running all tests..."
        go test -v ./...
        ;;
    *)
        echo "Unknown test type: $TYPE"
        echo "Available types: all, unit, integration, race, fuzz"
        exit 1
        ;;
esac

echo "✓ All tests passed!"
