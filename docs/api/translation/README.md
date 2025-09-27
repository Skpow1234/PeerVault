# Protocol Translation API

The Protocol Translation API provides seamless cross-protocol communication between different API types in PeerVault. It enables real-time message translation between WebSocket, Server-Sent Events (SSE), MQTT, and CoAP protocols.

## Overview

The Protocol Translation API acts as a bridge between different communication protocols, allowing clients using one protocol to communicate with services using another protocol. This is particularly useful in IoT environments, web applications, and distributed systems where different components may use different communication protocols.

## Features

- **Cross-Protocol Translation**: Translate messages between WebSocket, SSE, MQTT, and CoAP
- **Real-Time Processing**: Low-latency message translation with configurable timeouts
- **Protocol Bridging**: Seamless communication between different protocol endpoints
- **Message Transformation**: Intelligent mapping of message types, headers, and metadata
- **Analytics & Monitoring**: Comprehensive translation analytics and performance metrics
- **Error Handling**: Robust error handling with retry mechanisms
- **Connection Management**: Automatic connection pooling and management

## Supported Protocol Translations

### WebSocket Translations

- WebSocket → SSE
- WebSocket → MQTT
- WebSocket → CoAP

### SSE Translations

- SSE → WebSocket
- SSE → MQTT
- SSE → CoAP

### MQTT Translations

- MQTT → WebSocket
- MQTT → SSE
- MQTT → CoAP

### CoAP Translations

- CoAP → WebSocket
- CoAP → SSE
- CoAP → MQTT

## API Endpoints

### Base URL

```bash
http://localhost:8086
```

### Endpoints

#### General Translation

- `POST /translate` - Translate messages between any supported protocols

#### Protocol-Specific Translation

- `POST /translate/websocket` - Translate from WebSocket to other protocols
- `POST /translate/sse` - Translate from SSE to other protocols
- `POST /translate/mqtt` - Translate from MQTT to other protocols
- `POST /translate/coap` - Translate from CoAP to other protocols

#### Analytics & Monitoring

- `GET /translate/analytics` - Get translation analytics and metrics
- `GET /translate/health` - Health check endpoint

## Message Format

### Translation Request

```json
{
  "from_protocol": "websocket",
  "to_protocol": "mqtt",
  "message": {
    "id": "msg_1234567890",
    "protocol": "websocket",
    "type": "text",
    "topic": "sensors/temperature",
    "payload": "25.5",
    "headers": {
      "Content-Type": "text/plain"
    },
    "metadata": {
      "qos": 1,
      "retain": false
    },
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

### Translation Response

```json
{
  "success": true,
  "message": {
    "id": "msg_1234567890",
    "protocol": "mqtt",
    "type": "publish",
    "topic": "ws/sensors/temperature",
    "payload": "25.5",
    "headers": {
      "content_type": "text/plain"
    },
    "metadata": {
      "mqtt_qos": 1,
      "mqtt_retain": false,
      "mqtt_dup": false
    },
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "engine": "websocket-to-mqtt"
}
```

## Protocol-Specific Translation CURL

### WebSocket to MQTT

```bash
curl -X POST http://localhost:8086/translate/websocket \
  -H "Content-Type: application/json" \
  -d '{
    "to_protocol": "mqtt",
    "type": "text",
    "topic": "sensors/temperature",
    "payload": "25.5",
    "metadata": {
      "qos": 1,
      "retain": false
    }
  }'
```

### SSE to WebSocket

```bash
curl -X POST http://localhost:8086/translate/sse \
  -H "Content-Type: application/json" \
  -d '{
    "to_protocol": "websocket",
    "type": "data",
    "topic": "notifications",
    "payload": "New message received",
    "metadata": {
      "sse_event": "message",
      "sse_id": "12345"
    }
  }'
```

### MQTT to CoAP

```bash
curl -X POST http://localhost:8086/translate/mqtt \
  -H "Content-Type: application/json" \
  -d '{
    "to_protocol": "coap",
    "type": "publish",
    "topic": "sensors/humidity",
    "payload": "60.2",
    "metadata": {
      "mqtt_qos": 0,
      "mqtt_retain": true
    }
  }'
