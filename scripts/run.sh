#!/bin/bash

# PeerVault Run Script for Unix-like systems
# Usage: ./scripts/run.sh [target] [args...]

set -e

TARGET=${1:-"main"}
shift || true
ARGS="$@"

echo "PeerVault Run Script for Unix-like systems"
echo "Target: $TARGET"

case $TARGET in
    "main")
        echo "Running main application (all-in-one)..."
        go run ./cmd/peervault
        ;;
    "node")
        NODE_ARGS=${ARGS:-"--listen :3000"}
        echo "Running individual node with args: $NODE_ARGS"
        go run ./cmd/peervault-node $NODE_ARGS
        ;;
    "demo")
        DEMO_ARGS=${ARGS:-"--target localhost:5000"}
        echo "Running demo client with args: $DEMO_ARGS"
        go run ./cmd/peervault-demo $DEMO_ARGS
        ;;
    *)
        echo "Unknown target: $TARGET"
        echo "Available targets: main, node, demo"
        exit 1
        ;;
esac
