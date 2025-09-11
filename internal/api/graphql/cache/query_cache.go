package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cache"
)

// QueryCache represents a GraphQL query result cache
type QueryCache struct {
	// Cache storage
	storage cache.Cache[*CachedQueryResult]

	// Cache configuration
	config *QueryCacheConfig

	// Invalidation channels
	invalidationChannels map[string][]chan string

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	logger *slog.Logger

	// Metrics
	metrics *CacheMetrics
}

// QueryCacheConfig holds configuration for the query cache
type QueryCacheConfig struct {
	DefaultTTL           time.Duration        `json:"defaultTTL"`
	MaxCacheSize         int                  `json:"maxCacheSize"`
	EnableCompression    bool                 `json:"enableCompression"`
	EnableMetrics        bool                 `json:"enableMetrics"`
	InvalidationStrategy InvalidationStrategy `json:"invalidationStrategy"`
}

// InvalidationStrategy defines how cache invalidation works
type InvalidationStrategy string

const (
	// InvalidationStrategyTTL - Invalidate based on TTL only
	InvalidationStrategyTTL InvalidationStrategy = "ttl"
	// InvalidationStrategyEvent - Invalidate based on events
	InvalidationStrategyEvent InvalidationStrategy = "event"
	// InvalidationStrategyHybrid - Use both TTL and events
	InvalidationStrategyHybrid InvalidationStrategy = "hybrid"
)

// DefaultQueryCacheConfig returns the default cache configuration
func DefaultQueryCacheConfig() *QueryCacheConfig {
	return &QueryCacheConfig{
		DefaultTTL:           5 * time.Minute,
		MaxCacheSize:         1000,
		EnableCompression:    true,
		EnableMetrics:        true,
		InvalidationStrategy: InvalidationStrategyHybrid,
	}
}

// CachedQueryResult represents a cached GraphQL query result
type CachedQueryResult struct {
	Data       interface{}            `json:"data"`
	Errors     []interface{}          `json:"errors,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	TTL        time.Duration          `json:"ttl"`
	QueryHash  string                 `json:"queryHash"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// CacheMetrics holds cache performance metrics
type CacheMetrics struct {
	Hits          int64     `json:"hits"`
	Misses        int64     `json:"misses"`
	Sets          int64     `json:"sets"`
	Deletes       int64     `json:"deletes"`
	Invalidations int64     `json:"invalidations"`
	Size          int       `json:"size"`
	LastReset     time.Time `json:"lastReset"`
}

// NewQueryCache creates a new GraphQL query cache
func NewQueryCache(config *QueryCacheConfig, logger *slog.Logger) (*QueryCache, error) {
	if config == nil {
		config = DefaultQueryCacheConfig()
	}

	// Create underlying cache storage
	storage := cache.NewMemoryCache[*CachedQueryResult](config.MaxCacheSize)

	qc := &QueryCache{
		storage:              storage,
		config:               config,
		invalidationChannels: make(map[string][]chan string),
		logger:               logger,
		metrics: &CacheMetrics{
			LastReset: time.Now(),
		},
	}

	// Start cleanup goroutine
	go qc.startCleanup()

	return qc, nil
}

// Get retrieves a cached query result
func (qc *QueryCache) Get(ctx context.Context, query string, variables map[string]interface{}) (*CachedQueryResult, bool) {
	key := qc.generateCacheKey(query, variables)

	result, found := qc.storage.Get(ctx, key)
	if found {
		// Check if result is still valid
		if time.Since(result.Timestamp) < result.TTL {
			qc.metrics.Hits++
			qc.logger.Debug("Cache hit", "key", key, "age", time.Since(result.Timestamp))
			return result, true
		} else {
			// Result expired, remove it
			if err := qc.storage.Delete(ctx, key); err != nil {
				qc.logger.Warn("Failed to delete expired cache entry", "key", key, "error", err)
			}
			qc.metrics.Misses++
			qc.logger.Debug("Cache miss (expired)", "key", key)
			return nil, false
		}
	}

	qc.metrics.Misses++
	qc.logger.Debug("Cache miss", "key", key)
	return nil, false
}

// Set stores a query result in the cache
func (qc *QueryCache) Set(ctx context.Context, query string, variables map[string]interface{}, data interface{}, errors []interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = qc.config.DefaultTTL
	}

	key := qc.generateCacheKey(query, variables)

	result := &CachedQueryResult{
		Data:      data,
		Errors:    errors,
		Timestamp: time.Now(),
		TTL:       ttl,
		QueryHash: qc.hashQuery(query),
		Variables: variables,
	}

	if err := qc.storage.Set(ctx, key, result, ttl); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	qc.metrics.Sets++
	qc.logger.Debug("Cache set", "key", key, "ttl", ttl)
	return nil
}

// Delete removes a cached query result
func (qc *QueryCache) Delete(ctx context.Context, query string, variables map[string]interface{}) error {
	key := qc.generateCacheKey(query, variables)

	if err := qc.storage.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	qc.metrics.Deletes++
	qc.logger.Debug("Cache delete", "key", key)
	return nil
}

// InvalidateByPattern invalidates cache entries matching a pattern
func (qc *QueryCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	// This is a simplified implementation
	// In a real implementation, you would use pattern matching
	// For now, we'll invalidate all entries

	keys, err := qc.storage.Keys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	var invalidated int
	for _, key := range keys {
		if err := qc.storage.Delete(ctx, key); err == nil {
			invalidated++
		}
	}

	qc.metrics.Invalidations += int64(invalidated)
	qc.logger.Info("Cache invalidation by pattern", "pattern", pattern, "invalidated", invalidated)
	return nil
}

