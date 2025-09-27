package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/mqtt"
	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func main() {
	// Parse command line flags
	var (
		port       = flag.Int("port", 1883, "MQTT broker port")
		host       = flag.String("host", "localhost", "MQTT broker host")
		listenAddr = flag.String("listen", ":3002", "P2P listen address")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
		wsPort     = flag.Int("ws-port", 8085, "MQTT over WebSocket port")
		enableWS   = flag.Bool("enable-ws", true, "Enable MQTT over WebSocket")
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

	// Create file server instance (simplified for MQTT API)
	fileServer := createFileServer(*listenAddr, logger)

	// Create MQTT broker configuration
	brokerConfig := &mqtt.BrokerConfig{
		Port:            *port,
		Host:            *host,
		EnableWebSocket: *enableWS,
		WebSocketPort:   *wsPort,
		KeepAlive:       60 * time.Second,
		MaxConnections:  1000,
		MaxMessageSize:  256 * 1024, // 256KB
		RetainEnabled:   true,
		WillEnabled:     true,
		CleanSession:    true,
	}

	// Create MQTT broker
	broker := mqtt.NewBroker(fileServer, brokerConfig, logger)

	// Start MQTT broker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start TCP listener
	go func() {
		addr := fmt.Sprintf("%s:%d", *host, *port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			logger.Error("Failed to start MQTT TCP listener", "error", err, "addr", addr)
			os.Exit(1)
		}
		defer listener.Close()

		logger.Info("Starting MQTT broker",
			"host", *host,
			"port", *port,
			"protocol", "TCP",
		)

		if err := broker.ServeTCP(ctx, listener); err != nil {
			logger.Error("MQTT TCP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Start WebSocket listener if enabled
	if *enableWS {
		go func() {
			wsAddr := fmt.Sprintf("%s:%d", *host, *wsPort)
			logger.Info("Starting MQTT over WebSocket",
				"host", *host,
				"port", *wsPort,
				"protocol", "WebSocket",
			)

			if err := broker.ServeWebSocket(ctx, wsAddr); err != nil {
				logger.Error("MQTT WebSocket server failed", "error", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for interrupt signal to gracefully shutdown the broker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down MQTT broker...")

	// Cancel context to stop all listeners
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	logger.Info("MQTT broker stopped")
}

// createFileServer creates a file server instance for the MQTT API
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
		BootstrapNodes:    []string{}, // No bootstrap nodes for MQTT API
		ResourceLimits:    peer.DefaultResourceLimits(),
	}
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
