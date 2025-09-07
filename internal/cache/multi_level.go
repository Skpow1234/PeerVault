package cache

import (
	"context"
	"sync"
	"time"
)

// MultiLevelCache implements a multi-level cache with L1 (memory) and L2 (disk) caches
type MultiLevelCache[T any] struct {
	l1Cache Cache[T]
	l2Cache Cache[T]
	mu      sync.RWMutex
	stats   MultiLevelCacheStats
}

// MultiLevelCacheStats holds statistics for multi-level cache
type MultiLevelCacheStats struct {
	L1Stats     CacheStats `json:"l1_stats"`
	L2Stats     CacheStats `json:"l2_stats"`
	TotalHits   int64      `json:"total_hits"`
	TotalMisses int64      `json:"total_misses"`
	L1Hits      int64      `json:"l1_hits"`
	L2Hits      int64      `json:"l2_hits"`
	L2Misses    int64      `json:"l2_misses"`
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache[T any](l1MaxSize int, l2MaxSize int) *MultiLevelCache[T] {
	return &MultiLevelCache[T]{
		l1Cache: NewMemoryCache[T](l1MaxSize),
		l2Cache: NewMemoryCache[T](l2MaxSize), // In a real implementation, this would be a disk cache
	}
}

// Get retrieves a value from the cache, checking L1 first, then L2
func (mlc *MultiLevelCache[T]) Get(ctx context.Context, key string) (T, bool) {
	// Try L1 cache first
	if value, found := mlc.l1Cache.Get(ctx, key); found {
		mlc.mu.Lock()
		mlc.stats.TotalHits++
		mlc.stats.L1Hits++
		mlc.mu.Unlock()
		return value, true
	}

	// Try L2 cache
	if value, found := mlc.l2Cache.Get(ctx, key); found {
		// Promote to L1 cache
		_ = mlc.l1Cache.Set(ctx, key, value, 5*time.Minute) // Ignore error for promotion

		mlc.mu.Lock()
		mlc.stats.TotalHits++
		mlc.stats.L2Hits++
		mlc.mu.Unlock()
		return value, true
	}

	// Cache miss
	mlc.mu.Lock()
	mlc.stats.TotalMisses++
	mlc.stats.L2Misses++
	mlc.mu.Unlock()

	var zero T
	return zero, false
}

// Set stores a value in both L1 and L2 caches
func (mlc *MultiLevelCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	// Set in L1 cache with shorter TTL
	l1TTL := ttl
	if l1TTL > 5*time.Minute {
		l1TTL = 5 * time.Minute
	}

	if err := mlc.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
		return err
	}

	// Set in L2 cache with full TTL
	return mlc.l2Cache.Set(ctx, key, value, ttl)
}

// Delete removes a value from both caches
func (mlc *MultiLevelCache[T]) Delete(ctx context.Context, key string) error {
	// Delete from both caches
	_ = mlc.l1Cache.Delete(ctx, key) // Ignore L1 delete error
	return mlc.l2Cache.Delete(ctx, key)
}

// Clear removes all values from both caches
func (mlc *MultiLevelCache[T]) Clear(ctx context.Context) error {
	// Clear both caches
	_ = mlc.l1Cache.Clear(ctx) // Ignore L1 clear error
	return mlc.l2Cache.Clear(ctx)
}

// Stats returns combined statistics from both caches
func (mlc *MultiLevelCache[T]) Stats() MultiLevelCacheStats {
	mlc.mu.RLock()
	defer mlc.mu.RUnlock()

	stats := mlc.stats
	stats.L1Stats = mlc.l1Cache.Stats()
	stats.L2Stats = mlc.l2Cache.Stats()

	return stats
}

// Close closes both caches
func (mlc *MultiLevelCache[T]) Close() error {
	_ = mlc.l1Cache.(*MemoryCache[T]).Close() // Ignore L1 close error
	_ = mlc.l2Cache.(*MemoryCache[T]).Close() // Ignore L2 close error
	return nil
}

// CacheWarmer provides functionality to warm up caches
type CacheWarmer[T any] struct {
	cache Cache[T]
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer[T any](cache Cache[T]) *CacheWarmer[T] {
	return &CacheWarmer[T]{
		cache: cache,
	}
}

// WarmUp warms up the cache with the provided key-value pairs
func (cw *CacheWarmer[T]) WarmUp(ctx context.Context, items map[string]T, ttl time.Duration) error {
	for key, value := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := cw.cache.Set(ctx, key, value, ttl); err != nil {
				return err
			}
		}
	}
	return nil
}

// WarmUpAsync warms up the cache asynchronously
func (cw *CacheWarmer[T]) WarmUpAsync(ctx context.Context, items map[string]T, ttl time.Duration) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		if err := cw.WarmUp(ctx, items, ttl); err != nil {
			errChan <- err
		}
	}()

	return errChan
}
