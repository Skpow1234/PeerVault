#!/bin/bash

# Shugo Build Script for Unix-like systems
# Usage: ./scripts/build.sh [target]

set -e

TARGET=${1:-"all"}

echo "Shugo Build Script for Unix-like systems"
echo "Target: $TARGET"

# Create bin directory if it doesn't exist
mkdir -p bin

case $TARGET in
    "main")
        echo "Building main application..."
        go build -o bin/shugo ./cmd/shugo
        echo "Main application built successfully"
        ;;
    "node")
        echo "Building node binary..."
        go build -o bin/shugo-node ./cmd/shugo-node
        echo "Node binary built successfully"
        ;;
    "demo")
        echo "Building demo client..."
        go build -o bin/shugo-demo ./cmd/shugo-demo
        echo "Demo client built successfully"
        ;;
    "all")
        echo "Building all binaries..."
        ./scripts/build.sh main
        ./scripts/build.sh node
        ./scripts/build.sh demo
        echo "All binaries built successfully"
        ;;
    "clean")
        echo "Cleaning build artifacts..."
        rm -rf bin/
        rm -rf *_network/
        rm -rf node*_network/
        rm -rf demo-client-data/
        rm -rf shugo-*_data/
        go clean -cache -testcache
        echo "Cleaned build artifacts"
        ;;
    *)
        echo "Unknown target: $TARGET"
        echo "Available targets: main, node, demo, all, clean"
        exit 1
        ;;
esac

echo "Build completed!"
