package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure for PeerVault
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server" json:"server"`

	// Storage configuration
	Storage StorageConfig `yaml:"storage" json:"storage"`

	// Network configuration
	Network NetworkConfig `yaml:"network" json:"network"`

	// Security configuration
	Security SecurityConfig `yaml:"security" json:"security"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// API configuration
	API APIConfig `yaml:"api" json:"api"`

	// Peer configuration
	Peer PeerConfig `yaml:"peer" json:"peer"`

	// Performance configuration
	Performance PerformanceConfig `yaml:"performance" json:"performance"`
}

// ServerConfig contains server-specific configuration
type ServerConfig struct {
	// Node ID (auto-generated if not provided)
	NodeID string `yaml:"node_id" json:"node_id" env:"PEERVAULT_NODE_ID"`

	// Listen address for the server
	ListenAddr string `yaml:"listen_addr" json:"listen_addr" env:"PEERVAULT_LISTEN_ADDR" default:":3000"`

	// Enable debug mode
	Debug bool `yaml:"debug" json:"debug" env:"PEERVAULT_DEBUG" default:"false"`

	// Graceful shutdown timeout
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout" env:"PEERVAULT_SHUTDOWN_TIMEOUT" default:"30s"`
}

// StorageConfig contains storage-specific configuration
type StorageConfig struct {
	// Storage root directory
	Root string `yaml:"root" json:"root" env:"PEERVAULT_STORAGE_ROOT" default:"./storage"`

	// Maximum file size in bytes
	MaxFileSize int64 `yaml:"max_file_size" json:"max_file_size" env:"PEERVAULT_MAX_FILE_SIZE" default:"1073741824"` // 1GB

	// Enable compression
	Compression bool `yaml:"compression" json:"compression" env:"PEERVAULT_COMPRESSION" default:"false"`

	// Compression level (1-9)
	CompressionLevel int `yaml:"compression_level" json:"compression_level" env:"PEERVAULT_COMPRESSION_LEVEL" default:"6"`

	// Enable deduplication
	Deduplication bool `yaml:"deduplication" json:"deduplication" env:"PEERVAULT_DEDUPLICATION" default:"false"`

	// Storage cleanup interval
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval" env:"PEERVAULT_CLEANUP_INTERVAL" default:"1h"`

	// Retention period for deleted files
	RetentionPeriod time.Duration `yaml:"retention_period" json:"retention_period" env:"PEERVAULT_RETENTION_PERIOD" default:"24h"`
}

// NetworkConfig contains network-specific configuration
type NetworkConfig struct {
	// Bootstrap nodes (comma-separated)
	BootstrapNodes []string `yaml:"bootstrap_nodes" json:"bootstrap_nodes" env:"PEERVAULT_BOOTSTRAP_NODES"`

	// Connection timeout
	ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout" env:"PEERVAULT_CONNECTION_TIMEOUT" default:"30s"`

	// Read timeout
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout" env:"PEERVAULT_READ_TIMEOUT" default:"60s"`

	// Write timeout
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" env:"PEERVAULT_WRITE_TIMEOUT" default:"60s"`

	// Keep-alive interval
	KeepAliveInterval time.Duration `yaml:"keep_alive_interval" json:"keep_alive_interval" env:"PEERVAULT_KEEP_ALIVE_INTERVAL" default:"30s"`

	// Maximum message size
	MaxMessageSize int64 `yaml:"max_message_size" json:"max_message_size" env:"PEERVAULT_MAX_MESSAGE_SIZE" default:"1048576"` // 1MB
}

// SecurityConfig contains security-specific configuration
type SecurityConfig struct {
	// Cluster key for encryption
	ClusterKey string `yaml:"cluster_key" json:"cluster_key" env:"PEERVAULT_CLUSTER_KEY"`

	// Authentication token
	AuthToken string `yaml:"auth_token" json:"auth_token" env:"PEERVAULT_AUTH_TOKEN" default:"demo-token"`

	// Enable TLS
	TLS bool `yaml:"tls" json:"tls" env:"PEERVAULT_TLS" default:"false"`

	// TLS certificate file
	TLSCertFile string `yaml:"tls_cert_file" json:"tls_cert_file" env:"PEERVAULT_TLS_CERT_FILE"`

	// TLS key file
	TLSKeyFile string `yaml:"tls_key_file" json:"tls_key_file" env:"PEERVAULT_TLS_KEY_FILE"`

	// Key rotation interval
	KeyRotationInterval time.Duration `yaml:"key_rotation_interval" json:"key_rotation_interval" env:"PEERVAULT_KEY_ROTATION_INTERVAL" default:"24h"`

	// Enable encryption at rest
	EncryptionAtRest bool `yaml:"encryption_at_rest" json:"encryption_at_rest" env:"PEERVAULT_ENCRYPTION_AT_REST" default:"true"`

	// Enable encryption in transit
	EncryptionInTransit bool `yaml:"encryption_in_transit" json:"encryption_in_transit" env:"PEERVAULT_ENCRYPTION_IN_TRANSIT" default:"true"`
}

