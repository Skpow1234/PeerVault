package fileserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/crypto"
	"github.com/Skpow1234/Peervault/internal/dto"
	"github.com/Skpow1234/Peervault/internal/peer"
	"github.com/Skpow1234/Peervault/internal/storage"
	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

const defaultGetTimeout = 15 * time.Second

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

type streamResult struct {
	data []byte
	err  error
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
	fileOpManager   *FileOperationManager

	pendingMu   sync.Mutex
	pendingGets map[string]chan streamResult

	registry  *fileRegistry
	startTime time.Time
}

// getEncryptionKey returns the current encryption key, preferring KeyManager over the legacy EncKey
func (s *Server) getEncryptionKey() []byte {
	if s.KeyManager != nil {
		return s.KeyManager.GetEncryptionKey()
	}
	return s.EncKey
}

// storageKey returns the content-addressed key used on disk and on the wire.
func storageKey(logicalKey string) string {
	return crypto.HashKey(logicalKey)
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
		Options:         opts,
		KeyManager:      keyManager,
		store:           storage.NewStore(storeOpts),
		quitch:          make(chan struct{}),
		peers:           make(map[string]netp2p.Peer),
		pendingGets:     make(map[string]chan streamResult),
		fileOpManager:   NewFileOperationManager(20),
		resourceManager: peer.NewResourceManager(opts.ResourceLimits),
		registry:        newFileRegistry(),
		startTime:       time.Now().UTC(),
	}

	server.initializeHealthManager()
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

func (s *Server) handlePeerDisconnect(address string) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	if _, exists := s.peers[address]; exists {
		delete(s.peers, address)
		slog.Info("peer disconnected and removed", "address", address)
	}

	if s.resourceManager != nil {
		s.resourceManager.RemovePeer(address)
	}
}

func (s *Server) handlePeerReconnect(address string, newPeer netp2p.Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[address] = newPeer
	slog.Info("peer reconnected", "address", address)
}

// dialPeer dials a peer address via the transport and returns the connected peer.
func (s *Server) dialPeer(address string) (netp2p.Peer, error) {
	if s.Transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return s.Transport.DialPeer(address)
}

type Message struct{ Payload any }

func (s *Server) listPeers() []netp2p.Peer {
	if s.healthManager != nil {
		if peers := s.healthManager.GetHealthyPeers(); len(peers) > 0 {
			return peers
		}
	}

	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	peers := make([]netp2p.Peer, 0, len(s.peers))
	for _, p := range s.peers {
		peers = append(peers, p)
	}
	return peers
}

func (s *Server) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	payload := buf.Bytes()
	for _, p := range s.listPeers() {
		frameWriter := netp2p.NewFrameWriter(p)
		if err := frameWriter.WriteMessage(payload); err != nil {
			slog.Warn("failed to send message to peer", "peer", p.RemoteAddr(), "error", err)
			if s.healthManager != nil {
				s.healthManager.UpdatePeerHealth(p.RemoteAddr().String(), peer.StatusUnhealthy)
			}
		}
	}
	return nil
}

func (s *Server) readDecrypt(key string) (io.Reader, error) {
	_, encryptedReader, err := s.store.Read(key)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closer, ok := encryptedReader.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	var decryptedBuffer bytes.Buffer
	if _, err := crypto.CopyDecrypt(s.getEncryptionKey(), encryptedReader, &decryptedBuffer); err != nil {
		return nil, fmt.Errorf("failed to decrypt file: %w", err)
	}
	return bytes.NewReader(decryptedBuffer.Bytes()), nil
}

func (s *Server) registerPendingGet(key string) chan streamResult {
	ch := make(chan streamResult, 1)
	s.pendingMu.Lock()
	s.pendingGets[key] = ch
	s.pendingMu.Unlock()
	return ch
}

func (s *Server) unregisterPendingGet(key string) {
	s.pendingMu.Lock()
	delete(s.pendingGets, key)
	s.pendingMu.Unlock()
}

