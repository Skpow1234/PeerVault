# Client-Side Load Balancing

This document describes the client-side load balancing implementation for PeerVault's gRPC services, including health checking, failover, and service discovery.

## Overview

The client-side load balancing implementation provides:

- **Multiple Load Balancing Strategies**: Round-robin, random, weighted, and least connections
- **Health Checking**: Automatic health monitoring of server instances
- **Failover**: Automatic failover to healthy servers
- **Service Discovery**: Integration with various service discovery providers
- **Sticky Sessions**: Session affinity for stateful applications
- **Connection Management**: Efficient connection pooling and management

## Architecture

```text
Client Request
    ↓
LoadBalancedClient
    ↓
LoadBalancer
    ↓
Server Selection (Strategy)
    ↓
Health Check
    ↓
gRPC Server Instance
```

## Core Components

### 1. LoadBalancer

The `LoadBalancer` is the main component that manages server instances and implements load balancing strategies.

**Features:**

- Multiple load balancing strategies
- Health status tracking
- Connection counting
- Sticky session support
- Server management

**Configuration:**

```go
config := &balancing.LoadBalancerConfig{
    Strategy:           "round_robin", // "round_robin", "random", "weighted", "least_connections"
    HealthCheckInterval: 30 * time.Second,
    HealthCheckTimeout:  5 * time.Second,
    MaxRetries:         3,
    RetryDelay:         time.Second,
    FailoverTimeout:    10 * time.Second,
    StickySession:      false,
    SessionTimeout:     5 * time.Minute,
}
```

### 2. HealthChecker

The `HealthChecker` monitors the health of server instances.

**Features:**

- Periodic health checks
- Configurable intervals and timeouts
- Health status tracking
- Automatic server health updates

**Configuration:**

```go
healthChecker := balancing.NewHealthChecker(
    30 * time.Second, // check interval
    5 * time.Second,  // check timeout
    logger,
)
```

### 3. ServiceDiscovery

The `ServiceDiscovery` provides service discovery functionality.

**Features:**

- Multiple provider support (static, Consul, etcd, Kubernetes)
- Automatic service refresh
- Health status integration
- Service instance management

**Configuration:**

```go
config := &balancing.ServiceDiscoveryConfig{
    Provider:        "static", // "static", "consul", "etcd", "kubernetes"
    RefreshInterval: 30 * time.Second,
    Timeout:         10 * time.Second,
    MaxRetries:      3,
    RetryDelay:      time.Second,
}
```

### 4. LoadBalancedClient

The `LoadBalancedClient` provides a gRPC client with load balancing capabilities.

**Features:**

- Automatic failover
- Retry logic
- Connection management
- Stream support
- Session management

## Load Balancing Strategies

### 1. Round Robin

Distributes requests evenly across all healthy servers in a circular fashion.

```go
config := &balancing.LoadBalancerConfig{
    Strategy: "round_robin",
}
```

**Use Cases:**

- Even distribution of load
- Simple and predictable behavior
- Stateless applications

### 2. Random

Randomly selects a server from the pool of healthy servers.

```go
config := &balancing.LoadBalancerConfig{
    Strategy: "random",
}
```

**Use Cases:**

- Avoiding thundering herd problems
- Simple load distribution
- When server capabilities are similar

### 3. Weighted

Selects servers based on their assigned weights.

```go
// Add servers with different weights
loadBalancer.AddServer("server1", "localhost", 50051, 3) // weight 3
loadBalancer.AddServer("server2", "localhost", 50052, 1) // weight 1
```

**Use Cases:**

- Servers with different capacities
- Gradual traffic migration
- A/B testing scenarios

### 4. Least Connections

Selects the server with the fewest active connections.

```go
config := &balancing.LoadBalancerConfig{
    Strategy: "least_connections",
}
```

**Use Cases:**

- Long-running connections
- Stateful applications
- When connection count matters

## Health Checking

### Health Check Configuration

```go
healthChecker := balancing.NewHealthChecker(
    30 * time.Second, // check interval
    5 * time.Second,  // check timeout
    logger,
)
```

### Health Status Types

- **Healthy**: Server is responding normally
- **Unhealthy**: Server is not responding or returning errors
- **Unknown**: Health status has not been determined yet

