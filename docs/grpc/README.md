# PeerVault gRPC API

The PeerVault gRPC API provides high-performance, streaming-capable access to the PeerVault distributed file storage system.

## Overview

The gRPC API is designed for:

- **High-performance applications** requiring low-latency operations
- **Streaming file operations** for large file transfers
- **Real-time event streaming** for monitoring and notifications
- **Service-to-service communication** in microservices architectures

## Features

- **Bidirectional Streaming**: Real-time file upload/download with chunked transfer
- **Service Discovery**: Built-in service discovery and load balancing support
- **High Throughput**: Optimized for high-performance applications
- **Type Safety**: Strongly typed with protobuf definitions
- **Authentication**: Token-based authentication with metadata
- **Event Streaming**: Real-time events for file operations, peer status, and system metrics

## Quick Start

### Prerequisites

- Go 1.24+
- Protocol Buffers compiler (protoc) - for full implementation
- gRPC client libraries

### Running the Server

```bash
# Build the gRPC server
go build -o peervault-grpc ./cmd/peervault-grpc

# Run with default settings
./peervault-grpc

# Run with custom port and auth token
./peervault-grpc -port 8082 -auth-token your-secure-token
```

### Server Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8082` | gRPC server port |
| `-auth-token` | `demo-token` | Authentication token |

## API Services

### PeerVaultService

The main service providing all PeerVault operations:

#### File Operations

- `UploadFile(stream FileChunk) returns (FileResponse)` - Streaming file upload
- `DownloadFile(FileRequest) returns (stream FileChunk)` - Streaming file download
- `ListFiles(ListFilesRequest) returns (ListFilesResponse)` - List files with pagination
- `GetFile(FileRequest) returns (FileResponse)` - Get file metadata
- `DeleteFile(FileRequest) returns (DeleteFileResponse)` - Delete a file
- `UpdateFileMetadata(UpdateFileMetadataRequest) returns (FileResponse)` - Update file metadata

#### Peer Operations

- `ListPeers(Empty) returns (ListPeersResponse)` - List all peers
- `GetPeer(PeerRequest) returns (PeerResponse)` - Get peer information
- `AddPeer(AddPeerRequest) returns (PeerResponse)` - Add a new peer
- `RemovePeer(PeerRequest) returns (RemovePeerResponse)` - Remove a peer
- `GetPeerHealth(PeerRequest) returns (PeerHealthResponse)` - Get peer health status

#### System Operations

- `GetSystemInfo(Empty) returns (SystemInfoResponse)` - Get system information
- `GetMetrics(Empty) returns (MetricsResponse)` - Get system metrics
- `HealthCheck(Empty) returns (HealthResponse)` - Health check

#### Streaming Operations

- `StreamFileOperations(Empty) returns (stream FileOperationEvent)` - Real-time file operation events
- `StreamPeerEvents(Empty) returns (stream PeerEvent)` - Real-time peer status events
- `StreamSystemEvents(Empty) returns (stream SystemEvent)` - Real-time system events

## Authentication

The gRPC API uses token-based authentication via metadata:

```go
// Set authentication metadata
ctx := metadata.AppendToOutgoingContext(context.Background(), 
    "authorization", "Bearer your-token-here")

// Make authenticated request
response, err := client.GetFile(ctx, &FileRequest{Key: "file1"})
```

## Message Types

### File Types

```protobuf
message FileChunk {
    string file_key = 1;
    bytes data = 2;
    int64 offset = 3;
    bool is_last = 4;
    string checksum = 5;
}

message FileResponse {
    string key = 1;
    string name = 2;
    int64 size = 3;
    string content_type = 4;
    string hash = 5;
    google.protobuf.Timestamp created_at = 6;
    google.protobuf.Timestamp updated_at = 7;
    map<string, string> metadata = 8;
    repeated FileReplica replicas = 9;
}
```

### Peer Types

```protobuf
message PeerResponse {
    string id = 1;
    string address = 2;
    int32 port = 3;
    string status = 4;
    google.protobuf.Timestamp last_seen = 5;
    google.protobuf.Timestamp created_at = 6;
    map<string, string> metadata = 7;
}

message PeerHealthResponse {
    string peer_id = 1;
    string status = 2;
    double latency_ms = 3;
    int32 uptime_seconds = 4;
    map<string, string> metrics = 5;
}
```

### System Types

```protobuf
message SystemInfoResponse {
    string version = 1;
    int64 uptime_seconds = 2;
    google.protobuf.Timestamp start_time = 3;
    int64 storage_used = 4;
    int64 storage_total = 5;
    int32 peer_count = 6;
    int32 file_count = 7;
}

message MetricsResponse {
    int64 requests_total = 1;
    double requests_per_minute = 2;
    int32 active_connections = 3;
    double storage_usage_percent = 4;
    google.protobuf.Timestamp last_updated = 5;
}
```

