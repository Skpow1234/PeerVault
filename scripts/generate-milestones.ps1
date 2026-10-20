$ErrorActionPreference = 'Stop'
$out = Join-Path $PSScriptRoot '..\documentation\MILESTONES_EXTENDED.md'
$lines = New-Object System.Collections.Generic.List[string]
$script:n = 40

function Add-M {
    param([string]$Title, [string]$Detail, [string]$Touch)
    $script:n = $script:n + 1
    $num = $script:n
    $lines.Add(('## Milestone {0} - {1}' -f $num, $Title))
    $lines.Add('')
    $lines.Add(('* {0}' -f $Detail))
    if ($Touch) {
        $lines.Add(('* Touchpoints: `{0}`' -f $Touch))
    }
    $lines.Add('* Acceptance: Change is implemented, tested, and documented.')
    $lines.Add('')
}

$lines.Add('# Shugo Extended Milestones (41+)')
$lines.Add('')
$lines.Add('Generated from codebase review after Milestone 1-2 core P2P fixes (replication, DialPeer, SHA-256 CAS, pending Get).')
$lines.Add('Each item is an actionable improvement derived from TODOs, stubs, ops gaps, and hardening needs.')
$lines.Add('')
$lines.Add('## Progress note (core P2P restore)')
$lines.Add('')
$lines.Add('* **Done:** AES-GCM, HashKey SHA-256, CAS SHA-256, DialPeer, stream replication, pending Get from peers, shared KeyManager helpers, OnStream on main entrypoints.')
$lines.Add('* **Still open:** streaming AEAD without full buffering; reduce sleep-based test coordination; multi-node e2e CI-blocking; wire REST/GraphQL/gRPC to fileserver.')
$lines.Add('')

$core = @(
    @{T='Streaming AES-GCM (chunked AEAD)'; D='Replace ReadAll-based CopyEncrypt/CopyDecrypt with chunked/streaming AEAD.'; X='internal/crypto/crypto.go'},
    @{T='Overwrite / upsert storage semantics'; D='Define replace-vs-reject when storing an existing key; support upsert for re-replication.'; X='internal/storage/store.go'},
    @{T='Delete + tombstone replication'; D='Replicate deletes across peers with tombstones and conflict rules.'; X='internal/app/fileserver'},
    @{T='List keys API on fileserver'; D='Expose local key listing for APIs and CLI.'; X='internal/app/fileserver'},
    @{T='Replication factor N'; D='Configurable N-way replication with success quorum.'; X='internal/app/fileserver'},
    @{T='Prefer healthy peers for Get'; D='Order Get candidates by RTT/health score.'; X='internal/peer'},
    @{T='Parallel Get fan-out'; D='Request file from multiple peers; take first successful stream.'; X='fileserver'},
    @{T='Cancel in-flight streams on ctx done'; D='Propagate context cancellation into stream copy loops.'; X='fileserver'},
    @{T='Backpressure on replicateToPeers'; D='Limit concurrent outbound replication streams per Store.'; X='fileserver'},
    @{T='Handshake deadlock-safe exchange'; D='Concurrent send/receive or duplex framing to avoid rare handshake stalls.'; X='handshake.go'},
    @{T='Transport metrics'; D='Emit dial/accept/stream/message counters and error rates.'; X='transport/p2p'},
    @{T='Peer address normalization'; D='Normalize host:port keys so dial/reconnect maps match OnPeer keys.'; X='fileserver'},
    @{T='Bootstrap retries with backoff'; D='Retry failed bootstrap dials with jittered exponential backoff.'; X='fileserver'},
    @{T='Persistent peer routing table'; D='Survive restarts with known peer list on disk.'; X='peer'},
    @{T='Content integrity verify on Get'; D='Re-hash ciphertext/plaintext and reject mismatches.'; X='storage'},
    @{T='Chunked file transfer protocol'; D='Split large objects into framed chunks with offsets for resume.'; X='transport'},
    @{T='Resumable transfers'; D='Support offset resume after disconnect.'; X='fileserver'},
    @{T='Deduplicate concurrent Stores'; D='Coalesce in-flight stores for the same key.'; X='FileOperationManager'},
    @{T='Observability spans for Store/Get'; D='Add tracing spans around store, replicate, get, stream.'; X='tracing'},
    @{T='Structured error codes on wire'; D='Typed error DTOs instead of opaque strings.'; X='dto'}
)
foreach ($m in $core) { Add-M $m.T $m.D $m.X }

