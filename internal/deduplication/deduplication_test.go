package deduplication

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDeduplicationConfig(t *testing.T) {
	config := DefaultDeduplicationConfig()

	assert.NotNil(t, config)
	assert.Equal(t, int64(64*1024), config.ChunkSize)
	assert.Equal(t, "sha256", config.Algorithm)
}

func TestNewDeduplicator(t *testing.T) {
	chunkStore := NewMemoryChunkStore()

	tests := []struct {
		name   string
		config *DeduplicationConfig
	}{
		{
			name:   "with config",
			config: &DeduplicationConfig{ChunkSize: 32 * 1024, Algorithm: "sha256"},
		},
		{
			name:   "with nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deduplicator := NewDeduplicator(tt.config, chunkStore)

			assert.NotNil(t, deduplicator)
			assert.NotNil(t, deduplicator.hasher)
			assert.Equal(t, chunkStore, deduplicator.chunkStore)
			assert.NotNil(t, deduplicator.chunkIndex)

			if tt.config != nil {
				assert.Equal(t, tt.config.ChunkSize, deduplicator.chunkSize)
			} else {
				assert.Equal(t, int64(64*1024), deduplicator.chunkSize) // Default
			}
		})
	}
}

func TestDeduplicator_ProcessFile(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int // Expected number of chunks
	}{
		{
			name:     "small file",
			data:     []byte("hello"),
			expected: 1,
		},
		{
			name:     "exactly one chunk",
			data:     []byte("1234567890"), // 10 bytes
			expected: 1,
		},
		{
			name:     "multiple chunks",
			data:     []byte("12345678901234567890"), // 20 bytes
			expected: 2,
		},
		{
			name:     "empty file",
			data:     []byte(""),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh deduplicator for each test case
			chunkStore := NewMemoryChunkStore()
			config := &DeduplicationConfig{
				ChunkSize: 10, // Small chunk size for testing
				Algorithm: "sha256",
			}
			deduplicator := NewDeduplicator(config, chunkStore)

			reader := bytes.NewReader(tt.data)
			chunks, err := deduplicator.ProcessFile(context.Background(), reader)

			assert.NoError(t, err)
			assert.Len(t, chunks, tt.expected)

			// Verify chunk properties
			for i, chunk := range chunks {
				assert.NotEmpty(t, chunk.ID)
				assert.NotEmpty(t, chunk.Hash)
				// For "multiple chunks" test, both chunks are references to the same deduplicated chunk
				if tt.name == "multiple chunks" {
					assert.Equal(t, int64(2), chunk.RefCount)
				} else {
					assert.Equal(t, int64(1), chunk.RefCount)
				}
				assert.NotNil(t, chunk.Data)

				// Verify chunk size (except for the last chunk which might be smaller)
				expectedSize := int64(len(tt.data) - i*10)
				if expectedSize > 10 {
					expectedSize = 10
				}
				assert.Equal(t, expectedSize, chunk.Size)
			}
		})
	}
}

func TestDeduplicator_ProcessFile_ContextCancellation(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, chunkStore)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reader := bytes.NewReader([]byte("hello world"))
	chunks, err := deduplicator.ProcessFile(ctx, reader)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, chunks)
}

func TestDeduplicator_ProcessFile_Deduplication(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, chunkStore)

	// Process the same data twice
	data := []byte("hello world")
	reader1 := bytes.NewReader(data)
	reader2 := bytes.NewReader(data)

	chunks1, err1 := deduplicator.ProcessFile(context.Background(), reader1)
	chunks2, err2 := deduplicator.ProcessFile(context.Background(), reader2)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Should have the same number of chunks
	assert.Len(t, chunks1, len(chunks2))

	// Chunks should be identical (same ID and hash)
	for i := range chunks1 {
		assert.Equal(t, chunks1[i].ID, chunks2[i].ID)
		assert.Equal(t, chunks1[i].Hash, chunks2[i].Hash)
		assert.Equal(t, int64(2), chunks1[i].RefCount) // Should be referenced twice
		assert.Equal(t, int64(2), chunks2[i].RefCount)
	}
}

func TestDeduplicator_ReconstructFile(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, chunkStore)

	// Process a file
	originalData := []byte("hello world this is a test")
	reader := bytes.NewReader(originalData)
	chunks, err := deduplicator.ProcessFile(context.Background(), reader)
	require.NoError(t, err)

	// Reconstruct the file
	reconstructedReader, err := deduplicator.ReconstructFile(context.Background(), chunks)
	require.NoError(t, err)

	// Read the reconstructed data
	reconstructedData, err := io.ReadAll(reconstructedReader)
	require.NoError(t, err)

	// Should match the original data
	assert.Equal(t, originalData, reconstructedData)
}

