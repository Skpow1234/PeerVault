package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// DefaultValidator provides default validation rules for configuration
type DefaultValidator struct{}

// Validate validates the configuration using default rules
func (v *DefaultValidator) Validate(config *Config) error {
	var errors []ValidationError

	// Validate server configuration
	if err := v.validateServer(config.Server); err != nil {
		errors = append(errors, *err)
	}

	// Validate storage configuration
	if err := v.validateStorage(config.Storage); err != nil {
		errors = append(errors, *err)
	}

	// Validate network configuration
	if err := v.validateNetwork(config.Network); err != nil {
		errors = append(errors, *err)
	}

	// Validate security configuration
	if err := v.validateSecurity(config.Security); err != nil {
		errors = append(errors, *err)
	}

	// Validate logging configuration
	if err := v.validateLogging(config.Logging); err != nil {
		errors = append(errors, *err)
	}

	// Validate API configuration
	if err := v.validateAPI(config.API); err != nil {
		errors = append(errors, *err)
	}

	// Validate peer configuration
	if err := v.validatePeer(config.Peer); err != nil {
		errors = append(errors, *err)
	}

	// Validate performance configuration
	if err := v.validatePerformance(config.Performance); err != nil {
		errors = append(errors, *err)
	}

	// Return combined errors
	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}

	return nil
}

// validateServer validates server configuration
func (v *DefaultValidator) validateServer(config ServerConfig) *ValidationError {
	// Validate listen address
	if config.ListenAddr == "" {
		return &ValidationError{Field: "server.listen_addr", Message: "listen address cannot be empty"}
	}

	// Validate listen address format
	if _, err := net.ResolveTCPAddr("tcp", config.ListenAddr); err != nil {
		return &ValidationError{Field: "server.listen_addr", Message: fmt.Sprintf("invalid listen address: %v", err)}
	}

	// Validate shutdown timeout
	if config.ShutdownTimeout <= 0 {
		return &ValidationError{Field: "server.shutdown_timeout", Message: "shutdown timeout must be positive"}
	}

	return nil
}

// validateStorage validates storage configuration
func (v *DefaultValidator) validateStorage(config StorageConfig) *ValidationError {
	// Validate storage root
	if config.Root == "" {
		return &ValidationError{Field: "storage.root", Message: "storage root cannot be empty"}
	}

	// Validate storage root is absolute or can be made absolute
	if !filepath.IsAbs(config.Root) {
		absPath, err := filepath.Abs(config.Root)
		if err != nil {
			return &ValidationError{Field: "storage.root", Message: fmt.Sprintf("invalid storage root path: %v", err)}
		}
		config.Root = absPath
	}

	// Validate max file size
	if config.MaxFileSize <= 0 {
		return &ValidationError{Field: "storage.max_file_size", Message: "max file size must be positive"}
	}

	// Validate compression level
	if config.CompressionLevel < 1 || config.CompressionLevel > 9 {
		return &ValidationError{Field: "storage.compression_level", Message: "compression level must be between 1 and 9"}
	}

	// Validate cleanup interval
	if config.CleanupInterval <= 0 {
		return &ValidationError{Field: "storage.cleanup_interval", Message: "cleanup interval must be positive"}
	}

	// Validate retention period
	if config.RetentionPeriod <= 0 {
		return &ValidationError{Field: "storage.retention_period", Message: "retention period must be positive"}
	}

	return nil
}

