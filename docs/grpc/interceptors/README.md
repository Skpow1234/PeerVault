# gRPC Interceptors

This document describes the gRPC interceptors implemented in PeerVault, providing authentication, logging, monitoring, and other cross-cutting concerns.

## Overview

gRPC interceptors provide a way to implement cross-cutting concerns such as authentication, logging, monitoring, and validation without modifying the core business logic. PeerVault implements a comprehensive set of interceptors that can be configured and combined as needed.

## Architecture

```text
Client Request
    ↓
Circuit Breaker Interceptor
    ↓
Rate Limit Interceptor
    ↓
Authentication Interceptor
    ↓
Validation Interceptor
    ↓
Cache Interceptor
    ↓
Logging Interceptor
    ↓
Monitoring Interceptor
    ↓
Business Logic Handler
```

## Available Interceptors

### 1. Authentication Interceptor

Provides JWT-based authentication for gRPC services.

**Features:**

- JWT token validation
- User role extraction
- Configurable token expiry
- Skip methods for public endpoints
- Rate limiting integration

**Configuration:**

```go
config := &interceptors.AuthConfig{
    SecretKey:     "your-secret-key",
    TokenExpiry:   24 * time.Hour,
    Issuer:        "peervault",
    Audience:      "peervault-api",
    RequireAuth:   true,
    SkipMethods:   []string{"/peervault.PeerVaultService/HealthCheck"},
    RateLimitRPS:  100,
    RateLimitBurst: 200,
}
```

**Usage:**

```go
authInterceptor := interceptors.NewAuthInterceptor(config, logger)
server := grpc.NewServer(
    grpc.UnaryInterceptor(authInterceptor.UnaryAuthInterceptor()),
    grpc.StreamInterceptor(authInterceptor.StreamAuthInterceptor()),
)
```

### 2. Logging Interceptor

Provides comprehensive logging for gRPC requests and responses.

**Features:**

- Request/response logging
- Error logging
- Duration tracking
- User ID tracking
- Configurable log levels
- Payload truncation
- Stream logging

**Configuration:**

```go
config := &interceptors.LoggingConfig{
    LogRequests:     true,
    LogResponses:    true,
    LogErrors:       true,
    LogDuration:     true,
    LogUserID:       true,
    LogMetadata:     false,
    LogPayload:      false,
    MaxPayloadSize:  1024,
    SkipMethods:     []string{"/peervault.PeerVaultService/HealthCheck"},
    LogLevel:        slog.LevelInfo,
}
```

**Usage:**

```go
loggingInterceptor := interceptors.NewLoggingInterceptor(config, logger)
server := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor.UnaryLoggingInterceptor()),
    grpc.StreamInterceptor(loggingInterceptor.StreamLoggingInterceptor()),
)
```

### 3. Monitoring Interceptor

Provides metrics collection and tracing for gRPC services.

**Features:**

- Request count metrics
- Duration histograms
- Error rate tracking
- Active stream monitoring
- Distributed tracing
- Performance profiling
- Custom metrics

**Configuration:**

```go
config := &interceptors.MonitoringConfig{
    EnableMetrics:     true,
    EnableTracing:     true,
    EnableProfiling:   false,
    MetricsPrefix:     "grpc",
    HistogramBuckets:  []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    CounterLabels:     []string{"method", "status_code", "user_id"},
    SkipMethods:       []string{"/peervault.PeerVaultService/HealthCheck"},
    SampleRate:        1.0,
    MaxTraceDuration:  30 * time.Second,
}
```

**Usage:**

```go
monitoringInterceptor := interceptors.NewMonitoringInterceptor(config, logger)
server := grpc.NewServer(
    grpc.UnaryInterceptor(monitoringInterceptor.UnaryMonitoringInterceptor()),
    grpc.StreamInterceptor(monitoringInterceptor.StreamMonitoringInterceptor()),
)
```

### 4. Rate Limit Interceptor

Provides rate limiting functionality to prevent abuse.

**Features:**

- Per-client rate limiting
- Configurable requests per second
- Burst size control
- Client identification
- Automatic rate limit headers

**Configuration:**

```go
rateLimitInterceptor := interceptors.NewRateLimitInterceptor(
    100, // requests per second
    200, // burst size
    logger,
)
```

