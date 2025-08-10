package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	fs "github.com/anthdm/foreverstore/internal/app/fileserver"
	"github.com/anthdm/foreverstore/internal/crypto"
	"github.com/anthdm/foreverstore/internal/storage"
	netp2p "github.com/anthdm/foreverstore/internal/transport/p2p"
)

func makeServer(listenAddr string, nodes ...string) *fs.Server {
	tcptransportOpts := netp2p.TCPTransportOpts{ListenAddr: listenAddr, HandshakeFunc: netp2p.NOPHandshakeFunc, Decoder: netp2p.DefaultDecoder{}}
	tcpTransport := netp2p.NewTCPTransport(tcptransportOpts)
	fileServerOpts := fs.Options{
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
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7000", "")
	s3 := makeServer(":5000", ":3000", ":7000")
	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(2 * time.Second)
	go s3.Start()
	time.Sleep(2 * time.Second)
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