### Health Check Implementation

The health checker performs the following checks:

1. **Connection State**: Verifies the gRPC connection is ready
2. **Ping Test**: Performs a simple connectivity test
3. **Timeout Handling**: Respects configured timeouts
4. **Error Handling**: Properly handles connection errors

## Service Discovery

### Static Service Discovery

For simple deployments, use static service discovery:

```go
// Create static service provider
staticProvider := balancing.NewStaticServiceProvider(logger)

// Add services
instances := []*balancing.ServiceInstance{
    {
        ID:      "server1",
        Address: "localhost",
        Port:    50051,
        Weight:  1,
        Health:  balancing.HealthStatusHealthy,
    },
    {
        ID:      "server2",
        Address: "localhost",
        Port:    50052,
        Weight:  1,
        Health:  balancing.HealthStatusHealthy,
    },
}

staticProvider.AddStaticService("peervault", instances)
```

### Consul Service Discovery

For dynamic service discovery, integrate with Consul:

```go
// Register Consul provider
consulProvider := balancing.NewConsulServiceProvider(consulConfig, logger)
serviceDiscovery.RegisterProvider("consul", consulProvider)

// Use Consul for service discovery
config := &balancing.ServiceDiscoveryConfig{
    Provider: "consul",
}
```

## Usage Examples

### Basic Setup

```go
// Create logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Create load balancer
config := balancing.DefaultLoadBalancerConfig()
loadBalancer := balancing.NewLoadBalancer(config, logger)

// Add servers
loadBalancer.AddServer("server1", "localhost", 50051, 1)
loadBalancer.AddServer("server2", "localhost", 50052, 1)
loadBalancer.AddServer("server3", "localhost", 50053, 1)

// Create load-balanced client
clientConfig := balancing.DefaultLoadBalancedClientConfig()
client := balancing.NewLoadBalancedClient(loadBalancer, clientConfig, logger)

// Use client for gRPC calls
err := client.Invoke(ctx, "/peervault.PeerVaultService/HealthCheck", req, resp)
```

### Advanced Configuration

```go
// Advanced load balancer configuration
config := &balancing.LoadBalancerConfig{
    Strategy:           "weighted",
    HealthCheckInterval: 15 * time.Second,
    HealthCheckTimeout:  3 * time.Second,
    MaxRetries:         5,
    RetryDelay:         2 * time.Second,
    FailoverTimeout:    15 * time.Second,
    StickySession:      true,
    SessionTimeout:     10 * time.Minute,
}

loadBalancer := balancing.NewLoadBalancer(config, logger)

// Add servers with different weights
loadBalancer.AddServer("primary", "primary.example.com", 50051, 5)
loadBalancer.AddServer("secondary", "secondary.example.com", 50051, 2)
loadBalancer.AddServer("backup", "backup.example.com", 50051, 1)

// Advanced client configuration
clientConfig := &balancing.LoadBalancedClientConfig{
    MaxRetries:         5,
    RetryDelay:         2 * time.Second,
    FailoverTimeout:    15 * time.Second,
    ConnectionTimeout:  30 * time.Second,
    KeepAliveInterval:  30 * time.Second,
    KeepAliveTimeout:   5 * time.Second,
    MaxReceiveSize:     8 * 1024 * 1024, // 8MB
    MaxSendSize:        8 * 1024 * 1024, // 8MB
    UserAgent:          "peervault-grpc-client/1.0",
}

client := balancing.NewLoadBalancedClient(loadBalancer, clientConfig, logger)
```

### Service Discovery Integration

```go
// Create service discovery
serviceDiscovery := balancing.NewServiceDiscovery(logger)

// Configure service discovery
config := &balancing.ServiceDiscoveryConfig{
    Provider:        "consul",
    RefreshInterval: 30 * time.Second,
    Timeout:         10 * time.Second,
    MaxRetries:      3,
    RetryDelay:      time.Second,
}

serviceDiscovery.SetConfig(config)

// Start service discovery
serviceDiscovery.Start()

// Get service instances
instances, err := serviceDiscovery.GetHealthyServiceInstances("peervault")
if err != nil {
    log.Fatal(err)
}

// Add discovered instances to load balancer
for _, instance := range instances {
    loadBalancer.AddServer(instance.ID, instance.Address, instance.Port, instance.Weight)
}
```

