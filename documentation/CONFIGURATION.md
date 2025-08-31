# PeerVault Configuration Management

## Overview

PeerVault provides a comprehensive configuration management system that supports hierarchical configuration loading, environment variable overrides, validation, and hot reloading. The system is designed to be flexible, secure, and easy to use across different deployment environments.

## Features

- **Hierarchical Configuration**: Load from YAML/JSON files with environment variable overrides
- **Environment Variable Support**: Override any configuration value using environment variables
- **Validation**: Built-in validation with custom validation rules
- **Hot Reloading**: Watch configuration files for changes and reload automatically
- **Type Safety**: Strongly typed configuration structures
- **Documentation**: Self-documenting configuration with examples

## Configuration Structure

The PeerVault configuration is organized into logical sections:

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
  # Cluster key for encryption (set via environment variable PEERVAULT_CLUSTER_KEY)
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

All configuration values can be overridden using environment variables. The environment variable names follow the pattern `PEERVAULT_<SECTION>_<FIELD>`.

### Server Environment Variables

- `PEERVAULT_NODE_ID` - Node ID
- `PEERVAULT_LISTEN_ADDR` - Listen address
- `PEERVAULT_DEBUG` - Enable debug mode
- `PEERVAULT_SHUTDOWN_TIMEOUT` - Shutdown timeout

### Storage Environment Variables

- `PEERVAULT_STORAGE_ROOT` - Storage root directory
- `PEERVAULT_MAX_FILE_SIZE` - Maximum file size
- `PEERVAULT_COMPRESSION` - Enable compression
- `PEERVAULT_COMPRESSION_LEVEL` - Compression level
- `PEERVAULT_DEDUPLICATION` - Enable deduplication
- `PEERVAULT_CLEANUP_INTERVAL` - Cleanup interval
- `PEERVAULT_RETENTION_PERIOD` - Retention period

### Network Environment Variables

- `PEERVAULT_BOOTSTRAP_NODES` - Bootstrap nodes (comma-separated)
- `PEERVAULT_CONNECTION_TIMEOUT` - Connection timeout
- `PEERVAULT_READ_TIMEOUT` - Read timeout
- `PEERVAULT_WRITE_TIMEOUT` - Write timeout
- `PEERVAULT_KEEP_ALIVE_INTERVAL` - Keep-alive interval
- `PEERVAULT_MAX_MESSAGE_SIZE` - Maximum message size

### Security Environment Variables

- `PEERVAULT_CLUSTER_KEY` - Cluster key for encryption
- `PEERVAULT_AUTH_TOKEN` - Authentication token
- `PEERVAULT_TLS` - Enable TLS
- `PEERVAULT_TLS_CERT_FILE` - TLS certificate file
- `PEERVAULT_TLS_KEY_FILE` - TLS key file
- `PEERVAULT_KEY_ROTATION_INTERVAL` - Key rotation interval
- `PEERVAULT_ENCRYPTION_AT_REST` - Enable encryption at rest
- `PEERVAULT_ENCRYPTION_IN_TRANSIT` - Enable encryption in transit

### Logging Environment Variables

- `PEERVAULT_LOG_LEVEL` - Log level (debug, info, warn, error)
- `PEERVAULT_LOG_FORMAT` - Log format (json, text)
- `PEERVAULT_LOG_FILE` - Log file path
- `PEERVAULT_LOG_STRUCTURED` - Enable structured logging
- `PEERVAULT_LOG_INCLUDE_SOURCE` - Include source location
- `PEERVAULT_LOG_MAX_SIZE` - Max log file size (MB)
- `PEERVAULT_LOG_MAX_FILES` - Max log files to keep
- `PEERVAULT_LOG_MAX_AGE` - Max log file age
- `PEERVAULT_LOG_COMPRESS` - Compress rotated logs

### API Environment Variables

#### REST API

- `PEERVAULT_REST_ENABLED` - Enable REST API
- `PEERVAULT_REST_PORT` - REST API port
- `PEERVAULT_REST_ALLOWED_ORIGINS` - Allowed origins (comma-separated)
- `PEERVAULT_REST_RATE_LIMIT` - Rate limit per minute
- `PEERVAULT_REST_AUTH_TOKEN` - REST auth token

#### GraphQL API

