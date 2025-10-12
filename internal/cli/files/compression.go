package files

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/cli/client"
)

// CompressionManager manages file compression
type CompressionManager struct {
	client    *client.Client
	configDir string
	settings  *CompressionSettings
	stats     *CompressionStats
	mu        sync.RWMutex
}

// CompressionSettings represents compression configuration
type CompressionSettings struct {
	Enabled          bool   `json:"enabled"`
	Algorithm        string `json:"algorithm"` // "gzip", "zlib", "none"
	Level            int    `json:"level"`     // 1-9 for gzip/zlib
	MinSize          int64  `json:"min_size"`  // Minimum file size to compress
	MaxSize          int64  `json:"max_size"`  // Maximum file size to compress
	AutoCompress     bool   `json:"auto_compress"`
	CompressOnUpload bool   `json:"compress_on_upload"`
}

// CompressionStats represents compression statistics
type CompressionStats struct {
	TotalFiles          int64     `json:"total_files"`
	CompressedFiles     int64     `json:"compressed_files"`
	UncompressedFiles   int64     `json:"uncompressed_files"`
	TotalOriginalSize   int64     `json:"total_original_size"`
	TotalCompressedSize int64     `json:"total_compressed_size"`
	CompressionRatio    float64   `json:"compression_ratio"`
	SpaceSaved          int64     `json:"space_saved"`
	LastUpdated         time.Time `json:"last_updated"`
}

// CompressionResult represents the result of a compression operation
type CompressionResult struct {
	OriginalSize     int64         `json:"original_size"`
	CompressedSize   int64         `json:"compressed_size"`
	CompressionRatio float64       `json:"compression_ratio"`
	Algorithm        string        `json:"algorithm"`
	Level            int           `json:"level"`
	TimeTaken        time.Duration `json:"time_taken"`
	Success          bool          `json:"success"`
	Error            string        `json:"error,omitempty"`
}

// NewCompressionManager creates a new compression manager
func NewCompressionManager(client *client.Client, configDir string) *CompressionManager {
	cm := &CompressionManager{
		client:    client,
		configDir: configDir,
		settings:  getDefaultCompressionSettings(),
		stats:     &CompressionStats{},
	}

	cm.loadSettings()
	cm.loadStats()
	return cm
}

// CompressFile compresses a file
func (cm *CompressionManager) CompressFile(fileID string, algorithm string, level int) (*CompressionResult, error) {
	start := time.Now()

	// Create temporary file for download
	tempFile := fmt.Sprintf("/tmp/compress_%s", fileID)
	err := cm.client.DownloadFile(context.Background(), fileID, tempFile)
	if err != nil {
		return &CompressionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to download file: %v", err),
		}, err
	}

	// Read file data
	fileData, err := os.ReadFile(tempFile)
	if err != nil {
		os.Remove(tempFile) // Clean up
		return &CompressionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}
	os.Remove(tempFile) // Clean up

	originalSize := int64(len(fileData))

	// Compress data
	var compressedData []byte
	switch algorithm {
	case "gzip":
		compressedData, err = cm.compressGzip(fileData, level)
	case "zlib":
		compressedData, err = cm.compressZlib(fileData, level)
	default:
		return &CompressionResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported algorithm: %s", algorithm),
		}, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	if err != nil {
		return &CompressionResult{
			Success: false,
			Error:   fmt.Sprintf("compression failed: %v", err),
		}, err
	}

	compressedSize := int64(len(compressedData))
	compressionRatio := float64(compressedSize) / float64(originalSize)
	timeTaken := time.Since(start)

	result := &CompressionResult{
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		CompressionRatio: compressionRatio,
		Algorithm:        algorithm,
		Level:            level,
		TimeTaken:        timeTaken,
		Success:          true,
	}

	// Update statistics
	cm.updateStats(result)

	return result, nil
}

