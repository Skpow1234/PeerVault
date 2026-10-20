# Milestones

Canonical detail: [documentation/ROADMAP.md](../documentation/ROADMAP.md) · Extended backlog: [documentation/MILESTONES_EXTENDED.md](../documentation/MILESTONES_EXTENDED.md)

## Current status

**MVP / core P2P restore in progress.** Recent work completed:

- Cross-peer **stream replication** and **pending Get** from peers
- **`DialPeer`** on TCP transport (health reconnect can dial)
- **SHA-256** CAS path transform (was SHA-1)
- Shared **`KeyManager`** helpers for multi-node demos/tests
- **OnStream** wired on main node/API entrypoints

Still open before treating Milestone 1–2 as fully closed:

- Streaming AEAD (crypto still buffers full payloads)
- Make multi-node e2e **CI-blocking**
- Wire REST / GraphQL / real gRPC to the fileserver

## Core milestones (P0–P8)

| # | Theme | Priority | Notes |
|---|--------|----------|--------|
| 1 | Correctness & security | P0 | Mostly done; streaming AEAD still open |
| 2 | Transport, streaming, storage | P1 | Replication restored; polish acks/backpressure |
| 3 | Reliability, ops, DX | P2 | Health/dial improved; e2e CI next |
| 4 | API interfaces | P3 | Scaffolding exists; wire to fileserver |
| 5 | Developer experience | P4 | Docs reorganized under `docs/` |
| 6 | Performance & scale | P5 | See extended backlog |
| 7 | Observability & prod readiness | P6 | See extended backlog |
| 8 | Security hardening & compliance | P7 | CI present; demo tokens remain |
| 9 | Advanced features & ecosystem | P8 | [milestone9](milestone9/README.md) — deprioritize vs core |

## Extended backlog (41+)

**903** additional milestones (41–943) cover API wiring, protocol TODOs, CI, security, Helm/K8s, observability, and hardening — generated from code review.

See [MILESTONES_EXTENDED.md](../documentation/MILESTONES_EXTENDED.md).

## Suggested next slices

1. Streaming chunked AES-GCM (no full-buffer encrypt/decrypt)
2. CI-blocking multi-node Store-on-A / Get-on-B test
3. Wire REST `FileService` to fileserver
4. Then true gRPC (replace HTTP/JSON temporary server)

## Related

- [deployment.md](deployment.md) · [endpoints.md](endpoints.md) · [architecture.md](architecture.md)
