package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Skpow1234/Peervault/internal/api/grpc"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8082", "gRPC server port")
	authToken := flag.String("auth-token", "demo-token", "Authentication token")
	flag.Parse()

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create server configuration
	config := grpc.DefaultConfig()
	config.Port = ":" + *port
	config.AuthToken = *authToken

	// Create and start server
	server := grpc.NewServer(config, logger)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
		if err := server.Stop(); err != nil {
			logger.Error("Error stopping server", "error", err)
		}
	}()

	// Start server
	logger.Info("Starting PeerVault gRPC server", "port", config.Port)

	if err := server.Start(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	logger.Info("Server shutdown complete")
}
