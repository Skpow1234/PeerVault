# PeerVault Extended Milestones (41+)

Generated from codebase review after Milestone 1-2 core P2P fixes (replication, DialPeer, SHA-256 CAS, pending Get).
Each item is an actionable improvement derived from TODOs, stubs, ops gaps, and hardening needs.

## Progress note (core P2P restore)

* **Done:** AES-GCM, HashKey SHA-256, CAS SHA-256, DialPeer, stream replication, pending Get from peers, shared KeyManager helpers, OnStream on main entrypoints.
* **Still open:** streaming AEAD without full buffering; reduce sleep-based test coordination; multi-node e2e CI-blocking; wire REST/GraphQL/gRPC to fileserver.

## Milestone 41 - Streaming AES-GCM (chunked AEAD)

* Replace ReadAll-based CopyEncrypt/CopyDecrypt with chunked/streaming AEAD.
* Touchpoints: `internal/crypto/crypto.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 42 - Overwrite / upsert storage semantics

* Define replace-vs-reject when storing an existing key; support upsert for re-replication.
* Touchpoints: `internal/storage/store.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 43 - Delete + tombstone replication

* Replicate deletes across peers with tombstones and conflict rules.
* Touchpoints: `internal/app/fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 44 - List keys API on fileserver

* Expose local key listing for APIs and CLI.
* Touchpoints: `internal/app/fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 45 - Replication factor N

* Configurable N-way replication with success quorum.
* Touchpoints: `internal/app/fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 46 - Prefer healthy peers for Get

* Order Get candidates by RTT/health score.
* Touchpoints: `internal/peer`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 47 - Parallel Get fan-out

* Request file from multiple peers; take first successful stream.
* Touchpoints: `fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 48 - Cancel in-flight streams on ctx done

* Propagate context cancellation into stream copy loops.
* Touchpoints: `fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 49 - Backpressure on replicateToPeers

* Limit concurrent outbound replication streams per Store.
* Touchpoints: `fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 50 - Handshake deadlock-safe exchange

* Concurrent send/receive or duplex framing to avoid rare handshake stalls.
* Touchpoints: `handshake.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 51 - Transport metrics

* Emit dial/accept/stream/message counters and error rates.
* Touchpoints: `transport/p2p`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 52 - Peer address normalization

* Normalize host:port keys so dial/reconnect maps match OnPeer keys.
* Touchpoints: `fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 53 - Bootstrap retries with backoff

* Retry failed bootstrap dials with jittered exponential backoff.
* Touchpoints: `fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 54 - Persistent peer routing table

* Survive restarts with known peer list on disk.
* Touchpoints: `peer`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 55 - Content integrity verify on Get

* Re-hash ciphertext/plaintext and reject mismatches.
* Touchpoints: `storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 56 - Chunked file transfer protocol

* Split large objects into framed chunks with offsets for resume.
* Touchpoints: `transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 57 - Resumable transfers

* Support offset resume after disconnect.
* Touchpoints: `fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 58 - Deduplicate concurrent Stores

* Coalesce in-flight stores for the same key.
* Touchpoints: `FileOperationManager`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 59 - Observability spans for Store/Get

* Add tracing spans around store, replicate, get, stream.
* Touchpoints: `tracing`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 60 - Structured error codes on wire

* Typed error DTOs instead of opaque strings.
* Touchpoints: `dto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 61 - Wire REST FileService to fileserver

* Replace TODOs in file_service_impl with real Store/Get/List/Delete.
* Touchpoints: `file_service_impl.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 62 - Wire REST PeerService to peer map

* Implement list/add/remove using fileserver peer state.
* Touchpoints: `peer_service_impl.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 63 - Wire REST SystemService

* Real health/metrics from health and metrics packages.
* Touchpoints: `system_service_impl.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 64 - Wire GraphQL resolvers to fileserver

* Implement all resolver TODOs.
* Touchpoints: `resolver.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 65 - GraphQL upload uses Store

* fileserver Store from upload mutation.
* Touchpoints: `graphql/server.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 66 - GraphQL download uses Get

* Stream download via Get.
* Touchpoints: `graphql/server.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 67 - Replace gRPC HTTP/JSON with real gRPC

* Generate protobuf stubs and serve grpc.Server.
* Touchpoints: `api/grpc`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 68 - gRPC streaming Upload/Download

* Bidirectional chunk streaming mapped to fileserver.
* Touchpoints: `grpc/streaming`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 69 - grpc-gateway REST mapping

* Expose protobuf HTTP annotations for gateway.
* Touchpoints: `proto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 70 - Auth middleware shared across APIs

* One token/JWT validator for REST/GraphQL/gRPC.
* Touchpoints: `api`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 71 - Rate limit per peer and per API key

* Unified limiter config.
* Touchpoints: `ratelimit`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 72 - OpenAPI regenerated from code

* Keep OpenAPI in sync with routes.
* Touchpoints: `docs/api`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 73 - Contract tests CI-blocking

* Fail PR if consumer contracts break.
* Touchpoints: `tests/contracts`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 74 - Remove mock event generators in gRPC

