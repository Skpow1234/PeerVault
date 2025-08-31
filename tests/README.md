# PeerVault Testing

This directory contains all tests for the PeerVault project, organized by type and functionality.

## Directory Structure

```bash
tests/
├── README.md              # This file
├── unit/                  # Unit tests for individual components
│   ├── crypto/           # Cryptographic function tests
│   ├── storage/          # Storage system tests
│   ├── transport/        # Transport layer tests
│   └── peer/             # Peer management tests
├── integration/          # Integration tests
│   ├── end-to-end/       # Full system tests
│   ├── multi-node/       # Multi-node scenarios
│   └── performance/      # Performance and stress tests
├── fuzz/                 # Fuzz testing
│   └── transport/        # Transport protocol fuzz tests
└── fixtures/             # Test data and fixtures
    ├── files/            # Test files of various sizes
    └── configs/          # Test configurations
```

## Test Categories

### Unit Tests (`tests/unit/`)

Unit tests focus on testing individual components in isolation:

- **Crypto Tests**: Test encryption/decryption, key management, hashing
- **Storage Tests**: Test file operations, path transforms, CAS functionality
- **Transport Tests**: Test message framing, handshakes, peer connections
- **Peer Tests**: Test peer health, resource management, lifecycle

### Integration Tests (`tests/integration/`)

Integration tests verify that components work together correctly:

- **End-to-End Tests**: Full store/get/replicate workflows
- **Multi-Node Tests**: Multi-node scenarios and network behavior
- **Performance Tests**: Stress testing and performance benchmarks

### Fuzz Tests (`tests/fuzz/`)

Fuzz tests use random data to find edge cases and vulnerabilities:

- **Transport Fuzz Tests**: Test message framing with random data
- **Protocol Fuzz Tests**: Test handshake and message parsing

## Running Tests

### Using Task (Cross-platform)

```bash
# Run all tests
task test

# Run unit tests only
task test-unit

# Run integration tests
task test-integration

# Run tests with race detector
task test-race

# Run fuzz tests
task test-fuzz
```

### Using Scripts

**Windows (PowerShell):**

```powershell
.\scripts\test.ps1 all
.\scripts\test.ps1 unit
.\scripts\test.ps1 integration
```

**Unix-like systems:**

```bash
./scripts/test.sh all
./scripts/test.sh unit
./scripts/test.sh integration
```

### Using Make

```bash
make test
make test-unit
make test-race
make test-fuzz
```

### Direct Go Commands

```bash
# Run all tests
go test -v ./...

# Run specific test package
go test -v ./tests/unit/crypto

# Run with race detector
go test -race -v ./...

# Run fuzz tests
go test -run ^$ -fuzz=Fuzz -fuzztime=30s ./tests/fuzz/transport
```

## Test Guidelines

### Writing Unit Tests

1. **Test Structure**: Use table-driven tests for multiple scenarios
2. **Naming**: Use descriptive test names that explain the scenario
3. **Assertions**: Use clear assertions with helpful error messages
4. **Isolation**: Each test should be independent and not rely on others
5. **Coverage**: Aim for high test coverage, especially for critical paths

### Writing Integration Tests

1. **Setup/Teardown**: Properly clean up resources after tests
2. **Timeouts**: Use appropriate timeouts for network operations
3. **Retries**: Implement retry logic for flaky network operations
4. **Logging**: Use structured logging for debugging test failures
5. **Parallelization**: Run independent tests in parallel when possible

### Writing Fuzz Tests

1. **Seed Corpus**: Provide good seed data for fuzz testing
2. **Validation**: Validate that fuzz-generated data is handled correctly
3. **Crash Detection**: Ensure crashes are properly reported
4. **Performance**: Keep fuzz tests fast to maximize coverage

## Test Data

### Fixtures (`tests/fixtures/`)

- **Files**: Various test files (small, medium, large, binary, text)
- **Configs**: Test configurations for different scenarios
- **Certificates**: Test certificates for TLS testing (if applicable)

### Test Utilities

Common test utilities are available in `tests/utils/`:

- **TestServer**: Helper for starting test servers
- **TestClient**: Helper for creating test clients
- **TestData**: Helper for generating test data
- **TestNetwork**: Helper for creating test networks

## Continuous Integration

Tests are automatically run in CI/CD pipelines:

- **Unit Tests**: Run on every commit
- **Integration Tests**: Run on pull requests
- **Fuzz Tests**: Run periodically and on releases
- **Performance Tests**: Run on releases and major changes

## Coverage Reports

Generate coverage reports:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Generate coverage report for specific package
go test -coverprofile=coverage.out ./internal/crypto
go tool cover -func=coverage.out
```

## Debugging Tests

### Verbose Output

```bash
go test -v ./tests/unit/crypto
```

### Test Logging

Tests use structured logging for debugging:

```bash
# Run with debug logging
LOG_LEVEL=debug go test -v ./tests/integration/end-to-end
```

### Race Detection

```bash
go test -race -v ./...
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./tests/performance

# Memory profiling
go test -memprofile=mem.prof ./tests/performance
```