// LoggingConfig contains logging-specific configuration
type LoggingConfig struct {
	// Log level (debug, info, warn, error)
	Level string `yaml:"level" json:"level" env:"PEERVAULT_LOG_LEVEL" default:"info"`

	// Log format (json, text)
	Format string `yaml:"format" json:"format" env:"PEERVAULT_LOG_FORMAT" default:"json"`

	// Log file path (optional)
	File string `yaml:"file" json:"file" env:"PEERVAULT_LOG_FILE"`

	// Enable structured logging
	Structured bool `yaml:"structured" json:"structured" env:"PEERVAULT_LOG_STRUCTURED" default:"true"`

	// Include source location in logs
	IncludeSource bool `yaml:"include_source" json:"include_source" env:"PEERVAULT_LOG_INCLUDE_SOURCE" default:"false"`

	// Log rotation settings
	Rotation LogRotationConfig `yaml:"rotation" json:"rotation"`
}

// LogRotationConfig contains log rotation settings
type LogRotationConfig struct {
	// Maximum log file size in MB
	MaxSize int `yaml:"max_size" json:"max_size" env:"PEERVAULT_LOG_MAX_SIZE" default:"100"`

	// Maximum number of log files to keep
	MaxFiles int `yaml:"max_files" json:"max_files" env:"PEERVAULT_LOG_MAX_FILES" default:"10"`

	// Maximum age of log files
	MaxAge time.Duration `yaml:"max_age" json:"max_age" env:"PEERVAULT_LOG_MAX_AGE" default:"168h"` // 7 days

	// Compress rotated log files
	Compress bool `yaml:"compress" json:"compress" env:"PEERVAULT_LOG_COMPRESS" default:"true"`
}

// APIConfig contains API-specific configuration
type APIConfig struct {
	// REST API configuration
	REST RESTConfig `yaml:"rest" json:"rest"`

	// GraphQL API configuration
	GraphQL GraphQLConfig `yaml:"graphql" json:"graphql"`

	// gRPC API configuration
	GRPC GRPCConfig `yaml:"grpc" json:"grpc"`
}

// RESTConfig contains REST API configuration
type RESTConfig struct {
	// Enable REST API
	Enabled bool `yaml:"enabled" json:"enabled" env:"PEERVAULT_REST_ENABLED" default:"true"`

	// REST API port
	Port int `yaml:"port" json:"port" env:"PEERVAULT_REST_PORT" default:"8080"`

	// Allowed origins for CORS
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins" env:"PEERVAULT_REST_ALLOWED_ORIGINS" default:"*"`

	// Rate limit per minute
	RateLimitPerMin int `yaml:"rate_limit_per_min" json:"rate_limit_per_min" env:"PEERVAULT_REST_RATE_LIMIT" default:"100"`

	// Authentication token
	AuthToken string `yaml:"auth_token" json:"auth_token" env:"PEERVAULT_REST_AUTH_TOKEN" default:"demo-token"`
}

// GraphQLConfig contains GraphQL API configuration
type GraphQLConfig struct {
	// Enable GraphQL API
	Enabled bool `yaml:"enabled" json:"enabled" env:"PEERVAULT_GRAPHQL_ENABLED" default:"true"`

	// GraphQL API port
	Port int `yaml:"port" json:"port" env:"PEERVAULT_GRAPHQL_PORT" default:"8081"`

	// Enable GraphQL Playground
	EnablePlayground bool `yaml:"enable_playground" json:"enable_playground" env:"PEERVAULT_GRAPHQL_PLAYGROUND" default:"true"`

	// GraphQL endpoint path
	GraphQLPath string `yaml:"graphql_path" json:"graphql_path" env:"PEERVAULT_GRAPHQL_PATH" default:"/graphql"`

	// Playground path
	PlaygroundPath string `yaml:"playground_path" json:"playground_path" env:"PEERVAULT_GRAPHQL_PLAYGROUND_PATH" default:"/playground"`

	// Allowed origins for CORS
	AllowedOrigins []string `yaml:"allowed_origins" json:"allowed_origins" env:"PEERVAULT_GRAPHQL_ALLOWED_ORIGINS" default:"*"`
}

