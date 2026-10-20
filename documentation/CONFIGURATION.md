# Shugo Configuration Management

## Overview

Shugo provides a comprehensive configuration management system that supports hierarchical configuration loading, environment variable overrides, validation, and hot reloading. The system is designed to be flexible, secure, and easy to use across different deployment environments.

## Features

- **Hierarchical Configuration**: Load from YAML/JSON files with environment variable overrides
- **Environment Variable Support**: Override any configuration value using environment variables
- **Validation**: Built-in validation with custom validation rules
- **Hot Reloading**: Watch configuration files for changes and reload automatically
- **Type Safety**: Strongly typed configuration structures
- **Documentation**: Self-documenting configuration with examples

## Configuration Structure

The Shugo configuration is organized into logical sections:

### Server Configuration

```yaml
server:
  # Node ID (auto-generated if not provided)
  node_id: ""
  
  # Listen address for the server
  listen_addr: ":3000"
  
  # Enable debug mode
  debug: false
  
  # Graceful shutdown timeout
  shutdown_timeout: "30s"
```

### Storage Configuration

```yaml
storage:
  # Storage root directory
  root: "./storage"
  
  # Maximum file size in bytes (1GB)
  max_file_size: 1073741824
  
  # Enable compression
  compression: false
  
  # Compression level (1-9)
  compression_level: 6
  
  # Enable deduplication
  deduplication: false
  
  # Storage cleanup interval
  cleanup_interval: "1h"
  
  # Retention period for deleted files
  retention_period: "24h"
```

### Network Configuration

```yaml
network:
  # Bootstrap nodes (comma-separated)
  bootstrap_nodes:
    - "localhost:3000"
    - "localhost:7000"
  
  # Connection timeout
  connection_timeout: "30s"
  
  # Read timeout
  read_timeout: "60s"
  
  # Write timeout
  write_timeout: "60s"
  
  # Keep-alive interval
  keep_alive_interval: "30s"
  
  # Maximum message size (1MB)
  max_message_size: 1048576
```

### Security Configuration

```yaml
security:
  # Cluster key for encryption (set via environment variable SHUGO_CLUSTER_KEY)
  cluster_key: ""
  
  # Authentication token
  auth_token: "demo-token"
  
  # Enable TLS
  tls: false
  
  # TLS certificate file
  tls_cert_file: ""
  
  # TLS key file
  tls_key_file: ""
  
  # Key rotation interval
  key_rotation_interval: "24h"
  
  # Enable encryption at rest
  encryption_at_rest: true
  
  # Enable encryption in transit
  encryption_in_transit: true
```

### Logging Configuration

```yaml
logging:
  # Log level (debug, info, warn, error)
  level: "info"
  
  # Log format (json, text)
  format: "json"
  
  # Log file path (optional)
  file: ""
  
  # Enable structured logging
  structured: true
  
  # Include source location in logs
  include_source: false
  
  # Log rotation settings
  rotation:
    # Maximum log file size in MB
    max_size: 100
    
    # Maximum number of log files to keep
    max_files: 10
    
    # Maximum age of log files (7 days)
    max_age: "168h"
    
    # Compress rotated log files
    compress: true
```

### API Configuration

```yaml
api:
  # REST API Configuration
  rest:
    # Enable REST API
    enabled: true
    
    # REST API port
    port: 8080
    
    # Allowed origins for CORS
    allowed_origins:
      - "*"
    
    # Rate limit per minute
    rate_limit_per_min: 100
    
    # Authentication token
    auth_token: "demo-token"
  
  # GraphQL API Configuration
  graphql:
    # Enable GraphQL API
    enabled: true
    
    # GraphQL API port
    port: 8081
    
    # Enable GraphQL Playground
    enable_playground: true
    
    # GraphQL endpoint path
    graphql_path: "/graphql"
    
    # Playground path
    playground_path: "/playground"
    
    # Allowed origins for CORS
    allowed_origins:
      - "*"
  
  # gRPC API Configuration
  grpc:
    # Enable gRPC API
    enabled: true
    
    # gRPC API port
    port: 8082
    
    # Authentication token
    auth_token: "demo-token"
    
    # Enable reflection
    enable_reflection: true
    
    # Maximum concurrent streams
    max_concurrent_streams: 100
```

### Peer Configuration

