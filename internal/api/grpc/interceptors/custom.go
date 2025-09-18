package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CustomInterceptor represents a custom interceptor
type CustomInterceptor struct {
	name        string
	description string
	enabled     bool
	logger      *slog.Logger
}

// CustomInterceptorFunc represents a function that can be used as a custom interceptor
type CustomInterceptorFunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)

// CustomStreamInterceptorFunc represents a function that can be used as a custom stream interceptor
type CustomStreamInterceptorFunc func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error

// NewCustomInterceptor creates a new custom interceptor
func NewCustomInterceptor(name, description string, logger *slog.Logger) *CustomInterceptor {
	if logger == nil {
		logger = slog.Default()
	}

	return &CustomInterceptor{
		name:        name,
		description: description,
		enabled:     true,
		logger:      logger,
	}
}

// Enable enables the interceptor
func (ci *CustomInterceptor) Enable() {
	ci.enabled = true
}

// Disable disables the interceptor
func (ci *CustomInterceptor) Disable() {
	ci.enabled = false
}

// IsEnabled returns whether the interceptor is enabled
func (ci *CustomInterceptor) IsEnabled() bool {
	return ci.enabled
}

// GetName returns the interceptor name
func (ci *CustomInterceptor) GetName() string {
	return ci.name
}

// GetDescription returns the interceptor description
func (ci *CustomInterceptor) GetDescription() string {
	return ci.description
}

// RateLimitInterceptor provides rate limiting functionality
type RateLimitInterceptor struct {
	*CustomInterceptor
	requestsPerSecond int
	burstSize         int
	requestCounts     map[string]int64
	lastReset         time.Time
}

// NewRateLimitInterceptor creates a new rate limit interceptor
func NewRateLimitInterceptor(requestsPerSecond, burstSize int, logger *slog.Logger) *RateLimitInterceptor {
	base := NewCustomInterceptor("rate_limit", "Rate limiting interceptor", logger)

	return &RateLimitInterceptor{
		CustomInterceptor: base,
		requestsPerSecond: requestsPerSecond,
		burstSize:         burstSize,
		requestCounts:     make(map[string]int64),
		lastReset:         time.Now(),
	}
}

// UnaryRateLimitInterceptor returns a unary server interceptor for rate limiting
func (rli *RateLimitInterceptor) UnaryRateLimitInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !rli.IsEnabled() {
			return handler(ctx, req)
		}

		// Get client identifier (IP address or user ID)
		clientID := rli.getClientID(ctx)

		// Check rate limit
		if !rli.checkRateLimit(clientID) {
			rli.logger.Warn("Rate limit exceeded", "client_id", clientID, "method", info.FullMethod)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// StreamRateLimitInterceptor returns a stream server interceptor for rate limiting
func (rli *RateLimitInterceptor) StreamRateLimitInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !rli.IsEnabled() {
			return handler(srv, ss)
		}

		// Get client identifier
		clientID := rli.getClientID(ss.Context())

		// Check rate limit
		if !rli.checkRateLimit(clientID) {
			rli.logger.Warn("Rate limit exceeded for stream", "client_id", clientID, "method", info.FullMethod)
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// getClientID extracts client identifier from context
func (rli *RateLimitInterceptor) getClientID(ctx context.Context) string {
	// Try to get user ID first
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}

	// Fall back to IP address (would need to be set by another interceptor)
	if ip, ok := ctx.Value("client_ip").(string); ok {
		return ip
	}

	// Default to "unknown"
	return "unknown"
}

// checkRateLimit checks if the client is within rate limits
func (rli *RateLimitInterceptor) checkRateLimit(clientID string) bool {
	now := time.Now()

	// Reset counters if needed
	if now.Sub(rli.lastReset) >= time.Second {
		rli.requestCounts = make(map[string]int64)
		rli.lastReset = now
	}

	// Check current count
	currentCount := rli.requestCounts[clientID]
	if currentCount >= int64(rli.requestsPerSecond) {
		return false
	}

	// Increment count
	rli.requestCounts[clientID]++
	return true
}

// ValidationInterceptor provides request validation functionality
type ValidationInterceptor struct {
	*CustomInterceptor
	validators map[string]func(interface{}) error
}

// NewValidationInterceptor creates a new validation interceptor
func NewValidationInterceptor(logger *slog.Logger) *ValidationInterceptor {
	base := NewCustomInterceptor("validation", "Request validation interceptor", logger)

	return &ValidationInterceptor{
		CustomInterceptor: base,
		validators:        make(map[string]func(interface{}) error),
	}
}

// AddValidator adds a validator for a specific method
func (vi *ValidationInterceptor) AddValidator(method string, validator func(interface{}) error) {
	vi.validators[method] = validator
}

// UnaryValidationInterceptor returns a unary server interceptor for validation
func (vi *ValidationInterceptor) UnaryValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !vi.IsEnabled() {
			return handler(ctx, req)
		}

		// Check if validator exists for this method
		validator, exists := vi.validators[info.FullMethod]
		if !exists {
			return handler(ctx, req)
		}

		// Validate request
		if err := validator(req); err != nil {
			vi.logger.Error("Request validation failed", "method", info.FullMethod, "error", err)
			return nil, status.Error(codes.InvalidArgument, "request validation failed: "+err.Error())
		}

		return handler(ctx, req)
	}
}

// StreamValidationInterceptor returns a stream server interceptor for validation
func (vi *ValidationInterceptor) StreamValidationInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !vi.IsEnabled() {
			return handler(srv, ss)
		}

		// For streams, we validate the first message
		// This is a simplified implementation
		return handler(srv, ss)
	}
}

