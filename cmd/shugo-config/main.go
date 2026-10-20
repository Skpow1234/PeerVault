package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Skpow1234/Shugo/internal/config"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		outputPath = flag.String("output", "", "Output path for generated configuration")
		format     = flag.String("format", "yaml", "Output format (yaml, json)")
		validate   = flag.Bool("validate", false, "Validate configuration")
		generate   = flag.Bool("generate", false, "Generate default configuration")
		show       = flag.Bool("show", false, "Show current configuration")
		env        = flag.Bool("env", false, "Show environment variable mappings")
		watch      = flag.Bool("watch", false, "Watch configuration file for changes")
	)
	flag.Parse()

	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Determine config path
	if *configPath == "" {
		*configPath = "config/shugo.yaml"
	}

	// Handle different commands
	switch {
	case *generate:
		if err := generateConfig(nil, *outputPath, *format); err != nil {
			logger.Error("Failed to generate configuration", "error", err)
			os.Exit(1)
		}
		logger.Info("Configuration generated successfully", "path", *outputPath)

	default:
		// Create configuration manager for other commands
		manager := config.NewManager(*configPath)

		// Add validators (use configuration setting for demo tokens)
		manager.AddValidator(config.NewSecurityValidator(false)) // Will use config.Security.AllowDemoToken
		manager.AddValidator(&config.DefaultValidator{})
		manager.AddValidator(&config.PortValidator{})
		manager.AddValidator(&config.StorageValidator{})

		switch {
		case *validate:
			if err := validateConfig(manager); err != nil {
				// Check if this is a validation result with warnings only
				if validationResult, ok := err.(*config.ValidationResult); ok {
					if validationResult.HasErrors() {
						logger.Error("Configuration validation failed", "error", err)
						os.Exit(1)
					} else if validationResult.HasWarnings() {
						logger.Warn("Configuration validation completed with warnings", "warnings", err)
						// Don't exit with error code for warnings only
					}
				} else {
					logger.Error("Configuration validation failed", "error", err)
					os.Exit(1)
				}
			} else {
				logger.Info("Configuration validation passed")
			}

		case *show:
			if err := showConfig(manager, *format); err != nil {
				// Check if this is a validation result with warnings only
				if validationResult, ok := err.(*config.ValidationResult); ok {
					if validationResult.HasErrors() {
						logger.Error("Failed to show configuration", "error", err)
						os.Exit(1)
					} else if validationResult.HasWarnings() {
						logger.Warn("Configuration loaded with warnings", "warnings", err)
						// Continue with showing config even with warnings
					}
				} else {
					logger.Error("Failed to show configuration", "error", err)
					os.Exit(1)
				}
			}

		case *env:
			showEnvironmentMappings()

		case *watch:
			if err := watchConfig(manager, logger); err != nil {
				logger.Error("Failed to watch configuration", "error", err)
				os.Exit(1)
			}

		default:
			// Default behavior: validate and show
			if err := validateConfig(manager); err != nil {
				// Check if this is a validation result with warnings only
				if validationResult, ok := err.(*config.ValidationResult); ok {
					if validationResult.HasErrors() {
						logger.Error("Configuration validation failed", "error", err)
						os.Exit(1)
					} else if validationResult.HasWarnings() {
						logger.Warn("Configuration validation completed with warnings", "warnings", err)
						// Don't exit with error code for warnings only
					}
				} else {
					logger.Error("Configuration validation failed", "error", err)
					os.Exit(1)
				}
			} else {
				logger.Info("Configuration validation passed")
			}
		}
	}
}

