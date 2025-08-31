package services

import (
	"context"
	"github.com/Skpow1234/Peervault/internal/api/rest/types"
)

// FileService defines the interface for file operations
type FileService interface {
	// ListFiles retrieves all files
	ListFiles(ctx context.Context) ([]types.File, error)
	
	// GetFile retrieves a file by key
	GetFile(ctx context.Context, key string) (*types.File, error)
	
	// UploadFile uploads a new file
	UploadFile(ctx context.Context, name string, data []byte, contentType string, metadata map[string]string) (*types.File, error)
	
	// DeleteFile deletes a file by key
	DeleteFile(ctx context.Context, key string) error
	
	// UpdateFileMetadata updates file metadata
	UpdateFileMetadata(ctx context.Context, key string, metadata map[string]string) (*types.File, error)
}
