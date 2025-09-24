# PeerVault — P2P encrypted file store (Go)

Lightweight peer-to-peer file store demo written in Go. Nodes communicate over TCP, replicate files across peers, and encrypt file transfers using AES-GCM with authentication.

The included entrypoint at `cmd/peervault/main.go` boots 3 nodes locally and runs a simple store/get flow to demonstrate replication.

## MVP Features

PeerVault MVP focuses on core P2P file storage with a single, well-implemented API interface.

### ✅ Core P2P Functionality (MVP)

- **Encrypted file streaming** over TCP using AES-GCM with authentication
- **Content-addressable storage** with SHA-256 based path transform
- **Length-prefixed message framing** for reliable transport (1MB max frame size)
- **Basic peer discovery** and connection management
- **File replication** across connected peers
- **Simple demo** that launches 3 local nodes and exchanges files

### ✅ Primary API Interface (MVP)

- **gRPC API** with grpc-gateway for REST compatibility
- **Protocol Buffers** for type-safe communication
- **Basic file operations**: store, retrieve, list files
- **Health checks** and system status endpoints
- **Authentication** with HMAC-SHA256 and timestamp validation

### ✅ Essential Security (MVP)

- **AES-GCM encryption** with secure nonce management
- **HMAC-SHA256 authentication** for peer connections
- **Key derivation** from cluster key with rotation support
- **Basic vulnerability scanning** with govulncheck and semgrep
- **GitHub Actions security** with explicit permissions and least privilege

### ✅ Basic Observability (MVP)

- **Structured logging** with configurable levels
- **Health check endpoints** for monitoring
- **Basic metrics** for file operations and peer connections

## 🚧 Future Features (Post-MVP)

The following features are planned for future releases and are **not included in the MVP**:

### API Interfaces

- **GraphQL API** - Flexible queries and real-time subscriptions
- **Standalone REST API** - Direct REST without grpc-gateway
- **Interactive documentation** - GraphQL Playground and Swagger UI

### Advanced Security & Compliance

- **Enterprise compliance** - SOC 2, GDPR, ISO 27001 assessments
- **Role-Based Access Control (RBAC)** - Comprehensive authorization
- **Audit logging** - Complete security event logging
- **PKI & Certificate Management** - Public Key Infrastructure
- **Container security** - Trivy scanning and SLSA provenance

### Performance & Optimization

- **Memory optimization** - Buffer pooling, object pooling, connection pooling
- **Multi-level caching** - LRU cache with TTL support
- **Data efficiency** - Compression and content-based deduplication
- **Connection management** - Advanced pooling and multiplexing

### Advanced Features & Ecosystem

- **IPFS compatibility** - Full IPFS protocol support with CID
- **Blockchain integration** - Smart contracts and decentralized identity
- **Machine learning** - AI-based file classification and optimization
- **Edge computing** - Distributed edge node management
- **IoT support** - Device management and sensor data processing

### Developer Experience

- **Plugin architecture** - Extensible system for custom integrations
- **SDK documentation** - Comprehensive guides and examples
- **Advanced observability** - Prometheus metrics, distributed tracing
- **Performance benchmarks** - Comprehensive benchmarking suite

See [ROADMAP.md](documentation/ROADMAP.md) for detailed release planning.

## Content Addressing

PeerVault uses content-addressed storage to ensure data integrity and enable efficient deduplication.

### SHA-256 Based Content Addressing (MVP)

The MVP uses a simple SHA-256 based path transform for content addressing:

#### Path Layout

```bash
Storage Root/
├── ab/                    # First 2 characters of SHA-256 hash
│   └── cdef1234...        # Remaining 62 characters as filename
├── cd/
│   └── ef567890...
└── ...
```

#### Implementation Details

- **Hash Algorithm**: SHA-256 (256-bit hash)
- **Path Depth**: 2-level fan-out (first 2 chars as directory)
- **Filename**: Remaining 62 characters of hex-encoded hash
- **Collision Handling**: SHA-256 provides 2^256 possible values, making collisions cryptographically infeasible

#### Example

```go
// Input: file content "hello world"
// SHA-256: b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
// Path: b9/4d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### IPFS-Compatible Content Addressing (Future)

For future IPFS compatibility, PeerVault includes a more sophisticated content addressing system:

#### CID (Content Identifier) Support

- **Version**: CIDv1 format
- **Codec**: Raw data codec (extensible to other codecs)
- **Hash**: SHA-256 multihash
- **Format**: `bafy...` (base32 encoded) or `Qm...` (base58 encoded)

#### Advanced Path Layout

```bash
IPFS Storage/
├── bafy/                  # CID prefix
│   ├── bei/              # First 3 characters
│   │   └── 2a...         # Remaining characters
│   └── ...
└── ...
```

### Collision Handling

#### SHA-256 Collision Resistance

- **Theoretical Collision Probability**: ~1 in 2^256 (practically impossible)
- **Birthday Paradox**: Even with 2^64 files, collision probability is negligible
- **Cryptographic Security**: SHA-256 is considered cryptographically secure

#### Implementation Safeguards

- **Hash Verification**: Content is verified against its hash on retrieval
- **Integrity Checks**: SHA-256 provides built-in integrity verification
- **Error Handling**: Hash mismatches are detected and reported

### Storage Efficiency

#### Deduplication Benefits

- **Content-Based**: Identical content always maps to the same path
- **Automatic**: No manual deduplication required
- **Space Savings**: Duplicate files are automatically deduplicated

#### Path Distribution

- **Even Distribution**: SHA-256 provides uniform distribution across directories
- **Load Balancing**: 2-level fan-out prevents directory overload
- **Scalability**: System scales to millions of files without performance degradation

## Message Framing

The system uses a robust length-prefixed framing protocol for reliable message transport:

- **Frame Format**: `[type:u8][len:u32][payload:len]`
- **Message Types**:
  - `0x1`: Regular message (with payload)
  - `0x2`: Stream header (no payload)
- **Maximum Frame Size**: 1MB per frame
- **Network Resilience**: Handles partial reads and network interruptions

### Frame Structure

```bash
[Message Frame]
┌─────────┬─────────┬─────────────┐
│ Type    │ Length  │ Payload     │
│ (1 byte)│ (4 bytes)│ (N bytes)  │
└─────────┴─────────┴─────────────┘

