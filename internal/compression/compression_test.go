package compression

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressionConstants(t *testing.T) {
	// Test compression type constants
	assert.Equal(t, CompressionType("none"), CompressionTypeNone)
	assert.Equal(t, CompressionType("gzip"), CompressionTypeGzip)
	assert.Equal(t, CompressionType("zlib"), CompressionTypeZlib)

	// Test compression level constants
	assert.Equal(t, CompressionLevel(0), CompressionLevelFast)
	assert.Equal(t, CompressionLevel(1), CompressionLevelDefault)
	assert.Equal(t, CompressionLevel(2), CompressionLevelBest)
}

func TestGzipCompressor(t *testing.T) {
	tests := []struct {
		name  string
		level CompressionLevel
		data  []byte
	}{
		{
			name:  "fast level",
			level: CompressionLevelFast,
			data:  []byte("Hello, World! This is a test string for compression."),
		},
		{
			name:  "default level",
			level: CompressionLevelDefault,
			data:  []byte("Hello, World! This is a test string for compression."),
		},
		{
			name:  "best level",
			level: CompressionLevelBest,
			data:  []byte("Hello, World! This is a test string for compression."),
		},
		{
			name:  "empty data",
			level: CompressionLevelDefault,
			data:  []byte(""),
		},
		{
			name:  "single character",
			level: CompressionLevelDefault,
			data:  []byte("a"),
		},
		{
			name:  "large data",
			level: CompressionLevelDefault,
			data:  []byte("This is a very long string that should compress well. " + 
				"Repeating this text multiple times to make it longer. " +
				"This is a very long string that should compress well. " +
				"Repeating this text multiple times to make it longer. " +
				"This is a very long string that should compress well. " +
				"Repeating this text multiple times to make it longer."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewGzipCompressor(tt.level)
			ctx := context.Background()

			// Test GetType
			assert.Equal(t, CompressionTypeGzip, compressor.GetType())

			// Test GetLevel
			assert.Equal(t, tt.level, compressor.GetLevel())

			// Test Compress
			compressed, err := compressor.Compress(ctx, tt.data)
			assert.NoError(t, err)
			assert.NotNil(t, compressed)

			// Test Decompress
			decompressed, err := compressor.Decompress(ctx, compressed)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, decompressed)

			// Test round-trip
			recompressed, err := compressor.Compress(ctx, decompressed)
			assert.NoError(t, err)
			assert.Equal(t, compressed, recompressed)
		})
	}
}

func TestGzipCompressor_InvalidData(t *testing.T) {
	compressor := NewGzipCompressor(CompressionLevelDefault)
	ctx := context.Background()

	// Test decompressing invalid data
	invalidData := []byte("This is not valid gzip data")
	_, err := compressor.Decompress(ctx, invalidData)
	assert.Error(t, err)
}

func TestZlibCompressor(t *testing.T) {
	tests := []struct {
		name  string
		level CompressionLevel
		data  []byte
	}{
		{
			name:  "fast level",
			level: CompressionLevelFast,
			data:  []byte("Hello, World! This is a test string for compression."),
		},
		{
			name:  "default level",
			level: CompressionLevelDefault,
			data:  []byte("Hello, World! This is a test string for compression."),
		},
		{
			name:  "best level",
			level: CompressionLevelBest,
			data:  []byte("Hello, World! This is a test string for compression."),
		},
		{
			name:  "empty data",
			level: CompressionLevelDefault,
			data:  []byte(""),
		},
		{
			name:  "single character",
			level: CompressionLevelDefault,
			data:  []byte("a"),
		},
		{
			name:  "large data",
			level: CompressionLevelDefault,
			data:  []byte("This is a very long string that should compress well. " + 
				"Repeating this text multiple times to make it longer. " +
				"This is a very long string that should compress well. " +
				"Repeating this text multiple times to make it longer. " +
				"This is a very long string that should compress well. " +
				"Repeating this text multiple times to make it longer."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewZlibCompressor(tt.level)
			ctx := context.Background()

			// Test GetType
			assert.Equal(t, CompressionTypeZlib, compressor.GetType())

			// Test GetLevel
			assert.Equal(t, tt.level, compressor.GetLevel())

			// Test Compress
			compressed, err := compressor.Compress(ctx, tt.data)
			assert.NoError(t, err)
			assert.NotNil(t, compressed)

			// Test Decompress
			decompressed, err := compressor.Decompress(ctx, compressed)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, decompressed)

			// Test round-trip
			recompressed, err := compressor.Compress(ctx, decompressed)
			assert.NoError(t, err)
			assert.Equal(t, compressed, recompressed)
		})
	}
}

func TestZlibCompressor_InvalidData(t *testing.T) {
	compressor := NewZlibCompressor(CompressionLevelDefault)
	ctx := context.Background()

	// Test decompressing invalid data
	invalidData := []byte("This is not valid zlib data")
	_, err := compressor.Decompress(ctx, invalidData)
	assert.Error(t, err)
}

