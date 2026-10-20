# Milestones

Canonical detail: [documentation/ROADMAP.md](../documentation/ROADMAP.md) · Extended backlog: [documentation/MILESTONES_EXTENDED.md](../documentation/MILESTONES_EXTENDED.md)

## Current status

**Core P2P MVP path is largely unblocked.** Completed recently:

- Cross-peer **stream replication** and **pending Get**
- **`DialPeer`** for health reconnect
- **SHA-256** CAS
- **Chunked AES-GCM** (`PVCH` container; legacy single-blob still decrypts)
- **REST FileService** wired to fileserver (upload/list/get/delete/metadata)
- **CI-blocking** `p2p-e2e` job (`TestReplicationStoreAGetB`)

Still open:

- GraphQL resolvers / true gRPC (not HTTP/JSON)
- Peer/System REST services still stubs
- Reduce sleep-based waits in older e2e helpers

## Core milestones (P0–P8)

| # | Theme | Priority | Notes |
|---|--------|----------|--------|
| 1 | Correctness & security | P0 | Largely done |
| 2 | Transport, streaming, storage | P1 | Chunked crypto + replication done |
| 3 | Reliability, ops, DX | P2 | Critical e2e gated in CI |
| 4 | API interfaces | P3 | REST files wired; GraphQL/gRPC next |
| 5–9 | DX / perf / obs / security / ecosystem | P4–P8 | See extended backlog |

## Extended backlog (41+)

See [MILESTONES_EXTENDED.md](../documentation/MILESTONES_EXTENDED.md) (900+ items).

## Suggested next slices

1. Wire GraphQL resolvers to fileserver  
2. Replace gRPC HTTP/JSON with real protobuf gRPC  
3. Wire REST PeerService / SystemService  
4. Handshake duplex + bootstrap readiness (no sleeps)

## Related

- [deployment.md](deployment.md) · [endpoints.md](endpoints.md) · [architecture.md](architecture.md)