[Stream Frame]
┌─────────┬─────────┐
│ Type    │ Length  │
│ (1 byte)│ (4 bytes)│
└─────────┴─────────┘
```

This framing ensures reliable message delivery and eliminates the need for `time.Sleep` coordination.

## Encryption & Security

### AES-GCM Encryption

PeerVault uses **AES-GCM (Galois/Counter Mode)** for authenticated encryption, providing both confidentiality and integrity protection:

- **Algorithm**: AES-256-GCM with 12-byte nonces
- **Nonce Management**: Cryptographically secure random nonces, never reused
- **Authentication**: Built-in authentication prevents tampering
- **Key Size**: 256-bit encryption keys derived from cluster key

### Nonce Management & Security Guarantees

- **Nonce Generation**: Each message uses a fresh, cryptographically secure random 12-byte nonce
- **Nonce Reuse Prevention**: Nonces are never reused - each encryption operation gets a unique nonce
- **Replay Protection**: Timestamp validation in handshake prevents replay attacks
- **Forward Secrecy**: Keys are derived per-session and rotated regularly

## Key Management

The system now supports advanced key management with the following features:

- **Key Derivation**: Encryption keys are derived from a cluster key using HMAC-SHA256
- **Key Rotation**: Keys can be rotated automatically (every 24 hours by default)
- **Environment Configuration**: Set `PEERVAULT_CLUSTER_KEY` environment variable for shared cluster keys

### Using a Shared Cluster Key

For production deployments, set a shared cluster key across all nodes:

```bash
# Set the cluster key (32-byte hex string)
export PEERVAULT_CLUSTER_KEY="a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"

# Run the application
go run ./cmd/peervault
```

If no cluster key is provided, each node generates its own key (suitable for development/testing).

## Authentication

The system now supports authenticated peer connections with the following features:

- **Peer Authentication**: All peer connections are authenticated using HMAC-SHA256 signatures
- **Node Identity**: Each node has a unique ID that's verified during handshake
- **Timestamp Validation**: Handshake messages include timestamps to prevent replay attacks
- **Environment Configuration**: Set `PEERVAULT_AUTH_TOKEN` environment variable for shared authentication

### Using Authentication

For production deployments, set a shared auth token across all nodes:

```bash
# Set the auth token
export PEERVAULT_AUTH_TOKEN="your-secure-auth-token-here"

# Run the application
go run ./cmd/peervault
```

If no auth token is provided, a default demo token is used (suitable for development/testing).

## Requirements

- Go 1.24.4+ (required for security fixes)
- Make (optional; for Unix-like systems)

## Security

### Security Features

#### ✅ **Implemented (MVP)**

- **AES-GCM Encryption**: Authenticated encryption with secure nonce management
- **HMAC-SHA256 Authentication**: Peer connection authentication with timestamp validation
- **Key Derivation**: HMAC-SHA256 based key derivation from cluster key
- **Basic Vulnerability Scanning**: govulncheck and semgrep integration
- **Secure Nonce Management**: Cryptographically secure random nonces, never reused

#### 🚧 **Planned (Post-MVP)**

- **Advanced Vulnerability Scanning**: semgrep, detect-secrets, Trivy integration
- **Compliance Guidance**: Example controls for SOC 2, GDPR, ISO 27001 (guidance only, not certification)
- **Role-Based Access Control (RBAC)**: Comprehensive authorization system
- **Audit Logging**: Comprehensive security event logging
- **PKI & Certificate Management**: Public Key Infrastructure
- **Container Security**: Trivy scanning and SLSA provenance
- **Security Policies**: Access control and data classification policies

**Important**: Compliance features provide guidance and example controls only. They do not provide actual certification or guarantee compliance with regulatory requirements.

### Security Tools & Scripts

#### Local Security Validation

```bash
# Run comprehensive security checks
./scripts/security-check.sh

# Run specific security checks
./scripts/security-check.sh --vulnerability
./scripts/security-check.sh --compliance
./scripts/security-check.sh --test

# Install security tools
./scripts/security-check.sh --install-tools
```

#### Windows PowerShell Support

```powershell
# Run all security checks
.\scripts\security-check.ps1

# Run specific checks
.\scripts\security-check.ps1 -Vulnerability
.\scripts\security-check.ps1 -Compliance
.\scripts\security-check.ps1 -Test

# Install security tools
.\scripts\security-check.ps1 -InstallTools
```

### CI/CD Security Pipeline

The project includes comprehensive CI/CD security integration:

#### **Security Pipeline** (`.github/workflows/security.yml`)

- **Vulnerability Scanning**: govulncheck, gosec, semgrep, detect-secrets
- **Compliance Checking**: SOC 2, GDPR, ISO 27001 assessments
- **Container Security**: Trivy vulnerability scanning
- **Security Integration Tests**: RBAC, audit, privacy, PKI testing
- **Daily Security Scans**: Automated daily security assessments

#### **Development Security Checks** (`.github/workflows/security-dev.yml`)

- **Quick Security Validation**: For development changes
- **Security Module Testing**: Compilation and functionality testing
- **Policy Validation**: Security policy syntax checking
- **Documentation Checks**: Security documentation completeness

### Vulnerability Fixes

This project includes fixes for the following security vulnerabilities:

#### GO-2025-3750: Inconsistent handling of O_CREATE|O_EXCL on Unix and Windows

**Status:** ✅ Fixed  
**Go Version:** 1.24.4+  
**Impact:** Race conditions in file creation, potential security bypasses

**Fix Applied:**

- Updated to Go 1.24.4+ (minimum required version)
- Implemented atomic file creation with `O_CREATE|O_EXCL` flags
- Added proper error handling for file existence checks

**Files Modified:**

- `internal/storage/store.go` - Added `createFileAtomic()` function
- `go.mod` - Updated to Go 1.24.4
- `.github/workflows/ci.yml` - Updated CI to use Go 1.24.4

### Security Scanning

The CI pipeline includes automated security scanning:

- **govulncheck** - Scans for known vulnerabilities in dependencies
- **semgrep** - Static analysis for security issues
- **detect-secrets** - Secrets and credentials detection
- **Trivy** - Container vulnerability scanning
- **Dependabot** - Automated dependency updates with security patches

## Windows Defender Setup

If you're developing on Windows, you may encounter Windows Defender popups when running the application or tests. This is because Go applications that create network connections and access the file system are often flagged as potentially suspicious.

**⚠️ Security Note**: The trust script adds Windows Defender exclusions for the project folder. This is necessary for development but should only be used in trusted environments. Never run this on production systems or with untrusted code.

### Quick Fix (Recommended)

Run the quick trust script as Administrator:

```cmd
# Right-click and "Run as Administrator"
scripts\quick_trust.bat
```

### Full Setup (Advanced)

For a complete setup with code signing, run the PowerShell script as Administrator:

```powershell
# Right-click PowerShell and "Run as Administrator"
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\scripts\trust_application.ps1
```

This will:

- Add Windows Defender exclusions for the project folder
- Create a self-signed certificate for code signing
- Set up a build script that signs your binaries
- Configure your PowerShell profile for development

### Manual Setup

If you prefer to do it manually:

1. Open **Windows Security**
2. Go to **Virus & threat protection**
3. Click **Manage settings**
4. Under **Exclusions**, add:
   - Folder: Your project directory
   - Process: `peervault.exe`, `peervault-node.exe`, `peervault-demo.exe`, `go.exe`

### Why This Happens

Windows Defender flags Go applications because they:

- Create network connections (P2P networking)
- Access the file system (storage operations)
- Generate random data (crypto operations)
- Are compiled from source (not signed by trusted publishers)

This is normal for development and doesn't indicate any security issues.

## Install

```bash
# Clone your copy of this repository and change into it
git clone https://github.com/Skpow1234/PeerVault
cd peervault
go mod download
```

## Build

### Linux/macOS

```bash
make build
# binary at ./bin/peervault

