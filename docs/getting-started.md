# Getting started

## Requirements

- Go 1.24.4+ (if building locally)
- Docker Engine 20.10+ and Docker Compose 2.0+ (recommended path)
- Make or Task (optional)

## Quick start (Docker)

**APIs + P2P:**

```bash
docker compose -f docker-compose.apis.yml up -d --build
```

**P2P only:**

```bash
docker compose up -d --build
```

Then open:

- REST docs: `http://localhost:8080/docs`
- GraphQL playground: `http://localhost:8081/playground`
- gRPC/HTTP health: `http://localhost:8082/health`

Endpoint reference: [endpoints.md](endpoints.md) · Deploy options: [deployment.md](deployment.md).

## How it works (short)

- `cmd/peervault` (or Compose nodes) boot TCP peers via `internal/transport/p2p`
- Files land in content-addressed storage under `internal/storage`
- Store writes locally and is intended to replicate to peers; get fetches locally or from peers
- APIs (`cmd/peervault-api`, `-graphql`, `-grpc`) expose HTTP/GraphQL/gRPC-style access

More: [architecture.md](architecture.md).

## Windows Defender

Go binaries that open sockets and touch the filesystem often trigger Defender. Prefer the **Docker-only** workflow to avoid local exclusions.

If you must run native binaries: add folder/process exclusions under Windows Security → Virus & threat protection → Exclusions (project folder; `peervault*.exe`, `go.exe`). Only do this in trusted environments. Scripts under `scripts/` (`quick_trust.bat`, `trust_application.ps1`) exist for local setup — not for production.

## Clean up local data

Demo nodes write under per-node storage roots (e.g. Compose volumes or local folders like `ggnetwork`). Remove those directories/volumes, or call `Store.Clear()` from your own code.

```bash
docker compose down -v
```

## Next steps

- [Features](features.md) — what MVP includes vs later
- [Milestones](milestones.md) — roadmap map
- [Security](security.md) — keys and auth env vars
- Portal guide: [portal/guides/getting-started.md](portal/guides/getting-started.md)
