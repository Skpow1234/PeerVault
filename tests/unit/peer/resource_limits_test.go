package main

import (
	"context"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/peer"
)

func TestResourceLimits(t *testing.T) {
	// Test default resource limits
	limits := peer.DefaultResourceLimits()
	if limits.MaxConcurrentStreams != 3 {
		t.Errorf("Expected MaxConcurrentStreams to be 3, got %d", limits.MaxConcurrentStreams)
	}
	if limits.StreamTimeout != 5*time.Minute {
		t.Errorf("Expected StreamTimeout to be 5 minutes, got %v", limits.StreamTimeout)
	}
}

func TestStreamTracker(t *testing.T) {
	limits := peer.ResourceLimits{
		MaxConcurrentStreams: 2,
		StreamTimeout:        1 * time.Second,
		RateLimit:            10,
		BurstLimit:           5,
	}

	tracker := peer.NewStreamTracker(limits)

	// Test acquiring streams within limit
	ctx := context.Background()
	stream1, err := tracker.AcquireStream(ctx, "stream1")
	if err != nil {
		t.Fatalf("Failed to acquire stream1: %v", err)
	}
	defer tracker.ReleaseStream("stream1")

	// Check that stream1 context is not cancelled
	select {
	case <-stream1.Done():
		t.Error("Stream1 context should not be cancelled")
	default:
		// Expected
	}

	stream2, err := tracker.AcquireStream(ctx, "stream2")
	if err != nil {
		t.Fatalf("Failed to acquire stream2: %v", err)
	}
	defer tracker.ReleaseStream("stream2")

	// Check that stream2 context is not cancelled
	select {
	case <-stream2.Done():
		t.Error("Stream2 context should not be cancelled")
	default:
		// Expected
	}

	// Test exceeding concurrent stream limit
	_, err = tracker.AcquireStream(ctx, "stream3")
	if err == nil {
		t.Error("Expected error when exceeding concurrent stream limit")
	}

	// Test stream timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = tracker.AcquireStream(timeoutCtx, "stream4")
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}

	// Test releasing streams
	tracker.ReleaseStream("stream1")
	tracker.ReleaseStream("stream2")

	// Should be able to acquire streams again after release
	_, err = tracker.AcquireStream(ctx, "stream5")
	if err != nil {
		t.Fatalf("Failed to acquire stream after release: %v", err)
	}
	defer tracker.ReleaseStream("stream5")
}

func TestResourceManager(t *testing.T) {
	limits := peer.ResourceLimits{
		MaxConcurrentStreams: 2,
		StreamTimeout:        1 * time.Second,
		RateLimit:            10,
		BurstLimit:           5,
	}

	manager := peer.NewResourceManager(limits)

	// Test adding peers
	manager.AddPeer("peer1")
	manager.AddPeer("peer2")

	// Test acquiring streams for peers
	ctx := context.Background()
	stream1, err := manager.AcquireStreamForPeer(ctx, "peer1", "stream1")
	if err != nil {
		t.Fatalf("Failed to acquire stream for peer1: %v", err)
	}
	defer manager.ReleaseStreamForPeer("peer1", "stream1")

	// Check that stream1 context is not cancelled
	select {
	case <-stream1.Done():
		t.Error("Stream1 context should not be cancelled")
	default:
		// Expected
	}

	stream2, err := manager.AcquireStreamForPeer(ctx, "peer2", "stream2")
	if err != nil {
		t.Fatalf("Failed to acquire stream for peer2: %v", err)
	}
	defer manager.ReleaseStreamForPeer("peer2", "stream2")

	// Check that stream2 context is not cancelled
	select {
	case <-stream2.Done():
		t.Error("Stream2 context should not be cancelled")
	default:
		// Expected
	}

	// Test getting stats
	stats := manager.GetAllStats()
	if stats["peer1"] != 1 {
		t.Errorf("Expected peer1 to have 1 active stream, got %d", stats["peer1"])
	}
	if stats["peer2"] != 1 {
		t.Errorf("Expected peer2 to have 1 active stream, got %d", stats["peer2"])
	}

	// Test removing peer
	manager.RemovePeer("peer1")

	// Should not be able to acquire stream for removed peer
	_, err = manager.AcquireStreamForPeer(ctx, "peer1", "stream3")
	if err == nil {
		t.Error("Expected error when acquiring stream for removed peer")
	}

	// Test shutdown
	manager.Shutdown()

	// Should not be able to acquire streams after shutdown
	_, err = manager.AcquireStreamForPeer(ctx, "peer2", "stream4")
	if err == nil {
		t.Error("Expected error when acquiring stream after shutdown")
	}
}

func TestContextCancellation(t *testing.T) {
	limits := peer.DefaultResourceLimits()
	tracker := peer.NewStreamTracker(limits)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Acquire a stream
	streamCtx, err := tracker.AcquireStream(ctx, "test-stream")
	if err != nil {
		t.Fatalf("Failed to acquire stream: %v", err)
	}
	defer tracker.ReleaseStream("test-stream")

	// Cancel the parent context
	cancel()

	// The stream context should be cancelled
	select {
	case <-streamCtx.Done():
		// Expected
	default:
		t.Error("Stream context should be cancelled when parent is cancelled")
	}
}

func TestRateLimiting(t *testing.T) {
	limits := peer.ResourceLimits{
		MaxConcurrentStreams: 10,
		StreamTimeout:        1 * time.Second,
		RateLimit:            2, // 2 operations per second
		BurstLimit:           1,
	}

	tracker := peer.NewStreamTracker(limits)
	ctx := context.Background()

	// First stream should succeed
	_, err := tracker.AcquireStream(ctx, "stream1")
	if err != nil {
		t.Fatalf("Failed to acquire first stream: %v", err)
	}
	defer tracker.ReleaseStream("stream1")

	// Second stream should fail due to rate limiting
	_, err = tracker.AcquireStream(ctx, "stream2")
	if err == nil {
		t.Error("Expected rate limit error for second stream")
	}

	// Wait for rate limit to reset
	time.Sleep(600 * time.Millisecond)

	// Should be able to acquire stream again
	_, err = tracker.AcquireStream(ctx, "stream3")
	if err != nil {
		t.Fatalf("Failed to acquire stream after rate limit reset: %v", err)
	}
	defer tracker.ReleaseStream("stream3")
}
