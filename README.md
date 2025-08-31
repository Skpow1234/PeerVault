# PeerVault — P2P encrypted file store (Go)

Lightweight peer-to-peer file store demo written in Go. Nodes communicate over TCP, replicate files across peers, and encrypt file transfers using AES-CTR.

The included entrypoint at `cmd/peervault/main.go` boots 3 nodes locally and runs a simple store/get flow to demonstrate replication.

## Features

- Encrypted file streaming over TCP (AES-GCM with authentication)
- Advanced key management with derivation and rotation
- Authenticated peer connections with HMAC-SHA256 signatures
- Length-prefixed message framing for reliable transport
- Simple P2P transport abstraction (`internal/transport/p2p`)
- Content-addressable storage layout (SHA-256 based path transform)
- **GraphQL API** for flexible queries, mutations, and real-time subscriptions
- **REST API** for simple CRUD operations and webhook integrations
- **Interactive GraphQL Playground** for testing and development
- **Swagger UI** for REST API documentation and testing
- **Real-time monitoring** with health checks and metrics
- Minimal example that launches 3 local nodes and exchanges files

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
- **gosec** - Static analysis for security issues
- **Dependabot** - Automated dependency updates with security patches

## Windows Defender Setup

If you're developing on Windows, you may encounter Windows Defender popups when running the application or tests. This is because Go applications that create network connections and access the file system are often flagged as potentially suspicious.

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
```

### Windows (PowerShell)

```powershell
go build -o bin\peervault.exe .\cmd\peervault

# Build GraphQL server
go build -o bin\peervault-graphql.exe .\cmd\peervault-graphql
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
│   └── peervault-graphql/       # GraphQL API server binary
├── internal/                     # Core application code
│   ├── api/                     # API interfaces
│   │   └── graphql/             # GraphQL API implementation
│   │       ├── schema/          # GraphQL schema definitions
│   │       ├── types/           # Go types for GraphQL
│   │       ├── resolvers/       # GraphQL resolvers
│   │       └── server.go        # GraphQL HTTP server
│   ├── app/                     # Application logic
│   │   └── fileserver/          # Core file server implementation
│   ├── crypto/                  # Cryptographic functions and key management
│   ├── domain/                  # Domain entities and business logic
│   ├── dto/                     # Data transfer objects for network communication
│   ├── logging/                 # Logging utilities and configuration
│   ├── mapper/                  # Data mapping between domain and DTOs
│   ├── peer/                    # Peer management and health monitoring
│   ├── storage/                 # Content-addressable storage implementation
│   └── transport/               # Network transport layer
│       └── p2p/                 # P2P transport implementation
├── tests/                       # Comprehensive test suite
│   ├── unit/                    # Unit tests for all components
│   │   ├── concurrency/         # Concurrency safety tests
│   │   ├── crypto/              # Cryptographic function tests
│   │   ├── logging/             # Logging system tests
│   │   ├── peer/                # Peer management tests
│   │   ├── storage/             # Storage layer tests
│   │   └── transport/           # Transport layer tests
│   ├── integration/             # Integration tests
│   │   ├── end-to-end/          # End-to-end workflow tests
│   │   ├── graphql/             # GraphQL API integration tests
│   │   ├── multi-node/          # Multi-node network tests
│   │   └── performance/         # Performance and benchmark tests
│   ├── fuzz/                    # Fuzz testing for robustness
│   │   ├── crypto/              # Crypto layer fuzz tests
│   │   ├── storage/             # Storage layer fuzz tests
│   │   └── transport/           # Transport layer fuzz tests
│   ├── utils/                   # Test utilities and helpers
│   └── fixtures/                # Test data and fixtures
├── documentation/               # Project documentation
│   ├── README.md               # Documentation index
│   ├── CONTRIBUTING.md         # Contribution guidelines
│   ├── SECURITY.md             # Security policy
│   ├── ROADMAP.md              # Project roadmap
│   ├── ENCRYPTION.md           # Encryption implementation details
│   ├── LOGGING.md              # Logging system documentation
│   └── CONTAINERIZATION.md     # Docker and deployment guide
├── docs/                        # API and feature documentation
│   └── graphql/                # GraphQL API documentation
│       ├── README.md           # GraphQL API guide
│       └── schema.graphql      # GraphQL schema definition
├── scripts/                     # Build and automation scripts
│   ├── build.sh                # Unix build script
│   ├── build.ps1               # Windows build script
│   ├── run.sh                  # Unix run script
│   ├── run.ps1                 # Windows run script
│   ├── test.sh                 # Unix test script
│   └── test.ps1                # Windows test script
├── docker/                      # Containerization files
│   ├── Dockerfile              # Main application container
│   ├── Dockerfile.node         # Node-specific container
│   ├── Dockerfile.demo         # Demo client container
│   ├── docker-compose.yml      # Production multi-container setup
│   └── docker-compose.dev.yml  # Development container setup
├── .github/                     # GitHub configuration
│   ├── workflows/              # CI/CD pipeline configuration
│   ├── ISSUE_TEMPLATE/         # Issue and PR templates
│   └── dependabot.yml          # Automated dependency updates
├── bin/                         # Build artifacts (generated)
├── Makefile                     # Build automation for Unix systems
├── Taskfile.yml                 # Cross-platform task runner
├── .gitignore                   # Git ignore patterns
├── .golangci.yml               # Linting configuration
├── codecov.yml                 # Code coverage configuration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── README.md                   # This file - project overview
```

### Key Components

- **`cmd/`**: Application entrypoints for different use cases
- **`internal/`**: Core application code organized by domain
- **`internal/api/`**: API interfaces including GraphQL implementation
- **`tests/`**: Comprehensive test suite with unit, integration, and fuzz tests
- **`documentation/`**: Complete project documentation
- **`docs/`**: API and feature documentation including GraphQL
- **`scripts/`**: Cross-platform build and automation scripts
- **`docker/`**: Containerization for development and production
- **`.github/`**: CI/CD pipeline and GitHub automation

## Test

```bash
go test ./...
# or
make test

# Test GraphQL API specifically
go test ./tests/integration/graphql/ -v

# Test REST API specifically
go test ./tests/integration/rest/ -v
```

## Lint

```bash
# Run the linter with custom configuration
golangci-lint run --config config/.golangci.yml

# Run with default configuration
golangci-lint run

# Run specific linters
golangci-lint run --disable-all --enable=errcheck,gosec,gofmt

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
- **Security Scanning**: `gosec` and `govulncheck`
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
