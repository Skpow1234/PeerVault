package ratelimit

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest/versioning"
)

// Algorithm represents the rate limiting algorithm to use
type Algorithm int

const (
	TokenBucket Algorithm = iota
	SlidingWindow
	LeakyBucket
)

// String returns the string representation of the algorithm
func (a Algorithm) String() string {
	switch a {
	case TokenBucket:
		return "token_bucket"
	case SlidingWindow:
		return "sliding_window"
	case LeakyBucket:
		return "leaky_bucket"
	default:
		return "unknown"
	}
}

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	Algorithm       Algorithm
	RequestsPerMin  int
	BurstSize       int           // For token bucket
	WindowSize      time.Duration // For sliding window
	CleanupInterval time.Duration
	Enabled         bool
}

// DefaultConfig returns a default rate limit configuration
func DefaultConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Algorithm:       TokenBucket,
		RequestsPerMin:  100,
		BurstSize:       20,
		WindowSize:      time.Minute,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	}
}

// ClientState represents the state of a client for rate limiting
type ClientState struct {
	Key       string
	Algorithm Algorithm
	LastSeen  time.Time

	// Token Bucket fields
	Tokens     float64
	LastRefill time.Time

	// Sliding Window fields
	Requests []time.Time

	// Leaky Bucket fields
	WaterLevel float64
	LastLeak   time.Time

	// Abuse detection
	ConsecutiveViolations int
	TotalViolations       int
	BannedUntil           time.Time
}

// IsAllowed checks if the request is allowed based on the current state
func (cs *ClientState) IsAllowed(config *RateLimitConfig, now time.Time) bool {
	// Check if client is banned
	if now.Before(cs.BannedUntil) {
		return false
	}

	switch cs.Algorithm {
	case TokenBucket:
		return cs.isAllowedTokenBucket(config, now)
	case SlidingWindow:
		return cs.isAllowedSlidingWindow(config, now)
	case LeakyBucket:
		return cs.isAllowedLeakyBucket(config, now)
	default:
		return true // Allow by default for unknown algorithms
	}
}

// UpdateState updates the client state after a request
func (cs *ClientState) UpdateState(config *RateLimitConfig, now time.Time, allowed bool) {
	cs.LastSeen = now

	if !allowed {
		cs.ConsecutiveViolations++
		cs.TotalViolations++

		// Implement progressive banning
		if cs.ConsecutiveViolations >= 5 {
			banDuration := time.Duration(cs.ConsecutiveViolations) * time.Minute
			if banDuration > 30*time.Minute {
				banDuration = 30 * time.Minute
			}
			cs.BannedUntil = now.Add(banDuration)
		}
	} else if cs.ConsecutiveViolations > 0 {
		// Reset consecutive violations on successful request
		cs.ConsecutiveViolations--
	}

	switch cs.Algorithm {
	case TokenBucket:
		cs.updateTokenBucket(config, now, allowed)
	case SlidingWindow:
		cs.updateSlidingWindow(config, now, allowed)
	case LeakyBucket:
		cs.updateLeakyBucket(config, now, allowed)
	}
}

// Token Bucket implementation
func (cs *ClientState) isAllowedTokenBucket(config *RateLimitConfig, now time.Time) bool {
	// Refill tokens
	timePassed := now.Sub(cs.LastRefill)
	// Calculate tokens to add: (requests per minute) * (minutes passed)
	tokensToAdd := float64(config.RequestsPerMin) * timePassed.Minutes()
	cs.Tokens = math.Min(float64(config.BurstSize), cs.Tokens+tokensToAdd)
	cs.LastRefill = now

	return cs.Tokens >= 1.0
}

func (cs *ClientState) updateTokenBucket(_ *RateLimitConfig, _ time.Time, allowed bool) {
	if allowed {
		cs.Tokens -= 1.0
	}
}

// Sliding Window implementation
func (cs *ClientState) isAllowedSlidingWindow(config *RateLimitConfig, now time.Time) bool {
	// Remove old requests outside the window
	cutoff := now.Add(-config.WindowSize)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range cs.Requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	cs.Requests = validRequests

	return len(cs.Requests) < config.RequestsPerMin
}

func (cs *ClientState) updateSlidingWindow(_ *RateLimitConfig, now time.Time, allowed bool) {
	if allowed {
		cs.Requests = append(cs.Requests, now)
	}
}

