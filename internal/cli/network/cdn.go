package network

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// CDNManager manages Content Delivery Network operations
type CDNManager struct {
	client    *client.Client
	configDir string
	nodes     map[string]*CDNNode
	config    *CDNConfig
	stats     *CDNStats
	mu        sync.RWMutex
}

// CDNNode represents a CDN edge node
type CDNNode struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Location       string                 `json:"location"`
	URL            string                 `json:"url"`
	Region         string                 `json:"region"`
	Country        string                 `json:"country"`
	Latitude       float64                `json:"latitude"`
	Longitude      float64                `json:"longitude"`
	IsActive       bool                   `json:"is_active"`
	Bandwidth      int64                  `json:"bandwidth"`    // in bytes per second
	Storage        int64                  `json:"storage"`      // in bytes
	UsedStorage    int64                  `json:"used_storage"` // in bytes
	LastSync       time.Time              `json:"last_sync"`
	ResponseTime   int64                  `json:"response_time"` // in milliseconds
	CacheHitRate   float64                `json:"cache_hit_rate"`
	TotalRequests  int64                  `json:"total_requests"`
	FailedRequests int64                  `json:"failed_requests"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CDNConfig represents CDN configuration
type CDNConfig struct {
	AutoSync         bool          `json:"auto_sync"`
	SyncInterval     time.Duration `json:"sync_interval"`
	CacheTTL         time.Duration `json:"cache_ttl"`
	MaxBandwidth     int64         `json:"max_bandwidth"`     // in bytes per second
	MaxStorage       int64         `json:"max_storage"`       // in bytes
	CompressionLevel int           `json:"compression_level"` // 1-9
	EnableSSL        bool          `json:"enable_ssl"`
	EnableGzip       bool          `json:"enable_gzip"`
	EnableBrotli     bool          `json:"enable_brotli"`
}

// CDNStats represents CDN statistics
type CDNStats struct {
	TotalNodes          int       `json:"total_nodes"`
	ActiveNodes         int       `json:"active_nodes"`
	InactiveNodes       int       `json:"inactive_nodes"`
	TotalBandwidth      int64     `json:"total_bandwidth"` // in bytes per second
	UsedBandwidth       int64     `json:"used_bandwidth"`  // in bytes per second
	TotalStorage        int64     `json:"total_storage"`   // in bytes
	UsedStorage         int64     `json:"used_storage"`    // in bytes
	TotalRequests       int64     `json:"total_requests"`
	CacheHits           int64     `json:"cache_hits"`
	CacheMisses         int64     `json:"cache_misses"`
	CacheHitRate        float64   `json:"cache_hit_rate"`
	AverageResponseTime float64   `json:"average_response_time"`
	LastUpdated         time.Time `json:"last_updated"`
}

// CDNFile represents a file in the CDN
type CDNFile struct {
	FileID       string                 `json:"file_id"`
	NodeID       string                 `json:"node_id"`
	Size         int64                  `json:"size"`
	Checksum     string                 `json:"checksum"`
	CachedAt     time.Time              `json:"cached_at"`
	LastAccessed time.Time              `json:"last_accessed"`
	AccessCount  int64                  `json:"access_count"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewCDNManager creates a new CDN manager
func NewCDNManager(client *client.Client, configDir string) *CDNManager {
	cdn := &CDNManager{
		client:    client,
		configDir: configDir,
		nodes:     make(map[string]*CDNNode),
		config:    getDefaultCDNConfig(),
		stats:     &CDNStats{},
	}

	cdn.loadConfig()
	cdn.loadNodes()
	cdn.loadStats()

	// Start sync routine
	go cdn.startSyncRoutine()

	return cdn
}

// AddNode adds a new CDN node
func (cdn *CDNManager) AddNode(node *CDNNode) error {
	cdn.mu.Lock()
	defer cdn.mu.Unlock()

	// Check if node already exists
	if _, exists := cdn.nodes[node.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", node.ID)
	}

	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	node.IsActive = true

	cdn.nodes[node.ID] = node
	cdn.saveNodes()
	cdn.updateStats()

	return nil
}

// RemoveNode removes a CDN node
func (cdn *CDNManager) RemoveNode(nodeID string) error {
	cdn.mu.Lock()
	defer cdn.mu.Unlock()

	if _, exists := cdn.nodes[nodeID]; !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	delete(cdn.nodes, nodeID)
	cdn.saveNodes()
	cdn.updateStats()

	return nil
}

// GetNode returns a CDN node by ID
func (cdn *CDNManager) GetNode(nodeID string) (*CDNNode, error) {
	cdn.mu.RLock()
	defer cdn.mu.RUnlock()

	node, exists := cdn.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node with ID %s not found", nodeID)
	}

	// Return a copy
	nodeCopy := *node
	return &nodeCopy, nil
}

