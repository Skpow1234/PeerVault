package peer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	netp2p "github.com/Skpow1234/Peervault/internal/transport/p2p"
)

// HealthStatus represents the health status of a peer
type HealthStatus int

const (
	StatusHealthy HealthStatus = iota
	StatusUnhealthy
	StatusDisconnected
)

// PeerInfo contains information about a peer including health status
type PeerInfo struct {
	Peer                 netp2p.Peer
	Address              string
	Status               HealthStatus
	LastSeen             time.Time
	LastHeartbeat        time.Time
	ReconnectAttempts    int
	MaxReconnectAttempts int
	ReconnectBackoff     time.Duration
	mu                   sync.RWMutex
}

// HealthManager manages peer health monitoring and reconnection
type HealthManager struct {
	peers                map[string]*PeerInfo
	mu                   sync.RWMutex
	heartbeatInterval    time.Duration
	healthTimeout        time.Duration
	reconnectInterval    time.Duration
	maxReconnectAttempts int
	ctx                  context.Context
	cancel               context.CancelFunc
	wg                   sync.WaitGroup
	onPeerDisconnect     func(string)
	onPeerReconnect      func(string, netp2p.Peer)
	dialFunc             func(string) (netp2p.Peer, error)
}

// NewHealthManager creates a new peer health manager
func NewHealthManager(opts HealthManagerOpts) *HealthManager {
	if opts.HeartbeatInterval == 0 {
		opts.HeartbeatInterval = 30 * time.Second
	}
	if opts.HealthTimeout == 0 {
		opts.HealthTimeout = 90 * time.Second
	}
	if opts.ReconnectInterval == 0 {
		opts.ReconnectInterval = 5 * time.Second
	}
	if opts.MaxReconnectAttempts == 0 {
		opts.MaxReconnectAttempts = 5
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &HealthManager{
		peers:                make(map[string]*PeerInfo),
		heartbeatInterval:    opts.HeartbeatInterval,
		healthTimeout:        opts.HealthTimeout,
		reconnectInterval:    opts.ReconnectInterval,
		maxReconnectAttempts: opts.MaxReconnectAttempts,
		ctx:                  ctx,
		cancel:               cancel,
		onPeerDisconnect:     opts.OnPeerDisconnect,
		onPeerReconnect:      opts.OnPeerReconnect,
		dialFunc:             opts.DialFunc,
	}
}

// HealthManagerOpts contains configuration options for the health manager
type HealthManagerOpts struct {
	HeartbeatInterval    time.Duration
	HealthTimeout        time.Duration
	ReconnectInterval    time.Duration
	MaxReconnectAttempts int
	OnPeerDisconnect     func(string)
	OnPeerReconnect      func(string, netp2p.Peer)
	DialFunc             func(string) (netp2p.Peer, error)
}

// AddPeer adds a new peer to health monitoring
func (hm *HealthManager) AddPeer(peer netp2p.Peer) {
	address := peer.RemoteAddr().String()

	hm.mu.Lock()
	defer hm.mu.Unlock()

	peerInfo := &PeerInfo{
		Peer:                 peer,
		Address:              address,
		Status:               StatusHealthy,
		LastSeen:             time.Now(),
		LastHeartbeat:        time.Now(),
		ReconnectAttempts:    0,
		MaxReconnectAttempts: hm.maxReconnectAttempts,
		ReconnectBackoff:     hm.reconnectInterval,
	}

	hm.peers[address] = peerInfo
	slog.Info("peer added to health monitoring", "address", address)
}

// RemovePeer removes a peer from health monitoring
func (hm *HealthManager) RemovePeer(address string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if _, exists := hm.peers[address]; exists {
		delete(hm.peers, address)
		slog.Info("peer removed from health monitoring", "address", address)
	}
}

// UpdatePeerHealth updates the health status of a peer
func (hm *HealthManager) UpdatePeerHealth(address string, status HealthStatus) {
	hm.mu.RLock()
	peerInfo, exists := hm.peers[address]
	hm.mu.RUnlock()

	if !exists {
		return
	}

	peerInfo.mu.Lock()
	defer peerInfo.mu.Unlock()

	oldStatus := peerInfo.Status
	peerInfo.Status = status
	peerInfo.LastSeen = time.Now()

	if status == StatusHealthy {
		peerInfo.LastHeartbeat = time.Now()
		peerInfo.ReconnectAttempts = 0
		peerInfo.ReconnectBackoff = hm.reconnectInterval
	}

	if oldStatus != status {
		slog.Info("peer health status changed",
			"address", address,
			"old_status", oldStatus,
			"new_status", status)

		if status == StatusDisconnected && hm.onPeerDisconnect != nil {
			hm.onPeerDisconnect(address)
		}
	}
}

// GetHealthyPeers returns a list of healthy peers
func (hm *HealthManager) GetHealthyPeers() []netp2p.Peer {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var healthyPeers []netp2p.Peer
	for _, peerInfo := range hm.peers {
		peerInfo.mu.RLock()
		if peerInfo.Status == StatusHealthy {
			healthyPeers = append(healthyPeers, peerInfo.Peer)
		}
		peerInfo.mu.RUnlock()
	}

	return healthyPeers
}

// GetPeerStatus returns the health status of a specific peer
func (hm *HealthManager) GetPeerStatus(address string) (HealthStatus, bool) {
	hm.mu.RLock()
	peerInfo, exists := hm.peers[address]
	hm.mu.RUnlock()

	if !exists {
		return StatusDisconnected, false
	}

	peerInfo.mu.RLock()
	defer peerInfo.mu.RUnlock()

	return peerInfo.Status, true
}

// Start begins the health monitoring process
func (hm *HealthManager) Start() {
	slog.Info("starting peer health monitoring")

	hm.wg.Add(2)
	go hm.heartbeatLoop()
	go hm.healthCheckLoop()
}

// Stop stops the health monitoring process
func (hm *HealthManager) Stop() {
	slog.Info("stopping peer health monitoring")
	hm.cancel()
	hm.wg.Wait()
}

// heartbeatLoop sends periodic heartbeats to all peers
func (hm *HealthManager) heartbeatLoop() {
	defer hm.wg.Done()

	ticker := time.NewTicker(hm.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.sendHeartbeats()
		}
	}
}

