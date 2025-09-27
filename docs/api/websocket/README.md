# WebSocket API Documentation

The PeerVault WebSocket API provides real-time bidirectional communication for dynamic applications, enabling live updates, real-time messaging, and connection management.

## Overview

The WebSocket API is designed for applications that need:

- Real-time file operations notifications
- Live peer network status updates
- Instant system metrics and alerts
- Bidirectional communication with the PeerVault network

## Endpoints

### WebSocket Connection

- **URL**: `ws://localhost:8083/ws`
- **Protocol**: WebSocket
- **Authentication**: Optional (via connection headers)

### HTTP Endpoints

- **Health Check**: `GET /ws/health`
- **Metrics**: `GET /ws/metrics`
- **Status**: `GET /ws/status`

## Connection Management

### Establishing a Connection

```javascript
const ws = new WebSocket('ws://localhost:8083/ws');

ws.onopen = function(event) {
    console.log('Connected to PeerVault WebSocket API');
    
    // Subscribe to file operations
    ws.send(JSON.stringify({
        type: 'subscribe',
        topic: 'file_operations'
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received message:', message);
};

ws.onclose = function(event) {
    console.log('Connection closed:', event.code, event.reason);
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};
```

### Connection Configuration

The WebSocket server supports the following configuration options:

- **Read Timeout**: 30 seconds
- **Write Timeout**: 30 seconds
- **Ping Period**: 54 seconds
- **Pong Wait**: 60 seconds
- **Max Message Size**: 1MB

## Message Format

All WebSocket messages follow a consistent JSON format:

```json
{
    "type": "message_type",
    "topic": "optional_topic",
    "data": {},
    "timestamp": "2024-01-01T00:00:00Z",
    "clientId": "optional_client_id"
}
```

### Message Types

#### 1. Subscription Messages

#### Subscribe to Topic

```json
{
    "type": "subscribe",
    "topic": "file_operations"
}
```

#### Unsubscribe from Topic

```json
{
    "type": "unsubscribe",
    "topic": "file_operations"
}
```

#### 2. File Operation Messages

#### File Uploaded