```yaml
peer:
  # Maximum number of peers
  max_peers: 100
  
  # Heartbeat interval
  heartbeat_interval: "30s"
  
  # Health timeout
  health_timeout: "90s"
  
  # Reconnection backoff
  reconnect_backoff: "5s"
  
  # Maximum reconnection attempts
  max_reconnect_attempts: 10
```

### Performance Configuration

```yaml
performance:
  # Maximum concurrent streams per peer
  max_concurrent_streams_per_peer: 10
  
  # Buffer size for streaming (64KB)
  stream_buffer_size: 65536
  
  # Connection pool size
  connection_pool_size: 10
  
  # Enable connection multiplexing
  enable_multiplexing: true
  
  # Cache size in MB
  cache_size: 100
  
  # Cache TTL
  cache_ttl: "1h"
```

## Environment Variables

All configuration values can be overridden using environment variables. The environment variable names follow the pattern `SHUGO_<SECTION>_<FIELD>`.

### Server Environment Variables

- `SHUGO_NODE_ID` - Node ID
- `SHUGO_LISTEN_ADDR` - Listen address
- `SHUGO_DEBUG` - Enable debug mode
- `SHUGO_SHUTDOWN_TIMEOUT` - Shutdown timeout

### Storage Environment Variables

- `SHUGO_STORAGE_ROOT` - Storage root directory
- `SHUGO_MAX_FILE_SIZE` - Maximum file size
- `SHUGO_COMPRESSION` - Enable compression
- `SHUGO_COMPRESSION_LEVEL` - Compression level
- `SHUGO_DEDUPLICATION` - Enable deduplication
- `SHUGO_CLEANUP_INTERVAL` - Cleanup interval
- `SHUGO_RETENTION_PERIOD` - Retention period

### Network Environment Variables

- `SHUGO_BOOTSTRAP_NODES` - Bootstrap nodes (comma-separated)
- `SHUGO_CONNECTION_TIMEOUT` - Connection timeout
- `SHUGO_READ_TIMEOUT` - Read timeout
- `SHUGO_WRITE_TIMEOUT` - Write timeout
- `SHUGO_KEEP_ALIVE_INTERVAL` - Keep-alive interval
- `SHUGO_MAX_MESSAGE_SIZE` - Maximum message size

### Security Environment Variables

- `SHUGO_CLUSTER_KEY` - Cluster key for encryption
- `SHUGO_AUTH_TOKEN` - Authentication token
- `SHUGO_TLS` - Enable TLS
- `SHUGO_TLS_CERT_FILE` - TLS certificate file
- `SHUGO_TLS_KEY_FILE` - TLS key file
- `SHUGO_KEY_ROTATION_INTERVAL` - Key rotation interval
- `SHUGO_ENCRYPTION_AT_REST` - Enable encryption at rest
- `SHUGO_ENCRYPTION_IN_TRANSIT` - Enable encryption in transit

### Logging Environment Variables

- `SHUGO_LOG_LEVEL` - Log level (debug, info, warn, error)
- `SHUGO_LOG_FORMAT` - Log format (json, text)
- `SHUGO_LOG_FILE` - Log file path
- `SHUGO_LOG_STRUCTURED` - Enable structured logging
- `SHUGO_LOG_INCLUDE_SOURCE` - Include source location
- `SHUGO_LOG_MAX_SIZE` - Max log file size (MB)
- `SHUGO_LOG_MAX_FILES` - Max log files to keep
- `SHUGO_LOG_MAX_AGE` - Max log file age
- `SHUGO_LOG_COMPRESS` - Compress rotated logs

### API Environment Variables

#### REST API

- `SHUGO_REST_ENABLED` - Enable REST API
- `SHUGO_REST_PORT` - REST API port
- `SHUGO_REST_ALLOWED_ORIGINS` - Allowed origins (comma-separated)
- `SHUGO_REST_RATE_LIMIT` - Rate limit per minute
- `SHUGO_REST_AUTH_TOKEN` - REST auth token

#### GraphQL API

- `SHUGO_GRAPHQL_ENABLED` - Enable GraphQL API
- `SHUGO_GRAPHQL_PORT` - GraphQL API port
- `SHUGO_GRAPHQL_PLAYGROUND` - Enable GraphQL Playground
- `SHUGO_GRAPHQL_PATH` - GraphQL endpoint path
- `SHUGO_GRAPHQL_PLAYGROUND_PATH` - Playground path
- `SHUGO_GRAPHQL_ALLOWED_ORIGINS` - Allowed origins (comma-separated)