// healthCheckLoop checks peer health and handles reconnection
func (hm *HealthManager) healthCheckLoop() {
	defer hm.wg.Done()

	ticker := time.NewTicker(hm.healthTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.checkPeerHealth()
		}
	}
}

// sendHeartbeats sends heartbeat messages to all peers
func (hm *HealthManager) sendHeartbeats() {
	hm.mu.RLock()
	peers := make([]*PeerInfo, 0, len(hm.peers))
	for _, peerInfo := range hm.peers {
		peers = append(peers, peerInfo)
	}
	hm.mu.RUnlock()

	for _, peerInfo := range peers {
		peerInfo.mu.RLock()
		if peerInfo.Status == StatusHealthy {
			peer := peerInfo.Peer
			address := peerInfo.Address
			peerInfo.mu.RUnlock()

			// Send heartbeat message
			heartbeat := &HeartbeatMessage{
				Timestamp: time.Now().Unix(),
				NodeID:    "heartbeat", // This should be the actual node ID
			}

			if err := sendHeartbeat(peer, heartbeat); err != nil {
				slog.Warn("failed to send heartbeat", "address", address, "error", err)
				hm.UpdatePeerHealth(address, StatusUnhealthy)
			} else {
				hm.UpdatePeerHealth(address, StatusHealthy)
			}
		} else {
			peerInfo.mu.RUnlock()
		}
	}
}

// checkPeerHealth checks the health of all peers and handles reconnection
func (hm *HealthManager) checkPeerHealth() {
	hm.mu.RLock()
	peers := make([]*PeerInfo, 0, len(hm.peers))
	for _, peerInfo := range hm.peers {
		peers = append(peers, peerInfo)
	}
	hm.mu.RUnlock()

	for _, peerInfo := range peers {
		peerInfo.mu.Lock()

		// Check if peer is healthy
		if peerInfo.Status == StatusHealthy {
			// Check if peer has responded recently
			if time.Since(peerInfo.LastHeartbeat) > hm.healthTimeout {
				slog.Warn("peer heartbeat timeout", "address", peerInfo.Address)
				peerInfo.Status = StatusUnhealthy
			}
		}

		// Handle unhealthy peers
		if peerInfo.Status == StatusUnhealthy {
			if peerInfo.ReconnectAttempts < peerInfo.MaxReconnectAttempts {
				// Try to reconnect
				go hm.attemptReconnect(peerInfo)
			} else {
				// Mark as disconnected after max attempts
				peerInfo.Status = StatusDisconnected
				slog.Error("peer marked as disconnected after max reconnect attempts",
					"address", peerInfo.Address,
					"attempts", peerInfo.ReconnectAttempts)
			}
		}

		peerInfo.mu.Unlock()
	}
}

// attemptReconnect attempts to reconnect to an unhealthy peer
func (hm *HealthManager) attemptReconnect(peerInfo *PeerInfo) {
	peerInfo.mu.Lock()
	address := peerInfo.Address
	attempts := peerInfo.ReconnectAttempts
	backoff := peerInfo.ReconnectBackoff
	peerInfo.ReconnectAttempts++
	peerInfo.ReconnectBackoff = time.Duration(float64(backoff) * 1.5) // Exponential backoff
	peerInfo.mu.Unlock()

	slog.Info("attempting to reconnect to peer",
		"address", address,
		"attempt", attempts+1)

	// Wait for backoff period
	time.Sleep(backoff)

	if hm.dialFunc != nil {
		if newPeer, err := hm.dialFunc(address); err == nil {
			// Reconnection successful
			peerInfo.mu.Lock()
			peerInfo.Peer = newPeer
			peerInfo.Status = StatusHealthy
			peerInfo.LastSeen = time.Now()
			peerInfo.LastHeartbeat = time.Now()
			peerInfo.ReconnectAttempts = 0
			peerInfo.ReconnectBackoff = hm.reconnectInterval
			peerInfo.mu.Unlock()

			slog.Info("successfully reconnected to peer", "address", address)

			if hm.onPeerReconnect != nil {
				hm.onPeerReconnect(address, newPeer)
			}
		} else {
			slog.Warn("reconnection attempt failed",
				"address", address,
				"attempt", attempts+1,
				"error", err)
		}
	}
}

// HeartbeatMessage represents a heartbeat message
type HeartbeatMessage struct {
	Timestamp int64  `json:"timestamp"`
	NodeID    string `json:"node_id"`
}

// sendHeartbeat sends a heartbeat message to a peer
func sendHeartbeat(peer netp2p.Peer, msg *HeartbeatMessage) error {
	// For now, we'll use a simple ping mechanism
	// In a real implementation, this would use the proper message protocol
	_, err := peer.Write([]byte("PING"))
	return err
}

// String returns a string representation of HealthStatus
func (hs HealthStatus) String() string {
	switch hs {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	case StatusDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}