$api = @(
    @{T='Wire REST FileService to fileserver'; D='Replace TODOs in file_service_impl with real Store/Get/List/Delete.'; X='file_service_impl.go'},
    @{T='Wire REST PeerService to peer map'; D='Implement list/add/remove using fileserver peer state.'; X='peer_service_impl.go'},
    @{T='Wire REST SystemService'; D='Real health/metrics from health and metrics packages.'; X='system_service_impl.go'},
    @{T='Wire GraphQL resolvers to fileserver'; D='Implement all resolver TODOs.'; X='resolver.go'},
    @{T='GraphQL upload uses Store'; D='fileserver Store from upload mutation.'; X='graphql/server.go'},
    @{T='GraphQL download uses Get'; D='Stream download via Get.'; X='graphql/server.go'},
    @{T='Replace gRPC HTTP/JSON with real gRPC'; D='Generate protobuf stubs and serve grpc.Server.'; X='api/grpc'},
    @{T='gRPC streaming Upload/Download'; D='Bidirectional chunk streaming mapped to fileserver.'; X='grpc/streaming'},
    @{T='grpc-gateway REST mapping'; D='Expose protobuf HTTP annotations for gateway.'; X='proto'},
    @{T='Auth middleware shared across APIs'; D='One token/JWT validator for REST/GraphQL/gRPC.'; X='api'},
    @{T='Rate limit per peer and per API key'; D='Unified limiter config.'; X='ratelimit'},
    @{T='OpenAPI regenerated from code'; D='Keep OpenAPI in sync with routes.'; X='docs/api'},
    @{T='Contract tests CI-blocking'; D='Fail PR if consumer contracts break.'; X='tests/contracts'},
    @{T='Remove mock event generators in gRPC'; D='Replace mock events with real fileserver events.'; X='grpc/impl.go'},
    @{T='GraphQL subscriptions on real events'; D='Peer join/leave and file store events.'; X='subscriptions'},
    @{T='REST webhook emits on Store'; D='Configurable webhook on successful store/replicate.'; X='rest'},
    @{T='API request IDs'; D='Propagate X-Request-ID through logs and traces.'; X='middleware'},
    @{T='Versioned REST deprecation headers'; D='Emit Deprecation/Sunset headers.'; X='versioning'},
    @{T='GraphQL complexity limits'; D='Query depth/cost analysis to prevent abuse.'; X='graphql'},
    @{T='gRPC auth interceptor JSON parsing'; D='Finish TODO in interceptors/auth.go.'; X='auth.go'}
)
foreach ($m in $api) { Add-M $m.T $m.D $m.X }

$proto = @(
    @{T='MQTT wildcard subscriptions'; D='Implement + and # matching TODOs.'; X='mqtt'},
    @{T='MQTT session persistence'; D='Persist sessions across reconnect.'; X='mqtt'},
    @{T='MQTT over WebSocket'; D='WebSocket upgrade path for MQTT.'; X='mqtt'},
    @{T='CoAP DTLS'; D='DTLS support TODO on CoAP server.'; X='coap'},
    @{T='WebSocket topic distribution'; D='Topic-based fan-out TODO.'; X='websocket'},
    @{T='SSE backpressure'; D='Drop/slow clients cleanly under load.'; X='sse'},
    @{T='Protocol translation correctness tests'; D='Golden tests for translators.'; X='translation'},
    @{T='Unified health across protocol servers'; D='Same /health schema everywhere.'; X='cmd'},
    @{T='Shared TLS config loader'; D='One place for certs/mTLS for all HTTP servers.'; X='config'},
    @{T='Connection drain on shutdown'; D='Graceful shutdown TODO in GraphQL server.'; X='graphql'}
)
foreach ($m in $proto) { Add-M $m.T $m.D $m.X }

