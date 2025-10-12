package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// DeduplicationManager manages file deduplication
type DeduplicationManager struct {
	client    *client.Client
	configDir string
	chunks    map[string]*ChunkInfo
	stats     *DeduplicationStats
	mu        sync.RWMutex
}

// ChunkInfo represents information about a data chunk
type ChunkInfo struct {
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	RefCount     int       `json:"ref_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	FileIDs      []string  `json:"file_ids"`
}

// DeduplicationStats represents deduplication statistics
type DeduplicationStats struct {
	TotalFiles         int64     `json:"total_files"`
	UniqueChunks       int64     `json:"unique_chunks"`
	DuplicateChunks    int64     `json:"duplicate_chunks"`
	TotalSize          int64     `json:"total_size"`
	DeduplicatedSize   int64     `json:"deduplicated_size"`
	SpaceSaved         int64     `json:"space_saved"`
	DeduplicationRatio float64   `json:"deduplication_ratio"`
	LastUpdated        time.Time `json:"last_updated"`
}

// DeduplicationResult represents the result of a deduplication operation
type DeduplicationResult struct {
	FileID             string        `json:"file_id"`
	OriginalSize       int64         `json:"original_size"`
	DeduplicatedSize   int64         `json:"deduplicated_size"`
	ChunksCreated      int           `json:"chunks_created"`
	ChunksReused       int           `json:"chunks_reused"`
	SpaceSaved         int64         `json:"space_saved"`
	DeduplicationRatio float64       `json:"deduplication_ratio"`
	TimeTaken          time.Duration `json:"time_taken"`
	Success            bool          `json:"success"`
	Error              string        `json:"error,omitempty"`
}

// ChunkSize represents the size of chunks for deduplication
type ChunkSize int64

const (
	ChunkSizeSmall  ChunkSize = 64 * 1024   // 64KB
	ChunkSizeMedium ChunkSize = 256 * 1024  // 256KB
	ChunkSizeLarge  ChunkSize = 1024 * 1024 // 1MB
)

// NewDeduplicationManager creates a new deduplication manager
func NewDeduplicationManager(client *client.Client, configDir string) *DeduplicationManager {
	dm := &DeduplicationManager{
		client:    client,
		configDir: configDir,
		chunks:    make(map[string]*ChunkInfo),
		stats:     &DeduplicationStats{},
	}

	dm.loadChunks()
	dm.loadStats()
	return dm
}

// DeduplicateFile deduplicates a file by chunking it
func (dm *DeduplicationManager) DeduplicateFile(fileID string, chunkSize ChunkSize) (*DeduplicationResult, error) {
	start := time.Now()

	// Create temporary file for download
	tempFile := fmt.Sprintf("/tmp/dedup_%s", fileID)
	err := dm.client.DownloadFile(context.Background(), fileID, tempFile)
	if err != nil {
		return &DeduplicationResult{
			Success: false,
			Error:   fmt.Sprintf("failed to download file: %v", err),
		}, err
	}

	// Read file data
	fileData, err := os.ReadFile(tempFile)
	if err != nil {
		os.Remove(tempFile) // Clean up
		return &DeduplicationResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}
	os.Remove(tempFile) // Clean up

	originalSize := int64(len(fileData))

	// Chunk the file
	chunks, err := dm.chunkData(fileData, int64(chunkSize))
	if err != nil {
		return &DeduplicationResult{
			Success: false,
			Error:   fmt.Sprintf("failed to chunk file: %v", err),
		}, err
	}

	// Process chunks
	chunksCreated := 0
	chunksReused := 0
	deduplicatedSize := int64(0)

	for _, chunk := range chunks {
		chunkHash := dm.calculateHash(chunk)

		dm.mu.Lock()
		chunkInfo, exists := dm.chunks[chunkHash]
		if exists {
			// Chunk already exists, increment reference count
			chunkInfo.RefCount++
			chunkInfo.LastAccessed = time.Now()
			chunkInfo.FileIDs = append(chunkInfo.FileIDs, fileID)
			chunksReused++
			deduplicatedSize += int64(len(chunk))
		} else {
			// New chunk, store it
			dm.chunks[chunkHash] = &ChunkInfo{
				Hash:         chunkHash,
				Size:         int64(len(chunk)),
				RefCount:     1,
				CreatedAt:    time.Now(),
				LastAccessed: time.Now(),
				FileIDs:      []string{fileID},
			}
			chunksCreated++
			deduplicatedSize += int64(len(chunk))
		}
		dm.mu.Unlock()
	}

	spaceSaved := originalSize - deduplicatedSize
	deduplicationRatio := float64(deduplicatedSize) / float64(originalSize)
	timeTaken := time.Since(start)

	result := &DeduplicationResult{
		FileID:             fileID,
		OriginalSize:       originalSize,
		DeduplicatedSize:   deduplicatedSize,
		ChunksCreated:      chunksCreated,
		ChunksReused:       chunksReused,
		SpaceSaved:         spaceSaved,
		DeduplicationRatio: deduplicationRatio,
		TimeTaken:          timeTaken,
		Success:            true,
	}

	// Update statistics
	dm.updateStats(result)
	dm.saveChunks()

	return result, nil
}

// ReconstructFile reconstructs a file from its chunks
func (dm *DeduplicationManager) ReconstructFile(fileID string) ([]byte, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Find all chunks for this file
	var fileChunks [][]byte
	for _, chunkInfo := range dm.chunks {
		for _, id := range chunkInfo.FileIDs {
			if id == fileID {
				// Get chunk data (in real implementation, this would fetch from storage)
				chunkData := make([]byte, chunkInfo.Size)
				fileChunks = append(fileChunks, chunkData)
				break
			}
		}
	}

	if len(fileChunks) == 0 {
		return nil, fmt.Errorf("no chunks found for file: %s", fileID)
	}

	// Reconstruct file by concatenating chunks
	var reconstructed []byte
	for _, chunk := range fileChunks {
		reconstructed = append(reconstructed, chunk...)
	}

	return reconstructed, nil
}

// GetDeduplicationStats returns deduplication statistics
func (dm *DeduplicationManager) GetDeduplicationStats() *DeduplicationStats {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Return a copy
	stats := *dm.stats
	return &stats
}

// GetChunkInfo returns information about a specific chunk
func (dm *DeduplicationManager) GetChunkInfo(chunkHash string) (*ChunkInfo, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	chunkInfo, exists := dm.chunks[chunkHash]
	if !exists {
		return nil, fmt.Errorf("chunk not found: %s", chunkHash)
	}

	// Return a copy
	info := *chunkInfo
	return &info, nil
}

// ListChunks returns all chunks with their information
func (dm *DeduplicationManager) ListChunks() []*ChunkInfo {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	var chunks []*ChunkInfo
	for _, chunkInfo := range dm.chunks {
		// Return a copy
		info := *chunkInfo
		chunks = append(chunks, &info)
	}

	return chunks
}

// RemoveFileChunks removes chunks associated with a file
func (dm *DeduplicationManager) RemoveFileChunks(fileID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for hash, chunkInfo := range dm.chunks {
		// Remove file ID from chunk
		var newFileIDs []string
		for _, id := range chunkInfo.FileIDs {
			if id != fileID {
				newFileIDs = append(newFileIDs, id)
			}
		}

		if len(newFileIDs) == 0 {
			// No more references, remove chunk
			delete(dm.chunks, hash)
		} else {
			// Update file IDs and decrement reference count
			chunkInfo.FileIDs = newFileIDs
			chunkInfo.RefCount--
		}
	}

	dm.saveChunks()
	return nil
}

// CleanupUnusedChunks removes chunks with no references
func (dm *DeduplicationManager) CleanupUnusedChunks() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var toDelete []string
	for hash, chunkInfo := range dm.chunks {
		if chunkInfo.RefCount <= 0 {
			toDelete = append(toDelete, hash)
		}
	}

	for _, hash := range toDelete {
		delete(dm.chunks, hash)
	}

	if len(toDelete) > 0 {
		dm.saveChunks()
	}

	return nil
}

// ResetStats resets deduplication statistics
func (dm *DeduplicationManager) ResetStats() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.stats = &DeduplicationStats{
		LastUpdated: time.Now(),
	}
	dm.saveStats()
}

// Utility functions
func (dm *DeduplicationManager) chunkData(data []byte, chunkSize int64) ([][]byte, error) {
	var chunks [][]byte

	for i := int64(0); i < int64(len(data)); i += chunkSize {
		end := i + chunkSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}
		chunks = append(chunks, data[i:end])
	}

	return chunks, nil
}

func (dm *DeduplicationManager) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Statistics management
func (dm *DeduplicationManager) updateStats(result *DeduplicationResult) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.stats.TotalFiles++
	dm.stats.TotalSize += result.OriginalSize
	dm.stats.DeduplicatedSize += result.DeduplicatedSize
	dm.stats.SpaceSaved += result.SpaceSaved
	dm.stats.UniqueChunks += int64(result.ChunksCreated)
	dm.stats.DuplicateChunks += int64(result.ChunksReused)

	if dm.stats.TotalSize > 0 {
		dm.stats.DeduplicationRatio = float64(dm.stats.DeduplicatedSize) / float64(dm.stats.TotalSize)
	}

	dm.stats.LastUpdated = time.Now()
	dm.saveStats()
}

// Data persistence
func (dm *DeduplicationManager) loadChunks() error {
	chunksFile := filepath.Join(dm.configDir, "chunks.json")
	if _, err := os.Stat(chunksFile); os.IsNotExist(err) {
		return nil // File doesn't exist, start with empty chunks
	}

	data, err := os.ReadFile(chunksFile)
	if err != nil {
		return fmt.Errorf("failed to read chunks file: %w", err)
	}

	var chunks map[string]*ChunkInfo
	if err := json.Unmarshal(data, &chunks); err != nil {
		return fmt.Errorf("failed to unmarshal chunks: %w", err)
	}

	dm.chunks = chunks
	return nil
}

func (dm *DeduplicationManager) saveChunks() error {
	chunksFile := filepath.Join(dm.configDir, "chunks.json")

	data, err := json.MarshalIndent(dm.chunks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chunks: %w", err)
	}

	return os.WriteFile(chunksFile, data, 0644)
}

func (dm *DeduplicationManager) loadStats() error {
	statsFile := filepath.Join(dm.configDir, "deduplication_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil // Use default stats
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats DeduplicationStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	dm.stats = &stats
	return nil
}

func (dm *DeduplicationManager) saveStats() error {
	statsFile := filepath.Join(dm.configDir, "deduplication_stats.json")

	data, err := json.MarshalIndent(dm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
