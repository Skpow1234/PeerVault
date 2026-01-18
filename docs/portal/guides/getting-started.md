# Getting Started with PeerVault

This guide will help you get up and running with PeerVault in just a few minutes. We'll cover installation, basic configuration, and your first API call.

## Prerequisites

- Docker

## Installation

### Docker (Required)

The Docker-only workflow uses the repo compose files:

```bash
docker compose -f docker-compose.apis.yml up -d --build
```

## Configuration

Create a configuration file `config.yaml`:

```yaml
# Server configuration
server:
  rest:
    port: 8080
    host: "0.0.0.0"
  graphql:
    port: 8081
    host: "0.0.0.0"
  grpc:
    port: 9090
    host: "0.0.0.0"

# Storage configuration
storage:
  root: "/data"
  max_file_size: "1GB"
  replication_factor: 3

# Security configuration
security:
  jwt_secret: "your-secret-key"
  encryption_key: "your-encryption-key"

# Logging configuration
logging:
  level: "info"
  format: "json"
```

## Your First API Call

### Using a REST Client

Use any REST client to call the API endpoints under `http://localhost:8080/api/v1`.

### Using Go SDK

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/peervault/peervault-sdk-go"
)

func main() {
    // Create client
    client, err := peervault.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    
    // Upload a file
    file, err := os.Open("example.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    result, err := client.Files.Upload(context.Background(), "my-first-file", file)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Uploaded file: %s (size: %d)\n", result.Key, result.Size)
}
```

### Using JavaScript SDK

```javascript
import { PeerVaultClient } from '@peervault/sdk';

async function main() {
    // Create client
    const client = new PeerVaultClient({
        baseURL: 'http://localhost:8080'
    });
    
    // Upload a file
    const file = new File(['Hello, PeerVault!'], 'example.txt');
    const result = await client.files.upload('my-first-file', file);
    
    console.log(`Uploaded file: ${result.key} (size: ${result.size})`);
}

main().catch(console.error);
```

### Using Python SDK

```python
from peervault import PeerVaultClient

def main():
    # Create client
    client = PeerVaultClient('http://localhost:8080')
    
    # Upload a file
    with open('example.txt', 'rb') as file:
        result = client.files.upload('my-first-file', file)
    
    print(f"Uploaded file: {result.key} (size: {result.size})")

if __name__ == '__main__':
    main()
```

## GraphQL Example

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
    key: "my-graphql-file"
    content: "Hello from GraphQL!"
  }) {
    key
    size
    success
  }
}

# Subscribe to file events
subscription {
  fileEvents {
    type
    file {
      key
      size
    }
    timestamp
  }
}
```

## gRPC Example

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "github.com/peervault/peervault/proto"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Create client
    client := pb.NewFileServiceClient(conn)
    
    // Upload file
    stream, err := client.UploadFile(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // Send file data
    data := []byte("Hello from gRPC!")
    stream.Send(&pb.UploadRequest{
        Key:  "my-grpc-file",
        Data: data,
    })
    
    response, err := stream.CloseAndRecv()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Uploaded: %s", response.Key)
}
```

## Next Steps

Now that you have PeerVault running and have made your first API call, here's what you can do next:

1. **Explore the APIs**:
   - [REST API Documentation](../api/peervault-rest-api.yaml)
   - [GraphQL Playground](../graphql-playground/)
   - [Swagger UI](../swagger/)

2. **Learn about file operations**:
   - [File Operations Guide](file-operations.md)
   - [Authentication Guide](authentication.md)
   - [Error Handling Guide](error-handling.md)

3. **Build your application**:
   - [Web Application Examples](../../examples/web/)
   - [Mobile Application Examples](../../examples/mobile/)
   - [Backend Service Examples](../../examples/backend/)

4. **Join the community**:
   - [GitHub Repository](https://github.com/peervault/peervault)
   - [Discord Community](https://discord.gg/peervault)
   - [Stack Overflow](https://stackoverflow.com/questions/tagged/peervault)

## Troubleshooting

### Common Issues

**Server won't start**:

- Check if ports 8080, 8081, and 9090 are available
- Verify your configuration file is valid YAML
- Check the logs for error messages

**Connection refused**:

- Ensure the server is running
- Check the correct port and host
- Verify firewall settings

**Authentication errors**:

- Check your JWT token is valid
- Verify the token hasn't expired
- Ensure you're using the correct secret key

### Getting Help

- Check the [FAQ](../support/faq.md)
- Review the [Troubleshooting Guide](../support/troubleshooting.md)
- Join our [Discord Community](https://discord.gg/peervault)
- Open an issue on [GitHub](https://github.com/peervault/peervault/issues)

---

Congratulations! You've successfully set up PeerVault and made your first API call. You're now ready to build amazing applications with distributed file storage!