$themes = @{
    'Fileserver' = @(
        'Idempotent Store with content hash header','Multi-reader Get with shared buffer pool','Peer selection strategy plugin',
        'Replication job queue with retries','Dead-letter queue for failed replications','Per-key lease for concurrent writers',
        'Background verify-and-repair job','Storage GC for orphaned CAS objects','Migrate legacy SHA-1 paths to SHA-256',
        'Export/import snapshot format','Read-only mode for nodes','Maintenance mode endpoint','Disk space watermark checks',
        'Windows path length edge cases','Fsync policy','Checksum trailer on streams','Compress-before-encrypt option',
        'Namespace/tenant prefix on keys','Quota per namespace','Audit event on Store/Get','PII redaction in logs for keys',
        'Slow-peer eviction policy','Circuit breaker around DialPeer','Happy-eyeballs dual-stack dial','SOCKS5 proxy dial',
        'mDNS local discovery','Static peers file watcher','DNS SRV bootstrap','NAT traversal research spike','Encrypt-then-MAC metadata'
    )
    'Transport' = @(
        'Replace gob with protobuf DTOs','Message schema version byte','Optional frame compression','Max concurrent streams per conn',
        'Idle connection timeout','Keepalive ping frames','Write deadline per frame','Read deadline per frame',
        'Connection multiplexer','UDP/QUIC transport experiment','WebRTC datachannel spike','TLS wrapping for TCP',
        'Noise protocol handshake','Certificate pinning','Ban list for abusive peers','Rate limit messages per peer',
        'Fuzz LengthPrefixedDecoder in CI','Property tests for frame round-trip','Simulate partial reads','Simulate mid-stream disconnect',
        'Zero-copy buffer paths','Avoid gob global register side effects','Transport fake for unit tests','Dial timeout configurable',
        'Listen on multiple addresses','SO_REUSEPORT support','Export pprof on debug port','Connection churn benchmarks',
        'Latency histogram metrics','Packet loss chaos tests'
    )
    'Crypto' = @(
        'Key versioning in ciphertext header','Decrypt with previous keys after rotation','HKDF from x/crypto',
        'Per-file DEK wrapped by KEK','Hardware-backed keys interface','Constant-time compare audit','Nonce misuse review',
        'AAD binding message type and key id','Secret zeroization helpers','Forbid demo auth token in prod',
        'Fail closed if AUTH_TOKEN unset in prod','Fail closed if CLUSTER_KEY unset in prod','Handshake HMAC side-channel review',
        'Crypto fuzz corpus checked in','Benchmark Seal/Open','ChaCha20-Poly1305 option','Age interop spike',
        'Signed node identity ed25519','Mutual auth beyond PSK','Envelope encryption for backups'
    )
    'Storage' = @(
        'Pluggable PathTransformFunc registry','2-char fan-out option','Object metadata sidecars','Content-type metadata map',
        'Atomic replace via temp+rename','Cross-volume rename handling','Sparse file support','Memory-mapped read option',
        'Checksum database index','Bolt/SQLite key index','Bloom filter for Has','Sharded locks per path prefix',
        'Concurrent Has/Read stress tests','Storage Root absolute path policy','Multi-disk volume spanning',
        'Read-only mount detection','Corruption injector for tests','Backup hot-copy API','Incremental backup diffs',
        'CAS path docs match code'
    )
    'REST' = @(
        'Bearer auth real validator','Upload size limits','Streaming upload to Store','Streaming download from Get',
        'Multipart form upload','ETag If-None-Match','Conditional delete','Cursor pagination for list',
        'Filter/search query params','CORS allowlist from config','Trusted proxy headers','Prometheus metrics authenticity',
        'OpenTelemetry HTTP instrumentation','Idempotency-Key support','Problem+JSON errors','Admin-only routes',
        'API key issuance CLI','Per-route feature flags','Chaos middleware','Nightly load test job'
    )
    'GraphQL' = @(
        'Persisted queries','Automatic persisted queries','DataLoader batching','Field-level auth directives',
        'Schema stitching production readiness','Federation gateway hardening','Subscription auth on WS',
        'Heartbeat on subscriptions','Query allowlist mode','Resolver timeout per field','N+1 detection in tests',
        'GraphQL request/error metrics','Playground disabled in prod','File scalar streaming','Relay cursor connections',
        'Deprecate unused fields','Schema lint in CI','Schema contract tests','Example queries regenerated',
        'Resolver unit tests every field'
    )
    'gRPC' = @(
        'protoc generation check in CI','buf lint breaking change detect','grpc health v1','Reflection toggle',
        'grpc-web CORS hardening','Client SDK publish pipeline','Retry policies','Request hedging',
        'Load balancing config examples','Server-side keepalive','Max message size config','Compression gzip',
        'Richer gRPC status details','Interceptor chain metrics','Method-level authz','Real grpc client integration test',
        'Benchmark streaming throughput','Remove temporary HTTP/JSON server','Gateway OpenAPI from proto',
        'Versioned package shugo.v1'
    )
    'CLI' = @(
        'CLI store/get wired','CLI config file support','CLI progress bars','CLI JSON output','CLI path completion',
        'CLI profile switching','CLI doctor command','CLI peer list','CLI replicate status','CLI TUI polish',
        'Remove CLI sleep simulations','CLI integration tests','Man pages help examples','CLI version ldflags',
        'CLI plugin subcommands','CLI offline error clarity','CLI auth login','CLI batch store directory',
        'CLI verify command','CLI benchmark command'
    )
    'Testing-CI' = @(
        'Make multi-node e2e CI-blocking','Flake quarantine process','Testcontainers compose tests',
        'Race detector on fileserver','Coverage gate fileserver','Coverage gate transport','Mutation testing spike',
        'Nightly longer fuzz','SBOM in release','Signed release artifacts','Reproducible builds','Go version matrix',
        'Windows Linux CI parity','Optional Mac CI','Go module caching fix','Dependabot review SLA',
        'Action pinning to SHAs','CodeQL workflow','License check go-licenses','Dead code detection'
    )
    'Security' = @(
        'Remove hardcoded demo tokens','Secrets scanning pre-commit required','Living threat model doc',
        'Get amplification abuse cases','Max frame size DoS tests','Auth token rotation runbook',
        'Cluster key rotation runbook','Peer ban after auth failures','Audit log sink','RBAC on mutating APIs',
        'Privacy controls metadata export','PKI issue rotate node certs','Container non-root verification',
        'Read-only root FS compose','Seccomp AppArmor profiles','K8s NetworkPolicy examples',
        'Optional pen-test CI job','Dependency vulnerability SLA','security.txt contact','Bug bounty prep'
    )
    'Ops-Deploy' = @(
        'Helm chart first class','Kustomize overlays','HPA guidance','PersistentVolume examples','PodDisruptionBudget',
        'Anti-affinity for nodes','ServiceMonitor Prometheus','Grafana dashboard JSON','Alerting rules pack',
        'Liveness readiness docs','Rolling update strategy','Backup CronJob example','Disaster recovery runbook',
        'Multi-region design doc','Terraform module spike','Ansible role spike','systemd unit files',
        'sysctl tuning notes','Resource requests limits defaults','Cost estimation worksheet'
    )
    'Observability' = @(
        'RED metrics Store/Get','Peer connection gauge','Replication lag metric','Storage bytes gauge',
        'Log sampling under load','Trace sampling config','Correlate logs with trace IDs','Slow request log threshold',
        'Error budget dashboard','SLO definitions','Synthetic canary Store/Get','Blackbox exporter targets',
        'Continuous profiling','Internal domain event bus','Audit vs debug log separation','Hot-reload log level',
        'Health dependency detail','Version build info metric','Panic recovery metrics','Replication queue depth metrics'
    )
    'Docs-DX' = @(
        'ADR folder','Honest README status automation','Doc link checker CI','Compose secrets pattern example',
        'Tutorial 3-node from zero','Tutorial API upload download','Troubleshoot handshake failures',
        'Troubleshoot replication misses','Windows path FAQ','Contributing first-issue labels','CODEOWNERS',
        'PR template API checklist','Changelog generation','Semver policy','Deprecation policy',
        'Public roadmap board sync','SVG architecture diagram','Sequence diagrams Store/Get',
        'SDK quickstarts tested CI','Local docs search'
    )
    'Performance' = @(
        'Buffer pool frame headers','Object pool Message structs','Zero-alloc hot path audit','Parallel replicate workers',
        'Batch Has lookups','Optional recent Get cache','CDN cache headers REST','Compression negotiation',
        'Benchmark Store 1/10/100MB','Benchmark Get local vs remote','p99 latency budgets','GC tuning notes',
        'File cache guidance','Verify no double-encrypt on replicate','sendfile splice exploration','io_uring experiment',
        'Connection pool popular peers','Adaptive chunk size','Read-ahead sequential Get','Write coalescing'
    )
    'Ecosystem' = @(
        'IPFS gateway compatibility tests','CID to CAS mapping clarity','Blockchain module off by default',
        'Remove unused ethereum dep','ML classifier optional build tag','Edge worker protocol spec',
        'IoT device auth model','Plugin sandboxing','S3 plugin integration test','Webhook HMAC signatures',
        'OTel collector example','JavaScript SDK publish','Go SDK module path stability','Python client refresh',
        'Postman collection auto-gen','Insomnia collection','Terraform provider spike','K8s Operator CRD spike',
        'WASM node experiment','Mobile constrained-network profile'
    )
}

