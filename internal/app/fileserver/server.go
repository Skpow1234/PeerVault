package fileserver

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/anthdm/foreverstore/internal/crypto"
	"github.com/anthdm/foreverstore/internal/dto"
	"github.com/anthdm/foreverstore/internal/storage"
	netp2p "github.com/anthdm/foreverstore/internal/transport/p2p"
)

type Options struct {
	ID                string
	EncKey            []byte
	KeyManager        *crypto.KeyManager
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         netp2p.Transport
	BootstrapNodes    []string
}

type Server struct {
	Options
	KeyManager *crypto.KeyManager
	peerLock   sync.Mutex
	peers      map[string]netp2p.Peer
	store      *storage.Store
	quitch     chan struct{}
}

// getEncryptionKey returns the current encryption key, preferring KeyManager over the legacy EncKey
func (s *Server) getEncryptionKey() []byte {
	if s.KeyManager != nil {
		return s.KeyManager.GetEncryptionKey()
	}
	return s.EncKey
}

func New(opts Options) *Server {
	storeOpts := storage.StoreOpts{Root: opts.StorageRoot, PathTransformFunc: opts.PathTransformFunc}
	if len(opts.ID) == 0 {
		opts.ID = crypto.GenerateID()
	}

	// Initialize KeyManager if not provided
	var keyManager *crypto.KeyManager
	if opts.KeyManager == nil {
		var err error
		keyManager, err = crypto.NewKeyManager()
		if err != nil {
			// Fall back to legacy key generation
			keyManager = nil
		}
	} else {
		keyManager = opts.KeyManager
	}

	return &Server{
		Options:    opts,
		KeyManager: keyManager,
		store:      storage.NewStore(storeOpts),
		quitch:     make(chan struct{}),
		peers:      make(map[string]netp2p.Peer),
	}
}

type Message struct{ Payload any }

func (s *Server) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	payload := buf.Bytes()

	for _, peer := range s.peers {
		frameWriter := netp2p.NewFrameWriter(peer)
		if err := frameWriter.WriteMessage(payload); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(key)
		return r, err
	}
	fmt.Printf("[%s] dont have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)
	msg := Message{Payload: dto.GetFile{ID: s.ID, Key: crypto.HashKey(key)}}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.WriteDecrypt(crypto.CopyDecrypt, s.getEncryptionKey(), key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s)", s.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}
	_, r, err := s.store.Read(key)
	return r, err
}

func (s *Server) Store(key string, r io.Reader) error {
	// First, store the file locally without buffering
	size, err := s.store.Write(key, r)
	if err != nil {
		return err
	}

	// Broadcast the store message to peers
	msg := Message{Payload: dto.StoreFile{ID: s.ID, Key: crypto.HashKey(key), Size: size + 28}}
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	// Stream the file from disk to all peers without buffering in memory
	return s.streamFileToPeers(key, size)
}

// streamFileToPeers streams a file from disk to all peers without buffering in memory
func (s *Server) streamFileToPeers(key string, fileSize int64) error {
	// Read the file from disk
	_, fileReader, err := s.store.Read(key)
	if err != nil {
		return fmt.Errorf("failed to read file for streaming: %w", err)
	}
	defer func() {
		if rc, ok := fileReader.(io.ReadCloser); ok {
			rc.Close()
		}
	}()

	// Create a streaming writer that encrypts and sends to all peers
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	if len(peers) == 0 {
		// No peers to replicate to
		return nil
	}

	// Create a multi-writer for all peers
	mw := io.MultiWriter(peers...)

	// Write stream header using frame writer
	frameWriter := netp2p.NewFrameWriter(mw)
	if err := frameWriter.WriteStreamHeader(); err != nil {
		return fmt.Errorf("failed to write stream header: %w", err)
	}

	// Stream the file directly from disk to peers with encryption
	n, err := crypto.CopyEncrypt(s.getEncryptionKey(), fileReader, mw)
	if err != nil {
		return fmt.Errorf("failed to stream encrypted file: %w", err)
	}

	fmt.Printf("[%s] streamed (%d) bytes to %d peers\n", s.Transport.Addr(), n, len(peers))
	return nil
}

func (s *Server) Stop() { close(s.quitch) }

func (s *Server) OnPeer(p netp2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected with remote %s", p.RemoteAddr())
	return nil
}

func (s *Server) loop() {
	defer func() { log.Println("file server stopped due to error or user quit action"); s.Transport.Close() }()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *Server) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case dto.StoreFile:
		return s.handleMessageStoreFile(from, v)
	case dto.GetFile:
		return s.handleMessageGetFile(from, v)
	}
	return nil
}

func (s *Server) handleMessageGetFile(from string, msg dto.GetFile) error {
	// Check if we have the file
	hasFile := s.store.Has(msg.Key)
	var fileSize int64

	if hasFile {
		// Get file size for acknowledgment
		size, _, err := s.store.Read(msg.Key)
		if err != nil {
			fileSize = 0
		} else {
			fileSize = size
		}
	}

	// Send acknowledgment
	ack := dto.GetFileAck{
		RequestID: msg.ID, // Use the request ID
		Key:       msg.Key,
		HasFile:   hasFile,
		FileSize:  fileSize,
	}

	ackMsg := Message{Payload: ack}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(ackMsg); err != nil {
		return err
	}

	// Send acknowledgment to the requesting peer
	peer, ok := s.peers[from]
	if ok {
		frameWriter := netp2p.NewFrameWriter(peer)
		if err := frameWriter.WriteMessage(buf.Bytes()); err != nil {
			return err
		}
	}

	// If we have the file, serve it
	if hasFile {
		fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)
		fileSize, r, err := s.store.Read(msg.Key)
		if err != nil {
			return err
		}
		if rc, ok := r.(io.ReadCloser); ok {
			fmt.Println("closing readCloser")
			defer rc.Close()
		}
		peer, ok := s.peers[from]
		if !ok {
			return fmt.Errorf("peer %s not in map", from)
		}
		peer.Send([]byte{netp2p.IncomingStream})
		binary.Write(peer, binary.LittleEndian, fileSize)
		n, err := io.Copy(peer, r)
		if err != nil {
			return err
		}
		fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)
	}

	return nil
}

func (s *Server) handleMessageStoreFile(from string, msg dto.StoreFile) error {
	// Send acknowledgment immediately
	ack := dto.StoreFileAck{
		RequestID: msg.ID, // Use the request ID
		Key:       msg.Key,
		Success:   true,
		Error:     "",
	}

	ackMsg := Message{Payload: ack}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(ackMsg); err != nil {
		return err
	}

	// Send acknowledgment to the requesting peer
	peer, ok := s.peers[from]
	if ok {
		frameWriter := netp2p.NewFrameWriter(peer)
		if err := frameWriter.WriteMessage(buf.Bytes()); err != nil {
			return err
		}
	}

	// Now receive and store the file
	peer, ok = s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written %d bytes to disk\n", s.Transport.Addr(), n)
	peer.CloseStream()
	return nil
}

func (s *Server) BootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", s.Transport.Addr(), addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}
	return nil
}

func (s *Server) Start() error {
	fmt.Printf("[%s] starting fileserver...\n", s.Transport.Addr())
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.BootstrapNetwork()
	s.loop()
	return nil
}

func init() {
	gob.Register(dto.StoreFile{})
	gob.Register(dto.GetFile{})
	gob.Register(dto.StoreFileAck{})
	gob.Register(dto.GetFileAck{})
}