# Build GraphQL server
go build -o bin/peervault-graphql ./cmd/peervault-graphql

# Build REST API server
go build -o bin/peervault-api ./cmd/peervault-api

# Build Mock server
go build -o bin/peervault-mock ./cmd/peervault-mock
```

### Windows (PowerShell)

```powershell
go build -o bin\peervault.exe .\cmd\peervault

# Build GraphQL server
go build -o bin\peervault-graphql.exe .\cmd\peervault-graphql

# Build Mock server
go build -o bin\peervault-mock.exe .\cmd\peervault-mock
```

## Run

This repository’s `main.go` starts 3 nodes on localhost: `:3000`, `:7000`, `:5000`, then stores and fetches sample files via the third node.

### Easiest: go run

```bash
go run ./cmd/peervault
```

### Linux/macOS with Make

```bash
make run
```

### Windows

- If you built with `go build -o bin\peervault.exe .\cmd\peervault`:

```powershell
.\bin\peervault.exe
```

### What you should see

- Logs for each node starting up and connecting
- 20 demo files being stored and fetched
- Output lines like:

  - `[::]:5000 starting fileserver...`
  - `received and written (...) bytes to disk`
  - `my big data file here!`

#### Important note for Windows users

By default, each node’s storage root is set to the listen address plus `_network` (for example `":3000_network"`). The colon `:` is not a valid character in Windows directory names, which can cause errors when creating folders.

Two simple options:

- Recommended: Run via WSL or Git Bash (Unix-like environment), or
- Update `main.go` to use a Windows-friendly storage root. For example, change the `StorageRoot` in `makeServer` to something like:

```go
StorageRoot: fmt.Sprintf("node%s_network", strings.TrimPrefix(listenAddr, ":")),
```

or even hardcode per node (e.g., `"node3000_network"`, `"node7000_network"`, `"node5000_network"`).

File to edit: `cmd/peervault/main.go`, function `makeServer`.

## GraphQL API

PeerVault includes a comprehensive GraphQL API for interacting with the distributed storage system.

### Running the GraphQL Server

```bash
# Build and run the GraphQL server
go build -o peervault-graphql.exe ./cmd/peervault-graphql
./peervault-graphql.exe

# Or run directly
go run ./cmd/peervault-graphql
```

### GraphQL Endpoints

- **GraphQL API**: `http://localhost:8080/graphql`
- **GraphQL Playground**: `http://localhost:8080/playground`
- **Health Check**: `http://localhost:8080/health`
- **Metrics**: `http://localhost:8080/metrics`

### Example Queries

```bash
# Health check
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ health { status timestamp } }"}'

# Get system metrics
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ systemMetrics { storage { totalSpace usedSpace } } }"}'
```

For complete GraphQL API documentation, see [docs/graphql/README.md](docs/graphql/README.md).

## REST API

PeerVault includes a REST API for simple operations and webhook integrations, built with a clean architecture using entities, DTOs, mappers, endpoints, services, and service implementations.

### Architecture

The REST API follows a clean, layered architecture pattern with consolidated types:

```bash
internal/api/rest/
├── types/            # Consolidated types, entities, DTOs, and mappers
│   ├── entities.go   # Core business entities
│   ├── requests/     # API request DTOs
│   │   ├── file_requests.go
│   │   ├── peer_requests.go
│   │   └── system_requests.go
│   ├── responses/    # API response DTOs
│   │   ├── file_responses.go
│   │   ├── peer_responses.go
│   │   └── system_responses.go
│   └── mappers.go    # Entity-DTO mapping functions
├── endpoints/        # HTTP endpoint handlers
│   ├── file_endpoints.go
│   ├── peer_endpoints.go
│   └── system_endpoints.go
├── services/         # Business logic interfaces
│   ├── file_service.go
│   ├── peer_service.go
│   └── system_service.go
├── implementations/  # Service implementations
│   ├── file_service_impl.go
│   ├── peer_service_impl.go
│   └── system_service_impl.go
└── server.go        # Main server configuration
```

### Running the REST API Server

```bash
# Build and run the REST API server
go build -o peervault-api.exe ./cmd/peervault-api
./peervault-api.exe

# Or run directly
go run ./cmd/peervault-api
```

### REST API Endpoints

- **REST API**: `http://localhost:8081/api/v1`
- **Swagger UI**: `http://localhost:8081/docs`
- **Health Check**: `http://localhost:8081/health`
- **Metrics**: `http://localhost:8081/metrics`

### Example Requests

```bash
# Health check
curl http://localhost:8081/health

# List files
curl http://localhost:8081/api/v1/files

# Upload a file
curl -X POST http://localhost:8081/api/v1/files \
  -F "file=@example.txt" \
  -F "metadata={\"owner\":\"user1\"}"

# Add a peer
curl -X POST http://localhost:8081/api/v1/peers \
  -H "Content-Type: application/json" \
  -d '{"address": "192.168.1.100", "port": 8080}'
```

### Architecture Benefits

- **Consolidated Types**: All types, entities, DTOs, and mappers in one organized package
- **Separation of Concerns**: Clear separation between types, endpoints, services, and implementations
- **Testability**: Each layer can be tested independently
- **Maintainability**: Easy to modify or extend individual components
- **Scalability**: Services can be easily replaced or enhanced
- **Type Safety**: Strong typing throughout the application
- **Reduced Complexity**: Simplified import structure and better organization

For complete REST API documentation, see [docs/api/README.md](docs/api/README.md).

### Swagger Documentation

The REST API includes comprehensive OpenAPI/Swagger documentation:

- **Interactive Swagger UI**: `http://localhost:8081/docs` - Explore and test endpoints directly in your browser
- **OpenAPI Specification**: `http://localhost:8081/swagger.json` - Machine-readable API specification
- **Complete Documentation**: [docs/api/peervault-rest-api.yaml](docs/api/peervault-rest-api.yaml) - Full OpenAPI 3.0 specification

## gRPC API

PeerVault includes a high-performance gRPC API for streaming operations and service-to-service communication, built with Protocol Buffers and designed for high-throughput applications.

### Architecture grpc

The gRPC API follows a service-oriented architecture:

```bash
internal/api/grpc/
├── types/            # Type definitions (temporary, will be replaced by protobuf)
│   └── types.go      # Basic type definitions
├── services/         # Service implementations
│   ├── file_service.go
│   ├── peer_service.go
│   └── system_service.go
├── server.go         # Main gRPC server
└── proto/            # Protocol Buffer definitions
    └── peervault.proto
```

