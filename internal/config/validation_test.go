package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "test.field",
		Message: "test error message",
	}

	expected := "validation error for field 'test.field': test error message"
	assert.Equal(t, expected, err.Error())
}

func TestValidationWarning(t *testing.T) {
	warning := ValidationWarning{
		Field:   "test.field",
		Message: "test warning message",
	}

	expected := "validation warning for field 'test.field': test warning message"
	assert.Equal(t, expected, warning.Error())
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{}

	// Test initial state
	assert.False(t, result.HasErrors())
	assert.False(t, result.HasWarnings())
	assert.Equal(t, "no validation issues", result.Error())

	// Add error
	result.AddError("field1", "error message 1")
	assert.True(t, result.HasErrors())
	assert.False(t, result.HasWarnings())

	// Add warning
	result.AddWarning("field2", "warning message 1")
	assert.True(t, result.HasErrors())
	assert.True(t, result.HasWarnings())

	// Test error message format
	errorMsg := result.Error()
	assert.Contains(t, errorMsg, "validation errors:")
	assert.Contains(t, errorMsg, "validation warnings:")
	assert.Contains(t, errorMsg, "error message 1")
	assert.Contains(t, errorMsg, "warning message 1")
}

func TestDefaultValidator_Validate_ValidConfig(t *testing.T) {
	validator := &DefaultValidator{}
	config := DefaultConfig()

	err := validator.Validate(config)
	assert.NoError(t, err)
}

