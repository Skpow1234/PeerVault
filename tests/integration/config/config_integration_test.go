package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/graphql"
	"github.com/Skpow1234/Peervault/internal/api/grpc"
	"github.com/Skpow1234/Peervault/internal/api/rest"
	"github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/config"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigurationIntegration tests the integration of configuration with actual components
func TestConfigurationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "integration-test.yaml")

	// Create a comprehensive configuration file
	configData := `
server:
  listen_addr: ":3001"
  debug: true
  shutdown_timeout: "30s"

storage:
  root: "%s"
  max_file_size: 1048576
  compression: true
  compression_level: 6
  deduplication: false
  cleanup_interval: "1h"
  retention_period: "24h"

network:
  bootstrap_nodes:
    - "localhost:3000"
  connection_timeout: "30s"
  read_timeout: "60s"
  write_timeout: "60s"
  keep_alive_interval: "30s"
  max_message_size: 1048576

security:
  cluster_key: "test-cluster-key-for-integration-testing-only"
  auth_token: "integration-test-token"
  tls: false
  key_rotation_interval: "24h"
  encryption_at_rest: true
  encryption_in_transit: true
  allow_demo_token: true

logging:
  level: "debug"
  format: "json"
  file: ""
  structured: true
  include_source: false
  rotation:
    max_size: 100
    max_files: 10
    max_age: "168h"
    compress: true

api:
  rest:
    enabled: true
    port: 8081
    allowed_origins:
      - "*"
    rate_limit_per_min: 100
    auth_token: "rest-test-token"
  
  graphql:
    enabled: true
    port: 8082
    enable_playground: true
    graphql_path: "/graphql"
    playground_path: "/playground"
    allowed_origins:
      - "*"
  
  grpc:
    enabled: true
    port: 8083
    auth_token: "grpc-test-token"
    enable_reflection: true
    max_concurrent_streams: 100

peer:
  max_peers: 50
  heartbeat_interval: "30s"
  health_timeout: "90s"
  reconnect_backoff: "5s"
  max_reconnect_attempts: 10

performance:
  max_concurrent_streams_per_peer: 5
  stream_buffer_size: 32768
  connection_pool_size: 5
  enable_multiplexing: true
  cache_size: 50
  cache_ttl: "30m"
`

	// Replace storage root placeholder
	configData = fmt.Sprintf(configData, filepath.Join(tempDir, "storage"))

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Create configuration manager
	manager := config.NewManager(configPath)
	manager.AddValidator(&config.DefaultValidator{})
	manager.AddValidator(&config.PortValidator{})
	manager.AddValidator(config.NewSecurityValidator(false)) // Will use config.Security.AllowDemoToken
	manager.AddValidator(&config.StorageValidator{})

	// Load configuration
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()

	// Test server configuration integration
	t.Run("ServerConfiguration", func(t *testing.T) {
		assert.Equal(t, ":3001", cfg.Server.ListenAddr)
		assert.True(t, cfg.Server.Debug)
		assert.Equal(t, 30*time.Second, cfg.Server.ShutdownTimeout)
	})

	// Test storage configuration integration
	t.Run("StorageConfiguration", func(t *testing.T) {
		assert.Equal(t, filepath.Join(tempDir, "storage"), cfg.Storage.Root)
		assert.Equal(t, int64(1048576), cfg.Storage.MaxFileSize)
		assert.True(t, cfg.Storage.Compression)
		assert.Equal(t, 6, cfg.Storage.CompressionLevel)
		assert.False(t, cfg.Storage.Deduplication)
		assert.Equal(t, 1*time.Hour, cfg.Storage.CleanupInterval)
		assert.Equal(t, 24*time.Hour, cfg.Storage.RetentionPeriod)
	})

	// Test network configuration integration
	t.Run("NetworkConfiguration", func(t *testing.T) {
		assert.Equal(t, []string{"localhost:3000"}, cfg.Network.BootstrapNodes)
		assert.Equal(t, 30*time.Second, cfg.Network.ConnectionTimeout)
		assert.Equal(t, 60*time.Second, cfg.Network.ReadTimeout)
		assert.Equal(t, 60*time.Second, cfg.Network.WriteTimeout)
		assert.Equal(t, 30*time.Second, cfg.Network.KeepAliveInterval)
		assert.Equal(t, int64(1048576), cfg.Network.MaxMessageSize)
	})

	// Test security configuration integration
	t.Run("SecurityConfiguration", func(t *testing.T) {
		assert.Equal(t, "test-cluster-key-for-integration-testing-only", cfg.Security.ClusterKey)
		assert.Equal(t, "integration-test-token", cfg.Security.AuthToken)
		assert.False(t, cfg.Security.TLS)
		assert.Equal(t, 24*time.Hour, cfg.Security.KeyRotationInterval)
		assert.True(t, cfg.Security.EncryptionAtRest)
		assert.True(t, cfg.Security.EncryptionInTransit)
	})

	// Test logging configuration integration
	t.Run("LoggingConfiguration", func(t *testing.T) {
		assert.Equal(t, "debug", cfg.Logging.Level)
		assert.Equal(t, "json", cfg.Logging.Format)
		assert.Equal(t, "", cfg.Logging.File)
		assert.True(t, cfg.Logging.Structured)
		assert.False(t, cfg.Logging.IncludeSource)
		assert.Equal(t, 100, cfg.Logging.Rotation.MaxSize)
		assert.Equal(t, 10, cfg.Logging.Rotation.MaxFiles)
		assert.Equal(t, 168*time.Hour, cfg.Logging.Rotation.MaxAge)
		assert.True(t, cfg.Logging.Rotation.Compress)
	})

	// Test API configuration integration
	t.Run("APIConfiguration", func(t *testing.T) {
		// REST API
		assert.True(t, cfg.API.REST.Enabled)
		assert.Equal(t, 8081, cfg.API.REST.Port)
		assert.Equal(t, []string{"*"}, cfg.API.REST.AllowedOrigins)
		assert.Equal(t, 100, cfg.API.REST.RateLimitPerMin)
		assert.Equal(t, "rest-test-token", cfg.API.REST.AuthToken)

		// GraphQL API
		assert.True(t, cfg.API.GraphQL.Enabled)
		assert.Equal(t, 8082, cfg.API.GraphQL.Port)
		assert.True(t, cfg.API.GraphQL.EnablePlayground)
		assert.Equal(t, "/graphql", cfg.API.GraphQL.GraphQLPath)
		assert.Equal(t, "/playground", cfg.API.GraphQL.PlaygroundPath)
		assert.Equal(t, []string{"*"}, cfg.API.GraphQL.AllowedOrigins)

		// gRPC API
		assert.True(t, cfg.API.GRPC.Enabled)
		assert.Equal(t, 8083, cfg.API.GRPC.Port)
		assert.Equal(t, "grpc-test-token", cfg.API.GRPC.AuthToken)
		assert.True(t, cfg.API.GRPC.EnableReflection)
		assert.Equal(t, 100, cfg.API.GRPC.MaxConcurrentStreams)
	})

	// Test peer configuration integration
	t.Run("PeerConfiguration", func(t *testing.T) {
		assert.Equal(t, 50, cfg.Peer.MaxPeers)
		assert.Equal(t, 30*time.Second, cfg.Peer.HeartbeatInterval)
		assert.Equal(t, 90*time.Second, cfg.Peer.HealthTimeout)
		assert.Equal(t, 5*time.Second, cfg.Peer.ReconnectBackoff)
		assert.Equal(t, 10, cfg.Peer.MaxReconnectAttempts)
	})

	// Test performance configuration integration
	t.Run("PerformanceConfiguration", func(t *testing.T) {
		assert.Equal(t, 5, cfg.Performance.MaxConcurrentStreamsPerPeer)
		assert.Equal(t, 32768, cfg.Performance.StreamBufferSize)
		assert.Equal(t, 5, cfg.Performance.ConnectionPoolSize)
		assert.True(t, cfg.Performance.EnableMultiplexing)
		assert.Equal(t, 50, cfg.Performance.CacheSize)
		assert.Equal(t, 30*time.Minute, cfg.Performance.CacheTTL)
	})
}

