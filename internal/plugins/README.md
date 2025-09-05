# PeerVault Plugin Architecture

The PeerVault plugin system allows you to extend the core functionality with custom storage backends, authentication methods, transport protocols, and more.

## Plugin Types

### Storage Plugins

- **Purpose**: Custom storage backends (S3, Azure Blob, Google Cloud Storage, etc.)
- **Interface**: `StoragePlugin`
- **Location**: `internal/plugins/storage/`

### Authentication Plugins

- **Purpose**: Custom authentication methods (OAuth, LDAP, SAML, etc.)
- **Interface**: `AuthPlugin`
- **Location**: `internal/plugins/auth/`

### Transport Plugins

- **Purpose**: Custom transport protocols (WebRTC, QUIC, etc.)
- **Interface**: `TransportPlugin`
- **Location**: `internal/plugins/transport/`

### Processing Plugins

- **Purpose**: File processing (compression, encryption, virus scanning, etc.)
- **Interface**: `ProcessingPlugin`
- **Location**: `internal/plugins/processing/`

## Plugin Interface

All plugins must implement the base `Plugin` interface:

```go
type Plugin interface {
    // Plugin metadata
    Name() string
    Version() string
    Description() string
    
    // Lifecycle methods
    Initialize(config map[string]interface{}) error
    Start() error
    Stop() error
    
    // Configuration
    GetConfigSchema() map[string]interface{}
    ValidateConfig(config map[string]interface{}) error
}
```

## Plugin Discovery

Plugins are discovered automatically through:

1. **Built-in plugins**: Located in `internal/plugins/`
2. **External plugins**: Located in `plugins/` directory
3. **Dynamic loading**: Load plugins at runtime from specified paths

## Configuration

Plugins are configured in the main configuration file:

```yaml
plugins:
  enabled:
    - name: "s3-storage"
      type: "storage"
      config:
        bucket: "my-bucket"
        region: "us-west-2"
        access_key: "${AWS_ACCESS_KEY}"
        secret_key: "${AWS_SECRET_KEY}"
    
    - name: "oauth-auth"
      type: "auth"
      config:
        provider: "google"
        client_id: "${GOOGLE_CLIENT_ID}"
        client_secret: "${GOOGLE_CLIENT_SECRET}"
        redirect_url: "http://localhost:8080/auth/callback"
```

## Plugin Development

### Creating a Storage Plugin

```go
package main

import (
    "context"
    "io"
    
    "github.com/peervault/peervault/internal/plugins"
)

type S3StoragePlugin struct {
    bucket string
    region string
    // ... other fields
}

func (p *S3StoragePlugin) Name() string {
    return "s3-storage"
}

func (p *S3StoragePlugin) Version() string {
    return "1.0.0"
}

func (p *S3StoragePlugin) Description() string {
    return "Amazon S3 storage backend"
}

func (p *S3StoragePlugin) Initialize(config map[string]interface{}) error {
    // Initialize S3 client
    p.bucket = config["bucket"].(string)
    p.region = config["region"].(string)
    return nil
}

func (p *S3StoragePlugin) Start() error {
    // Start the plugin
    return nil
}

func (p *S3StoragePlugin) Stop() error {
    // Cleanup resources
    return nil
}

func (p *S3StoragePlugin) GetConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "bucket": map[string]interface{}{
            "type": "string",
            "required": true,
            "description": "S3 bucket name",
        },
        "region": map[string]interface{}{
            "type": "string",
            "required": true,
            "description": "AWS region",
        },
    }
}

func (p *S3StoragePlugin) ValidateConfig(config map[string]interface{}) error {
    // Validate configuration
    return nil
}

// StoragePlugin interface methods
func (p *S3StoragePlugin) Store(ctx context.Context, key string, data io.Reader) error {
    // Store data in S3
    return nil
}

func (p *S3StoragePlugin) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
    // Retrieve data from S3
    return nil, nil
}

func (p *S3StoragePlugin) Delete(ctx context.Context, key string) error {
    // Delete data from S3
    return nil
}

func (p *S3StoragePlugin) List(ctx context.Context, prefix string) ([]string, error) {
    // List objects in S3
    return nil, nil
}

// Register the plugin
func init() {
    plugins.RegisterStoragePlugin(&S3StoragePlugin{})
}
```

### Creating an Authentication Plugin

