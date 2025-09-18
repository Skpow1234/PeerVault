# Advanced Health Checking

This document describes the comprehensive health checking implementation for PeerVault's gRPC services, including detailed status reporting, dependency management, and monitoring.

## Overview

The advanced health checking system provides:

- **Comprehensive Health Monitoring**: Monitor all system components
- **Dependency Management**: Track component dependencies and health
- **Detailed Status Reporting**: Rich health information with metrics and traces
- **Health Aggregation**: Aggregate health status from multiple components
- **Performance Profiling**: Track health check performance over time
- **Event Streaming**: Real-time health event streaming
- **Custom Health Checks**: Support for custom health check implementations

## Architecture

```text
Health Request
    ↓
AdvancedHealthChecker
    ↓
ComponentHealthChecker (per component)
    ↓
Health Check Function
    ↓
HealthResult
    ↓
HealthAggregator
    ↓
HealthResponse
```

## Core Components

### 1. AdvancedHealthChecker

The main health checker that orchestrates all health checks.

**Features:**

- Component registration and management
- Dependency tracking
- Health check scheduling
- Result aggregation
- Metrics collection
- Tracing and profiling

**Configuration:**

```go
config := &health.HealthConfig{
    CheckInterval:     30 * time.Second,
    CheckTimeout:      5 * time.Second,
    MaxRetries:        3,
    RetryDelay:        time.Second,
    EnableMetrics:     true,
    EnableTracing:     true,
    EnableProfiling:   false,
    MetricsInterval:   60 * time.Second,
    TraceInterval:     300 * time.Second,
    ProfileInterval:   600 * time.Second,
}
```

### 2. ComponentHealthChecker

Manages health checks for individual components.

**Features:**

- Component-specific health checks
- Dependency validation
- Result caching
- Performance tracking
- Error handling

### 3. HealthAggregator

Aggregates health results from multiple components.

**Features:**

- Overall health status calculation
- Dependency-aware aggregation
- Health percentage calculation
- Metrics aggregation

### 4. HealthMetrics

Collects and manages health-related metrics.

**Features:**

- Component performance metrics
- Success/failure rates
- Duration tracking
- Historical data

### 5. HealthTracer

Traces health check operations for debugging.

**Features:**

- Operation tracing
- Performance analysis
- Error tracking
- Historical traces

### 6. HealthProfiler

Profiles health check performance over time.

**Features:**

- Performance profiling
- Trend analysis
- Bottleneck identification
- Optimization insights

## Built-in Health Checks

### Database Health Check

```go
// Register database health check
healthChecker.RegisterComponent("database", health.DatabaseHealthCheck, []string{})
```

**Checks:**

- Connection status
- Query performance
- Connection pool health
- Database version

### Cache Health Check

```go
// Register cache health check
healthChecker.RegisterComponent("cache", health.CacheHealthCheck, []string{})
```

**Checks:**

- Cache connectivity
- Hit/miss rates
- Memory usage
- Cache version

### Storage Health Check

```go
// Register storage health check
healthChecker.RegisterComponent("storage", health.StorageHealthCheck, []string{})
```

**Checks:**

- Disk space
- I/O performance
- File system health
- Storage type

### Network Health Check

```go
// Register network health check
healthChecker.RegisterComponent("network", health.NetworkHealthCheck, []string{})
```

**Checks:**

- Network connectivity
- Latency
- Packet loss
- Bandwidth

### Service Health Check

```go
// Register external service health check
serviceCheck := health.ServiceHealthCheck("external-api", "https://api.example.com")
healthChecker.RegisterComponent("external-api", serviceCheck, []string{"network"})
```

**Checks:**

- Service availability
- Response time
- Status codes
- Endpoint health

## Custom Health Checks

### Simple Custom Check

```go
// Create a custom health check
customCheck := health.CustomHealthCheck("my-component", func(ctx context.Context) error {
    // Perform health check logic
    if someCondition {
        return fmt.Errorf("component is unhealthy")
    }
    return nil
})

// Register the custom check
healthChecker.RegisterComponent("my-component", customCheck, []string{"database"})
```

### Advanced Custom Check

