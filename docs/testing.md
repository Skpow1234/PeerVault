# Testing & quality

Prefer running tests through **CI / containers**. Local non-Docker workflows were trimmed in favor of Docker-only DX.

## What CI runs

### Critical (fail the pipeline)

- Unit + race: `./tests/unit/...`, `./internal/...`
- Fuzz (time-boxed) under `./tests/fuzz/...`
- Security scans (semgrep, govulncheck)
- Cross-platform builds
- Docker image builds

### Non-critical (warnings)

- Integration / e2e under `./tests/integration/...`
- golangci-lint, gofmt, some style/doc checks

See `.github/workflows/ci.yml` and [`.github/workflows/README.md`](../.github/workflows/README.md).

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

Lint runs in CI. Install [golangci-lint](https://golangci-lint.run/) locally if you want the same checks before push (`.golangci.yml`).
