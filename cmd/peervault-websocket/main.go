package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/websocket"
	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func main() {
	// Parse command line flags
	var (
		port       = flag.Int("port", 8083, "WebSocket server port")
		host       = flag.String("host", "localhost", "WebSocket server host")
		listenAddr = flag.String("listen", ":3000", "P2P listen address")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Create file server instance (simplified for WebSocket API)
	fileServer := createFileServer(*listenAddr, logger)

	// Create WebSocket API server
	wsConfig := &websocket.Config{
		Port:           *port,
		Host:           *host,
		AllowedOrigins: []string{"*"}, // TODO: Make configurable
		EnableCORS:     true,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		PingPeriod:     54 * time.Second,
		PongWait:       60 * time.Second,
	}

	wsServer := websocket.NewServer(fileServer, wsConfig, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *host, *port),
		Handler:      wsServer,
		ReadTimeout:  wsConfig.ReadTimeout,
		WriteTimeout: wsConfig.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting WebSocket API server",
			"host", *host,
			"port", *port,
			"endpoints", []string{
				"/ws",
				"/ws/health",
				"/ws/metrics",
				"/ws/status",
			},
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("WebSocket server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down WebSocket server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("WebSocket server stopped")
}

// createFileServer creates a file server instance for the WebSocket API
func createFileServer(listenAddr string, _ *slog.Logger) *fs.Server {
	// Generate a unique node ID for this server
	nodeID := crypto.GenerateID()

	tcptransportOpts := netp2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	}
	tcpTransport := netp2p.NewTCPTransport(tcptransportOpts)
	fileServerOpts := fs.Options{
		ID:                nodeID,
		EncKey:            crypto.NewEncryptionKey(),
		StorageRoot:       storage.SanitizeStorageRootFromAddr(listenAddr),
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    []string{}, // No bootstrap nodes for WebSocket API
		ResourceLimits:    peer.DefaultResourceLimits(),
	}
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
