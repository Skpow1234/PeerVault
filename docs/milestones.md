# Milestones

Canonical detail lives in [documentation/ROADMAP.md](../documentation/ROADMAP.md). This page is a short map of priorities and where to look next.

## Current status

**MVP development** — focus on a stable P2P core (store/get, framing, encryption, handshake) and a single primary API path (gRPC with gateway). GraphQL/REST and ecosystem features are secondary until replication and peer fetch are solid.

Known gaps that still track early roadmap items:

- Cross-peer **replication** is currently disabled in `internal/app/fileserver`
- CAS path transform still uses **SHA-1** in storage (roadmap calls for SHA-256)
- gRPC server is still largely **HTTP/JSON** pending full protobuf streaming

## Core milestones (P0–P8)

| # | Theme | Priority | Notes |
|---|--------|----------|--------|
| 1 | Correctness & security | P0 | ID scoping, AEAD, SHA-256 CAS, keys, handshake, framing |
| 2 | Transport, streaming, storage | P1 | Remove sleep coordination, true streaming, discovery |
| 3 | Reliability, ops, DX | P2 | Config, health, Windows paths, tests |
| 4 | API interfaces | P3 | gRPC / REST / GraphQL integration |
| 5 | Developer experience | P4 | Docs, SDKs, portal |
| 6 | Performance & scale | P5 | Pools, cache, compression, dedup |
| 7 | Observability & prod readiness | P6 | Metrics, tracing, hardening |
| 8 | Security hardening & compliance | P7 | CI security pipeline (also see `documentation/PIPELINE_UPDATES.md`) |
| 9 | Advanced features & ecosystem | P8 | IPFS, chain, ML, edge, IoT — see [milestone9/README.md](milestone9/README.md) |

## API / protocol milestones (P9)

| # | Theme |
|---|--------|
| 10 | Advanced GraphQL |
| 11 | REST advanced features |
| 12 | gRPC advanced features |
| 13 | API documentation & testing |
| 14 | Multi-protocol (WS, MQTT, CoAP, …) |
| 15 | API analytics & monitoring |

## Later themes (P10+)

Roadmap continues through blockchain (16–20), AI ethics (25), edge/IoT (26–30), enterprise (31–35), and speculative future tech (36–40). Treat these as long-range planning, not near-term delivery.

## Suggested focus order

1. Finish Milestone **1–2** acceptance (replication, dial, SHA-256, no sleep-based protocol)
2. Make multi-node e2e tests **CI-blocking**
3. Wire one real API to the fileserver
4. Then expand APIs / Milestone 9+ ecosystem work

## Related

- Full roadmap: [documentation/ROADMAP.md](../documentation/ROADMAP.md)
- Milestone 9 feature notes: [milestone9/README.md](milestone9/README.md)
- Pipeline / Milestone 8: [documentation/PIPELINE_UPDATES.md](../documentation/PIPELINE_UPDATES.md)
