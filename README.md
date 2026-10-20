# PeerVault

Peer-to-peer encrypted file storage in Go. Nodes talk over TCP, store content-addressed blobs, and encrypt with AES-GCM.

> **Status:** MVP development — core P2P and APIs are evolving. See [docs/milestones.md](docs/milestones.md).

[![Go](https://img.shields.io/badge/Go-1.24.6-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com/)
[![AES-GCM](https://img.shields.io/badge/AES--GCM-Encryption-green.svg)](https://en.wikipedia.org/wiki/Galois/Counter_Mode)
[![CI](https://img.shields.io/badge/GitHub%20Actions-CI-blue.svg)](https://github.com/features/actions)

## Quick start

```bash
# APIs + P2P nodes
docker compose -f docker-compose.apis.yml up -d --build

# P2P network only
docker compose up -d --build
```

| Service | URL |
|---------|-----|
| REST + Swagger | http://localhost:8080/docs |
| GraphQL Playground | http://localhost:8081/playground |
| gRPC/HTTP health | http://localhost:8082/health |

More options: **[docs/deployment.md](docs/deployment.md)** · Endpoints: **[docs/endpoints.md](docs/endpoints.md)**

## What it does

- Encrypted P2P file store over TCP (`internal/app/fileserver`, `internal/transport/p2p`)
- Content-addressable on-disk storage (`internal/storage`)
- AES-GCM encryption and HMAC peer handshake (`internal/crypto`)
- REST, GraphQL, and gRPC/HTTP-JSON API surfaces (`cmd/peervault-*`)

Feature list: [docs/features.md](docs/features.md) · How it works: [docs/architecture.md](docs/architecture.md)

## Documentation

| Doc | Description |
|-----|-------------|
| [docs/README.md](docs/README.md) | Documentation index |
| [docs/getting-started.md](docs/getting-started.md) | Requirements and first run |
| [docs/deployment.md](docs/deployment.md) | Docker Compose deploy paths |
| [docs/endpoints.md](docs/endpoints.md) | REST / GraphQL / gRPC routes |
| [docs/milestones.md](docs/milestones.md) | Roadmap milestones summary |
| [docs/security.md](docs/security.md) | Encryption, keys, auth |
| [docs/testing.md](docs/testing.md) | Tests and CI |
| [docs/plugins.md](docs/plugins.md) | Plugin overview |

API deep-dives: [REST](docs/api/README.md) · [GraphQL](docs/graphql/README.md) · [gRPC](docs/grpc/README.md) · [SDKs](docs/sdk/README.md)

Project docs: [Roadmap](documentation/ROADMAP.md) · [Containerization](documentation/CONTAINERIZATION.md) · [Contributing](documentation/CONTRIBUTING.md)

## Layout

```text
cmd/              # Binaries (node, demo, APIs, …)
internal/         # Core library (fileserver, transport, crypto, APIs)
docker/           # Dockerfiles
docs/             # User-facing documentation (this reorg)
documentation/    # Roadmap, policies, deep technical notes
tests/            # Unit, integration, fuzz, API tests
proto/            # Protobuf definitions
```

## Develop / CI

Tests and lint run in GitHub Actions (see [docs/testing.md](docs/testing.md)). Unit and fuzz checks are critical; some integration suites are non-blocking.

```bash
# Optional local helpers
make test    # if Make is available
task test    # if Task is available
```

## License

See [LICENSE](LICENSE).
