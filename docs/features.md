# Features

## MVP focus

Shugo MVP targets core P2P file storage plus one primary API surface.

### Core P2P

- Encrypted file streaming over TCP (AES-GCM)
- Content-addressable storage (roadmap: SHA-256; see [milestones.md](milestones.md))
- Length-prefixed message framing (1MB max frame)
- Basic peer discovery and connection management
- File replication across connected peers (intended; see status in milestones)
- Local multi-node demo (`cmd/shugo` / Compose)

### Primary API (MVP intent)

- gRPC API with gateway-style REST compatibility
- Protocol Buffers for typed communication
- Store / retrieve / list files
- Health and system status
- HMAC-SHA256 auth with timestamp validation for peers

### Essential security

- AES-GCM with secure nonces
- HMAC-SHA256 peer authentication
- Key derivation from cluster key (`SHUGO_CLUSTER_KEY`)
- Vulnerability scanning in CI (govulncheck, semgrep, …)

### Basic observability

- Structured logging
- Health endpoints
- Basic metrics for files and peers

## Post-MVP (planned)

Not required for MVP completion:

- GraphQL subscriptions, standalone REST polish, interactive docs
- Enterprise compliance guidance (SOC 2 / GDPR / ISO — guidance only)
- RBAC, audit logging, PKI
- Caching, pools, compression, deduplication
- IPFS, blockchain, ML, edge, IoT
- Plugin ecosystem, richer SDKs, Prometheus/tracing/benchmarks

Full planning: [milestones.md](milestones.md) · [documentation/ROADMAP.md](../documentation/ROADMAP.md).