```go
// Create an advanced custom health check
func MyComponentHealthCheck(ctx context.Context) (*health.HealthResult, error) {
    start := time.Now()
    
    // Perform health check logic
    status := health.HealthStatusHealthy
    message := "Component healthy"
    metrics := make(map[string]interface{})
    details := make(map[string]interface{})
    
    // Check component status
    if err := checkComponentStatus(); err != nil {
        status = health.HealthStatusUnhealthy
        message = err.Error()
    }
    
    // Collect metrics
    metrics["active_connections"] = getActiveConnections()
    metrics["memory_usage"] = getMemoryUsage()
    
    // Collect details
    details["version"] = getComponentVersion()
    details["uptime"] = getUptime()
    
    return &health.HealthResult{
        Component: "my-component",
        Status:    status,
        Message:   message,
        Timestamp: time.Now(),
        Duration:  time.Since(start),
        Metrics:   metrics,
        Details:   details,
    }, nil
}

// Register the advanced check
healthChecker.RegisterComponent("my-component", MyComponentHealthCheck, []string{"database", "cache"})
```

## Usage Examples

### Basic Setup

```go
// Create logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Create health checker
config := health.DefaultHealthConfig()
healthChecker := health.NewAdvancedHealthChecker(config, logger)

// Register components
healthChecker.RegisterComponent("database", health.DatabaseHealthCheck, []string{})
healthChecker.RegisterComponent("cache", health.CacheHealthCheck, []string{})
healthChecker.RegisterComponent("storage", health.StorageHealthCheck, []string{})
healthChecker.RegisterComponent("network", health.NetworkHealthCheck, []string{})

// Start health checker
healthChecker.Start()

// Create health service
healthService := health.NewHealthService(healthChecker, logger)

// Register with gRPC server
grpcServer := grpc.NewServer()
peervault.RegisterPeerVaultServiceServer(grpcServer, healthService)
```

### Advanced Configuration

```go
// Advanced health checker configuration
config := &health.HealthConfig{
    CheckInterval:     15 * time.Second,  // More frequent checks
    CheckTimeout:      3 * time.Second,   // Shorter timeout
    MaxRetries:        5,                 // More retries
    RetryDelay:        2 * time.Second,   // Longer retry delay
    EnableMetrics:     true,              // Enable metrics
    EnableTracing:     true,              // Enable tracing
    EnableProfiling:   true,              // Enable profiling
    MetricsInterval:   30 * time.Second,  // More frequent metrics
    TraceInterval:     60 * time.Second,  // More frequent traces
    ProfileInterval:   300 * time.Second, // More frequent profiles
}

healthChecker := health.NewAdvancedHealthChecker(config, logger)

// Register components with dependencies
healthChecker.RegisterComponent("database", health.DatabaseHealthCheck, []string{})
healthChecker.RegisterComponent("cache", health.CacheHealthCheck, []string{"database"})
healthChecker.RegisterComponent("storage", health.StorageHealthCheck, []string{})
healthChecker.RegisterComponent("network", health.NetworkHealthCheck, []string{})
healthChecker.RegisterComponent("external-api", externalApiCheck, []string{"network"})
healthChecker.RegisterComponent("my-service", myServiceCheck, []string{"database", "cache", "external-api"})

// Start health checker
healthChecker.Start()
```

### Dependency Management

```go
// Register components with dependencies
healthChecker.RegisterComponent("database", health.DatabaseHealthCheck, []string{})
healthChecker.RegisterComponent("cache", health.CacheHealthCheck, []string{"database"})
healthChecker.RegisterComponent("storage", health.StorageHealthCheck, []string{})
healthChecker.RegisterComponent("network", health.NetworkHealthCheck, []string{})
healthChecker.RegisterComponent("external-api", externalApiCheck, []string{"network"})
healthChecker.RegisterComponent("my-service", myServiceCheck, []string{"database", "cache", "external-api"})

// Components will be checked in dependency order:
// 1. database, storage, network (no dependencies)
// 2. cache (depends on database)
// 3. external-api (depends on network)
// 4. my-service (depends on database, cache, external-api)
```