// validateNetwork validates network configuration
func (v *DefaultValidator) validateNetwork(config NetworkConfig) *ValidationError {
	// Validate bootstrap nodes
	for i, node := range config.BootstrapNodes {
		if node == "" {
			continue // Skip empty nodes
		}
		if _, err := net.ResolveTCPAddr("tcp", node); err != nil {
			return &ValidationError{Field: fmt.Sprintf("network.bootstrap_nodes[%d]", i), Message: fmt.Sprintf("invalid bootstrap node address: %v", err)}
		}
	}

	// Validate connection timeout
	if config.ConnectionTimeout <= 0 {
		return &ValidationError{Field: "network.connection_timeout", Message: "connection timeout must be positive"}
	}

	// Validate read timeout
	if config.ReadTimeout <= 0 {
		return &ValidationError{Field: "network.read_timeout", Message: "read timeout must be positive"}
	}

	// Validate write timeout
	if config.WriteTimeout <= 0 {
		return &ValidationError{Field: "network.write_timeout", Message: "write timeout must be positive"}
	}

	// Validate keep-alive interval
	if config.KeepAliveInterval <= 0 {
		return &ValidationError{Field: "network.keep_alive_interval", Message: "keep-alive interval must be positive"}
	}

	// Validate max message size
	if config.MaxMessageSize <= 0 {
		return &ValidationError{Field: "network.max_message_size", Message: "max message size must be positive"}
	}

	return nil
}

// validateSecurity validates security configuration
func (v *DefaultValidator) validateSecurity(config SecurityConfig) *ValidationError {
	// Validate auth token
	if config.AuthToken == "" {
		return &ValidationError{Field: "security.auth_token", Message: "auth token cannot be empty"}
	}

	// Validate TLS configuration
	if config.TLS {
		if config.TLSCertFile == "" {
			return &ValidationError{Field: "security.tls_cert_file", Message: "TLS certificate file is required when TLS is enabled"}
		}
		if config.TLSKeyFile == "" {
			return &ValidationError{Field: "security.tls_key_file", Message: "TLS key file is required when TLS is enabled"}
		}

		// Check if certificate file exists
		if _, err := os.Stat(config.TLSCertFile); os.IsNotExist(err) {
			return &ValidationError{Field: "security.tls_cert_file", Message: "TLS certificate file does not exist"}
		}

		// Check if key file exists
		if _, err := os.Stat(config.TLSKeyFile); os.IsNotExist(err) {
			return &ValidationError{Field: "security.tls_key_file", Message: "TLS key file does not exist"}
		}
	}

	// Validate key rotation interval
	if config.KeyRotationInterval <= 0 {
		return &ValidationError{Field: "security.key_rotation_interval", Message: "key rotation interval must be positive"}
	}

	return nil
}

// validateLogging validates logging configuration
func (v *DefaultValidator) validateLogging(config LoggingConfig) *ValidationError {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[strings.ToLower(config.Level)] {
		return &ValidationError{Field: "logging.level", Message: "log level must be one of: debug, info, warn, error"}
	}

	// Validate log format
	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[strings.ToLower(config.Format)] {
		return &ValidationError{Field: "logging.format", Message: "log format must be one of: json, text"}
	}

	// Validate log file path if specified
	if config.File != "" {
		dir := filepath.Dir(config.File)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return &ValidationError{Field: "logging.file", Message: fmt.Sprintf("log file directory does not exist: %s", dir)}
		}
	}

	// Validate log rotation settings
	if err := v.validateLogRotation(config.Rotation); err != nil {
		return err
	}

	return nil
}

// validateLogRotation validates log rotation configuration
func (v *DefaultValidator) validateLogRotation(config LogRotationConfig) *ValidationError {
	// Validate max size
	if config.MaxSize <= 0 {
		return &ValidationError{Field: "logging.rotation.max_size", Message: "max size must be positive"}
	}

	// Validate max files
	if config.MaxFiles <= 0 {
		return &ValidationError{Field: "logging.rotation.max_files", Message: "max files must be positive"}
	}

	// Validate max age
	if config.MaxAge <= 0 {
		return &ValidationError{Field: "logging.rotation.max_age", Message: "max age must be positive"}
	}

	return nil
}

// validateAPI validates API configuration
func (v *DefaultValidator) validateAPI(config APIConfig) *ValidationError {
	// Validate REST configuration
	if err := v.validateREST(config.REST); err != nil {
		return err
	}

	// Validate GraphQL configuration
	if err := v.validateGraphQL(config.GraphQL); err != nil {
		return err
	}

	// Validate gRPC configuration
	if err := v.validateGRPC(config.GRPC); err != nil {
		return err
	}

	return nil
}

