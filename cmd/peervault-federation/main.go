package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/graphql/federation"
)

func main() {
	var (
		port                = flag.Int("port", 8081, "Port for the federation gateway")
		serviceTimeout      = flag.Duration("service-timeout", 30*time.Second, "Timeout for service requests")
		healthCheckInterval = flag.Duration("health-check-interval", 30*time.Second, "Interval for health checks")
		enableHealthChecks  = flag.Bool("enable-health-checks", true, "Enable health checks for services")
		configFile          = flag.String("config", "", "Configuration file path")
		verbose             = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Setup logging
	var logLevel slog.Level
	if *verbose {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Create federation configuration
	config := &federation.FederationConfig{
		GatewayPort:         *port,
		ServiceTimeout:      *serviceTimeout,
		HealthCheckInterval: *healthCheckInterval,
		EnableHealthChecks:  *enableHealthChecks,
	}

	// Load configuration from file if provided
	if *configFile != "" {
		if err := loadConfigFromFile(*configFile, config); err != nil {
			logger.Error("Failed to load configuration", "error", err)
			os.Exit(1)
		}
	}

	// Create federation gateway
	gateway := federation.NewFederationGateway(logger)

	// Register default services if any
	registerDefaultServices(gateway, logger)

	// Create federation server
	server := federation.NewFederationServer(gateway, config)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Start the server
	logger.Info("Starting PeerVault GraphQL Federation Gateway",
		"port", config.GatewayPort,
		"healthChecks", config.EnableHealthChecks,
		"healthCheckInterval", config.HealthCheckInterval,
	)

	if err := server.Start(ctx); err != nil {
		logger.Error("Failed to start federation server", "error", err)
		os.Exit(1)
	}
}

// loadConfigFromFile loads configuration from a file
func loadConfigFromFile(_ string, _ *federation.FederationConfig) error {
	// TODO: Implement configuration file loading
	// For now, just return nil
	return nil
}

// registerDefaultServices registers default services
func registerDefaultServices(gateway *federation.FederationGateway, logger *slog.Logger) {
	// Register the main PeerVault service
	mainService := &federation.FederatedService{
		Name:        "peervault-main",
		URL:         "http://localhost:8080/graphql",
		HealthCheck: "http://localhost:8080/health",
		Capabilities: map[string]bool{
			"files":   true,
			"nodes":   true,
			"storage": true,
			"network": true,
			"metrics": true,
		},
		Metadata: map[string]string{
			"version": "1.0.0",
			"region":  "us-east-1",
		},
	}

	if err := gateway.RegisterService(mainService); err != nil {
		logger.Error("Failed to register main service", "error", err)
	} else {
		logger.Info("Registered main PeerVault service", "name", mainService.Name, "url", mainService.URL)
	}

	// Register additional services if they exist
	additionalServices := []*federation.FederatedService{
		{
			Name:        "peervault-analytics",
			URL:         "http://localhost:8082/graphql",
			HealthCheck: "http://localhost:8082/health",
			Capabilities: map[string]bool{
				"analytics": true,
				"metrics":   true,
				"reporting": true,
			},
			Metadata: map[string]string{
				"version": "1.0.0",
				"region":  "us-east-1",
			},
		},
		{
			Name:        "peervault-storage",
			URL:         "http://localhost:8083/graphql",
			HealthCheck: "http://localhost:8083/health",
			Capabilities: map[string]bool{
				"storage":     true,
				"files":       true,
				"replication": true,
			},
			Metadata: map[string]string{
				"version": "1.0.0",
				"region":  "us-east-1",
			},
		},
	}

	for _, service := range additionalServices {
		if err := gateway.RegisterService(service); err != nil {
			logger.Warn("Failed to register additional service", "name", service.Name, "error", err)
		} else {
			logger.Info("Registered additional service", "name", service.Name, "url", service.URL)
		}
	}
}