* Replace mock events with real fileserver events.
* Touchpoints: `grpc/impl.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 75 - GraphQL subscriptions on real events

* Peer join/leave and file store events.
* Touchpoints: `subscriptions`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 76 - REST webhook emits on Store

* Configurable webhook on successful store/replicate.
* Touchpoints: `rest`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 77 - API request IDs

* Propagate X-Request-ID through logs and traces.
* Touchpoints: `middleware`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 78 - Versioned REST deprecation headers

* Emit Deprecation/Sunset headers.
* Touchpoints: `versioning`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 79 - GraphQL complexity limits

* Query depth/cost analysis to prevent abuse.
* Touchpoints: `graphql`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 80 - gRPC auth interceptor JSON parsing

* Finish TODO in interceptors/auth.go.
* Touchpoints: `auth.go`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 81 - MQTT wildcard subscriptions

* Implement + and # matching TODOs.
* Touchpoints: `mqtt`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 82 - MQTT session persistence

* Persist sessions across reconnect.
* Touchpoints: `mqtt`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 83 - MQTT over WebSocket

* WebSocket upgrade path for MQTT.
* Touchpoints: `mqtt`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 84 - CoAP DTLS

* DTLS support TODO on CoAP server.
* Touchpoints: `coap`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 85 - WebSocket topic distribution

* Topic-based fan-out TODO.
* Touchpoints: `websocket`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 86 - SSE backpressure

* Drop/slow clients cleanly under load.
* Touchpoints: `sse`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 87 - Protocol translation correctness tests

* Golden tests for translators.
* Touchpoints: `translation`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 88 - Unified health across protocol servers

* Same /health schema everywhere.
* Touchpoints: `cmd`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 89 - Shared TLS config loader

* One place for certs/mTLS for all HTTP servers.
* Touchpoints: `config`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 90 - Connection drain on shutdown

* Graceful shutdown TODO in GraphQL server.
* Touchpoints: `graphql`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: CLI

## Milestone 91 - CLI-1 CLI store/get wired

* Improve CLI: CLI store/get wired.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 92 - CLI-2 CLI config file support

* Improve CLI: CLI config file support.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 93 - CLI-3 CLI progress bars

* Improve CLI: CLI progress bars.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 94 - CLI-4 CLI JSON output

* Improve CLI: CLI JSON output.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 95 - CLI-5 CLI path completion

* Improve CLI: CLI path completion.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 96 - CLI-6 CLI profile switching

* Improve CLI: CLI profile switching.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 97 - CLI-7 CLI doctor command

* Improve CLI: CLI doctor command.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 98 - CLI-8 CLI peer list

* Improve CLI: CLI peer list.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 99 - CLI-9 CLI replicate status

* Improve CLI: CLI replicate status.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 100 - CLI-10 CLI TUI polish

* Improve CLI: CLI TUI polish.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 101 - CLI-11 Remove CLI sleep simulations

* Improve CLI: Remove CLI sleep simulations.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 102 - CLI-12 CLI integration tests

* Improve CLI: CLI integration tests.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 103 - CLI-13 Man pages help examples

* Improve CLI: Man pages help examples.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 104 - CLI-14 CLI version ldflags

* Improve CLI: CLI version ldflags.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 105 - CLI-15 CLI plugin subcommands

* Improve CLI: CLI plugin subcommands.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 106 - CLI-16 CLI offline error clarity

* Improve CLI: CLI offline error clarity.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 107 - CLI-17 CLI auth login

* Improve CLI: CLI auth login.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 108 - CLI-18 CLI batch store directory

* Improve CLI: CLI batch store directory.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 109 - CLI-19 CLI verify command

* Improve CLI: CLI verify command.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 110 - CLI-20 CLI benchmark command

* Improve CLI: CLI benchmark command.
* Touchpoints: `CLI`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Crypto

## Milestone 111 - Crypto-1 Key versioning in ciphertext header

* Improve Crypto: Key versioning in ciphertext header.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 112 - Crypto-2 Decrypt with previous keys after rotation

* Improve Crypto: Decrypt with previous keys after rotation.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 113 - Crypto-3 HKDF from x/crypto

* Improve Crypto: HKDF from x/crypto.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 114 - Crypto-4 Per-file DEK wrapped by KEK

* Improve Crypto: Per-file DEK wrapped by KEK.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 115 - Crypto-5 Hardware-backed keys interface

* Improve Crypto: Hardware-backed keys interface.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 116 - Crypto-6 Constant-time compare audit

* Improve Crypto: Constant-time compare audit.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 117 - Crypto-7 Nonce misuse review

* Improve Crypto: Nonce misuse review.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 118 - Crypto-8 AAD binding message type and key id

* Improve Crypto: AAD binding message type and key id.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 119 - Crypto-9 Secret zeroization helpers

* Improve Crypto: Secret zeroization helpers.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 120 - Crypto-10 Forbid demo auth token in prod

* Improve Crypto: Forbid demo auth token in prod.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 121 - Crypto-11 Fail closed if AUTH_TOKEN unset in prod

* Improve Crypto: Fail closed if AUTH_TOKEN unset in prod.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 122 - Crypto-12 Fail closed if CLUSTER_KEY unset in prod

* Improve Crypto: Fail closed if CLUSTER_KEY unset in prod.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 123 - Crypto-13 Handshake HMAC side-channel review

* Improve Crypto: Handshake HMAC side-channel review.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 124 - Crypto-14 Crypto fuzz corpus checked in

* Improve Crypto: Crypto fuzz corpus checked in.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 125 - Crypto-15 Benchmark Seal/Open

* Improve Crypto: Benchmark Seal/Open.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 126 - Crypto-16 ChaCha20-Poly1305 option

* Improve Crypto: ChaCha20-Poly1305 option.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 127 - Crypto-17 Age interop spike

* Improve Crypto: Age interop spike.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 128 - Crypto-18 Signed node identity ed25519

* Improve Crypto: Signed node identity ed25519.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 129 - Crypto-19 Mutual auth beyond PSK

* Improve Crypto: Mutual auth beyond PSK.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 130 - Crypto-20 Envelope encryption for backups

* Improve Crypto: Envelope encryption for backups.
* Touchpoints: `Crypto`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Docs-DX

## Milestone 131 - Docs-DX-1 ADR folder

* Improve Docs-DX: ADR folder.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 132 - Docs-DX-2 Honest README status automation

* Improve Docs-DX: Honest README status automation.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 133 - Docs-DX-3 Doc link checker CI

* Improve Docs-DX: Doc link checker CI.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 134 - Docs-DX-4 Compose secrets pattern example

* Improve Docs-DX: Compose secrets pattern example.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 135 - Docs-DX-5 Tutorial 3-node from zero

* Improve Docs-DX: Tutorial 3-node from zero.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 136 - Docs-DX-6 Tutorial API upload download

* Improve Docs-DX: Tutorial API upload download.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 137 - Docs-DX-7 Troubleshoot handshake failures

* Improve Docs-DX: Troubleshoot handshake failures.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 138 - Docs-DX-8 Troubleshoot replication misses

* Improve Docs-DX: Troubleshoot replication misses.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 139 - Docs-DX-9 Windows path FAQ

* Improve Docs-DX: Windows path FAQ.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 140 - Docs-DX-10 Contributing first-issue labels

* Improve Docs-DX: Contributing first-issue labels.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 141 - Docs-DX-11 CODEOWNERS

* Improve Docs-DX: CODEOWNERS.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 142 - Docs-DX-12 PR template API checklist

* Improve Docs-DX: PR template API checklist.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 143 - Docs-DX-13 Changelog generation

* Improve Docs-DX: Changelog generation.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 144 - Docs-DX-14 Semver policy

* Improve Docs-DX: Semver policy.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 145 - Docs-DX-15 Deprecation policy

* Improve Docs-DX: Deprecation policy.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 146 - Docs-DX-16 Public roadmap board sync

* Improve Docs-DX: Public roadmap board sync.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 147 - Docs-DX-17 SVG architecture diagram

* Improve Docs-DX: SVG architecture diagram.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 148 - Docs-DX-18 Sequence diagrams Store/Get

* Improve Docs-DX: Sequence diagrams Store/Get.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 149 - Docs-DX-19 SDK quickstarts tested CI

* Improve Docs-DX: SDK quickstarts tested CI.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 150 - Docs-DX-20 Local docs search

* Improve Docs-DX: Local docs search.
* Touchpoints: `Docs-DX`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Ecosystem

## Milestone 151 - Ecosystem-1 IPFS gateway compatibility tests

* Improve Ecosystem: IPFS gateway compatibility tests.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 152 - Ecosystem-2 CID to CAS mapping clarity

* Improve Ecosystem: CID to CAS mapping clarity.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 153 - Ecosystem-3 Blockchain module off by default

* Improve Ecosystem: Blockchain module off by default.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 154 - Ecosystem-4 Remove unused ethereum dep

* Improve Ecosystem: Remove unused ethereum dep.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 155 - Ecosystem-5 ML classifier optional build tag

* Improve Ecosystem: ML classifier optional build tag.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 156 - Ecosystem-6 Edge worker protocol spec

* Improve Ecosystem: Edge worker protocol spec.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 157 - Ecosystem-7 IoT device auth model

* Improve Ecosystem: IoT device auth model.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 158 - Ecosystem-8 Plugin sandboxing

* Improve Ecosystem: Plugin sandboxing.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 159 - Ecosystem-9 S3 plugin integration test

* Improve Ecosystem: S3 plugin integration test.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 160 - Ecosystem-10 Webhook HMAC signatures

* Improve Ecosystem: Webhook HMAC signatures.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 161 - Ecosystem-11 OTel collector example

* Improve Ecosystem: OTel collector example.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 162 - Ecosystem-12 JavaScript SDK publish

* Improve Ecosystem: JavaScript SDK publish.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 163 - Ecosystem-13 Go SDK module path stability

* Improve Ecosystem: Go SDK module path stability.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 164 - Ecosystem-14 Python client refresh

* Improve Ecosystem: Python client refresh.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 165 - Ecosystem-15 Postman collection auto-gen

* Improve Ecosystem: Postman collection auto-gen.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 166 - Ecosystem-16 Insomnia collection

* Improve Ecosystem: Insomnia collection.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 167 - Ecosystem-17 Terraform provider spike

* Improve Ecosystem: Terraform provider spike.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 168 - Ecosystem-18 K8s Operator CRD spike

* Improve Ecosystem: K8s Operator CRD spike.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 169 - Ecosystem-19 WASM node experiment

* Improve Ecosystem: WASM node experiment.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 170 - Ecosystem-20 Mobile constrained-network profile

* Improve Ecosystem: Mobile constrained-network profile.
* Touchpoints: `Ecosystem`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Fileserver

## Milestone 171 - Fileserver-1 Idempotent Store with content hash header

* Improve Fileserver: Idempotent Store with content hash header.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 172 - Fileserver-2 Multi-reader Get with shared buffer pool

* Improve Fileserver: Multi-reader Get with shared buffer pool.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 173 - Fileserver-3 Peer selection strategy plugin

* Improve Fileserver: Peer selection strategy plugin.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 174 - Fileserver-4 Replication job queue with retries

* Improve Fileserver: Replication job queue with retries.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 175 - Fileserver-5 Dead-letter queue for failed replications

* Improve Fileserver: Dead-letter queue for failed replications.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 176 - Fileserver-6 Per-key lease for concurrent writers

* Improve Fileserver: Per-key lease for concurrent writers.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 177 - Fileserver-7 Background verify-and-repair job

* Improve Fileserver: Background verify-and-repair job.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 178 - Fileserver-8 Storage GC for orphaned CAS objects

* Improve Fileserver: Storage GC for orphaned CAS objects.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 179 - Fileserver-9 Migrate legacy SHA-1 paths to SHA-256

* Improve Fileserver: Migrate legacy SHA-1 paths to SHA-256.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 180 - Fileserver-10 Export/import snapshot format

* Improve Fileserver: Export/import snapshot format.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 181 - Fileserver-11 Read-only mode for nodes

* Improve Fileserver: Read-only mode for nodes.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 182 - Fileserver-12 Maintenance mode endpoint

* Improve Fileserver: Maintenance mode endpoint.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 183 - Fileserver-13 Disk space watermark checks

* Improve Fileserver: Disk space watermark checks.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 184 - Fileserver-14 Windows path length edge cases

* Improve Fileserver: Windows path length edge cases.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 185 - Fileserver-15 Fsync policy

* Improve Fileserver: Fsync policy.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 186 - Fileserver-16 Checksum trailer on streams

* Improve Fileserver: Checksum trailer on streams.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 187 - Fileserver-17 Compress-before-encrypt option

* Improve Fileserver: Compress-before-encrypt option.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 188 - Fileserver-18 Namespace/tenant prefix on keys

* Improve Fileserver: Namespace/tenant prefix on keys.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 189 - Fileserver-19 Quota per namespace

* Improve Fileserver: Quota per namespace.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 190 - Fileserver-20 Audit event on Store/Get

* Improve Fileserver: Audit event on Store/Get.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 191 - Fileserver-21 PII redaction in logs for keys

* Improve Fileserver: PII redaction in logs for keys.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 192 - Fileserver-22 Slow-peer eviction policy

* Improve Fileserver: Slow-peer eviction policy.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 193 - Fileserver-23 Circuit breaker around DialPeer

* Improve Fileserver: Circuit breaker around DialPeer.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 194 - Fileserver-24 Happy-eyeballs dual-stack dial

* Improve Fileserver: Happy-eyeballs dual-stack dial.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 195 - Fileserver-25 SOCKS5 proxy dial

* Improve Fileserver: SOCKS5 proxy dial.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 196 - Fileserver-26 mDNS local discovery

* Improve Fileserver: mDNS local discovery.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 197 - Fileserver-27 Static peers file watcher

* Improve Fileserver: Static peers file watcher.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 198 - Fileserver-28 DNS SRV bootstrap

* Improve Fileserver: DNS SRV bootstrap.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 199 - Fileserver-29 NAT traversal research spike

* Improve Fileserver: NAT traversal research spike.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 200 - Fileserver-30 Encrypt-then-MAC metadata

* Improve Fileserver: Encrypt-then-MAC metadata.
* Touchpoints: `Fileserver`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: GraphQL

## Milestone 201 - GraphQL-1 Persisted queries

* Improve GraphQL: Persisted queries.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 202 - GraphQL-2 Automatic persisted queries

* Improve GraphQL: Automatic persisted queries.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 203 - GraphQL-3 DataLoader batching

* Improve GraphQL: DataLoader batching.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 204 - GraphQL-4 Field-level auth directives

* Improve GraphQL: Field-level auth directives.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 205 - GraphQL-5 Schema stitching production readiness

* Improve GraphQL: Schema stitching production readiness.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 206 - GraphQL-6 Federation gateway hardening

* Improve GraphQL: Federation gateway hardening.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 207 - GraphQL-7 Subscription auth on WS

* Improve GraphQL: Subscription auth on WS.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 208 - GraphQL-8 Heartbeat on subscriptions

* Improve GraphQL: Heartbeat on subscriptions.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 209 - GraphQL-9 Query allowlist mode

* Improve GraphQL: Query allowlist mode.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 210 - GraphQL-10 Resolver timeout per field

* Improve GraphQL: Resolver timeout per field.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 211 - GraphQL-11 N+1 detection in tests

* Improve GraphQL: N+1 detection in tests.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 212 - GraphQL-12 GraphQL request/error metrics

* Improve GraphQL: GraphQL request/error metrics.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 213 - GraphQL-13 Playground disabled in prod

* Improve GraphQL: Playground disabled in prod.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 214 - GraphQL-14 File scalar streaming

* Improve GraphQL: File scalar streaming.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 215 - GraphQL-15 Relay cursor connections

* Improve GraphQL: Relay cursor connections.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 216 - GraphQL-16 Deprecate unused fields

* Improve GraphQL: Deprecate unused fields.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 217 - GraphQL-17 Schema lint in CI

* Improve GraphQL: Schema lint in CI.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 218 - GraphQL-18 Schema contract tests

* Improve GraphQL: Schema contract tests.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 219 - GraphQL-19 Example queries regenerated

* Improve GraphQL: Example queries regenerated.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 220 - GraphQL-20 Resolver unit tests every field

* Improve GraphQL: Resolver unit tests every field.
* Touchpoints: `GraphQL`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: gRPC

## Milestone 221 - gRPC-1 protoc generation check in CI

* Improve gRPC: protoc generation check in CI.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 222 - gRPC-2 buf lint breaking change detect

* Improve gRPC: buf lint breaking change detect.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 223 - gRPC-3 grpc health v1

* Improve gRPC: grpc health v1.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 224 - gRPC-4 Reflection toggle

* Improve gRPC: Reflection toggle.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 225 - gRPC-5 grpc-web CORS hardening

* Improve gRPC: grpc-web CORS hardening.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 226 - gRPC-6 Client SDK publish pipeline

* Improve gRPC: Client SDK publish pipeline.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 227 - gRPC-7 Retry policies

* Improve gRPC: Retry policies.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 228 - gRPC-8 Request hedging

* Improve gRPC: Request hedging.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 229 - gRPC-9 Load balancing config examples

* Improve gRPC: Load balancing config examples.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 230 - gRPC-10 Server-side keepalive

* Improve gRPC: Server-side keepalive.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 231 - gRPC-11 Max message size config

* Improve gRPC: Max message size config.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 232 - gRPC-12 Compression gzip

* Improve gRPC: Compression gzip.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 233 - gRPC-13 Richer gRPC status details

* Improve gRPC: Richer gRPC status details.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 234 - gRPC-14 Interceptor chain metrics

* Improve gRPC: Interceptor chain metrics.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 235 - gRPC-15 Method-level authz

* Improve gRPC: Method-level authz.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 236 - gRPC-16 Real grpc client integration test

* Improve gRPC: Real grpc client integration test.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 237 - gRPC-17 Benchmark streaming throughput

* Improve gRPC: Benchmark streaming throughput.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 238 - gRPC-18 Remove temporary HTTP/JSON server

* Improve gRPC: Remove temporary HTTP/JSON server.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 239 - gRPC-19 Gateway OpenAPI from proto

* Improve gRPC: Gateway OpenAPI from proto.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 240 - gRPC-20 Versioned package peervault.v1

* Improve gRPC: Versioned package peervault.v1.
* Touchpoints: `gRPC`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Observability

## Milestone 241 - Observability-1 RED metrics Store/Get

* Improve Observability: RED metrics Store/Get.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 242 - Observability-2 Peer connection gauge

* Improve Observability: Peer connection gauge.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 243 - Observability-3 Replication lag metric

* Improve Observability: Replication lag metric.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 244 - Observability-4 Storage bytes gauge

* Improve Observability: Storage bytes gauge.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 245 - Observability-5 Log sampling under load

* Improve Observability: Log sampling under load.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 246 - Observability-6 Trace sampling config

* Improve Observability: Trace sampling config.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 247 - Observability-7 Correlate logs with trace IDs

* Improve Observability: Correlate logs with trace IDs.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 248 - Observability-8 Slow request log threshold

* Improve Observability: Slow request log threshold.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 249 - Observability-9 Error budget dashboard

* Improve Observability: Error budget dashboard.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 250 - Observability-10 SLO definitions

* Improve Observability: SLO definitions.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 251 - Observability-11 Synthetic canary Store/Get

* Improve Observability: Synthetic canary Store/Get.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 252 - Observability-12 Blackbox exporter targets

* Improve Observability: Blackbox exporter targets.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 253 - Observability-13 Continuous profiling

* Improve Observability: Continuous profiling.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 254 - Observability-14 Internal domain event bus

* Improve Observability: Internal domain event bus.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 255 - Observability-15 Audit vs debug log separation

* Improve Observability: Audit vs debug log separation.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 256 - Observability-16 Hot-reload log level

* Improve Observability: Hot-reload log level.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 257 - Observability-17 Health dependency detail

* Improve Observability: Health dependency detail.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 258 - Observability-18 Version build info metric

* Improve Observability: Version build info metric.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 259 - Observability-19 Panic recovery metrics

* Improve Observability: Panic recovery metrics.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 260 - Observability-20 Replication queue depth metrics

* Improve Observability: Replication queue depth metrics.
* Touchpoints: `Observability`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Ops-Deploy

## Milestone 261 - Ops-Deploy-1 Helm chart first class

* Improve Ops-Deploy: Helm chart first class.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 262 - Ops-Deploy-2 Kustomize overlays

* Improve Ops-Deploy: Kustomize overlays.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 263 - Ops-Deploy-3 HPA guidance

* Improve Ops-Deploy: HPA guidance.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 264 - Ops-Deploy-4 PersistentVolume examples

* Improve Ops-Deploy: PersistentVolume examples.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 265 - Ops-Deploy-5 PodDisruptionBudget

* Improve Ops-Deploy: PodDisruptionBudget.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 266 - Ops-Deploy-6 Anti-affinity for nodes

* Improve Ops-Deploy: Anti-affinity for nodes.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 267 - Ops-Deploy-7 ServiceMonitor Prometheus

* Improve Ops-Deploy: ServiceMonitor Prometheus.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 268 - Ops-Deploy-8 Grafana dashboard JSON

* Improve Ops-Deploy: Grafana dashboard JSON.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 269 - Ops-Deploy-9 Alerting rules pack

* Improve Ops-Deploy: Alerting rules pack.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 270 - Ops-Deploy-10 Liveness readiness docs

* Improve Ops-Deploy: Liveness readiness docs.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 271 - Ops-Deploy-11 Rolling update strategy

* Improve Ops-Deploy: Rolling update strategy.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 272 - Ops-Deploy-12 Backup CronJob example

* Improve Ops-Deploy: Backup CronJob example.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 273 - Ops-Deploy-13 Disaster recovery runbook

* Improve Ops-Deploy: Disaster recovery runbook.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 274 - Ops-Deploy-14 Multi-region design doc

* Improve Ops-Deploy: Multi-region design doc.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 275 - Ops-Deploy-15 Terraform module spike

* Improve Ops-Deploy: Terraform module spike.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 276 - Ops-Deploy-16 Ansible role spike

* Improve Ops-Deploy: Ansible role spike.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 277 - Ops-Deploy-17 systemd unit files

* Improve Ops-Deploy: systemd unit files.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 278 - Ops-Deploy-18 sysctl tuning notes

* Improve Ops-Deploy: sysctl tuning notes.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 279 - Ops-Deploy-19 Resource requests limits defaults

* Improve Ops-Deploy: Resource requests limits defaults.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 280 - Ops-Deploy-20 Cost estimation worksheet

* Improve Ops-Deploy: Cost estimation worksheet.
* Touchpoints: `Ops-Deploy`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Performance

## Milestone 281 - Performance-1 Buffer pool frame headers

* Improve Performance: Buffer pool frame headers.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 282 - Performance-2 Object pool Message structs

* Improve Performance: Object pool Message structs.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 283 - Performance-3 Zero-alloc hot path audit

* Improve Performance: Zero-alloc hot path audit.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 284 - Performance-4 Parallel replicate workers

* Improve Performance: Parallel replicate workers.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 285 - Performance-5 Batch Has lookups

* Improve Performance: Batch Has lookups.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 286 - Performance-6 Optional recent Get cache

* Improve Performance: Optional recent Get cache.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 287 - Performance-7 CDN cache headers REST

* Improve Performance: CDN cache headers REST.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 288 - Performance-8 Compression negotiation

* Improve Performance: Compression negotiation.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 289 - Performance-9 Benchmark Store 1/10/100MB

* Improve Performance: Benchmark Store 1/10/100MB.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 290 - Performance-10 Benchmark Get local vs remote

* Improve Performance: Benchmark Get local vs remote.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 291 - Performance-11 p99 latency budgets

* Improve Performance: p99 latency budgets.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 292 - Performance-12 GC tuning notes

* Improve Performance: GC tuning notes.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 293 - Performance-13 File cache guidance

* Improve Performance: File cache guidance.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 294 - Performance-14 Verify no double-encrypt on replicate

* Improve Performance: Verify no double-encrypt on replicate.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 295 - Performance-15 sendfile splice exploration

* Improve Performance: sendfile splice exploration.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 296 - Performance-16 io_uring experiment

* Improve Performance: io_uring experiment.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 297 - Performance-17 Connection pool popular peers

* Improve Performance: Connection pool popular peers.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 298 - Performance-18 Adaptive chunk size

* Improve Performance: Adaptive chunk size.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 299 - Performance-19 Read-ahead sequential Get

* Improve Performance: Read-ahead sequential Get.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 300 - Performance-20 Write coalescing

* Improve Performance: Write coalescing.
* Touchpoints: `Performance`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: REST

## Milestone 301 - REST-1 Bearer auth real validator

* Improve REST: Bearer auth real validator.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 302 - REST-2 Upload size limits

* Improve REST: Upload size limits.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 303 - REST-3 Streaming upload to Store

* Improve REST: Streaming upload to Store.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 304 - REST-4 Streaming download from Get

* Improve REST: Streaming download from Get.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 305 - REST-5 Multipart form upload

* Improve REST: Multipart form upload.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 306 - REST-6 ETag If-None-Match

* Improve REST: ETag If-None-Match.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 307 - REST-7 Conditional delete

* Improve REST: Conditional delete.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 308 - REST-8 Cursor pagination for list

* Improve REST: Cursor pagination for list.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 309 - REST-9 Filter/search query params

* Improve REST: Filter/search query params.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 310 - REST-10 CORS allowlist from config

* Improve REST: CORS allowlist from config.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 311 - REST-11 Trusted proxy headers

* Improve REST: Trusted proxy headers.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 312 - REST-12 Prometheus metrics authenticity

* Improve REST: Prometheus metrics authenticity.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 313 - REST-13 OpenTelemetry HTTP instrumentation

* Improve REST: OpenTelemetry HTTP instrumentation.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 314 - REST-14 Idempotency-Key support

* Improve REST: Idempotency-Key support.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 315 - REST-15 Problem+JSON errors

* Improve REST: Problem+JSON errors.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 316 - REST-16 Admin-only routes

* Improve REST: Admin-only routes.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 317 - REST-17 API key issuance CLI

* Improve REST: API key issuance CLI.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 318 - REST-18 Per-route feature flags

* Improve REST: Per-route feature flags.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 319 - REST-19 Chaos middleware

* Improve REST: Chaos middleware.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 320 - REST-20 Nightly load test job

* Improve REST: Nightly load test job.
* Touchpoints: `REST`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Security

## Milestone 321 - Security-1 Remove hardcoded demo tokens

* Improve Security: Remove hardcoded demo tokens.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 322 - Security-2 Secrets scanning pre-commit required

* Improve Security: Secrets scanning pre-commit required.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 323 - Security-3 Living threat model doc

* Improve Security: Living threat model doc.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 324 - Security-4 Get amplification abuse cases

* Improve Security: Get amplification abuse cases.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 325 - Security-5 Max frame size DoS tests

* Improve Security: Max frame size DoS tests.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 326 - Security-6 Auth token rotation runbook

* Improve Security: Auth token rotation runbook.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 327 - Security-7 Cluster key rotation runbook

* Improve Security: Cluster key rotation runbook.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 328 - Security-8 Peer ban after auth failures

* Improve Security: Peer ban after auth failures.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 329 - Security-9 Audit log sink

* Improve Security: Audit log sink.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 330 - Security-10 RBAC on mutating APIs

* Improve Security: RBAC on mutating APIs.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 331 - Security-11 Privacy controls metadata export

* Improve Security: Privacy controls metadata export.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 332 - Security-12 PKI issue rotate node certs

* Improve Security: PKI issue rotate node certs.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 333 - Security-13 Container non-root verification

* Improve Security: Container non-root verification.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 334 - Security-14 Read-only root FS compose

* Improve Security: Read-only root FS compose.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 335 - Security-15 Seccomp AppArmor profiles

* Improve Security: Seccomp AppArmor profiles.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 336 - Security-16 K8s NetworkPolicy examples

* Improve Security: K8s NetworkPolicy examples.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 337 - Security-17 Optional pen-test CI job

* Improve Security: Optional pen-test CI job.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 338 - Security-18 Dependency vulnerability SLA

* Improve Security: Dependency vulnerability SLA.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 339 - Security-19 security.txt contact

* Improve Security: security.txt contact.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 340 - Security-20 Bug bounty prep

* Improve Security: Bug bounty prep.
* Touchpoints: `Security`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Storage

## Milestone 341 - Storage-1 Pluggable PathTransformFunc registry

* Improve Storage: Pluggable PathTransformFunc registry.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 342 - Storage-2 2-char fan-out option

* Improve Storage: 2-char fan-out option.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 343 - Storage-3 Object metadata sidecars

* Improve Storage: Object metadata sidecars.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 344 - Storage-4 Content-type metadata map

* Improve Storage: Content-type metadata map.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 345 - Storage-5 Atomic replace via temp+rename

* Improve Storage: Atomic replace via temp+rename.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 346 - Storage-6 Cross-volume rename handling

* Improve Storage: Cross-volume rename handling.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 347 - Storage-7 Sparse file support

* Improve Storage: Sparse file support.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 348 - Storage-8 Memory-mapped read option

* Improve Storage: Memory-mapped read option.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 349 - Storage-9 Checksum database index

* Improve Storage: Checksum database index.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 350 - Storage-10 Bolt/SQLite key index

* Improve Storage: Bolt/SQLite key index.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 351 - Storage-11 Bloom filter for Has

* Improve Storage: Bloom filter for Has.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 352 - Storage-12 Sharded locks per path prefix

* Improve Storage: Sharded locks per path prefix.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 353 - Storage-13 Concurrent Has/Read stress tests

* Improve Storage: Concurrent Has/Read stress tests.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 354 - Storage-14 Storage Root absolute path policy

* Improve Storage: Storage Root absolute path policy.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 355 - Storage-15 Multi-disk volume spanning

* Improve Storage: Multi-disk volume spanning.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 356 - Storage-16 Read-only mount detection

* Improve Storage: Read-only mount detection.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 357 - Storage-17 Corruption injector for tests

* Improve Storage: Corruption injector for tests.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 358 - Storage-18 Backup hot-copy API

* Improve Storage: Backup hot-copy API.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 359 - Storage-19 Incremental backup diffs

* Improve Storage: Incremental backup diffs.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 360 - Storage-20 CAS path docs match code

* Improve Storage: CAS path docs match code.
* Touchpoints: `Storage`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Testing-CI

## Milestone 361 - Testing-CI-1 Make multi-node e2e CI-blocking

* Improve Testing-CI: Make multi-node e2e CI-blocking.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 362 - Testing-CI-2 Flake quarantine process

* Improve Testing-CI: Flake quarantine process.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 363 - Testing-CI-3 Testcontainers compose tests

* Improve Testing-CI: Testcontainers compose tests.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 364 - Testing-CI-4 Race detector on fileserver

* Improve Testing-CI: Race detector on fileserver.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 365 - Testing-CI-5 Coverage gate fileserver

* Improve Testing-CI: Coverage gate fileserver.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 366 - Testing-CI-6 Coverage gate transport

* Improve Testing-CI: Coverage gate transport.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 367 - Testing-CI-7 Mutation testing spike

* Improve Testing-CI: Mutation testing spike.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 368 - Testing-CI-8 Nightly longer fuzz

* Improve Testing-CI: Nightly longer fuzz.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 369 - Testing-CI-9 SBOM in release

* Improve Testing-CI: SBOM in release.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 370 - Testing-CI-10 Signed release artifacts

* Improve Testing-CI: Signed release artifacts.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 371 - Testing-CI-11 Reproducible builds

* Improve Testing-CI: Reproducible builds.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 372 - Testing-CI-12 Go version matrix

* Improve Testing-CI: Go version matrix.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 373 - Testing-CI-13 Windows Linux CI parity

* Improve Testing-CI: Windows Linux CI parity.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 374 - Testing-CI-14 Optional Mac CI

* Improve Testing-CI: Optional Mac CI.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 375 - Testing-CI-15 Go module caching fix

* Improve Testing-CI: Go module caching fix.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 376 - Testing-CI-16 Dependabot review SLA

* Improve Testing-CI: Dependabot review SLA.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 377 - Testing-CI-17 Action pinning to SHAs

* Improve Testing-CI: Action pinning to SHAs.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 378 - Testing-CI-18 CodeQL workflow

* Improve Testing-CI: CodeQL workflow.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 379 - Testing-CI-19 License check go-licenses

* Improve Testing-CI: License check go-licenses.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 380 - Testing-CI-20 Dead code detection

* Improve Testing-CI: Dead code detection.
* Touchpoints: `Testing-CI`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Transport

## Milestone 381 - Transport-1 Replace gob with protobuf DTOs

* Improve Transport: Replace gob with protobuf DTOs.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 382 - Transport-2 Message schema version byte

* Improve Transport: Message schema version byte.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 383 - Transport-3 Optional frame compression

* Improve Transport: Optional frame compression.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 384 - Transport-4 Max concurrent streams per conn

* Improve Transport: Max concurrent streams per conn.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 385 - Transport-5 Idle connection timeout

* Improve Transport: Idle connection timeout.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 386 - Transport-6 Keepalive ping frames

* Improve Transport: Keepalive ping frames.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 387 - Transport-7 Write deadline per frame

* Improve Transport: Write deadline per frame.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 388 - Transport-8 Read deadline per frame

* Improve Transport: Read deadline per frame.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 389 - Transport-9 Connection multiplexer

* Improve Transport: Connection multiplexer.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 390 - Transport-10 UDP/QUIC transport experiment

* Improve Transport: UDP/QUIC transport experiment.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 391 - Transport-11 WebRTC datachannel spike

* Improve Transport: WebRTC datachannel spike.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 392 - Transport-12 TLS wrapping for TCP

* Improve Transport: TLS wrapping for TCP.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 393 - Transport-13 Noise protocol handshake

* Improve Transport: Noise protocol handshake.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 394 - Transport-14 Certificate pinning

* Improve Transport: Certificate pinning.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 395 - Transport-15 Ban list for abusive peers

* Improve Transport: Ban list for abusive peers.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 396 - Transport-16 Rate limit messages per peer

* Improve Transport: Rate limit messages per peer.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 397 - Transport-17 Fuzz LengthPrefixedDecoder in CI

* Improve Transport: Fuzz LengthPrefixedDecoder in CI.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 398 - Transport-18 Property tests for frame round-trip

* Improve Transport: Property tests for frame round-trip.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 399 - Transport-19 Simulate partial reads

* Improve Transport: Simulate partial reads.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 400 - Transport-20 Simulate mid-stream disconnect

* Improve Transport: Simulate mid-stream disconnect.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 401 - Transport-21 Zero-copy buffer paths

* Improve Transport: Zero-copy buffer paths.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 402 - Transport-22 Avoid gob global register side effects

* Improve Transport: Avoid gob global register side effects.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 403 - Transport-23 Transport fake for unit tests

* Improve Transport: Transport fake for unit tests.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 404 - Transport-24 Dial timeout configurable

* Improve Transport: Dial timeout configurable.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 405 - Transport-25 Listen on multiple addresses

* Improve Transport: Listen on multiple addresses.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 406 - Transport-26 SO_REUSEPORT support

* Improve Transport: SO_REUSEPORT support.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 407 - Transport-27 Export pprof on debug port

* Improve Transport: Export pprof on debug port.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 408 - Transport-28 Connection churn benchmarks

* Improve Transport: Connection churn benchmarks.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 409 - Transport-29 Latency histogram metrics

* Improve Transport: Latency histogram metrics.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 410 - Transport-30 Packet loss chaos tests

* Improve Transport: Packet loss chaos tests.
* Touchpoints: `Transport`
* Acceptance: Change is implemented, tested, and documented.

---

## Theme: Code-derived extras

## Milestone 411 - REST file list from store index

* Code-derived improvement: REST file list from store index.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 412 - REST get-by-key streaming

* Code-derived improvement: REST get-by-key streaming.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 413 - REST upload streaming

* Code-derived improvement: REST upload streaming.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 414 - REST delete with peer notify

* Code-derived improvement: REST delete with peer notify.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 415 - REST metadata update persistence

* Code-derived improvement: REST metadata update persistence.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 416 - Peer list from health manager

* Code-derived improvement: Peer list from health manager.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 417 - Peer get with health detail

* Code-derived improvement: Peer get with health detail.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 418 - Peer add via DialPeer

* Code-derived improvement: Peer add via DialPeer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 419 - Peer remove with connection close

* Code-derived improvement: Peer remove with connection close.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 420 - System info version/uptime/runtime

* Code-derived improvement: System info version/uptime/runtime.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 421 - System metrics store/get counters

* Code-derived improvement: System metrics store/get counters.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 422 - System health disk + peer quorum

* Code-derived improvement: System health disk + peer quorum.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 423 - GraphQL File type fields complete

* Code-derived improvement: GraphQL File type fields complete.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 424 - GraphQL Node type fields complete

* Code-derived improvement: GraphQL Node type fields complete.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 425 - GraphQL PeerConnection type complete

* Code-derived improvement: GraphQL PeerConnection type complete.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 426 - GraphQL Mutation storeFile

* Code-derived improvement: GraphQL Mutation storeFile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 427 - GraphQL Mutation deleteFile

* Code-derived improvement: GraphQL Mutation deleteFile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 428 - GraphQL Mutation updateMetadata

* Code-derived improvement: GraphQL Mutation updateMetadata.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 429 - GraphQL Mutation addPeer

* Code-derived improvement: GraphQL Mutation addPeer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 430 - GraphQL Mutation removePeer

* Code-derived improvement: GraphQL Mutation removePeer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 431 - GraphQL Mutation updateConfig

* Code-derived improvement: GraphQL Mutation updateConfig.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 432 - GraphQL Query files pagination

* Code-derived improvement: GraphQL Query files pagination.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 433 - GraphQL Query peerNetwork graph

* Code-derived improvement: GraphQL Query peerNetwork graph.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 434 - GraphQL Query performanceStats

* Code-derived improvement: GraphQL Query performanceStats.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 435 - GraphQL Query storageStats

* Code-derived improvement: GraphQL Query storageStats.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 436 - GraphQL Query health

* Code-derived improvement: GraphQL Query health.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 437 - gRPC HealthCheck real status

* Code-derived improvement: gRPC HealthCheck real status.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 438 - gRPC ListFiles real data

* Code-derived improvement: gRPC ListFiles real data.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 439 - gRPC GetFile stream

* Code-derived improvement: gRPC GetFile stream.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 440 - gRPC DeleteFile

* Code-derived improvement: gRPC DeleteFile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 441 - gRPC ListPeers

* Code-derived improvement: gRPC ListPeers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 442 - gRPC AddPeer DialPeer

* Code-derived improvement: gRPC AddPeer DialPeer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 443 - gRPC RemovePeer

* Code-derived improvement: gRPC RemovePeer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 444 - gRPC PeerHealth

* Code-derived improvement: gRPC PeerHealth.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 445 - gRPC SystemInfo

* Code-derived improvement: gRPC SystemInfo.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 446 - gRPC Metrics

* Code-derived improvement: gRPC Metrics.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 447 - Monitoring alert silence persistence

* Code-derived improvement: Monitoring alert silence persistence.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 448 - Monitoring alert resolve persistence

* Code-derived improvement: Monitoring alert resolve persistence.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 449 - Monitoring recommendations from SLOs

* Code-derived improvement: Monitoring recommendations from SLOs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 450 - Analytics popularity from store events

* Code-derived improvement: Analytics popularity from store events.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 451 - Analytics privacy review

* Code-derived improvement: Analytics privacy review.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 452 - WebSocket /ws/health contract

* Code-derived improvement: WebSocket /ws/health contract.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 453 - WebSocket /ws/metrics contract

* Code-derived improvement: WebSocket /ws/metrics contract.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 454 - CLI analytics without sleep stubs

* Code-derived improvement: CLI analytics without sleep stubs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 455 - CLI network CDN simulation removal

* Code-derived improvement: CLI network CDN simulation removal.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 456 - Terminal resize handling

* Code-derived improvement: Terminal resize handling.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 457 - Completion for storage keys

* Code-derived improvement: Completion for storage keys.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 458 - Federation gateway auth propagation

* Code-derived improvement: Federation gateway auth propagation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 459 - Cache invalidation on Store

* Code-derived improvement: Cache invalidation on Store.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 460 - Query cache TTL tuning

* Code-derived improvement: Query cache TTL tuning.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 461 - Stitching remote schema refresh

* Code-derived improvement: Stitching remote schema refresh.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 462 - Balancer health checker vs DialPeer

* Code-derived improvement: Balancer health checker vs DialPeer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 463 - Static discovery watch docs

* Code-derived improvement: Static discovery watch docs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 464 - Optional Consul discovery module

* Code-derived improvement: Optional Consul discovery module.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 465 - Optional etcd discovery module

* Code-derived improvement: Optional etcd discovery module.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 466 - Optional K8s discovery module

* Code-derived improvement: Optional K8s discovery module.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 467 - CoAP observe integration test

* Code-derived improvement: CoAP observe integration test.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 468 - MQTT retained message persistence

* Code-derived improvement: MQTT retained message persistence.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 469 - SSE last-event-id resume

* Code-derived improvement: SSE last-event-id resume.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 470 - Translation analytics accuracy tests

* Code-derived improvement: Translation analytics accuracy tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 471 - Mock server scenario DSL docs

* Code-derived improvement: Mock server scenario DSL docs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 472 - Contract test /health JSON schema

* Code-derived improvement: Contract test /health JSON schema.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 473 - k6 script for Store/Get path

* Code-derived improvement: k6 script for Store/Get path.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 474 - Security test oversized frame rejected

* Code-derived improvement: Security test oversized frame rejected.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 475 - Security test bad HMAC rejected

* Code-derived improvement: Security test bad HMAC rejected.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 476 - Security test path traversal rejected

* Code-derived improvement: Security test path traversal rejected.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 477 - Fuzz CASPathTransformFunc

* Code-derived improvement: Fuzz CASPathTransformFunc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 478 - Fuzz handshake serialize/deserialize

* Code-derived improvement: Fuzz handshake serialize/deserialize.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 479 - Unit dialPeer success path

* Code-derived improvement: Unit dialPeer success path.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 480 - Unit replicateToPeers no peers

* Code-derived improvement: Unit replicateToPeers no peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 481 - Unit Get timeout when no peers

* Code-derived improvement: Unit Get timeout when no peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 482 - Unit OnStream invalid key length

* Code-derived improvement: Unit OnStream invalid key length.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 483 - Unit OnStream size mismatch

* Code-derived improvement: Unit OnStream size mismatch.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 484 - Integration store A get B

* Code-derived improvement: Integration store A get B.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 485 - Integration store A get C via B

* Code-derived improvement: Integration store A get C via B.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 486 - Integration peer reconnect after kill

* Code-derived improvement: Integration peer reconnect after kill.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 487 - Integration mismatched cluster key

* Code-derived improvement: Integration mismatched cluster key.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 488 - Chaos kill peer mid-stream

* Code-derived improvement: Chaos kill peer mid-stream.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 489 - Chaos slow disk

* Code-derived improvement: Chaos slow disk.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 490 - Chaos clock skew beyond handshake window

* Code-derived improvement: Chaos clock skew beyond handshake window.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 491 - Docs update milestones.md after M1-2

* Code-derived improvement: Docs update milestones.md after M1-2.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 492 - Docs architecture sequence diagrams

* Code-derived improvement: Docs architecture sequence diagrams.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 493 - Release v0.1.0 MVP criteria

* Code-derived improvement: Release v0.1.0 MVP criteria.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 494 - Release ghcr publish

* Code-derived improvement: Release ghcr publish.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 495 - License header check

* Code-derived improvement: License header check.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 496 - SPDX identifiers

* Code-derived improvement: SPDX identifiers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 497 - Third-party notices file

* Code-derived improvement: Third-party notices file.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 498 - Feature flag framework

* Code-derived improvement: Feature flag framework.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 499 - Remote config provider

* Code-derived improvement: Remote config provider.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 500 - Canary peer percentage rollout

* Code-derived improvement: Canary peer percentage rollout.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 501 - Shadow reads verification

* Code-derived improvement: Shadow reads verification.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 502 - Dual-write migration helper

* Code-derived improvement: Dual-write migration helper.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 503 - Read repair on checksum fail

* Code-derived improvement: Read repair on checksum fail.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 504 - Anti-entropy merkle sync

* Code-derived improvement: Anti-entropy merkle sync.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 505 - Rendezvous hashing ownership

* Code-derived improvement: Rendezvous hashing ownership.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 506 - Consistent hashing ring

* Code-derived improvement: Consistent hashing ring.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 507 - Optional raft spike

* Code-derived improvement: Optional raft spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 508 - CRDT for metadata only

* Code-derived improvement: CRDT for metadata only.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 509 - Vector clocks for conflicts

* Code-derived improvement: Vector clocks for conflicts.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 510 - Last-writer-wins policy doc

* Code-derived improvement: Last-writer-wins policy doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 511 - Manual conflict resolve API

* Code-derived improvement: Manual conflict resolve API.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 512 - Snapshot + log shipping

* Code-derived improvement: Snapshot + log shipping.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 513 - Point-in-time restore

* Code-derived improvement: Point-in-time restore.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 514 - Encryption key escrow

* Code-derived improvement: Encryption key escrow.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 515 - Customer-managed keys

* Code-derived improvement: Customer-managed keys.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 516 - KMS integration AWS/GCP/Azure

* Code-derived improvement: KMS integration AWS/GCP/Azure.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 517 - HSM integration spike

* Code-derived improvement: HSM integration spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 518 - FIPS build mode investigation

* Code-derived improvement: FIPS build mode investigation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 519 - Data residency tags

* Code-derived improvement: Data residency tags.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 520 - Retention policy engine

* Code-derived improvement: Retention policy engine.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 521 - Legal hold flag

* Code-derived improvement: Legal hold flag.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 522 - GDPR delete cascade

* Code-derived improvement: GDPR delete cascade.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 523 - Export user data package

* Code-derived improvement: Export user data package.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 524 - Access review reports

* Code-derived improvement: Access review reports.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 525 - Break-glass admin procedure

* Code-derived improvement: Break-glass admin procedure.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 526 - On-call runbooks pack

* Code-derived improvement: On-call runbooks pack.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 527 - Incident severity matrix

* Code-derived improvement: Incident severity matrix.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 528 - Postmortem template

* Code-derived improvement: Postmortem template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 529 - Capacity planning model

* Code-derived improvement: Capacity planning model.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 530 - Load shedding strategy

* Code-derived improvement: Load shedding strategy.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 531 - Graceful degradation mode

* Code-derived improvement: Graceful degradation mode.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 532 - Read-only fallback

* Code-derived improvement: Read-only fallback.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 533 - Cache-only mode

* Code-derived improvement: Cache-only mode.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 534 - Emergency peer quarantine

* Code-derived improvement: Emergency peer quarantine.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 535 - ACME cert renew

* Code-derived improvement: ACME cert renew.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 536 - Internal CA for node mesh

* Code-derived improvement: Internal CA for node mesh.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 537 - SPIFFE/SPIRE spike

* Code-derived improvement: SPIFFE/SPIRE spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 538 - Zero-trust network notes

* Code-derived improvement: Zero-trust network notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 539 - Supply chain SLSA L2

* Code-derived improvement: Supply chain SLSA L2.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 540 - SLSA L3 provenance

* Code-derived improvement: SLSA L3 provenance.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 541 - Cosign image signing

* Code-derived improvement: Cosign image signing.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 542 - OPA policy for APIs

* Code-derived improvement: OPA policy for APIs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 543 - Admission controller examples

* Code-derived improvement: Admission controller examples.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 544 - Gateway API examples

* Code-derived improvement: Gateway API examples.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 545 - Service mesh notes

* Code-derived improvement: Service mesh notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 546 - eBPF observability spike

* Code-derived improvement: eBPF observability spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 547 - WASM plugin runtime spike

* Code-derived improvement: WASM plugin runtime spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 548 - SQLite metadata backend

* Code-derived improvement: SQLite metadata backend.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 549 - Postgres metadata backend

* Code-derived improvement: Postgres metadata backend.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 550 - S3 CAS backend plugin

* Code-derived improvement: S3 CAS backend plugin.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 551 - MinIO integration test

* Code-derived improvement: MinIO integration test.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 552 - Azure blob plugin

* Code-derived improvement: Azure blob plugin.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 553 - GCS plugin

* Code-derived improvement: GCS plugin.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 554 - Restic integration example

* Code-derived improvement: Restic integration example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 555 - Velero for K8s volumes

* Code-derived improvement: Velero for K8s volumes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 556 - Multi-tenant isolation tests

* Code-derived improvement: Multi-tenant isolation tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 557 - Noisy neighbor limits

* Code-derived improvement: Noisy neighbor limits.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 558 - Fair queuing per tenant

* Code-derived improvement: Fair queuing per tenant.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 559 - Billing meters bytes stored/transferred

* Code-derived improvement: Billing meters bytes stored/transferred.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 560 - Usage export CSV

* Code-derived improvement: Usage export CSV.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 561 - Admin UI spike read-only

* Code-derived improvement: Admin UI spike read-only.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 562 - Peer map visualization

* Code-derived improvement: Peer map visualization.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 563 - File browser UI spike

* Code-derived improvement: File browser UI spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 564 - Slack alert sink

* Code-derived improvement: Slack alert sink.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 565 - PagerDuty integration

* Code-derived improvement: PagerDuty integration.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 566 - Status page generator

* Code-derived improvement: Status page generator.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 567 - Public uptime SLA page

* Code-derived improvement: Public uptime SLA page.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 568 - Pentest remediation tracker

* Code-derived improvement: Pentest remediation tracker.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 569 - Vulnerability disclosure program

* Code-derived improvement: Vulnerability disclosure program.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 570 - Tabletop exercise materials

* Code-derived improvement: Tabletop exercise materials.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 571 - Game day region loss

* Code-derived improvement: Game day region loss.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 572 - Game day key compromise

* Code-derived improvement: Game day key compromise.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 573 - Game day thundering herd Get

* Code-derived improvement: Game day thundering herd Get.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 574 - Benchmark vs IPFS baseline

* Code-derived improvement: Benchmark vs IPFS baseline.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 575 - Benchmark vs MinIO baseline

* Code-derived improvement: Benchmark vs MinIO baseline.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 576 - RFC wire protocol v2

* Code-derived improvement: RFC wire protocol v2.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 577 - RFC auth v2

* Code-derived improvement: RFC auth v2.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 578 - Community RFC process

* Code-derived improvement: Community RFC process.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 579 - Weekly triage bot

* Code-derived improvement: Weekly triage bot.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 580 - Stale issue automation

* Code-derived improvement: Stale issue automation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 581 - Good first issue pack

* Code-derived improvement: Good first issue pack.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 582 - Mentor checklist for contributors

* Code-derived improvement: Mentor checklist for contributors.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 583 - Internal Go style guide

* Code-derived improvement: Internal Go style guide.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 584 - Error wrapping conventions enforced

* Code-derived improvement: Error wrapping conventions enforced.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 585 - Context first-arg lint

* Code-derived improvement: Context first-arg lint.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 586 - No panic in library lint

* Code-derived improvement: No panic in library lint.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 587 - slog key consistency lint

* Code-derived improvement: slog key consistency lint.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 588 - Import alias standards

* Code-derived improvement: Import alias standards.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 589 - Package dependency direction rules

* Code-derived improvement: Package dependency direction rules.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 590 - Forbid cmd/internal import cycles

* Code-derived improvement: Forbid cmd/internal import cycles.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 591 - API compatibility checker

* Code-derived improvement: API compatibility checker.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 592 - Proto compatibility checker

* Code-derived improvement: Proto compatibility checker.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 593 - GraphQL schema compatibility checker

* Code-derived improvement: GraphQL schema compatibility checker.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 594 - OpenAPI compatibility checker

* Code-derived improvement: OpenAPI compatibility checker.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 595 - Semantic release pipeline

* Code-derived improvement: Semantic release pipeline.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 596 - Changelog fragments per PR

* Code-derived improvement: Changelog fragments per PR.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 597 - Docs preview per PR

* Code-derived improvement: Docs preview per PR.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 598 - Ephemeral preview environments

* Code-derived improvement: Ephemeral preview environments.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 599 - Feature branch compose stacks

* Code-derived improvement: Feature branch compose stacks.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 600 - kind/k3d quickstart

* Code-derived improvement: kind/k3d quickstart.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 601 - Devcontainer definition

* Code-derived improvement: Devcontainer definition.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 602 - Codespaces config

* Code-derived improvement: Codespaces config.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 603 - Editorconfig enforce

* Code-derived improvement: Editorconfig enforce.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 604 - Pre-commit golangci required

* Code-derived improvement: Pre-commit golangci required.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 605 - Pre-commit gofmt required

* Code-derived improvement: Pre-commit gofmt required.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 606 - Makefile/Taskfile parity

* Code-derived improvement: Makefile/Taskfile parity.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 607 - PowerShell ops module

* Code-derived improvement: PowerShell ops module.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 608 - Bash completion install script

* Code-derived improvement: Bash completion install script.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 609 - Homebrew formula spike

* Code-derived improvement: Homebrew formula spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 610 - Scoop package spike

* Code-derived improvement: Scoop package spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 611 - Apt package spike

* Code-derived improvement: Apt package spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 612 - Nix flake spike

* Code-derived improvement: Nix flake spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 613 - Winget manifest

* Code-derived improvement: Winget manifest.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 614 - Chocolatey package

* Code-derived improvement: Chocolatey package.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 615 - Distroless runtime image

* Code-derived improvement: Distroless runtime image.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 616 - Chainguard/wolfi image option

* Code-derived improvement: Chainguard/wolfi image option.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 617 - Image CVE gate in CI

* Code-derived improvement: Image CVE gate in CI.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 618 - Base image refresh automation

* Code-derived improvement: Base image refresh automation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 619 - Multi-arch amd64/arm64 images

* Code-derived improvement: Multi-arch amd64/arm64 images.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 620 - arm64 performance notes

* Code-derived improvement: arm64 performance notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 621 - Raspberry Pi edge profile

* Code-derived improvement: Raspberry Pi edge profile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 622 - Low-memory 512MB profile

* Code-derived improvement: Low-memory 512MB profile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 623 - High-throughput profile

* Code-derived improvement: High-throughput profile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 624 - Air-gapped install pack

* Code-derived improvement: Air-gapped install pack.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 625 - Offline module proxy docs

* Code-derived improvement: Offline module proxy docs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 626 - Corporate MITM TLS docs

* Code-derived improvement: Corporate MITM TLS docs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 627 - Proxy env support verified

* Code-derived improvement: Proxy env support verified.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 628 - IPv6-only deployment test

* Code-derived improvement: IPv6-only deployment test.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 629 - Dual-stack deployment test

* Code-derived improvement: Dual-stack deployment test.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 630 - Split-horizon DNS notes

* Code-derived improvement: Split-horizon DNS notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 631 - Corporate firewall ports doc

* Code-derived improvement: Corporate firewall ports doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 632 - SELinux labeling notes

* Code-derived improvement: SELinux labeling notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 633 - AppArmor profile sample

* Code-derived improvement: AppArmor profile sample.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 634 - ulimit recommendations

* Code-derived improvement: ulimit recommendations.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 635 - fd exhaustion tests

* Code-derived improvement: fd exhaustion tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 636 - goroutine leak tests

* Code-derived improvement: goroutine leak tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 637 - memory leak soak 24h

* Code-derived improvement: memory leak soak 24h.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 638 - disk fill behavior test

* Code-derived improvement: disk fill behavior test.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 639 - clock jump behavior test

* Code-derived improvement: clock jump behavior test.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 640 - NTP requirements doc

* Code-derived improvement: NTP requirements doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 641 - Milestone tracker sync automation

* Code-derived improvement: Milestone tracker sync automation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 642 - Streaming decrypt pipe for Get

* Code-derived improvement: Streaming decrypt pipe for Get.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 643 - Shared cluster key required in docker-compose

* Code-derived improvement: Shared cluster key required in docker-compose.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 644 - Document PEERVAULT_CLUSTER_KEY in compose samples

* Code-derived improvement: Document PEERVAULT_CLUSTER_KEY in compose samples.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 645 - DialPeer unit test with fake listener

* Code-derived improvement: DialPeer unit test with fake listener.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 646 - OnStream skip when Has already true

* Code-derived improvement: OnStream skip when Has already true.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 647 - Store announcement without duplicate encrypt

* Code-derived improvement: Store announcement without duplicate encrypt.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 648 - Get pending channel leak on timeout fix verification

* Code-derived improvement: Get pending channel leak on timeout fix verification.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 649 - Normalize RemoteAddr string keys

* Code-derived improvement: Normalize RemoteAddr string keys.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 650 - Health manager uses DialPeer reconnect e2e

* Code-derived improvement: Health manager uses DialPeer reconnect e2e.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 651 - Resource manager stream slots during replicate

* Code-derived improvement: Resource manager stream slots during replicate.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 652 - FrameWriter endianness consistency audit

* Code-derived improvement: FrameWriter endianness consistency audit.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 653 - LittleEndian key header vs BigEndian frame header docs

* Code-derived improvement: LittleEndian key header vs BigEndian frame header docs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 654 - Max key length DoS guard tests

* Code-derived improvement: Max key length DoS guard tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 655 - Negative size stream rejected tests

* Code-derived improvement: Negative size stream rejected tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 656 - fileserver package race test in CI

* Code-derived improvement: fileserver package race test in CI.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 657 - cmd/peervault demo waits for bootstrap without fixed sleep

* Code-derived improvement: cmd/peervault demo waits for bootstrap without fixed sleep.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 658 - Replace remaining time.Sleep in e2e with readiness probe

* Code-derived improvement: Replace remaining time.Sleep in e2e with readiness probe.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 659 - Port conflict retry in tests

* Code-derived improvement: Port conflict retry in tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 660 - Unique storage roots per test parallel safe

* Code-derived improvement: Unique storage roots per test parallel safe.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 661 - Cleanup storage roots in test teardown always

* Code-derived improvement: Cleanup storage roots in test teardown always.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 662 - Windows Defender exclusion note in getting-started only

* Code-derived improvement: Windows Defender exclusion note in getting-started only.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 663 - SECURITY.md link from milestones

* Code-derived improvement: SECURITY.md link from milestones.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 664 - CONTAINERIZATION.md link Helm absence note

* Code-derived improvement: CONTAINERIZATION.md link Helm absence note.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 665 - Update architecture.md SHA-256 now accurate

* Code-derived improvement: Update architecture.md SHA-256 now accurate.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 666 - Mark M1.3 CAS complete in ROADMAP

* Code-derived improvement: Mark M1.3 CAS complete in ROADMAP.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 667 - Mark M1 dial/replication progress in ROADMAP

* Code-derived improvement: Mark M1 dial/replication progress in ROADMAP.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 668 - Add fileserver README

* Code-derived improvement: Add fileserver README.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 669 - Add transport README

* Code-derived improvement: Add transport README.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 670 - Public metrics redaction guide

* Code-derived improvement: Public metrics redaction guide.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 671 - Admin vs public endpoint split

* Code-derived improvement: Admin vs public endpoint split.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 672 - Readiness excludes bootstrap-incomplete optional

* Code-derived improvement: Readiness excludes bootstrap-incomplete optional.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 673 - Liveness always process-up

* Code-derived improvement: Liveness always process-up.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 674 - Startup probe for slow disk

* Code-derived improvement: Startup probe for slow disk.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 675 - Horizontal scaling peer list gossip

* Code-derived improvement: Horizontal scaling peer list gossip.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 676 - Epidemic anti-entropy design

* Code-derived improvement: Epidemic anti-entropy design.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 677 - Merkle tree per namespace

* Code-derived improvement: Merkle tree per namespace.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 678 - Chunk manifest object format

* Code-derived improvement: Chunk manifest object format.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 679 - Erasure coding research spike

* Code-derived improvement: Erasure coding research spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 680 - Reed-Solomon optional module

* Code-derived improvement: Reed-Solomon optional module.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 681 - Parity peer role

* Code-derived improvement: Parity peer role.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 682 - Witness peer role

* Code-derived improvement: Witness peer role.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 683 - Observer/read-only peer role

* Code-derived improvement: Observer/read-only peer role.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 684 - Bootstrap peer privileged allowlist

* Code-derived improvement: Bootstrap peer privileged allowlist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 685 - Operator signed peer allowlist

* Code-derived improvement: Operator signed peer allowlist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 686 - Revocation list distribution

* Code-derived improvement: Revocation list distribution.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 687 - CRL for node certs

* Code-derived improvement: CRL for node certs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 688 - OCSP stapling for API TLS

* Code-derived improvement: OCSP stapling for API TLS.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 689 - HSTS on all HTTP APIs

* Code-derived improvement: HSTS on all HTTP APIs.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 690 - CSP for playground/swagger

* Code-derived improvement: CSP for playground/swagger.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 691 - Secure cookies if sessions added

* Code-derived improvement: Secure cookies if sessions added.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 692 - SameSite cookie policy

* Code-derived improvement: SameSite cookie policy.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 693 - CSRF for browser cookie auth

* Code-derived improvement: CSRF for browser cookie auth.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 694 - Turnstile/captcha for public upload rate

* Code-derived improvement: Turnstile/captcha for public upload rate.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 695 - Proof-of-work optional anti-spam

* Code-derived improvement: Proof-of-work optional anti-spam.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 696 - Token bucket per IP

* Code-derived improvement: Token bucket per IP.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 697 - Token bucket per API key

* Code-derived improvement: Token bucket per API key.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 698 - Geo-IP allow/deny lists

* Code-derived improvement: Geo-IP allow/deny lists.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 699 - Tor exit node policy doc

* Code-derived improvement: Tor exit node policy doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 700 - Sanctioned region block list optional

* Code-derived improvement: Sanctioned region block list optional.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 701 - Export control classification

* Code-derived improvement: Export control classification.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 702 - Crypto export notice

* Code-derived improvement: Crypto export notice.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 703 - Dual-use statement

* Code-derived improvement: Dual-use statement.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 704 - Responsible AI statement for ML module

* Code-derived improvement: Responsible AI statement for ML module.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 705 - Model card for classifier

* Code-derived improvement: Model card for classifier.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 706 - Dataset provenance for ML

* Code-derived improvement: Dataset provenance for ML.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 707 - Bias evaluation for ML labels

* Code-derived improvement: Bias evaluation for ML labels.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 708 - Human review hook for ML

* Code-derived improvement: Human review hook for ML.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 709 - Edge offline sync conflict UI notes

* Code-derived improvement: Edge offline sync conflict UI notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 710 - IoT device certificate enrollment

* Code-derived improvement: IoT device certificate enrollment.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 711 - CoAP block-wise transfer

* Code-derived improvement: CoAP block-wise transfer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 712 - MQTT QoS 2 exactly-once path

* Code-derived improvement: MQTT QoS 2 exactly-once path.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 713 - AMQP bridge spike

* Code-derived improvement: AMQP bridge spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 714 - NATS bridge spike

* Code-derived improvement: NATS bridge spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 715 - Kafka connect spike

* Code-derived improvement: Kafka connect spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 716 - Pulsar bridge spike

* Code-derived improvement: Pulsar bridge spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 717 - Redis streams bridge spike

* Code-derived improvement: Redis streams bridge spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 718 - Webhook retry with backoff

* Code-derived improvement: Webhook retry with backoff.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 719 - Webhook DLQ

* Code-derived improvement: Webhook DLQ.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 720 - Outbox pattern for events

* Code-derived improvement: Outbox pattern for events.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 721 - Inbox pattern for inbound

* Code-derived improvement: Inbox pattern for inbound.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 722 - Idempotent consumer keys

* Code-derived improvement: Idempotent consumer keys.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 723 - Exactly-once caveats doc

* Code-derived improvement: Exactly-once caveats doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 724 - At-least-once default doc

* Code-derived improvement: At-least-once default doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 725 - Ordering guarantees doc

* Code-derived improvement: Ordering guarantees doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 726 - Clock sync requirement SLO

* Code-derived improvement: Clock sync requirement SLO.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 727 - Logical clock for events

* Code-derived improvement: Logical clock for events.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 728 - Hybrid logical clock spike

* Code-derived improvement: Hybrid logical clock spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 729 - Serialize Store per key globally

* Code-derived improvement: Serialize Store per key globally.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 730 - Serializable isolation notes

* Code-derived improvement: Serializable isolation notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 731 - Snapshot isolation for metadata DB

* Code-derived improvement: Snapshot isolation for metadata DB.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 732 - Read committed for indexes

* Code-derived improvement: Read committed for indexes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 733 - Dirty read never for ciphertext

* Code-derived improvement: Dirty read never for ciphertext.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 734 - Constant-time Has for missing keys timing

* Code-derived improvement: Constant-time Has for missing keys timing.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 735 - Padding oracle review on GCM misuse

* Code-derived improvement: Padding oracle review on GCM misuse.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 736 - Compress-then-encrypt warning in ENCRYPTION.md

* Code-derived improvement: Compress-then-encrypt warning in ENCRYPTION.md.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 737 - Key stretching if password-derived keys added

* Code-derived improvement: Key stretching if password-derived keys added.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 738 - Argon2id for password KDF if needed

* Code-derived improvement: Argon2id for password KDF if needed.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 739 - Passphrase UX never for cluster key

* Code-derived improvement: Passphrase UX never for cluster key.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 740 - QR code cluster key exchange spike

* Code-derived improvement: QR code cluster key exchange spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 741 - Shamir secret sharing for cluster key

* Code-derived improvement: Shamir secret sharing for cluster key.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 742 - Threshold decrypt research

* Code-derived improvement: Threshold decrypt research.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 743 - Multi-party compute out of scope note

* Code-derived improvement: Multi-party compute out of scope note.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 744 - Formal verification spike for framing

* Code-derived improvement: Formal verification spike for framing.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 745 - TLA+ spec for replication

* Code-derived improvement: TLA+ spec for replication.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 746 - Alloy model for peer membership

* Code-derived improvement: Alloy model for peer membership.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 747 - Property-based tests with gopter/rapid

* Code-derived improvement: Property-based tests with gopter/rapid.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 748 - Hypothesis-style stateful tests

* Code-derived improvement: Hypothesis-style stateful tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 749 - Linearizability checker for Get/Store

* Code-derived improvement: Linearizability checker for Get/Store.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 750 - Jepsen-style harness spike

* Code-derived improvement: Jepsen-style harness spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 751 - Fault injection with toxiproxy

* Code-derived improvement: Fault injection with toxiproxy.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 752 - Chaos mesh examples

* Code-derived improvement: Chaos mesh examples.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 753 - Litmus chaos experiments

* Code-derived improvement: Litmus chaos experiments.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 754 - Gameday calendar quarterly

* Code-derived improvement: Gameday calendar quarterly.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 755 - Error budget policy

* Code-derived improvement: Error budget policy.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 756 - Pager rotation

* Code-derived improvement: Pager rotation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 757 - Runbook automation buttons

* Code-derived improvement: Runbook automation buttons.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 758 - ChatOps deploy freeze

* Code-derived improvement: ChatOps deploy freeze.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 759 - Feature freeze checklist

* Code-derived improvement: Feature freeze checklist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 760 - Release candidate soak

* Code-derived improvement: Release candidate soak.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 761 - Canary 1 percent peers

* Code-derived improvement: Canary 1 percent peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 762 - Canary 10 percent peers

* Code-derived improvement: Canary 10 percent peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 763 - Canary 50 percent peers

* Code-derived improvement: Canary 50 percent peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 764 - Auto rollback on error spike

* Code-derived improvement: Auto rollback on error spike.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 765 - Manual rollback runbook

* Code-derived improvement: Manual rollback runbook.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 766 - Blue/green for API pods

* Code-derived improvement: Blue/green for API pods.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 767 - Shadow traffic to new protocol

* Code-derived improvement: Shadow traffic to new protocol.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 768 - Protocol v2 dual-stack period

* Code-derived improvement: Protocol v2 dual-stack period.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 769 - Deprecate protocol v1 timeline

* Code-derived improvement: Deprecate protocol v1 timeline.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 770 - Compat matrix published

* Code-derived improvement: Compat matrix published.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 771 - Client SDK support windows

* Code-derived improvement: Client SDK support windows.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 772 - Server support windows

* Code-derived improvement: Server support windows.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 773 - LTS branch policy

* Code-derived improvement: LTS branch policy.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 774 - Security patch backport policy

* Code-derived improvement: Security patch backport policy.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 775 - CVE response SLA

* Code-derived improvement: CVE response SLA.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 776 - Embargo process

* Code-derived improvement: Embargo process.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 777 - Credit for reporters

* Code-derived improvement: Credit for reporters.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 778 - Hall of fame page

* Code-derived improvement: Hall of fame page.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 779 - OpenSSF scorecard target A

* Code-derived improvement: OpenSSF scorecard target A.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 780 - Best practices badge

* Code-derived improvement: Best practices badge.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 781 - OpenSSF scorecard CI

* Code-derived improvement: OpenSSF scorecard CI.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 782 - Fuzz introspector

* Code-derived improvement: Fuzz introspector.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 783 - OSS-Fuzz onboarding

* Code-derived improvement: OSS-Fuzz onboarding.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 784 - Coverity/semgrep tune rules

* Code-derived improvement: Coverity/semgrep tune rules.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 785 - Custom semgrep rules for gob

* Code-derived improvement: Custom semgrep rules for gob.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 786 - Custom semgrep for unsafe os.Open

* Code-derived improvement: Custom semgrep for unsafe os.Open.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 787 - Ban ioutil deprecated

* Code-derived improvement: Ban ioutil deprecated.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 788 - Ban interface{} in new code prefer any

* Code-derived improvement: Ban interface{} in new code prefer any.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 789 - Prefer errors.Join

* Code-derived improvement: Prefer errors.Join.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 790 - Prefer slog over log

* Code-derived improvement: Prefer slog over log.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 791 - Prefer context.WithTimeout helpers

* Code-derived improvement: Prefer context.WithTimeout helpers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 792 - Standardize port flags across cmd

* Code-derived improvement: Standardize port flags across cmd.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 793 - Standardize log-level flags across cmd

* Code-derived improvement: Standardize log-level flags across cmd.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 794 - Standardize storage-root flags across cmd

* Code-derived improvement: Standardize storage-root flags across cmd.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 795 - Unify makeServer helpers into internal/testutil

* Code-derived improvement: Unify makeServer helpers into internal/testutil.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 796 - Shared KeyManager fixture package

* Code-derived improvement: Shared KeyManager fixture package.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 797 - Golden files for frames

* Code-derived improvement: Golden files for frames.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 798 - Record/replay transport tests

* Code-derived improvement: Record/replay transport tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 799 - Deterministic GenerateID for tests option

* Code-derived improvement: Deterministic GenerateID for tests option.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 800 - Leaktest for goroutines

* Code-derived improvement: Leaktest for goroutines.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 801 - goleak in TestMain

* Code-derived improvement: goleak in TestMain.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 802 - testify require consistently

* Code-derived improvement: testify require consistently.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 803 - Table-driven tests style guide

* Code-derived improvement: Table-driven tests style guide.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 804 - Fuzz time budgets documented

* Code-derived improvement: Fuzz time budgets documented.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 805 - Benchmark baselines checked in

* Code-derived improvement: Benchmark baselines checked in.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 806 - perf regression CI threshold

* Code-derived improvement: perf regression CI threshold.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 807 - CPU profile on fail

* Code-derived improvement: CPU profile on fail.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 808 - Heap profile on fail

* Code-derived improvement: Heap profile on fail.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 809 - Block profile optional

* Code-derived improvement: Block profile optional.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 810 - Mutex profile optional

* Code-derived improvement: Mutex profile optional.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 811 - Execution trace on fail

* Code-derived improvement: Execution trace on fail.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 812 - OpenTelemetry baggage for key hash

* Code-derived improvement: OpenTelemetry baggage for key hash.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 813 - Span attributes peer addr redacted

* Code-derived improvement: Span attributes peer addr redacted.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 814 - Metric exemplars

* Code-derived improvement: Metric exemplars.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 815 - Log cardinality limits

* Code-derived improvement: Log cardinality limits.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 816 - Prometheus histogram buckets tuned

* Code-derived improvement: Prometheus histogram buckets tuned.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 817 - Exemplars for Store latency

* Code-derived improvement: Exemplars for Store latency.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 818 - Tail sampling for errors

* Code-derived improvement: Tail sampling for errors.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 819 - Head sampling for success

* Code-derived improvement: Head sampling for success.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 820 - Adaptive sampling

* Code-derived improvement: Adaptive sampling.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 821 - Remote write to Grafana Cloud example

* Code-derived improvement: Remote write to Grafana Cloud example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 822 - Loki log ship example

* Code-derived improvement: Loki log ship example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 823 - Tempo traces example

* Code-derived improvement: Tempo traces example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 824 - Pyroscope continuous profile example

* Code-derived improvement: Pyroscope continuous profile example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 825 - Parca alternative

* Code-derived improvement: Parca alternative.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 826 - Thanos long-term metrics

* Code-derived improvement: Thanos long-term metrics.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 827 - Mimir example

* Code-derived improvement: Mimir example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 828 - VictoriaMetrics example

* Code-derived improvement: VictoriaMetrics example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 829 - Grafana Alloy config

* Code-derived improvement: Grafana Alloy config.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 830 - OpenTelemetry Operator

* Code-derived improvement: OpenTelemetry Operator.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 831 - Service graph for peers

* Code-derived improvement: Service graph for peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 832 - RED dashboard template

* Code-derived improvement: RED dashboard template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 833 - USE dashboard template

* Code-derived improvement: USE dashboard template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 834 - Golden signals dashboard

* Code-derived improvement: Golden signals dashboard.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 835 - Saturation alerts

* Code-derived improvement: Saturation alerts.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 836 - Traffic alerts

* Code-derived improvement: Traffic alerts.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 837 - Errors alerts

* Code-derived improvement: Errors alerts.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 838 - Latency alerts

* Code-derived improvement: Latency alerts.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 839 - Burn-rate alerts

* Code-derived improvement: Burn-rate alerts.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 840 - Weekly ops review template

* Code-derived improvement: Weekly ops review template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 841 - Monthly capacity review

* Code-derived improvement: Monthly capacity review.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 842 - Quarterly architecture review

* Code-derived improvement: Quarterly architecture review.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 843 - Annual threat model review

* Code-derived improvement: Annual threat model review.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 844 - Dependency update party cadence

* Code-derived improvement: Dependency update party cadence.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 845 - Go toolchain upgrade playbook

* Code-derived improvement: Go toolchain upgrade playbook.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 846 - Protobuf upgrade playbook

* Code-derived improvement: Protobuf upgrade playbook.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 847 - Major dep upgrade checklist

* Code-derived improvement: Major dep upgrade checklist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 848 - Remove zombie packages

* Code-derived improvement: Remove zombie packages.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 849 - Archive unused cmd binaries

* Code-derived improvement: Archive unused cmd binaries.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 850 - Feature flag to disable ML cmd

* Code-derived improvement: Feature flag to disable ML cmd.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 851 - Feature flag to disable chain cmd

* Code-derived improvement: Feature flag to disable chain cmd.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 852 - Reduce binary size build tags

* Code-derived improvement: Reduce binary size build tags.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 853 - UPX evaluation (security tradeoff)

* Code-derived improvement: UPX evaluation (security tradeoff).
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 854 - Strip symbols in release

* Code-derived improvement: Strip symbols in release.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 855 - Buildinfo embed

* Code-derived improvement: Buildinfo embed.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 856 - vcs revision embed

* Code-derived improvement: vcs revision embed.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 857 - Dirty build detect

* Code-derived improvement: Dirty build detect.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 858 - Source map for panics

* Code-derived improvement: Source map for panics.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 859 - Sentry/GlitchTip optional

* Code-derived improvement: Sentry/GlitchTip optional.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 860 - Panic webhook

* Code-derived improvement: Panic webhook.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 861 - Watchdog process supervisor

* Code-derived improvement: Watchdog process supervisor.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 862 - Auto-restart policy compose

* Code-derived improvement: Auto-restart policy compose.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 863 - Restart budget

* Code-derived improvement: Restart budget.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 864 - Crash loop backoff

* Code-derived improvement: Crash loop backoff.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 865 - OOM kill diagnostics

* Code-derived improvement: OOM kill diagnostics.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 866 - cgroup memory limit tests

* Code-derived improvement: cgroup memory limit tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 867 - CPU quota tests

* Code-derived improvement: CPU quota tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 868 - IO throttle tests

* Code-derived improvement: IO throttle tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 869 - Network bandwidth throttle tests

* Code-derived improvement: Network bandwidth throttle tests.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 870 - Fair share between tenants

* Code-derived improvement: Fair share between tenants.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 871 - Priority queues for Get

* Code-derived improvement: Priority queues for Get.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 872 - Priority queues for Store

* Code-derived improvement: Priority queues for Store.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 873 - Admin bypass queue

* Code-derived improvement: Admin bypass queue.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 874 - Maintenance window scheduler

* Code-derived improvement: Maintenance window scheduler.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 875 - Freeze writes API

* Code-derived improvement: Freeze writes API.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 876 - Drain peer before shutdown

* Code-derived improvement: Drain peer before shutdown.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 877 - Undrain peer

* Code-derived improvement: Undrain peer.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 878 - Quiesce replication

* Code-derived improvement: Quiesce replication.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 879 - Catch-up replication after quiesce

* Code-derived improvement: Catch-up replication after quiesce.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 880 - Backfill tool

* Code-derived improvement: Backfill tool.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 881 - Rebalance tool

* Code-derived improvement: Rebalance tool.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 882 - Repair tool

* Code-derived improvement: Repair tool.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 883 - Verify tool

* Code-derived improvement: Verify tool.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 884 - Doctor tool cluster-wide

* Code-derived improvement: Doctor tool cluster-wide.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 885 - Topology visualizer CLI

* Code-derived improvement: Topology visualizer CLI.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 886 - Dot graph export of peers

* Code-derived improvement: Dot graph export of peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 887 - JSON topology export

* Code-derived improvement: JSON topology export.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 888 - Prometheus peer target discovery file SD

* Code-derived improvement: Prometheus peer target discovery file SD.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 889 - HTTP SD for peers

* Code-derived improvement: HTTP SD for peers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 890 - Consul SD documented

* Code-derived improvement: Consul SD documented.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 891 - K8s headless service peer discovery

* Code-derived improvement: K8s headless service peer discovery.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 892 - StatefulSet ordinal bootstrap

* Code-derived improvement: StatefulSet ordinal bootstrap.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 893 - DaemonSet edge profile

* Code-derived improvement: DaemonSet edge profile.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 894 - Job for batch verify

* Code-derived improvement: Job for batch verify.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 895 - CronJob for GC

* Code-derived improvement: CronJob for GC.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 896 - CronJob for backup

* Code-derived improvement: CronJob for backup.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 897 - CronJob for key rotation reminder

* Code-derived improvement: CronJob for key rotation reminder.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 898 - Secret rotation operator

* Code-derived improvement: Secret rotation operator.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 899 - ExternalSecrets example

* Code-derived improvement: ExternalSecrets example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 900 - SealedSecrets example

* Code-derived improvement: SealedSecrets example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 901 - SOPS example

* Code-derived improvement: SOPS example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 902 - age encrypted env files

* Code-derived improvement: age encrypted env files.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 903 - direnv example

* Code-derived improvement: direnv example.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 904 - dotenv never commit

* Code-derived improvement: dotenv never commit.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 905 - gitignore secrets patterns audit

* Code-derived improvement: gitignore secrets patterns audit.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 906 - pre-commit detect-private-key

* Code-derived improvement: pre-commit detect-private-key.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 907 - gitleaks in CI

* Code-derived improvement: gitleaks in CI.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 908 - trufflehog optional

* Code-derived improvement: trufflehog optional.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 909 - AWS access key patterns

* Code-derived improvement: AWS access key patterns.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 910 - GitHub token patterns

* Code-derived improvement: GitHub token patterns.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 911 - Cluster key entropy check

* Code-derived improvement: Cluster key entropy check.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 912 - Auth token entropy check

* Code-derived improvement: Auth token entropy check.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 913 - Reject weak keys

* Code-derived improvement: Reject weak keys.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 914 - Key stretch documentation

* Code-derived improvement: Key stretch documentation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 915 - HSM lunch-and-learn notes

* Code-derived improvement: HSM lunch-and-learn notes.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 916 - SOC2 control mapping living doc

* Code-derived improvement: SOC2 control mapping living doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 917 - ISO27001 control mapping

* Code-derived improvement: ISO27001 control mapping.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 918 - GDPR DPA template

* Code-derived improvement: GDPR DPA template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 919 - HIPAA out of scope statement

* Code-derived improvement: HIPAA out of scope statement.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 920 - PCI out of scope statement

* Code-derived improvement: PCI out of scope statement.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 921 - FedRAMP path evaluation

* Code-derived improvement: FedRAMP path evaluation.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 922 - CJIS out of scope

* Code-derived improvement: CJIS out of scope.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 923 - Customer questionnaire pack

* Code-derived improvement: Customer questionnaire pack.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 924 - SIG Lite answers

* Code-derived improvement: SIG Lite answers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 925 - CAIQ lite answers

* Code-derived improvement: CAIQ lite answers.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 926 - Trust center markdown

* Code-derived improvement: Trust center markdown.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 927 - Status history archive

* Code-derived improvement: Status history archive.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 928 - Incident public comms templates

* Code-derived improvement: Incident public comms templates.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 929 - Postmortem blameless culture doc

* Code-derived improvement: Postmortem blameless culture doc.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 930 - Learning review template

* Code-derived improvement: Learning review template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 931 - Near miss log

* Code-derived improvement: Near miss log.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 932 - Bug bash events

* Code-derived improvement: Bug bash events.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 933 - Hackathon backlog grooming

* Code-derived improvement: Hackathon backlog grooming.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 934 - Design doc template

* Code-derived improvement: Design doc template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 935 - RFC template

* Code-derived improvement: RFC template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 936 - Experiment template

* Code-derived improvement: Experiment template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 937 - Spike timebox 2 days rule

* Code-derived improvement: Spike timebox 2 days rule.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 938 - Definition of done checklist

* Code-derived improvement: Definition of done checklist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 939 - Definition of ready checklist

* Code-derived improvement: Definition of ready checklist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 940 - Story point guide

* Code-derived improvement: Story point guide.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 941 - Milestone exit criteria template

* Code-derived improvement: Milestone exit criteria template.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 942 - MVP exit criteria checklist

* Code-derived improvement: MVP exit criteria checklist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

## Milestone 943 - v1.0 exit criteria checklist

* Code-derived improvement: v1.0 exit criteria checklist.
* Touchpoints: `codebase`
* Acceptance: Change is implemented, tested, and documented.

---

## Summary

* Extended milestones in this file: **903** (numbers 41-943).
* Canonical early milestones 1-40 remain in `documentation/ROADMAP.md`.
* Prefer fileserver/crypto streaming/API wiring/CI-blocking e2e before ecosystem items.