func TestDeduplicator_ReconstructFile_EmptyChunks(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	deduplicator := NewDeduplicator(nil, chunkStore)

	// Reconstruct with empty chunks
	reconstructedReader, err := deduplicator.ReconstructFile(context.Background(), []*Chunk{})
	require.NoError(t, err)

	// Read the reconstructed data
	reconstructedData, err := io.ReadAll(reconstructedReader)
	require.NoError(t, err)

	// Should be empty
	assert.Empty(t, reconstructedData)
}

func TestDeduplicator_DeleteFile(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, chunkStore)

	// Process a file
	data := []byte("hello world")
	reader := bytes.NewReader(data)
	chunks, err := deduplicator.ProcessFile(context.Background(), reader)
	require.NoError(t, err)

	// Verify chunks exist
	for _, chunk := range chunks {
		_, err := chunkStore.Get(context.Background(), chunk.ID)
		assert.NoError(t, err)
	}

	// Delete the file
	err = deduplicator.DeleteFile(context.Background(), chunks)
	assert.NoError(t, err)

	// Verify chunks are deleted (ref count should be 0)
	for _, chunk := range chunks {
		_, err := chunkStore.Get(context.Background(), chunk.ID)
		assert.Error(t, err) // Should not exist anymore
	}
}

func TestDeduplicator_DeleteFile_MultipleReferences(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, chunkStore)

	// Process the same data twice (creates multiple references)
	data := []byte("hello world")
	reader1 := bytes.NewReader(data)
	reader2 := bytes.NewReader(data)

	chunks1, err1 := deduplicator.ProcessFile(context.Background(), reader1)
	chunks2, err2 := deduplicator.ProcessFile(context.Background(), reader2)
	require.NoError(t, err1)
	require.NoError(t, err2)

	// Delete first file
	err := deduplicator.DeleteFile(context.Background(), chunks1)
	assert.NoError(t, err)

	// Chunks should still exist (ref count should be 1)
	for _, chunk := range chunks1 {
		_, err := chunkStore.Get(context.Background(), chunk.ID)
		assert.NoError(t, err)
	}

	// Delete second file
	err = deduplicator.DeleteFile(context.Background(), chunks2)
	assert.NoError(t, err)

	// Now chunks should be deleted
	for _, chunk := range chunks2 {
		_, err := chunkStore.Get(context.Background(), chunk.ID)
		assert.Error(t, err)
	}
}

func TestDeduplicator_GetStats(t *testing.T) {
	chunkStore := NewMemoryChunkStore()
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, chunkStore)

	// Initially no chunks
	stats := deduplicator.GetStats()
	assert.Equal(t, 0, stats.TotalChunks)
	assert.Equal(t, int64(10), stats.ChunkSize)
	assert.Equal(t, "sha256", stats.Algorithm)

	// Process a file
	data := []byte("hello world")
	reader := bytes.NewReader(data)
	_, err := deduplicator.ProcessFile(context.Background(), reader)
	require.NoError(t, err)

	// Should have chunks now
	stats = deduplicator.GetStats()
	assert.Greater(t, stats.TotalChunks, 0)
}

func TestMemoryChunkStore_Store(t *testing.T) {
	store := NewMemoryChunkStore()

	chunk := &Chunk{
		ID:       "test-chunk",
		Hash:     "test-hash",
		Size:     100,
		RefCount: 1,
		Data:     []byte("test data"),
	}

	err := store.Store(context.Background(), chunk)
	assert.NoError(t, err)

	// Verify chunk was stored
	retrieved, err := store.Get(context.Background(), "test-chunk")
	assert.NoError(t, err)
	assert.Equal(t, chunk, retrieved)
}

