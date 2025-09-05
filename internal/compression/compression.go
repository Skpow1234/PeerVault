package compression

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"
	"sync"
)

// CompressionType represents the type of compression
type CompressionType string

const (
	CompressionTypeNone CompressionType = "none"
	CompressionTypeGzip CompressionType = "gzip"
	CompressionTypeZlib CompressionType = "zlib"
)

// CompressionLevel represents the compression level
type CompressionLevel int

const (
	CompressionLevelFast CompressionLevel = iota
	CompressionLevelDefault
	CompressionLevelBest
)

// Compressor interface defines compression operations
type Compressor interface {
	Compress(ctx context.Context, data []byte) ([]byte, error)
	Decompress(ctx context.Context, data []byte) ([]byte, error)
	GetType() CompressionType
	GetLevel() CompressionLevel
}

// GzipCompressor implements gzip compression
type GzipCompressor struct {
	level CompressionLevel
}

// NewGzipCompressor creates a new gzip compressor
func NewGzipCompressor(level CompressionLevel) *GzipCompressor {
	return &GzipCompressor{
		level: level,
	}
}

// Compress compresses data using gzip
func (gc *GzipCompressor) Compress(ctx context.Context, data []byte) ([]byte, error) {
	var buf bytes.Buffer

	var writer *gzip.Writer
	switch gc.level {
	case CompressionLevelFast:
		writer, _ = gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	case CompressionLevelBest:
		writer, _ = gzip.NewWriterLevel(&buf, gzip.BestCompression)
	default:
		writer = gzip.NewWriter(&buf)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses gzip data
func (gc *GzipCompressor) Decompress(ctx context.Context, data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// GetType returns the compression type
func (gc *GzipCompressor) GetType() CompressionType {
	return CompressionTypeGzip
}

// GetLevel returns the compression level
func (gc *GzipCompressor) GetLevel() CompressionLevel {
	return gc.level
}

// ZlibCompressor implements zlib compression
type ZlibCompressor struct {
	level CompressionLevel
}

// NewZlibCompressor creates a new zlib compressor
func NewZlibCompressor(level CompressionLevel) *ZlibCompressor {
	return &ZlibCompressor{
		level: level,
	}
}

// Compress compresses data using zlib
func (zc *ZlibCompressor) Compress(ctx context.Context, data []byte) ([]byte, error) {
	var buf bytes.Buffer

	var writer *zlib.Writer
	switch zc.level {
	case CompressionLevelFast:
		writer, _ = zlib.NewWriterLevel(&buf, zlib.BestSpeed)
	case CompressionLevelBest:
		writer, _ = zlib.NewWriterLevel(&buf, zlib.BestCompression)
	default:
		writer = zlib.NewWriter(&buf)
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses zlib data
func (zc *ZlibCompressor) Decompress(ctx context.Context, data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// GetType returns the compression type
func (zc *ZlibCompressor) GetType() CompressionType {
	return CompressionTypeZlib
}

// GetLevel returns the compression level
func (zc *ZlibCompressor) GetLevel() CompressionLevel {
	return zc.level
}

// CompressionManager manages multiple compressors
type CompressionManager struct {
	compressors map[CompressionType]Compressor
	mu          sync.RWMutex
}

// NewCompressionManager creates a new compression manager
func NewCompressionManager() *CompressionManager {
	cm := &CompressionManager{
		compressors: make(map[CompressionType]Compressor),
	}

	// Add default compressors
	cm.RegisterCompressor(NewGzipCompressor(CompressionLevelDefault))
	cm.RegisterCompressor(NewZlibCompressor(CompressionLevelDefault))

	return cm
}

// RegisterCompressor registers a new compressor
func (cm *CompressionManager) RegisterCompressor(compressor Compressor) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.compressors[compressor.GetType()] = compressor
}

// GetCompressor returns a compressor by type
func (cm *CompressionManager) GetCompressor(compressionType CompressionType) (Compressor, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	compressor, exists := cm.compressors[compressionType]
	if !exists {
		return nil, fmt.Errorf("compressor type %s not found", compressionType)
	}

	return compressor, nil
}

// Compress compresses data using the specified compression type
func (cm *CompressionManager) Compress(ctx context.Context, data []byte, compressionType CompressionType) ([]byte, error) {
	compressor, err := cm.GetCompressor(compressionType)
	if err != nil {
		return nil, err
	}

	return compressor.Compress(ctx, data)
}

// Decompress decompresses data using the specified compression type
func (cm *CompressionManager) Decompress(ctx context.Context, data []byte, compressionType CompressionType) ([]byte, error) {
	compressor, err := cm.GetCompressor(compressionType)
	if err != nil {
		return nil, err
	}

	return compressor.Decompress(ctx, data)
}

// GetAvailableTypes returns all available compression types
func (cm *CompressionManager) GetAvailableTypes() []CompressionType {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	types := make([]CompressionType, 0, len(cm.compressors))
	for compressionType := range cm.compressors {
		types = append(types, compressionType)
	}

	return types
}
