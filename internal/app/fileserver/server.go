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
	peerLock   sync.RWMutex
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

	// Copy peer list under read lock to avoid race conditions
	s.peerLock.RLock()
	peers := make([]netp2p.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.peerLock.RUnlock()

	// Send to all peers
	for _, peer := range peers {
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
		// Read encrypted data from disk and decrypt it
		_, encryptedReader, err := s.store.Read(key)
		if err != nil {
			return nil, err
		}
		
		// Create a buffer to hold the decrypted data
		var decryptedBuffer bytes.Buffer
		_, err = crypto.CopyDecrypt(s.getEncryptionKey(), encryptedReader, &decryptedBuffer)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt file: %w", err)
		}
		
		return bytes.NewReader(decryptedBuffer.Bytes()), nil
	}
	fmt.Printf("[%s] dont have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)
	msg := Message{Payload: dto.GetFile{ID: s.ID, Key: crypto.HashKey(key)}}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	
	// Copy peer list under read lock to avoid race conditions
	s.peerLock.RLock()
	peers := make([]netp2p.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.peerLock.RUnlock()
	
	for _, peer := range peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := s.store.WriteDecrypt(crypto.CopyDecrypt, s.getEncryptionKey(), key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		fmt.Printf("[%s] received (%d) bytes over the network from (%s)", s.Transport.Addr(), n, peer.RemoteAddr())
		peer.CloseStream()
	}
	
	// Read and decrypt the file from disk
	_, encryptedReader, err := s.store.Read(key)
	if err != nil {
		return nil, err
	}
	
	var decryptedBuffer bytes.Buffer
	_, err = crypto.CopyDecrypt(s.getEncryptionKey(), encryptedReader, &decryptedBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %w", err)
	}
	
	return bytes.NewReader(decryptedBuffer.Bytes()), nil
}

func (s *Server) Store(key string, r io.Reader) error {
	// Store the file locally with encryption at rest
	size, err := s.store.WriteDecrypt(crypto.CopyEncrypt, s.getEncryptionKey(), key, r)
	if err != nil {
		return err
	}

	// Broadcast the store message to peers
	msg := Message{Payload: dto.StoreFile{ID: s.ID, Key: crypto.HashKey(key), Size: size}}
	if err := s.broadcast(&msg); err != nil {
		return err
	}

	// Stream the file from disk to all peers with resilient replication
	return s.resilientStreamToPeers(key, size)
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

// getPeer safely retrieves a peer under read lock
func (s *Server) getPeer(from string) (netp2p.Peer, bool) {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	peer, ok := s.peers[from]
	return peer, ok
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
	peer, ok := s.getPeer(from)
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
		peer, ok := s.getPeer(from)
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
	peer, ok := s.getPeer(from)
	if ok {
		frameWriter := netp2p.NewFrameWriter(peer)
		if err := frameWriter.WriteMessage(buf.Bytes()); err != nil {
			return err
		}
	}

	// Now receive and store the file with encryption at rest
	peer, ok = s.getPeer(from)
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	n, err := s.store.WriteDecrypt(crypto.CopyEncrypt, s.getEncryptionKey(), msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s] written %d bytes to disk (encrypted)\n", s.Transport.Addr(), n)
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

// resilientStreamToPeers streams a file to peers with retry logic
func (s *Server) resilientStreamToPeers(key string, fileSize int64) error {
	maxRetries := 3

	s.peerLock.RLock()
	peers := make([]netp2p.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.peerLock.RUnlock()

	if len(peers) == 0 {
		log.Printf("[%s] no peers available for replication", s.Transport.Addr())
		return nil
	}

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

	// Try to stream to each peer with retries
	successCount := 0
	for _, peer := range peers {
		peerAddr := peer.RemoteAddr().String()

		for attempt := 0; attempt < maxRetries; attempt++ {
			if err := s.streamToSinglePeer(key, fileReader, peer); err != nil {
				log.Printf("[%s] attempt %d failed to stream to peer %s: %v", s.Transport.Addr(), attempt+1, peerAddr, err)

				if attempt == maxRetries-1 {
					log.Printf("[%s] failed to stream to peer %s after %d attempts", s.Transport.Addr(), peerAddr, maxRetries)
				}
				continue
			}

			// Success
			successCount++
			log.Printf("[%s] successfully streamed to peer %s", s.Transport.Addr(), peerAddr)
			break
		}
	}

	if successCount == 0 {
		return fmt.Errorf("failed to stream to any peer after retries")
	}

	log.Printf("[%s] successfully streamed to %d/%d peers", s.Transport.Addr(), successCount, len(peers))
	return nil
}

// streamToSinglePeer streams a file to a single peer
func (s *Server) streamToSinglePeer(key string, fileReader io.Reader, peer netp2p.Peer) error {
	// Write stream header
	frameWriter := netp2p.NewFrameWriter(peer)
	if err := frameWriter.WriteStreamHeader(); err != nil {
		return fmt.Errorf("failed to write stream header: %w", err)
	}

	// Stream the file with encryption
	n, err := crypto.CopyEncrypt(s.getEncryptionKey(), fileReader, peer)
	if err != nil {
		return fmt.Errorf("failed to stream encrypted file: %w", err)
	}

	log.Printf("[%s] streamed (%d) bytes to peer %s", s.Transport.Addr(), n, peer.RemoteAddr())
	return nil
}