## Examples

### File Upload (Streaming)

```go
func uploadFile(client PeerVaultServiceClient, filePath, fileKey string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    stream, err := client.UploadFile(context.Background())
    if err != nil {
        return err
    }

    buffer := make([]byte, 1024) // 1KB chunks
    offset := int64(0)

    for {
        n, err := file.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        chunk := &FileChunk{
            FileKey: fileKey,
            Data:    buffer[:n],
            Offset:  offset,
            IsLast:  false,
        }

        if err := stream.Send(chunk); err != nil {
            return err
        }

        offset += int64(n)
    }

    // Send final chunk
    finalChunk := &FileChunk{
        FileKey: fileKey,
        Data:    []byte{},
        Offset:  offset,
        IsLast:  true,
    }

    if err := stream.Send(finalChunk); err != nil {
        return err
    }

    response, err := stream.CloseAndRecv()
    if err != nil {
        return err
    }

    fmt.Printf("File uploaded: %s, size: %d\n", response.Key, response.Size)
    return nil
}
```

### File Download (Streaming)

```go
func downloadFile(client PeerVaultServiceClient, fileKey, outputPath string) error {
    req := &FileRequest{Key: fileKey}
    stream, err := client.DownloadFile(context.Background(), req)
    if err != nil {
        return err
    }

    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        if _, err := file.WriteAt(chunk.Data, chunk.Offset); err != nil {
            return err
        }

        if chunk.IsLast {
            break
        }
    }

    return nil
}
```

### Event Streaming

```go
func streamFileOperations(client PeerVaultServiceClient) error {
    stream, err := client.StreamFileOperations(context.Background(), &Empty{})
    if err != nil {
        return err
    }

    for {
        event, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        fmt.Printf("File operation: %s on %s by %s at %v\n",
            event.EventType, event.FileKey, event.PeerId, event.Timestamp)
    }

    return nil
}
```

## Error Handling

The gRPC API uses standard gRPC status codes:

- `OK` - Operation completed successfully
- `INVALID_ARGUMENT` - Invalid request parameters
- `NOT_FOUND` - Resource not found
- `PERMISSION_DENIED` - Authentication/authorization failed
- `INTERNAL` - Internal server error
- `UNAVAILABLE` - Service temporarily unavailable

## Performance Considerations

- **Chunk Size**: Use 1KB-64KB chunks for optimal streaming performance
- **Connection Pooling**: Reuse gRPC connections for better performance
- **Concurrent Streams**: The server supports up to 100 concurrent streams by default
- **Compression**: Enable gRPC compression for large file transfers

## Security

- **TLS**: Enable TLS for production deployments
- **Authentication**: Use strong authentication tokens
- **Rate Limiting**: Implement client-side rate limiting
- **Input Validation**: Validate all input data

## Integration

### Go Client

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/Skpow1234/Peervault/proto/peervault"
)

conn, err := grpc.Dial("localhost:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewPeerVaultServiceClient(conn)
```

### Other Languages

The gRPC API can be used with any language that supports gRPC:

- **Python**: Use `grpcio` and `grpcio-tools`
- **JavaScript/TypeScript**: Use `@grpc/grpc-js`
- **Java**: Use `grpc-java`
- **C#**: Use `Grpc.Net.Client`

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/Skpow1234/Peervault.git
cd Peervault

# Install dependencies
go mod tidy

# Build the gRPC server
go build -o peervault-grpc ./cmd/peervault-grpc

# Run tests
go test ./internal/api/grpc/...
```

### Protocol Buffer Generation

```bash
# Install protoc and plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/peervault.proto
```

## Status

**Current Status**: 🚧 **In Development**

The gRPC API is currently in development with the following status:

- ✅ **Basic Structure**: Server framework and service definitions
- ✅ **Type Definitions**: Protobuf message types
- ✅ **Service Interfaces**: File, peer, and system service interfaces
- 🚧 **Streaming Implementation**: Streaming operations in progress
- 🚧 **Authentication**: Basic token-based auth implemented
- 🚧 **Error Handling**: Standard gRPC error codes
- 📋 **Documentation**: API documentation and examples
- 📋 **Testing**: Integration and unit tests
- 📋 **Client Libraries**: Multi-language client examples

## Roadmap

- [ ] Complete streaming file operations
- [ ] Implement real-time event streaming
- [ ] Add TLS support
- [ ] Create client libraries for multiple languages
- [ ] Add comprehensive testing
- [ ] Performance optimization
- [ ] Production deployment guides

## Contributing

Contributions are welcome! Please see the main project README for contribution guidelines.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