### Running the gRPC API Server

```bash
# Build and run the gRPC API server
go build -o peervault-grpc.exe ./cmd/peervault-grpc
./peervault-grpc.exe

# Or run directly
go run ./cmd/peervault-grpc

# Run with custom configuration
go run ./cmd/peervault-grpc -port 8082 -auth-token your-secure-token
```

### gRPC API Features

- ✅ **HTTP/JSON API**: Working server with JSON endpoints (temporary solution)
- ✅ **Health checks**: `/health` endpoint with system status
- ✅ **System information**: `/system/info` endpoint with version and metrics
- ✅ **Metrics**: `/system/metrics` endpoint with performance data
- ✅ **File operations**: `/files` endpoints for file management
- ✅ **Peer management**: `/peers` endpoints for peer operations
- 🔄 **Bidirectional Streaming**: Real-time file upload/download with chunked transfer (planned)
- 🔄 **Service Discovery**: Built-in service discovery and load balancing support (planned)
- 🔄 **High Throughput**: Optimized for high-performance applications (planned)
- 🔄 **Type Safety**: Strongly typed with protobuf definitions (planned)
- 🔄 **Authentication**: Token-based authentication with metadata (planned)
- 🔄 **Event Streaming**: Real-time events for file operations, peer status, and system metrics (planned)

### gRPC Services

- **File Operations**: Streaming upload/download, metadata management
- **Peer Operations**: Peer discovery, health monitoring, network management
- **System Operations**: Metrics collection, health checks, system information
- **Event Streaming**: Real-time events for monitoring and notifications

### Example Usage

```go
// Connect to gRPC server
conn, err := grpc.Dial("localhost:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewPeerVaultServiceClient(conn)

// Health check
response, err := client.HealthCheck(context.Background(), &pb.Empty{})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\n", response.Status)
```

### gRPC API Benefits

- **High Performance**: Optimized for low-latency, high-throughput operations
- **Streaming Support**: Efficient handling of large file transfers
- **Type Safety**: Strongly typed with Protocol Buffers
- **Multi-language Support**: Client libraries available for multiple languages
- **Service Discovery**: Built-in service discovery capabilities
- **Real-time Events**: Streaming events for monitoring and notifications

For complete gRPC API documentation, see [docs/grpc/README.md](docs/grpc/README.md).

## API Testing & Quality Assurance

PeerVault includes comprehensive API testing capabilities for ensuring reliability, performance, and security across all API interfaces.

### 🧪 **API Testing Features**

#### ✅ **Interactive API Testing**
- **Postman Collections**: Pre-configured collections for all API endpoints
- **Environment Management**: Support for multiple environments (dev, staging, prod)
- **Automated Testing**: Newman CLI integration for CI/CD pipelines
- **Test Reporting**: Comprehensive test results and coverage reports

#### ✅ **API Mocking**
- **Mock Server**: Standalone mock server for development and testing
- **OpenAPI Integration**: Automatic mock generation from OpenAPI specifications
- **Scenario Testing**: Customizable response scenarios and conditions
- **Analytics**: Request/response monitoring and analytics

#### ✅ **Contract Testing**
- **Consumer-Driven Contracts**: Pact-compatible contract testing
- **Request/Response Validation**: Schema validation and compatibility checks
- **Contract Evolution**: Track API changes and breaking changes
- **Provider Verification**: Automated provider contract verification

#### ✅ **Performance Testing**
- **Load Testing**: Configurable concurrency and duration testing
- **Stress Testing**: System limits and bottleneck identification
- **Response Time Analysis**: Detailed performance metrics and distributions
- **k6 Integration**: Advanced performance testing scenarios

#### ✅ **Security Testing**
- **OWASP Top 10**: Comprehensive API security vulnerability testing
- **Injection Testing**: SQL injection, XSS, and command injection tests
- **Authentication Testing**: Auth bypass and token validation tests
- **Security Headers**: Security header validation and compliance

### 🚀 **Quick Start**

#### Run All API Tests
```bash
# Run comprehensive API test suite
./scripts/api-testing/run-tests.sh

# Run with custom configuration
./scripts/api-testing/run-tests.sh --base-url http://localhost:8080 --verbose
```

#### Start Mock Server
```bash
# Start mock server for development
go run cmd/peervault-mock/main.go --config config/mock-server.yaml

# Generate scenarios from OpenAPI spec
go run cmd/peervault-mock/main.go --generate --spec docs/api/peervault-rest-api.yaml
```

#### Run Individual Test Suites
```bash
# Contract tests
go test ./tests/contracts/...

# Performance tests
go test -bench=. ./tests/performance/...

# Security tests
go test ./tests/security/...

# Postman tests
newman run tests/api/collections/peervault-postman.json
```

### 📊 **Test Coverage**

The API testing framework provides comprehensive coverage:

- **Contract Tests**: 19 test cases covering all major API endpoints
- **Performance Tests**: 8 test scenarios with configurable load patterns
- **Security Tests**: 6 OWASP Top 10 security test categories
- **Mock Scenarios**: 15+ pre-configured mock response scenarios

### 🔧 **Configuration**

#### Mock Server Configuration
```yaml
# config/mock-server.yaml
port: 3001
host: "localhost"
openapi_spec: "docs/api/peervault-rest-api.yaml"
mock_data_dir: "tests/api/mock-data"
response_delay: 100ms
enable_analytics: true
```

#### Test Environment Variables
```bash
# API Testing Configuration
export BASE_URL="http://localhost:3000"      # Target API URL
export MOCK_URL="http://localhost:3001"      # Mock server URL
export TEST_TIMEOUT="30s"                    # Test timeout duration
export VERBOSE="true"                        # Enable verbose output
```

### 📈 **CI/CD Integration**

The API testing framework integrates seamlessly with CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run API Tests
  run: |
    ./scripts/api-testing/run-tests.sh --base-url ${{ env.API_URL }}
```

### 📚 **Documentation**

- **API Testing Guide**: [docs/api/testing/README.md](docs/api/testing/README.md)
- **Implementation Details**: [docs/api/testing/IMPLEMENTATION.md](docs/api/testing/IMPLEMENTATION.md)
- **Security Configuration**: [.github/workflows/SECURITY.md](.github/workflows/SECURITY.md)

For complete API testing documentation, see [docs/api/testing/README.md](docs/api/testing/README.md).

## Plugin Architecture

PeerVault includes a comprehensive plugin architecture for extensibility:

### Plugin Types

- **Storage Plugins**: Custom storage backends (S3, Azure, GCP, etc.)
- **Authentication Plugins**: Custom authentication mechanisms
- **Transport Plugins**: Custom transport protocols
- **Processing Plugins**: Data processing and transformation

### Example: S3 Storage Plugin

```go
// Example S3 storage plugin
package s3_storage

import (
    "github.com/Skpow1234/Peervault/internal/plugins"
    "github.com/aws/aws-sdk-go/service/s3"
)