### Sticky Sessions

```go
// Enable sticky sessions
config := &balancing.LoadBalancerConfig{
    Strategy:      "round_robin",
    StickySession: true,
    SessionTimeout: 5 * time.Minute,
}

loadBalancer := balancing.NewLoadBalancer(config, logger)

// Use session ID in context
ctx := context.WithValue(context.Background(), "session_id", "user123")
err := client.Invoke(ctx, "/peervault.PeerVaultService/GetUserData", req, resp)
```

## Monitoring and Statistics

### Load Balancer Statistics

```go
stats := loadBalancer.GetServerStats()
// Returns:
// {
//     "total_servers": 3,
//     "healthy_servers": 2,
//     "unhealthy_servers": 1,
//     "unknown_servers": 0,
//     "total_connections": 15,
//     "strategy": "round_robin",
//     "sticky_session": false,
//     "active_sessions": 0,
//     "servers": [...]
// }
```

### Health Check Statistics

```go
healthStats := healthChecker.GetHealthStats()
// Returns:
// {
//     "total_servers": 3,
//     "healthy_servers": 2,
//     "unhealthy_servers": 1,
//     "unknown_servers": 0,
//     "check_interval": "30s",
//     "check_timeout": "5s",
//     "servers": [...]
// }
```

### Client Statistics

```go
clientStats := client.GetStats()
// Returns:
// {
//     "total_connections": 3,
//     "max_retries": 3,
//     "retry_delay": "1s",
//     "failover_timeout": "10s",
//     "load_balancer": {...}
// }
```

## Best Practices

### 1. Server Configuration

- **Health Check Intervals**: Set appropriate intervals based on your requirements
- **Timeouts**: Configure timeouts to match your service characteristics
- **Weights**: Use weights to reflect server capacities
- **Connection Limits**: Set appropriate connection limits

### 2. Client Configuration

- **Retry Logic**: Configure retry attempts and delays appropriately
- **Connection Management**: Use connection pooling for efficiency
- **Timeout Settings**: Set timeouts based on your service requirements
- **Buffer Sizes**: Configure appropriate buffer sizes for your use case

### 3. Health Checking

- **Check Frequency**: Balance between responsiveness and overhead
- **Timeout Values**: Set timeouts to detect failures quickly
- **Error Handling**: Implement proper error handling and logging
- **Monitoring**: Monitor health check results and patterns

### 4. Service Discovery

- **Provider Selection**: Choose the appropriate service discovery provider
- **Refresh Intervals**: Set refresh intervals based on your deployment model
- **Error Handling**: Handle service discovery failures gracefully
- **Fallback**: Implement fallback mechanisms for service discovery failures

### 5. Monitoring

- **Metrics Collection**: Collect comprehensive metrics
- **Alerting**: Set up alerts for critical failures
- **Logging**: Implement structured logging for debugging
- **Dashboards**: Create dashboards for monitoring

## Troubleshooting

### Common Issues

1. **No Healthy Servers**: Check server health and connectivity
2. **Connection Failures**: Verify network connectivity and firewall rules
3. **Health Check Failures**: Check server health check implementation
4. **Service Discovery Issues**: Verify service discovery configuration
5. **Sticky Session Problems**: Check session timeout and server health

### Debug Mode

Enable debug logging for detailed information:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

### Health Check Debugging

Force health checks for debugging:

```go
// Force health check for specific server
err := healthChecker.ForceHealthCheck("server1")

// Force health check for all servers
healthChecker.ForceHealthCheckAll()
```

## Future Enhancements

- [ ] Advanced load balancing algorithms (consistent hashing, etc.)
- [ ] Circuit breaker integration
- [ ] Rate limiting per server
- [ ] Automatic scaling based on load
- [ ] Geographic load balancing
- [ ] Machine learning-based server selection
- [ ] Advanced health check strategies
- [ ] Integration with more service discovery providers
- [ ] Real-time configuration updates
- [ ] Performance optimization