func (s *Server) fulfillPendingGet(key string, data []byte, err error) {
	s.pendingMu.Lock()
	ch, ok := s.pendingGets[key]
	if ok {
		delete(s.pendingGets, key)
	}
	s.pendingMu.Unlock()
	if !ok {
		return
	}
	select {
	case ch <- streamResult{data: data, err: err}:
	default:
	}
}

func (s *Server) Get(ctx context.Context, key string) (io.Reader, error) {
	sk := storageKey(key)

	if s.store.Has(sk) {
		slog.Info("serving file locally", "key", key, "addr", s.Transport.Addr())
		return s.readDecrypt(sk)
	}

	slog.Info("dont have file locally, requesting peers", "key", key, "addr", s.Transport.Addr())
	pending := s.registerPendingGet(sk)
	defer s.unregisterPendingGet(sk)

	msg := Message{Payload: dto.GetFile{ID: crypto.GenerateID(), Key: sk}}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	timeout := defaultGetTimeout
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}

	select {
	case res := <-pending:
		if res.err != nil {
			return nil, res.err
		}
		return bytes.NewReader(res.data), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("file not found on any peer")
	}
}

func (s *Server) Store(ctx context.Context, key string, r io.Reader) error {
	return s.StoreWithMeta(ctx, key, r, "", nil)
}

// StoreWithMeta stores content and records API metadata (name/content-type/etc).
func (s *Server) StoreWithMeta(ctx context.Context, key string, r io.Reader, contentType string, metadata map[string]string) error {
	_ = ctx
	sk := storageKey(key)

	// Tee plaintext so we can hash/size for the registry without a second read.
	var plainBuf bytes.Buffer
	tee := io.TeeReader(r, &plainBuf)

	size, err := s.store.WriteDecrypt(crypto.CopyEncrypt, s.getEncryptionKey(), sk, tee)
	if err != nil {
		return err
	}

	plain := plainBuf.Bytes()
	sum := sha256.Sum256(plain)
	now := time.Now().UTC()
	name := key
	if metadata != nil {
		if n := metadata["name"]; n != "" {
			name = n
		}
	}
	s.registry.put(&FileInfo{
		Key:         key,
		Name:        name,
		Size:        int64(len(plain)),
		ContentType: contentType,
		Hash:        hex.EncodeToString(sum[:]),
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    metadata,
	})

	msg := Message{Payload: dto.StoreFile{ID: s.ID, Key: sk, Size: size}}
	if err := s.broadcast(&msg); err != nil {
		slog.Warn("failed to broadcast store announcement", "error", err)
	}

	if err := s.replicateToPeers(sk); err != nil {
		slog.Warn("replication completed with errors", "key", key, "error", err)
	}

	slog.Info("file stored and replicated", "key", key, "storage_key", sk, "size", size)
	return nil
}

// Has reports whether the logical key exists locally.
func (s *Server) Has(key string) bool {
	return s.store.Has(storageKey(key))
}

// Delete removes a logical key from local storage.
func (s *Server) Delete(ctx context.Context, key string) error {
	_ = ctx
	sk := storageKey(key)
	if !s.store.Has(sk) {
		return fmt.Errorf("file not found: %s", key)
	}
	if err := s.store.Delete(sk); err != nil {
		return err
	}
	s.registry.delete(key)
	return nil
}

// NodeID returns this node's identifier.
func (s *Server) NodeID() string {
	return s.ID
}

// ListFiles returns registered file metadata.
func (s *Server) ListFiles() []FileInfo {
	return s.registry.list()
}

// GetFileInfo returns metadata for a logical key.
func (s *Server) GetFileInfo(key string) (*FileInfo, bool) {
	return s.registry.get(key)
}

