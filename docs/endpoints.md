# API Endpoints

Ports below match `docker-compose.apis.yml` (API stack). P2P nodes use `3000`, `5000`, and `7000`.

| Service | Default port | Base URL |
|---------|--------------|----------|
| REST API | `8080` | `http://localhost:8080` |
| GraphQL API | `8081` | `http://localhost:8081` |
| gRPC / HTTP-JSON API | `8082` | `http://localhost:8082` |
| P2P node 1 | `3000` | TCP |
| P2P node 3 (client) | `5000` | TCP |
| P2P node 2 | `7000` | TCP |

Start the API stack:

```bash
docker compose -f docker-compose.apis.yml up -d --build
```

---

## REST API (`:8080`)

Docs UI: `/docs` · OpenAPI JSON: `/swagger.json` · Spec file: [api/shugo-rest-api.yaml](api/shugo-rest-api.yaml)

### System

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Metrics |
| `GET` | `/system` | System info |
| `GET` | `/api` | API root / discovery |
| `POST` | `/webhook` | Webhook receiver |
| `GET` | `/docs` | Swagger UI |
| `GET` | `/swagger.json` | OpenAPI document |

### Files (`/api/v1`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/files` | List files |
| `GET` | `/api/v1/files/get` | Get file by key |
| `POST` | `/api/v1/files` | Upload file |
| `DELETE` | `/api/v1/files` | Delete file |
| `PUT` | `/api/v1/files/metadata` | Update metadata |

### Peers (`/api/v1`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/peers` | List peers |
| `GET` | `/api/v1/peers/get` | Get peer by id |
| `POST` | `/api/v1/peers` | Add peer |
| `DELETE` | `/api/v1/peers` | Remove peer |

### Analytics (`/api/v1`)

| Method | Path |
|--------|------|
| `GET` | `/api/v1/analytics/summary` |
| `GET` | `/api/v1/analytics/usage` |
| `GET` | `/api/v1/analytics/endpoint` |
| `GET` | `/api/v1/analytics/user` |
| `GET` | `/api/v1/analytics/trends` |
| `GET` | `/api/v1/analytics/popularity` |
| `GET` | `/api/v1/analytics/calls` |
| `GET` | `/api/v1/analytics/dashboard` |

See [api/analytics/README.md](api/analytics/README.md).

### Monitoring (`/api/v1`)

| Method | Path |
|--------|------|
| `GET` | `/api/v1/monitoring/snapshot` |
| `GET` | `/api/v1/monitoring/realtime` |
| `GET` | `/api/v1/monitoring/response-time` |
| `GET` | `/api/v1/monitoring/throughput` |
| `GET` | `/api/v1/monitoring/endpoint` |
| `GET` | `/api/v1/monitoring/trend` |
| `GET` | `/api/v1/monitoring/report` |
| `GET` | `/api/v1/monitoring/alerts` |
| `POST` | `/api/v1/monitoring/alerts/silence` |
| `POST` | `/api/v1/monitoring/alerts/resolve` |
| `GET` | `/api/v1/monitoring/recommendations` |
| `GET` | `/api/v1/monitoring/dashboard` |
| `GET` | `/api/v1/monitoring/health-score` |
| `GET` | `/api/v1/monitoring/summary` |

See [api/monitoring/README.md](api/monitoring/README.md).

---

## GraphQL API (`:8081`)

| Path | Description |
|------|-------------|
| `/graphql` | GraphQL HTTP endpoint |
| `/playground` | GraphQL Playground |
| `/health` | Health check |
| `/metrics` | Metrics |

Full guide: [graphql/README.md](graphql/README.md).

---

## gRPC / HTTP-JSON API (`:8082`)

The current server exposes HTTP/JSON routes (true protobuf gRPC streaming is still in progress). See [grpc/README.md](grpc/README.md).

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/system/info` | System information |
| `GET` | `/system/metrics` | Metrics |
| `GET` | `/files` | List files |
| `GET` | `/files/{key}` | Get file |
| `DELETE` | `/files/{key}` | Delete file |
| `GET` | `/peers` | List peers |
| `GET` | `/peers/{id}` | Get peer |
| `POST` | `/peers` | Add peer |
| `DELETE` | `/peers/{id}` | Remove peer |
| `GET` | `/peers/{id}/health` | Peer health |

Auth token for the compose stack default: `demo-token` (set via `--auth-token`).

---

## Related protocols

Additional servers exist as separate binaries/docs (not all started by default compose):

- WebSocket — [api/websocket/README.md](api/websocket/README.md)
- Protocol translation — [api/translation/README.md](api/translation/README.md)
- MQTT / CoAP / SSE — see `cmd/shugo-mqtt`, `cmd/shugo-coap`, `cmd/shugo-sse`