#### gRPC API

- `SHUGO_GRPC_ENABLED` - Enable gRPC API
- `SHUGO_GRPC_PORT` - gRPC API port
- `SHUGO_GRPC_AUTH_TOKEN` - gRPC auth token
- `SHUGO_GRPC_REFLECTION` - Enable reflection
- `SHUGO_GRPC_MAX_STREAMS` - Max concurrent streams

### Peer Environment Variables

- `SHUGO_MAX_PEERS` - Maximum number of peers
- `SHUGO_HEARTBEAT_INTERVAL` - Heartbeat interval
- `SHUGO_HEALTH_TIMEOUT` - Health timeout
- `SHUGO_RECONNECT_BACKOFF` - Reconnection backoff
- `SHUGO_MAX_RECONNECT_ATTEMPTS` - Max reconnection attempts

### Performance Environment Variables

- `SHUGO_MAX_STREAMS_PER_PEER` - Max concurrent streams per peer
- `SHUGO_STREAM_BUFFER_SIZE` - Stream buffer size
- `SHUGO_CONNECTION_POOL_SIZE` - Connection pool size
- `SHUGO_ENABLE_MULTIPLEXING` - Enable connection multiplexing
- `SHUGO_CACHE_SIZE` - Cache size (MB)
- `SHUGO_CACHE_TTL` - Cache TTL

## Usage

### Basic Configuration Loading

```go
package main

import (
    "log"
    "github.com/Skpow1234/Shugo/internal/config"
)

func main() {
    // Create configuration manager
    manager := config.NewManager("config/shugo.yaml")
    
    // Add validators
    manager.AddValidator(&config.DefaultValidator{})
    manager.AddValidator(&config.PortValidator{})
    manager.AddValidator(config.NewSecurityValidator(false)) // Will use config.Security.AllowDemoToken
    manager.AddValidator(&config.StorageValidator{})
    
    // Load configuration
    if err := manager.Load(); err != nil {
        log.Fatal("Failed to load configuration:", err)
    }
    
    // Get configuration
    cfg := manager.Get()
    
    // Use configuration
    fmt.Printf("Server listening on: %s\n", cfg.Server.ListenAddr)
    fmt.Printf("Storage root: %s\n", cfg.Storage.Root)
    fmt.Printf("Log level: %s\n", cfg.Logging.Level)
}
```

### Environment Variable Overrides

Set these environment variables via Docker compose.

### Hot Reloading

```go
// Start watching for configuration changes
err := manager.Watch(func(cfg *config.Config) {
    log.Printf("Configuration reloaded: %s", cfg.Server.ListenAddr)
    // Handle configuration changes
})
if err != nil {
    log.Fatal("Failed to start config watcher:", err)
}

// Stop watching when done
defer manager.Stop()
```

### Custom Validation

```go
// Create custom validator
type CustomValidator struct{}

func (v *CustomValidator) Validate(config *config.Config) error {
    // Add custom validation logic
    if config.Storage.MaxFileSize > 10*1024*1024*1024 { // 10GB
        return fmt.Errorf("file size too large for this deployment")
    }
    return nil
}

// Add custom validator
manager.AddValidator(&CustomValidator{})
```

## Configuration Management Tool

Shugo includes a command-line tool for configuration management:

### Generate Default Configuration

```bash
# Generate default YAML configuration
./shugo-config -generate -format yaml -output config/shugo.yaml

# Generate JSON configuration
./shugo-config -generate -format json -output config/shugo.json
```

### Validate Configuration

```bash
# Validate configuration file
./shugo-config -validate -config config/shugo.yaml
```

### Show Configuration

```bash
# Show current configuration in YAML format
./shugo-config -show -format yaml

# Show current configuration in JSON format
./shugo-config -show -format json
```

### Show Environment Mappings

```bash
# Show all environment variable mappings
./shugo-config -env
```

### Watch Configuration

```bash
# Watch configuration file for changes
./shugo-config -watch -config config/shugo.yaml
```

## Validation Rules

The configuration system includes built-in validation rules:

### Default Validator

