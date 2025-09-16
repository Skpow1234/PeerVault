package ratelimit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest/versioning"
)

func TestTokenBucketAlgorithm(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       TokenBucket,
		RequestsPerMin:  120, // 2 requests per second for faster testing
		BurstSize:       5,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0

	// Should allow burst requests
	for i := 0; i < 5; i++ {
		if !rl.IsAllowed(req, version) {
			t.Errorf("Request %d should be allowed in burst", i+1)
		}
	}

	// Should deny further requests
	if rl.IsAllowed(req, version) {
		t.Error("Request should be denied after burst")
	}

	// Wait for token refill (need about 0.5 seconds for 1 token at 120 requests/minute)
	time.Sleep(600 * time.Millisecond)

	// Should allow some requests after refill
	if !rl.IsAllowed(req, version) {
		t.Error("Request should be allowed after token refill")
	}
}

func TestSlidingWindowAlgorithm(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       SlidingWindow,
		RequestsPerMin:  5,
		WindowSize:      time.Minute,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0

	// Should allow up to limit
	for i := 0; i < 5; i++ {
		if !rl.IsAllowed(req, version) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Should deny further requests
	if rl.IsAllowed(req, version) {
		t.Error("Request should be denied after limit")
	}
}

func TestLeakyBucketAlgorithm(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       LeakyBucket,
		RequestsPerMin:  60, // 1 request per second for faster testing
		BurstSize:       3,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0

	// Should allow burst requests
	for i := 0; i < 3; i++ {
		if !rl.IsAllowed(req, version) {
			t.Errorf("Request %d should be allowed in burst", i+1)
		}
	}

	// Should deny further requests (burst limit exceeded)
	if rl.IsAllowed(req, version) {
		t.Error("Request should be denied after burst")
	}

	// Wait for leak (at 60 requests/minute = 1 request/second, wait 1.1 seconds)
	time.Sleep(1100 * time.Millisecond)

	// Should allow requests after leak
	if !rl.IsAllowed(req, version) {
		t.Error("Request should be allowed after leak")
	}
}

func TestAbuseDetection(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       TokenBucket,
		RequestsPerMin:  1,
		BurstSize:       1,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0

	// Make one allowed request
	if !rl.IsAllowed(req, version) {
		t.Error("First request should be allowed")
	}

	// Make several denied requests to trigger abuse detection
	for i := 0; i < 6; i++ {
		rl.IsAllowed(req, version) // This will be denied
	}

	// Check if client is banned
	state := rl.GetClientState(req, version)
	if state == nil {
		t.Fatal("Client state should exist")
		return // This return is unreachable due to t.Fatal, but satisfies staticcheck
	}

	if time.Now().Before(state.BannedUntil) {
		t.Log("Client correctly banned due to abuse")
	} else {
		t.Error("Client should be banned due to abuse")
	}
}

func TestDisabledRateLimiting(t *testing.T) {
	config := &RateLimitConfig{
		Enabled: false,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0

	// Should always allow when disabled
	for i := 0; i < 100; i++ {
		if !rl.IsAllowed(req, version) {
			t.Errorf("Request %d should be allowed when rate limiting is disabled", i+1)
		}
	}
}

func TestMiddleware(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       TokenBucket,
		RequestsPerMin:  1,
		BurstSize:       1,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	middleware := rl.Middleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	// First request should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0
	ctx := context.WithValue(req.Context(), versioning.VersionContextKey{}, version)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Second request should be rate limited
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	// Check headers
	if w.Header().Get("X-Rate-Limit-Exceeded") != "true" {
		t.Error("X-Rate-Limit-Exceeded header should be set")
	}
	if w.Header().Get("Retry-After") != "60" {
		t.Error("Retry-After header should be set to 60")
	}
}

func TestStats(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       TokenBucket,
		RequestsPerMin:  10,
		BurstSize:       5,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	// Make some requests
	req := httptest.NewRequest("GET", "/test", nil)
	version := versioning.Version_1_0_0

	for i := 0; i < 3; i++ {
		rl.IsAllowed(req, version)
	}

	stats := rl.GetStats()

	if stats["algorithm"] != "token_bucket" {
		t.Errorf("Expected algorithm token_bucket, got %v", stats["algorithm"])
	}
	if stats["total_clients"] != 1 {
		t.Errorf("Expected 1 total client, got %v", stats["total_clients"])
	}
	if stats["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", stats["enabled"])
	}
}

func TestVersionSpecificRateLimiting(t *testing.T) {
	config := &RateLimitConfig{
		Algorithm:       TokenBucket,
		RequestsPerMin:  2,
		BurstSize:       2,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}

	rl := NewRateLimiter(config)
	defer rl.Stop()

	req := httptest.NewRequest("GET", "/test", nil)

	// Make requests with different versions
	v1 := versioning.Version_1_0_0
	v2 := versioning.Version_1_1_0

	// Should allow requests for both versions independently
	for i := 0; i < 2; i++ {
		if !rl.IsAllowed(req, v1) {
			t.Errorf("Request %d for v1.0.0 should be allowed", i+1)
		}
		if !rl.IsAllowed(req, v2) {
			t.Errorf("Request %d for v1.1.0 should be allowed", i+1)
		}
	}

	// Both versions should be rate limited independently
	if rl.IsAllowed(req, v1) {
		t.Error("v1.0.0 should be rate limited")
	}
	if rl.IsAllowed(req, v2) {
		t.Error("v1.1.0 should be rate limited")
	}
}
