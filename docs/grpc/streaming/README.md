# Advanced Streaming Patterns

This document describes the advanced streaming patterns implemented in PeerVault's gRPC services, including error recovery, backpressure, and flow control.

## Overview

The advanced streaming implementation provides robust, production-ready streaming capabilities with:

- **Error Recovery**: Automatic retry mechanisms with exponential backoff
- **Backpressure**: Flow control to prevent memory exhaustion
- **Bidirectional Streaming**: Full-duplex communication with flow control
- **Stream Multiplexing**: Multiple streams over a single connection
- **Health Monitoring**: Real-time stream health and statistics

## Architecture

```text
Client
    ↓ gRPC Stream
StreamManager
    ↓ Stream Management
Stream (with Error Recovery & Backpressure)
    ↓ Event Broadcasting
EventBroadcaster
    ↓ Multiplexed Events
Multiple Subscribers
```

## Core Components

### StreamManager

The `StreamManager` is responsible for creating, managing, and monitoring streams.

```go
type StreamManager struct {
    config     *StreamConfig
    logger     *slog.Logger
    streams    map[string]*Stream
    streamMux  sync.RWMutex
    ctx        context.Context
    cancel     context.CancelFunc
}
```

**Features:**

- Stream lifecycle management
- Configuration management
- Health monitoring
- Graceful shutdown

### Stream

The `Stream` represents a managed stream with error recovery and backpressure.

```go
type Stream struct {
    ID           string
    Config       *StreamConfig
    Logger       *slog.Logger
    ctx          context.Context
    cancel       context.CancelFunc
    messageChan  chan interface{}
    errorChan    chan error
    retryCount   int
    lastActivity time.Time
    backpressure bool
    mutex        sync.RWMutex
}
```

**Features:**

- Message buffering with configurable size
- Automatic error recovery with retry logic
- Backpressure detection and handling
- Heartbeat monitoring
- Activity tracking

### BidirectionalStream

The `BidirectionalStream` provides full-duplex communication with flow control.

```go
type BidirectionalStream struct {
    *Stream
    SendChan    chan interface{}
    ReceiveChan chan interface{}
    FlowControl *FlowController
}
```

**Features:**

- Flow control with sliding window
- Bidirectional message passing
- Automatic acknowledgment handling
- Window size management

### StreamMultiplexer

The `StreamMultiplexer` allows multiple streams to share a single connection.

```go
type StreamMultiplexer struct {
    Streams map[string]*Stream
    Mutex   sync.RWMutex
    Logger  *slog.Logger
}
```

**Features:**

- Multiple stream management
- Broadcast messaging
- Stream statistics
- Resource sharing

## Configuration

### StreamConfig

```go
type StreamConfig struct {
    BufferSize        int           // Buffer size for the stream
    MaxRetries        int           // Maximum number of retry attempts
    RetryDelay        time.Duration // Delay between retries
    BackpressureLimit int           // Maximum pending messages before backpressure
    HeartbeatInterval time.Duration // Interval for heartbeat messages
    Timeout           time.Duration // Stream timeout
}
```

### Default Configuration

```go
func DefaultStreamConfig() *StreamConfig {
    return &StreamConfig{
        BufferSize:        1000,
        MaxRetries:        3,
        RetryDelay:        time.Second,
        BackpressureLimit: 100,
        HeartbeatInterval: 30 * time.Second,
        Timeout:           5 * time.Minute,
    }
}
```

## Error Recovery

### Retry Logic

The streaming implementation includes sophisticated retry logic:

1. **Exponential Backoff**: Retry delays increase exponentially
2. **Maximum Retries**: Configurable maximum retry attempts
3. **Error Classification**: Different retry strategies for different error types
4. **Circuit Breaker**: Prevents cascading failures

```go
// Retry logic example
if retryCount < maxRetries {
    retryCount++
    time.Sleep(stream.Config.RetryDelay * time.Duration(retryCount))
    continue
} else {
    return status.Error(codes.Internal, "max retries exceeded")
}
```

### Error Types

- **Transient Errors**: Automatically retried
- **Permanent Errors**: Not retried, logged and reported
- **Timeout Errors**: Retried with backoff
- **Resource Exhaustion**: Triggers backpressure

## Backpressure

### Detection

Backpressure is detected when:

- Message buffer is full
- Flow control window is exhausted
- Memory usage exceeds limits
- Processing rate is slower than message rate

### Handling

When backpressure is detected:

1. **Flow Control**: Reduce message sending rate
2. **Buffer Management**: Clear old messages if necessary
3. **Client Notification**: Inform clients to slow down
4. **Resource Monitoring**: Monitor system resources