- Validates listen address format
- Ensures positive timeouts and intervals
- Validates log levels and formats
- Checks file paths and permissions
- Validates port ranges

### Port Validator

- Detects port conflicts between APIs
- Ensures unique ports for enabled services

### Security Validator

- Warns about weak authentication tokens
- Checks for empty cluster keys
- Validates key strength requirements

### Storage Validator

- Verifies storage directory writability
- Checks for reasonable file size limits
- Validates storage configuration

## Best Practices

### Security

1. **Use Strong Keys**: Always use strong, randomly generated cluster keys in production
2. **Environment Variables**: Store sensitive configuration in environment variables, not in files
3. **TLS**: Enable TLS for production deployments
4. **Authentication**: Use strong authentication tokens

### Performance

1. **Resource Limits**: Set appropriate resource limits based on your hardware
2. **Caching**: Configure caching based on your workload
3. **Connection Pooling**: Enable connection pooling for high-throughput scenarios

### Monitoring

1. **Logging**: Use structured logging with appropriate log levels
2. **Metrics**: Enable metrics collection for monitoring
3. **Health Checks**: Configure health check endpoints

### Deployment

1. **Configuration Files**: Use configuration files for static settings
2. **Environment Variables**: Use environment variables for environment-specific settings
3. **Secrets Management**: Use proper secrets management for sensitive data
4. **Configuration Validation**: Always validate configuration before deployment

## Examples

### Development Configuration

```yaml
server:
  listen_addr: ":3000"
  debug: true

storage:
  root: "./storage"

logging:
  level: "debug"
  format: "text"

api:
  rest:
    enabled: true
    port: 8080
  graphql:
    enabled: true
    port: 8081
  grpc:
    enabled: true
    port: 8082

security:
  auth_token: "dev-token"
```

### Production Configuration

```yaml
server:
  listen_addr: ":3000"
  debug: false
  shutdown_timeout: "60s"

storage:
  root: "/data/shugo"
  max_file_size: 1073741824
  compression: true
  compression_level: 6

logging:
  level: "info"
  format: "json"
  file: "/var/log/shugo/app.log"
  structured: true
  rotation:
    max_size: 100
    max_files: 10
    max_age: "168h"
    compress: true

api:
  rest:
    enabled: true
    port: 8080
    allowed_origins:
      - "https://example.com"
    rate_limit_per_min: 1000
  graphql:
    enabled: true
    port: 8081
    enable_playground: false
  grpc:
    enabled: true
    port: 8082

security:
  tls: true
  tls_cert_file: "/etc/ssl/certs/shugo.crt"
  tls_key_file: "/etc/ssl/private/shugo.key"
  encryption_at_rest: true
  encryption_in_transit: true

performance:
  max_concurrent_streams_per_peer: 20
  stream_buffer_size: 131072
  connection_pool_size: 20
  enable_multiplexing: true
  cache_size: 500
  cache_ttl: "2h"
```

### Docker Configuration

```yaml
server:
  listen_addr: ":3000"
  debug: false

storage:
  root: "/app/data"

logging:
  level: "info"
  format: "json"

api:
  rest:
    enabled: true
    port: 8080
  graphql:
    enabled: true
    port: 8081
  grpc:
    enabled: true
    port: 8082

security:
  auth_token: "${SHUGO_AUTH_TOKEN}"
  cluster_key: "${SHUGO_CLUSTER_KEY}"

network:
  bootstrap_nodes: "${SHUGO_BOOTSTRAP_NODES}"
```

## Troubleshooting

### Common Issues

1. **Configuration Not Loading**: Check file permissions and path
2. **Validation Errors**: Review validation error messages and fix configuration
3. **Environment Variables Not Working**: Ensure variable names follow the correct pattern
4. **Hot Reloading Not Working**: Check file permissions and ensure the file exists

### Debug Configuration

```bash
# Enable debug logging
export SHUGO_LOG_LEVEL=debug

# Show configuration with debug info
./shugo-config -show -config config/shugo.yaml
```

### Validation Errors

Common validation errors and solutions:

- **Invalid listen address**: Use format `host:port` or `:port`
- **Port conflicts**: Ensure unique ports for each API
- **File size too large**: Reduce `max_file_size` or increase system limits
- **Storage not writable**: Check directory permissions
- **Weak security**: Use stronger authentication tokens and cluster keys