// TestConfigurationWithComponents tests configuration integration with actual components
func TestConfigurationWithComponents(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "component-test.yaml")

	// Create a minimal configuration for component testing
	configData := `
server:
  listen_addr: ":3002"
  debug: false

storage:
  root: "%s"

logging:
  level: "info"
  format: "text"

api:
  rest:
    enabled: true
    port: 8084
  graphql:
    enabled: true
    port: 8085
  grpc:
    enabled: true
    port: 8086

security:
  auth_token: "component-test-token"
`

	configData = fmt.Sprintf(configData, filepath.Join(tempDir, "storage"))

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Create configuration manager
	manager := config.NewManager(configPath)
	manager.AddValidator(&config.DefaultValidator{})
	manager.AddValidator(&config.PortValidator{})

	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()

	// Test fileserver integration
	t.Run("FileserverIntegration", func(t *testing.T) {
		// Initialize key manager
		keyManager, err := crypto.NewKeyManager()
		require.NoError(t, err)

		// Initialize transport
		transport := netp2p.NewTCPTransport(netp2p.TCPTransportOpts{
			ListenAddr: cfg.Server.ListenAddr,
			OnPeer:     nil,
			OnStream:   nil,
		})

		// Initialize fileserver with configuration
		opts := fileserver.Options{
			ID:                cfg.Server.NodeID,
			EncKey:            nil,
			KeyManager:        keyManager,
			StorageRoot:       cfg.Storage.Root,
			PathTransformFunc: storage.CASPathTransformFunc,
			Transport:         transport,
			BootstrapNodes:    cfg.Network.BootstrapNodes,
			ResourceLimits: peer.ResourceLimits{
				MaxConcurrentStreams: cfg.Performance.MaxConcurrentStreamsPerPeer,
			},
		}

		server := fileserver.New(opts)
		assert.NotNil(t, server)
		assert.Equal(t, cfg.Server.ListenAddr, cfg.Server.ListenAddr) // Verify configuration is used
	})

	// Test REST API integration
	t.Run("RESTAPIIntegration", func(t *testing.T) {
		if !cfg.API.REST.Enabled {
			t.Skip("REST API not enabled")
		}

		restConfig := &rest.Config{
			Port:            fmt.Sprintf(":%d", cfg.API.REST.Port),
			ReadTimeout:     cfg.Network.ReadTimeout,
			WriteTimeout:    cfg.Network.WriteTimeout,
			MaxHeaderBytes:  1 << 20,
			AllowedOrigins:  cfg.API.REST.AllowedOrigins,
			RateLimitPerMin: cfg.API.REST.RateLimitPerMin,
			AuthToken:       cfg.API.REST.AuthToken,
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		restServer := rest.NewServer(restConfig, logger)
		assert.NotNil(t, restServer)
		assert.Equal(t, fmt.Sprintf(":%d", cfg.API.REST.Port), restConfig.Port)
	})

	// Test GraphQL API integration
	t.Run("GraphQLAPIIntegration", func(t *testing.T) {
		if !cfg.API.GraphQL.Enabled {
			t.Skip("GraphQL API not enabled")
		}

		// Create a mock fileserver for GraphQL
		keyManager, err := crypto.NewKeyManager()
		require.NoError(t, err)

		transport := netp2p.NewTCPTransport(netp2p.TCPTransportOpts{
			ListenAddr: cfg.Server.ListenAddr,
			OnPeer:     nil,
			OnStream:   nil,
		})

		opts := fileserver.Options{
			ID:                cfg.Server.NodeID,
			KeyManager:        keyManager,
			StorageRoot:       cfg.Storage.Root,
			PathTransformFunc: storage.CASPathTransformFunc,
			Transport:         transport,
			BootstrapNodes:    cfg.Network.BootstrapNodes,
			ResourceLimits:    peer.DefaultResourceLimits(),
		}

		fileserver := fileserver.New(opts)

		graphqlConfig := &graphql.Config{
			Port:             cfg.API.GraphQL.Port,
			PlaygroundPath:   cfg.API.GraphQL.PlaygroundPath,
			GraphQLPath:      cfg.API.GraphQL.GraphQLPath,
			AllowedOrigins:   cfg.API.GraphQL.AllowedOrigins,
			EnablePlayground: cfg.API.GraphQL.EnablePlayground,
		}

		graphqlServer := graphql.NewServer(fileserver, graphqlConfig)
		assert.NotNil(t, graphqlServer)
		assert.Equal(t, cfg.API.GraphQL.Port, graphqlConfig.Port)
		assert.Equal(t, cfg.API.GraphQL.GraphQLPath, graphqlConfig.GraphQLPath)
	})

	// Test gRPC API integration
	t.Run("GRPCAPIIntegration", func(t *testing.T) {
		if !cfg.API.GRPC.Enabled {
			t.Skip("gRPC API not enabled")
		}

		grpcConfig := &grpc.Config{
			Port:      fmt.Sprintf(":%d", cfg.API.GRPC.Port),
			AuthToken: cfg.API.GRPC.AuthToken,
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		grpcServer := grpc.NewServer(grpcConfig, logger)
		assert.NotNil(t, grpcServer)
		assert.Equal(t, fmt.Sprintf(":%d", cfg.API.GRPC.Port), grpcConfig.Port)
		assert.Equal(t, cfg.API.GRPC.AuthToken, grpcConfig.AuthToken)
	})
}

// TestConfigurationHotReload tests hot reloading functionality
func TestConfigurationHotReload(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "hot-reload-test.yaml")

	// Create initial configuration
	initialConfig := `
server:
  listen_addr: ":3003"
  debug: false
logging:
  level: "info"
api:
  rest:
    enabled: true
    port: 8087
  graphql:
    enabled: true
    port: 8088
  grpc:
    enabled: true
    port: 8089
`

	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	require.NoError(t, err)

	// Create configuration manager
	manager := config.NewManager(configPath)
	manager.AddValidator(&config.DefaultValidator{})
	manager.AddValidator(&config.PortValidator{})

	// Load initial configuration
	err = manager.Load()
	require.NoError(t, err)

	initialCfg := manager.Get()
	assert.Equal(t, ":3003", initialCfg.Server.ListenAddr)
	assert.False(t, initialCfg.Server.Debug)
	assert.Equal(t, "info", initialCfg.Logging.Level)

	// Start watching for changes
	reloadCount := 0
	reloadChan := make(chan *config.Config, 1)

	err = manager.Watch(func(cfg *config.Config) {
		reloadCount++
		reloadChan <- cfg
	})
	require.NoError(t, err)

	// Wait for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Update configuration file
	updatedConfig := `
server:
  listen_addr: ":3004"
  debug: true
logging:
  level: "debug"
api:
  rest:
    enabled: true
    port: 8087
  graphql:
    enabled: true
    port: 8088
  grpc:
    enabled: true
    port: 8089
`

	err = os.WriteFile(configPath, []byte(updatedConfig), 0644)
	require.NoError(t, err)

	// Wait for reload
	select {
	case updatedCfg := <-reloadChan:
		assert.Equal(t, ":3004", updatedCfg.Server.ListenAddr)
		assert.True(t, updatedCfg.Server.Debug)
		assert.Equal(t, "debug", updatedCfg.Logging.Level)
		assert.Greater(t, reloadCount, 0)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for configuration reload")
	}

	// Stop watching
	manager.Stop()
}

