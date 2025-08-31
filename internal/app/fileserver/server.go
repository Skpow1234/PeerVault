package fileserver

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/dto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

type Options struct {
	ID                string
	EncKey            []byte
	KeyManager        *crypto.KeyManager
	StorageRoot       string
	PathTransformFunc storage.PathTransformFunc
	Transport         netp2p.Transport
	BootstrapNodes    []string
	ResourceLimits    peer.ResourceLimits
}

type Server struct {
	Options
	KeyManager      *crypto.KeyManager
	peerLock        sync.RWMutex
	peers           map[string]netp2p.Peer
	store           *storage.Store
	quitch          chan struct{}
	healthManager   *peer.HealthManager
	resourceManager *peer.ResourceManager
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

	// Use default resource limits if not provided
	if opts.ResourceLimits.MaxConcurrentStreams == 0 {
		opts.ResourceLimits = peer.DefaultResourceLimits()
	}

	server := &Server{
		Options:    opts,
		KeyManager: keyManager,
		store:      storage.NewStore(storeOpts),
		quitch:     make(chan struct{}),
		peers:      make(map[string]netp2p.Peer),
	}

	// Initialize health manager
	server.initializeHealthManager()

	// Initialize resource manager
	server.resourceManager = peer.NewResourceManager(opts.ResourceLimits)

	return server
}

// initializeHealthManager sets up the peer health monitoring system
func (s *Server) initializeHealthManager() {
	opts := peer.HealthManagerOpts{
		HeartbeatInterval:    30 * time.Second,
		HealthTimeout:        90 * time.Second,
		ReconnectInterval:    5 * time.Second,
		MaxReconnectAttempts: 5,
		OnPeerDisconnect:     s.handlePeerDisconnect,
		OnPeerReconnect:      s.handlePeerReconnect,
		DialFunc:             s.dialPeer,
	}

	s.healthManager = peer.NewHealthManager(opts)
}

// handlePeerDisconnect is called when a peer is disconnected
func (s *Server) handlePeerDisconnect(address string) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	if _, exists := s.peers[address]; exists {
		delete(s.peers, address)
		slog.Info("peer disconnected and removed", "address", address)
	}

	// Remove peer from resource management
	if s.resourceManager != nil {
		s.resourceManager.RemovePeer(address)
	}
}

// handlePeerReconnect is called when a peer reconnects
func (s *Server) handlePeerReconnect(address string, newPeer netp2p.Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[address] = newPeer
	slog.Info("peer reconnected", "address", address)
}

// dialPeer attempts to dial a peer address
func (s *Server) dialPeer(address string) (netp2p.Peer, error) {
	// This would use the transport's dial function
	// For now, we'll return an error as this needs to be implemented
	return nil, fmt.Errorf("dial not implemented yet")
}

type Message struct{ Payload any }

func (s *Server) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	payload := buf.Bytes()

	// Get healthy peers from health manager if available
	var peers []netp2p.Peer
	if s.healthManager != nil {
		peers = s.healthManager.GetHealthyPeers()
	} else {
		// Fallback to all peers if health manager is not available
		s.peerLock.RLock()
		peers = make([]netp2p.Peer, 0, len(s.peers))
		for _, peer := range s.peers {
			peers = append(peers, peer)
		}
		s.peerLock.RUnlock()
	}

	// Send to healthy peers only
	for _, p := range peers {
		frameWriter := netp2p.NewFrameWriter(p)
		if err := frameWriter.WriteMessage(payload); err != nil {
			slog.Warn("failed to send message to peer", "peer", p.RemoteAddr(), "error", err)
			// Update peer health status
			if s.healthManager != nil {
				s.healthManager.UpdatePeerHealth(p.RemoteAddr().String(), peer.StatusUnhealthy)
			}
		}
	}
	return nil
}