type S3StoragePlugin struct {
    bucket   string
    s3Client *s3.S3
}

func (p *S3StoragePlugin) Store(key string, data []byte) error {
    // S3 storage implementation
    return nil
}

func (p *S3StoragePlugin) Retrieve(key string) ([]byte, error) {
    // S3 retrieval implementation
    return nil, nil
}
```

### Plugin Development

1. **Create Plugin Module**: Create a new Go module for your plugin
2. **Implement Interfaces**: Implement the required plugin interfaces
3. **Register Plugin**: Register your plugin with the PeerVault system
4. **Configure Plugin**: Add plugin configuration to your setup

For complete plugin development guide, see [internal/plugins/README.md](internal/plugins/README.md).

## Developer Tools & SDKs

### SDK Documentation

- **Go SDK**: [docs/sdk/go/README.md](docs/sdk/go/README.md)
- **JavaScript/TypeScript SDK**: [docs/sdk/javascript/README.md](docs/sdk/javascript/README.md)
- **Developer Portal**: [docs/portal/README.md](docs/portal/README.md)

### Code Examples

- **REST API Examples**: [docs/examples/rest/](docs/examples/rest/)
- **GraphQL Examples**: [docs/examples/graphql/](docs/examples/graphql/)
- **JavaScript Browser Examples**: [docs/examples/javascript/browser/](docs/examples/javascript/browser/)

### Interactive Documentation

- **Swagger UI**: `http://localhost:8081/docs` - Interactive REST API documentation
- **GraphQL Playground**: `http://localhost:8080/playground` - Interactive GraphQL testing
- **API Reference**: Complete API documentation with examples

### Getting Started Guide

For a comprehensive getting started guide, see [docs/portal/guides/getting-started.md](docs/portal/guides/getting-started.md).

## Project Status & Roadmap

### 🎯 **Current Status: MVP Development**

**PeerVault is currently in MVP development phase** focusing on core P2P file storage functionality:

- **✅ Core P2P**: Basic peer-to-peer file storage with encryption
- **✅ gRPC API**: Primary API interface with grpc-gateway for REST
- **✅ Security**: AES-GCM encryption and HMAC-SHA256 authentication
- **✅ Storage**: Content-addressable storage with SHA-256
- **🔄 In Progress**: Message framing, peer discovery, basic replication

### 📋 **MVP Goals (Current Focus)**

1. **Stable P2P Core**: Reliable file storage and replication between peers
2. **Single API Interface**: gRPC with grpc-gateway (no GraphQL/REST complexity)
3. **Essential Security**: Proper encryption and authentication
4. **Basic Observability**: Health checks and structured logging
5. **Simple Demo**: Working 3-node example with file exchange

### 🚧 **Post-MVP Roadmap**

After MVP completion, features will be added incrementally:

#### **Phase 1: API Expansion**

- GraphQL API for flexible queries
- Standalone REST API
- Interactive documentation (Swagger UI, GraphQL Playground)

#### **Phase 2: Advanced Security**

- Enterprise compliance (SOC 2, GDPR guidance)
- Role-Based Access Control (RBAC)
- PKI and certificate management
- Advanced audit logging

#### **Phase 3: Performance & Scale**

- Memory optimization and caching
- Data compression and deduplication
- Advanced connection management
- Performance benchmarking

#### **Phase 4: Ecosystem Integration**

- IPFS compatibility
- Blockchain integration
- Machine learning features
- Edge computing and IoT support

### 📊 **Development Approach**

- **MVP First**: Focus on core functionality before adding complexity
- **Incremental**: Add features one at a time with proper testing
- **Quality**: Each feature must be production-ready before moving to next
- **Documentation**: Clear status indicators for what exists vs. planned

For detailed roadmap information, see [documentation/ROADMAP.md](documentation/ROADMAP.md).

## gRPC Implementation Status

### ✅ **Completed**

- **HTTP/JSON Server**: Working server with JSON endpoints
- **Health Endpoints**: `/health`, `/system/info`, `/system/metrics`
- **File Operations**: `/files` endpoints for file management
- **Peer Operations**: `/peers` endpoints for peer management
- **Event Broadcasting**: Background event generation for testing
- **Service Layer**: Complete service implementations for file, peer, and system operations

### 🔄 **In Progress**

- **Full protobuf generation**: The protobuf Go code generation is currently using a manual implementation. The next step is to properly install and configure `protoc` to generate the complete Go code from the `.proto` files.
- **True gRPC implementation**: Currently using HTTP/JSON endpoints as a temporary solution. Will be replaced with proper gRPC streaming once protobuf generation is working.

### 📋 **Next Steps**

1. **Install protoc compiler**: Set up proper protobuf compilation environment
2. **Generate protobuf code**: Use `protoc` to generate proper Go code from `.proto` files
3. **Implement true gRPC**: Replace HTTP/JSON with actual gRPC streaming
4. **Add streaming endpoints**: Implement bidirectional streaming for file operations
5. **Add authentication**: Implement proper gRPC authentication and authorization

### 🚨 **Current Limitations**

- The server uses HTTP/JSON instead of true gRPC due to protobuf marshaling issues
- Streaming functionality is simulated with background event generation
- Authentication is currently disabled for development purposes

The Swagger documentation includes:

- Detailed endpoint descriptions with examples
- Request/response schemas for all data types
- Authentication and security information
- Error handling documentation
- Rate limiting details

## Docker

PeerVault supports multiple containerization approaches:

### All-in-One Container (Development)

Build and run the demo in a single container:

```bash
docker build -t peervault .
docker run --rm -p 3000:3000 -p 5000:5000 -p 7000:7000 peervault
```

### Multi-Container Deployment (Production)

For production-like environments with separate containers:

```bash
# Build and run all services
docker-compose -f docker/docker-compose.yml up --build

# Run in background
docker-compose -f docker/docker-compose.yml up -d --build

# View logs
docker-compose -f docker/docker-compose.yml logs -f

# Stop all services
docker-compose -f docker/docker-compose.yml down
```

### Development Setup

For development and testing:

```bash
docker-compose -f docker/docker-compose.dev.yml up --build
```

### Individual Nodes

Run individual nodes for custom topologies:

```bash
# Build node image
docker build -f docker/Dockerfile.node -t peervault-node .

# Run bootstrap node
docker run -d --name node1 -p 3000:3000 peervault-node --listen :3000

# Run client node
docker run -d --name node2 -p 5000:5000 peervault-node \
  --listen :5000 --bootstrap node1:3000
```

For detailed containerization documentation, see [documentation/CONTAINERIZATION.md](documentation/CONTAINERIZATION.md).

## How it works (high level)

- `cmd/peervault/main.go` creates 3 servers and bootstraps them together using the TCP transport in `internal/transport/p2p`.
- Files are written to disk under a content-addressed path derived from a SHA-1 of the key (`CASPathTransformFunc` in `internal/storage`).
- On store:
  - The file is written locally.
  - A control message is broadcast to peers so they can pull the encrypted stream and persist it.
