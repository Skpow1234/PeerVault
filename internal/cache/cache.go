package cache

import (
	"context"
	"sync"
	"time"
)

// CacheItem represents a cached item
type CacheItem[T any] struct {
	Value       T
	ExpiresAt   time.Time
	CreatedAt   time.Time
	AccessCount int64
	LastAccess  time.Time
}

// IsExpired checks if the cache item has expired
func (item *CacheItem[T]) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

// Touch updates the access time and count
func (item *CacheItem[T]) Touch() {
	item.LastAccess = time.Now()
	item.AccessCount++
}

// Cache interface defines the basic cache operations
type Cache[T any] interface {
	Get(ctx context.Context, key string) (T, bool)
	Set(ctx context.Context, key string, value T, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Keys(ctx context.Context) ([]string, error)
	Stats() CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Size      int     `json:"size"`
	MaxSize   int     `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
	Evictions int64   `json:"evictions"`
}

// MemoryCache implements an in-memory cache with LRU eviction
type MemoryCache[T any] struct {
	items   map[string]*CacheItem[T]
	mu      sync.RWMutex
	maxSize int
	stats   CacheStats
	cleanup *time.Ticker
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache[T any](maxSize int) *MemoryCache[T] {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &MemoryCache[T]{
		items:   make(map[string]*CacheItem[T]),
		maxSize: maxSize,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Start cleanup routine
	cache.cleanup = time.NewTicker(1 * time.Minute)
	go cache.cleanupRoutine()

	return cache
}

// Get retrieves a value from the cache
func (mc *MemoryCache[T]) Get(ctx context.Context, key string) (T, bool) {
	mc.mu.RLock()
	item, exists := mc.items[key]
	mc.mu.RUnlock()

	if !exists {
		mc.mu.Lock()
		mc.stats.Misses++
		mc.mu.Unlock()
		var zero T
		return zero, false
	}

	// Check if expired
	if item.IsExpired() {
		mc.mu.Lock()
		delete(mc.items, key)
		mc.stats.Misses++
		mc.mu.Unlock()
		var zero T
		return zero, false
	}

	// Update access info
	item.Touch()

	mc.mu.Lock()
	mc.stats.Hits++
	mc.mu.Unlock()

	return item.Value, true
}

// Set stores a value in the cache
func (mc *MemoryCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict
	if len(mc.items) >= mc.maxSize {
		mc.evictLRU()
	}

	now := time.Now()
	mc.items[key] = &CacheItem[T]{
		Value:       value,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		AccessCount: 1,
		LastAccess:  now,
	}

	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCache[T]) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
	return nil
}

// Clear removes all values from the cache
func (mc *MemoryCache[T]) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*CacheItem[T])
	return nil
}

// Keys returns all keys in the cache
func (mc *MemoryCache[T]) Keys(ctx context.Context) ([]string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	keys := make([]string, 0, len(mc.items))
	for key := range mc.items {
		keys = append(keys, key)
	}

	return keys, nil
}

// Stats returns cache statistics
func (mc *MemoryCache[T]) Stats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := mc.stats
	stats.Size = len(mc.items)
	stats.MaxSize = mc.maxSize

	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRate = float64(stats.Hits) / float64(total)
	}

	return stats
}

// Close closes the cache and cleans up resources
func (mc *MemoryCache[T]) Close() error {
	mc.cancel()
	mc.cleanup.Stop()
	return nil
}

// evictLRU removes the least recently used item
func (mc *MemoryCache[T]) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range mc.items {
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
		}
	}

	if oldestKey != "" {
		delete(mc.items, oldestKey)
		mc.stats.Evictions++
	}
}

// cleanupRoutine periodically removes expired items
func (mc *MemoryCache[T]) cleanupRoutine() {
	for {
		select {
		case <-mc.cleanup.C:
			mc.cleanupExpired()
		case <-mc.ctx.Done():
			return
		}
	}
}

// cleanupExpired removes expired items
func (mc *MemoryCache[T]) cleanupExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for key, item := range mc.items {
		if item.IsExpired() {
			delete(mc.items, key)
		}
	}
}
