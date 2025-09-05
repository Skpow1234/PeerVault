package deduplication

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"sync"
)

// Chunk represents a deduplicated chunk of data
type Chunk struct {
	ID       string `json:"id"`
	Hash     string `json:"hash"`
	Size     int64  `json:"size"`
	RefCount int64  `json:"ref_count"`
	Data     []byte `json:"data,omitempty"`
}

// ChunkStore interface for storing and retrieving chunks
type ChunkStore interface {
	Store(ctx context.Context, chunk *Chunk) error
	Get(ctx context.Context, id string) (*Chunk, error)
	Delete(ctx context.Context, id string) error
	IncrementRef(ctx context.Context, id string) error
	DecrementRef(ctx context.Context, id string) error
}

// Deduplicator implements content-based deduplication
type Deduplicator struct {
	chunkSize  int64
	hasher     hash.Hash
	chunkStore ChunkStore
	chunkIndex map[string]string // hash -> chunk ID mapping
	mu         sync.RWMutex
}

// DeduplicationConfig holds configuration for deduplication
type DeduplicationConfig struct {
	ChunkSize int64  `yaml:"chunk_size"`
	Algorithm string `yaml:"algorithm"`
}

// DefaultDeduplicationConfig returns default deduplication configuration
func DefaultDeduplicationConfig() *DeduplicationConfig {
	return &DeduplicationConfig{
		ChunkSize: 64 * 1024, // 64KB chunks
		Algorithm: "sha256",
	}
}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator(config *DeduplicationConfig, chunkStore ChunkStore) *Deduplicator {
	if config == nil {
		config = DefaultDeduplicationConfig()
	}

	var hasher hash.Hash
	switch config.Algorithm {
	case "sha256":
		hasher = sha256.New()
	default:
		hasher = sha256.New()
	}

	return &Deduplicator{
		chunkSize:  config.ChunkSize,
		hasher:     hasher,
		chunkStore: chunkStore,
		chunkIndex: make(map[string]string),
	}
}

// ProcessFile processes a file and returns deduplicated chunks
func (d *Deduplicator) ProcessFile(ctx context.Context, reader io.Reader) ([]*Chunk, error) {
	var chunks []*Chunk
	buffer := make([]byte, d.chunkSize)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		chunkData := buffer[:n]
		chunk, err := d.processChunk(ctx, chunkData)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// processChunk processes a single chunk
func (d *Deduplicator) processChunk(ctx context.Context, data []byte) (*Chunk, error) {
	// Calculate hash
	d.hasher.Reset()
	d.hasher.Write(data)
	hash := fmt.Sprintf("%x", d.hasher.Sum(nil))

	d.mu.RLock()
	chunkID, exists := d.chunkIndex[hash]
	d.mu.RUnlock()

	if exists {
		// Chunk already exists, increment reference count
		if err := d.chunkStore.IncrementRef(ctx, chunkID); err != nil {
			return nil, err
		}

		chunk, err := d.chunkStore.Get(ctx, chunkID)
		if err != nil {
			return nil, err
		}

		return chunk, nil
	}

	// New chunk, create and store it
	chunkID = fmt.Sprintf("chunk_%s", hash)
	chunk := &Chunk{
		ID:       chunkID,
		Hash:     hash,
		Size:     int64(len(data)),
		RefCount: 1,
		Data:     make([]byte, len(data)),
	}
	copy(chunk.Data, data)

	if err := d.chunkStore.Store(ctx, chunk); err != nil {
		return nil, err
	}

	// Update index
	d.mu.Lock()
	d.chunkIndex[hash] = chunkID
	d.mu.Unlock()

	return chunk, nil
}

// ReconstructFile reconstructs a file from chunks
func (d *Deduplicator) ReconstructFile(ctx context.Context, chunks []*Chunk) (io.Reader, error) {
	readers := make([]io.Reader, len(chunks))

	for i, chunk := range chunks {
		// Get the actual chunk data
		fullChunk, err := d.chunkStore.Get(ctx, chunk.ID)
		if err != nil {
			return nil, err
		}

		readers[i] = io.NewSectionReader(bytes.NewReader(fullChunk.Data), 0, chunk.Size)
	}

	return io.MultiReader(readers...), nil
}

// DeleteFile removes references to chunks for a file
func (d *Deduplicator) DeleteFile(ctx context.Context, chunks []*Chunk) error {
	for _, chunk := range chunks {
		if err := d.chunkStore.DecrementRef(ctx, chunk.ID); err != nil {
			return err
		}

		// Check if chunk should be deleted
		fullChunk, err := d.chunkStore.Get(ctx, chunk.ID)
		if err != nil {
			continue
		}

		if fullChunk.RefCount <= 0 {
			if err := d.chunkStore.Delete(ctx, chunk.ID); err != nil {
				return err
			}

			// Remove from index
			d.mu.Lock()
			delete(d.chunkIndex, chunk.Hash)
			d.mu.Unlock()
		}
	}

	return nil
}

// GetStats returns deduplication statistics
func (d *Deduplicator) GetStats() DeduplicationStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return DeduplicationStats{
		TotalChunks: len(d.chunkIndex),
		ChunkSize:   d.chunkSize,
		Algorithm:   "sha256",
	}
}

// DeduplicationStats holds deduplication statistics
type DeduplicationStats struct {
	TotalChunks int    `json:"total_chunks"`
	ChunkSize   int64  `json:"chunk_size"`
	Algorithm   string `json:"algorithm"`
}

// MemoryChunkStore implements an in-memory chunk store
type MemoryChunkStore struct {
	chunks map[string]*Chunk
	mu     sync.RWMutex
}

// NewMemoryChunkStore creates a new in-memory chunk store
func NewMemoryChunkStore() *MemoryChunkStore {
	return &MemoryChunkStore{
		chunks: make(map[string]*Chunk),
	}
}

// Store stores a chunk
func (mcs *MemoryChunkStore) Store(ctx context.Context, chunk *Chunk) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	mcs.chunks[chunk.ID] = chunk
	return nil
}

// Get retrieves a chunk
func (mcs *MemoryChunkStore) Get(ctx context.Context, id string) (*Chunk, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	chunk, exists := mcs.chunks[id]
	if !exists {
		return nil, fmt.Errorf("chunk %s not found", id)
	}

	return chunk, nil
}

// Delete deletes a chunk
func (mcs *MemoryChunkStore) Delete(ctx context.Context, id string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	delete(mcs.chunks, id)
	return nil
}

// IncrementRef increments the reference count for a chunk
func (mcs *MemoryChunkStore) IncrementRef(ctx context.Context, id string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	chunk, exists := mcs.chunks[id]
	if !exists {
		return fmt.Errorf("chunk %s not found", id)
	}

	chunk.RefCount++
	return nil
}

// DecrementRef decrements the reference count for a chunk
func (mcs *MemoryChunkStore) DecrementRef(ctx context.Context, id string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	chunk, exists := mcs.chunks[id]
	if !exists {
		return fmt.Errorf("chunk %s not found", id)
	}

	chunk.RefCount--
	return nil
}