func TestDefaultValidator_ValidateServer(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   ServerConfig
		hasError bool
		field    string
	}{
		{
			name: "valid server config",
			config: ServerConfig{
				ListenAddr:      ":3000",
				ShutdownTimeout: 30 * time.Second,
			},
			hasError: false,
		},
		{
			name: "empty listen address",
			config: ServerConfig{
				ListenAddr:      "",
				ShutdownTimeout: 30 * time.Second,
			},
			hasError: true,
			field:    "server.listen_addr",
		},
		{
			name: "invalid listen address",
			config: ServerConfig{
				ListenAddr:      "invalid-address",
				ShutdownTimeout: 30 * time.Second,
			},
			hasError: true,
			field:    "server.listen_addr",
		},
		{
			name: "zero shutdown timeout",
			config: ServerConfig{
				ListenAddr:      ":3000",
				ShutdownTimeout: 0,
			},
			hasError: true,
			field:    "server.shutdown_timeout",
		},
		{
			name: "negative shutdown timeout",
			config: ServerConfig{
				ListenAddr:      ":3000",
				ShutdownTimeout: -1 * time.Second,
			},
			hasError: true,
			field:    "server.shutdown_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateServer(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateStorage(t *testing.T) {
	validator := &DefaultValidator{}
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		config   StorageConfig
		hasError bool
		field    string
	}{
		{
			name: "valid storage config",
			config: StorageConfig{
				Root:             tempDir,
				MaxFileSize:      1024 * 1024,
				CompressionLevel: 6,
				CleanupInterval:  time.Hour,
				RetentionPeriod:  24 * time.Hour,
			},
			hasError: false,
		},
		{
			name: "empty storage root",
			config: StorageConfig{
				Root:             "",
				MaxFileSize:      1024 * 1024,
				CompressionLevel: 6,
				CleanupInterval:  time.Hour,
				RetentionPeriod:  24 * time.Hour,
			},
			hasError: true,
			field:    "storage.root",
		},
		{
			name: "zero max file size",
			config: StorageConfig{
				Root:             tempDir,
				MaxFileSize:      0,
				CompressionLevel: 6,
				CleanupInterval:  time.Hour,
				RetentionPeriod:  24 * time.Hour,
			},
			hasError: true,
			field:    "storage.max_file_size",
		},
		{
			name: "invalid compression level - too low",
			config: StorageConfig{
				Root:             tempDir,
				MaxFileSize:      1024 * 1024,
				CompressionLevel: 0,
				CleanupInterval:  time.Hour,
				RetentionPeriod:  24 * time.Hour,
			},
			hasError: true,
			field:    "storage.compression_level",
		},
		{
			name: "invalid compression level - too high",
			config: StorageConfig{
				Root:             tempDir,
				MaxFileSize:      1024 * 1024,
				CompressionLevel: 10,
				CleanupInterval:  time.Hour,
				RetentionPeriod:  24 * time.Hour,
			},
			hasError: true,
			field:    "storage.compression_level",
		},
		{
			name: "zero cleanup interval",
			config: StorageConfig{
				Root:             tempDir,
				MaxFileSize:      1024 * 1024,
				CompressionLevel: 6,
				CleanupInterval:  0,
				RetentionPeriod:  24 * time.Hour,
			},
			hasError: true,
			field:    "storage.cleanup_interval",
		},
		{
			name: "zero retention period",
			config: StorageConfig{
				Root:             tempDir,
				MaxFileSize:      1024 * 1024,
				CompressionLevel: 6,
				CleanupInterval:  time.Hour,
				RetentionPeriod:  0,
			},
			hasError: true,
			field:    "storage.retention_period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateStorage(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateNetwork(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   NetworkConfig
		hasError bool
		field    string
	}{
		{
			name: "valid network config",
			config: NetworkConfig{
				BootstrapNodes:    []string{"127.0.0.1:3000", "127.0.0.1:3001"},
				ConnectionTimeout: time.Second,
				ReadTimeout:       time.Second,
				WriteTimeout:      time.Second,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    1024,
			},
			hasError: false,
		},
		{
			name: "invalid bootstrap node",
			config: NetworkConfig{
				BootstrapNodes:    []string{"invalid-address"},
				ConnectionTimeout: time.Second,
				ReadTimeout:       time.Second,
				WriteTimeout:      time.Second,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    1024,
			},
			hasError: true,
			field:    "network.bootstrap_nodes[0]",
		},
		{
			name: "empty bootstrap nodes are skipped",
			config: NetworkConfig{
				BootstrapNodes:    []string{"", "127.0.0.1:3000"},
				ConnectionTimeout: time.Second,
				ReadTimeout:       time.Second,
				WriteTimeout:      time.Second,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    1024,
			},
			hasError: false,
		},
		{
			name: "zero connection timeout",
			config: NetworkConfig{
				BootstrapNodes:    []string{},
				ConnectionTimeout: 0,
				ReadTimeout:       time.Second,
				WriteTimeout:      time.Second,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    1024,
			},
			hasError: true,
			field:    "network.connection_timeout",
		},
		{
			name: "zero read timeout",
			config: NetworkConfig{
				BootstrapNodes:    []string{},
				ConnectionTimeout: time.Second,
				ReadTimeout:       0,
				WriteTimeout:      time.Second,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    1024,
			},
			hasError: true,
			field:    "network.read_timeout",
		},
		{
			name: "zero write timeout",
			config: NetworkConfig{
				BootstrapNodes:    []string{},
				ConnectionTimeout: time.Second,
				ReadTimeout:       time.Second,
				WriteTimeout:      0,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    1024,
			},
			hasError: true,
			field:    "network.write_timeout",
		},
		{
			name: "zero keep-alive interval",
			config: NetworkConfig{
				BootstrapNodes:    []string{},
				ConnectionTimeout: time.Second,
				ReadTimeout:       time.Second,
				WriteTimeout:      time.Second,
				KeepAliveInterval: 0,
				MaxMessageSize:    1024,
			},
			hasError: true,
			field:    "network.keep_alive_interval",
		},
		{
			name: "zero max message size",
			config: NetworkConfig{
				BootstrapNodes:    []string{},
				ConnectionTimeout: time.Second,
				ReadTimeout:       time.Second,
				WriteTimeout:      time.Second,
				KeepAliveInterval: time.Second,
				MaxMessageSize:    0,
			},
			hasError: true,
			field:    "network.max_message_size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateNetwork(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateSecurity(t *testing.T) {
	validator := &DefaultValidator{}
	tempDir := t.TempDir()

	// Create test certificate and key files
	certFile := filepath.Join(tempDir, "cert.pem")
	keyFile := filepath.Join(tempDir, "key.pem")

	require.NoError(t, os.WriteFile(certFile, []byte("test cert"), 0644))
	require.NoError(t, os.WriteFile(keyFile, []byte("test key"), 0644))

	tests := []struct {
		name     string
		config   SecurityConfig
		hasError bool
		field    string
	}{
		{
			name: "valid security config without TLS",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 false,
				KeyRotationInterval: time.Hour,
			},
			hasError: false,
		},
		{
			name: "valid security config with TLS",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 true,
				TLSCertFile:         certFile,
				TLSKeyFile:          keyFile,
				KeyRotationInterval: time.Hour,
			},
			hasError: false,
		},
		{
			name: "empty auth token",
			config: SecurityConfig{
				AuthToken:           "",
				TLS:                 false,
				KeyRotationInterval: time.Hour,
			},
			hasError: true,
			field:    "security.auth_token",
		},
		{
			name: "TLS enabled but no cert file",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 true,
				TLSCertFile:         "",
				TLSKeyFile:          keyFile,
				KeyRotationInterval: time.Hour,
			},
			hasError: true,
			field:    "security.tls_cert_file",
		},
		{
			name: "TLS enabled but no key file",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 true,
				TLSCertFile:         certFile,
				TLSKeyFile:          "",
				KeyRotationInterval: time.Hour,
			},
			hasError: true,
			field:    "security.tls_key_file",
		},
		{
			name: "TLS enabled but cert file doesn't exist",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 true,
				TLSCertFile:         "/nonexistent/cert.pem",
				TLSKeyFile:          keyFile,
				KeyRotationInterval: time.Hour,
			},
			hasError: true,
			field:    "security.tls_cert_file",
		},
		{
			name: "TLS enabled but key file doesn't exist",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 true,
				TLSCertFile:         certFile,
				TLSKeyFile:          "/nonexistent/key.pem",
				KeyRotationInterval: time.Hour,
			},
			hasError: true,
			field:    "security.tls_key_file",
		},
		{
			name: "zero key rotation interval",
			config: SecurityConfig{
				AuthToken:           "valid-token",
				TLS:                 false,
				KeyRotationInterval: 0,
			},
			hasError: true,
			field:    "security.key_rotation_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSecurity(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateLogging(t *testing.T) {
	validator := &DefaultValidator{}
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		config   LoggingConfig
		hasError bool
		field    string
	}{
		{
			name: "valid logging config",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				File:   "",
				Rotation: LogRotationConfig{
					MaxSize:  100 * 1024 * 1024,
					MaxFiles: 5,
					MaxAge:   7 * 24 * time.Hour,
				},
			},
			hasError: false,
		},
		{
			name: "invalid log level",
			config: LoggingConfig{
				Level:  "invalid",
				Format: "json",
				File:   "",
				Rotation: LogRotationConfig{
					MaxSize:  100 * 1024 * 1024,
					MaxFiles: 5,
					MaxAge:   7 * 24 * time.Hour,
				},
			},
			hasError: true,
			field:    "logging.level",
		},
		{
			name: "case insensitive log level",
			config: LoggingConfig{
				Level:  "DEBUG",
				Format: "json",
				File:   "",
				Rotation: LogRotationConfig{
					MaxSize:  100 * 1024 * 1024,
					MaxFiles: 5,
					MaxAge:   7 * 24 * time.Hour,
				},
			},
			hasError: false,
		},
		{
			name: "invalid log format",
			config: LoggingConfig{
				Level:  "info",
				Format: "invalid",
				File:   "",
				Rotation: LogRotationConfig{
					MaxSize:  100 * 1024 * 1024,
					MaxFiles: 5,
					MaxAge:   7 * 24 * time.Hour,
				},
			},
			hasError: true,
			field:    "logging.format",
		},
		{
			name: "log file with nonexistent directory",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				File:   "/nonexistent/dir/log.txt",
				Rotation: LogRotationConfig{
					MaxSize:  100 * 1024 * 1024,
					MaxFiles: 5,
					MaxAge:   7 * 24 * time.Hour,
				},
			},
			hasError: true,
			field:    "logging.file",
		},
		{
			name: "log file with valid directory",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				File:   filepath.Join(tempDir, "log.txt"),
				Rotation: LogRotationConfig{
					MaxSize:  100 * 1024 * 1024,
					MaxFiles: 5,
					MaxAge:   7 * 24 * time.Hour,
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateLogging(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateLogRotation(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   LogRotationConfig
		hasError bool
		field    string
	}{
		{
			name: "valid log rotation config",
			config: LogRotationConfig{
				MaxSize:  100 * 1024 * 1024,
				MaxFiles: 5,
				MaxAge:   7 * 24 * time.Hour,
			},
			hasError: false,
		},
		{
			name: "zero max size",
			config: LogRotationConfig{
				MaxSize:  0,
				MaxFiles: 5,
				MaxAge:   7 * 24 * time.Hour,
			},
			hasError: true,
			field:    "logging.rotation.max_size",
		},
		{
			name: "zero max files",
			config: LogRotationConfig{
				MaxSize:  100 * 1024 * 1024,
				MaxFiles: 0,
				MaxAge:   7 * 24 * time.Hour,
			},
			hasError: true,
			field:    "logging.rotation.max_files",
		},
		{
			name: "zero max age",
			config: LogRotationConfig{
				MaxSize:  100 * 1024 * 1024,
				MaxFiles: 5,
				MaxAge:   0,
			},
			hasError: true,
			field:    "logging.rotation.max_age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateLogRotation(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateREST(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   RESTConfig
		hasError bool
		field    string
	}{
		{
			name: "valid REST config",
			config: RESTConfig{
				Enabled:         true,
				Port:            8080,
				RateLimitPerMin: 100,
				AuthToken:       "valid-token",
			},
			hasError: false,
		},
		{
			name: "disabled REST config",
			config: RESTConfig{
				Enabled:         false,
				Port:            0,
				RateLimitPerMin: 0,
				AuthToken:       "",
			},
			hasError: false,
		},
		{
			name: "invalid port - too low",
			config: RESTConfig{
				Enabled:         true,
				Port:            0,
				RateLimitPerMin: 100,
				AuthToken:       "valid-token",
			},
			hasError: true,
			field:    "api.rest.port",
		},
		{
			name: "invalid port - too high",
			config: RESTConfig{
				Enabled:         true,
				Port:            65536,
				RateLimitPerMin: 100,
				AuthToken:       "valid-token",
			},
			hasError: true,
			field:    "api.rest.port",
		},
		{
			name: "zero rate limit",
			config: RESTConfig{
				Enabled:         true,
				Port:            8080,
				RateLimitPerMin: 0,
				AuthToken:       "valid-token",
			},
			hasError: true,
			field:    "api.rest.rate_limit_per_min",
		},
		{
			name: "empty auth token",
			config: RESTConfig{
				Enabled:         true,
				Port:            8080,
				RateLimitPerMin: 100,
				AuthToken:       "",
			},
			hasError: true,
			field:    "api.rest.auth_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateREST(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateGraphQL(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   GraphQLConfig
		hasError bool
		field    string
	}{
		{
			name: "valid GraphQL config",
			config: GraphQLConfig{
				Enabled:          true,
				Port:             8080,
				GraphQLPath:      "/graphql",
				EnablePlayground: false,
			},
			hasError: false,
		},
		{
			name: "valid GraphQL config with playground",
			config: GraphQLConfig{
				Enabled:          true,
				Port:             8080,
				GraphQLPath:      "/graphql",
				EnablePlayground: true,
				PlaygroundPath:   "/playground",
			},
			hasError: false,
		},
		{
			name: "disabled GraphQL config",
			config: GraphQLConfig{
				Enabled:          false,
				Port:             0,
				GraphQLPath:      "",
				EnablePlayground: false,
			},
			hasError: false,
		},
		{
			name: "invalid port",
			config: GraphQLConfig{
				Enabled:          true,
				Port:             0,
				GraphQLPath:      "/graphql",
				EnablePlayground: false,
			},
			hasError: true,
			field:    "api.graphql.port",
		},
		{
			name: "empty GraphQL path",
			config: GraphQLConfig{
				Enabled:          true,
				Port:             8080,
				GraphQLPath:      "",
				EnablePlayground: false,
			},
			hasError: true,
			field:    "api.graphql.graphql_path",
		},
		{
			name: "playground enabled but no path",
			config: GraphQLConfig{
				Enabled:          true,
				Port:             8080,
				GraphQLPath:      "/graphql",
				EnablePlayground: true,
				PlaygroundPath:   "",
			},
			hasError: true,
			field:    "api.graphql.playground_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateGraphQL(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidateGRPC(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   GRPCConfig
		hasError bool
		field    string
	}{
		{
			name: "valid gRPC config",
			config: GRPCConfig{
				Enabled:              true,
				Port:                 8080,
				AuthToken:            "valid-token",
				MaxConcurrentStreams: 100,
			},
			hasError: false,
		},
		{
			name: "disabled gRPC config",
			config: GRPCConfig{
				Enabled:              false,
				Port:                 0,
				AuthToken:            "",
				MaxConcurrentStreams: 0,
			},
			hasError: false,
		},
		{
			name: "invalid port",
			config: GRPCConfig{
				Enabled:              true,
				Port:                 0,
				AuthToken:            "valid-token",
				MaxConcurrentStreams: 100,
			},
			hasError: true,
			field:    "api.grpc.port",
		},
		{
			name: "empty auth token",
			config: GRPCConfig{
				Enabled:              true,
				Port:                 8080,
				AuthToken:            "",
				MaxConcurrentStreams: 100,
			},
			hasError: true,
			field:    "api.grpc.auth_token",
		},
		{
			name: "zero max concurrent streams",
			config: GRPCConfig{
				Enabled:              true,
				Port:                 8080,
				AuthToken:            "valid-token",
				MaxConcurrentStreams: 0,
			},
			hasError: true,
			field:    "api.grpc.max_concurrent_streams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateGRPC(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidatePeer(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   PeerConfig
		hasError bool
		field    string
	}{
		{
			name: "valid peer config",
			config: PeerConfig{
				MaxPeers:             100,
				HeartbeatInterval:    time.Second,
				HealthTimeout:        time.Second,
				ReconnectBackoff:     time.Second,
				MaxReconnectAttempts: 5,
			},
			hasError: false,
		},
		{
			name: "zero max peers",
			config: PeerConfig{
				MaxPeers:             0,
				HeartbeatInterval:    time.Second,
				HealthTimeout:        time.Second,
				ReconnectBackoff:     time.Second,
				MaxReconnectAttempts: 5,
			},
			hasError: true,
			field:    "peer.max_peers",
		},
		{
			name: "zero heartbeat interval",
			config: PeerConfig{
				MaxPeers:             100,
				HeartbeatInterval:    0,
				HealthTimeout:        time.Second,
				ReconnectBackoff:     time.Second,
				MaxReconnectAttempts: 5,
			},
			hasError: true,
			field:    "peer.heartbeat_interval",
		},
		{
			name: "zero health timeout",
			config: PeerConfig{
				MaxPeers:             100,
				HeartbeatInterval:    time.Second,
				HealthTimeout:        0,
				ReconnectBackoff:     time.Second,
				MaxReconnectAttempts: 5,
			},
			hasError: true,
			field:    "peer.health_timeout",
		},
		{
			name: "zero reconnect backoff",
			config: PeerConfig{
				MaxPeers:             100,
				HeartbeatInterval:    time.Second,
				HealthTimeout:        time.Second,
				ReconnectBackoff:     0,
				MaxReconnectAttempts: 5,
			},
			hasError: true,
			field:    "peer.reconnect_backoff",
		},
		{
			name: "zero max reconnect attempts",
			config: PeerConfig{
				MaxPeers:             100,
				HeartbeatInterval:    time.Second,
				HealthTimeout:        time.Second,
				ReconnectBackoff:     time.Second,
				MaxReconnectAttempts: 0,
			},
			hasError: true,
			field:    "peer.max_reconnect_attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePeer(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDefaultValidator_ValidatePerformance(t *testing.T) {
	validator := &DefaultValidator{}

	tests := []struct {
		name     string
		config   PerformanceConfig
		hasError bool
		field    string
	}{
		{
			name: "valid performance config",
			config: PerformanceConfig{
				MaxConcurrentStreamsPerPeer: 10,
				StreamBufferSize:            1024,
				ConnectionPoolSize:          100,
				CacheSize:                   1024 * 1024,
				CacheTTL:                    time.Hour,
			},
			hasError: false,
		},
		{
			name: "zero max concurrent streams per peer",
			config: PerformanceConfig{
				MaxConcurrentStreamsPerPeer: 0,
				StreamBufferSize:            1024,
				ConnectionPoolSize:          100,
				CacheSize:                   1024 * 1024,
				CacheTTL:                    time.Hour,
			},
			hasError: true,
			field:    "performance.max_concurrent_streams_per_peer",
		},
		{
			name: "zero stream buffer size",
			config: PerformanceConfig{
				MaxConcurrentStreamsPerPeer: 10,
				StreamBufferSize:            0,
				ConnectionPoolSize:          100,
				CacheSize:                   1024 * 1024,
				CacheTTL:                    time.Hour,
			},
			hasError: true,
			field:    "performance.stream_buffer_size",
		},
		{
			name: "zero connection pool size",
			config: PerformanceConfig{
				MaxConcurrentStreamsPerPeer: 10,
				StreamBufferSize:            1024,
				ConnectionPoolSize:          0,
				CacheSize:                   1024 * 1024,
				CacheTTL:                    time.Hour,
			},
			hasError: true,
			field:    "performance.connection_pool_size",
		},
		{
			name: "negative cache size",
			config: PerformanceConfig{
				MaxConcurrentStreamsPerPeer: 10,
				StreamBufferSize:            1024,
				ConnectionPoolSize:          100,
				CacheSize:                   -1,
				CacheTTL:                    time.Hour,
			},
			hasError: true,
			field:    "performance.cache_size",
		},
		{
			name: "negative cache TTL",
			config: PerformanceConfig{
				MaxConcurrentStreamsPerPeer: 10,
				StreamBufferSize:            1024,
				ConnectionPoolSize:          100,
				CacheSize:                   1024 * 1024,
				CacheTTL:                    -1 * time.Hour,
			},
			hasError: true,
			field:    "performance.cache_ttl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validatePerformance(tt.config)
			if tt.hasError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.field, err.Field)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestPortValidator_Validate(t *testing.T) {
	validator := &PortValidator{}

	tests := []struct {
		name     string
		config   *Config
		hasError bool
		contains string
	}{
		{
			name: "no port conflicts",
			config: &Config{
				API: APIConfig{
					REST: RESTConfig{
						Enabled: true,
						Port:    8080,
					},
					GraphQL: GraphQLConfig{
						Enabled: true,
						Port:    8081,
					},
					GRPC: GRPCConfig{
						Enabled: true,
						Port:    8082,
					},
				},
			},
			hasError: false,
		},
		{
			name: "REST and GraphQL port conflict",
			config: &Config{
				API: APIConfig{
					REST: RESTConfig{
						Enabled: true,
						Port:    8080,
					},
					GraphQL: GraphQLConfig{
						Enabled: true,
						Port:    8080,
					},
					GRPC: GRPCConfig{
						Enabled: true,
						Port:    8082,
					},
				},
			},
			hasError: true,
			contains: "port conflict",
		},
		{
			name: "disabled APIs don't conflict",
			config: &Config{
				API: APIConfig{
					REST: RESTConfig{
						Enabled: false,
						Port:    8080,
					},
					GraphQL: GraphQLConfig{
						Enabled: false,
						Port:    8080,
					},
					GRPC: GRPCConfig{
						Enabled: false,
						Port:    8080,
					},
				},
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.contains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		allowDemo   bool
		hasError    bool
		hasWarnings bool
		contains    string
	}{
		{
			name: "valid security config",
			config: &Config{
				Security: SecurityConfig{
					AuthToken:  "secure-token",
					ClusterKey: "a-very-long-cluster-key-that-is-secure-enough",
				},
			},
			allowDemo:   false,
			hasError:    false,
			hasWarnings: false,
		},
		{
			name: "demo token allowed",
			config: &Config{
				Security: SecurityConfig{
					AuthToken:  "demo-token",
					ClusterKey: "a-very-long-cluster-key-that-is-secure-enough",
				},
			},
			allowDemo:   true,
			hasError:    false,
			hasWarnings: false,
		},
		{
			name: "demo token not allowed",
			config: &Config{
				Security: SecurityConfig{
					AuthToken:  "demo-token",
					ClusterKey: "a-very-long-cluster-key-that-is-secure-enough",
				},
			},
			allowDemo:   false,
			hasError:    false,
			hasWarnings: true,
			contains:    "demo token",
		},
		{
			name: "empty cluster key",
			config: &Config{
				Security: SecurityConfig{
					AuthToken:  "secure-token",
					ClusterKey: "",
				},
			},
			allowDemo:   false,
			hasError:    false,
			hasWarnings: true,
			contains:    "cluster key",
		},
		{
			name: "short cluster key",
			config: &Config{
				Security: SecurityConfig{
					AuthToken:  "secure-token",
					ClusterKey: "short",
				},
			},
			allowDemo:   false,
			hasError:    false,
			hasWarnings: true,
			contains:    "32 characters",
		},
		{
			name: "config allows demo token",
			config: &Config{
				Security: SecurityConfig{
					AuthToken:      "demo-token",
					ClusterKey:     "a-very-long-cluster-key-that-is-secure-enough",
					AllowDemoToken: true,
				},
			},
			allowDemo:   false, // Validator doesn't allow, but config does
			hasError:    false,
			hasWarnings: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSecurityValidator(tt.allowDemo)
			err := validator.Validate(tt.config)

			if tt.hasError {
				assert.Error(t, err)
			} else if tt.hasWarnings {
				assert.Error(t, err) // Warnings are returned as errors
				if result, ok := err.(*ValidationResult); ok {
					assert.True(t, result.HasWarnings())
					assert.Contains(t, result.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStorageValidator_Validate(t *testing.T) {
	tempDir := t.TempDir()
	validator := &StorageValidator{}

	tests := []struct {
		name        string
		config      *Config
		hasError    bool
		hasWarnings bool
		contains    string
	}{
		{
			name: "valid storage config",
			config: &Config{
				Storage: StorageConfig{
					Root:        tempDir,
					MaxFileSize: 1024 * 1024,
				},
			},
			hasError:    false,
			hasWarnings: false,
		},
		{
			name: "large file size warning",
			config: &Config{
				Storage: StorageConfig{
					Root:        tempDir,
					MaxFileSize: 11 * 1024 * 1024 * 1024, // 11GB
				},
			},
			hasError:    false,
			hasWarnings: true,
			contains:    "very large",
		},
		{
			name: "unwritable storage directory",
			config: &Config{
				Storage: StorageConfig{
					Root:        "/root/restricted",
					MaxFileSize: 1024 * 1024,
				},
			},
			hasError:    runtime.GOOS != "windows", // Skip on Windows
			hasWarnings: false,
			contains:    "storage validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)

			if tt.hasError {
				assert.Error(t, err)
				if result, ok := err.(*ValidationResult); ok {
					assert.True(t, result.HasErrors())
					assert.Contains(t, result.Error(), tt.contains)
				}
			} else if tt.hasWarnings {
				assert.Error(t, err) // Warnings are returned as errors
				if result, ok := err.(*ValidationResult); ok {
					assert.True(t, result.HasWarnings())
					assert.Contains(t, result.Error(), tt.contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStorageValidator_CheckStorageWritable(t *testing.T) {
	validator := &StorageValidator{}
	tempDir := t.TempDir()

	// Test writable directory
	err := validator.checkStorageWritable(tempDir)
	assert.NoError(t, err)

	// Test creating directory
	newDir := filepath.Join(tempDir, "newdir")
	err = validator.checkStorageWritable(newDir)
	assert.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(newDir)
	assert.NoError(t, err)

	// Test unwritable directory (this might fail on some systems)
	// Skip this test on Windows as it might not have the same permission model
	if runtime.GOOS != "windows" {
		err = validator.checkStorageWritable("/root/restricted")
		assert.Error(t, err)
		// The error could be either from directory creation or write test
		assert.True(t, strings.Contains(err.Error(), "storage directory is not writable") ||
			strings.Contains(err.Error(), "failed to create storage directory") ||
			strings.Contains(err.Error(), "permission denied"))
	}
}
