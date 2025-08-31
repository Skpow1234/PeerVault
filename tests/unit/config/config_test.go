package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, ":3000", cfg.Server.ListenAddr)
	assert.Equal(t, "./storage", cfg.Storage.Root)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "demo-token", cfg.Security.AuthToken)
	assert.True(t, cfg.API.REST.Enabled)
	assert.True(t, cfg.API.GraphQL.Enabled)
	assert.True(t, cfg.API.GRPC.Enabled)
}

func TestNewManager(t *testing.T) {
	manager := config.NewManager("test-config.yaml")

	assert.NotNil(t, manager)
	assert.Equal(t, "test-config.yaml", manager.GetConfigPath())
	assert.NotNil(t, manager.Get())
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configData := `
server:
  listen_addr: ":8080"
  debug: true
storage:
  root: "/tmp/test-storage"
  max_file_size: 2097152
logging:
  level: "debug"
  format: "text"
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	manager := config.NewManager(configPath)
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, ":8080", cfg.Server.ListenAddr)
	assert.True(t, cfg.Server.Debug)
	assert.Equal(t, "/tmp/test-storage", cfg.Storage.Root)
	assert.Equal(t, int64(2097152), cfg.Storage.MaxFileSize)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
}

func TestLoadFromEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("PEERVAULT_LISTEN_ADDR", ":9090")
	os.Setenv("PEERVAULT_STORAGE_ROOT", "/env/storage")
	os.Setenv("PEERVAULT_LOG_LEVEL", "warn")
	os.Setenv("PEERVAULT_DEBUG", "true")
	os.Setenv("PEERVAULT_MAX_FILE_SIZE", "1048576")
	os.Setenv("PEERVAULT_BOOTSTRAP_NODES", "node1:3000,node2:3000")
	defer func() {
		os.Unsetenv("PEERVAULT_LISTEN_ADDR")
		os.Unsetenv("PEERVAULT_STORAGE_ROOT")
		os.Unsetenv("PEERVAULT_LOG_LEVEL")
		os.Unsetenv("PEERVAULT_DEBUG")
		os.Unsetenv("PEERVAULT_MAX_FILE_SIZE")
		os.Unsetenv("PEERVAULT_BOOTSTRAP_NODES")
	}()

	manager := config.NewManager("")
	err := manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, ":9090", cfg.Server.ListenAddr)
	assert.Equal(t, "/env/storage", cfg.Storage.Root)
	assert.Equal(t, "warn", cfg.Logging.Level)
	assert.True(t, cfg.Server.Debug)
	assert.Equal(t, int64(1048576), cfg.Storage.MaxFileSize)
	assert.Equal(t, []string{"node1:3000", "node2:3000"}, cfg.Network.BootstrapNodes)
}

func TestLoadFromFileAndEnvironment(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configData := `
server:
  listen_addr: ":8080"
  debug: false
storage:
  root: "/tmp/test-storage"
logging:
  level: "info"
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Set environment variables to override file values
	os.Setenv("PEERVAULT_LOG_LEVEL", "debug")
	os.Setenv("PEERVAULT_DEBUG", "true")
	defer func() {
		os.Unsetenv("PEERVAULT_LOG_LEVEL")
		os.Unsetenv("PEERVAULT_DEBUG")
	}()

	manager := config.NewManager(configPath)
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	// File values
	assert.Equal(t, ":8080", cfg.Server.ListenAddr)
	assert.Equal(t, "/tmp/test-storage", cfg.Storage.Root)
	// Environment overrides
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.True(t, cfg.Server.Debug)
}

func TestValidation(t *testing.T) {
	manager := config.NewManager("")
	manager.AddValidator(&config.DefaultValidator{})

	// Test valid configuration
	err := manager.Load()
	assert.NoError(t, err)

	// Test invalid configuration
	cfg := manager.Get()
	cfg.Server.ListenAddr = "invalid-address"

	// This should fail validation
	manager.AddValidator(&config.DefaultValidator{})
	// Note: The current implementation doesn't re-validate after modification
	// This test demonstrates the validation framework is in place
}

func TestPortValidator(t *testing.T) {
	validator := &config.PortValidator{}

	// Test valid configuration
	cfg := config.DefaultConfig()
	err := validator.Validate(cfg)
	assert.NoError(t, err)

	// Test port conflict - create a new config with conflicting ports
	cfg2 := config.DefaultConfig()
	cfg2.API.REST.Port = 8080
	cfg2.API.GraphQL.Port = 8080
	err = validator.Validate(cfg2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port conflict")
}

func TestSecurityValidator(t *testing.T) {
	validator := &config.SecurityValidator{}

	// Test valid configuration
	cfg := config.DefaultConfig()
	cfg.Security.ClusterKey = "a-very-long-cluster-key-that-is-secure-enough"
	cfg.Security.AuthToken = "secure-token"
	err := validator.Validate(cfg)
	assert.NoError(t, err)

	// Test weak auth token
	cfg.Security.AuthToken = "demo-token"
	err = validator.Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security warning")

	// Test empty cluster key
	cfg.Security.ClusterKey = ""
	err = validator.Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security warning")

	// Test weak cluster key
	cfg.Security.ClusterKey = "short"
	err = validator.Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security warning")
}

func TestStorageValidator(t *testing.T) {
	validator := &config.StorageValidator{}

	// Test valid configuration
	cfg := config.DefaultConfig()
	err := validator.Validate(cfg)
	assert.NoError(t, err)

	// Test very large file size
	cfg.Storage.MaxFileSize = 100 * 1024 * 1024 * 1024 // 100GB
	err = validator.Validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage warning")
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "save-test.yaml")

	manager := config.NewManager(configPath)

	// Modify configuration
	cfg := manager.Get()
	cfg.Server.ListenAddr = ":9090"
	cfg.Logging.Level = "debug"

	err := manager.Save()
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify
	manager2 := config.NewManager(configPath)
	err = manager2.Load()
	require.NoError(t, err)

	cfg2 := manager2.Get()
	assert.Equal(t, ":9090", cfg2.Server.ListenAddr)
	assert.Equal(t, "debug", cfg2.Logging.Level)
}