func (s *Server) Get(ctx context.Context, key string) (io.Reader, error) {
	if s.store.Has(key) {
		slog.Info("serving file", "key", key, "addr", s.Transport.Addr())
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
	slog.Info("dont have file", "key", key, "addr", s.Transport.Addr())
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
		if err := binary.Read(peer, binary.LittleEndian, &fileSize); err != nil {
			slog.Error("failed to read file size from peer", "peer", peer.RemoteAddr(), "err", err)
			continue
		}
		n, err := s.store.WriteDecrypt(crypto.CopyDecrypt, s.getEncryptionKey(), key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}
		slog.Info("received", "bytes", n, "peer", peer.RemoteAddr())
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

func (s *Server) Store(ctx context.Context, key string, r io.Reader) error {
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
	return s.resilientStreamToPeers(ctx, key, size)
}

func (s *Server) Stop() {
	// Stop health manager
	if s.healthManager != nil {
		s.healthManager.Stop()
	}

	// Shutdown resource manager
	if s.resourceManager != nil {
		s.resourceManager.Shutdown()
	}

	close(s.quitch)
}

func (s *Server) OnPeer(p netp2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p

	// Add peer to health monitoring
	if s.healthManager != nil {
		s.healthManager.AddPeer(p)
	}

	// Add peer to resource management
	if s.resourceManager != nil {
		s.resourceManager.AddPeer(p.RemoteAddr().String())
	}

	slog.Info("connected", "peer", p.RemoteAddr())
	return nil
}

func (s *Server) loop() {
	defer func() {
		slog.Info("file server stopped")
		if err := s.Transport.Close(); err != nil {
			slog.Error("failed to close transport", "err", err)
		}
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				slog.Error("decoding error", "err", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				slog.Error("handle message error", "err", err)
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
		slog.Info("serving file", "key", msg.Key, "addr", s.Transport.Addr())
		fileSize, r, err := s.store.Read(msg.Key)
		if err != nil {
			return err
		}
		// r is already an io.ReadCloser, so we can close it directly
		defer func() {
			if err := r.Close(); err != nil {
				slog.Error("failed to close file reader", "err", err)
			}
		}()
		peer, ok := s.getPeer(from)
		if !ok {
			return fmt.Errorf("peer %s not in map", from)
		}
		if err := peer.Send([]byte{netp2p.IncomingStream}); err != nil {
			slog.Error("failed to send stream header", "peer", from, "err", err)
			return err
		}
		if err := binary.Write(peer, binary.LittleEndian, fileSize); err != nil {
			slog.Error("failed to write file size", "peer", from, "err", err)
			return err
		}
		n, err := io.Copy(peer, r)
		if err != nil {
			return err
		}
		slog.Info("written", "bytes", n, "peer", from)
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
	slog.Info("written", "bytes", n, "addr", s.Transport.Addr())
	peer.CloseStream()
	return nil
}

func (s *Server) BootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			slog.Info("attemping to connect", "addr", addr, "current_addr", s.Transport.Addr())
			if err := s.Transport.Dial(addr); err != nil {
				slog.Error("dial error", "err", err)
			}
		}(addr)
	}
	return nil
}

func (s *Server) Start() error {
	slog.Info("starting fileserver", "addr", s.Transport.Addr())

	// Start health manager
	if s.healthManager != nil {
		s.healthManager.Start()
	}

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	if err := s.BootstrapNetwork(); err != nil {
		slog.Error("failed to bootstrap network", "err", err)
		// Don't return error here as we can still function without bootstrap
	}
	s.loop()
	return nil
}

func init() {
	gob.Register(dto.StoreFile{})
	gob.Register(dto.GetFile{})
	gob.Register(dto.StoreFileAck{})
	gob.Register(dto.GetFileAck{})
}

// contextReader wraps an io.Reader with context cancellation support
type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (cr *contextReader) Read(p []byte) (n int, err error) {
	// Check if context is cancelled before reading
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
	}

	return cr.reader.Read(p)
}

// resilientStreamToPeers streams a file to peers with retry logic
func (s *Server) resilientStreamToPeers(ctx context.Context, key string, fileSize int64) error {
	maxRetries := 3

	s.peerLock.RLock()
	peers := make([]netp2p.Peer, 0, len(s.peers))
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	s.peerLock.RUnlock()

	if len(peers) == 0 {
		slog.Info("no peers available for replication", "addr", s.Transport.Addr())
		return nil
	}

	// Read the file from disk
	_, fileReader, err := s.store.Read(key)
	if err != nil {
		return fmt.Errorf("failed to read file for streaming: %w", err)
	}
	defer func() {
		if closer, ok := fileReader.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				slog.Error("failed to close file reader", "err", err)
			}
		}
	}()

	// Try to stream to each peer with retries
	successCount := 0
	for _, peer := range peers {
		peerAddr := peer.RemoteAddr().String()

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Acquire stream slot with resource management
		streamID := fmt.Sprintf("%s_%s_%d", key, peerAddr, time.Now().UnixNano())
		streamCtx, err := s.resourceManager.AcquireStreamForPeer(ctx, peerAddr, streamID)
		if err != nil {
			slog.Warn("failed to acquire stream slot", "peer", peerAddr, "error", err)
			continue
		}

		// Ensure stream is released when done
		defer s.resourceManager.ReleaseStreamForPeer(peerAddr, streamID)

		for attempt := 0; attempt < maxRetries; attempt++ {
			// Check if stream context is cancelled
			select {
			case <-streamCtx.Done():
				return streamCtx.Err()
			default:
			}

			if err := s.streamToSinglePeer(streamCtx, key, fileReader, peer); err != nil {
				slog.Error("attempt failed to stream to peer", "attempt", attempt+1, "peer", peerAddr, "err", err)

				if attempt == maxRetries-1 {
					slog.Error("failed to stream to peer after retries", "peer", peerAddr, "max_retries", maxRetries)
				}
				continue
			}

			// Success
			successCount++
			slog.Info("successfully streamed to peer", "peer", peerAddr)
			break
		}
	}

	if successCount == 0 {
		return fmt.Errorf("failed to stream to any peer after retries")
	}

	slog.Info("successfully streamed to peers", "success_count", successCount, "total_peers", len(peers))
	return nil
}

// streamToSinglePeer streams a file to a single peer
func (s *Server) streamToSinglePeer(ctx context.Context, key string, fileReader io.Reader, peer netp2p.Peer) error {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Write stream header
	frameWriter := netp2p.NewFrameWriter(peer)
	if err := frameWriter.WriteStreamHeader(); err != nil {
		return fmt.Errorf("failed to write stream header: %w", err)
	}

	// Create a context-aware reader that checks for cancellation
	ctxReader := &contextReader{
		ctx:    ctx,
		reader: fileReader,
	}

	// Stream the file with encryption
	n, err := crypto.CopyEncrypt(s.getEncryptionKey(), ctxReader, peer)
	if err != nil {
		return fmt.Errorf("failed to stream encrypted file: %w", err)
	}

	slog.Info("streamed", "bytes", n, "peer", peer.RemoteAddr())
	return nil
}