// generateConfig generates a default configuration file
func generateConfig(manager *config.Manager, outputPath, format string) error {
	// Get default configuration without validation
	cfg := config.DefaultConfig()

	// Determine output path
	if outputPath == "" {
		switch format {
		case "yaml", "yml":
			outputPath = "config/shugo.yaml"
		case "json":
			outputPath = "config/shugo.json"
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save configuration directly
	var data []byte
	var err error
	switch format {
	case "yaml", "yml":
		data, err = config.MarshalYAML(cfg)
	case "json":
		data, err = json.MarshalIndent(cfg, "", "  ")
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

// validateConfig validates the configuration
func validateConfig(manager *config.Manager) error {
	if err := manager.Load(); err != nil {
		// Check if this is a validation result with only warnings
		if validationResult, ok := err.(*config.ValidationResult); ok {
			if validationResult.HasErrors() {
				return fmt.Errorf("failed to load configuration: %w", err)
			} else if validationResult.HasWarnings() {
				// Return the validation result directly for warnings
				cfg := manager.Get()
				fmt.Printf("Configuration loaded from: %s\n", manager.GetConfigPath())
				fmt.Printf("Node ID: %s\n", cfg.Server.NodeID)
				fmt.Printf("Listen Address: %s\n", cfg.Server.ListenAddr)
				fmt.Printf("Storage Root: %s\n", cfg.Storage.Root)
				fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
				return err
			}
		} else {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	cfg := manager.Get()
	fmt.Printf("Configuration loaded from: %s\n", manager.GetConfigPath())
	fmt.Printf("Node ID: %s\n", cfg.Server.NodeID)
	fmt.Printf("Listen Address: %s\n", cfg.Server.ListenAddr)
	fmt.Printf("Storage Root: %s\n", cfg.Storage.Root)
	fmt.Printf("Log Level: %s\n", cfg.Logging.Level)

	return nil
}

// showConfig displays the current configuration
func showConfig(manager *config.Manager, format string) error {
	if err := manager.Load(); err != nil {
		// Check if this is a validation result with only warnings
		if validationResult, ok := err.(*config.ValidationResult); ok {
			if validationResult.HasErrors() {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			// If it has warnings, continue with showing config
			// The warning will be handled by the caller
		} else {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	cfg := manager.Get()

	switch format {
	case "json":
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal configuration: %w", err)
		}
		fmt.Println(string(data))

	case "yaml", "yml":
		data, err := config.MarshalYAML(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal configuration: %w", err)
		}
		fmt.Println(string(data))

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// showEnvironmentMappings displays environment variable mappings
func showEnvironmentMappings() {
	fmt.Println("Environment Variable Mappings:")
	fmt.Println("================================")
	fmt.Println("Server Configuration:")
	fmt.Println("  SHUGO_NODE_ID              - Node ID")
	fmt.Println("  SHUGO_LISTEN_ADDR          - Listen address")
	fmt.Println("  SHUGO_DEBUG                - Enable debug mode")
	fmt.Println("  SHUGO_SHUTDOWN_TIMEOUT     - Shutdown timeout")
	fmt.Println()
	fmt.Println("Storage Configuration:")
	fmt.Println("  SHUGO_STORAGE_ROOT         - Storage root directory")
	fmt.Println("  SHUGO_MAX_FILE_SIZE        - Maximum file size")
	fmt.Println("  SHUGO_COMPRESSION          - Enable compression")
	fmt.Println("  SHUGO_COMPRESSION_LEVEL    - Compression level")
	fmt.Println("  SHUGO_DEDUPLICATION        - Enable deduplication")
	fmt.Println("  SHUGO_CLEANUP_INTERVAL     - Cleanup interval")
	fmt.Println("  SHUGO_RETENTION_PERIOD     - Retention period")
	fmt.Println()
	fmt.Println("Network Configuration:")
	fmt.Println("  SHUGO_BOOTSTRAP_NODES      - Bootstrap nodes (comma-separated)")
	fmt.Println("  SHUGO_CONNECTION_TIMEOUT   - Connection timeout")
	fmt.Println("  SHUGO_READ_TIMEOUT         - Read timeout")
	fmt.Println("  SHUGO_WRITE_TIMEOUT        - Write timeout")
	fmt.Println("  SHUGO_KEEP_ALIVE_INTERVAL  - Keep-alive interval")
	fmt.Println("  SHUGO_MAX_MESSAGE_SIZE     - Maximum message size")
	fmt.Println()
	fmt.Println("Security Configuration:")
	fmt.Println("  SHUGO_CLUSTER_KEY          - Cluster key for encryption")
	fmt.Println("  SHUGO_AUTH_TOKEN           - Authentication token")
	fmt.Println("  SHUGO_TLS                  - Enable TLS")
	fmt.Println("  SHUGO_TLS_CERT_FILE        - TLS certificate file")
	fmt.Println("  SHUGO_TLS_KEY_FILE         - TLS key file")
	fmt.Println("  SHUGO_KEY_ROTATION_INTERVAL - Key rotation interval")
	fmt.Println("  SHUGO_ENCRYPTION_AT_REST   - Enable encryption at rest")
	fmt.Println("  SHUGO_ENCRYPTION_IN_TRANSIT - Enable encryption in transit")
	fmt.Println()
	fmt.Println("Logging Configuration:")
	fmt.Println("  SHUGO_LOG_LEVEL            - Log level (debug, info, warn, error)")
	fmt.Println("  SHUGO_LOG_FORMAT           - Log format (json, text)")
	fmt.Println("  SHUGO_LOG_FILE             - Log file path")
	fmt.Println("  SHUGO_LOG_STRUCTURED       - Enable structured logging")
	fmt.Println("  SHUGO_LOG_INCLUDE_SOURCE   - Include source location")
	fmt.Println("  SHUGO_LOG_MAX_SIZE         - Max log file size (MB)")
	fmt.Println("  SHUGO_LOG_MAX_FILES        - Max log files to keep")
	fmt.Println("  SHUGO_LOG_MAX_AGE          - Max log file age")
	fmt.Println("  SHUGO_LOG_COMPRESS         - Compress rotated logs")
	fmt.Println()
	fmt.Println("API Configuration:")
	fmt.Println("REST API:")
	fmt.Println("  SHUGO_REST_ENABLED         - Enable REST API")
	fmt.Println("  SHUGO_REST_PORT            - REST API port")
	fmt.Println("  SHUGO_REST_ALLOWED_ORIGINS - Allowed origins (comma-separated)")
	fmt.Println("  SHUGO_REST_RATE_LIMIT      - Rate limit per minute")
	fmt.Println("  SHUGO_REST_AUTH_TOKEN      - REST auth token")
	fmt.Println()
	fmt.Println("GraphQL API:")
	fmt.Println("  SHUGO_GRAPHQL_ENABLED      - Enable GraphQL API")
	fmt.Println("  SHUGO_GRAPHQL_PORT         - GraphQL API port")
	fmt.Println("  SHUGO_GRAPHQL_PLAYGROUND   - Enable GraphQL Playground")
	fmt.Println("  SHUGO_GRAPHQL_PATH         - GraphQL endpoint path")
	fmt.Println("  SHUGO_GRAPHQL_PLAYGROUND_PATH - Playground path")
	fmt.Println("  SHUGO_GRAPHQL_ALLOWED_ORIGINS - Allowed origins (comma-separated)")
	fmt.Println()
	fmt.Println("gRPC API:")
	fmt.Println("  SHUGO_GRPC_ENABLED         - Enable gRPC API")
	fmt.Println("  SHUGO_GRPC_PORT            - gRPC API port")
	fmt.Println("  SHUGO_GRPC_AUTH_TOKEN      - gRPC auth token")
	fmt.Println("  SHUGO_GRPC_REFLECTION      - Enable reflection")
	fmt.Println("  SHUGO_GRPC_MAX_STREAMS     - Max concurrent streams")
	fmt.Println()
	fmt.Println("Peer Configuration:")
	fmt.Println("  SHUGO_MAX_PEERS            - Maximum number of peers")
	fmt.Println("  SHUGO_HEARTBEAT_INTERVAL   - Heartbeat interval")
	fmt.Println("  SHUGO_HEALTH_TIMEOUT       - Health timeout")
	fmt.Println("  SHUGO_RECONNECT_BACKOFF    - Reconnection backoff")
	fmt.Println("  SHUGO_MAX_RECONNECT_ATTEMPTS - Max reconnection attempts")
	fmt.Println()
	fmt.Println("Performance Configuration:")
	fmt.Println("  SHUGO_MAX_STREAMS_PER_PEER - Max concurrent streams per peer")
	fmt.Println("  SHUGO_STREAM_BUFFER_SIZE   - Stream buffer size")
	fmt.Println("  SHUGO_CONNECTION_POOL_SIZE - Connection pool size")
	fmt.Println("  SHUGO_ENABLE_MULTIPLEXING  - Enable connection multiplexing")
	fmt.Println("  SHUGO_CACHE_SIZE           - Cache size (MB)")
	fmt.Println("  SHUGO_CACHE_TTL            - Cache TTL")
}

// watchConfig watches the configuration file for changes
func watchConfig(manager *config.Manager, logger *slog.Logger) error {
	if err := manager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("Starting configuration watcher", "file", manager.GetConfigPath())

	// Start watching for changes
	if err := manager.Watch(func(cfg *config.Config) {
		logger.Info("Configuration reloaded",
			"node_id", cfg.Server.NodeID,
			"listen_addr", cfg.Server.ListenAddr,
			"log_level", cfg.Logging.Level)
	}); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	// Keep the watcher running
	select {}
}