// InvalidateByTags invalidates cache entries with specific tags
func (qc *QueryCache) InvalidateByTags(ctx context.Context, tags []string) error {
	// This would require extending the cache to support tags
	// For now, we'll implement a simple pattern-based invalidation

	var totalInvalidated int
	for _, tag := range tags {
		if err := qc.InvalidateByPattern(ctx, tag); err == nil {
			totalInvalidated++
		}
	}

	qc.logger.Info("Cache invalidation by tags", "tags", tags, "invalidated", totalInvalidated)
	return nil
}

// Clear clears all cached entries
func (qc *QueryCache) Clear(ctx context.Context) error {
	if err := qc.storage.Clear(ctx); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	qc.logger.Info("Cache cleared")
	return nil
}

// GetMetrics returns cache performance metrics
func (qc *QueryCache) GetMetrics() *CacheMetrics {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	// Get current cache size
	keys, _ := qc.storage.Keys(context.Background())
	qc.metrics.Size = len(keys)

	// Create a copy to avoid race conditions
	metrics := *qc.metrics
	return &metrics
}

// ResetMetrics resets the cache metrics
func (qc *QueryCache) ResetMetrics() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	qc.metrics = &CacheMetrics{
		LastReset: time.Now(),
	}
}

// generateCacheKey generates a cache key from query and variables
func (qc *QueryCache) generateCacheKey(query string, variables map[string]interface{}) string {
	// Create a deterministic key from query and variables
	keyData := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	keyBytes, _ := json.Marshal(keyData)
	hash := sha256.Sum256(keyBytes)
	return hex.EncodeToString(hash[:])
}

// hashQuery generates a hash for the query
func (qc *QueryCache) hashQuery(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:])
}

// startCleanup starts the cache cleanup goroutine
func (qc *QueryCache) startCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		qc.cleanup()
	}
}

// cleanup removes expired entries from the cache
func (qc *QueryCache) cleanup() {
	ctx := context.Background()
	keys, err := qc.storage.Keys(ctx)
	if err != nil {
		qc.logger.Error("Failed to get keys for cleanup", "error", err)
		return
	}

	var cleaned int
	for _, key := range keys {
		result, found := qc.storage.Get(ctx, key)
		if found && time.Since(result.Timestamp) >= result.TTL {
			if err := qc.storage.Delete(ctx, key); err == nil {
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		qc.logger.Debug("Cache cleanup completed", "cleaned", cleaned)
	}
}

// SubscribeToInvalidation subscribes to cache invalidation events
func (qc *QueryCache) SubscribeToInvalidation(pattern string) <-chan string {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	ch := make(chan string, 10)
	if qc.invalidationChannels[pattern] == nil {
		qc.invalidationChannels[pattern] = make([]chan string, 0)
	}
	qc.invalidationChannels[pattern] = append(qc.invalidationChannels[pattern], ch)

	return ch
}

// UnsubscribeFromInvalidation unsubscribes from cache invalidation events
func (qc *QueryCache) UnsubscribeFromInvalidation(pattern string, ch <-chan string) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	if channels, exists := qc.invalidationChannels[pattern]; exists {
		for i, channel := range channels {
			if channel == ch {
				qc.invalidationChannels[pattern] = append(channels[:i], channels[i+1:]...)
				close(channel)
				break
			}
		}
	}
}

// PublishInvalidation publishes an invalidation event
func (qc *QueryCache) PublishInvalidation(pattern string) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	if channels, exists := qc.invalidationChannels[pattern]; exists {
		for _, ch := range channels {
			select {
			case ch <- pattern:
			default:
				// Channel is full, skip
			}
		}
	}
}

// WarmCache preloads the cache with common queries
func (qc *QueryCache) WarmCache(ctx context.Context, queries []CacheWarmQuery) error {
	for _, warmQuery := range queries {
		// Execute the query and cache the result
		// This would integrate with your GraphQL execution engine
		qc.logger.Info("Warming cache", "query", warmQuery.Query, "ttl", warmQuery.TTL)

		// For now, just log the warm query
		// In a real implementation, you would execute the query and cache the result
	}

	return nil
}

// CacheWarmQuery represents a query to warm the cache
type CacheWarmQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	TTL       time.Duration          `json:"ttl"`
	Priority  int                    `json:"priority"`
}

// GetCacheStats returns detailed cache statistics
func (qc *QueryCache) GetCacheStats() map[string]interface{} {
	metrics := qc.GetMetrics()

	hitRate := float64(0)
	if metrics.Hits+metrics.Misses > 0 {
		hitRate = float64(metrics.Hits) / float64(metrics.Hits+metrics.Misses)
	}

	return map[string]interface{}{
		"hits":          metrics.Hits,
		"misses":        metrics.Misses,
		"sets":          metrics.Sets,
		"deletes":       metrics.Deletes,
		"invalidations": metrics.Invalidations,
		"size":          metrics.Size,
		"hitRate":       hitRate,
		"lastReset":     metrics.LastReset,
		"config": map[string]interface{}{
			"defaultTTL":           qc.config.DefaultTTL,
			"maxCacheSize":         qc.config.MaxCacheSize,
			"enableCompression":    qc.config.EnableCompression,
			"invalidationStrategy": qc.config.InvalidationStrategy,
		},
	}
}