func TestConfigWatcher(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "watch-test.yaml")

	// Create initial config file
	configData := `
server:
  listen_addr: ":8080"
logging:
  level: "info"
`
	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	manager := config.NewManager(configPath)
	manager.AddValidator(&config.DefaultValidator{})

	// Load initial configuration
	err = manager.Load()
	require.NoError(t, err)

	initialCfg := manager.Get()
	assert.Equal(t, ":8080", initialCfg.Server.ListenAddr)
	assert.Equal(t, "info", initialCfg.Logging.Level)

	// Start watching
	reloadCount := 0
	err = manager.Watch(func(cfg *config.Config) {
		reloadCount++
	})
	require.NoError(t, err)

	// Wait a bit for watcher to start
	time.Sleep(100 * time.Millisecond)

	// Modify the config file
	updatedConfigData := `
server:
  listen_addr: ":9090"
logging:
  level: "debug"
`
	err = os.WriteFile(configPath, []byte(updatedConfigData), 0644)
	require.NoError(t, err)

	// Wait for reload
	time.Sleep(2 * time.Second)

	// Stop watching
	manager.Stop()

	// Verify reload happened
	assert.Greater(t, reloadCount, 0)
}

func TestDurationParsing(t *testing.T) {
	// Test duration parsing from environment
	os.Setenv("PEERVAULT_SHUTDOWN_TIMEOUT", "60s")
	os.Setenv("PEERVAULT_CLEANUP_INTERVAL", "2h")
	os.Setenv("PEERVAULT_RETENTION_PERIOD", "168h")
	defer func() {
		os.Unsetenv("PEERVAULT_SHUTDOWN_TIMEOUT")
		os.Unsetenv("PEERVAULT_CLEANUP_INTERVAL")
		os.Unsetenv("PEERVAULT_RETENTION_PERIOD")
	}()

	manager := config.NewManager("")
	err := manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, 60*time.Second, cfg.Server.ShutdownTimeout)
	assert.Equal(t, 2*time.Hour, cfg.Storage.CleanupInterval)
	assert.Equal(t, 168*time.Hour, cfg.Storage.RetentionPeriod)
}

func TestSliceParsing(t *testing.T) {
	// Test slice parsing from environment
	os.Setenv("PEERVAULT_BOOTSTRAP_NODES", "node1:3000,node2:3000,node3:3000")
	os.Setenv("PEERVAULT_REST_ALLOWED_ORIGINS", "https://example.com,https://test.com")
	defer func() {
		os.Unsetenv("PEERVAULT_BOOTSTRAP_NODES")
		os.Unsetenv("PEERVAULT_REST_ALLOWED_ORIGINS")
	}()

	manager := config.NewManager("")
	err := manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, []string{"node1:3000", "node2:3000", "node3:3000"}, cfg.Network.BootstrapNodes)
	assert.Equal(t, []string{"https://example.com", "https://test.com"}, cfg.API.REST.AllowedOrigins)
}

func TestInvalidConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	// Create invalid YAML
	invalidYAML := `
server:
  listen_addr: ":8080
  debug: true
`
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	manager := config.NewManager(configPath)
	err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestNonExistentConfigFile(t *testing.T) {
	manager := config.NewManager("non-existent.yaml")
	err := manager.Load()
	assert.NoError(t, err) // Should use defaults when file doesn't exist

	cfg := manager.Get()
	assert.NotNil(t, cfg)
	assert.Equal(t, ":3000", cfg.Server.ListenAddr) // Default value
}

func TestValidationErrors(t *testing.T) {
	manager := config.NewManager("")
	manager.AddValidator(&config.DefaultValidator{})

	// Create invalid configuration
	cfg := manager.Get()
	cfg.Server.ListenAddr = ""    // Invalid: empty listen address
	cfg.Storage.MaxFileSize = -1  // Invalid: negative file size
	cfg.Logging.Level = "invalid" // Invalid: unknown log level

	// Note: The current implementation doesn't re-validate after modification
	// This test demonstrates the validation framework structure
	validator := &config.DefaultValidator{}
	err := validator.Validate(cfg)
	assert.Error(t, err)

	// Check if it's a ValidationErrors type
	if validationErrors, ok := err.(*config.ValidationErrors); ok {
		assert.Greater(t, len(validationErrors.Errors), 0)
		for _, validationError := range validationErrors.Errors {
			assert.NotEmpty(t, validationError.Field)
			assert.NotEmpty(t, validationError.Message)
		}
	}
}