- On get:
  - If not present locally, a request is broadcast and another peer streams the file back.
- Network messages are framed by a minimal protocol in `internal/transport/p2p` with small control bytes to distinguish messages vs. streams.
- **GraphQL API**: The `cmd/peervault-graphql` binary provides a GraphQL interface for querying files, monitoring peers, and accessing system metrics through HTTP endpoints.

## Project Structure

```bash
peervault/
├── cmd/                          # Application entrypoints
│   ├── peervault/               # Main application binary
│   ├── peervault-node/          # Standalone node binary
│   ├── peervault-demo/          # Demo client binary
│   ├── peervault-graphql/       # GraphQL API server binary
│   ├── peervault-api/           # REST API server binary
│   ├── peervault-grpc/          # gRPC API server binary
│   ├── peervault-mock/          # API mock server binary
│   ├── peervault-config/        # Configuration management tool
│   ├── peervault-ipfs/          # IPFS compatibility tool
│   ├── peervault-chain/         # Blockchain integration tool
│   ├── peervault-ml/            # Machine learning tool
│   └── peervault-edge/          # Edge computing tool
├── internal/                     # Core application code
│   ├── api/                     # API interfaces
│   │   ├── graphql/             # GraphQL API implementation
│   │   │   ├── schema/          # GraphQL schema definitions
│   │   │   ├── types/           # Go types for GraphQL
│   │   │   ├── resolvers/       # GraphQL resolvers
│   │   │   └── server.go        # GraphQL HTTP server
│   │   ├── rest/                # REST API implementation
│   │   │   ├── endpoints/       # HTTP endpoint handlers
│   │   │   ├── implementations/ # Service implementations
│   │   │   ├── services/        # Business logic interfaces
│   │   │   ├── types/           # API types and DTOs
│   │   │   └── server.go        # REST HTTP server
│   │   ├── grpc/                # gRPC API implementation
│   │   │   ├── services/        # gRPC service implementations
│   │   │   ├── types/           # gRPC type definitions
│   │   │   └── server.go        # gRPC server
│   │   ├── mocking/             # API mocking framework
│   │   │   └── mock_server.go   # Mock server implementation
│   │   ├── contracts/           # Contract testing framework
│   │   │   ├── contract.go      # Contract verification logic
│   │   │   └── contract_test.go # Contract test implementation
│   │   ├── performance/         # Performance testing framework
│   │   │   ├── performance.go   # Performance testing logic
│   │   │   └── load_test.go     # Load test implementation
│   │   └── security/            # Security testing framework
│   │       ├── security.go      # Security testing logic
│   │       └── security_test.go # Security test implementation
│   ├── blockchain/              # Blockchain integration
│   │   └── integration.go       # Blockchain integration and smart contracts
│   ├── content/                 # Content addressing
│   │   └── addressing.go        # Content addressing and CID support
│   ├── edge/                    # Edge computing
│   │   └── computing.go         # Edge computing and task distribution
│   ├── ipfs/                    # IPFS compatibility
│   │   └── compatibility.go     # IPFS protocol compatibility
│   ├── iot/                     # IoT device management
│   │   └── devices.go           # IoT devices and sensor data
│   ├── ml/                      # Machine learning
│   │   └── classification.go    # ML classification and optimization
│   ├── app/                     # Application logic
│   │   ├── fileserver/          # Core file server implementation
│   │   └── service.go           # Main application service
│   ├── auth/                    # Authentication and authorization
│   │   └── rbac.go              # Role-Based Access Control
│   ├── audit/                   # Audit logging and monitoring
│   │   └── audit.go             # Comprehensive audit logging
│   ├── backup/                  # Backup and disaster recovery
│   │   └── backup.go            # Backup management system
│   ├── cache/                   # Caching system
│   │   ├── cache.go             # LRU cache implementation
│   │   └── multi_level.go       # Multi-level caching
│   ├── compression/             # Data compression
│   │   └── compression.go       # Compression utilities
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Configuration structures
│   │   ├── validation.go        # Configuration validation
│   │   └── watcher.go           # Configuration file watching
│   ├── crypto/                  # Cryptographic functions and key management
│   │   └── crypto.go            # Encryption and key management
│   ├── deduplication/           # Data deduplication
│   │   └── deduplication.go     # Content-based deduplication
│   ├── domain/                  # Domain entities and business logic
│   │   └── entity.go            # Core domain entities
│   ├── dto/                     # Data transfer objects for network communication
│   │   └── messages.go          # Network message DTOs
│   ├── health/                  # Health checking
│   │   └── health.go            # Health check system
│   ├── logging/                 # Logging utilities and configuration
│   │   └── logger.go            # Structured logging
│   ├── mapper/                  # Data mapping between domain and DTOs
│   │   └── message_mapper.go    # Message mapping utilities
│   ├── metrics/                 # Metrics collection
│   │   └── metrics.go           # Prometheus-compatible metrics
│   ├── peer/                    # Peer management and health monitoring
│   │   ├── health.go            # Peer health monitoring
│   │   └── resource_manager.go  # Resource management
│   ├── pki/                     # PKI and certificate management
│   │   └── pki.go               # Public Key Infrastructure
│   ├── plugins/                 # Plugin architecture
│   │   ├── plugin.go            # Plugin interfaces and management
│   │   └── README.md            # Plugin development guide
│   ├── pool/                    # Object and connection pooling
│   │   ├── buffer_pool.go       # Buffer pooling
│   │   ├── connection_pool.go   # Connection pooling
│   │   └── object_pool.go       # Object pooling
│   ├── privacy/                 # Data privacy and compliance
│   │   └── privacy.go           # Privacy controls and compliance
│   ├── storage/                 # Content-addressable storage implementation
│   │   ├── store.go             # Storage implementation
│   │   ├── path_utils.go        # Path utilities
│   │   └── ggnetwork/           # Storage data directory
│   ├── tracing/                 # Distributed tracing
│   │   └── tracing.go           # Tracing implementation
│   └── transport/               # Network transport layer
│       └── p2p/                 # P2P transport implementation
│           ├── encoding.go      # Message encoding
│           ├── handshake.go     # Connection handshake
│           ├── tcp_transport.go # TCP transport implementation
│           └── transport.go     # Transport interface
├── plugins/                     # Plugin implementations
│   └── s3-storage/              # S3 storage plugin
│       ├── s3_storage.go        # S3 plugin implementation
│       └── go.mod               # Plugin module definition
├── security/                    # Security infrastructure
│   ├── audit/                   # Security audit tools
│   │   ├── scanner.go           # Vulnerability scanner
│   │   └── compliance.go        # Compliance checker
│   ├── policies/                # Security policies
│   │   ├── access-control.yaml  # Access control policies
│   │   └── data-classification.yaml # Data classification policies
│   ├── tools/                   # Security tools
│   │   ├── security-scan.sh     # Security scanning script
│   │   ├── penetration-test.sh  # Penetration testing script
│   │   └── compliance-check.sh  # Compliance checking script
│   └── README.md                # Security documentation
├── tests/                       # Comprehensive test suite
│   ├── unit/                    # Unit tests for all components
│   │   ├── concurrency/         # Concurrency safety tests
│   │   ├── config/              # Configuration tests
│   │   ├── crypto/              # Cryptographic function tests
│   │   ├── logging/             # Logging system tests
│   │   ├── peer/                # Peer management tests
│   │   ├── storage/             # Storage layer tests
│   │   └── transport/           # Transport layer tests
│   ├── integration/             # Integration tests
│   │   ├── config/              # Configuration integration tests
│   │   ├── end-to-end/          # End-to-end workflow tests
│   │   ├── graphql/             # GraphQL API integration tests
│   │   ├── grpc/                # gRPC API integration tests
│   │   ├── milestone9/          # Milestone 9 advanced features tests
│   │   ├── multi-node/          # Multi-node network tests
│   │   ├── performance/         # Performance and benchmark tests
│   │   └── rest/                # REST API integration tests
│   ├── api/                     # API testing framework
│   │   ├── collections/         # Postman/Insomnia collections
│   │   └── mock-data/           # Mock server data and scenarios
│   ├── contracts/               # Contract testing definitions
│   │   ├── health_contract.json # Health check contract
│   │   └── contract_test.go     # Contract verification tests
│   ├── performance/             # Performance testing
│   │   ├── load_test.go         # Go performance tests
│   │   └── load-test.js         # k6 performance tests
│   ├── security/                # Security testing
│   │   └── security_test.go     # Security test implementation
│   ├── fuzz/                    # Fuzz testing for robustness
│   │   ├── crypto/              # Crypto layer fuzz tests
│   │   ├── storage/             # Storage layer fuzz tests
│   │   └── transport/           # Transport layer fuzz tests
│   ├── utils/                   # Test utilities and helpers
│   │   └── test_server.go       # Test server utilities
│   └── fixtures/                # Test data and fixtures
│       ├── configs/             # Test configuration files
│       └── files/               # Test files
├── documentation/               # Project documentation
│   ├── README.md               # Documentation index
│   ├── CONTRIBUTING.md         # Contribution guidelines
│   ├── SECURITY.md             # Security policy
│   ├── ROADMAP.md              # Project roadmap
│   ├── ENCRYPTION.md           # Encryption implementation details
│   ├── LOGGING.md              # Logging system documentation
│   ├── CONTAINERIZATION.md     # Docker and deployment guide
│   ├── CONFIGURATION.md        # Configuration management guide
│   ├── PROJECT_MATURITY_ROADMAP.md # Project maturity roadmap
│   └── PIPELINE_UPDATES.md     # CI/CD pipeline updates
├── docs/                        # API and feature documentation
│   ├── api/                     # REST API documentation
│   │   ├── peervault-rest-api.yaml # OpenAPI specification
│   │   └── README.md            # REST API guide
│   ├── graphql/                 # GraphQL API documentation
│   │   ├── README.md            # GraphQL API guide
│   │   └── schema.graphql       # GraphQL schema definition
│   ├── grpc/                    # gRPC API documentation
│   │   └── README.md            # gRPC API guide
│   ├── sdk/                     # SDK documentation
│   │   ├── go/                  # Go SDK documentation
│   │   └── javascript/          # JavaScript SDK documentation
│   ├── examples/                # Code examples
│   │   ├── rest/                # REST API examples
│   │   ├── graphql/             # GraphQL examples
│   │   └── javascript/          # JavaScript examples
│   ├── portal/                  # Developer portal
│   │   ├── guides/              # Getting started guides
│   │   └── README.md            # Developer portal
│   ├── milestone9/              # Milestone 9 documentation
│   │   └── README.md            # Advanced features documentation
│   ├── swagger/                 # Swagger UI
│   │   └── index.html           # Interactive API documentation
│   └── graphql-playground/      # GraphQL Playground
│       └── index.html           # Interactive GraphQL testing
├── scripts/                     # Build and automation scripts
│   ├── build.sh                # Unix build script
│   ├── build.ps1               # Windows build script
│   ├── run.sh                  # Unix run script
│   ├── run.ps1                 # Windows run script
│   ├── test.sh                 # Unix test script
│   ├── test.ps1                # Windows test script
│   ├── pre-commit.sh           # Pre-commit hooks (Unix)
│   ├── pre-commit.ps1          # Pre-commit hooks (Windows)
│   ├── security-check.sh       # Security validation script (Unix)
│   ├── security-check.ps1      # Security validation script (Windows)
│   ├── quick_trust.bat         # Windows Defender trust script
│   ├── trust_application.ps1   # Windows Defender setup script
│   └── api-testing/            # API testing scripts
│       └── run-tests.sh        # Comprehensive API test runner
├── docker/                      # Containerization files
│   ├── Dockerfile              # Main application container
│   ├── Dockerfile.node         # Node-specific container
│   ├── Dockerfile.demo         # Demo client container
│   ├── Dockerfile.graphql-api  # GraphQL API container
│   ├── Dockerfile.grpc-api     # gRPC API container
│   ├── Dockerfile.rest-api     # REST API container
│   ├── docker-compose.yml      # Production multi-container setup
│   ├── docker-compose.dev.yml  # Development container setup
│   ├── docker-compose.apis.yml # API services setup
│   └── README.md               # Docker documentation
├── .github/                     # GitHub configuration
│   ├── workflows/              # CI/CD pipeline configuration
│   │   ├── ci.yml              # Main CI pipeline
│   │   ├── security.yml        # Security pipeline
│   │   ├── security-dev.yml    # Development security checks
│   │   └── README.md           # Pipeline documentation
│   ├── ISSUE_TEMPLATE/         # Issue and PR templates
│   │   ├── bug_report.md       # Bug report template
│   │   └── feature_request.md  # Feature request template
│   ├── pull_request_template.md # Pull request template
│   └── dependabot.yml          # Automated dependency updates
├── proto/                       # Protocol Buffer definitions
│   ├── go.mod                  # Proto module definition
│   └── peervault/              # Generated protobuf code
│       ├── go.mod              # Generated module definition
│       ├── go.sum              # Generated module checksums
│       ├── peervault.pb.go     # Generated protobuf code
│       └── peervault_grpc.pb.go # Generated gRPC code
├── config/                      # Configuration files
│   ├── peervault.yaml          # Main configuration
│   ├── test-demo.yaml          # Demo configuration
│   ├── test-demo2.yaml         # Demo configuration 2
│   ├── test-generated.yaml     # Generated test configuration
│   ├── mock-server.yaml        # Mock server configuration
│   └── codecov.yml             # Code coverage configuration
├── bin/                         # Build artifacts (generated)
├── storage/                     # Storage data directory (generated)
├── Makefile                     # Build automation for Unix systems
├── Taskfile.yml                 # Cross-platform task runner
├── .gitignore                   # Git ignore patterns
├── .golangci.yml               # Linting configuration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── README.md                   # This file - project overview
```

