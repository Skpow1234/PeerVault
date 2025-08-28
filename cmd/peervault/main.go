package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"os"

	fs "github.com/anthdm/foreverstore/internal/app/fileserver"
	"github.com/anthdm/foreverstore/internal/crypto"
	"github.com/anthdm/foreverstore/internal/logging"
	"github.com/anthdm/foreverstore/internal/storage"
	netp2p "github.com/anthdm/foreverstore/internal/transport/p2p"
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
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: storage.CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
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
	go s3.Start()
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		s3.Store(key, data)
		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
