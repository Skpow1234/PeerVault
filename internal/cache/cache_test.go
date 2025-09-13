package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheItem_IsExpired(t *testing.T) {
	now := time.Now()

	// Not expired
	item := &CacheItem[string]{
		ExpiresAt: now.Add(1 * time.Hour),
	}
	assert.False(t, item.IsExpired())

	// Expired
	expiredItem := &CacheItem[string]{
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	assert.True(t, expiredItem.IsExpired())
}

func TestCacheItem_Touch(t *testing.T) {
	initialTime := time.Now().Add(-1 * time.Hour)
	item := &CacheItem[string]{
		Value:       "test",
		AccessCount: 5,
		LastAccess:  initialTime,
	}

	item.Touch()

	assert.Equal(t, int64(6), item.AccessCount)
	assert.True(t, item.LastAccess.After(initialTime))
}

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache[string](100)
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.items)
	assert.Equal(t, 100, cache.maxSize)

	// Cleanup
	cache.Close()
}

func TestMemoryCache_SetGet(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)

	// Get the value
	value, exists := cache.Get(ctx, "key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Check stats
	stats := cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 1, stats.Size)
	assert.Equal(t, 100, stats.MaxSize)
}

func TestMemoryCache_Get_NonExistent(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Get non-existent key
	value, exists := cache.Get(ctx, "nonexistent")
	assert.False(t, exists)
	assert.Equal(t, "", value)

	// Check stats
	stats := cache.Stats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestMemoryCache_Get_Expired(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Set a value with short TTL
	err := cache.Set(ctx, "key1", "value1", 1*time.Millisecond)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Get the expired value
	value, exists := cache.Get(ctx, "key1")
	assert.False(t, exists)
	assert.Equal(t, "", value)

	// Check stats (should count as miss)
	stats := cache.Stats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 0, stats.Size) // Item should be cleaned up
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)

	// Verify it exists
	_, exists := cache.Get(ctx, "key1")
	assert.True(t, exists)

	// Delete it
	err = cache.Delete(ctx, "key1")
	require.NoError(t, err)

	// Verify it's gone
	_, exists = cache.Get(ctx, "key1")
	assert.False(t, exists)
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values
	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)
	err = cache.Set(ctx, "key2", "value2", 1*time.Hour)
	require.NoError(t, err)

	// Verify they exist
	_, exists1 := cache.Get(ctx, "key1")
	_, exists2 := cache.Get(ctx, "key2")
	assert.True(t, exists1)
	assert.True(t, exists2)

	// Clear all
	err = cache.Clear(ctx)
	require.NoError(t, err)

	// Verify they're gone
	_, exists1 = cache.Get(ctx, "key1")
	_, exists2 = cache.Get(ctx, "key2")
	assert.False(t, exists1)
	assert.False(t, exists2)

	assert.Equal(t, 0, cache.Stats().Size)
}

func TestMemoryCache_Keys(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Set multiple values
	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)
	err = cache.Set(ctx, "key2", "value2", 1*time.Hour)
	require.NoError(t, err)

	// Get keys
	keys, err := cache.Keys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestMemoryCache_LRUEviction(t *testing.T) {
	cache := NewMemoryCache[string](2) // Very small cache
	defer cache.Close()

	ctx := context.Background()

	// Fill the cache
	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamps

	err = cache.Set(ctx, "key2", "value2", 1*time.Hour)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	// Access key1 to make it more recently used
	_, _ = cache.Get(ctx, "key1")

	// Add a third item, should evict the least recently used
	err = cache.Set(ctx, "key3", "value3", 1*time.Hour)
	require.NoError(t, err)

	// Check which items exist
	_, exists1 := cache.Get(ctx, "key1")
	_, exists2 := cache.Get(ctx, "key2")
	_, exists3 := cache.Get(ctx, "key3")

	// Only 2 items should exist (cache size is 2)
	existsCount := 0
	if exists1 {
		existsCount++
	}
	if exists2 {
		existsCount++
	}
	if exists3 {
		existsCount++
	}
	assert.Equal(t, 2, existsCount)

	// key1 should exist (was recently accessed)
	assert.True(t, exists1, "key1 should exist as it was recently accessed")

	// key3 should exist (newly added)
	assert.True(t, exists3, "key3 should exist as it was just added")

	// Check eviction stats
	stats := cache.Stats()
	assert.Equal(t, int64(1), stats.Evictions)
}

func TestMemoryCache_Stats(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Initial stats
	stats := cache.Stats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, 100, stats.MaxSize)
	assert.Equal(t, float64(0), stats.HitRate)

	// Set a value
	err := cache.Set(ctx, "key1", "value1", 1*time.Hour)
	require.NoError(t, err)

	// Get existing value (hit)
	_, _ = cache.Get(ctx, "key1")

	// Get non-existing value (miss)
	_, _ = cache.Get(ctx, "nonexistent")

	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, 1, stats.Size)
	assert.Equal(t, float64(0.5), stats.HitRate) // 1 hit out of 2 requests
}

func TestMemoryCache_CleanupExpired(t *testing.T) {
	cache := NewMemoryCache[string](100)
	defer cache.Close()

	ctx := context.Background()

	// Set values with different TTLs
	err := cache.Set(ctx, "short", "value1", 10*time.Millisecond)
	require.NoError(t, err)
	err = cache.Set(ctx, "long", "value2", 1*time.Hour)
	require.NoError(t, err)

	// Verify both exist
	_, exists1 := cache.Get(ctx, "short")
	_, exists2 := cache.Get(ctx, "long")
	assert.True(t, exists1)
	assert.True(t, exists2)

	// Wait for short TTL to expire
	time.Sleep(50 * time.Millisecond)

	// Manually trigger cleanup (normally done by cleanup routine)
	cache.cleanupExpired()

	// Short TTL item should be gone
	_, exists1 = cache.Get(ctx, "short")
	assert.False(t, exists1)

	// Long TTL item should still exist
	value2, exists2 := cache.Get(ctx, "long")
	assert.True(t, exists2)
	assert.Equal(t, "value2", value2)
}

func TestMemoryCache_WithGenericTypes(t *testing.T) {
	// Test with different types
	intCache := NewMemoryCache[int](10)
	defer intCache.Close()

	ctx := context.Background()

	err := intCache.Set(ctx, "int_key", 42, 1*time.Hour)
	require.NoError(t, err)

	value, exists := intCache.Get(ctx, "int_key")
	assert.True(t, exists)
	assert.Equal(t, 42, value)

	// Test with struct
	type TestStruct struct {
		Name  string
		Value int
	}

	structCache := NewMemoryCache[TestStruct](10)
	defer structCache.Close()

	testStruct := TestStruct{Name: "test", Value: 123}
	err = structCache.Set(ctx, "struct_key", testStruct, 1*time.Hour)
	require.NoError(t, err)

	result, exists := structCache.Get(ctx, "struct_key")
	assert.True(t, exists)
	assert.Equal(t, testStruct, result)
}
