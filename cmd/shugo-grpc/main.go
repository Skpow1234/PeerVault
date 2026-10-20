package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Skpow1234/Peervault/internal/api/grpc"
	"github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func main() {
	port := flag.String("port", "8082", "HTTP/JSON API port (temporary gRPC facade)")
	authToken := flag.String("auth-token", "demo-token", "Authentication token")
	listen := flag.String("listen", ":3020", "P2P listen address for embedded fileserver")
	storageRoot := flag.String("storage", "peervault-grpc-storage", "Storage root")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	keyManager, err := crypto.NewKeyManager()
	if err != nil {
		logger.Error("key manager", "error", err)
		os.Exit(1)
	}
	nodeID := crypto.GenerateID()
	tr := netp2p.NewTCPTransport(netp2p.TCPTransportOpts{
		ListenAddr:    *listen,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	})
	fsServer := fileserver.New(fileserver.Options{
		ID:                nodeID,
		KeyManager:        keyManager,
		StorageRoot:       *storageRoot,
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tr,
		ResourceLimits:    peer.DefaultResourceLimits(),
	})
	tr.OnPeer = fsServer.OnPeer
	tr.OnStream = fsServer.OnStream
	go func() {
		if err := fsServer.Start(); err != nil {
			logger.Error("fileserver error", "error", err)
		}
	}()

	config := grpc.DefaultConfig()
	config.Port = ":" + *port
	config.AuthToken = *authToken
	server := grpc.NewServer(config, logger, fsServer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
		fsServer.Stop()
		if err := server.Stop(); err != nil {
			logger.Error("Error stopping server", "error", err)
		}
	}()

	logger.Info("Starting PeerVault gRPC/HTTP API", "port", config.Port)
	if err := server.Start(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
	<-ctx.Done()
	logger.Info("Server shutdown complete")
}