```json
{
    "type": "file_uploaded",
    "topic": "file_operations",
    "data": {
        "key": "file_hash",
        "size": 1024,
        "node": "node_id",
        "timestamp": "2024-01-01T00:00:00Z"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### File Deleted

```json
{
    "type": "file_deleted",
    "topic": "file_operations",
    "data": {
        "key": "file_hash",
        "node": "node_id",
        "timestamp": "2024-01-01T00:00:00Z"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 3. Peer Network Messages

#### Peer Connected

```json
{
    "type": "peer_connected",
    "topic": "peer_network",
    "data": {
        "id": "peer_id",
        "address": "192.168.1.100",
        "port": 8080,
        "capabilities": ["storage", "compute"],
        "timestamp": "2024-01-01T00:00:00Z"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### Peer Disconnected

```json
{
    "type": "peer_disconnected",
    "topic": "peer_network",
    "data": {
        "id": "peer_id",
        "reason": "timeout",
        "timestamp": "2024-01-01T00:00:00Z"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 4. System Messages

#### System Metrics Update

```json
{
    "type": "system_metrics",
    "topic": "system",
    "data": {
        "storage": {
            "totalSpace": 1073741824,
            "usedSpace": 536870912,
            "availableSpace": 536870912
        },
        "network": {
            "activeConnections": 5,
            "totalBytesTransferred": 1048576
        },
        "performance": {
            "averageResponseTime": 0.1,
            "requestsPerSecond": 10.5
        }
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### Performance Alert

```json
{
    "type": "performance_alert",
    "topic": "alerts",
    "data": {
        "level": "warning",
        "message": "High memory usage detected",
        "metric": "memory_usage",
        "value": 85.5,
        "threshold": 80.0,
        "timestamp": "2024-01-01T00:00:00Z"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Available Topics

### File Operations (`file_operations`)

- `file_uploaded` - File has been uploaded to the network
- `file_deleted` - File has been deleted from the network
- `file_updated` - File metadata has been updated
- `file_replicated` - File has been replicated to another node

### Peer Network (`peer_network`)

- `peer_connected` - New peer has joined the network
- `peer_disconnected` - Peer has left the network
- `peer_health_changed` - Peer health status has changed
- `network_topology_changed` - Network topology has been updated

### System (`system`)

- `system_metrics` - System performance metrics update
- `storage_metrics` - Storage usage and capacity metrics
- `network_metrics` - Network performance metrics
- `performance_metrics` - Application performance metrics

### Alerts (`alerts`)

- `performance_alert` - Performance threshold exceeded
- `security_alert` - Security event detected
- `storage_alert` - Storage capacity warning
- `network_alert` - Network connectivity issue

## Error Handling

### Connection Errors

#### Connection Refused

```json
{
    "type": "error",
    "data": {
        "code": "CONNECTION_REFUSED",
        "message": "Server is not accepting connections"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### Authentication Failed

```json
{
    "type": "error",
    "data": {
        "code": "AUTH_FAILED",
        "message": "Invalid authentication credentials"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

#### Invalid Message Format

```json
{
    "type": "error",
    "data": {
        "code": "INVALID_MESSAGE",
        "message": "Message format is invalid"
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Rate Limiting

The WebSocket API implements rate limiting to prevent abuse:

- **Messages per second**: 100 messages per client
- **Connection limit**: 1000 concurrent connections
- **Message size limit**: 1MB per message

Rate limit exceeded responses:

```json
{
    "type": "error",
    "data": {
        "code": "RATE_LIMIT_EXCEEDED",
        "message": "Rate limit exceeded. Please slow down your requests.",
        "retryAfter": 60
    },
    "timestamp": "2024-01-01T00:00:00Z"
}
```

## Security

### Authentication

WebSocket connections can be authenticated using:

1. **Query Parameters**: `ws://localhost:8083/ws?token=your_token`
2. **Headers**: `Authorization: Bearer your_token`
3. **Connection Init Message**: Send auth message after connection

### CORS Configuration

The WebSocket server supports CORS with configurable origins:

- Default: Allow all origins (`*`)
- Production: Configure specific allowed origins

### Message Validation

All incoming messages are validated for:

- JSON format compliance
- Required fields presence
- Message size limits
- Topic subscription permissions

## Examples

### Real-time File Monitoring

```javascript
const ws = new WebSocket('ws://localhost:8083/ws');

ws.onopen = function() {
    // Subscribe to file operations
    ws.send(JSON.stringify({
        type: 'subscribe',
        topic: 'file_operations'
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'file_uploaded':
            console.log(`File uploaded: ${message.data.key}`);
            updateFileList(message.data);
            break;
        case 'file_deleted':
            console.log(`File deleted: ${message.data.key}`);
            removeFromFileList(message.data.key);
            break;
    }
};
```

### Peer Network Monitoring

```javascript
const ws = new WebSocket('ws://localhost:8083/ws');

ws.onopen = function() {
    // Subscribe to peer network events
    ws.send(JSON.stringify({
        type: 'subscribe',
        topic: 'peer_network'
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'peer_connected':
            addPeerToNetwork(message.data);
            break;
        case 'peer_disconnected':
            removePeerFromNetwork(message.data.id);
            break;
        case 'peer_health_changed':
            updatePeerHealth(message.data);
            break;
    }
};
```

### System Metrics Dashboard

```javascript
const ws = new WebSocket('ws://localhost:8083/ws');

ws.onopen = function() {
    // Subscribe to system metrics
    ws.send(JSON.stringify({
        type: 'subscribe',
        topic: 'system'
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    if (message.type === 'system_metrics') {
        updateDashboard(message.data);
    }
};
```

## Server Configuration

### Environment Variables

```bash
# WebSocket server configuration
PEERVAULT_WS_PORT=8083
PEERVAULT_WS_HOST=localhost
PEERVAULT_WS_CORS_ORIGINS=*
PEERVAULT_WS_READ_TIMEOUT=30s
PEERVAULT_WS_WRITE_TIMEOUT=30s
PEERVAULT_WS_PING_PERIOD=54s
PEERVAULT_WS_PONG_WAIT=60s
```

### Configuration File

```yaml
websocket:
  port: 8083
  host: localhost
  cors:
    enabled: true
    origins: ["*"]
  timeouts:
    read: 30s
    write: 30s
    ping_period: 54s
    pong_wait: 60s
  limits:
    max_connections: 1000
    max_message_size: 1048576  # 1MB
    messages_per_second: 100
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if the WebSocket server is running
   - Verify the port and host configuration
   - Check firewall settings

2. **Authentication Errors**
   - Verify authentication token is valid
   - Check token expiration
   - Ensure proper authentication method

3. **Message Format Errors**
   - Validate JSON format
   - Check required fields
   - Verify message size limits

4. **Rate Limiting**
   - Reduce message frequency
   - Implement client-side rate limiting
   - Check server rate limit configuration

### Debug Mode

Enable debug logging for troubleshooting:

```bash
go run ./cmd/peervault-websocket -verbose
```

This will provide detailed logs for:

- Connection events
- Message processing
- Error conditions
- Performance metrics

## Performance Considerations

### Connection Optimization

- Use connection pooling for multiple clients
- Implement reconnection logic with exponential backoff
- Monitor connection health with ping/pong

### Message Handling

- Batch multiple updates when possible
- Use message queuing for high-volume scenarios
- Implement client-side message filtering

### Resource Usage

- Monitor memory usage with many connections
- Implement connection limits per client
- Use message compression for large payloads

## Migration from HTTP APIs

When migrating from HTTP APIs to WebSocket:

1. **Replace polling with subscriptions**
2. **Implement real-time updates**
3. **Handle connection state management**
4. **Add error handling and reconnection logic**
5. **Optimize message frequency and size**
