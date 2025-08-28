# PeerVault Roadmap

This roadmap organizes improvements by priority and theme. It reflects issues and opportunities identified across the codebase (`internal/app/fileserver`, `internal/transport/p2p`, `internal/storage`, `internal/crypto`, `cmd/peervault`).

## Guiding goals

- Correctness across nodes (consistent addressing and file discovery)
- Secure by default (authenticated encryption, authenticated transport)
- Robust transport (proper framing, no sleeps, backpressure)
- Efficient streaming (no unnecessary buffering)
- Operable and testable (observability, race-safety, e2e tests)
- Developer experience and portability (Windows-safe paths, simple run)

---

## Milestone 2 — Transport, Streaming, and Storage (P1)

1 Remove `time.Sleep`-based coordination ✅

- Problem: Relies on arbitrary delays instead of explicit acknowledgments.
- Solution: Replace with proper request-response protocol using acknowledgments.
- Acceptance: No `time.Sleep` calls in protocol logic; explicit acks for all operations.
- Touchpoints: `internal/app/fileserver/server.go`, `cmd/peervault/main.go`.

2 True streaming replication without buffering to memory ✅

- Problem: Entire files are buffered in memory before replication.
- Solution: Stream directly from disk to network peers without buffering.
- Acceptance: Large files can be replicated without memory exhaustion.
- Touchpoints: `internal/app/fileserver/server.go`.

3 Clarify encryption-at-rest vs in-transit

- Decide and document: encrypt at rest and in transit, or only in transit; keep paths consistent on store/get.
- If encrypting at rest, store ciphertext; decrypt on `Get` before returning.

4 Map concurrency safety

- Guard `peers` map with a `sync.RWMutex` and copy under lock before iteration.
- Provide helper methods `ListPeers()` and `WithPeers(fn)`.
- Touchpoints: `internal/app/fileserver/server.go`.

5 Logging and error context

- Switch to structured logging (zap/zerolog/log/slog).
- Wrap errors with context; standardize messages and include peer IDs.
- Touchpoints: transport and fileserver.

---

## Milestone 3 — Reliability, Ops, and DX (P2)

1 Peer lifecycle and health

- Heartbeats with timeouts; remove dead peers; reconnect with exponential backoff.
- Touchpoints: transport and fileserver.

2 Resource limits and backpressure

- Cap per-peer concurrent streams; add throttling; propagate cancellations with `context.Context`.

3 Windows portability

- Sanitize `StorageRoot` to avoid `:` in directory names; fix in code (not only README).
- Touchpoints: `cmd/peervault/main.go`, `internal/storage` defaults.

4 Containerization and multi-node runs

- Provide multi-container examples (one node per container) with a compose file; document ports and bootstrap.

5 Developer tooling

- Cross-platform run scripts (PowerShell + bash) or Taskfile; improve `Makefile` targets.

---

## Testing Plan

Unit tests

- Crypto: AEAD encrypt/decrypt roundtrip; tamper detection; nonce uniqueness.
- Storage: CAS path transform updates; read/write/delete happy/edge paths.
- Transport: decoder framing tests; fuzz `Decode` for partial reads.

Integration tests

- End-to-end replication: A stores; B fetches from network; validate bytes and size.
- Large file streaming (e.g., 100MB) within time/memory bounds.
- Disconnect mid-stream and ensure proper cleanup; subsequent retries succeed.
- Auth handshake: unauthenticated peer rejected; authenticated accepted.

Tooling

- Race detector: `go test -race ./...`
- Fuzzing for decoder: `go test -run ^$ -fuzz=Fuzz -fuzztime=30s ./internal/transport/p2p`

---

## Implementation Notes (by area)

Fileserver (`internal/app/fileserver/server.go`)

- Remove `time.Sleep`; wait for explicit acks or use size headers.
- Guard `peers` with `RWMutex`; add helper methods to broadcast safely.
- Rework `Store` to stream: `tee := io.TeeReader(r, localWriter)` and `io.Copy` into an `io.MultiWriter` over peers via an AEAD-wrapped writer.
- Align ID/path usage with new storage layout.

Transport (`internal/transport/p2p`)

- `encoding.go`: implement length-prefixed framing and robust reads.
- `handshake.go`: define `PeerInfo`, perform authentication (PSK demo; Noise later).
- `tcp_transport.go`: on stream header, block for size; remove reliance on goroutine `wg` coordination via magic control bytes.

Storage (`internal/storage`)

- Switch to SHA-256; update block segmentation size and tests.
- Consider removing `ID` from on-disk layout; store by `hashedKey` only, or store under `ownerID` not `requesterID`.

Crypto (`internal/crypto`)

- Implement AEAD helpers with nonce management, AAD, and constant-time checks.
- Replace MD5 with SHA-256/HMAC-SHA-256; provide `HashKey()` and `HashKeyHMAC(secret, key)`.

Entrypoint (`cmd/peervault/main.go`)

- Provide config/env for a cluster key (demo); sanitize `StorageRoot` for Windows.
- Optionally split nodes into separate processes for more realistic demos.

---

## Acceptance Checklist

- [ ] Cross-node fetch works consistently with new storage scoping
- [ ] AEAD enforced; tampering detected; unit tests cover crypto edge cases
- [ ] No `time.Sleep` for protocol timing; framed messages/streams only
- [ ] Decoder tolerant to partial reads; fuzz tests pass
- [ ] No data races (race detector clean); `peers` access synchronized
- [ ] SHA-256/HMAC adopted; CAS tests updated
- [ ] Windows path safe by default
- [ ] Structured logs with peer IDs and error wrapping
- [ ] Example e2e tests for store/get/replication/large files

---

## Future Enhancements

- Discovery (mDNS, DHT, static bootstrap with health checks)
- Replication factors and consistency (N-way replication, quorum reads)
- Content integrity (Merkle trees, chunking, resumable transfers)
- Pluggable codecs (JSON/CBOR/protobuf) and versioned message schemas
- Observability (metrics, tracing, pprof)
