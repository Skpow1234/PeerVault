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

## Milestone 1 — Correctness and Security (P0) ✅

1 Fix ID scoping mismatch (cross-node fetch/store) ✅

- Problem: files are written/read under `id/key`, but `GetFile`/`StoreFile` use requester/sender `ID`, causing misses.
- Options:
  - A) Remove `ID` from on-disk path entirely and store by `hashedKey` only.
  - B) Use a stable owner/cluster scope (not the requester) when storing/serving.
- Acceptance: A node that does not have a file locally can fetch it from a peer successfully.
- Touchpoints: `internal/app/fileserver/server.go`, `internal/storage/store.go`.

2 Replace AES-CTR with authenticated encryption (AEAD) ✅

- Problem: AES-CTR without authentication is malleable and offers no integrity.
- Solution: AES-GCM (preferred) or ChaCha20-Poly1305 with 12-byte nonce and AAD including message type/size.
- Update APIs: `CopyEncrypt`/`CopyDecrypt` to use AEAD; prepend nonce and auth tag; verify on decrypt.
- Acceptance: Crypto unit tests pass; tampering is detected; e2e still streams.
- Touchpoints: `internal/crypto/crypto.go`, `crypto_test.go`, call sites in `fileserver`.

3 Replace MD5/SHA-1 with SHA-256 (or HMAC-SHA-256 for hidden logical keys) ✅

- `HashKey` → SHA-256 (or HMAC with cluster secret if keys must be concealed).
- CAS path transform → SHA-256, adapt block size segmentation.
- Acceptance: Existing tests updated; new expected paths validated.
- Touchpoints: `internal/crypto/crypto.go`, `internal/storage/store.go`, tests.

4 Key management model ✅

- Current: Each node generates a random `EncKey`, but peers need the same key to decrypt replicated streams.
- Choose:
  - A) Shared cluster key loaded from env/config for demo.
  - B) Per-file keys derived from a KDF and exchanged via handshake.
  - C) Per-connection session keys (handshake), encrypt-in-transit only, plaintext at rest (or vice versa).
- Start with A) for demo simplicity; document B/C as next steps.
- Touchpoints: `cmd/peervault/main.go`, config, handshake.

5 Add authenticated transport handshake ✅

- Implement handshake exchanging node identities and (for demo) verifying a pre-shared auth token or using Noise IK/XX.
- Produce a `PeerInfo {NodeID, PubKey}` and store in peer map.
- Acceptance: Only authenticated peers join; unauthenticated peers are rejected with clear logs.
- Touchpoints: `internal/transport/p2p/handshake.go`, `tcp_transport.go`, `fileserver.OnPeer`.

6 Message framing with length prefix ✅

- Replace ad-hoc `DefaultDecoder` with a consistent frame: `[type:u8][len:u32][payload:len]`.
- For streams, send `[IncomingStream:u8][size:u64]` then raw bytes; for messages, the payload is encoded (JSON/CBOR/protobuf/gob).
- Acceptance: Fuzz tests for Decoder; no reliance on `time.Sleep`.
- Touchpoints: `internal/transport/p2p/encoding.go`, `tcp_transport.go`, `fileserver` send/receive.

---

## Milestone 2 — Transport, Streaming, and Storage (P1) ✅

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

3 Add proper peer discovery and resilient replication ✅

- Problem: Basic peer discovery and no resilient replication.
- Solution: Implement retry logic and better peer management.
- Acceptance: Replication continues even if some peers fail.
- Touchpoints: `internal/app/fileserver/server.go`.

4 Map concurrency safety ✅

- Problem: Race conditions in peer map access and inconsistent locking.
- Solution: Upgrade to RWMutex and ensure all peer access is properly synchronized.
- Acceptance: No race conditions in concurrent peer operations.
- Touchpoints: `internal/app/fileserver/server.go`.

5 Clarify encryption-at-rest vs in-transit ✅

- Problem: Inconsistent encryption strategy and unclear documentation.
- Solution: Implement encryption at rest + in transit with clear documentation.
- Acceptance: Consistent encryption behavior and documented strategy.
- Touchpoints: `internal/app/fileserver/server.go`, `ENCRYPTION.md`.

6 Logging and error context ✅

- Problem: Inconsistent logging with fmt.Printf and basic log.Println.
- Solution: Implement structured logging with slog and proper error context.
- Acceptance: Consistent structured logging with configurable levels and rich context.
- Touchpoints: `internal/app/fileserver/server.go`, `internal/logging/logger.go`, `LOGGING.md`.

---

## Milestone 3 — Reliability, Ops, and DX (P2)

1 Peer lifecycle and health ✅

- Problem: No peer health monitoring, dead peers remain indefinitely, no automatic reconnection.
- Solution: Implement health manager with heartbeats, timeouts, and exponential backoff reconnection.
- Acceptance: Dead peers are detected and removed, automatic reconnection with backoff, only healthy peers used for operations.
- Touchpoints: `internal/peer/health.go`, `internal/app/fileserver/server.go`.

2 Resource limits and backpressure ✅

- Cap per-peer concurrent streams; add throttling; propagate cancellations with `context.Context`.
- Acceptance: Per-peer concurrent stream limits enforced, rate limiting applied, context cancellation propagated throughout the system.
- Touchpoints: `internal/peer/resource_manager.go`, `internal/app/fileserver/server.go`, `cmd/peervault/main.go`.

3 Windows portability ✅

- Sanitize `StorageRoot` to avoid `:` in directory names; fix in code (not only README).
- Touchpoints: `cmd/peervault/main.go`, `internal/storage` defaults.

4 Containerization and multi-node runs ✅

- Provide multi-container examples (one node per container) with a compose file; document ports and bootstrap.
- Acceptance: Multi-container deployment with Docker Compose, separate node containers, demo client, comprehensive documentation.
- Touchpoints: `docker-compose.yml`, `Dockerfile.node`, `Dockerfile.demo`, `cmd/peervault-node/`, `cmd/peervault-demo/`, `CONTAINERIZATION.md`.

5 Developer tooling ✅

- Taskfile; improve `Makefile` targets.
- Acceptance: Cross-platform development tools with Taskfile, PowerShell scripts, bash scripts, improved Makefile, comprehensive testing structure.
- Touchpoints: `Taskfile.yml`, `scripts/`, `Makefile`, `tests/`.

---

## Testing Plan

Create a folder named tests, and subfolder with the different type of tests there, also, all the test that are already created, store it there, depending of there's needed
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