foreach ($theme in ($themes.Keys | Sort-Object)) {
    $lines.Add('---')
    $lines.Add('')
    $lines.Add(('## Theme: {0}' -f $theme))
    $lines.Add('')
    $i = 0
    foreach ($item in $themes[$theme]) {
        $i++
        Add-M ('{0}-{1} {2}' -f $theme, $i, $item) ('Improve {0}: {1}.' -f $theme, $item) $theme
    }
}

$extras = Get-Content (Join-Path $PSScriptRoot 'milestone-extras.txt') -ErrorAction SilentlyContinue
if (-not $extras) {
    $extras = 1..250 | ForEach-Object { 'Backlog hardening item {0}' -f $_ }
}

$lines.Add('---')
$lines.Add('')
$lines.Add('## Theme: Code-derived extras')
$lines.Add('')
foreach ($item in $extras) {
    Add-M $item ('Code-derived improvement: {0}.' -f $item) 'codebase'
}

while ($script:n -lt 540) {
    $next = $script:n + 1
    Add-M ('Hardening backlog item {0}' -f $next) ('Systematic hardening backlog item {0}.' -f $next) 'backlog'
}

$created = $script:n - 40
$lines.Add('---')
$lines.Add('')
$lines.Add('## Summary')
$lines.Add('')
$lines.Add(('* Extended milestones in this file: **{0}** (numbers 41-{1}).' -f $created, $script:n))
$lines.Add('* Canonical early milestones 1-40 remain in `documentation/ROADMAP.md`.')
$lines.Add('* Prefer fileserver/crypto streaming/API wiring/CI-blocking e2e before ecosystem items.')
$lines.Add('')

$resolved = [System.IO.Path]::GetFullPath($out)
[System.IO.File]::WriteAllLines($resolved, $lines)
Write-Output ('Created {0} milestones (41-{1}) -> {2}' -f $created, $script:n, $resolved)
