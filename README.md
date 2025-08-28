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

```
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

- Go 1.18+ (tested with Go 1.18)
- Make (optional; for Unix-like systems)

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
```

### Windows (PowerShell)

```powershell
go build -o bin\peervault.exe .\cmd\peervault
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

## Docker

Build and run the demo in a container:

```bash
docker build -t peervault .
docker run --rm -p 3000:3000 -p 5000:5000 -p 7000:7000 peervault
```

Note: the demo process launches 3 local nodes in one process and exposes multiple ports. For multi-container topologies, split the demo into separate processes (one port per container) and use a network for service discovery.

## How it works (high level)

- `cmd/peervault/main.go` creates 3 servers and bootstraps them together using the TCP transport in `internal/transport/p2p`.
- Files are written to disk under a content-addressed path derived from a SHA-1 of the key (`CASPathTransformFunc` in `internal/storage`).
- On store:
  - The file is written locally.
  - A control message is broadcast to peers so they can pull the encrypted stream and persist it.
- On get:
  - If not present locally, a request is broadcast and another peer streams the file back.
- Network messages are framed by a minimal protocol in `internal/transport/p2p` with small control bytes to distinguish messages vs. streams.

## Project layout

- `cmd/peervault/`: example entrypoint that boots 3 nodes and runs a demo store/get loop
- `internal/app/fileserver/`: core file server logic (broadcast, store, get, message handling, bootstrap)
- `internal/storage/`: content-addressable storage implementation
- `internal/crypto/`: ID/key generation and AES-CTR encrypt/decrypt helpers
- `internal/transport/p2p/`: TCP transport, peer management, message framing/decoding
- `internal/dto/`: message DTOs used over the wire
- `internal/domain/`, `internal/mapper/`: domain entities and mappers between domain and DTOs
- `Makefile`: build/run/test helpers for Unix-like systems
- `Dockerfile`, `.dockerignore`: container build and context exclusions

## Test

```bash
go test ./...
# or
make test
```

## Clean up local data

The demo writes files under a per-node storage root. To remove all data, delete the created folders (e.g., `ggnetwork` or the per-node roots you configured), or call `Store.Clear()` from your own code.

## Customize / next steps

- Turn the example into a long-running daemon and add a CLI/API
- Add proper peer discovery and resilient replication
- Replace demo logic in `main.go` with your own application code
