# Structured Logging

## Overview

This document describes the structured logging system implemented in the PeerVault distributed storage system.

## Logging Architecture

### Structured Logging with slog

The application uses Go's built-in `log/slog` package for structured logging, providing:

- **JSON Output**: Machine-readable log format for production
- **Log Levels**: Configurable logging levels (debug, info, warn, error)
- **Context Fields**: Rich metadata with each log entry
- **Source Location**: File and line information for debugging

## Configuration

### Environment Variables

```bash
# Set log level (debug, info, warn, error)
export LOG_LEVEL=info

# Default: info
```

### Log Levels

- **DEBUG**: Detailed debugging information
- **INFO**: General operational information
- **WARN**: Warning messages for potential issues
- **ERROR**: Error conditions that need attention

## Usage

### Basic Logging

```go
import "log/slog"

// Info level logging
slog.Info("file stored", "key", "example.txt", "bytes", 1024)

// Error level logging
slog.Error("connection failed", "peer", "127.0.0.1:3000", "error", err.Error())

// Debug level logging
slog.Debug("processing request", "request_id", "abc123")
```

### Context Helpers

The logging package provides helper functions for common contexts:

```go
import "github.com/Skpow1234/Peervault/internal/logging"

// Component-specific logger
logger := logging.Logger("fileserver")

// Error context
errorLogger := logging.WithError(err)

// Peer context
peerLogger := logging.WithPeer("127.0.0.1:3000")

// File key context
keyLogger := logging.WithKey("example.txt")

// Byte count context
bytesLogger := logging.WithBytes(1024)
```

## Log Entry Structure

### JSON Output Format

```json
{
  "time": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "msg": "file stored",
  "key": "example.txt",
  "bytes": 1024,
  "peer": "127.0.0.1:3000",
  "component": "fileserver",
  "source": {
    "function": "Store",
    "file": "server.go",
    "line": 145
  }
}
```

### Common Fields

- **time**: Timestamp of the log entry
- **level**: Log level (DEBUG, INFO, WARN, ERROR)
- **msg**: Human-readable message
- **component**: Application component (fileserver, transport, etc.)
- **peer**: Network peer address
- **key**: File key being processed
- **bytes**: Number of bytes transferred
- **error**: Error message (for error logs)
- **source**: Source code location (when enabled)

## Log Categories

### File Operations

```go
// File storage
slog.Info("file stored", "key", key, "bytes", size, "addr", s.Transport.Addr())

// File retrieval
slog.Info("serving file", "key", key, "addr", s.Transport.Addr())

// File not found
slog.Info("dont have file", "key", key, "addr", s.Transport.Addr())
```

### Network Operations

```go
// Peer connection
slog.Info("connected", "peer", p.RemoteAddr())

// Network transmission
slog.Info("written", "bytes", n, "peer", from)

// Network reception
slog.Info("received", "bytes", n, "peer", peer.RemoteAddr())

// Connection errors
slog.Error("dial error", "error", err.Error())
```

### Replication Operations

```go
// Replication success
slog.Info("successfully streamed to peer", "peer", peerAddr)

// Replication failure
slog.Error("attempt failed to stream to peer", "attempt", attempt+1, "peer", peerAddr, "error", err)

// Replication summary
slog.Info("successfully streamed to peers", "success_count", successCount, "total_peers", len(peers))
```

### System Operations

```go
// Server startup
slog.Info("starting fileserver", "addr", s.Transport.Addr())

// Server shutdown
slog.Info("file server stopped")

// Transport listening
slog.Info("TCP transport listening on port", "port", t.ListenAddr)
```

## Error Context

### Error Logging Best Practices

```go
// Good: Include context with errors
slog.Error("failed to decrypt file", "key", key, "error", err.Error())

// Good: Use structured error logging
slog.Error("network operation failed",
    "operation", "file_transfer",
    "peer", peerAddr,
    "error", err.Error(),
    "retry_attempt", attempt,
)

// Avoid: Just logging the error
slog.Error("error occurred", "error", err.Error()) // Too generic
```

### Error Propagation

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to decrypt file %s: %w", key, err)
}
```

## Performance Considerations

### Log Level Filtering

- **Production**: Use INFO level or higher
- **Development**: Use DEBUG level for detailed information
- **Performance**: Higher log levels reduce I/O overhead

### Structured vs Unstructured

```go
// Good: Structured logging
slog.Info("file processed", "key", key, "size", size, "duration", duration)

// Avoid: String concatenation
slog.Info(fmt.Sprintf("file %s processed, size: %d, duration: %v", key, size, duration))
```

## Monitoring and Analysis

### Log Aggregation

The JSON format enables easy integration with:

- **ELK Stack**: Elasticsearch, Logstash, Kibana
- **Splunk**: Log analysis and monitoring
- **Grafana**: Metrics and visualization
- **Custom Tools**: JSON parsing for analysis

### Metrics Extraction

Common metrics that can be extracted:

- **File operations per second**
- **Network transfer rates**
- **Error rates by component**
- **Peer connection status**
- **Replication success rates**

### Alerting

Set up alerts for:

- **High error rates**
- **Failed peer connections**
- **Replication failures**
- **System startup/shutdown events**

## Testing

### Log Testing

```go
func TestLogging(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))
    
    logger.Info("test message", "key", "value")
    
    var logEntry map[string]interface{}
    json.Unmarshal(buf.Bytes(), &logEntry)
    
    if logEntry["msg"] != "test message" {
        t.Error("log message mismatch")
    }
}
```

### Log Level Testing

```go
func TestLogLevels(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
    logger := slog.New(handler)
    
    // Debug should be filtered
    logger.Debug("debug message")
    if buf.Len() > 0 {
        t.Error("debug message should be filtered")
    }
    
    // Info should be logged
    logger.Info("info message")
    if buf.Len() == 0 {
        t.Error("info message should be logged")
    }
}
```

## Configuration Examples

### Development Environment

```bash
export LOG_LEVEL=debug
./peervault
```

### Production Environment

```bash
export LOG_LEVEL=info
./peervault > app.log 2>&1
```

### Docker Environment

```dockerfile
ENV LOG_LEVEL=info
CMD ["./peervault"]
```

## Best Practices

### 1. Use Appropriate Log Levels

- **DEBUG**: Detailed debugging information
- **INFO**: General operational events
- **WARN**: Potential issues that don't stop operation
- **ERROR**: Errors that need immediate attention

### 2. Include Relevant Context

```go
// Good: Rich context
slog.Info("file operation completed",
    "operation", "store",
    "key", key,
    "size", size,
    "duration", duration,
    "peer_count", len(peers),
)

// Avoid: Minimal context
slog.Info("operation completed")
```

### 3. Consistent Field Names

Use consistent field names across the application:

- **peer**: Network peer address
- **key**: File key
- **bytes**: Byte count
- **error**: Error message
- **component**: Application component

### 4. Avoid Sensitive Information

Never log:

- **Encryption keys**
- **Authentication tokens**
- **User passwords**
- **Personal data**

### 5. Performance Considerations

- Use appropriate log levels in production
- Avoid expensive operations in log statements
- Consider log rotation and retention policies

## Conclusion

The structured logging system provides comprehensive observability for the PeerVault application, enabling effective monitoring, debugging, and operational insights while maintaining good performance characteristics.