```

## Message Type Mapping

### WebSocket Message Types

- `text` → MQTT: `publish`, SSE: `data`, CoAP: `request`
- `binary` → MQTT: `publish`, SSE: `data`, CoAP: `request`
- `ping` → MQTT: `pingreq`, SSE: `ping`, CoAP: `ping`
- `pong` → MQTT: `pingresp`, SSE: `pong`, CoAP: `pong`
- `close` → MQTT: `disconnect`, SSE: `close`, CoAP: `reset`
- `error` → MQTT: `disconnect`, SSE: `error`, CoAP: `reset`

### MQTT Message Types

- `publish` → WebSocket: `text`, SSE: `data`, CoAP: `request`
- `subscribe` → WebSocket: `subscribe`, SSE: `data`, CoAP: `observe`
- `unsubscribe` → WebSocket: `unsubscribe`, SSE: `data`, CoAP: `unobserve`
- `pingreq` → WebSocket: `ping`, SSE: `ping`, CoAP: `ping`
- `pingresp` → WebSocket: `pong`, SSE: `pong`, CoAP: `pong`
- `disconnect` → WebSocket: `close`, SSE: `close`, CoAP: `reset`

### SSE Message Types

- `data` → WebSocket: `text`, MQTT: `publish`, CoAP: `request`
- `ping` → WebSocket: `ping`, MQTT: `pingreq`, CoAP: `ping`
- `pong` → WebSocket: `pong`, MQTT: `pingresp`, CoAP: `pong`
- `close` → WebSocket: `close`, MQTT: `disconnect`, CoAP: `reset`
- `error` → WebSocket: `error`, MQTT: `disconnect`, CoAP: `reset`

### CoAP Message Types

- `request` → WebSocket: `text`, MQTT: `publish`, SSE: `data`
- `response` → WebSocket: `text`, MQTT: `publish`, SSE: `data`
- `ping` → WebSocket: `ping`, MQTT: `pingreq`, SSE: `ping`
- `pong` → WebSocket: `pong`, MQTT: `pingresp`, SSE: `pong`
- `reset` → WebSocket: `close`, MQTT: `disconnect`, SSE: `close`

## Topic Mapping

### WebSocket Topics

- WebSocket topics are prefixed when translating to other protocols:
  - To MQTT: `ws/{original_topic}`
  - To CoAP: `/ws{original_topic}`

### MQTT Topics

- MQTT topics are prefixed when translating to other protocols:
  - To WebSocket: Remove `ws/` prefix if present
  - To CoAP: `/mqtt{original_topic}`

### SSE Topics

- SSE topics are prefixed when translating to other protocols:
  - To WebSocket: No prefix change
  - To MQTT: `sse/{original_topic}`
  - To CoAP: `/sse{original_topic}`

### CoAP Topics

- CoAP URIs are prefixed when translating to other protocols:
  - To WebSocket: Remove `/ws` prefix if present
  - To MQTT: `coap{original_uri}`

## QoS and Reliability Mapping

### MQTT QoS Levels

- QoS 0 → CoAP: `NON` (Non-confirmable)
- QoS 1 → CoAP: `CON` (Confirmable)
- QoS 2 → CoAP: `CON` (Confirmable)

### CoAP Reliability

- `CON` (Confirmable) → MQTT: QoS 1
- `NON` (Non-confirmable) → MQTT: QoS 0

## Analytics

### Translation Analytics

```bash
curl http://localhost:8086/translate/analytics
```

Response:

```json
{
  "summary": {
    "total_translations": 1250,
    "total_errors": 15,
    "success_rate": 98.8,
    "total_bytes_translated": 1048576,
    "active_engines": 12,
    "active_protocols": 4,
    "uptime": "2h30m15s"
  },
  "translation_engines": {
    "websocket-to-mqtt": {
      "engine_name": "websocket-to-mqtt",
      "from_protocol": "websocket",
      "to_protocol": "mqtt",
      "total_requests": 450,
      "successful_translations": 445,
      "failed_translations": 5,
      "total_bytes_in": 45000,
      "total_bytes_out": 45000,
      "avg_latency": "2.5ms",
      "max_latency": "15ms",
      "min_latency": "1ms",
      "last_activity": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-15T08:00:00Z"
    }
  },
  "protocol_stats": {
    "websocket": {
      "protocol": "websocket",
      "total_requests": 600,
      "total_bytes": 60000,
      "avg_latency": "2.1ms",
      "error_rate": 1.2,
      "last_activity": "2024-01-15T10:30:00Z"
    }
  },
  "performance": {
    "avg_translation_time": "2.3ms",
    "max_translation_time": "20ms",
    "min_translation_time": "0.5ms",
    "p95_latency": "5ms",
    "p99_latency": "10ms",
    "throughput_per_second": 125.5
  },
  "error_analysis": {
    "total_errors": 15,
    "error_rate": 1.2,
    "error_types": {
      "connection_timeout": 8,
      "invalid_message_format": 4,
      "protocol_mismatch": 3
    },
    "most_common_error": "connection_timeout",
    "last_error_time": "2024-01-15T10:25:00Z"
  },
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": "Translation from websocket to unsupported_protocol not supported",
  "engine": "websocket-to-unsupported_protocol"
}
```

### Common Error Types

- `connection_timeout`: Connection to target protocol server timed out
- `invalid_message_format`: Message format is invalid for the source protocol
- `protocol_mismatch`: Requested translation is not supported
- `target_server_unavailable`: Target protocol server is not available
- `message_too_large`: Message exceeds size limits for target protocol

## Configuration

### Server Configuration

```yaml
translation:
  port: 8086
  host: localhost
  max_connections: 1000
  read_timeout: 30s
  write_timeout: 30s
  
  # Protocol endpoints
  websocket_addr: localhost:8083
  sse_addr: localhost:8084
  mqtt_addr: localhost:1883
  coap_addr: localhost:5683
  
  # Translation settings
  enable_analytics: true
  buffer_size: 1024
  retry_attempts: 3
  retry_delay: 1s
