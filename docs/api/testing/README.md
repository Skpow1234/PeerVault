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

Testing runs in CI using containerized workflows. Local non-Docker commands have been removed.

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
