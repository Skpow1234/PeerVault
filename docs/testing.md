# Testing & quality

Prefer running tests through **CI / containers**. Local non-Docker workflows were trimmed in favor of Docker-only DX.

## What CI runs

See [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) and [`.github/workflows/README.md`](../.github/workflows/README.md).

### Required (fail the merge gate — `CI Success`)

- Lint & format: `gofmt`, `go mod tidy` drift, golangci-lint (`config/.golangci.yml`), trailing whitespace
- Unit + race: `./tests/unit/...`, `./internal/...` (websocket soft-excluded under race)
- Core coverage floor (~40%) on crypto / storage / p2p / fileserver
- P2P replication E2E: `TestReplicationStoreAGetB`
- REST API smoke: `./tests/integration/rest/...`
- `govulncheck ./...`
- Cross-platform build (Linux / Windows / macOS)
- Docker compose validate + node/demo image build + smoke

### Soft (`continue-on-error` — warn only)

- Broader integration (gRPC, GraphQL, multi-node, full e2e, websocket)
- Time-boxed fuzz under `./tests/fuzz/...`
- Extra static security (staticcheck, Semgrep)
- Benchmarks (artifact)
- Docs sanity

Promote soft suites to required once they are stable. Deeper security lives in `security.yml`.

## Local equivalents

```bash
gofmt -s -l .
go mod tidy
golangci-lint run --config=config/.golangci.yml
go test -race ./tests/unit/... ./internal/...
go test -count=1 -timeout=3m ./tests/integration/end-to-end/ -run TestReplicationStoreAGetB
go test ./tests/integration/rest/...
govulncheck ./...
```

## API testing

- Postman collections: `tests/api/collections/`
- Mock server + OpenAPI-driven mocks
- Contract tests under `tests/contracts/`
- Performance (Go + k6) under `tests/performance/`
- Security/OWASP-oriented API tests under `tests/security/` and `internal/api/security/`

Guides:

- [api/testing/README.md](api/testing/README.md)
- [api/testing/IMPLEMENTATION.md](api/testing/IMPLEMENTATION.md)

Mock config example: `config/mock-server.yaml`.

## Lint

```bash
golangci-lint run --config=config/.golangci.yml
```