// validateREST validates REST API configuration
func (v *DefaultValidator) validateREST(config RESTConfig) *ValidationError {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	// Validate port
	if config.Port <= 0 || config.Port > 65535 {
		return &ValidationError{Field: "api.rest.port", Message: "port must be between 1 and 65535"}
	}

	// Validate rate limit
	if config.RateLimitPerMin <= 0 {
		return &ValidationError{Field: "api.rest.rate_limit_per_min", Message: "rate limit must be positive"}
	}

	// Validate auth token
	if config.AuthToken == "" {
		return &ValidationError{Field: "api.rest.auth_token", Message: "auth token cannot be empty"}
	}

	return nil
}

// validateGraphQL validates GraphQL API configuration
func (v *DefaultValidator) validateGraphQL(config GraphQLConfig) *ValidationError {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	// Validate port
	if config.Port <= 0 || config.Port > 65535 {
		return &ValidationError{Field: "api.graphql.port", Message: "port must be between 1 and 65535"}
	}

	// Validate GraphQL path
	if config.GraphQLPath == "" {
		return &ValidationError{Field: "api.graphql.graphql_path", Message: "GraphQL path cannot be empty"}
	}

	// Validate playground path
	if config.EnablePlayground && config.PlaygroundPath == "" {
		return &ValidationError{Field: "api.graphql.playground_path", Message: "playground path cannot be empty when playground is enabled"}
	}

	return nil
}

// validateGRPC validates gRPC API configuration
func (v *DefaultValidator) validateGRPC(config GRPCConfig) *ValidationError {
	if !config.Enabled {
		return nil // Skip validation if disabled
	}

	// Validate port
	if config.Port <= 0 || config.Port > 65535 {
		return &ValidationError{Field: "api.grpc.port", Message: "port must be between 1 and 65535"}
	}

	// Validate auth token
	if config.AuthToken == "" {
		return &ValidationError{Field: "api.grpc.auth_token", Message: "auth token cannot be empty"}
	}

	// Validate max concurrent streams
	if config.MaxConcurrentStreams <= 0 {
		return &ValidationError{Field: "api.grpc.max_concurrent_streams", Message: "max concurrent streams must be positive"}
	}

	return nil
}

// validatePeer validates peer configuration
func (v *DefaultValidator) validatePeer(config PeerConfig) *ValidationError {
	// Validate max peers
	if config.MaxPeers <= 0 {
		return &ValidationError{Field: "peer.max_peers", Message: "max peers must be positive"}
	}

	// Validate heartbeat interval
	if config.HeartbeatInterval <= 0 {
		return &ValidationError{Field: "peer.heartbeat_interval", Message: "heartbeat interval must be positive"}
	}

	// Validate health timeout
	if config.HealthTimeout <= 0 {
		return &ValidationError{Field: "peer.health_timeout", Message: "health timeout must be positive"}
	}

	// Validate reconnect backoff
	if config.ReconnectBackoff <= 0 {
		return &ValidationError{Field: "peer.reconnect_backoff", Message: "reconnect backoff must be positive"}
	}

	// Validate max reconnect attempts
	if config.MaxReconnectAttempts <= 0 {
		return &ValidationError{Field: "peer.max_reconnect_attempts", Message: "max reconnect attempts must be positive"}
	}

	return nil
}