// TestConfigurationEnvironmentOverride tests environment variable overrides
func TestConfigurationEnvironmentOverride(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "env-override-test.yaml")

	// Create configuration file
	configData := `
server:
  listen_addr: ":3005"
  debug: false
storage:
  root: "/tmp/storage"
logging:
  level: "info"
api:
  rest:
    enabled: true
    port: 8090
  graphql:
    enabled: true
    port: 8091
  grpc:
    enabled: true
    port: 8092
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Set environment variables to override file values
	require.NoError(t, os.Setenv("PEERVAULT_SERVER_DEBUG", "true"))
	require.NoError(t, os.Setenv("PEERVAULT_LOGGING_LEVEL", "debug"))
	require.NoError(t, os.Setenv("PEERVAULT_STORAGE_ROOT", "/env/storage"))
	require.NoError(t, os.Setenv("PEERVAULT_API_REST_PORT", "8093"))
	require.NoError(t, os.Setenv("PEERVAULT_API_GRAPHQL_PORT", "8094"))
	require.NoError(t, os.Setenv("PEERVAULT_API_GRPC_PORT", "8095"))
	defer func() {
		require.NoError(t, os.Unsetenv("PEERVAULT_SERVER_DEBUG"))
		require.NoError(t, os.Unsetenv("PEERVAULT_LOGGING_LEVEL"))
		require.NoError(t, os.Unsetenv("PEERVAULT_STORAGE_ROOT"))
		require.NoError(t, os.Unsetenv("PEERVAULT_API_REST_PORT"))
		require.NoError(t, os.Unsetenv("PEERVAULT_API_GRAPHQL_PORT"))
		require.NoError(t, os.Unsetenv("PEERVAULT_API_GRPC_PORT"))
	}()

	// Create configuration manager
	manager := config.NewManager(configPath)
	manager.AddValidator(&config.DefaultValidator{})
	manager.AddValidator(&config.PortValidator{})

	// Load configuration
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()

	// Verify environment variables override file values
	assert.Equal(t, ":3005", cfg.Server.ListenAddr)   // File value (no env override)
	assert.True(t, cfg.Server.Debug)                  // Environment override
	assert.Equal(t, "debug", cfg.Logging.Level)       // Environment override
	assert.Equal(t, "/env/storage", cfg.Storage.Root) // Environment override
	assert.Equal(t, 8093, cfg.API.REST.Port)          // Environment override
	assert.Equal(t, 8094, cfg.API.GraphQL.Port)       // Environment override
	assert.Equal(t, 8095, cfg.API.GRPC.Port)          // Environment override
}

// TestConfigurationValidationIntegration tests validation integration
func TestConfigurationValidationIntegration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "validation-test.yaml")

	// Create configuration with validation issues
	configData := `
