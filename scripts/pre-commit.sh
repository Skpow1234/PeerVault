#!/bin/bash

# Pre-commit script to fix common formatting issues
# Run this before committing: ./scripts/pre-commit.sh

set -e

echo "🔧 Running pre-commit checks..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Format Go code
echo "📝 Formatting Go code..."
go fmt ./...
goimports -w .

# Check for trailing whitespace in code files
echo "🧹 Checking for trailing whitespace in code files..."
if grep -r --include="*.go" --include="*.yml" --include="*.yaml" '[[:space:]]$' .; then
    echo "❌ Found trailing whitespace in code files. Please remove it."
    echo "💡 Run: find . -name '*.go' -o -name '*.yml' -o -name '*.yaml' | xargs sed -i 's/[[:space:]]*$//'"
    exit 1
fi

# Run linter
echo "🔍 Running linter..."
if command -v golangci-lint > /dev/null 2>&1; then
    golangci-lint run ./...
else
    echo "⚠️  golangci-lint not found. Skipping linting."
fi

# Run tests
echo "🧪 Running tests..."
go test ./tests/unit/...
go test ./internal/...

echo "✅ Pre-commit checks passed!"
echo "💡 You can now commit your changes."