## Health Status Types

### HealthStatusHealthy

Component is functioning normally.

```go
result := &health.HealthResult{
    Component: "database",
    Status:    health.HealthStatusHealthy,
    Message:   "Database connection healthy",
    // ... other fields
}
```

### HealthStatusUnhealthy

Component is not functioning properly.

```go
result := &health.HealthResult{
    Component: "database",
    Status:    health.HealthStatusUnhealthy,
    Message:   "Database connection failed",
    Error:     fmt.Errorf("connection timeout"),
    // ... other fields
}
```

### HealthStatusDegraded

Component is functioning but with reduced performance.

```go
result := &health.HealthResult{
    Component: "cache",
    Status:    health.HealthStatusDegraded,
    Message:   "Cache hit rate below threshold",
    // ... other fields
}
```

### HealthStatusUnknown

Component health status has not been determined.

```go
result := &health.HealthResult{
    Component: "new-component",
    Status:    health.HealthStatusUnknown,
    Message:   "Health check not yet performed",
    // ... other fields
}
```

## gRPC Health Service

### Health Check Endpoint

```go
// Basic health check
response, err := client.HealthCheck(ctx, &emptypb.Empty{})
if err != nil {
    log.Printf("Health check failed: %v", err)
}

log.Printf("Health status: %s", response.Status)
log.Printf("Version: %s", response.Version)
```

### Detailed Health Check

```go
// Detailed health check
response, err := client.GetDetailedHealth(ctx, &emptypb.Empty{})
if err != nil {
    log.Printf("Detailed health check failed: %v", err)
}

log.Printf("Overall status: %s", response.Status)
for key, value := range response.Metadata {
    log.Printf("%s: %s", key, value)
}
```

### Component Health Check

```go
// Component-specific health check
req := &peervault.ComponentHealthRequest{
    Component: "database",
}

response, err := client.GetComponentHealth(ctx, req)
if err != nil {
    log.Printf("Component health check failed: %v", err)
}

log.Printf("Component: %s", response.Component)
log.Printf("Status: %s", response.Status)
log.Printf("Message: %s", response.Message)
log.Printf("Duration: %d ns", response.Duration)
```

### Force Health Check

```go
// Force health check for specific component
req := &peervault.ForceHealthCheckRequest{
    Component: "database",
}

response, err := client.ForceHealthCheck(ctx, req)
if err != nil {
    log.Printf("Force health check failed: %v", err)
}

log.Printf("Success: %t", response.Success)
log.Printf("Message: %s", response.Message)
```

### Health Metrics

```go
// Get health metrics
response, err := client.GetHealthMetrics(ctx, &emptypb.Empty{})
if err != nil {
    log.Printf("Health metrics failed: %v", err)
}

for key, value := range response.Metrics {
    log.Printf("%s: %s", key, value)
}
```

### Health Traces

```go
// Get health traces
response, err := client.GetHealthTraces(ctx, &emptypb.Empty{})
if err != nil {
    log.Printf("Health traces failed: %v", err)
}

for _, trace := range response.Traces {
    log.Printf("Trace ID: %s", trace.Id)
    log.Printf("Component: %s", trace.Component)
    log.Printf("Status: %s", trace.Status)
    log.Printf("Duration: %d ns", trace.Duration)
}
```

### Health Profiles

```go
// Get health profiles
response, err := client.GetHealthProfiles(ctx, &emptypb.Empty{})
if err != nil {
    log.Printf("Health profiles failed: %v", err)
}

for _, profile := range response.Profiles {
    log.Printf("Component: %s", profile.Component)
    log.Printf("Check Count: %d", profile.CheckCount)
    log.Printf("Avg Duration: %d ns", profile.AvgDuration)
    log.Printf("Min Duration: %d ns", profile.MinDuration)
    log.Printf("Max Duration: %d ns", profile.MaxDuration)
}
```

### Health Event Streaming

