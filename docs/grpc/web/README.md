# gRPC-Web Support

This document describes the gRPC-Web implementation for PeerVault, which enables browser compatibility for gRPC services.

## Overview

gRPC-Web allows gRPC services to be consumed directly from web browsers by translating gRPC calls to HTTP/1.1 requests. This enables seamless integration of PeerVault's gRPC API with web applications.

## Features

- **Browser Compatibility**: Works in all modern browsers without HTTP/2 support
- **TypeScript Client**: Auto-generated TypeScript client with full type safety
- **CORS Support**: Configurable CORS policies for cross-origin requests
- **Authentication**: Bearer token authentication for secure access
- **Streaming Support**: Full support for gRPC streaming operations
- **Error Handling**: Comprehensive error handling and retry mechanisms

## Architecture

```text
Browser Client (TypeScript)
    ↓ HTTP/1.1 + gRPC-Web Protocol
gRPC-Web Server (Go)
    ↓ gRPC Protocol
gRPC Services (Go)
```

## Server Implementation

The gRPC-Web server is implemented in `internal/api/grpc/web/server.go` and provides:

### Configuration

```go
type Config struct {
    Port           string
    AllowedOrigins []string
    CORSEnabled    bool
    AuthToken      string
}
```

### Key Features

1. **CORS Middleware**: Handles cross-origin requests with configurable allowed origins
2. **Authentication Interceptors**: Validates Bearer tokens for all requests
3. **Health Check Endpoint**: Provides HTTP health check endpoint
4. **Graceful Shutdown**: Proper cleanup of resources on shutdown

### Usage

```go
config := &web.Config{
    Port: ":8080",
    AllowedOrigins: []string{
        "http://localhost:3000",
        "https://yourdomain.com",
    },
    CORSEnabled: true,
    AuthToken: "your-secret-token",
}

server := web.NewWebServer(config, logger)
server.Start()
```

## TypeScript Client

The TypeScript client is located in `sdk/typescript/grpc/` and provides:

### Installation

```bash
cd sdk/typescript/grpc
npm install
npm run build
```

### Client Usage

```typescript
import { PeerVaultClient } from '@peervault/grpc-client';

const client = new PeerVaultClient({
    endpoint: 'http://localhost:8080',
    authToken: 'your-secret-token',
    timeout: 30000,
    retryAttempts: 3,
});

// File operations
const fileData = new Uint8Array([1, 2, 3, 4]);
const response = await client.uploadFile('test.txt', fileData);

// Streaming operations
const unsubscribe = client.streamFileOperations((event) => {
    console.log('File operation:', event);
});
```

### Client Features

1. **Type Safety**: Full TypeScript support with generated types
2. **Streaming**: Support for all streaming operations
3. **Error Handling**: Comprehensive error handling with retry logic
4. **Authentication**: Automatic Bearer token injection
5. **Chunked Upload/Download**: Efficient file transfer with chunking

## Protocol Details

### gRPC-Web Protocol

The implementation uses the gRPC-Web protocol which:

1. **Translates gRPC calls to HTTP/1.1**: Uses POST requests with protobuf payloads
2. **Handles streaming**: Uses Server-Sent Events (SSE) for server streaming
3. **Manages metadata**: Converts gRPC metadata to HTTP headers
4. **Error handling**: Maps gRPC status codes to HTTP status codes

### Message Format

```typescript
// Request headers
Content-Type: application/grpc-web+proto
Authorization: Bearer <token>
X-Grpc-Web: 1

// Response headers
Content-Type: application/grpc-web+proto
X-Grpc-Status: 0
X-Grpc-Message: ""
```

## Security Considerations

### Authentication

- **Bearer Token**: All requests require valid Bearer token
- **Token Validation**: Server validates tokens on every request
- **Secure Transmission**: Use HTTPS in production

### CORS Configuration

```go
AllowedOrigins: []string{
    "https://yourdomain.com",     // Production domain
    "http://localhost:3000",      // Development
}
```

### Best Practices

1. **Use HTTPS**: Always use HTTPS in production
2. **Token Rotation**: Implement token rotation for enhanced security
3. **Rate Limiting**: Implement rate limiting to prevent abuse
4. **Input Validation**: Validate all inputs on both client and server

## Performance Considerations

### Chunking

- **File Upload**: Files are uploaded in 64KB chunks
- **File Download**: Files are downloaded in 64KB chunks
- **Memory Efficiency**: Streaming prevents memory issues with large files

### Connection Management

- **Keep-Alive**: HTTP keep-alive for connection reuse
- **Timeout Configuration**: Configurable timeouts for different operations
- **Retry Logic**: Automatic retry with exponential backoff

## Error Handling

### Client-Side Errors

```typescript
try {
    const response = await client.getFile('nonexistent.txt');
} catch (error) {
    if (error.message.includes('not found')) {
        // Handle file not found
    } else if (error.message.includes('unauthorized')) {
        // Handle authentication error
    }
}
```

### Server-Side Errors

```go
// Return appropriate gRPC status codes
return status.Error(codes.NotFound, "file not found")
return status.Error(codes.Unauthenticated, "invalid token")
return status.Error(codes.Internal, "internal server error")
```

## Development Setup

### Prerequisites

- Go 1.19+
- Node.js 16+
- Protocol Buffers compiler (protoc)
- gRPC-Web plugin

### Building the Server

```bash
# Build the gRPC-Web server
go build -o bin/peervault-grpc-web ./cmd/peervault-grpc-web
```

### Generating TypeScript Client

```bash
cd sdk/typescript/grpc
npm install
npm run generate
npm run compile
```

### Testing

```bash
# Start the server
./bin/peervault-grpc-web

# Test with curl
curl -X POST http://localhost:8080/peervault.PeerVaultService/HealthCheck \
  -H "Content-Type: application/grpc-web+proto" \
  -H "Authorization: Bearer your-secret-token"
```

## Browser Compatibility

| Browser | Version | Support |
|---------|---------|---------|
| Chrome | 60+ | ✅ Full |
| Firefox | 55+ | ✅ Full |
| Safari | 11+ | ✅ Full |
| Edge | 79+ | ✅ Full |
| IE | 11 | ⚠️ Limited |

## Troubleshooting

### Common Issues

1. **CORS Errors**: Check `AllowedOrigins` configuration
2. **Authentication Failures**: Verify token format and validity
3. **Streaming Issues**: Check browser WebSocket support
4. **Type Generation**: Ensure protoc and plugins are installed

### Debug Mode

Enable debug mode for detailed logging:

```typescript
const client = new PeerVaultClient({
    endpoint: 'http://localhost:8080',
    authToken: 'your-secret-token',
    debug: true, // Enable debug logging
});
```

## Future Enhancements

- [ ] WebSocket transport for better streaming performance
- [ ] Compression support for large payloads
- [ ] Client-side load balancing
- [ ] Automatic retry with circuit breaker
- [ ] Metrics and monitoring integration
