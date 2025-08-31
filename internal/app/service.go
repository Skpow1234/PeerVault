package app

import "github.com/Skpow1234/Peervault/internal/domain"

// FileService defines the high-level application operations.
type FileService interface {
	// Store persists and broadcasts file availability.
	Store(meta domain.FileMetadata, dataReader Readable) error

	// Get retrieves a file, possibly from peers, and returns a reader.
	Get(key string) (Readable, error)
}

// Readable is the minimal interface for readers we work with. Matches io.Reader
// without importing to keep app package lightweight.
type Readable interface {
	Read(p []byte) (n int, err error)
}