- `PEERVAULT_GRAPHQL_ENABLED` - Enable GraphQL API
- `PEERVAULT_GRAPHQL_PORT` - GraphQL API port
- `PEERVAULT_GRAPHQL_PLAYGROUND` - Enable GraphQL Playground
- `PEERVAULT_GRAPHQL_PATH` - GraphQL endpoint path
- `PEERVAULT_GRAPHQL_PLAYGROUND_PATH` - Playground path
- `PEERVAULT_GRAPHQL_ALLOWED_ORIGINS` - Allowed origins (comma-separated)

#### gRPC API

- `PEERVAULT_GRPC_ENABLED` - Enable gRPC API
- `PEERVAULT_GRPC_PORT` - gRPC API port
- `PEERVAULT_GRPC_AUTH_TOKEN` - gRPC auth token
- `PEERVAULT_GRPC_REFLECTION` - Enable reflection
- `PEERVAULT_GRPC_MAX_STREAMS` - Max concurrent streams

### Peer Environment Variables

- `PEERVAULT_MAX_PEERS` - Maximum number of peers
- `PEERVAULT_HEARTBEAT_INTERVAL` - Heartbeat interval
- `PEERVAULT_HEALTH_TIMEOUT` - Health timeout
- `PEERVAULT_RECONNECT_BACKOFF` - Reconnection backoff
- `PEERVAULT_MAX_RECONNECT_ATTEMPTS` - Max reconnection attempts

### Performance Environment Variables

- `PEERVAULT_MAX_STREAMS_PER_PEER` - Max concurrent streams per peer
- `PEERVAULT_STREAM_BUFFER_SIZE` - Stream buffer size
- `PEERVAULT_CONNECTION_POOL_SIZE` - Connection pool size
- `PEERVAULT_ENABLE_MULTIPLEXING` - Enable connection multiplexing
- `PEERVAULT_CACHE_SIZE` - Cache size (MB)
- `PEERVAULT_CACHE_TTL` - Cache TTL

## Usage

### Basic Configuration Loading

```go
package main

import (
    "log"
    "github.com/Skpow1234/Peervault/internal/config"
)

func main() {
    // Create configuration manager
    manager := config.NewManager("config/peervault.yaml")
    
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

```bash
# Override configuration using environment variables
export PEERVAULT_LISTEN_ADDR=":8080"
export PEERVAULT_LOG_LEVEL="debug"
export PEERVAULT_STORAGE_ROOT="/data/peervault"
export PEERVAULT_CLUSTER_KEY="your-secure-cluster-key"

# Run the application
go run main.go
```

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

PeerVault includes a command-line tool for configuration management:

### Generate Default Configuration

```bash
# Generate default YAML configuration
./peervault-config -generate -format yaml -output config/peervault.yaml

# Generate JSON configuration
./peervault-config -generate -format json -output config/peervault.json
```

### Validate Configuration

```bash
# Validate configuration file
./peervault-config -validate -config config/peervault.yaml
```

### Show Configuration

```bash
# Show current configuration in YAML format
./peervault-config -show -format yaml

# Show current configuration in JSON format
./peervault-config -show -format json
```

### Show Environment Mappings

```bash
# Show all environment variable mappings
./peervault-config -env
```

### Watch Configuration

```bash
# Watch configuration file for changes
./peervault-config -watch -config config/peervault.yaml
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
  root: "/data/peervault"
  max_file_size: 1073741824
  compression: true
  compression_level: 6

logging:
  level: "info"
  format: "json"
  file: "/var/log/peervault/app.log"
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
  tls_cert_file: "/etc/ssl/certs/peervault.crt"
  tls_key_file: "/etc/ssl/private/peervault.key"
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
  auth_token: "${PEERVAULT_AUTH_TOKEN}"
  cluster_key: "${PEERVAULT_CLUSTER_KEY}"

network:
  bootstrap_nodes: "${PEERVAULT_BOOTSTRAP_NODES}"
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
export PEERVAULT_LOG_LEVEL=debug

# Show configuration with debug info
./peervault-config -show -config config/peervault.yaml
```

### Validation Errors

Common validation errors and solutions:

- **Invalid listen address**: Use format `host:port` or `:port`
- **Port conflicts**: Ensure unique ports for each API
- **File size too large**: Reduce `max_file_size` or increase system limits
- **Storage not writable**: Check directory permissions
- **Weak security**: Use stronger authentication tokens and cluster keys