// GRPCConfig contains gRPC API configuration
type GRPCConfig struct {
	// Enable gRPC API
	Enabled bool `yaml:"enabled" json:"enabled" env:"PEERVAULT_GRPC_ENABLED" default:"true"`

	// gRPC API port
	Port int `yaml:"port" json:"port" env:"PEERVAULT_GRPC_PORT" default:"8082"`

	// Authentication token
	AuthToken string `yaml:"auth_token" json:"auth_token" env:"PEERVAULT_GRPC_AUTH_TOKEN" default:"demo-token"`

	// Enable reflection
	EnableReflection bool `yaml:"enable_reflection" json:"enable_reflection" env:"PEERVAULT_GRPC_REFLECTION" default:"true"`

	// Maximum concurrent streams
	MaxConcurrentStreams int `yaml:"max_concurrent_streams" json:"max_concurrent_streams" env:"PEERVAULT_GRPC_MAX_STREAMS" default:"100"`
}

// PeerConfig contains peer-specific configuration
type PeerConfig struct {
	// Maximum number of peers
	MaxPeers int `yaml:"max_peers" json:"max_peers" env:"PEERVAULT_MAX_PEERS" default:"100"`

	// Heartbeat interval
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" json:"heartbeat_interval" env:"PEERVAULT_HEARTBEAT_INTERVAL" default:"30s"`

	// Health timeout
	HealthTimeout time.Duration `yaml:"health_timeout" json:"health_timeout" env:"PEERVAULT_HEALTH_TIMEOUT" default:"90s"`

	// Reconnection backoff
	ReconnectBackoff time.Duration `yaml:"reconnect_backoff" json:"reconnect_backoff" env:"PEERVAULT_RECONNECT_BACKOFF" default:"5s"`

	// Maximum reconnection attempts
	MaxReconnectAttempts int `yaml:"max_reconnect_attempts" json:"max_reconnect_attempts" env:"PEERVAULT_MAX_RECONNECT_ATTEMPTS" default:"10"`
}

// PerformanceConfig contains performance-specific configuration
type PerformanceConfig struct {
	// Maximum concurrent streams per peer
	MaxConcurrentStreamsPerPeer int `yaml:"max_concurrent_streams_per_peer" json:"max_concurrent_streams_per_peer" env:"PEERVAULT_MAX_STREAMS_PER_PEER" default:"10"`

	// Buffer size for streaming
	StreamBufferSize int `yaml:"stream_buffer_size" json:"stream_buffer_size" env:"PEERVAULT_STREAM_BUFFER_SIZE" default:"65536"` // 64KB

	// Connection pool size
	ConnectionPoolSize int `yaml:"connection_pool_size" json:"connection_pool_size" env:"PEERVAULT_CONNECTION_POOL_SIZE" default:"10"`

	// Enable connection multiplexing
	EnableMultiplexing bool `yaml:"enable_multiplexing" json:"enable_multiplexing" env:"PEERVAULT_ENABLE_MULTIPLEXING" default:"true"`

	// Cache size in MB
	CacheSize int `yaml:"cache_size" json:"cache_size" env:"PEERVAULT_CACHE_SIZE" default:"100"`

	// Cache TTL
	CacheTTL time.Duration `yaml:"cache_ttl" json:"cache_ttl" env:"PEERVAULT_CACHE_TTL" default:"1h"`
}

// Manager handles configuration loading, validation, and hot reloading
type Manager struct {
	config     *Config
	configPath string
	watcher    *ConfigWatcher
	validators []Validator
}