// UpdateFileMetadata updates metadata for a registered file.
func (s *Server) UpdateFileMetadata(key string, metadata map[string]string) (*FileInfo, error) {
	if !s.Has(key) {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	info, ok := s.registry.updateMetadata(key, metadata)
	if !ok {
		// File exists on disk but not in registry — synthesize a record.
		now := time.Now().UTC()
		info = &FileInfo{Key: key, Name: key, CreatedAt: now, UpdatedAt: now, Metadata: metadata}
		s.registry.put(info)
		return info, nil
	}
	return info, nil
}

// FileCount returns the number of indexed files.
func (s *Server) FileCount() int {
	return s.registry.count()
}

// ListPeers returns connected peer information for APIs.
func (s *Server) ListPeers() []PeerInfo {
	now := time.Now().UTC()
	peers := s.listPeers()
	out := make([]PeerInfo, 0, len(peers)+1)
	out = append(out, PeerInfo{
		ID:       s.ID,
		Address:  s.Transport.Addr(),
		Status:   "self",
		LastSeen: now,
	})
	for _, p := range peers {
		addr := p.RemoteAddr().String()
		out = append(out, PeerInfo{
			ID:       addr,
			Address:  addr,
			Status:   "active",
			LastSeen: now,
		})
	}
	return out
}

// PeerCount returns connected remote peers (excluding self).
func (s *Server) PeerCount() int {
	return len(s.listPeers())
}

// AddPeer dials and registers a peer at host:port or host.
func (s *Server) AddPeer(ctx context.Context, address string, port int) (*PeerInfo, error) {
	_ = ctx
	addr := address
	if port > 0 && !strings.Contains(address, ":") {
		addr = net.JoinHostPort(address, strconv.Itoa(port))
	}
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(addr, "3000")
	}
	p, err := s.dialPeer(addr)
	if err != nil {
		return nil, err
	}
	return &PeerInfo{
		ID:       p.RemoteAddr().String(),
		Address:  p.RemoteAddr().String(),
		Status:   "active",
		LastSeen: time.Now().UTC(),
	}, nil
}

// RemovePeer disconnects a peer by address/id.
func (s *Server) RemovePeer(peerID string) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	p, ok := s.peers[peerID]
	if !ok {
		// Try matching by suffix / alternate form
		for addr, peer := range s.peers {
			if addr == peerID || strings.HasSuffix(addr, peerID) {
				p = peer
				ok = true
				peerID = addr
				break
			}
		}
	}
	if !ok {
		return fmt.Errorf("peer not found: %s", peerID)
	}
	_ = p.Close()
	delete(s.peers, peerID)
	if s.healthManager != nil {
		s.healthManager.RemovePeer(peerID)
	}
	if s.resourceManager != nil {
		s.resourceManager.RemovePeer(peerID)
	}
	return nil
}

// Uptime returns time since this server was constructed.
func (s *Server) Uptime() time.Duration {
	return time.Since(s.startTime)
}

// StartTime returns when this server instance was created.
func (s *Server) StartTime() time.Time {
	return s.startTime
}

// ListenAddr returns the transport listen address.
func (s *Server) ListenAddr() string {
	if s.Transport == nil {
		return ""
	}
	return s.Transport.Addr()
}