**Usage:**

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(rateLimitInterceptor.UnaryRateLimitInterceptor()),
    grpc.StreamInterceptor(rateLimitInterceptor.StreamRateLimitInterceptor()),
)
```

### 5. Validation Interceptor

Provides request validation functionality.

**Features:**

- Method-specific validators
- Custom validation functions
- Error handling
- Validation caching

**Configuration:**

```go
validationInterceptor := interceptors.NewValidationInterceptor(logger)

// Add validators for specific methods
validationInterceptor.AddValidator(
    "/peervault.PeerVaultService/UploadFile",
    func(req interface{}) error {
        // Custom validation logic
        return nil
    },
)
```

**Usage:**

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(validationInterceptor.UnaryValidationInterceptor()),
    grpc.StreamInterceptor(validationInterceptor.StreamValidationInterceptor()),
)
```

### 6. Cache Interceptor

Provides response caching functionality.

**Features:**

- Response caching
- Configurable TTL
- Cache key generation
- Automatic cleanup
- Memory management

**Configuration:**

```go
cacheInterceptor := interceptors.NewCacheInterceptor(
    5 * time.Minute, // cache TTL
    logger,
)
```

**Usage:**

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(cacheInterceptor.UnaryCacheInterceptor()),
)
```

### 7. Circuit Breaker Interceptor

Provides circuit breaker functionality for fault tolerance.

**Features:**

- Failure threshold detection
- Automatic circuit opening
- Timeout-based recovery
- Half-open state handling
- Failure count tracking

**Configuration:**

```go
circuitBreakerInterceptor := interceptors.NewCircuitBreakerInterceptor(
    5,                // failure threshold
    30 * time.Second, // timeout
    logger,
)
```

**Usage:**

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(circuitBreakerInterceptor.UnaryCircuitBreakerInterceptor()),
)
```

## Interceptor Manager

The `InterceptorManager` provides a centralized way to manage all interceptors.

### Creating an Interceptor Manager

```go
manager := interceptors.NewInterceptorManager(logger)

// Configure interceptors
manager.SetAuthInterceptor(authConfig)
manager.SetLoggingInterceptor(loggingConfig)
manager.SetMonitoringInterceptor(monitoringConfig)
manager.SetRateLimitInterceptor(100, 200)
manager.SetValidationInterceptor()
manager.SetCacheInterceptor(300) // 5 minutes
manager.SetCircuitBreakerInterceptor(5, 30) // 5 failures, 30 seconds
```

### Using Interceptors

```go
// Get all unary interceptors
unaryInterceptors := manager.GetUnaryInterceptors()

// Get all stream interceptors
streamInterceptors := manager.GetStreamInterceptors()

// Create server with interceptors
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(unaryInterceptors...),
    grpc.ChainStreamInterceptor(streamInterceptors...),
)
```

### Managing Interceptors

```go
// Enable/disable specific interceptors
manager.EnableInterceptor("auth")
manager.DisableInterceptor("cache")

// Get interceptor status
status := manager.GetInterceptorStatus()

// Reset metrics and traces
manager.ResetAllMetrics()
manager.ResetAllTraces()

// Get metrics and traces
metrics := manager.GetMetrics()
traces := manager.GetTraces()
```

## Configuration Examples

### Basic Configuration

```go
// Create logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Create interceptor manager
manager := interceptors.NewInterceptorManager(logger)

// Configure authentication
authConfig := interceptors.DefaultAuthConfig()
authConfig.SecretKey = "your-secret-key"
authConfig.RequireAuth = true
manager.SetAuthInterceptor(authConfig)

// Configure logging
loggingConfig := interceptors.DefaultLoggingConfig()
loggingConfig.LogRequests = true
loggingConfig.LogResponses = true
loggingConfig.LogErrors = true
manager.SetLoggingInterceptor(loggingConfig)

// Configure monitoring
monitoringConfig := interceptors.DefaultMonitoringConfig()
monitoringConfig.EnableMetrics = true
monitoringConfig.EnableTracing = true
manager.SetMonitoringInterceptor(monitoringConfig)

// Configure rate limiting
manager.SetRateLimitInterceptor(100, 200)

// Configure validation
manager.SetValidationInterceptor()

// Configure caching
manager.SetCacheInterceptor(300) // 5 minutes

// Configure circuit breaker
manager.SetCircuitBreakerInterceptor(5, 30) // 5 failures, 30 seconds
```