### Key Components

- **`cmd/`**: Application entrypoints for different use cases (main, node, demo, APIs, config)
- **`internal/`**: Core application code organized by domain
- **`internal/api/`**: API interfaces including GraphQL, REST, gRPC, and comprehensive testing frameworks
- **`internal/api/mocking/`**: API mocking framework for development and testing
- **`internal/api/contracts/`**: Contract testing framework for API compatibility
- **`internal/api/performance/`**: Performance testing framework with load testing
- **`internal/api/security/`**: Security testing framework with OWASP compliance
- **`internal/auth/`**: Authentication and authorization (RBAC)
- **`internal/audit/`**: Audit logging and security monitoring
- **`internal/backup/`**: Backup and disaster recovery
- **`internal/blockchain/`**: Blockchain integration and smart contracts
- **`internal/cache/`**: Multi-level caching system
- **`internal/compression/`**: Data compression utilities
- **`internal/config/`**: Configuration management and validation
- **`internal/content/`**: Content addressing and CID support
- **`internal/deduplication/`**: Content-based data deduplication
- **`internal/edge/`**: Edge computing and task distribution
- **`internal/health/`**: Health checking and monitoring
- **`internal/ipfs/`**: IPFS protocol compatibility
- **`internal/iot/`**: IoT device management and sensor data
- **`internal/metrics/`**: Prometheus-compatible metrics collection
- **`internal/ml/`**: Machine learning and AI integration
- **`internal/pki/`**: PKI and certificate management
- **`internal/plugins/`**: Plugin architecture and management
- **`internal/pool/`**: Object and connection pooling
- **`internal/privacy/`**: Data privacy and compliance controls
- **`internal/tracing/`**: Distributed tracing
- **`plugins/`**: Plugin implementations (S3 storage, etc.)
- **`security/`**: Security infrastructure, policies, and tools
- **`tests/`**: Comprehensive test suite with unit, integration, fuzz, and API testing
- **`tests/api/`**: API testing collections and mock data
- **`tests/contracts/`**: Contract testing definitions and verification
- **`tests/performance/`**: Performance testing with load and stress tests
- **`tests/security/`**: Security testing with OWASP compliance tests
- **`documentation/`**: Complete project documentation
- **`docs/`**: API documentation, SDKs, examples, and developer portal
- **`scripts/`**: Cross-platform build, automation, security, and API testing scripts
- **`scripts/api-testing/`**: API testing automation and test runners
- **`docker/`**: Containerization for development and production
- **`.github/`**: CI/CD pipeline with security integration and GitHub automation
- **`proto/`**: Protocol Buffer definitions and generated code
- **`config/`**: Configuration files and templates

