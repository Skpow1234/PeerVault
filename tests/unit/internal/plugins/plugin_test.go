package plugins_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Skpow1234/Peervault/internal/plugins"
	"github.com/stretchr/testify/assert"
)

func TestPluginTypeConstants(t *testing.T) {
	// Test plugin type constants
	assert.Equal(t, plugins.PluginType("storage"), plugins.PluginTypeStorage)
	assert.Equal(t, plugins.PluginType("auth"), plugins.PluginTypeAuth)
	assert.Equal(t, plugins.PluginType("transport"), plugins.PluginTypeTransport)
	assert.Equal(t, plugins.PluginType("processing"), plugins.PluginTypeProcessing)
}

func TestPluginTypeString(t *testing.T) {
	// Test plugin type string representation
	assert.Equal(t, "storage", string(plugins.PluginTypeStorage))
	assert.Equal(t, "auth", string(plugins.PluginTypeAuth))
	assert.Equal(t, "transport", string(plugins.PluginTypeTransport))
	assert.Equal(t, "processing", string(plugins.PluginTypeProcessing))
}

func TestPluginTypeValues(t *testing.T) {
	// Test plugin type values
	assert.Equal(t, "storage", string(plugins.PluginTypeStorage))
	assert.Equal(t, "auth", string(plugins.PluginTypeAuth))
	assert.Equal(t, "transport", string(plugins.PluginTypeTransport))
	assert.Equal(t, "processing", string(plugins.PluginTypeProcessing))
}

func TestPluginInterface(t *testing.T) {
	// Test plugin interface implementation
	plugin := &MockPlugin{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin",
		pluginType:  plugins.PluginTypeStorage,
	}
	
	assert.Equal(t, "test-plugin", plugin.Name())
	assert.Equal(t, "1.0.0", plugin.Version())
	assert.Equal(t, "Test plugin", plugin.Description())
	assert.Equal(t, plugins.PluginTypeStorage, plugin.Type())
}

func TestMockPlugin(t *testing.T) {
	// Test mock plugin functionality
	plugin := &MockPlugin{
		name:        "test-plugin",
		version:     "1.0.0",
		description: "Test plugin",
		pluginType:  plugins.PluginTypeStorage,
	}
	
	// Test initialization
	err := plugin.Initialize(map[string]interface{}{"enabled": true})
	assert.NoError(t, err)
	
	// Test start
	err = plugin.Start()
	assert.NoError(t, err)
	assert.True(t, plugin.started)
	
	// Test stop
	err = plugin.Stop()
	assert.NoError(t, err)
	assert.False(t, plugin.started)
	
	// Test config validation
	err = plugin.ValidateConfig(map[string]interface{}{"enabled": true})
	assert.NoError(t, err)
	
	err = plugin.ValidateConfig(map[string]interface{}{"enabled": false})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin is disabled")
}

// MockPlugin is a test implementation of the Plugin interface
type MockPlugin struct {
	name        string
	version     string
	description string
	pluginType  plugins.PluginType
	config      map[string]interface{}
	started     bool
}

func (mp *MockPlugin) Name() string {
	return mp.name
}

func (mp *MockPlugin) Version() string {
	return mp.version
}

func (mp *MockPlugin) Description() string {
	return mp.description
}

func (mp *MockPlugin) Type() plugins.PluginType {
	return mp.pluginType
}

func (mp *MockPlugin) Initialize(config map[string]interface{}) error {
	mp.config = config
	return nil
}

func (mp *MockPlugin) Start() error {
	mp.started = true
	return nil
}

func (mp *MockPlugin) Stop() error {
	mp.started = false
	return nil
}

func (mp *MockPlugin) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"enabled": map[string]interface{}{
				"type":    "boolean",
				"default": true,
			},
		},
	}
}

func (mp *MockPlugin) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	if enabled, ok := config["enabled"].(bool); ok && !enabled {
		return fmt.Errorf("plugin is disabled")
	}

	return nil
}

// MockStoragePlugin is a test implementation of the StoragePlugin interface
type MockStoragePlugin struct {
	*MockPlugin
	data map[string][]byte
}

func NewMockStoragePlugin() *MockStoragePlugin {
	return &MockStoragePlugin{
		MockPlugin: &MockPlugin{
			name:        "mock-storage",
			version:     "1.0.0",
			description: "Mock storage plugin",
			pluginType:  plugins.PluginTypeStorage,
		},
		data: make(map[string][]byte),
	}
}

func (msp *MockStoragePlugin) Store(ctx context.Context, key string, data io.Reader) error {
	bytes, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	msp.data[key] = bytes
	return nil
}

func (msp *MockStoragePlugin) Retrieve(ctx context.Context, key string) (io.ReadCloser, error) {
	data, exists := msp.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (msp *MockStoragePlugin) Delete(ctx context.Context, key string) error {
	delete(msp.data, key)
	return nil
}

func (msp *MockStoragePlugin) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for key := range msp.data {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (msp *MockStoragePlugin) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := msp.data[key]
	return exists, nil
}

func (msp *MockStoragePlugin) GetMetadata(ctx context.Context, key string) (*FileMetadata, error) {
	data, exists := msp.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return &FileMetadata{
		Key:  key,
		Size: int64(len(data)),
		// Add other metadata fields as needed
	}, nil
}
