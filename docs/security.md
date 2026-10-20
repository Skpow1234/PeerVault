# Security

Deep dive: [documentation/ENCRYPTION.md](../documentation/ENCRYPTION.md) · Policy: [documentation/SECURITY.md](../documentation/SECURITY.md).

## Encryption (AES-GCM)

- AES-256-GCM with 12-byte random nonces (never reused)
- **Chunked streaming format** (`PVCH`): encrypts in 64KiB plaintext chunks so large files are not fully buffered; each chunk binds AAD = chunk index
- Legacy single-blob ciphertext remains decryptable
- Keys derived from a cluster key via HMAC-SHA256

## Key management

- Set `PEERVAULT_CLUSTER_KEY` on all nodes for a shared cluster (via Compose env)
- Optional rotation (default interval documented as 24h in product docs)
- Without a cluster key, each node generates its own key (dev/demo only)

## Peer authentication

- HMAC-SHA256 handshake with node identity and timestamp skew checks
- Set `PEERVAULT_AUTH_TOKEN` across nodes for production-like demos
- Default demo token is for local development only

## MVP vs planned

**In MVP path:** AES-GCM, HMAC peer auth, key derivation, govulncheck/semgrep in CI, atomic file create (`O_CREATE|O_EXCL`).

**Planned / guidance:** broader compliance checklists, RBAC, audit, PKI, SLSA — not certifications. See Milestone 8 notes in [documentation/PIPELINE_UPDATES.md](../documentation/PIPELINE_UPDATES.md).

## CI security

Workflows under `.github/workflows/`:

- `security.yml` — vuln scan, secrets detection, container scan attempts
- `security-dev.yml` — lighter checks on development changes
- Dependabot for dependency updates

Tools commonly referenced: govulncheck, semgrep, detect-secrets, Trivy, gosec.

## Vulnerability note (GO-2025-3750)

Mitigated by requiring Go 1.24.4+ and atomic create helpers in `internal/storage/store.go`.

## Policies

Example policy YAML under `security/policies/` (access control, data classification). These are guidance artifacts, not runtime enforcers by themselves.