```go
// Backpressure detection
if len(s.messageChan) >= s.Config.BackpressureLimit {
    s.backpressure = true
    return status.Error(codes.ResourceExhausted, "backpressure detected")
}
```

## Flow Control

### Sliding Window

The flow control implementation uses a sliding window approach:

```go
type FlowController struct {
    WindowSize    int
    CurrentWindow int
    Mutex         sync.RWMutex
}
```

### Window Management

- **Initial Window**: Set to buffer size
- **Window Reduction**: Decremented on each send
- **Window Increase**: Incremented on acknowledgment
- **Window Reset**: Reset on connection recovery

## Usage Examples

### Creating a Stream

```go
// Create stream manager
config := streaming.DefaultStreamConfig()
streamManager := streaming.NewStreamManager(config, logger)

// Create a stream
stream := streamManager.CreateStream("my-stream-id")

// Send messages
err := stream.SendMessage("Hello, World!")
if err != nil {
    log.Printf("Failed to send message: %v", err)
}
```

### Bidirectional Streaming

```go
// Create bidirectional stream
stream := streamManager.NewBidirectionalStream("bidirectional-stream")

// Send message with flow control
err := stream.Send("Hello, Server!")
if err != nil {
    log.Printf("Send failed: %v", err)
}

// Receive message
message, err := stream.Receive()
if err != nil {
    log.Printf("Receive failed: %v", err)
}

// Acknowledge receipt
stream.Acknowledge()
```

### Stream Multiplexing

```go
// Create multiplexer
multiplexer := streaming.NewStreamMultiplexer(logger)

// Add streams
multiplexer.AddStream("stream1", stream1)
multiplexer.AddStream("stream2", stream2)

// Broadcast message to all streams
multiplexer.BroadcastMessage("Broadcast message")

// Get statistics
stats := multiplexer.GetStats()
```

### Event Broadcasting

```go
// Create streaming service
service := streaming.NewStreamingService(logger)

// Subscribe to events
eventChan, err := service.SubscribeToFileOperationEvents("subscriber-1")
if err != nil {
    log.Printf("Subscription failed: %v", err)
}

// Listen for events
for event := range eventChan {
    log.Printf("Received event: %v", event)
}

// Unsubscribe
err = service.UnsubscribeFromFileOperationEvents("subscriber-1")
```

## Monitoring and Statistics

### Stream Statistics

```go
stats := stream.GetStats()
// Returns:
// {
//     "id": "stream-1",
//     "buffer_size": 50,
//     "max_buffer": 1000,
//     "retry_count": 2,
//     "last_activity": "2023-01-01T12:00:00Z",
//     "backpressure": false,
//     "healthy": true,
//     "uptime": "5m30s"
// }
```

### Service Health Check

```go
health := service.HealthCheck()
// Returns:
// {
//     "status": "healthy",
//     "total_streams": 10,
//     "healthy_streams": 9,
//     "subscribers": 5,
//     "timestamp": "2023-01-01T12:00:00Z"
// }
```

## Best Practices

### Configuration BP

1. **Buffer Sizes**: Set appropriate buffer sizes based on expected load
2. **Retry Limits**: Balance between reliability and performance
3. **Timeouts**: Set reasonable timeouts for different operations
4. **Heartbeat Intervals**: Monitor stream health without excessive overhead

### Error Handling

1. **Graceful Degradation**: Handle errors without crashing
2. **Logging**: Log all errors with appropriate context
3. **Monitoring**: Monitor error rates and patterns
4. **Alerting**: Set up alerts for critical errors

### Performance

1. **Resource Management**: Monitor memory and CPU usage
2. **Connection Pooling**: Reuse connections when possible
3. **Batch Processing**: Process messages in batches when appropriate
4. **Load Balancing**: Distribute load across multiple instances

## Troubleshooting

### Common Issues

1. **High Memory Usage**: Check buffer sizes and backpressure settings
2. **Slow Performance**: Monitor flow control and retry settings
3. **Connection Drops**: Check timeout and heartbeat configurations
4. **Error Loops**: Verify retry logic and error classification

### Debug Mode

Enable debug logging for detailed stream information:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Metrics Collection

Collect metrics for monitoring:

```go
// Stream metrics
streamStats := stream.GetStats()

// Service metrics
serviceStats := service.GetStreamStats()

// Health metrics
healthStats := service.HealthCheck()
```

## Future Enhancements

- [ ] Adaptive flow control based on network conditions
- [ ] Compression for large messages
- [ ] Encryption for sensitive streams
- [ ] Stream persistence for reliability
- [ ] Advanced load balancing strategies
- [ ] Machine learning-based error prediction