## Test

```bash
go test ./...
# or
make test

# Test GraphQL API specifically
go test ./tests/integration/graphql/ -v

# Test REST API specifically
go test ./tests/integration/rest/ -v

# Test API testing frameworks
go test ./internal/api/contracts/... ./internal/api/performance/... ./internal/api/security/...

# Run comprehensive API test suite
./scripts/api-testing/run-tests.sh

# Run contract tests
go test ./tests/contracts/...

# Run performance tests
go test -bench=. ./tests/performance/...

# Run security tests
go test ./tests/security/...
```

## Lint

```bash
# Run the linter with custom configuration
golangci-lint run --config config/.golangci.yml

# Run with default configuration
golangci-lint run

# Run specific linters
golangci-lint run --disable-all --enable=errcheck,gofmt

# Fix formatting issues
go fmt ./...
goimports -w .

# Check for trailing whitespace in code files (CI check)
grep -r --include="*.go" --include="*.yml" --include="*.yaml" '[[:space:]]$' .

# Fix trailing whitespace (PowerShell)
Get-ChildItem -Recurse -Include "*.go", "*.yml", "*.yaml" | ForEach-Object { $content = Get-Content $_.FullName -Raw; $cleanContent = $content -replace '\s+$', ''; if ($content -ne $cleanContent) { Set-Content $_.FullName $cleanContent -NoNewline; Write-Host "Fixed trailing whitespace in: $($_.FullName)" } }

# Fix trailing whitespace (Unix/Linux)
find . -name "*.go" -o  -name "*.yml" -o -name "*.yaml" | xargs sed -i 's/[[:space:]]*$//'

# Check code formatting (CI check)
gofmt -s -l .

# Fix code formatting
gofmt -s -w .
```

### CI Pipeline Checks

The CI pipeline automatically runs these checks on every push and pull request:

#### **Critical Checks** (Pipeline fails if these fail)

- **Unit Tests**: `go test -v -race ./tests/unit/...` and `go test -v -race ./internal/...`
- **Fuzz Tests**: `go test -fuzz=Fuzz -fuzztime=30s ./tests/fuzz/...`
- **Security Scanning**: `semgrep` and `govulncheck`
- **Build Tests**: Cross-platform binary builds
- **Docker Tests**: Container builds and validation

#### **Non-Critical Checks** (Pipeline passes with warnings)

- **Integration Tests**: `go test -v -timeout=10m ./tests/integration/...` (shows warnings if failed - these are application logic bugs)
- **Linting**: `golangci-lint run ./...` (shows warnings if failed)
- **Code Formatting**: `gofmt -s -l .` (shows warnings if code is not formatted)
- **Trailing Whitespace**: Checks for trailing spaces in code files (Go, YAML, YML) - excludes markdown files
- **Code Quality**: Cyclomatic complexity and import checks
- **Documentation**: Exported function comments and README checks

**Pro tip**: Run the pre-commit script before pushing to keep code clean!

```bash
# Unix/Linux/macOS
./scripts/pre-commit.sh

# Windows PowerShell
.\scripts\pre-commit.ps1

# Or run individual commands:
go fmt ./...
goimports -w .
golangci-lint run ./...
```

**Note**: Lint and format failures won't block the pipeline, but it's good practice to keep code clean!

## Clean up local data

The demo writes files under a per-node storage root. To remove all data, delete the created folders (e.g., `ggnetwork` or the per-node roots you configured), or call `Store.Clear()` from your own code.

## Customize / next steps

- Turn the example into a long-running daemon and add a CLI/API
- Add proper peer discovery and resilient replication
- Replace demo logic in `main.go` with your own application code
- **GraphQL API**: Use the GraphQL API for building web applications and integrations
- **REST API**: Use the REST API for simple operations and webhook integrations
- **Real-time monitoring**: Leverage GraphQL subscriptions for real-time system monitoring
- **API extensions**: Extend the GraphQL schema with custom types and resolvers
