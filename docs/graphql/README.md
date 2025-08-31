# PeerVault GraphQL API

The PeerVault GraphQL API provides a flexible interface for interacting with the distributed storage system. It offers real-time queries, mutations, and subscriptions for file operations, peer management, and system monitoring.

## Features

- **File Operations**: Upload, download, delete, and query files
- **Peer Management**: Add, remove, and monitor peer nodes
- **System Monitoring**: Real-time metrics and health checks
- **Real-time Subscriptions**: Live updates for file and peer events
- **Interactive Playground**: Built-in GraphQL Playground for testing

## Quick Start

### Running the GraphQL Server

```bash
# Build the GraphQL server
go build -o peervault-graphql.exe cmd/peervault-graphql/main.go

# Run with default settings
./peervault-graphql.exe

# Run with custom settings
./peervault-graphql.exe -port 8080 -storage ./data -playground=true
```

### Command Line Options

- `-port`: Port to listen on (default: 8080)
- `-storage`: Storage root directory (default: ./storage)
- `-bootstrap`: Comma-separated list of bootstrap nodes
- `-playground`: Enable GraphQL Playground (default: true)
- `-log-level`: Log level (debug, info, warn, error)

## API Endpoints

### GraphQL Endpoint

- **URL**: `POST /graphql`
- **Content-Type**: `application/json`
- **Description**: Main GraphQL endpoint for queries, mutations, and subscriptions

### GraphQL Playground

- **URL**: `GET /playground`
- **Description**: Interactive GraphQL IDE for testing queries

### Health Check

- **URL**: `GET /health`
- **Description**: System health status

### Metrics

- **URL**: `GET /metrics`
- **Description**: System metrics and statistics

## Schema Overview

### Queries

#### File Operations

```graphql
# Get a single file by key
query {
  file(key: "example.txt") {
    id
    key
    size
    createdAt
    owner {
      id
      address
    }
  }
}

# List files with filtering
query {
  files(limit: 10, offset: 0, filter: {
    sizeMin: 1024
    tags: ["important", "backup"]
  }) {
    id
    key
    size
    createdAt
  }
}

# Get file metadata
query {
  fileMetadata(key: "example.txt") {
    contentType
    checksum
    tags
  }
}
```

#### Peer Network

```graphql
# Get peer network information
query {
  peerNetwork {
    nodes {
      id
      address
      port
      status
      health {
        isHealthy
        lastHeartbeat
      }
    }
    connections {
      from { id address }
      to { id address }
      status
      latency
    }
    topology {
      totalNodes
      connectedNodes
      averageLatency
    }
  }
}

# Get specific node
query {
  node(id: "node-123") {
    id
    address
    port
    status
    capabilities
  }
}
```

#### System Monitoring

```graphql
# Get system metrics
query {
  systemMetrics {
    storage {
      totalSpace
      usedSpace
      availableSpace
      fileCount
    }
    network {
      activeConnections
      totalBytesTransferred
      averageBandwidth
    }
    performance {
      averageResponseTime
      requestsPerSecond
      memoryUsage
      cpuUsage
    }
    uptime
  }
}

# Get health status
query {
  health {
    status
    timestamp
    details
  }
}
```

### Mutations

#### Files Operations

```graphql
# Upload a file
mutation {
  uploadFile(
    file: "file-upload-data"
    key: "example.txt"
    metadata: {
      contentType: "text/plain"
      tags: ["important", "backup"]
    }
  ) {
    id
    key
    size
    status
    progress
  }
}

# Delete a file
mutation {
  deleteFile(key: "example.txt")
}

# Update file metadata
mutation {
  updateFileMetadata(
    key: "example.txt"
    metadata: {
      tags: ["updated", "important"]
    }
  ) {
    contentType
    tags
  }
}
```

#### Peer Management

```graphql
# Add a peer
mutation {
  addPeer(address: "192.168.1.100", port: 3000) {
    id
    address
    port
    status
  }
}

# Remove a peer
mutation {
  removePeer(id: "node-123")
}
```

#### Configuration

```graphql
# Update system configuration
mutation {
  updateConfiguration(config: {
    storageRoot: "/new/storage/path"
    replicationFactor: 3
    maxFileSize: 1073741824
    encryptionEnabled: true
  })
}
```

### Subscriptions

#### Real-time File Events

```graphql
# Subscribe to file uploads
subscription {
  fileUploaded {
    id
    key
    size
    createdAt
    owner {
      id
      address
    }
  }
}

# Subscribe to file deletions
subscription {
  fileDeleted
}

# Subscribe to file updates
subscription {
  fileUpdated {
    id
    key
    size
    updatedAt
  }
}
```

#### Real-time Peer Events

```graphql
# Subscribe to peer connections
subscription {
  peerConnected {
    id
    address
    port
    status
  }
}

# Subscribe to peer disconnections
subscription {
  peerDisconnected {
    id
    address
    port
  }
}

# Subscribe to peer health changes
subscription {
  peerHealthChanged {
    isHealthy
    lastHeartbeat
    responseTime
    errors
  }
}
```

#### Real-time System Events

```graphql
# Subscribe to system metrics updates
subscription {
  systemMetricsUpdated {
    storage {
      usedSpace
      fileCount
    }
    network {
      activeConnections
      totalBytesTransferred
    }
    performance {
      averageResponseTime
      memoryUsage
    }
  }
}

# Subscribe to performance alerts
subscription {
  performanceAlert {
    type
    message
    severity
    timestamp
    metrics {
      averageResponseTime
      memoryUsage
      cpuUsage
    }
  }
}
```

## Examples

### JavaScript/TypeScript Client

```javascript
// Using fetch
const response = await fetch('http://localhost:8080/graphql', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    query: `
      query {
        health {
          status
          timestamp
        }
      }
    `
  })
});

const data = await response.json();
console.log(data.data.health);
```

### cURL Example

```bash
# Health check
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ health { status timestamp } }"
  }'

# Get files
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "{ files(limit: 5) { id key size createdAt } }"
  }'
```

## Error Handling

GraphQL responses include an `errors` field when something goes wrong:

```json
{
  "data": null,
  "errors": [
    {
      "message": "File not found",
      "locations": [{"line": 2, "column": 3}],
      "path": ["file"]
    }
  ]
}
```

## CORS Support

The GraphQL API includes CORS support for cross-origin requests. All origins are allowed by default, but this can be configured for production use.

## Security Considerations

- The current implementation allows all origins for CORS
- No authentication is implemented yet
- File uploads are limited to 32MB by default
- Consider implementing rate limiting for production use

## Future Enhancements

- Authentication and authorization
- Rate limiting
- File upload progress tracking
- Advanced filtering and search
- Batch operations
- Schema introspection
- Persisted queries
- Query complexity analysis

## Troubleshooting

### Common Issues

1. **Port already in use**: Change the port using the `-port` flag
2. **Storage permission errors**: Ensure the storage directory is writable
3. **CORS errors**: Check that the client is making requests to the correct endpoint
4. **File upload failures**: Verify file size limits and storage space

### Logs

The server logs important events and errors. Use the `-log-level` flag to control verbosity:

```bash
./peervault-graphql.exe -log-level debug
```

### Health Checks

Monitor the health endpoint to ensure the service is running:

```bash
curl http://localhost:8080/health
```
