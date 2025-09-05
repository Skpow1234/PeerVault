package plugins

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// PluginType represents the type of plugin
type PluginType string

const (
	PluginTypeStorage    PluginType = "storage"
	PluginTypeAuth       PluginType = "auth"
	PluginTypeTransport  PluginType = "transport"
	PluginTypeProcessing PluginType = "processing"
)

// Plugin is the base interface that all plugins must implement
type Plugin interface {
	// Plugin metadata
	Name() string
	Version() string
	Description() string
	Type() PluginType

	// Lifecycle methods
	Initialize(config map[string]interface{}) error
	Start() error
	Stop() error

	// Configuration
	GetConfigSchema() map[string]interface{}
	ValidateConfig(config map[string]interface{}) error
}

// StoragePlugin interface for storage backends
type StoragePlugin interface {
	Plugin

	// Storage operations
	Store(ctx context.Context, key string, data io.Reader) error
	Retrieve(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	Exists(ctx context.Context, key string) (bool, error)
	GetMetadata(ctx context.Context, key string) (*FileMetadata, error)
}

// AuthPlugin interface for authentication providers
type AuthPlugin interface {
	Plugin

	// Authentication operations
	Authenticate(ctx context.Context, credentials map[string]interface{}) (*AuthResult, error)
	Authorize(ctx context.Context, userID string, resource string, action string) (bool, error)
	GetUserInfo(ctx context.Context, userID string) (*UserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	RevokeToken(ctx context.Context, token string) error
}

// TransportPlugin interface for transport protocols
type TransportPlugin interface {
	Plugin

	// Transport operations
	Connect(ctx context.Context, address string) (Connection, error)
	Listen(ctx context.Context, address string) (Listener, error)
	GetProtocol() string
	GetCapabilities() []string
}

// ProcessingPlugin interface for file processing
type ProcessingPlugin interface {
	Plugin

	// Processing operations
	Process(ctx context.Context, input io.Reader, config map[string]interface{}) (io.Reader, error)
	GetSupportedFormats() []string
	GetProcessingOptions() map[string]interface{}
}

// Connection represents a transport connection
type Connection interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
	IsConnected() bool
}

// Listener represents a transport listener
type Listener interface {
	Accept() (Connection, error)
	Close() error
	Address() string
}

// FileMetadata contains file metadata
type FileMetadata struct {
	Key         string            `json:"key"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Checksum    string            `json:"checksum"`
	Metadata    map[string]string `json:"metadata"`
}

// AuthResult contains authentication result
type AuthResult struct {
	Success      bool     `json:"success"`
	UserID       string   `json:"user_id"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    string   `json:"expires_at"`
	Permissions  []string `json:"permissions"`
}

// UserInfo contains user information
type UserInfo struct {
	UserID    string            `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	FullName  string            `json:"full_name"`
	Avatar    string            `json:"avatar"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// Manager manages plugins
type Manager struct {
	plugins map[PluginType]map[string]Plugin
	mutex   sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[PluginType]map[string]Plugin),
	}
}

// RegisterPlugin registers a plugin
func (m *Manager) RegisterPlugin(plugin Plugin) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pluginType := plugin.Type()
	if m.plugins[pluginType] == nil {
		m.plugins[pluginType] = make(map[string]Plugin)
	}

	name := plugin.Name()
	if _, exists := m.plugins[pluginType][name]; exists {
		return fmt.Errorf("plugin %s of type %s already registered", name, pluginType)
	}

	m.plugins[pluginType][name] = plugin
	return nil
}

// GetPlugin retrieves a plugin by type and name
func (m *Manager) GetPlugin(pluginType PluginType, name string) (Plugin, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.plugins[pluginType] == nil {
		return nil, fmt.Errorf("no plugins of type %s registered", pluginType)
	}

	plugin, exists := m.plugins[pluginType][name]
	if !exists {
		return nil, fmt.Errorf("plugin %s of type %s not found", name, pluginType)
	}

	return plugin, nil
}

// GetPluginsByType retrieves all plugins of a specific type
func (m *Manager) GetPluginsByType(pluginType PluginType) map[string]Plugin {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.plugins[pluginType] == nil {
		return make(map[string]Plugin)
	}

	// Return a copy to prevent external modification
	result := make(map[string]Plugin)
	for name, plugin := range m.plugins[pluginType] {
		result[name] = plugin
	}

	return result
}

// ListPlugins lists all registered plugins
func (m *Manager) ListPlugins() map[PluginType][]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[PluginType][]string)
	for pluginType, plugins := range m.plugins {
		names := make([]string, 0, len(plugins))
		for name := range plugins {
			names = append(names, name)
		}
		result[pluginType] = names
	}

	return result
}