```

### Environment Variables

```bash
TRANSLATION_PORT=8086
TRANSLATION_HOST=localhost
TRANSLATION_WEBSOCKET_ADDR=localhost:8083
TRANSLATION_SSE_ADDR=localhost:8084
TRANSLATION_MQTT_ADDR=localhost:1883
TRANSLATION_COAP_ADDR=localhost:5683
TRANSLATION_ENABLE_ANALYTICS=true
TRANSLATION_BUFFER_SIZE=1024
TRANSLATION_RETRY_ATTEMPTS=3
TRANSLATION_RETRY_DELAY=1s
```

## Security

### Authentication

- API key authentication (optional)
- JWT token validation (optional)
- IP whitelisting (optional)

### Rate Limiting

- Per-client rate limiting
- Per-protocol rate limiting
- Global rate limiting

### Data Privacy

- Message payload encryption (optional)
- Header sanitization
- Metadata filtering

## Performance

### Optimization Features

- Connection pooling
- Message batching
- Compression support
- Caching mechanisms

### Monitoring

- Real-time metrics
- Performance dashboards
- Alerting system
- Health checks

## Use Cases

### IoT Device Communication

- MQTT devices communicating with web applications via WebSocket
- CoAP sensors sending data to MQTT brokers
- SSE dashboards receiving MQTT sensor data

### Web Application Integration

- Real-time chat applications using WebSocket with MQTT backend
- Live data feeds using SSE with CoAP data sources
- Cross-platform messaging systems

### Microservices Communication

- Service-to-service communication across different protocols
- Protocol translation for legacy system integration
- Real-time event streaming between services

## Examples

### JavaScript Client

```javascript
// Translate WebSocket message to MQTT
async function translateWebSocketToMQTT(message) {
  const response = await fetch('http://localhost:8086/translate/websocket', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      to_protocol: 'mqtt',
      type: 'text',
      topic: 'sensors/temperature',
      payload: message,
      metadata: {
        qos: 1,
        retain: false
      }
    })
  });
  
  const result = await response.json();
  return result.message;
}

// Translate MQTT message to SSE
async function translateMQTTToSSE(message) {
  const response = await fetch('http://localhost:8086/translate/mqtt', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      to_protocol: 'sse',
      type: 'publish',
      topic: 'notifications',
      payload: message,
      metadata: {
        mqtt_qos: 0,
        mqtt_retain: true
      }
    })
  });
  
  const result = await response.json();
  return result.message;
}
```

### Python Client

```python
import requests
import json

def translate_websocket_to_mqtt(message):
    """Translate WebSocket message to MQTT format"""
    url = 'http://localhost:8086/translate/websocket'
    data = {
        'to_protocol': 'mqtt',
        'type': 'text',
        'topic': 'sensors/temperature',
        'payload': message,
        'metadata': {
            'qos': 1,
            'retain': False
        }
    }
    
    response = requests.post(url, json=data)
    result = response.json()
    return result['message']

def translate_mqtt_to_coap(message):
    """Translate MQTT message to CoAP format"""
    url = 'http://localhost:8086/translate/mqtt'
    data = {
        'to_protocol': 'coap',
        'type': 'publish',
        'topic': 'sensors/humidity',
        'payload': message,
        'metadata': {
            'mqtt_qos': 0,
            'mqtt_retain': True
        }
    }
    
    response = requests.post(url, json=data)
    result = response.json()
    return result['message']

# Example usage
ws_message = "25.5"
mqtt_message = translate_websocket_to_mqtt(ws_message)
print(f"Translated to MQTT: {mqtt_message}")

coap_message = translate_mqtt_to_coap(mqtt_message['payload'])
print(f"Translated to CoAP: {coap_message}")
```

## Troubleshooting

### Common Issues

#### Connection Timeouts

- Check if target protocol servers are running
- Verify network connectivity
- Increase timeout values in configuration

#### Invalid Message Format

- Ensure message payload is valid for source protocol
- Check message type mapping
- Verify required headers are present

#### Protocol Mismatch

- Verify requested translation is supported
- Check protocol endpoint configuration
- Ensure target protocol server is accessible

### Debug Mode

Enable debug logging for detailed translation information:

```bash
./bin/peervault-translation.exe -verbose
```

### Health Check

Monitor server health:

```bash
curl http://localhost:8086/translate/health
```

## Contributing

### Adding New Protocol Support

1. Implement the `Translator` interface
2. Add protocol-specific message mapping
3. Update configuration and documentation
4. Add tests and examples

### Testing

```bash
# Run translation tests
go test ./internal/api/translation/...

# Run integration tests
go test ./tests/integration/translation/...
```

## License

This API is part of the PeerVault project and is licensed under the same terms as the main project.
