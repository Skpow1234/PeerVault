# Architecture

## High level

1. Entrypoints under `cmd/` start nodes or API servers.
2. `internal/app/fileserver` owns store/get and peer messaging.
3. `internal/transport/p2p` provides TCP transport, framing, and handshake.
4. `internal/storage` persists content-addressed blobs on disk.
5. `internal/crypto` handles AES-GCM and key derivation.
6. `internal/api/*` exposes REST, GraphQL, and gRPC/HTTP-JSON faces.

### Intended data flow

- **Store:** write locally (encrypted) → broadcast control message → peers pull encrypted stream.
- **Get:** read local CAS path → if missing, broadcast request → peer streams file back.

Framing: `[type:u8][len:u32][payload:len]` — message `0x1`, stream header `0x2`. Max frame 1MB.

### Content addressing

Paths are derived from a hash of the key (`CASPathTransformFunc` in `internal/storage`). Docs historically describe SHA-256; verify the implementation when changing storage. IPFS/CID work lives under `internal/content` and `internal/ipfs` (ecosystem / Milestone 9).

## Project layout (overview)

```text
cmd/                 # peervault, node, demo, api, graphql, grpc, …
internal/
  app/fileserver/    # Core P2P file server
  transport/p2p/     # TCP, framing, handshake
  storage/           # CAS disk store
  crypto/            # Encryption & keys
  peer/              # Health & resources
  api/               # REST, GraphQL, gRPC, analytics, …
tests/               # Unit, integration, fuzz, API, security
docker/              # Dockerfiles & compose helpers
docs/                # This documentation tree
documentation/       # Roadmap, security policy, containerization, …
proto/               # Protobuf definitions
config/              # YAML configs
scripts/             # Build / test / security helpers
```

Key binaries: `peervault`, `peervault-node`, `peervault-demo`, `peervault-api`, `peervault-graphql`, `peervault-grpc`.

## Related docs

- [Endpoints](endpoints.md)
- [Deployment](deployment.md)
- [Security](security.md)
- REST layer notes: [api/README.md](api/README.md)
