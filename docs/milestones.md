# Milestones

Canonical detail: [documentation/ROADMAP.md](../documentation/ROADMAP.md) · Extended backlog: [documentation/MILESTONES_EXTENDED.md](../documentation/MILESTONES_EXTENDED.md)

## Current status

**API surfaces are now backed by the fileserver** (not mock data) for the main file/peer/system paths.

Completed recently:

- Chunked AES-GCM, replication, DialPeer, SHA-256 CAS, CI `p2p-e2e`
- Shared **file registry** + peer APIs on `fileserver`
- **REST** File / Peer / System services wired to fileserver
- **GraphQL** resolvers + upload/download/health/metrics wired
- **gRPC HTTP/JSON facade** file service wired (true protobuf gRPC server still pending)

Still open:

- Full GraphQL schema engine (dispatcher is query-name based)
- True `google.golang.org/grpc` server replacing HTTP/JSON
- Subscriptions on real events
- Bootstrap readiness without sleeps

## Suggested next

1. Stand up real gRPC server from `proto/peervault`  
2. GraphQL library (gqlgen/graphql-go) over current dispatcher  
3. Event bus for subscriptions  
4. Handshake duplex + dial readiness probes  

## Related

- [endpoints.md](endpoints.md) · [architecture.md](architecture.md) · [deployment.md](deployment.md)