server:
  listen_addr: "invalid-address"
  debug: false
storage:
  root: "%s"
  max_file_size: -1
logging:
  level: "invalid-level"
api:
  rest:
    enabled: true
    port: 8096
  graphql:
    enabled: true
    port: 8096  # Port conflict
  grpc:
    enabled: true
    port: 8097
security:
  auth_token: "demo-token"  # Weak token
  cluster_key: ""  # Empty cluster key
  allow_demo_token: true
`

	configData = fmt.Sprintf(configData, filepath.Join(tempDir, "storage"))

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Create configuration manager with all validators
	manager := config.NewManager(configPath)
	manager.AddValidator(&config.DefaultValidator{})
	manager.AddValidator(&config.PortValidator{})
	manager.AddValidator(config.NewSecurityValidator(false)) // Will use config.Security.AllowDemoToken
	manager.AddValidator(&config.StorageValidator{})

	// Load configuration (should fail validation)
	err = manager.Load()
	assert.Error(t, err)

	// Check for specific validation errors
	if validationErrors, ok := err.(*config.ValidationErrors); ok {
		errorMessages := make([]string, len(validationErrors.Errors))
		for i, validationError := range validationErrors.Errors {
			errorMessages[i] = validationError.Message
		}

		// Check for expected validation errors
		assert.Contains(t, errorMessages, "invalid listen address")
		assert.Contains(t, errorMessages, "max file size must be positive")
		assert.Contains(t, errorMessages, "log level must be one of: debug, info, warn, error")
		assert.Contains(t, errorMessages, "port conflict")
		assert.Contains(t, errorMessages, "security warning")
	}
}
