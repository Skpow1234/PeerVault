package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test server config defaults
	assert.Equal(t, ":3000", cfg.Server.ListenAddr)
	assert.Equal(t, false, cfg.Server.Debug)
	assert.Equal(t, 30*time.Second, cfg.Server.ShutdownTimeout)

	// Test storage config defaults
	assert.Equal(t, "./storage", cfg.Storage.Root)
	assert.Equal(t, int64(1073741824), cfg.Storage.MaxFileSize) // 1GB
	assert.Equal(t, false, cfg.Storage.Compression)
	assert.Equal(t, 6, cfg.Storage.CompressionLevel)

	// Test security config defaults
	assert.Equal(t, "demo-token", cfg.Security.AuthToken)
	assert.Equal(t, true, cfg.Security.EncryptionAtRest)
	assert.Equal(t, true, cfg.Security.AllowDemoToken)

	// Test API config defaults
	assert.Equal(t, true, cfg.API.REST.Enabled)
	assert.Equal(t, 8080, cfg.API.REST.Port)
	assert.Equal(t, true, cfg.API.GraphQL.Enabled)
	assert.Equal(t, 8081, cfg.API.GraphQL.Port)
	assert.Equal(t, true, cfg.API.GRPC.Enabled)
	assert.Equal(t, 8082, cfg.API.GRPC.Port)
}

func TestNewManager(t *testing.T) {
	manager := NewManager("test-config.yaml")
	assert.NotNil(t, manager)
	assert.Equal(t, "test-config.yaml", manager.GetConfigPath())
	assert.NotNil(t, manager.Get())
}