// replicateToPeers streams the ciphertext already on disk to each connected peer.
func (s *Server) replicateToPeers(sk string) error {
	peers := s.listPeers()
	if len(peers) == 0 {
		return nil
	}

	var firstErr error
	for _, p := range peers {
		if err := s.sendFileStream(p, sk); err != nil {
			slog.Warn("failed to replicate to peer", "peer", p.RemoteAddr(), "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// sendFileStream writes a framed stream: [stream header][keyLen][key][size][ciphertext]
func (s *Server) sendFileStream(p netp2p.Peer, sk string) error {
	size, reader, err := s.store.Read(sk)
	if err != nil {
		return err
	}
	defer func() {
		if closer, ok := reader.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	fw := netp2p.NewFrameWriter(p)
	if err := fw.WriteStreamHeader(); err != nil {
		return fmt.Errorf("write stream header: %w", err)
	}

	keyBytes := []byte(sk)
	if err := binary.Write(p, binary.LittleEndian, uint32(len(keyBytes))); err != nil {
		return fmt.Errorf("write key length: %w", err)
	}
	if _, err := p.Write(keyBytes); err != nil {
		return fmt.Errorf("write key: %w", err)
	}
	if err := binary.Write(p, binary.LittleEndian, size); err != nil {
		return fmt.Errorf("write size: %w", err)
	}

	n, err := io.Copy(p, io.LimitReader(reader, size))
	if err != nil {
		return fmt.Errorf("stream file data: %w", err)
	}
	slog.Info("streamed file to peer", "peer", p.RemoteAddr(), "bytes", n, "key", sk)
	return nil
}

func (s *Server) Stop() {
	if s.healthManager != nil {
		s.healthManager.Stop()
	}
	if s.resourceManager != nil {
		s.resourceManager.Shutdown()
	}

	select {
	case <-s.quitch:
	default:
		close(s.quitch)
	}
}

func (s *Server) OnPeer(p netp2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p

	if s.healthManager != nil {
		s.healthManager.AddPeer(p)
	}
	if s.resourceManager != nil {
		s.resourceManager.AddPeer(p.RemoteAddr().String())
	}

	slog.Info("connected", "peer", p.RemoteAddr())
	return nil
}

// OnStream handles incoming file streams: [keyLen][key][size][ciphertext]
func (s *Server) OnStream(p netp2p.Peer, reader io.Reader) error {
	var keyLen uint32
	if err := binary.Read(reader, binary.LittleEndian, &keyLen); err != nil {
		return fmt.Errorf("failed to read key length: %w", err)
	}
	if keyLen == 0 || keyLen > 1024 {
		return fmt.Errorf("invalid key length: %d", keyLen)
	}

	keyBytes := make([]byte, keyLen)
	if _, err := io.ReadFull(reader, keyBytes); err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}
	sk := string(keyBytes)

	var size int64
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return fmt.Errorf("failed to read size: %w", err)
	}
	if size < 0 {
		return fmt.Errorf("invalid stream size: %d", size)
	}

	slog.Info("receiving file stream",
		slog.String("key", sk),
		slog.Int64("size", size),
		slog.String("peer", p.RemoteAddr().String()))

	limited := io.LimitReader(reader, size)
	ciphertext, err := io.ReadAll(limited)
	if err != nil {
		s.fulfillPendingGet(sk, nil, err)
		return fmt.Errorf("failed to read stream payload: %w", err)
	}

	// Persist ciphertext as-is (already encrypted by the origin).
	if !s.store.Has(sk) {
		if _, err := s.store.Write(sk, bytes.NewReader(ciphertext)); err != nil {
			s.fulfillPendingGet(sk, nil, err)
			return fmt.Errorf("failed to store streamed file: %w", err)
		}
		slog.Info("stored streamed file",
			slog.String("key", sk),
			slog.Int("bytes", len(ciphertext)),
			slog.String("peer", p.RemoteAddr().String()))
	}

	// Satisfy a pending Get with decrypted plaintext when requested.
	s.pendingMu.Lock()
	_, waiting := s.pendingGets[sk]
	s.pendingMu.Unlock()
	if waiting {
		var plain bytes.Buffer
		if _, err := crypto.CopyDecrypt(s.getEncryptionKey(), bytes.NewReader(ciphertext), &plain); err != nil {
			s.fulfillPendingGet(sk, nil, fmt.Errorf("failed to decrypt streamed file: %w", err))
			return err
		}
		s.fulfillPendingGet(sk, plain.Bytes(), nil)
	}

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
				continue
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
	case dto.StoreFileAck, dto.GetFileAck:
		return nil
	}
	return nil
}

func (s *Server) getPeer(from string) (netp2p.Peer, bool) {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	p, ok := s.peers[from]
	return p, ok
}

func (s *Server) handleMessageGetFile(from string, msg dto.GetFile) error {
	hasFile := s.store.Has(msg.Key)
	var fileSize int64
	if hasFile {
		size, _, err := s.store.Read(msg.Key)
		if err == nil {
			fileSize = size
		}
	}

	ack := dto.GetFileAck{
		RequestID: msg.ID,
		Key:       msg.Key,
		HasFile:   hasFile,
		FileSize:  fileSize,
	}
	ackMsg := Message{Payload: ack}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(ackMsg); err != nil {
		return err
	}

	p, ok := s.getPeer(from)
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	fw := netp2p.NewFrameWriter(p)
	if err := fw.WriteMessage(buf.Bytes()); err != nil {
		return err
	}

	if !hasFile {
		return nil
	}

	slog.Info("serving file to peer", "key", msg.Key, "peer", from)
	return s.sendFileStream(p, msg.Key)
}

func (s *Server) handleMessageStoreFile(from string, msg dto.StoreFile) error {
	// Announcement only — payload arrives via OnStream replication.
	ack := dto.StoreFileAck{
		RequestID: msg.ID,
		Key:       msg.Key,
		Success:   true,
	}
	ackMsg := Message{Payload: ack}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(ackMsg); err != nil {
		return err
	}

	p, ok := s.getPeer(from)
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}
	fw := netp2p.NewFrameWriter(p)
	return fw.WriteMessage(buf.Bytes())
}

func (s *Server) BootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			slog.Info("attempting to connect", "addr", addr, "current_addr", s.Transport.Addr())
			if err := s.Transport.Dial(addr); err != nil {
				slog.Error("dial error", "err", err)
			}
		}(addr)
	}
	return nil
}