// Leaky Bucket implementation
func (cs *ClientState) isAllowedLeakyBucket(config *RateLimitConfig, now time.Time) bool {
	// Leak water over time
	timePassed := now.Sub(cs.LastLeak)
	leakRate := float64(config.RequestsPerMin) / 60.0 // requests per second
	leaked := leakRate * timePassed.Seconds()
	cs.WaterLevel = math.Max(0, cs.WaterLevel-leaked)
	cs.LastLeak = now

	return cs.WaterLevel+1.0 <= float64(config.BurstSize)
}

func (cs *ClientState) updateLeakyBucket(_ *RateLimitConfig, _ time.Time, allowed bool) {
	if allowed {
		cs.WaterLevel += 1.0
	}
}

// RateLimiter manages rate limiting for multiple clients
type RateLimiter struct {
	config        *RateLimitConfig
	clients       map[string]*ClientState
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:   config,
		clients:  make(map[string]*ClientState),
		stopChan: make(chan struct{}),
	}

	if config.Enabled {
		rl.cleanupTicker = time.NewTicker(config.CleanupInterval)
		go rl.cleanupRoutine()
	}

	return rl
}

// Stop stops the rate limiter and cleanup routine
func (rl *RateLimiter) Stop() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
	close(rl.stopChan)
}

// cleanupRoutine periodically cleans up old client states
func (rl *RateLimiter) cleanupRoutine() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup removes old client states
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Keep clients for 24 hours

	for key, state := range rl.clients {
		if state.LastSeen.Before(cutoff) {
			delete(rl.clients, key)
		}
	}
}

// getClientKey generates a unique key for the client
func (rl *RateLimiter) getClientKey(r *http.Request, version versioning.APIVersion) string {
	// Use IP address as primary key
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	// Include API version in key for version-specific rate limiting
	return fmt.Sprintf("%s:v%s", ip, version.String())
}

// IsAllowed checks if the request is allowed
func (rl *RateLimiter) IsAllowed(r *http.Request, version versioning.APIVersion) bool {
	if !rl.config.Enabled {
		return true
	}

	key := rl.getClientKey(r, version)
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	state, exists := rl.clients[key]
	if !exists {
		// Create new client state
		state = &ClientState{
			Key:       key,
			Algorithm: rl.config.Algorithm,
			LastSeen:  now,
		}

		// Initialize based on algorithm
		switch rl.config.Algorithm {
		case TokenBucket:
			state.Tokens = float64(rl.config.BurstSize)
			state.LastRefill = now
		case SlidingWindow:
			state.Requests = make([]time.Time, 0)
		case LeakyBucket:
			state.WaterLevel = 0
			state.LastLeak = now
		}

		rl.clients[key] = state
	}

	allowed := state.IsAllowed(rl.config, now)
	state.UpdateState(rl.config, now, allowed)

	return allowed
}

// GetClientState returns the state of a client (for debugging/monitoring)
func (rl *RateLimiter) GetClientState(r *http.Request, version versioning.APIVersion) *ClientState {
	key := rl.getClientKey(r, version)

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if state, exists := rl.clients[key]; exists {
		return state
	}
	return nil
}

// GetStats returns rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	totalClients := len(rl.clients)
	activeClients := 0
	bannedClients := 0
	totalViolations := 0

	for _, state := range rl.clients {
		if now.Sub(state.LastSeen) < time.Hour {
			activeClients++
		}
		if now.Before(state.BannedUntil) {
			bannedClients++
		}
		totalViolations += state.TotalViolations
	}

	return map[string]interface{}{
		"algorithm":        rl.config.Algorithm.String(),
		"total_clients":    totalClients,
		"active_clients":   activeClients,
		"banned_clients":   bannedClients,
		"total_violations": totalViolations,
		"enabled":          rl.config.Enabled,
	}
}

// Middleware creates an HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API version from context
			version, ok := versioning.GetVersionFromContext(r.Context())
			if !ok {
				version = versioning.Version_1_0_0 // Default fallback
			}

			if !rl.IsAllowed(r, version) {
				// Rate limit exceeded
				w.Header().Set("X-Rate-Limit-Exceeded", "true")
				w.Header().Set("Retry-After", "60") // Retry after 1 minute
				http.Error(w, `{"error": "Rate limit exceeded", "message": "Too many requests. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			// Add rate limit headers
			state := rl.GetClientState(r, version)
			if state != nil {
				w.Header().Set("X-Rate-Limit-Algorithm", rl.config.Algorithm.String())
				if rl.config.Algorithm == TokenBucket {
					w.Header().Set("X-Rate-Limit-Tokens", fmt.Sprintf("%.1f", state.Tokens))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
