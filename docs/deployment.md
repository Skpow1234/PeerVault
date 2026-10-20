# Deployment

PeerVault is intended to run via **Docker Compose**. There is no Helm chart or Kubernetes manifests in the repo today (the README badge is aspirational). Production-like setups use multi-container Compose; Kubernetes is only mentioned as a future orchestration option in [CONTAINERIZATION.md](../documentation/CONTAINERIZATION.md).

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- ~1GB free memory
- Free ports as needed: `3000`, `5000`, `7000` (P2P) and/or `8080`–`8082` (APIs)

## Option A — Full stack (APIs + P2P)

```bash
docker compose -f docker-compose.apis.yml up -d --build
```

Starts REST (`8080`), GraphQL (`8081`), gRPC/HTTP-JSON (`8082`), and three P2P nodes (`3000`, `5000`, `7000`).

## Option B — P2P network only

```bash
docker compose up -d --build
```

Uses root `docker-compose.yml`: three nodes + optional demo client that stores/fetches sample files.

Useful log signals:

- `[::]:5000 starting fileserver...`
- `received and written (...) bytes to disk`

## Option C — Dev all-in-one

```bash
docker compose -f docker-compose.dev.yml up --build
```

Or a single image:

```bash
docker build -t peervault .
docker run --rm -p 3000:3000 -p 5000:5000 -p 7000:7000 peervault
```

## Option D — Individual nodes

```bash
docker build -f docker/Dockerfile.node -t peervault-node .

docker run -d --name node1 -p 3000:3000 peervault-node --listen :3000

docker run -d --name node2 -p 5000:5000 peervault-node \
  --listen :5000 --bootstrap node1:3000
```

## Compose layouts

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Multi-node P2P + demo |
| `docker-compose.apis.yml` | APIs + P2P |
| `docker-compose.dev.yml` | Development all-in-one |
| `docker/docker-compose*.yml` | Copies under `docker/` (see [docker/README.md](../docker/README.md)) |

Images: `docker/Dockerfile`, `Dockerfile.node`, `Dockerfile.demo`, `Dockerfile.rest-api`, `Dockerfile.grpc-api`, `Dockerfile.graphql-api`.

## Environment variables

Set via Compose `environment:` (preferred for shared cluster settings):

| Variable | Purpose |
|----------|---------|
| `PEERVAULT_CLUSTER_KEY` | Shared encryption key material for nodes |
| `PEERVAULT_AUTH_TOKEN` | Shared peer auth token |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` |

If unset, demo/default tokens and per-node keys are used (dev only).

### Node flags (`peervault-node`)

```text
--listen string          Listen address (default ":3000")
--bootstrap string       Comma-separated bootstrap addresses
--log-level string       Log level (default "info")
--storage-prefix string  Storage directory prefix
```

### Demo flags (`peervault-demo`)

```text
--target string     Target node (default "localhost:5000")
--bootstrap string  Bootstrap addresses
--iterations int    Store/get iterations (default 20)
```

## Ops commands

```bash
docker compose logs -f
docker compose down
docker compose down -v   # also remove volumes
```

## What is not in-repo

- **No Helm / K8s manifests** — scale with Compose today; Swarm/K8s called out as future options in containerization docs
- **No managed cloud deploy recipes** (ECS, GKE, etc.)

More detail: [documentation/CONTAINERIZATION.md](../documentation/CONTAINERIZATION.md) · [docker/README.md](../docker/README.md)
