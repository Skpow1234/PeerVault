# Contributing to PeerVault

Thank you for your interest in contributing to PeerVault! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Documentation](#documentation)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a feature branch** for your changes
4. **Make your changes** following the guidelines below
5. **Test your changes** thoroughly
6. **Submit a pull request**

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- Docker (for containerization tests)
- Make (optional, for build automation)

### Local Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/Peervault.git
cd Peervault

# Add upstream remote
git remote add upstream https://github.com/Skpow1234/Peervault.git

# Install dependencies
go mod download

# Run tests to verify setup
go test ./...
```

### Development Tools

We use several tools for code quality:

```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Run linting
golangci-lint run ./...

# Format code
go fmt ./...
goimports -w .
```

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feature/descriptive-feature-name`
- `bugfix/issue-description`
- `docs/documentation-update`
- `refactor/code-improvement`

### Commit Messages

Follow conventional commit format:

```bash
type(scope): description

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build/tooling changes

Examples:

```bash
feat(storage): add encryption at rest support
fix(transport): resolve connection timeout issue
docs(readme): update installation instructions
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific test categories
go test ./tests/unit/...
go test ./tests/integration/...
go test ./tests/fuzz/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. ./tests/integration/performance/...
```

### Test Guidelines

1. **Write tests for new features**
2. **Maintain test coverage above 80%**
3. **Use descriptive test names**
4. **Test both success and failure cases**
5. **Use table-driven tests for multiple scenarios**
6. **Mock external dependencies**

### Example Test Structure

```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FeatureFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FeatureFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("FeatureFunction() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Pull Request Process

### Before Submitting

1. **Ensure all tests pass**
2. **Run linting and fix issues**
3. **Update documentation if needed**
4. **Add tests for new functionality**
5. **Update CHANGELOG.md if applicable**

### PR Template

Use the provided PR template and fill in all sections:

- **Description**: What does this PR do?
- **Type of change**: Bug fix, feature, etc.
- **Testing**: How was this tested?
- **Checklist**: Ensure all items are completed

### Review Process

1. **Automated checks must pass**
2. **At least one maintainer must approve**
3. **All conversations must be resolved**
4. **Code review feedback must be addressed**

## Code Style

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines:

- **Package names**: Short, clear, no underscores
- **Function names**: MixedCaps or mixedCaps
- **Variable names**: Short in small scopes, descriptive in large scopes
- **Constants**: UPPER_CASE for exported, lower_case for unexported

### Code Organization

```bash
internal/
├── app/          # Application logic
├── crypto/       # Cryptographic functions
├── logging/      # Logging utilities
├── peer/         # Peer management
├── storage/      # Storage layer
└── transport/    # Network transport

cmd/
├── peervault/    # Main application
├── peervault-node/ # Node binary
└── peervault-demo/ # Demo client

tests/
├── unit/         # Unit tests
├── integration/  # Integration tests
├── fuzz/         # Fuzz tests
└── utils/        # Test utilities
```

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// Bad
if err != nil {
    return err
}
```

### Logging

```go
// Use structured logging
logger.Info("operation completed",
    "duration", duration,
    "items_processed", count,
    "status", "success",
)
```

## Documentation

### Code Comments

- **Export all exported functions, types, and packages**
- **Use complete sentences**
- **Start with the name of the thing being documented**

```go
// ProcessData handles the processing of input data and returns
// the processed result. It returns an error if processing fails.
func ProcessData(input []byte) ([]byte, error) {
    // Implementation
}
```

### README Updates

- **Update README.md for user-facing changes**
- **Add examples for new features**
- **Update installation instructions if needed**

## Reporting Issues

### Bug Reports

Use the bug report template and include:

- **Clear description** of the issue
- **Steps to reproduce**
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, etc.)
- **Logs or error messages**

### Feature Requests

Use the feature request template and include:

- **Clear description** of the feature
- **Use case** and motivation
- **Proposed implementation** (if applicable)
- **Alternatives considered**

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Documentation**: Check existing docs first
- **Code**: Look at existing code for examples

## Recognition

Contributors will be recognized in:

- **CHANGELOG.md** for significant contributions
- **README.md** for major contributors
- **Release notes** for each release

Thank you for contributing to PeerVault!
