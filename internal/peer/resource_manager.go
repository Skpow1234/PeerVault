package peer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ResourceLimits defines the resource constraints for peer operations
type ResourceLimits struct {
	MaxConcurrentStreams int           // Maximum concurrent streams per peer
	StreamTimeout        time.Duration // Timeout for stream operations
	RateLimit            rate.Limit    // Rate limit for operations per second
	BurstLimit           int           // Burst limit for rate limiting
}

// DefaultResourceLimits returns sensible default resource limits
func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxConcurrentStreams: 3,
		StreamTimeout:        5 * time.Minute,
		RateLimit:            rate.Limit(10), // 10 operations per second
		BurstLimit:           5,
	}
}

// StreamTracker tracks active streams for a peer
type StreamTracker struct {
	activeStreams map[string]context.CancelFunc
	mu            sync.RWMutex
	limiter       *rate.Limiter
	limits        ResourceLimits
}

// NewStreamTracker creates a new stream tracker for a peer
func NewStreamTracker(limits ResourceLimits) *StreamTracker {
	return &StreamTracker{
		activeStreams: make(map[string]context.CancelFunc),
		limiter:       rate.NewLimiter(limits.RateLimit, limits.BurstLimit),
		limits:        limits,
	}
}

// AcquireStream attempts to acquire a stream slot with context support
func (st *StreamTracker) AcquireStream(ctx context.Context, streamID string) (context.Context, error) {
	// Check rate limit
	if !st.limiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded for stream %s", streamID)
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	// Check concurrent stream limit
	if len(st.activeStreams) >= st.limits.MaxConcurrentStreams {
		return nil, fmt.Errorf("concurrent stream limit exceeded (%d/%d)",
			len(st.activeStreams), st.limits.MaxConcurrentStreams)
	}

	// Create a new context with timeout
	streamCtx, cancel := context.WithTimeout(ctx, st.limits.StreamTimeout)

	// Store the cancel function
	st.activeStreams[streamID] = cancel

	slog.Debug("stream acquired",
		"stream_id", streamID,
		"active_streams", len(st.activeStreams),
		"max_streams", st.limits.MaxConcurrentStreams)

	return streamCtx, nil
}

// ReleaseStream releases a stream slot
func (st *StreamTracker) ReleaseStream(streamID string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if cancel, exists := st.activeStreams[streamID]; exists {
		cancel()
		delete(st.activeStreams, streamID)
		slog.Debug("stream released",
			"stream_id", streamID,
			"active_streams", len(st.activeStreams))
	}
}

// CancelAllStreams cancels all active streams
func (st *StreamTracker) CancelAllStreams() {
	st.mu.Lock()
	defer st.mu.Unlock()

	for streamID, cancel := range st.activeStreams {
		cancel()
		slog.Debug("stream cancelled", "stream_id", streamID)
	}
	st.activeStreams = make(map[string]context.CancelFunc)
}

// GetActiveStreamCount returns the number of active streams
func (st *StreamTracker) GetActiveStreamCount() int {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return len(st.activeStreams)
}

// ResourceManager manages resource limits across all peers
type ResourceManager struct {
	peerTrackers map[string]*StreamTracker
	mu           sync.RWMutex
	limits       ResourceLimits
}

// NewResourceManager creates a new resource manager
func NewResourceManager(limits ResourceLimits) *ResourceManager {
	return &ResourceManager{
		peerTrackers: make(map[string]*StreamTracker),
		limits:       limits,
	}
}

// AddPeer adds a peer to resource management
func (rm *ResourceManager) AddPeer(peerAddress string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.peerTrackers[peerAddress]; !exists {
		rm.peerTrackers[peerAddress] = NewStreamTracker(rm.limits)
		slog.Info("peer added to resource management", "address", peerAddress)
	}
}

// RemovePeer removes a peer from resource management
func (rm *ResourceManager) RemovePeer(peerAddress string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if tracker, exists := rm.peerTrackers[peerAddress]; exists {
		tracker.CancelAllStreams()
		delete(rm.peerTrackers, peerAddress)
		slog.Info("peer removed from resource management", "address", peerAddress)
	}
}

// AcquireStreamForPeer attempts to acquire a stream for a specific peer
func (rm *ResourceManager) AcquireStreamForPeer(ctx context.Context, peerAddress, streamID string) (context.Context, error) {
	rm.mu.RLock()
	tracker, exists := rm.peerTrackers[peerAddress]
	rm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("peer %s not found in resource manager", peerAddress)
	}

	return tracker.AcquireStream(ctx, streamID)
}

// ReleaseStreamForPeer releases a stream for a specific peer
func (rm *ResourceManager) ReleaseStreamForPeer(peerAddress, streamID string) {
	rm.mu.RLock()
	tracker, exists := rm.peerTrackers[peerAddress]
	rm.mu.RUnlock()

	if exists {
		tracker.ReleaseStream(streamID)
	}
}

// GetPeerStats returns statistics for a peer
func (rm *ResourceManager) GetPeerStats(peerAddress string) (int, error) {
	rm.mu.RLock()
	tracker, exists := rm.peerTrackers[peerAddress]
	rm.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("peer %s not found in resource manager", peerAddress)
	}

	return tracker.GetActiveStreamCount(), nil
}

// GetAllStats returns statistics for all peers
func (rm *ResourceManager) GetAllStats() map[string]int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := make(map[string]int)
	for address, tracker := range rm.peerTrackers {
		stats[address] = tracker.GetActiveStreamCount()
	}
	return stats
}

// Shutdown cancels all streams and cleans up resources
func (rm *ResourceManager) Shutdown() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for address, tracker := range rm.peerTrackers {
		tracker.CancelAllStreams()
		slog.Info("cancelled all streams for peer", "address", address)
	}
	rm.peerTrackers = make(map[string]*StreamTracker)
}