// validatePerformance validates performance configuration
func (v *DefaultValidator) validatePerformance(config PerformanceConfig) *ValidationError {
	// Validate max concurrent streams per peer
	if config.MaxConcurrentStreamsPerPeer <= 0 {
		return &ValidationError{Field: "performance.max_concurrent_streams_per_peer", Message: "max concurrent streams per peer must be positive"}
	}

	// Validate stream buffer size
	if config.StreamBufferSize <= 0 {
		return &ValidationError{Field: "performance.stream_buffer_size", Message: "stream buffer size must be positive"}
	}

	// Validate connection pool size
	if config.ConnectionPoolSize <= 0 {
		return &ValidationError{Field: "performance.connection_pool_size", Message: "connection pool size must be positive"}
	}

	// Validate cache size
	if config.CacheSize < 0 {
		return &ValidationError{Field: "performance.cache_size", Message: "cache size cannot be negative"}
	}

	// Validate cache TTL
	if config.CacheTTL < 0 {
		return &ValidationError{Field: "performance.cache_ttl", Message: "cache TTL cannot be negative"}
	}

	return nil
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}

	return fmt.Sprintf("validation errors:\n%s", strings.Join(messages, "\n"))
}

// Custom validators

// PortValidator validates that ports are not conflicting
type PortValidator struct{}

// Validate checks for port conflicts in the configuration
func (v *PortValidator) Validate(config *Config) error {
	ports := make(map[int]string)

	// Collect all enabled API ports
	if config.API.REST.Enabled {
		if existing, exists := ports[config.API.REST.Port]; exists {
			return fmt.Errorf("port conflict: %s and REST API both use port %d", existing, config.API.REST.Port)
		}
		ports[config.API.REST.Port] = "REST API"
	}
	if config.API.GraphQL.Enabled {
		if existing, exists := ports[config.API.GraphQL.Port]; exists {
			return fmt.Errorf("port conflict: %s and GraphQL API both use port %d", existing, config.API.GraphQL.Port)
		}
		ports[config.API.GraphQL.Port] = "GraphQL API"
	}
	if config.API.GRPC.Enabled {
		if existing, exists := ports[config.API.GRPC.Port]; exists {
			return fmt.Errorf("port conflict: %s and gRPC API both use port %d", existing, config.API.GRPC.Port)
		}
		ports[config.API.GRPC.Port] = "gRPC API"
	}

	return nil
}

// SecurityValidator validates security-related configuration
type SecurityValidator struct {
	AllowDemoToken bool // Allow demo token in production
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(allowDemoToken bool) *SecurityValidator {
	return &SecurityValidator{
		AllowDemoToken: allowDemoToken,
	}
}

// Validate checks security configuration for potential issues
func (v *SecurityValidator) Validate(config *Config) error {
	// Check for weak auth tokens (only if demo token is not allowed)
	// Use configuration setting if not explicitly set in validator
	allowDemo := v.AllowDemoToken
	if !allowDemo {
		allowDemo = config.Security.AllowDemoToken
	}
	
	if !allowDemo && config.Security.AuthToken == "demo-token" {
		return fmt.Errorf("security warning: using default demo token in production")
	}

	// Check for empty cluster key in production (only if demo tokens are not allowed)
	if !allowDemo && config.Security.ClusterKey == "" {
		return fmt.Errorf("security warning: no cluster key specified")
	}

	// Check for weak cluster key
	if len(config.Security.ClusterKey) < 32 {
		return fmt.Errorf("security warning: cluster key should be at least 32 characters long")
	}

	return nil
}

// StorageValidator validates storage configuration
type StorageValidator struct{}

// Validate checks storage configuration for potential issues
func (v *StorageValidator) Validate(config *Config) error {
	// Check if storage directory is writable
	if err := v.checkStorageWritable(config.Storage.Root); err != nil {
		return fmt.Errorf("storage validation failed: %w", err)
	}

	// Check for reasonable file size limits
	if config.Storage.MaxFileSize > 10*1024*1024*1024 { // 10GB
		return fmt.Errorf("storage warning: max file size is very large (%d bytes)", config.Storage.MaxFileSize)
	}

	return nil
}

// checkStorageWritable checks if the storage directory is writable
func (v *StorageValidator) checkStorageWritable(path string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Test write access
	testFile := filepath.Join(path, ".test_write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("storage directory is not writable: %w", err)
	}

	// Clean up test file
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to clean up test file: %w", err)
	}

	return nil
}
