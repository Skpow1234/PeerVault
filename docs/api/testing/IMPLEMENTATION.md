# API Testing Implementation Guide

This document provides a comprehensive guide to the PeerVault API testing implementation for Milestone 13.

## Overview

The API testing implementation includes five major components:

1. **Interactive Testing** - Postman/Insomnia integration
2. **API Mocking** - Mock server generation from OpenAPI specs
3. **Contract Testing** - Consumer-driven contract testing
4. **Performance Testing** - Load and stress testing
5. **Security Testing** - OWASP API security testing

## Architecture

```bash
docs/api/testing/          # Documentation
├── README.md              # Overview and quick start
├── IMPLEMENTATION.md      # This file
└── ...

tests/api/                 # Test collections and data
├── collections/           # Postman/Insomnia collections
├── mock-data/            # Mock server data
└── ...

tests/contracts/           # Contract definitions
├── health_contract.json   # Health check contract
└── ...

tests/performance/         # Performance tests
├── load_test.go          # Go performance tests
├── load-test.js          # k6 performance tests
└── ...

tests/security/           # Security tests
├── security_test.go      # Security test implementation
└── ...

internal/api/             # Core implementation
├── mocking/              # Mock server implementation
├── contracts/            # Contract testing
├── performance/          # Performance testing
├── security/             # Security testing
└── ...

cmd/peervault-mock/       # Mock server binary
└── main.go              # Mock server entry point

scripts/api-testing/      # Testing scripts
└── run-tests.sh         # Comprehensive test runner
```

## Components

### 1. Interactive Testing

**Location**: `tests/api/collections/`

**Features**:

- Postman collection with pre-configured requests
- Environment variables for different environments
- Automated test assertions
- Request/response logging

**Usage**:

```bash
# Import into Postman
postman import tests/api/collections/peervault-postman.json

# Run with Newman
newman run tests/api/collections/peervault-postman.json
```

### 2. API Mocking

**Location**: `internal/api/mocking/`, `cmd/peervault-mock/`

**Features**:

- Mock server generation from OpenAPI specifications
- Scenario-based response customization
- Request matching with conditions
- Analytics and monitoring
- YAML/JSON configuration

**Usage**:

```bash
# Start mock server
go run cmd/peervault-mock/main.go --config config/mock-server.yaml

# Generate scenarios from OpenAPI spec
go run cmd/peervault-mock/main.go --generate --spec docs/api/peervault-rest-api.yaml
```

### 3. Contract Testing

**Location**: `internal/api/contracts/`, `tests/contracts/`

**Features**:

- Consumer-driven contract testing
- Request/response validation
- Schema validation
- Pact integration support
- Contract evolution tracking

**Usage**:

```bash
# Run contract tests
go test ./tests/contracts/...

# Verify with Pact
pact-verifier --provider-base-url=http://localhost:3000 tests/contracts/
```

### 4. Performance Testing

**Location**: `internal/api/performance/`, `tests/performance/`

**Features**:

- Load testing with configurable concurrency
- Stress testing capabilities
- Response time analysis
- Throughput measurement
- k6 integration for advanced scenarios

**Usage**:

```bash
# Run Go performance tests
go test -bench=. ./tests/performance/...

# Run k6 load tests
k6 run tests/performance/load-test.js
```

### 5. Security Testing

**Location**: `internal/api/security/`, `tests/security/`

**Features**:

- OWASP Top 10 API security testing
- Injection vulnerability testing
- Authentication and authorization testing
- Security header validation
- Custom security test framework

**Usage**:

```bash
# Run security tests
go test ./tests/security/...

# Run OWASP ZAP scan
zap-baseline.py -t http://localhost:3000
```

## Configuration

### Mock Server Configuration

**File**: `config/mock-server.yaml`

```yaml
port: 3001
host: "localhost"
openapi_spec: "docs/api/peervault-rest-api.yaml"
mock_data_dir: "tests/api/mock-data"
response_delay: 100ms
enable_analytics: true
```

### Test Configuration

**Environment Variables**:

- `BASE_URL`: Base URL for API testing (default: <http://localhost:3000>)
- `MOCK_URL`: Mock server URL (default: <http://localhost:3001>)
- `TEST_TIMEOUT`: Test timeout duration (default: 30s)
- `VERBOSE`: Enable verbose output (default: false)

## Running Tests

### Comprehensive Test Suite

```bash
# Run all tests
./scripts/api-testing/run-tests.sh

# Run with custom configuration
./scripts/api-testing/run-tests.sh --base-url http://localhost:8080 --verbose
```

### Individual Test Suites

```bash
# Unit tests
go test ./tests/contracts/... ./tests/performance/... ./tests/security/...

# Contract tests
go test ./tests/contracts/...

# Performance tests
go test -bench=. ./tests/performance/...

# Security tests
go test ./tests/security/...

# Postman tests
newman run tests/api/collections/peervault-postman.json
```

## Integration

### CI/CD Integration

The testing framework integrates with CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run API Tests
  run: |
    ./scripts/api-testing/run-tests.sh --base-url ${{ env.API_URL }}
```

### Development Workflow

1. **Development**: Use mock server for frontend development
2. **Testing**: Run contract tests to ensure API compatibility
3. **Performance**: Run load tests to validate performance
4. **Security**: Run security tests to identify vulnerabilities
5. **Deployment**: Run comprehensive test suite before deployment

## Extending the Framework

### Adding New Test Types

1. Create test implementation in appropriate `internal/api/` directory
2. Add test files in corresponding `tests/` directory
3. Update test runner script
4. Add documentation

### Adding New Mock Scenarios

1. Create scenario file in `tests/api/mock-data/`
2. Update mock server configuration
3. Test scenario with mock server

### Adding New Security Tests

1. Implement test in `internal/api/security/`
2. Add test case in `tests/security/`
3. Update OWASP test categories if needed

## Best Practices

1. **Test Isolation**: Each test should be independent
2. **Data Management**: Use test fixtures and mock data
3. **Error Handling**: Proper error handling and reporting
4. **Performance**: Optimize test execution time
5. **Documentation**: Keep tests and documentation in sync
6. **Maintenance**: Regular updates and cleanup

## Troubleshooting

### Common Issues

1. **Service Not Running**: Ensure PeerVault API is running
2. **Mock Server Issues**: Check configuration and port availability
3. **Test Failures**: Review logs and check service health
4. **Performance Issues**: Adjust test parameters and timeouts

### Debug Mode

```bash
# Enable verbose output
VERBOSE=true ./scripts/api-testing/run-tests.sh

# Run individual tests with debug
go test -v ./tests/contracts/...
```

## Future Enhancements

1. **GraphQL Testing**: Enhanced GraphQL-specific testing
2. **WebSocket Testing**: Real-time communication testing
3. **Database Testing**: Database integration testing
4. **Monitoring Integration**: Real-time monitoring and alerting
5. **Test Reporting**: Enhanced reporting and visualization
