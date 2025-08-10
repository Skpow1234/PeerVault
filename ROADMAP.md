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

## Milestone 1 — Correctness and Security (P0)

1 Fix ID scoping mismatch (cross-node fetch/store)

- Problem: files are written/read under `id/key`, but `GetFile`/`StoreFile` use requester/sender `ID`, causing misses.
- Options:
  - A) Remove `ID` from on-disk path entirely and store by `hashedKey` only.
  - B) Use a stable owner/cluster scope (not the requester) when storing/serving.
- Acceptance: A node that does not have a file locally can fetch it from a peer successfully.
- Touchpoints: `internal/app/fileserver/server.go`, `internal/storage/store.go`.

2 Replace AES-CTR with authenticated encryption (AEAD)

- Problem: AES-CTR without authentication is malleable and offers no integrity.
- Solution: AES-GCM (preferred) or ChaCha20-Poly1305 with 12-byte nonce and AAD including message type/size.
- Update APIs: `CopyEncrypt`/`CopyDecrypt` to use AEAD; prepend nonce and auth tag; verify on decrypt.
- Acceptance: Crypto unit tests pass; tampering is detected; e2e still streams.
- Touchpoints: `internal/crypto/crypto.go`, `crypto_test.go`, call sites in `fileserver`.

3 Replace MD5/SHA-1 with SHA-256 (or HMAC-SHA-256 for hidden logical keys)

- `HashKey` → SHA-256 (or HMAC with cluster secret if keys must be concealed).
- CAS path transform → SHA-256, adapt block size segmentation.
- Acceptance: Existing tests updated; new expected paths validated.
- Touchpoints: `internal/crypto/crypto.go`, `internal/storage/store.go`, tests.

4 Key management model

- Current: Each node generates a random `EncKey`, but peers need the same key to decrypt replicated streams.
- Choose:
  - A) Shared cluster key loaded from env/config for demo.
  - B) Per-file keys derived from a KDF and exchanged via handshake.
  - C) Per-connection session keys (handshake), encrypt-in-transit only, plaintext at rest (or vice versa).
- Start with A) for demo simplicity; document B/C as next steps.
- Touchpoints: `cmd/peervault/main.go`, config, handshake.

5 Add authenticated transport handshake

- Implement handshake exchanging node identities and (for demo) verifying a pre-shared auth token or using Noise IK/XX.
- Produce a `PeerInfo {NodeID, PubKey}` and store in peer map.
- Acceptance: Only authenticated peers join; unauthenticated peers are rejected with clear logs.
- Touchpoints: `internal/transport/p2p/handshake.go`, `tcp_transport.go`, `fileserver.OnPeer`.

6 Message framing with length prefix

- Replace ad-hoc `DefaultDecoder` with a consistent frame: `[type:u8][len:u32][payload:len]`.
- For streams, send `[IncomingStream:u8][size:u64]` then raw bytes; for messages, the payload is encoded (JSON/CBOR/protobuf/gob).
- Acceptance: Fuzz tests for Decoder; no reliance on `time.Sleep`.
- Touchpoints: `internal/transport/p2p/encoding.go`, `tcp_transport.go`, `fileserver` send/receive.

---

## Milestone 2 — Transport, Streaming, and Storage (P1)

1 Remove `time.Sleep`-based coordination

- Replace with explicit acks and length-prefixed frames; block read until header/size is fully received.
- Touchpoints: `internal/app/fileserver/server.go`.

2 True streaming replication without buffering to memory

- Use `io.Pipe`: write to disk and encrypt to peers concurrently.
- Ensure backpressure via flow control; avoid unbounded memory usage.
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