// DecompressFile decompresses a file
func (cm *CompressionManager) DecompressFile(fileID string, algorithm string) ([]byte, error) {
	// Create temporary file for download
	tempFile := fmt.Sprintf("/tmp/decompress_%s", fileID)
	err := cm.client.DownloadFile(context.Background(), fileID, tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// Read file data
	fileData, err := os.ReadFile(tempFile)
	if err != nil {
		os.Remove(tempFile) // Clean up
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	os.Remove(tempFile) // Clean up

	// Decompress data
	var decompressedData []byte
	switch algorithm {
	case "gzip":
		decompressedData, err = cm.decompressGzip(fileData)
	case "zlib":
		decompressedData, err = cm.decompressZlib(fileData)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressedData, nil
}

// ShouldCompress determines if a file should be compressed
func (cm *CompressionManager) ShouldCompress(fileSize int64, contentType string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.settings.Enabled {
		return false
	}

	// Check size limits
	if fileSize < cm.settings.MinSize || fileSize > cm.settings.MaxSize {
		return false
	}

	// Check if file type is already compressed
	compressedTypes := []string{
		"application/gzip",
		"application/zip",
		"application/x-rar-compressed",
		"application/x-7z-compressed",
		"image/jpeg",
		"image/png",
		"video/mp4",
		"video/avi",
		"audio/mp3",
		"audio/aac",
	}

	for _, compressedType := range compressedTypes {
		if contentType == compressedType {
			return false
		}
	}

	return true
}

// GetCompressionSettings returns current compression settings
func (cm *CompressionManager) GetCompressionSettings() *CompressionSettings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	settings := *cm.settings
	return &settings
}

// UpdateCompressionSettings updates compression settings
func (cm *CompressionManager) UpdateCompressionSettings(settings *CompressionSettings) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate settings
	if settings.Level < 1 || settings.Level > 9 {
		return fmt.Errorf("compression level must be between 1 and 9")
	}

	if settings.Algorithm != "gzip" && settings.Algorithm != "zlib" && settings.Algorithm != "none" {
		return fmt.Errorf("unsupported algorithm: %s", settings.Algorithm)
	}

	cm.settings = settings
	cm.saveSettings()

	return nil
}

// GetCompressionStats returns compression statistics
func (cm *CompressionManager) GetCompressionStats() *CompressionStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	stats := *cm.stats
	return &stats
}

// ResetStats resets compression statistics
func (cm *CompressionManager) ResetStats() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.stats = &CompressionStats{
		LastUpdated: time.Now(),
	}
	cm.saveStats()
}

// Compression algorithms
func (cm *CompressionManager) compressGzip(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (cm *CompressionManager) compressZlib(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (cm *CompressionManager) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (cm *CompressionManager) decompressZlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Statistics management
func (cm *CompressionManager) updateStats(result *CompressionResult) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.stats.TotalFiles++
	if result.Success {
		cm.stats.CompressedFiles++
		cm.stats.TotalOriginalSize += result.OriginalSize
		cm.stats.TotalCompressedSize += result.CompressedSize
		cm.stats.SpaceSaved += result.OriginalSize - result.CompressedSize

		if cm.stats.TotalOriginalSize > 0 {
			cm.stats.CompressionRatio = float64(cm.stats.TotalCompressedSize) / float64(cm.stats.TotalOriginalSize)
		}
	} else {
		cm.stats.UncompressedFiles++
	}

	cm.stats.LastUpdated = time.Now()
	cm.saveStats()
}

// Configuration management
func getDefaultCompressionSettings() *CompressionSettings {
	return &CompressionSettings{
		Enabled:          true,
		Algorithm:        "gzip",
		Level:            6,
		MinSize:          1024,              // 1KB
		MaxSize:          100 * 1024 * 1024, // 100MB
		AutoCompress:     true,
		CompressOnUpload: false,
	}
}

func (cm *CompressionManager) loadSettings() error {
	settingsFile := filepath.Join(cm.configDir, "compression.json")
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		return nil // Use default settings
	}

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings CompressionSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	cm.settings = &settings
	return nil
}

func (cm *CompressionManager) saveSettings() error {
	settingsFile := filepath.Join(cm.configDir, "compression.json")

	data, err := json.MarshalIndent(cm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	return os.WriteFile(settingsFile, data, 0644)
}

func (cm *CompressionManager) loadStats() error {
	statsFile := filepath.Join(cm.configDir, "compression_stats.json")
	if _, err := os.Stat(statsFile); os.IsNotExist(err) {
		return nil // Use default stats
	}

	data, err := os.ReadFile(statsFile)
	if err != nil {
		return fmt.Errorf("failed to read stats file: %w", err)
	}

	var stats CompressionStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	cm.stats = &stats
	return nil
}

func (cm *CompressionManager) saveStats() error {
	statsFile := filepath.Join(cm.configDir, "compression_stats.json")

	data, err := json.MarshalIndent(cm.stats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %w", err)
	}

	return os.WriteFile(statsFile, data, 0644)
}
