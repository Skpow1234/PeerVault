package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/logging"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func makeServer(listenAddr string, nodes ...string) *fs.Server {
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
		BootstrapNodes:    nodes,
		ResourceLimits:    peer.DefaultResourceLimits(),
	}
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	// Configure structured logging
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logging.ConfigureLogger(logLevel)

	slog.Info("starting PeerVault application", "log_level", logLevel)

	s1 := makeServer(":3000", "")
	s2 := makeServer(":7000", "")
	s3 := makeServer(":5000", ":3000", ":7000")
	go func() { log.Fatal(s1.Start()) }()
	go func() { log.Fatal(s2.Start()) }()
	go func() {
		if err := s3.Start(); err != nil {
			log.Printf("s3.Start() error: %v", err)
		}
	}()

	// Create a context with timeout for the operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		if err := s3.Store(ctx, key, data); err != nil {
			log.Fatal(err)
		}
		r, err := s3.Get(ctx, key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
