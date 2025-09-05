# PeerVault SDK Documentation

This directory contains SDK documentation and examples for integrating with PeerVault's APIs.

## Available SDKs

### Go SDK

- **Location**: `sdk/go/`
- **Documentation**: [Go SDK Guide](go/README.md)
- **Examples**: [Go Examples](go/examples/)

### JavaScript/TypeScript SDK

- **Location**: `sdk/javascript/`
- **Documentation**: [JavaScript SDK Guide](javascript/README.md)
- **Examples**: [JavaScript Examples](javascript/examples/)

### Python SDK

- **Location**: `sdk/python/`
- **Documentation**: [Python SDK Guide](python/README.md)
- **Examples**: [Python Examples](python/examples/)

### Java SDK

- **Location**: `sdk/java/`
- **Documentation**: [Java SDK Guide](java/README.md)
- **Examples**: [Java Examples](java/examples/)

## Quick Start

### REST API

```bash
# Upload a file
curl -X POST http://localhost:8080/api/v1/files \
  -H "Content-Type: multipart/form-data" \
  -F "file=@example.txt" \
  -F "key=my-file"

# Download a file
curl -X GET http://localhost:8080/api/v1/files/my-file \
  -o downloaded-file.txt
```

### GraphQL API

```graphql
# Query files
query {
  files {
    key
    size
    createdAt
    metadata
  }
}

# Upload a file
mutation {
  uploadFile(input: {
    key: "my-file"
    content: "file content"
  }) {
    key
    size
    success
  }
}
```

### gRPC API

```go
// Connect to gRPC service
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := pb.NewFileServiceClient(conn)

// Upload file
stream, err := client.UploadFile(context.Background())
```

## Authentication

All APIs support JWT-based authentication:

```bash
# Get token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "pass"}'

# Use token
curl -X GET http://localhost:8080/api/v1/files \
  -H "Authorization: Bearer <token>"
```

## API Endpoints

### REST API (Port 8080)

- **Files**: `/api/v1/files/*`
- **Peers**: `/api/v1/peers/*`
- **System**: `/api/v1/system/*`
- **Health**: `/api/v1/health`

### GraphQL API (Port 8081)

- **Endpoint**: `/graphql`
- **Playground**: `/graphql/playground`
- **Schema**: `/graphql/schema`

### gRPC API (Port 9090)

- **File Service**: `FileService`
- **Peer Service**: `PeerService`
- **System Service**: `SystemService`

## Error Handling

All APIs return consistent error responses:

```json
{
  "error": {
    "code": "FILE_NOT_FOUND",
    "message": "File with key 'example' not found",
    "details": {
      "key": "example",
      "timestamp": "2024-01-01T00:00:00Z"
    }
  }
}
```

## Rate Limiting

APIs implement rate limiting:

- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per user
- **Headers**: `X-RateLimit-*` headers included in responses

## Webhooks

Configure webhooks for real-time notifications:

```json
{
  "url": "https://your-app.com/webhook",
  "events": ["file.uploaded", "file.deleted", "peer.connected"],
  "secret": "webhook-secret"
}
```
