# API Testing Documentation

This directory contains comprehensive API testing tools and documentation for PeerVault.

## Overview

PeerVault provides multiple approaches to API testing:

1. **Interactive Testing** - Postman/Insomnia integration with pre-configured collections
2. **API Mocking** - Mock server generation from OpenAPI specifications
3. **Contract Testing** - Consumer-driven contract testing with Pact
4. **Performance Testing** - Load and stress testing tools
5. **Security Testing** - OWASP API security testing

## Quick Start

### Interactive Testing

```bash
# Import Postman collections
postman import tests/api/collections/peervault-postman.json

# Run automated tests
npm run test:api:postman
```

### API Mocking

```bash
# Start mock server
go run cmd/peervault-mock/main.go --config config/mock-server.yaml

# Generate mocks from OpenAPI spec
go run cmd/peervault-mock/main.go --generate --spec docs/api/peervault-rest-api.yaml
```

### Contract Testing

```bash
# Run contract tests
go test ./tests/contracts/...

# Verify provider contracts
pact-verifier --provider-base-url=http://localhost:3000 tests/contracts/
```

### Performance Testing

```bash
# Run load tests
go test -bench=. ./tests/performance/...

# Run stress tests
k6 run tests/performance/stress-test.js
```

### Security Testing

```bash
# Run security scans
go test ./tests/security/...

# Run OWASP ZAP scan
zap-baseline.py -t http://localhost:3000
```

## Collections

- `tests/api/collections/` - Postman/Insomnia collections
- `tests/contracts/` - Pact contract definitions
- `tests/performance/` - Load testing scripts
- `tests/security/` - Security testing tools

## Configuration

All testing tools can be configured via:

- Environment variables
- Configuration files in `config/`
- Command-line flags

## Integration

Testing tools integrate with:

- CI/CD pipelines
- Development workflows
- Monitoring systems
- Documentation generation