// CacheInterceptor provides caching functionality
type CacheInterceptor struct {
	*CustomInterceptor
	cache       map[string]interface{}
	cacheTTL    time.Duration
	lastCleanup time.Time
}

// NewCacheInterceptor creates a new cache interceptor
func NewCacheInterceptor(cacheTTL time.Duration, logger *slog.Logger) *CacheInterceptor {
	base := NewCustomInterceptor("cache", "Response caching interceptor", logger)

	return &CacheInterceptor{
		CustomInterceptor: base,
		cache:             make(map[string]interface{}),
		cacheTTL:          cacheTTL,
		lastCleanup:       time.Now(),
	}
}

// UnaryCacheInterceptor returns a unary server interceptor for caching
func (ci *CacheInterceptor) UnaryCacheInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !ci.IsEnabled() {
			return handler(ctx, req)
		}

		// Generate cache key
		cacheKey := ci.generateCacheKey(info.FullMethod, req)

		// Check cache
		if cached, exists := ci.cache[cacheKey]; exists {
			ci.logger.Debug("Cache hit", "method", info.FullMethod, "cache_key", cacheKey)
			return cached, nil
		}

		// Call handler
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		// Cache response
		ci.cache[cacheKey] = resp
		ci.logger.Debug("Response cached", "method", info.FullMethod, "cache_key", cacheKey)

		// Cleanup old cache entries
		ci.cleanupCache()

		return resp, err
	}
}

// generateCacheKey generates a cache key for the request
func (ci *CacheInterceptor) generateCacheKey(method string, _ interface{}) string {
	// Simple cache key generation - in production, use a proper hash function
	return method + "_" + time.Now().Format("20060102150405")
}

// cleanupCache removes expired cache entries
func (ci *CacheInterceptor) cleanupCache() {
	now := time.Now()
	if now.Sub(ci.lastCleanup) < ci.cacheTTL {
		return
	}

	// Simple cleanup - remove all entries
	// In production, implement proper TTL-based cleanup
	ci.cache = make(map[string]interface{})
	ci.lastCleanup = now
}

// CircuitBreakerInterceptor provides circuit breaker functionality
type CircuitBreakerInterceptor struct {
	*CustomInterceptor
	failureThreshold int
	timeout          time.Duration
	state            string // "closed", "open", "half-open"
	failureCount     int
	lastFailureTime  time.Time
}

// NewCircuitBreakerInterceptor creates a new circuit breaker interceptor
func NewCircuitBreakerInterceptor(failureThreshold int, timeout time.Duration, logger *slog.Logger) *CircuitBreakerInterceptor {
	base := NewCustomInterceptor("circuit_breaker", "Circuit breaker interceptor", logger)

	return &CircuitBreakerInterceptor{
		CustomInterceptor: base,
		failureThreshold:  failureThreshold,
		timeout:           timeout,
		state:             "closed",
		failureCount:      0,
	}
}

// UnaryCircuitBreakerInterceptor returns a unary server interceptor for circuit breaking
func (cbi *CircuitBreakerInterceptor) UnaryCircuitBreakerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !cbi.IsEnabled() {
			return handler(ctx, req)
		}

		// Check circuit breaker state
		if cbi.state == "open" {
			if time.Since(cbi.lastFailureTime) < cbi.timeout {
				cbi.logger.Warn("Circuit breaker is open", "method", info.FullMethod)
				return nil, status.Error(codes.Unavailable, "service temporarily unavailable")
			}
			// Try to close the circuit
			cbi.state = "half-open"
		}

		// Call handler
		resp, err := handler(ctx, req)

		// Update circuit breaker state
		if err != nil {
			cbi.failureCount++
			cbi.lastFailureTime = time.Now()

			if cbi.failureCount >= cbi.failureThreshold {
				cbi.state = "open"
				cbi.logger.Error("Circuit breaker opened", "method", info.FullMethod, "failure_count", cbi.failureCount)
			}
		} else {
			// Reset on success
			cbi.failureCount = 0
			cbi.state = "closed"
		}

		return resp, err
	}
}

// GetState returns the current circuit breaker state
func (cbi *CircuitBreakerInterceptor) GetState() string {
	return cbi.state
}

// GetFailureCount returns the current failure count
func (cbi *CircuitBreakerInterceptor) GetFailureCount() int {
	return cbi.failureCount
}

// StreamCircuitBreakerInterceptor returns a stream server interceptor for circuit breaking
func (cbi *CircuitBreakerInterceptor) StreamCircuitBreakerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !cbi.IsEnabled() {
			return handler(srv, ss)
		}

		// Check circuit breaker state
		if cbi.state == "open" {
			if time.Since(cbi.lastFailureTime) < cbi.timeout {
				cbi.logger.Warn("Circuit breaker is open for stream", "method", info.FullMethod)
				return status.Error(codes.Unavailable, "service temporarily unavailable")
			}
			// Try to close the circuit
			cbi.state = "half-open"
		}

		// Call handler
		err := handler(srv, ss)

		// Update circuit breaker state
		if err != nil {
			cbi.failureCount++
			cbi.lastFailureTime = time.Now()

			if cbi.failureCount >= cbi.failureThreshold {
				cbi.state = "open"
				cbi.logger.Error("Circuit breaker opened for stream", "method", info.FullMethod, "failure_count", cbi.failureCount)
			}
		} else {
			// Reset on success
			cbi.failureCount = 0
			cbi.state = "closed"
		}

		return err
	}
}