func TestMemoryChunkStore_Get(t *testing.T) {
	store := NewMemoryChunkStore()

	chunk := &Chunk{
		ID:       "test-chunk",
		Hash:     "test-hash",
		Size:     100,
		RefCount: 1,
		Data:     []byte("test data"),
	}

	// Store chunk
	err := store.Store(context.Background(), chunk)
	require.NoError(t, err)

	// Get existing chunk
	retrieved, err := store.Get(context.Background(), "test-chunk")
	assert.NoError(t, err)
	assert.Equal(t, chunk, retrieved)

	// Get non-existing chunk
	_, err = store.Get(context.Background(), "non-existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryChunkStore_Delete(t *testing.T) {
	store := NewMemoryChunkStore()

	chunk := &Chunk{
		ID:       "test-chunk",
		Hash:     "test-hash",
		Size:     100,
		RefCount: 1,
		Data:     []byte("test data"),
	}

	// Store chunk
	err := store.Store(context.Background(), chunk)
	require.NoError(t, err)

	// Verify chunk exists
	_, err = store.Get(context.Background(), "test-chunk")
	assert.NoError(t, err)

	// Delete chunk
	err = store.Delete(context.Background(), "test-chunk")
	assert.NoError(t, err)

	// Verify chunk is deleted
	_, err = store.Get(context.Background(), "test-chunk")
	assert.Error(t, err)
}

func TestMemoryChunkStore_IncrementRef(t *testing.T) {
	store := NewMemoryChunkStore()

	chunk := &Chunk{
		ID:       "test-chunk",
		Hash:     "test-hash",
		Size:     100,
		RefCount: 1,
		Data:     []byte("test data"),
	}

	// Store chunk
	err := store.Store(context.Background(), chunk)
	require.NoError(t, err)

	// Increment reference count
	err = store.IncrementRef(context.Background(), "test-chunk")
	assert.NoError(t, err)

	// Verify reference count increased
	retrieved, err := store.Get(context.Background(), "test-chunk")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), retrieved.RefCount)

	// Try to increment non-existing chunk
	err = store.IncrementRef(context.Background(), "non-existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryChunkStore_DecrementRef(t *testing.T) {
	store := NewMemoryChunkStore()

	chunk := &Chunk{
		ID:       "test-chunk",
		Hash:     "test-hash",
		Size:     100,
		RefCount: 2,
		Data:     []byte("test data"),
	}

	// Store chunk
	err := store.Store(context.Background(), chunk)
	require.NoError(t, err)

	// Decrement reference count
	err = store.DecrementRef(context.Background(), "test-chunk")
	assert.NoError(t, err)

	// Verify reference count decreased
	retrieved, err := store.Get(context.Background(), "test-chunk")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), retrieved.RefCount)

	// Try to decrement non-existing chunk
	err = store.DecrementRef(context.Background(), "non-existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryChunkStore_Concurrency(t *testing.T) {
	store := NewMemoryChunkStore()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			chunk := &Chunk{
				ID:       fmt.Sprintf("chunk-%d", id),
				Hash:     fmt.Sprintf("hash-%d", id),
				Size:     100,
				RefCount: 1,
				Data:     []byte(fmt.Sprintf("data-%d", id)),
			}

			// Store
			err := store.Store(context.Background(), chunk)
			assert.NoError(t, err)

			// Get
			retrieved, err := store.Get(context.Background(), chunk.ID)
			assert.NoError(t, err)
			assert.Equal(t, chunk, retrieved)

			// Increment ref
			err = store.IncrementRef(context.Background(), chunk.ID)
			assert.NoError(t, err)

			// Decrement ref
			err = store.DecrementRef(context.Background(), chunk.ID)
			assert.NoError(t, err)

			// Delete
			err = store.Delete(context.Background(), chunk.ID)
			assert.NoError(t, err)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for goroutines")
		}
	}
}

func TestDeduplicator_ProcessChunk_ErrorHandling(t *testing.T) {
	// Create a mock chunk store that returns errors
	mockStore := &mockChunkStore{shouldError: true}
	config := &DeduplicationConfig{
		ChunkSize: 10,
		Algorithm: "sha256",
	}
	deduplicator := NewDeduplicator(config, mockStore)

	// Process a file - should handle errors gracefully
	data := []byte("hello world")
	reader := bytes.NewReader(data)
	chunks, err := deduplicator.ProcessFile(context.Background(), reader)

	assert.Error(t, err)
	assert.Nil(t, chunks)
}

// mockChunkStore is a mock implementation for testing error conditions
type mockChunkStore struct {
	shouldError bool
}

func (m *mockChunkStore) Store(ctx context.Context, chunk *Chunk) error {
	if m.shouldError {
		return fmt.Errorf("mock store error")
	}
	return nil
}

func (m *mockChunkStore) Get(ctx context.Context, id string) (*Chunk, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock store error")
	}
	return &Chunk{ID: id}, nil
}

func (m *mockChunkStore) Delete(ctx context.Context, id string) error {
	if m.shouldError {
		return fmt.Errorf("mock store error")
	}
	return nil
}

func (m *mockChunkStore) IncrementRef(ctx context.Context, id string) error {
	if m.shouldError {
		return fmt.Errorf("mock store error")
	}
	return nil
}

func (m *mockChunkStore) DecrementRef(ctx context.Context, id string) error {
	if m.shouldError {
		return fmt.Errorf("mock store error")
	}
	return nil
}