```go
// Stream health events
stream, err := client.StreamHealthEvents(ctx, &emptypb.Empty{})
if err != nil {
    log.Printf("Health event stream failed: %v", err)
}

for {
    event, err := stream.Recv()
    if err != nil {
        if err == io.EOF {
            break
        }
        log.Printf("Stream error: %v", err)
        break
    }

    log.Printf("Event Type: %s", event.EventType)
    log.Printf("Component: %s", event.Component)
    log.Printf("Status: %s", event.Status)
    log.Printf("Message: %s", event.Message)
}
```

## Monitoring and Alerting

### Health Metrics Monitoring

```go
// Get health metrics
metrics := healthChecker.GetHealthMetrics()

// Example metrics:
// {
//     "database": {
//         "check_count": 100,
//         "failure_count": 2,
//         "success_rate": 98.0,
//         "last_duration": "5ms",
//         "last_status": "healthy"
//     },
//     "overall": {
//         "total_checks": 500,
//         "total_failures": 5,
//         "overall_success_rate": 99.0,
//         "total_duration": "250ms"
//     }
// }
```

### Health Traces Monitoring

```go
// Get health traces
traces := healthChecker.GetHealthTraces()

// Example traces:
// {
//     "trace_1": {
//         "id": "trace_1",
//         "component": "database",
//         "start_time": "2023-01-01T12:00:00Z",
//         "end_time": "2023-01-01T12:00:05Z",
//         "duration": "5ms",
//         "status": "healthy",
//         "details": {
//             "message": "Database connection healthy",
//             "metrics": "{\"connections\": 10}"
//         }
//     }
// }
```

### Health Profiles Monitoring

```go
// Get health profiles
profiles := healthChecker.GetHealthProfiles()

// Example profiles:
// {
//     "database": {
//         "component": "database",
//         "check_count": 100,
//         "total_duration": "500ms",
//         "avg_duration": "5ms",
//         "min_duration": "2ms",
//         "max_duration": "15ms",
//         "last_updated": "2023-01-01T12:00:00Z"
//     }
// }
```

## Best Practices

### 1. Component Design

- **Single Responsibility**: Each component should have a single, well-defined purpose
- **Dependency Management**: Clearly define component dependencies
- **Error Handling**: Implement proper error handling and logging
- **Performance**: Keep health checks lightweight and fast

### 2. Health Check Implementation

- **Timeout Handling**: Always implement timeouts for health checks
- **Resource Management**: Properly manage resources in health checks
- **Error Classification**: Distinguish between different types of errors
- **Metrics Collection**: Collect relevant metrics for monitoring

### 3. Configuration

- **Check Intervals**: Set appropriate check intervals based on component criticality
- **Timeouts**: Configure timeouts to match component characteristics
- **Retry Logic**: Implement appropriate retry logic for transient failures
- **Dependencies**: Order components based on their dependencies

### 4. Monitoring

- **Metrics Collection**: Collect comprehensive metrics
- **Alerting**: Set up alerts for critical health failures
- **Logging**: Implement structured logging for debugging
- **Dashboards**: Create dashboards for health monitoring

### 5. Performance

- **Efficient Checks**: Implement efficient health check logic
- **Caching**: Cache health check results when appropriate
- **Parallel Execution**: Run independent health checks in parallel
- **Resource Limits**: Set appropriate resource limits

## Troubleshooting

### Common Issues

1. **Health Check Timeouts**: Check component performance and network connectivity
2. **Dependency Failures**: Verify component dependencies and order
3. **Resource Exhaustion**: Monitor resource usage and limits
4. **Configuration Errors**: Verify health check configuration
5. **Network Issues**: Check network connectivity and firewall rules

### Debug Mode

Enable debug logging for detailed health check information:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Force Health Checks

Force health checks for debugging:

```go
// Force health check for specific component
err := healthChecker.ForceHealthCheck("database")

// Force health check for all components
healthChecker.ForceHealthCheckAll()
```

## Future Enhancements

- [ ] Machine learning-based health prediction
- [ ] Advanced dependency resolution
- [ ] Health check scheduling optimization
- [ ] Integration with external monitoring systems
- [ ] Advanced alerting and notification
- [ ] Health check result caching
- [ ] Distributed health checking
- [ ] Health check result persistence
- [ ] Advanced analytics and reporting
- [ ] Health check result visualization
