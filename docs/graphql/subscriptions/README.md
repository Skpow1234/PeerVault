# GraphQL Subscriptions

This document describes the real-time GraphQL subscription features implemented in PeerVault.

## Overview

GraphQL subscriptions provide real-time updates for various events in the PeerVault system. They use WebSocket connections to deliver live updates to clients.

## WebSocket Endpoint

The GraphQL subscription WebSocket endpoint is available at:

```bash
ws://localhost:8080/ws
```

## Supported Subscriptions

### File Operations

#### fileUploaded

Subscribe to file upload events.

```graphql
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
    replicas {
      node {
        id
        address
      }
      status
    }
  }
}
```

#### fileDeleted

Subscribe to file deletion events.

```graphql
subscription {
  fileDeleted
}
```

#### fileUpdated

Subscribe to file update events.

```graphql
subscription {
  fileUpdated {
    id
    key
    size
    updatedAt
    metadata {
      contentType
      tags
    }
  }
}
```

### Peer Network Events

#### peerConnected

Subscribe to peer connection events.

```graphql
subscription {
  peerConnected {
    id
    address
    port
    status
    health {
      isHealthy
      responseTime
      uptime
    }
    capabilities
  }
}
```

#### peerDisconnected

Subscribe to peer disconnection events.

```graphql
subscription {
  peerDisconnected {
    id
    address
    port
    status
    lastSeen
  }
}
```

#### peerHealthChanged

Subscribe to peer health status changes.

```graphql
subscription {
  peerHealthChanged {
    isHealthy
    lastHeartbeat
    responseTime
    uptime
    errors
  }
}
```

### System Monitoring

#### systemMetricsUpdated

Subscribe to system metrics updates.

```graphql
subscription {
  systemMetricsUpdated {
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
      errorRate
    }
    performance {
      averageResponseTime
      requestsPerSecond
      errorRate
      memoryUsage
      cpuUsage
    }
    uptime
  }
}
```

#### performanceAlert

Subscribe to performance alerts.

```graphql
subscription {
  performanceAlert {
    type
    message
    severity
    timestamp
    metrics {
      averageResponseTime
      requestsPerSecond
      errorRate
      memoryUsage
      cpuUsage
    }
  }
}
```

## WebSocket Protocol

The WebSocket connection follows the GraphQL over WebSocket protocol:

### Connection Initialization

1. Client connects to WebSocket endpoint
2. Server sends `connection_init` message
3. Client responds with `connection_ack`

### Subscription Messages

#### Start Subscription

```json
{
  "id": "subscription_id",
  "type": "start",
  "payload": {
    "query": "subscription { fileUploaded { id key } }"
  }
}
```

#### Stop Subscription

```json
{
  "id": "subscription_id",
  "type": "stop"
}
```

#### Data Message

```json
{
  "id": "subscription_id",
  "type": "data",
  "payload": {
    "data": {
      "fileUploaded": {
        "id": "file_123",
        "key": "example.txt"
      }
    }
  }
}
```

## Client Examples

### JavaScript/TypeScript

```typescript
import { createClient } from 'graphql-ws';

const client = createClient({
  url: 'ws://localhost:8080/ws',
});

// Subscribe to file uploads
const unsubscribe = client.subscribe(
  {
    query: `
      subscription {
        fileUploaded {
          id
          key
          size
          createdAt
        }
      }
    `,
  },
  {
    next: (data) => {
      console.log('File uploaded:', data.data.fileUploaded);
    },
    error: (err) => {
      console.error('Subscription error:', err);
    },
    complete: () => {
      console.log('Subscription completed');
    },
  }
);

// Unsubscribe when done
unsubscribe();
```

### Go Client

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/url"
    "time"

    "github.com/gorilla/websocket"
)

func main() {
    u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
    
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    defer conn.Close()

    // Send subscription
    subscription := map[string]interface{}{
        "id": "1",
        "type": "start",
        "payload": map[string]interface{}{
            "query": "subscription { fileUploaded { id key size } }",
        },
    }
    
    if err := conn.WriteJSON(subscription); err != nil {
        log.Fatal("write:", err)
    }

    // Read messages
    for {
        var message map[string]interface{}
        if err := conn.ReadJSON(&message); err != nil {
            log.Println("read:", err)
            return
        }
        
        fmt.Printf("Received: %+v\n", message)
    }
}
```

## Configuration

The WebSocket server can be configured in the GraphQL server configuration:

```go
config := &graphql.Config{
    Port:             8080,
    WebSocketPath:    "/ws",
    EnableWebSocket:  true,
    // ... other config
}
```

## Performance Considerations

- WebSocket connections are kept alive with ping/pong messages every 54 seconds
- Each client can have multiple active subscriptions
- Subscription data is broadcast to all subscribed clients for each topic
- Connection timeouts are set to 60 seconds for read operations

## Security

- CORS is enabled for WebSocket connections
- Origin checking can be configured in the WebSocket upgrader
- Client authentication can be added through connection initialization

## Monitoring

The WebSocket hub provides metrics:

- Total connected clients
- Subscriptions per topic
- Active topics
- Connection status

Access metrics at: `GET /metrics`
