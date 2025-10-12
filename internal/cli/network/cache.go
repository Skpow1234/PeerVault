package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// CacheManager manages caching for improved performance
type CacheManager struct {
	client    *client.Client
	configDir string
	cache     map[string]*CacheEntry
	config    *CacheConfig
	stats     *CacheStats
	mu        sync.RWMutex
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key          string                 `json:"key"`
	Value        interface{}            `json:"value"`
	Size         int64                  `json:"size"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	LastAccessed time.Time              `json:"last_accessed"`
	AccessCount  int64                  `json:"access_count"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	MaxSize         int64          `json:"max_size"`         // Maximum cache size in bytes
	MaxEntries      int            `json:"max_entries"`      // Maximum number of entries
	DefaultTTL      time.Duration  `json:"default_ttl"`      // Default time-to-live
	CleanupInterval time.Duration  `json:"cleanup_interval"` // Cleanup interval
	EvictionPolicy  EvictionPolicy `json:"eviction_policy"`  // Eviction policy
	Compression     bool           `json:"compression"`      // Enable compression
	Persistence     bool           `json:"persistence"`      // Enable persistence
}

// EvictionPolicy represents cache eviction strategies
type EvictionPolicy string

const (
	LRU  EvictionPolicy = "lru"  // Least Recently Used
	LFU  EvictionPolicy = "lfu"  // Least Frequently Used
	TTL  EvictionPolicy = "ttl"  // Time To Live
	FIFO EvictionPolicy = "fifo" // First In First Out
)

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries    int64     `json:"total_entries"`
	CurrentSize     int64     `json:"current_size"`
	MaxSize         int64     `json:"max_size"`
	HitCount        int64     `json:"hit_count"`
	MissCount       int64     `json:"miss_count"`
	HitRate         float64   `json:"hit_rate"`
	EvictionCount   int64     `json:"eviction_count"`
	ExpirationCount int64     `json:"expiration_count"`
	LastCleanup     time.Time `json:"last_cleanup"`
	LastUpdated     time.Time `json:"last_updated"`
}