func TestLoadFromFile_YAML(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configData := `
server:
  listen_addr: ":4000"
  debug: true
storage:
  root: "/tmp/storage"
  max_file_size: 2147483648
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	manager := NewManager(configPath)
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, ":4000", cfg.Server.ListenAddr)
	assert.Equal(t, true, cfg.Server.Debug)
	assert.Equal(t, "/tmp/storage", cfg.Storage.Root)
	assert.Equal(t, int64(2147483648), cfg.Storage.MaxFileSize)
}

func TestLoadFromFile_JSON(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	configData := `{
  "server": {
    "listen_addr": ":5000",
    "debug": false
  },
  "storage": {
    "root": "/tmp/json-storage",
    "compression": true
  }
}`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	manager := NewManager(configPath)
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, ":5000", cfg.Server.ListenAddr)
	assert.Equal(t, false, cfg.Server.Debug)
	assert.Equal(t, "/tmp/json-storage", cfg.Storage.Root)
	assert.Equal(t, true, cfg.Storage.Compression)
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	manager := NewManager("non-existent.yaml")
	err := manager.Load()
	// Should not error, should use defaults
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, ":3000", cfg.Server.ListenAddr) // default value
}

func TestLoadFromFile_UnsupportedFormat(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.txt")

	err := os.WriteFile(configPath, []byte("invalid format"), 0644)
	require.NoError(t, err)

	manager := NewManager(configPath)
	err = manager.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config file format")
}

func TestLoadFromEnvironment(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"PEERVAULT_LISTEN_ADDR":   ":6000",
		"PEERVAULT_DEBUG":         "true",
		"PEERVAULT_STORAGE_ROOT":  "/env/storage",
		"PEERVAULT_MAX_FILE_SIZE": "3221225472",
		"PEERVAULT_REST_PORT":     "9090",
		"PEERVAULT_GRAPHQL_PORT":  "9091",
		"PEERVAULT_GRPC_PORT":     "9092",
		"PEERVAULT_LOG_LEVEL":     "debug",
		"PEERVAULT_MAX_PEERS":     "50",
		"PEERVAULT_CACHE_SIZE":    "200",
	}

	// Set environment variables
	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	manager := NewManager("")
	err := manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	assert.Equal(t, ":6000", cfg.Server.ListenAddr)
	assert.Equal(t, true, cfg.Server.Debug)
	assert.Equal(t, "/env/storage", cfg.Storage.Root)
	assert.Equal(t, int64(3221225472), cfg.Storage.MaxFileSize)
	assert.Equal(t, 9090, cfg.API.REST.Port)
	assert.Equal(t, 9091, cfg.API.GraphQL.Port)
	assert.Equal(t, 9092, cfg.API.GRPC.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, 50, cfg.Peer.MaxPeers)
	assert.Equal(t, 200, cfg.Performance.CacheSize)
}

func TestLoadFromFileAndEnvironment(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configData := `
server:
  listen_addr: ":4000"
storage:
  root: "/file/storage"
`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Set environment variable to override file
	os.Setenv("PEERVAULT_LISTEN_ADDR", ":7000")
	os.Setenv("PEERVAULT_STORAGE_ROOT", "/env/storage")
	defer func() {
		os.Unsetenv("PEERVAULT_LISTEN_ADDR")
		os.Unsetenv("PEERVAULT_STORAGE_ROOT")
	}()

	manager := NewManager(configPath)
	err = manager.Load()
	require.NoError(t, err)

	cfg := manager.Get()
	// Environment should override file
	assert.Equal(t, ":7000", cfg.Server.ListenAddr)
	assert.Equal(t, "/env/storage", cfg.Storage.Root)
}

func TestSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved-config.yaml")

	manager := NewManager(configPath)

	// Modify the config
	cfg := manager.Get()
	cfg.Server.ListenAddr = ":8000"
	cfg.Storage.Root = "/saved/storage"

	err := manager.Save()
	require.NoError(t, err)

	// Verify file was created and has correct content
	assert.FileExists(t, configPath)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), ":8000")
	assert.Contains(t, string(data), "/saved/storage")
}

func TestSave_NoPath(t *testing.T) {
	manager := NewManager("")
	err := manager.Save()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no config path specified")
}

func TestSetFieldValue(t *testing.T) {
	manager := NewManager("")

	// Test string field
	var strField string
	field := reflect.ValueOf(&strField).Elem()
	err := manager.setFieldValue(field, "test_value")
	require.NoError(t, err)
	assert.Equal(t, "test_value", strField)

	// Test int field
	var intField int
	field = reflect.ValueOf(&intField).Elem()
	err = manager.setFieldValue(field, "42")
	require.NoError(t, err)
	assert.Equal(t, 42, intField)

	// Test bool field
	var boolField bool
	field = reflect.ValueOf(&boolField).Elem()
	err = manager.setFieldValue(field, "true")
	require.NoError(t, err)
	assert.Equal(t, true, boolField)

	// Test duration field
	var durationField time.Duration
	field = reflect.ValueOf(&durationField).Elem()
	err = manager.setFieldValue(field, "5m")
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, durationField)

	// Test string slice field
	var sliceField []string
	field = reflect.ValueOf(&sliceField).Elem()
	err = manager.setFieldValue(field, "a,b,c")
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, sliceField)
}

func TestSetFieldValue_InvalidValues(t *testing.T) {
	manager := NewManager("")

	// Test invalid duration
	var durationField time.Duration
	field := reflect.ValueOf(&durationField).Elem()
	err := manager.setFieldValue(field, "invalid")
	assert.Error(t, err)

	// Test invalid int
	var intField int
	field = reflect.ValueOf(&intField).Elem()
	err = manager.setFieldValue(field, "not_a_number")
	assert.Error(t, err)

	// Test invalid bool
	var boolField bool
	field = reflect.ValueOf(&boolField).Elem()
	err = manager.setFieldValue(field, "not_a_bool")
	assert.Error(t, err)
}

func TestMarshalYAML(t *testing.T) {
	cfg := DefaultConfig()
	data, err := MarshalYAML(cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's valid YAML by trying to unmarshal it back
	var unmarshaled Config
	err = yaml.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, cfg.Server.ListenAddr, unmarshaled.Server.ListenAddr)
}