```go
package main

import (
    "context"
    
    "github.com/peervault/peervault/internal/plugins"
)

type OAuthAuthPlugin struct {
    provider string
    clientID string
    clientSecret string
    // ... other fields
}

func (p *OAuthAuthPlugin) Name() string {
    return "oauth-auth"
}

func (p *OAuthAuthPlugin) Version() string {
    return "1.0.0"
}

func (p *OAuthAuthPlugin) Description() string {
    return "OAuth 2.0 authentication provider"
}

func (p *OAuthAuthPlugin) Initialize(config map[string]interface{}) error {
    p.provider = config["provider"].(string)
    p.clientID = config["client_id"].(string)
    p.clientSecret = config["client_secret"].(string)
    return nil
}

func (p *OAuthAuthPlugin) Start() error {
    return nil
}

func (p *OAuthAuthPlugin) Stop() error {
    return nil
}

func (p *OAuthAuthPlugin) GetConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "provider": map[string]interface{}{
            "type": "string",
            "required": true,
            "description": "OAuth provider (google, github, etc.)",
        },
        "client_id": map[string]interface{}{
            "type": "string",
            "required": true,
            "description": "OAuth client ID",
        },
        "client_secret": map[string]interface{}{
            "type": "string",
            "required": true,
            "description": "OAuth client secret",
        },
    }
}

func (p *OAuthAuthPlugin) ValidateConfig(config map[string]interface{}) error {
    return nil
}

// AuthPlugin interface methods
func (p *OAuthAuthPlugin) Authenticate(ctx context.Context, credentials map[string]interface{}) (*plugins.AuthResult, error) {
    // Authenticate user with OAuth
    return nil, nil
}

func (p *OAuthAuthPlugin) Authorize(ctx context.Context, userID string, resource string, action string) (bool, error) {
    // Check authorization
    return true, nil
}

func (p *OAuthAuthPlugin) GetUserInfo(ctx context.Context, userID string) (*plugins.UserInfo, error) {
    // Get user information
    return nil, nil
}

// Register the plugin
func init() {
    plugins.RegisterAuthPlugin(&OAuthAuthPlugin{})
}
```

## Plugin Management

### Loading Plugins

```go
package main

import (
    "github.com/peervault/peervault/internal/plugins"
)

func main() {
    // Load plugins from configuration
    pluginManager := plugins.NewManager()
    
    // Load built-in plugins
    pluginManager.LoadBuiltinPlugins()
    
    // Load external plugins
    pluginManager.LoadExternalPlugins("./plugins")
    
    // Initialize and start plugins
    pluginManager.InitializeAll(config)
    pluginManager.StartAll()
    
    // ... rest of application
}
```

### Plugin Lifecycle

1. **Discovery**: Plugins are discovered and registered
2. **Initialization**: Plugins are initialized with configuration
3. **Start**: Plugins are started and ready to handle requests
4. **Runtime**: Plugins handle requests and operations
5. **Stop**: Plugins are stopped gracefully
6. **Cleanup**: Resources are cleaned up

## Built-in Plugins

### Available Storage Plugins

- **Local Storage**: File system storage (default)
- **Memory Storage**: In-memory storage for testing
- **S3 Storage**: Amazon S3 integration
- **Azure Blob**: Azure Blob Storage integration
- **Google Cloud**: Google Cloud Storage integration

### Available Authentication Plugins

- **JWT Auth**: JWT-based authentication (default)
- **OAuth 2.0**: OAuth 2.0 integration
- **LDAP**: LDAP authentication
- **SAML**: SAML authentication
- **API Key**: API key authentication

### Available Transport Plugins

- **TCP**: TCP transport (default)
- **WebRTC**: WebRTC transport for browser clients
- **QUIC**: QUIC transport for improved performance
- **WebSocket**: WebSocket transport for real-time communication

### Available Processing Plugins

- **Compression**: File compression (gzip, lz4, etc.)
- **Encryption**: File encryption
- **Virus Scanning**: Virus scanning integration
- **Image Processing**: Image resizing, format conversion
- **Video Processing**: Video transcoding, thumbnail generation

## Plugin Development Tools

### Plugin Generator

```bash
# Generate a new plugin
peervault plugin generate --type storage --name my-storage-plugin

# Generate with template
peervault plugin generate --type auth --name my-auth-plugin --template oauth
```

### Plugin Testing

```bash
# Test a plugin
peervault plugin test ./plugins/my-plugin

# Run plugin tests
go test ./plugins/my-plugin/...
```

### Plugin Validation

```bash
# Validate plugin configuration
peervault plugin validate ./plugins/my-plugin/config.yaml

# Check plugin compatibility
peervault plugin check ./plugins/my-plugin
```

## Best Practices

### Plugin Design

1. **Single Responsibility**: Each plugin should have a single, well-defined purpose
2. **Stateless**: Plugins should be stateless when possible
3. **Error Handling**: Implement proper error handling and logging
4. **Configuration**: Use structured configuration with validation
5. **Testing**: Write comprehensive tests for your plugins

### Performance

1. **Connection Pooling**: Use connection pooling for external services
2. **Caching**: Implement caching where appropriate
3. **Async Operations**: Use async operations for I/O-bound tasks
4. **Resource Management**: Properly manage resources and cleanup

### Security

1. **Input Validation**: Validate all inputs
2. **Secret Management**: Use secure secret management
3. **Access Control**: Implement proper access controls
4. **Audit Logging**: Log security-relevant events

## Examples

See the `plugins/` directory for example plugins:

- [S3 Storage Plugin](plugins/s3-storage/) - Amazon S3 integration
- [OAuth Auth Plugin](plugins/oauth-auth/) - OAuth 2.0 authentication
- [WebRTC Transport Plugin](plugins/webrtc-transport/) - WebRTC transport
- [Compression Plugin](plugins/compression/) - File compression

## Support

- **Documentation**: [Plugin Development Guide](../docs/plugins/)
- **Examples**: [Plugin Examples](../examples/plugins/)
- **Community**: [Discord Server](https://discord.gg/peervault)
- **Issues**: [GitHub Issues](https://github.com/peervault/peervault/issues)