// CacheResult represents the result of a cache operation
type CacheResult struct {
	Hit       bool        `json:"hit"`
	Value     interface{} `json:"value,omitempty"`
	Size      int64       `json:"size"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(client *client.Client, configDir string) *CacheManager {
	cm := &CacheManager{
		client:    client,
		configDir: configDir,
		cache:     make(map[string]*CacheEntry),
		config:    getDefaultCacheConfig(),
		stats:     &CacheStats{},
	}

	_ = cm.loadConfig() // Ignore error for initialization
	_ = cm.loadCache()  // Ignore error for initialization
	_ = cm.loadStats()  // Ignore error for initialization

	// Start cleanup routine
	go cm.startCleanupRoutine()

	return cm
}

// Get retrieves a value from the cache
func (cm *CacheManager) Get(key string) (*CacheResult, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, exists := cm.cache[key]
	if !exists {
		cm.stats.MissCount++
		cm.updateHitRate()
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		delete(cm.cache, entry.Key)
		cm.stats.ExpirationCount++
		cm.stats.MissCount++
		cm.updateHitRate()
		return nil, fmt.Errorf("key expired: %s", key)
	}

	// Update access statistics
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	cm.stats.HitCount++
	cm.updateHitRate()
	_ = cm.saveStats() // Ignore error for demo purposes

	return &CacheResult{
		Hit:       true,
		Value:     entry.Value,
		Size:      entry.Size,
		CreatedAt: entry.CreatedAt,
		ExpiresAt: entry.ExpiresAt,
	}, nil
}

// Set stores a value in the cache
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration, tags []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Calculate size (simplified)
	size := int64(len(fmt.Sprintf("%v", value)))

	// Check if we need to evict entries
	if cm.needsEviction(size) {
		cm.evictEntries(size)
	}

	// Create cache entry
	now := time.Now()
	entry := &CacheEntry{
		Key:          key,
		Value:        value,
		Size:         size,
		CreatedAt:    now,
		ExpiresAt:    now.Add(ttl),
		LastAccessed: now,
		AccessCount:  1,
		Tags:         tags,
		Metadata:     make(map[string]interface{}),
	}

	// Remove existing entry if it exists
	if existingEntry, exists := cm.cache[key]; exists {
		cm.stats.CurrentSize -= existingEntry.Size
	}

	cm.cache[key] = entry
	cm.stats.CurrentSize += size
	cm.stats.TotalEntries++
	cm.stats.LastUpdated = now

	_ = cm.saveCache() // Ignore error for demo purposes
	_ = cm.saveStats() // Ignore error for demo purposes

	return nil
}

// Delete removes a value from the cache
func (cm *CacheManager) Delete(key string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, exists := cm.cache[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	delete(cm.cache, key)
	cm.stats.CurrentSize -= entry.Size
	cm.stats.TotalEntries--
	cm.stats.LastUpdated = time.Now()

	_ = cm.saveCache() // Ignore error for demo purposes
	_ = cm.saveStats() // Ignore error for demo purposes

	return nil
}

// Clear removes all entries from the cache
func (cm *CacheManager) Clear() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.cache = make(map[string]*CacheEntry)
	cm.stats.CurrentSize = 0
	cm.stats.TotalEntries = 0
	cm.stats.LastUpdated = time.Now()

	_ = cm.saveCache() // Ignore error for demo purposes
	_ = cm.saveStats() // Ignore error for demo purposes

	return nil
}

// GetByTags retrieves all entries with specific tags
func (cm *CacheManager) GetByTags(tags []string) ([]*CacheEntry, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var results []*CacheEntry
	for _, entry := range cm.cache {
		// Check if entry has any of the specified tags
		for _, tag := range tags {
			for _, entryTag := range entry.Tags {
				if tag == entryTag {
					// Return a copy
					entryCopy := *entry
					results = append(results, &entryCopy)
					break
				}
			}
		}
	}

	return results, nil
}

// DeleteByTags removes all entries with specific tags
func (cm *CacheManager) DeleteByTags(tags []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var keysToDelete []string
	for key, entry := range cm.cache {
		for _, tag := range tags {
			for _, entryTag := range entry.Tags {
				if tag == entryTag {
					keysToDelete = append(keysToDelete, key)
					break
				}
			}
		}
	}

	for _, key := range keysToDelete {
		entry := cm.cache[key]
		delete(cm.cache, key)
		cm.stats.CurrentSize -= entry.Size
		cm.stats.TotalEntries--
	}

	cm.stats.LastUpdated = time.Now()
	_ = cm.saveCache() // Ignore error for demo purposes
	_ = cm.saveStats() // Ignore error for demo purposes

	return nil
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	stats := *cm.stats
	return &stats
}

// UpdateConfig updates cache configuration
func (cm *CacheManager) UpdateConfig(config *CacheConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config = config
	_ = cm.saveConfig() // Ignore error for demo purposes

	return nil
}

// GetConfig returns current cache configuration
func (cm *CacheManager) GetConfig() *CacheConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	config := *cm.config
	return &config
}

// ListEntries returns all cache entries
func (cm *CacheManager) ListEntries() []*CacheEntry {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var entries []*CacheEntry
	for _, entry := range cm.cache {
		// Return a copy
		entryCopy := *entry
		entries = append(entries, &entryCopy)
	}

	return entries
}

// Cache file operations
func (cm *CacheManager) CacheFile(fileID string, ttl time.Duration) error {
	// Download file content
	tempFile := fmt.Sprintf("/tmp/cache_%s", fileID)
	err := cm.client.DownloadFile(context.Background(), fileID, tempFile)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		_ = os.Remove(tempFile) // Clean up
		return fmt.Errorf("failed to read file: %w", err)
	}
	_ = os.Remove(tempFile) // Clean up

	// Cache the content
	return cm.Set(fmt.Sprintf("file:%s", fileID), content, ttl, []string{"file", "content"})
}

// GetCachedFile retrieves cached file content
func (cm *CacheManager) GetCachedFile(fileID string) ([]byte, error) {
	result, err := cm.Get(fmt.Sprintf("file:%s", fileID))
	if err != nil {
		return nil, err
	}

	content, ok := result.Value.([]byte)
	if !ok {
		return nil, fmt.Errorf("cached value is not []byte")
	}

	return content, nil
}

// Utility methods
func (cm *CacheManager) needsEviction(additionalSize int64) bool {
	return cm.stats.CurrentSize+additionalSize > cm.config.MaxSize ||
		cm.stats.TotalEntries >= int64(cm.config.MaxEntries)
}

func (cm *CacheManager) evictEntries(requiredSize int64) {
	switch cm.config.EvictionPolicy {
	case LRU:
		cm.evictLRU(requiredSize)
	case LFU:
		cm.evictLFU(requiredSize)
	case TTL:
		cm.evictTTL(requiredSize)
	case FIFO:
		cm.evictFIFO(requiredSize)
	default:
		cm.evictLRU(requiredSize)
	}
}

func (cm *CacheManager) evictLRU(requiredSize int64) {
	// Find least recently used entries
	var entries []*CacheEntry
	for _, entry := range cm.cache {
		entries = append(entries, entry)
	}

	// Sort by last accessed time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].LastAccessed.After(entries[j].LastAccessed) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove entries until we have enough space
	evictedSize := int64(0)
	for _, entry := range entries {
		if evictedSize >= requiredSize {
			break
		}

		delete(cm.cache, entry.Key)
		evictedSize += entry.Size
		cm.stats.CurrentSize -= entry.Size
		cm.stats.TotalEntries--
		cm.stats.EvictionCount++
	}
}

func (cm *CacheManager) evictLFU(requiredSize int64) {
	// Find least frequently used entries
	var entries []*CacheEntry
	for _, entry := range cm.cache {
		entries = append(entries, entry)
	}

	// Sort by access count (lowest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].AccessCount > entries[j].AccessCount {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove entries until we have enough space
	evictedSize := int64(0)
	for _, entry := range entries {
		if evictedSize >= requiredSize {
			break
		}

		delete(cm.cache, entry.Key)
		evictedSize += entry.Size
		cm.stats.CurrentSize -= entry.Size
		cm.stats.TotalEntries--
		cm.stats.EvictionCount++
	}
}

func (cm *CacheManager) evictTTL(requiredSize int64) {
	// Find expired entries first
	now := time.Now()
	var expiredKeys []string
	for key, entry := range cm.cache {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired entries
	evictedSize := int64(0)
	for _, key := range expiredKeys {
		entry := cm.cache[key]
		delete(cm.cache, key)
		evictedSize += entry.Size
		cm.stats.CurrentSize -= entry.Size
		cm.stats.TotalEntries--
		cm.stats.ExpirationCount++
	}

	// If still need more space, evict by TTL (closest to expiry first)
	if evictedSize < requiredSize {
		var entries []*CacheEntry
		for _, entry := range cm.cache {
			entries = append(entries, entry)
		}

		// Sort by expiry time (closest to expiry first)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].ExpiresAt.After(entries[j].ExpiresAt) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Remove entries until we have enough space
		for _, entry := range entries {
			if evictedSize >= requiredSize {
				break
			}

			delete(cm.cache, entry.Key)
			evictedSize += entry.Size
			cm.stats.CurrentSize -= entry.Size
			cm.stats.TotalEntries--
			cm.stats.EvictionCount++
		}
	}
}

func (cm *CacheManager) evictFIFO(requiredSize int64) {
	// Find oldest entries
	var entries []*CacheEntry
	for _, entry := range cm.cache {
		entries = append(entries, entry)
	}

	// Sort by creation time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].CreatedAt.After(entries[j].CreatedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove entries until we have enough space
	evictedSize := int64(0)
	for _, entry := range entries {
		if evictedSize >= requiredSize {
			break
		}

		delete(cm.cache, entry.Key)
		evictedSize += entry.Size
		cm.stats.CurrentSize -= entry.Size
		cm.stats.TotalEntries--
		cm.stats.EvictionCount++
	}
}

func (cm *CacheManager) updateHitRate() {
	total := cm.stats.HitCount + cm.stats.MissCount
	if total > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(total)
	}
}

// Cleanup routine
func (cm *CacheManager) startCleanupRoutine() {
	ticker := time.NewTicker(cm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cm.cleanup()
	}
}

func (cm *CacheManager) cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	var expiredKeys []string

	for key, entry := range cm.cache {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		entry := cm.cache[key]
		delete(cm.cache, key)
		cm.stats.CurrentSize -= entry.Size
		cm.stats.TotalEntries--
		cm.stats.ExpirationCount++
	}

	cm.stats.LastCleanup = now
	_ = cm.saveCache() // Ignore error for demo purposes
	_ = cm.saveStats() // Ignore error for demo purposes
}

// Configuration management
func getDefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:         100 * 1024 * 1024, // 100MB
		MaxEntries:      1000,
		DefaultTTL:      1 * time.Hour,
		CleanupInterval: 5 * time.Minute,
		EvictionPolicy:  LRU,
		Compression:     true,
		Persistence:     true,
	}
}

func (cm *CacheManager) loadConfig() error {
	configFile := filepath.Join(cm.configDir, "cache.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // Use default config
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config CacheConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cm.config = &config
	return nil
}

func (cm *CacheManager) saveConfig() error {
	configFile := filepath.Join(cm.configDir, "cache.json")

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

func (cm *CacheManager) loadCache() error {
	if !cm.config.Persistence {
		return nil
	}

	cacheFile := filepath.Join(cm.configDir, "cache_data.json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty cache
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache map[string]*CacheEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	cm.cache = cache

	// Recalculate stats
	cm.stats.CurrentSize = 0
	cm.stats.TotalEntries = int64(len(cache))
	for _, entry := range cache {
		cm.stats.CurrentSize += entry.Size
	}

	return nil
}

func (cm *CacheManager) saveCache() error {
	if !cm.config.Persistence {
		return nil
	}

	cacheFile := filepath.Join(cm.configDir, "cache_data.json")

	data, err := json.MarshalIndent(cm.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	return os.WriteFile(cacheFile, data, 0644)
}

func (cm *CacheManager) loadStats() error {
	statsFile := filepath.Join(cm.configDir, "cache_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil // Use default stats
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats CacheStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	cm.stats = &stats
	return nil
}

func (cm *CacheManager) saveStats() error {
	statsFile := filepath.Join(cm.configDir, "cache_stats.json")

	data, err := json.MarshalIndent(cm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