### Production Configuration

```go
// Production-ready configuration
manager := interceptors.NewInterceptorManager(logger)

// Authentication with strict settings
authConfig := &interceptors.AuthConfig{
    SecretKey:     os.Getenv("JWT_SECRET"),
    TokenExpiry:   1 * time.Hour,
    Issuer:        "peervault-prod",
    Audience:      "peervault-api",
    RequireAuth:   true,
    SkipMethods:   []string{"/peervault.PeerVaultService/HealthCheck"},
    RateLimitRPS:  1000,
    RateLimitBurst: 2000,
}
manager.SetAuthInterceptor(authConfig)

// Logging with error focus
loggingConfig := &interceptors.LoggingConfig{
    LogRequests:     false, // Reduce log volume
    LogResponses:    false,
    LogErrors:       true,  // Focus on errors
    LogDuration:     true,
    LogUserID:       true,
    LogMetadata:     false,
    LogPayload:      false,
    MaxPayloadSize:  512,
    SkipMethods:     []string{"/peervault.PeerVaultService/HealthCheck"},
    LogLevel:        slog.LevelWarn,
}
manager.SetLoggingInterceptor(loggingConfig)

// Monitoring with full metrics
monitoringConfig := &interceptors.MonitoringConfig{
    EnableMetrics:     true,
    EnableTracing:     true,
    EnableProfiling:   false,
    MetricsPrefix:     "peervault_grpc",
    HistogramBuckets:  []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    CounterLabels:     []string{"method", "status_code", "user_id", "service"},
    SkipMethods:       []string{"/peervault.PeerVaultService/HealthCheck"},
    SampleRate:        0.1, // 10% sampling
    MaxTraceDuration:  60 * time.Second,
}
manager.SetMonitoringInterceptor(monitoringConfig)

// Aggressive rate limiting
manager.SetRateLimitInterceptor(500, 1000)

// Strict validation
manager.SetValidationInterceptor()

// Long-term caching
manager.SetCacheInterceptor(1800) // 30 minutes

// Conservative circuit breaker
manager.SetCircuitBreakerInterceptor(3, 60) // 3 failures, 60 seconds
```

## Best Practices

### 1. Interceptor Order

Interceptors are executed in the order they are added. Consider the following order:

1. **Circuit Breaker**: First to prevent cascading failures
2. **Rate Limiting**: Early rejection of excessive requests
3. **Authentication**: Verify identity before processing
4. **Validation**: Validate requests before business logic
5. **Caching**: Check cache before expensive operations
6. **Logging**: Log requests and responses
7. **Monitoring**: Collect metrics and traces

### 2. Error Handling

- Always return appropriate gRPC status codes
- Log errors with sufficient context
- Don't expose sensitive information in error messages
- Use structured logging for better observability

### 3. Performance Considerations

- Use sampling for high-volume operations
- Implement proper caching strategies
- Monitor interceptor performance
- Use appropriate buffer sizes

### 4. Security

- Validate all inputs
- Use secure token storage
- Implement proper rate limiting
- Monitor for suspicious activity

### 5. Monitoring

- Collect comprehensive metrics
- Set up alerts for critical failures
- Monitor interceptor performance
- Track error rates and patterns

## Troubleshooting

### Common Issues

1. **Authentication Failures**: Check token format and expiry
2. **Rate Limit Exceeded**: Adjust rate limits or implement backoff
3. **Validation Errors**: Verify request format and required fields
4. **Cache Misses**: Check cache key generation and TTL
5. **Circuit Breaker Open**: Investigate underlying service issues

### Debug Mode

Enable debug logging for detailed interceptor information:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Metrics and Traces

Monitor interceptor performance:

```go
// Get metrics
metrics := manager.GetMetrics()

// Get traces
traces := manager.GetTraces()

// Get interceptor status
status := manager.GetInterceptorStatus()
```

## Future Enhancements

- [ ] Distributed tracing integration
- [ ] Advanced caching strategies
- [ ] Machine learning-based rate limiting
- [ ] Automatic circuit breaker tuning
- [ ] Interceptor performance optimization
- [ ] Custom interceptor plugins
- [ ] A/B testing support
- [ ] Real-time configuration updates