// ListNodes returns all CDN nodes
func (cdn *CDNManager) ListNodes() []*CDNNode {
	cdn.mu.RLock()
	defer cdn.mu.RUnlock()

	var nodes []*CDNNode
	for _, node := range cdn.nodes {
		// Return a copy
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes
}

// GetNearestNode returns the nearest CDN node to a location
func (cdn *CDNManager) GetNearestNode(latitude, longitude float64) (*CDNNode, error) {
	cdn.mu.RLock()
	defer cdn.mu.RUnlock()

	if len(cdn.nodes) == 0 {
		return nil, fmt.Errorf("no CDN nodes available")
	}

	var nearestNode *CDNNode
	minDistance := float64(999999999)

	for _, node := range cdn.nodes {
		if !node.IsActive {
			continue
		}

		// Calculate distance (simplified)
		distance := cdn.calculateDistance(latitude, longitude, node.Latitude, node.Longitude)
		if distance < minDistance {
			minDistance = distance
			nearestNode = node
		}
	}

	if nearestNode == nil {
		return nil, fmt.Errorf("no active CDN nodes available")
	}

	// Return a copy
	nodeCopy := *nearestNode
	return &nodeCopy, nil
}

// CacheFile caches a file on a CDN node
func (cdn *CDNManager) CacheFile(fileID, nodeID string) error {
	cdn.mu.Lock()
	defer cdn.mu.Unlock()

	node, exists := cdn.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	if !node.IsActive {
		return fmt.Errorf("node %s is not active", nodeID)
	}

	// Check storage capacity
	fileSize := int64(1024 * 1024) // Assume 1MB for demo
	if node.UsedStorage+fileSize > node.Storage {
		return fmt.Errorf("insufficient storage on node %s", nodeID)
	}

	// Create CDN file entry
	cdnFile := &CDNFile{
		FileID:       fileID,
		NodeID:       nodeID,
		Size:         fileSize,
		Checksum:     fmt.Sprintf("checksum_%s", fileID),
		CachedAt:     time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  1,
		ExpiresAt:    time.Now().Add(cdn.config.CacheTTL),
		Metadata:     make(map[string]interface{}),
	}

	// Update node storage
	node.UsedStorage += fileSize
	node.UpdatedAt = time.Now()

	// Save CDN file (simplified - in real implementation, you'd save to a separate store)
	cdn.saveCDNFile(cdnFile)
	cdn.saveNodes()

	return nil
}

// GetCachedFile retrieves a cached file from CDN
func (cdn *CDNManager) GetCachedFile(fileID string) (*CDNFile, error) {
	cdn.mu.RLock()
	defer cdn.mu.RUnlock()

	// Find the file in CDN nodes
	for _, node := range cdn.nodes {
		if !node.IsActive {
			continue
		}

		// Check if file is cached on this node (simplified)
		cdnFile, err := cdn.getCDNFile(fileID, node.ID)
		if err == nil && time.Now().Before(cdnFile.ExpiresAt) {
			// Update access statistics
			cdnFile.LastAccessed = time.Now()
			cdnFile.AccessCount++
			node.TotalRequests++
			node.CacheHitRate = float64(node.TotalRequests) / float64(node.TotalRequests+node.FailedRequests)

			cdn.saveCDNFile(cdnFile)
			cdn.saveNodes()
			cdn.updateStats()

			return cdnFile, nil
		}
	}

	return nil, fmt.Errorf("file %s not found in CDN cache", fileID)
}

// SyncNode synchronizes a CDN node with the main server
func (cdn *CDNManager) SyncNode(nodeID string) error {
	cdn.mu.Lock()
	defer cdn.mu.Unlock()

	node, exists := cdn.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node with ID %s not found", nodeID)
	}

	// Simulate sync operation
	start := time.Now()
	time.Sleep(time.Millisecond * 100) // Simulate network delay
	responseTime := time.Since(start).Milliseconds()

	node.LastSync = time.Now()
	node.ResponseTime = responseTime
	node.UpdatedAt = time.Now()

	cdn.saveNodes()

	return nil
}

// GetStats returns CDN statistics
func (cdn *CDNManager) GetStats() *CDNStats {
	cdn.mu.RLock()
	defer cdn.mu.RUnlock()

	// Update current stats
	cdn.updateStats()

	// Return a copy
	stats := *cdn.stats
	return &stats
}

// UpdateConfig updates CDN configuration
func (cdn *CDNManager) UpdateConfig(config *CDNConfig) error {
	cdn.mu.Lock()
	defer cdn.mu.Unlock()

	cdn.config = config
	cdn.saveConfig()

	return nil
}

// GetConfig returns current CDN configuration
func (cdn *CDNManager) GetConfig() *CDNConfig {
	cdn.mu.RLock()
	defer cdn.mu.RUnlock()

	// Return a copy
	config := *cdn.config
	return &config
}

// Utility methods
func (cdn *CDNManager) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Simplified distance calculation (Haversine formula would be more accurate)
	dx := lat2 - lat1
	dy := lon2 - lon1
	return dx*dx + dy*dy
}

