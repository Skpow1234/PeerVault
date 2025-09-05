# PeerVault Go SDK

Official Go SDK for PeerVault distributed file storage system.

## Installation

```bash
go get github.com/peervault/peervault-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/peervault/peervault-sdk-go"
)

func main() {
    // Create client
    client, err := peervault.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    
    // Upload file
    file, err := os.Open("example.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    result, err := client.Files.Upload(context.Background(), "my-file", file)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Uploaded file: %s (size: %d)\n", result.Key, result.Size)
}
```

## Authentication

```go
// Login to get token
auth, err := client.Auth.Login(context.Background(), "username", "password")
if err != nil {
    log.Fatal(err)
}

// Set token for subsequent requests
client.SetToken(auth.Token)
```

## File Operations

### Upload File

```go
// From file
file, _ := os.Open("example.txt")
defer file.Close()
result, err := client.Files.Upload(ctx, "my-key", file)

// From bytes
content := []byte("file content")
result, err := client.Files.UploadBytes(ctx, "my-key", content)

// From reader
reader := strings.NewReader("file content")
result, err := client.Files.UploadReader(ctx, "my-key", reader)
```

### Download File

```go
// To file
file, _ := os.Create("downloaded.txt")
defer file.Close()
err := client.Files.Download(ctx, "my-key", file)

// To bytes
content, err := client.Files.DownloadBytes(ctx, "my-key")

// To reader
reader, err := client.Files.DownloadReader(ctx, "my-key")
```

### List Files

```go
files, err := client.Files.List(ctx, &peervault.ListOptions{
    Limit:  10,
    Offset: 0,
    Prefix: "documents/",
})

for _, file := range files {
    fmt.Printf("File: %s (size: %d)\n", file.Key, file.Size)
}
```

### Delete File

```go
err := client.Files.Delete(ctx, "my-key")
```

## Peer Operations

### List Peers

```go
peers, err := client.Peers.List(ctx)
for _, peer := range peers {
    fmt.Printf("Peer: %s (status: %s)\n", peer.ID, peer.Status)
}
```

### Get Peer Health

```go
health, err := client.Peers.GetHealth(ctx, "peer-id")
fmt.Printf("Health: %s (uptime: %s)\n", health.Status, health.Uptime)
```

## GraphQL Client

```go
// Create GraphQL client
gqlClient := peervault.NewGraphQLClient("http://localhost:8081")

// Execute query
var result struct {
    Files []struct {
        Key  string `json:"key"`
        Size int64  `json:"size"`
    } `json:"files"`
}

err := gqlClient.Query(ctx, `
    query {
        files {
            key
            size
        }
    }
`, &result)
```

## gRPC Client

```go
// Create gRPC client
grpcClient, err := peervault.NewGRPCClient("localhost:9090")
if err != nil {
    log.Fatal(err)
}

// Upload file via gRPC
stream, err := grpcClient.UploadFile(ctx)
if err != nil {
    log.Fatal(err)
}

// Send file data
file, _ := os.Open("example.txt")
defer file.Close()

buf := make([]byte, 1024)
for {
    n, err := file.Read(buf)
    if err == io.EOF {
        break
    }
    
    stream.Send(&pb.UploadRequest{
        Data: buf[:n],
    })
}

response, err := stream.CloseAndRecv()
```

## Configuration

```go
client, err := peervault.NewClient("http://localhost:8080", &peervault.Config{
    Timeout:     30 * time.Second,
    Retries:     3,
    RateLimit:   100, // requests per minute
    UserAgent:   "my-app/1.0",
})
```

## Error Handling

```go
result, err := client.Files.Upload(ctx, "key", file)
if err != nil {
    switch e := err.(type) {
    case *peervault.FileNotFoundError:
        fmt.Printf("File not found: %s\n", e.Key)
    case *peervault.RateLimitError:
        fmt.Printf("Rate limited: retry after %s\n", e.RetryAfter)
    case *peervault.AuthError:
        fmt.Printf("Authentication failed: %s\n", e.Message)
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

## Streaming

```go
// Stream upload
uploader := client.Files.NewUploader(ctx, "large-file")
defer uploader.Close()

file, _ := os.Open("large-file.bin")
defer file.Close()

buf := make([]byte, 64*1024) // 64KB chunks
for {
    n, err := file.Read(buf)
    if err == io.EOF {
        break
    }
    
    if err := uploader.Write(buf[:n]); err != nil {
        log.Fatal(err)
    }
}

result, err := uploader.Finish()
```

## Webhooks

```go
// Register webhook
webhook, err := client.Webhooks.Create(ctx, &peervault.WebhookConfig{
    URL:    "https://your-app.com/webhook",
    Events: []string{"file.uploaded", "file.deleted"},
    Secret: "webhook-secret",
})

// Verify webhook signature
func verifyWebhook(payload []byte, signature string, secret string) bool {
    return peervault.VerifyWebhookSignature(payload, signature, secret)
}
```

## Examples

See the [examples directory](examples/) for complete working examples:

- [Basic file operations](examples/basic/)
- [Authentication](examples/auth/)
- [GraphQL queries](examples/graphql/)
- [gRPC streaming](examples/grpc/)
- [Webhook handling](examples/webhooks/)
- [Error handling](examples/errors/)
