package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Skpow1234/Shugo/internal/api/rest"
	"github.com/Skpow1234/Shugo/internal/app/fileserver"
	"github.com/Skpow1234/Shugo/internal/crypto"
	"github.com/Skpow1234/Shugo/internal/peer"
	"github.com/Skpow1234/Shugo/internal/storage"
	netp2p "github.com/Skpow1234/Shugo/internal/transport/p2p"
)

func main() {
	port := flag.Int("port", 8081, "Port to listen on")
	listen := flag.String("listen", ":3010", "P2P listen address for the embedded fileserver")
	bootstrap := flag.String("bootstrap", "", "Comma-separated bootstrap peer addresses")
	storageRoot := flag.String("storage", "shugo-api-storage", "Storage root directory")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	keyManager, err := crypto.NewKeyManager()
	if err != nil {
		logger.Error("failed to create key manager", "error", err)
		os.Exit(1)
	}

	nodeID := crypto.GenerateID()
	tcpTransport := netp2p.NewTCPTransport(netp2p.TCPTransportOpts{
		ListenAddr:    *listen,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	})

	var bootstrapNodes []string
	if *bootstrap != "" {
		bootstrapNodes = []string{*bootstrap}
	}

	fsServer := fileserver.New(fileserver.Options{
		ID:                nodeID,
		KeyManager:        keyManager,
		StorageRoot:       *storageRoot,
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    bootstrapNodes,
		ResourceLimits:    peer.DefaultResourceLimits(),
	})
	tcpTransport.OnPeer = fsServer.OnPeer
	tcpTransport.OnStream = fsServer.OnStream

	go func() {
		if err := fsServer.Start(); err != nil {
			logger.Error("fileserver error", "error", err)
		}
	}()

	restConfig := rest.DefaultConfig()
	restConfig.Port = ":" + fmt.Sprintf("%d", *port)
	server := rest.NewServer(restConfig, logger, fsServer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		logger.Info("REST API listening", "port", restConfig.Port)
		if err := server.Start(); err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")
	fsServer.Stop()
	if err := server.Stop(ctx); err != nil {
		logger.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("Server stopped gracefully")
}