func (cdn *CDNManager) updateStats() {
	cdn.stats.TotalNodes = len(cdn.nodes)
	cdn.stats.ActiveNodes = 0
	cdn.stats.InactiveNodes = 0
	cdn.stats.TotalBandwidth = 0
	cdn.stats.UsedBandwidth = 0
	cdn.stats.TotalStorage = 0
	cdn.stats.UsedStorage = 0
	cdn.stats.TotalRequests = 0
	cdn.stats.CacheHits = 0
	cdn.stats.CacheMisses = 0

	for _, node := range cdn.nodes {
		if node.IsActive {
			cdn.stats.ActiveNodes++
		} else {
			cdn.stats.InactiveNodes++
		}

		cdn.stats.TotalBandwidth += node.Bandwidth
		cdn.stats.TotalStorage += node.Storage
		cdn.stats.UsedStorage += node.UsedStorage
		cdn.stats.TotalRequests += node.TotalRequests
		cdn.stats.CacheHits += node.TotalRequests
		cdn.stats.CacheMisses += node.FailedRequests
	}

	// Calculate cache hit rate
	if cdn.stats.TotalRequests > 0 {
		cdn.stats.CacheHitRate = float64(cdn.stats.CacheHits) / float64(cdn.stats.TotalRequests)
	}

	cdn.stats.LastUpdated = time.Now()
}

// Sync routine
func (cdn *CDNManager) startSyncRoutine() {
	if !cdn.config.AutoSync {
		return
	}

	ticker := time.NewTicker(cdn.config.SyncInterval)
	defer ticker.Stop()

	for range ticker.C {
		cdn.syncAllNodes()
	}
}

func (cdn *CDNManager) syncAllNodes() {
	cdn.mu.RLock()
	nodes := make([]*CDNNode, 0, len(cdn.nodes))
	for _, node := range cdn.nodes {
		nodes = append(nodes, node)
	}
	cdn.mu.RUnlock()

	for _, node := range nodes {
		if node.IsActive {
			go cdn.SyncNode(node.ID)
		}
	}
}

// CDN file management
func (cdn *CDNManager) saveCDNFile(cdnFile *CDNFile) error {
	cdnFilesFile := filepath.Join(cdn.configDir, fmt.Sprintf("cdn_file_%s_%s.json", cdnFile.FileID, cdnFile.NodeID))

	data, err := json.MarshalIndent(cdnFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal CDN file: %w", err)
	}

	return os.WriteFile(cdnFilesFile, data, 0644)
}

func (cdn *CDNManager) getCDNFile(fileID, nodeID string) (*CDNFile, error) {
	cdnFilesFile := filepath.Join(cdn.configDir, fmt.Sprintf("cdn_file_%s_%s.json", fileID, nodeID))

	if _, err := os.Stat(cdnFilesFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("CDN file not found")
	}

	data, err := os.ReadFile(cdnFilesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CDN file: %w", err)
	}

	var cdnFile CDNFile
	if err := json.Unmarshal(data, &cdnFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CDN file: %w", err)
	}

	return &cdnFile, nil
}

// Configuration management
func getDefaultCDNConfig() *CDNConfig {
	return &CDNConfig{
		AutoSync:         true,
		SyncInterval:     5 * time.Minute,
		CacheTTL:         1 * time.Hour,
		MaxBandwidth:     100 * 1024 * 1024,       // 100MB/s
		MaxStorage:       10 * 1024 * 1024 * 1024, // 10GB
		CompressionLevel: 6,
		EnableSSL:        true,
		EnableGzip:       true,
		EnableBrotli:     false,
	}
}

func (cdn *CDNManager) loadConfig() error {
	configFile := filepath.Join(cdn.configDir, "cdn.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // Use default config
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config CDNConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cdn.config = &config
	return nil
}

func (cdn *CDNManager) saveConfig() error {
	configFile := filepath.Join(cdn.configDir, "cdn.json")

	data, err := json.MarshalIndent(cdn.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configFile, data, 0644)
}

func (cdn *CDNManager) loadNodes() error {
	nodesFile := filepath.Join(cdn.configDir, "cdn_nodes.json")
	if _, err := os.Stat(nodesFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty nodes
	}

	data, err := os.ReadFile(nodesFile)
	if err != nil {
		return fmt.Errorf("failed to read nodes file: %w", err)
	}

	var nodes map[string]*CDNNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	cdn.nodes = nodes
	return nil
}

func (cdn *CDNManager) saveNodes() error {
	nodesFile := filepath.Join(cdn.configDir, "cdn_nodes.json")

	data, err := json.MarshalIndent(cdn.nodes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	return os.WriteFile(nodesFile, data, 0644)
}

func (cdn *CDNManager) loadStats() error {
	statsFile := filepath.Join(cdn.configDir, "cdn_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil // Use default stats
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats CDNStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	cdn.stats = &stats
	return nil
}

func (cdn *CDNManager) saveStats() error {
	statsFile := filepath.Join(cdn.configDir, "cdn_stats.json")

	data, err := json.MarshalIndent(cdn.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