func TestCompressionManager(t *testing.T) {
	manager := NewCompressionManager()
	ctx := context.Background()

	// Test initial state
	availableTypes := manager.GetAvailableTypes()
	assert.Len(t, availableTypes, 2)
	assert.Contains(t, availableTypes, CompressionTypeGzip)
	assert.Contains(t, availableTypes, CompressionTypeZlib)

	// Test data
	testData := []byte("Hello, World! This is a test string for compression.")

	// Test gzip compression
	compressedGzip, err := manager.Compress(ctx, testData, CompressionTypeGzip)
	assert.NoError(t, err)
	assert.NotNil(t, compressedGzip)

	decompressedGzip, err := manager.Decompress(ctx, compressedGzip, CompressionTypeGzip)
	assert.NoError(t, err)
	assert.Equal(t, testData, decompressedGzip)

	// Test zlib compression
	compressedZlib, err := manager.Compress(ctx, testData, CompressionTypeZlib)
	assert.NoError(t, err)
	assert.NotNil(t, compressedZlib)

	decompressedZlib, err := manager.Decompress(ctx, compressedZlib, CompressionTypeZlib)
	assert.NoError(t, err)
	assert.Equal(t, testData, decompressedZlib)

	// Test that gzip and zlib produce different compressed data
	assert.NotEqual(t, compressedGzip, compressedZlib)
}

func TestCompressionManager_RegisterCompressor(t *testing.T) {
	manager := NewCompressionManager()

	// Test registering a new compressor
	customGzip := NewGzipCompressor(CompressionLevelBest)
	manager.RegisterCompressor(customGzip)

	// Test that the compressor is available
	compressor, err := manager.GetCompressor(CompressionTypeGzip)
	assert.NoError(t, err)
	assert.Equal(t, CompressionLevelBest, compressor.GetLevel())

	// Test that available types still includes gzip
	availableTypes := manager.GetAvailableTypes()
	assert.Contains(t, availableTypes, CompressionTypeGzip)
}

func TestCompressionManager_GetCompressor_NotFound(t *testing.T) {
	manager := NewCompressionManager()

	// Test getting a non-existent compressor
	_, err := manager.GetCompressor(CompressionTypeNone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compressor type none not found")
}

func TestCompressionManager_Compress_NotFound(t *testing.T) {
	manager := NewCompressionManager()
	ctx := context.Background()

	// Test compressing with a non-existent compressor
	_, err := manager.Compress(ctx, []byte("test"), CompressionTypeNone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compressor type none not found")
}

func TestCompressionManager_Decompress_NotFound(t *testing.T) {
	manager := NewCompressionManager()
	ctx := context.Background()

	// Test decompressing with a non-existent compressor
	_, err := manager.Decompress(ctx, []byte("test"), CompressionTypeNone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compressor type none not found")
}

func TestCompressionManager_ConcurrentAccess(t *testing.T) {
	manager := NewCompressionManager()
	ctx := context.Background()

	// Test concurrent access to the compression manager
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Test concurrent compression
			data := []byte("Concurrent test data")
			compressed, err := manager.Compress(ctx, data, CompressionTypeGzip)
			assert.NoError(t, err)
			assert.NotNil(t, compressed)
			
			// Test concurrent decompression
			decompressed, err := manager.Decompress(ctx, compressed, CompressionTypeGzip)
			assert.NoError(t, err)
			assert.Equal(t, data, decompressed)
			
			// Test concurrent access to available types
			types := manager.GetAvailableTypes()
			assert.Len(t, types, 2)
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCompressionManager_EdgeCases(t *testing.T) {
	manager := NewCompressionManager()
	ctx := context.Background()

	// Test with nil data
	compressed, err := manager.Compress(ctx, nil, CompressionTypeGzip)
	assert.NoError(t, err)
	assert.NotNil(t, compressed)

	decompressed, err := manager.Decompress(ctx, compressed, CompressionTypeGzip)
	assert.NoError(t, err)
	assert.Nil(t, decompressed)

	// Test with very small data
	smallData := []byte("a")
	compressed, err = manager.Compress(ctx, smallData, CompressionTypeGzip)
	assert.NoError(t, err)
	assert.NotNil(t, compressed)

	decompressed, err = manager.Decompress(ctx, compressed, CompressionTypeGzip)
	assert.NoError(t, err)
	assert.Equal(t, smallData, decompressed)
}

func TestCompressionManager_GetAvailableTypes(t *testing.T) {
	manager := NewCompressionManager()

	// Test initial available types
	types := manager.GetAvailableTypes()
	assert.Len(t, types, 2)
	assert.Contains(t, types, CompressionTypeGzip)
	assert.Contains(t, types, CompressionTypeZlib)

	// Test that the slice is a copy (not the original map)
	types[0] = CompressionTypeNone
	typesAgain := manager.GetAvailableTypes()
	assert.Len(t, typesAgain, 2)
	assert.Contains(t, typesAgain, CompressionTypeGzip)
	assert.Contains(t, typesAgain, CompressionTypeZlib)
}
