package main

import (
	"flag"
	"log"
	"log/slog"
	"strings"

	fs "github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/logging"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

func main() {
	// Parse command line flags
	var (
		listenAddr     = flag.String("listen", ":3000", "Listen address for this node")
		bootstrapNodes = flag.String("bootstrap", "", "Comma-separated list of bootstrap node addresses")
		logLevel       = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		storagePrefix  = flag.String("storage-prefix", "peervault", "Prefix for storage directory")
	)
	flag.Parse()

	// Configure structured logging
	logging.ConfigureLogger(*logLevel)
	slog.Info("starting PeerVault node",
		"listen_addr", *listenAddr,
		"bootstrap_nodes", *bootstrapNodes,
		"log_level", *logLevel)

	// Parse bootstrap nodes
	var bootstrapList []string
	if *bootstrapNodes != "" {
		bootstrapList = strings.Split(*bootstrapNodes, ",")
		// Trim whitespace from each address
		for i, addr := range bootstrapList {
			bootstrapList[i] = strings.TrimSpace(addr)
		}
	}

	// Create server
	server := makeServer(*listenAddr, *storagePrefix, bootstrapList...)

	// Start the server
	slog.Info("starting PeerVault node server", "address", *listenAddr)
	if err := server.Start(); err != nil {
		log.Fatal("failed to start server:", err)
	}
}

func makeServer(listenAddr, storagePrefix string, bootstrapNodes ...string) *fs.Server {
	// Generate a unique node ID for this server
	nodeID := crypto.GenerateID()

	tcptransportOpts := netp2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: netp2p.AuthenticatedHandshakeFunc(nodeID),
		Decoder:       netp2p.LengthPrefixedDecoder{},
	}
	tcpTransport := netp2p.NewTCPTransport(tcptransportOpts)

	// Create storage root with prefix for better organization in containers
	storageRoot := storage.SanitizeStorageRootFromAddrWithPrefix(listenAddr, storagePrefix)

	fileServerOpts := fs.Options{
		ID:                nodeID,
		EncKey:            crypto.NewEncryptionKey(),
		StorageRoot:       storageRoot,
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    bootstrapNodes,
		ResourceLimits:    peer.DefaultResourceLimits(),
	}
	s := fs.New(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