// InitializePlugin initializes a plugin with configuration
func (m *Manager) InitializePlugin(pluginType PluginType, name string, config map[string]interface{}) error {
	plugin, err := m.GetPlugin(pluginType, name)
	if err != nil {
		return err
	}

	// Validate configuration
	if err := plugin.ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration for plugin %s: %w", name, err)
	}

	// Initialize plugin
	if err := plugin.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

	return nil
}

// StartPlugin starts a plugin
func (m *Manager) StartPlugin(pluginType PluginType, name string) error {
	plugin, err := m.GetPlugin(pluginType, name)
	if err != nil {
		return err
	}

	if err := plugin.Start(); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", name, err)
	}

	return nil
}

// StopPlugin stops a plugin
func (m *Manager) StopPlugin(pluginType PluginType, name string) error {
	plugin, err := m.GetPlugin(pluginType, name)
	if err != nil {
		return err
	}

	if err := plugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop plugin %s: %w", name, err)
	}

	return nil
}

// InitializeAll initializes all plugins with their configurations
func (m *Manager) InitializeAll(configs map[PluginType]map[string]map[string]interface{}) error {
	for pluginType, pluginConfigs := range configs {
		for name, config := range pluginConfigs {
			if err := m.InitializePlugin(pluginType, name, config); err != nil {
				return err
			}
		}
	}
	return nil
}

// StartAll starts all plugins
func (m *Manager) StartAll() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for pluginType, plugins := range m.plugins {
		for name := range plugins {
			if err := m.StartPlugin(pluginType, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// StopAll stops all plugins
func (m *Manager) StopAll() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for pluginType, plugins := range m.plugins {
		for name := range plugins {
			if err := m.StopPlugin(pluginType, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// Global plugin registry
var (
	storagePlugins    = make(map[string]StoragePlugin)
	authPlugins       = make(map[string]AuthPlugin)
	transportPlugins  = make(map[string]TransportPlugin)
	processingPlugins = make(map[string]ProcessingPlugin)
	registryMutex     sync.RWMutex
)

// RegisterStoragePlugin registers a storage plugin
func RegisterStoragePlugin(plugin StoragePlugin) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	storagePlugins[plugin.Name()] = plugin
}

// RegisterAuthPlugin registers an authentication plugin
func RegisterAuthPlugin(plugin AuthPlugin) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	authPlugins[plugin.Name()] = plugin
}

// RegisterTransportPlugin registers a transport plugin
func RegisterTransportPlugin(plugin TransportPlugin) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	transportPlugins[plugin.Name()] = plugin
}

// RegisterProcessingPlugin registers a processing plugin
func RegisterProcessingPlugin(plugin ProcessingPlugin) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	processingPlugins[plugin.Name()] = plugin
}

// GetStoragePlugin retrieves a storage plugin by name
func GetStoragePlugin(name string) (StoragePlugin, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	plugin, exists := storagePlugins[name]
	if !exists {
		return nil, fmt.Errorf("storage plugin %s not found", name)
	}
	return plugin, nil
}

// GetAuthPlugin retrieves an authentication plugin by name
func GetAuthPlugin(name string) (AuthPlugin, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	plugin, exists := authPlugins[name]
	if !exists {
		return nil, fmt.Errorf("auth plugin %s not found", name)
	}
	return plugin, nil
}

// GetTransportPlugin retrieves a transport plugin by name
func GetTransportPlugin(name string) (TransportPlugin, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	plugin, exists := transportPlugins[name]
	if !exists {
		return nil, fmt.Errorf("transport plugin %s not found", name)
	}
	return plugin, nil
}

// GetProcessingPlugin retrieves a processing plugin by name
func GetProcessingPlugin(name string) (ProcessingPlugin, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	plugin, exists := processingPlugins[name]
	if !exists {
		return nil, fmt.Errorf("processing plugin %s not found", name)
	}
	return plugin, nil
}

// ListRegisteredPlugins lists all registered plugins
func ListRegisteredPlugins() map[PluginType][]string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	result := make(map[PluginType][]string)

	// Storage plugins
	storageNames := make([]string, 0, len(storagePlugins))
	for name := range storagePlugins {
		storageNames = append(storageNames, name)
	}
	result[PluginTypeStorage] = storageNames

	// Auth plugins
	authNames := make([]string, 0, len(authPlugins))
	for name := range authPlugins {
		authNames = append(authNames, name)
	}
	result[PluginTypeAuth] = authNames

	// Transport plugins
	transportNames := make([]string, 0, len(transportPlugins))
	for name := range transportPlugins {
		transportNames = append(transportNames, name)
	}
	result[PluginTypeTransport] = transportNames

	// Processing plugins
	processingNames := make([]string, 0, len(processingPlugins))
	for name := range processingPlugins {
		processingNames = append(processingNames, name)
	}
	result[PluginTypeProcessing] = processingNames

	return result
}
