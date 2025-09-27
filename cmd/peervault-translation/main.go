package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/translation"
	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func main() {
	// Parse command line flags
	var (
		port       = flag.Int("port", 8086, "Translation server port")
		host       = flag.String("host", "localhost", "Translation server host")
		listenAddr = flag.String("listen", ":3004", "P2P listen address")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")

		// Protocol endpoints
		websocketAddr = flag.String("websocket", "localhost:8083", "WebSocket server address")
		sseAddr       = flag.String("sse", "localhost:8084", "SSE server address")
		mqttAddr      = flag.String("mqtt", "localhost:1883", "MQTT server address")
		coapAddr      = flag.String("coap", "localhost:5683", "CoAP server address")
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

	// Create file server instance (simplified for translation API)
	fileServer := createFileServer(*listenAddr, logger)

	// Create translation server configuration
	translationConfig := &translation.ServerConfig{
		Port:           *port,
		Host:           *host,
		MaxConnections: 1000,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,

		// Protocol endpoints
		WebSocketAddr: *websocketAddr,
		SSEAddr:       *sseAddr,
		MQTTAddr:      *mqttAddr,
		CoAPAddr:      *coapAddr,

		// Translation settings
		EnableAnalytics: true,
		BufferSize:      1024,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
	}

	// Create translation server
	server := translation.NewServer(fileServer, translationConfig, logger)

	// Start translation server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP listener
	go func() {
		addr := fmt.Sprintf("%s:%d", *host, *port)
		logger.Info("Starting Protocol Translation server",
			"host", *host,
			"port", *port,
			"protocol", "HTTP",
			"endpoints", []string{
				"/translate",
				"/translate/websocket",
				"/translate/sse",
				"/translate/mqtt",
				"/translate/coap",
				"/translate/analytics",
				"/translate/health",
			},
		)

		if err := server.ServeHTTP(ctx, addr); err != nil {
			logger.Error("Translation HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Protocol Translation server...")

	// Cancel context to stop all listeners
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	logger.Info("Protocol Translation server stopped")
}

// createFileServer creates a file server instance for the translation API
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
		BootstrapNodes:    []string{}, // No bootstrap nodes for translation API
		ResourceLimits:    peer.DefaultResourceLimits(),
	}
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