func (s *Server) Start() error {
	slog.Info("starting fileserver", "addr", s.Transport.Addr())

	if s.healthManager != nil {
		s.healthManager.Start()
	}

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	if err := s.BootstrapNetwork(); err != nil {
		slog.Error("failed to bootstrap network", "err", err)
	}

	go s.loop()
	return nil
}

func init() {
	gob.Register(dto.StoreFile{})
	gob.Register(dto.GetFile{})
	gob.Register(dto.StoreFileAck{})
	gob.Register(dto.GetFileAck{})
}

// FileOperationManager manages concurrent file operations
type FileOperationManager struct {
	semaphore  chan struct{}
	mu         sync.RWMutex
	operations map[string]*FileOperation
}

// FileOperation represents a file operation in progress
type FileOperation struct {
	Key       string
	Type      string // "store", "get", "delete"
	StartTime time.Time
	Status    string // "running", "completed", "failed"
	Error     error
}

// NewFileOperationManager creates a new file operation manager
func NewFileOperationManager(maxConcurrentOps int) *FileOperationManager {
	if maxConcurrentOps <= 0 {
		maxConcurrentOps = 10
	}

	return &FileOperationManager{
		semaphore:  make(chan struct{}, maxConcurrentOps),
		operations: make(map[string]*FileOperation),
	}
}

// AcquireOperation acquires a slot for a file operation
func (fom *FileOperationManager) AcquireOperation(key, opType string) (*FileOperation, error) {
	select {
	case fom.semaphore <- struct{}{}:
		op := &FileOperation{
			Key:       key,
			Type:      opType,
			StartTime: time.Now(),
			Status:    "running",
		}

		fom.mu.Lock()
		fom.operations[key] = op
		fom.mu.Unlock()

		return op, nil
	default:
		return nil, fmt.Errorf("too many concurrent file operations")
	}
}

// ReleaseOperation releases a file operation slot
func (fom *FileOperationManager) ReleaseOperation(key string, err error) {
	fom.mu.Lock()
	if op, exists := fom.operations[key]; exists {
		if err != nil {
			op.Status = "failed"
			op.Error = err
		} else {
			op.Status = "completed"
		}
		delete(fom.operations, key)
	}
	fom.mu.Unlock()

	<-fom.semaphore
}

// GetActiveOperations returns the number of active operations
func (fom *FileOperationManager) GetActiveOperations() int {
	fom.mu.RLock()
	defer fom.mu.RUnlock()
	return len(fom.operations)
}
