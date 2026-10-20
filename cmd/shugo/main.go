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

func makeServer(listenAddr string, keyManager *crypto.KeyManager, nodes ...string) *fs.Server {
	nodeID := crypto.GenerateID()

	tcptransportOpts := netp2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	}
	tcpTransport := netp2p.NewTCPTransport(tcptransportOpts)
	fileServerOpts := fs.Options{
		ID:                nodeID,
		KeyManager:        keyManager,
		StorageRoot:       storage.SanitizeStorageRootFromAddr(listenAddr),
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
		ResourceLimits:    peer.DefaultResourceLimits(),
	}
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	tcpTransport.OnStream = s.OnStream
	return s
}

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logging.ConfigureLogger(logLevel)

	slog.Info("starting PeerVault application", "log_level", logLevel)

	// Shared cluster key so all local demo nodes can decrypt replicated ciphertext.
	clusterKey := make([]byte, 32)
	copy(clusterKey, []byte("peervault-demo-cluster-key-32b!!"))
	keyManager, err := crypto.NewKeyManagerFromClusterKey(clusterKey)
	if err != nil {
		log.Fatal(err)
	}

	s1 := makeServer(":3000", keyManager, "")
	s2 := makeServer(":7000", keyManager, "")
	s3 := makeServer(":5000", keyManager, ":3000", ":7000")
	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(200 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(200 * time.Millisecond)
	go func() {
		if err := s3.Start(); err != nil {
			log.Printf("s3.Start() error: %v", err)
		}
	}()

	// Allow bootstrap dials to complete
	time.Sleep(1 * time.Second)

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