// Validator interface for configuration validation
type Validator interface {
	Validate(config *Config) error
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			NodeID:          "",
			ListenAddr:      ":3000",
			Debug:           false,
			ShutdownTimeout: 30 * time.Second,
		},
		Storage: StorageConfig{
			Root:             "./storage",
			MaxFileSize:      1073741824, // 1GB
			Compression:      false,
			CompressionLevel: 6,
			Deduplication:    false,
			CleanupInterval:  1 * time.Hour,
			RetentionPeriod:  24 * time.Hour,
		},
		Network: NetworkConfig{
			BootstrapNodes:    []string{},
			ConnectionTimeout: 30 * time.Second,
			ReadTimeout:       60 * time.Second,
			WriteTimeout:      60 * time.Second,
			KeepAliveInterval: 30 * time.Second,
			MaxMessageSize:    1048576, // 1MB
		},
		Security: SecurityConfig{
			ClusterKey:          "",
			AuthToken:           "demo-token",
			TLS:                 false,
			TLSCertFile:         "",
			TLSKeyFile:          "",
			KeyRotationInterval: 24 * time.Hour,
			EncryptionAtRest:    true,
			EncryptionInTransit: true,
		},
		Logging: LoggingConfig{
			Level:         "info",
			Format:        "json",
			File:          "",
			Structured:    true,
			IncludeSource: false,
			Rotation: LogRotationConfig{
				MaxSize:  100,
				MaxFiles: 10,
				MaxAge:   168 * time.Hour, // 7 days
				Compress: true,
			},
		},
		API: APIConfig{
			REST: RESTConfig{
				Enabled:         true,
				Port:            8080,
				AllowedOrigins:  []string{"*"},
				RateLimitPerMin: 100,
				AuthToken:       "demo-token",
			},
			GraphQL: GraphQLConfig{
				Enabled:          true,
				Port:             8081,
				EnablePlayground: true,
				GraphQLPath:      "/graphql",
				PlaygroundPath:   "/playground",
				AllowedOrigins:   []string{"*"},
			},
			GRPC: GRPCConfig{
				Enabled:              true,
				Port:                 8082,
				AuthToken:            "demo-token",
				EnableReflection:     true,
				MaxConcurrentStreams: 100,
			},
		},
		Peer: PeerConfig{
			MaxPeers:             100,
			HeartbeatInterval:    30 * time.Second,
			HealthTimeout:        90 * time.Second,
			ReconnectBackoff:     5 * time.Second,
			MaxReconnectAttempts: 10,
		},
		Performance: PerformanceConfig{
			MaxConcurrentStreamsPerPeer: 10,
			StreamBufferSize:            65536, // 64KB
			ConnectionPoolSize:          10,
			EnableMultiplexing:          true,
			CacheSize:                   100,
			CacheTTL:                    1 * time.Hour,
		},
	}
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		config:     DefaultConfig(),
		configPath: configPath,
		validators: []Validator{},
	}
}

// Load loads configuration from file and environment variables
func (m *Manager) Load() error {
	// Load from file if it exists
	if m.configPath != "" {
		if err := m.loadFromFile(); err != nil {
			return fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Override with environment variables
	if err := m.loadFromEnvironment(); err != nil {
		return fmt.Errorf("failed to load config from environment: %w", err)
	}

	// Validate configuration
	if err := m.validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// loadFromFile loads configuration from YAML or JSON file
func (m *Manager) loadFromFile() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil // File doesn't exist, use defaults
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(m.configPath)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, m.config); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, m.config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	return nil
}

// loadFromEnvironment loads configuration from environment variables
func (m *Manager) loadFromEnvironment() error {
	return m.loadStructFromEnv(reflect.ValueOf(m.config).Elem(), "")
}

// loadStructFromEnv recursively loads struct fields from environment variables
func (m *Manager) loadStructFromEnv(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get environment variable name from tag
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			// Check for nested structs even without env tag
			if field.Kind() == reflect.Struct {
				if err := m.loadStructFromEnv(field, prefix); err != nil {
					return err
				}
			}
			continue
		}

		// Get environment variable value
		envValue := os.Getenv(envTag)
		if envValue == "" {
			// Check for nested structs
			if field.Kind() == reflect.Struct {
				if err := m.loadStructFromEnv(field, envTag+"_"); err != nil {
					return err
				}
			}
			continue
		}

		// Set field value based on type
		if err := m.setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue sets a field value from a string
func (m *Manager) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration: %s", value)
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer: %s", value)
			}
			field.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer: %s", value)
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean: %s", value)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float: %s", value)
		}
		field.SetFloat(floatValue)
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			// Handle string slices (comma-separated values)
			values := strings.Split(value, ",")
			for i, v := range values {
				values[i] = strings.TrimSpace(v)
			}
			field.Set(reflect.ValueOf(values))
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// validate validates the configuration
func (m *Manager) validate() error {
	for _, validator := range m.validators {
		if err := validator.Validate(m.config); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the current configuration
func (m *Manager) Get() *Config {
	return m.config
}

// AddValidator adds a configuration validator
func (m *Manager) AddValidator(validator Validator) {
	m.validators = append(m.validators, validator)
}

// Save saves the current configuration to file
func (m *Manager) Save() error {
	if m.configPath == "" {
		return fmt.Errorf("no config path specified")
	}

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Watch starts watching for configuration file changes
func (m *Manager) Watch(callback func(*Config)) error {
	if m.configPath == "" {
		return fmt.Errorf("no config path specified")
	}

	m.watcher = NewConfigWatcher(m.configPath, func() {
		if err := m.Load(); err != nil {
			// Log error but don't fail
			fmt.Printf("Failed to reload config: %v\n", err)
			return
		}
		callback(m.config)
	})

	return m.watcher.Start()
}

// Stop stops watching for configuration changes
func (m *Manager) Stop() {
	if m.watcher != nil {
		m.watcher.Stop()
	}
}

// GetConfigPath returns the configuration file path
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// MarshalYAML marshals the configuration to YAML
func MarshalYAML(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
