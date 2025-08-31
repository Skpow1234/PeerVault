package services

import (
	"crypto/sha256"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/proto/peervault"
)

// FileService provides file-related operations
type FileService struct {
	// Mock storage for demonstration
	files map[string][]byte
}

// NewFileService creates a new file service instance
func NewFileService() *FileService {
	return &FileService{
		files: make(map[string][]byte),
	}
}

// UploadFile uploads a file and returns file metadata
func (s *FileService) UploadFile(fileKey string, data []byte) (*peervault.FileResponse, error) {
	// Store the file data
	s.files[fileKey] = data

	// Calculate hash
	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	// Create file response
	response := &peervault.FileResponse{
		Key:         fileKey,
		Name:        fileKey,
		Size:        int64(len(data)),
		ContentType: "application/octet-stream",
		Hash:        hash,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		Metadata: map[string]string{
			"uploaded_at": time.Now().Format(time.RFC3339),
		},
		Replicas: []*peervault.FileReplica{
			{
				PeerId:    "local",
				Status:    "active",
				CreatedAt: timestamppb.Now(),
			},
		},
	}

	return response, nil
}

// DownloadFile downloads a file by key
func (s *FileService) DownloadFile(key string) ([]byte, error) {
	data, exists := s.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	return data, nil
}

// ListFiles lists files with pagination and filtering
func (s *FileService) ListFiles(page, pageSize int, filter string) (*peervault.ListFilesResponse, error) {
	// Mock implementation
	files := []*peervault.FileResponse{
		{
			Key:         "file1.txt",
			Name:        "file1.txt",
			Size:        1024,
			ContentType: "text/plain",
			Hash:        "abc123",
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Metadata:    map[string]string{},
			Replicas:    []*peervault.FileReplica{},
		},
		{
			Key:         "file2.jpg",
			Name:        "file2.jpg",
			Size:        2048,
			ContentType: "image/jpeg",
			Hash:        "def456",
			CreatedAt:   timestamppb.Now(),
			UpdatedAt:   timestamppb.Now(),
			Metadata:    map[string]string{},
			Replicas:    []*peervault.FileReplica{},
		},
	}

	return &peervault.ListFilesResponse{
		Files:    files,
		Total:    int32(len(files)),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// GetFile retrieves file metadata by key
func (s *FileService) GetFile(key string) (*peervault.FileResponse, error) {
	data, exists := s.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	return &peervault.FileResponse{
		Key:         key,
		Name:        key,
		Size:        int64(len(data)),
		ContentType: "application/octet-stream",
		Hash:        hash,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		Metadata:    map[string]string{},
		Replicas:    []*peervault.FileReplica{},
	}, nil
}

// DeleteFile deletes a file by key
func (s *FileService) DeleteFile(key string) (bool, error) {
	if _, exists := s.files[key]; !exists {
		return false, fmt.Errorf("file not found: %s", key)
	}

	delete(s.files, key)
	return true, nil
}

// UpdateFileMetadata updates file metadata
func (s *FileService) UpdateFileMetadata(key string, metadata map[string]string) (*peervault.FileResponse, error) {
	data, exists := s.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	return &peervault.FileResponse{
		Key:         key,
		Name:        key,
		Size:        int64(len(data)),
		ContentType: "application/octet-stream",
		Hash:        hash,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		Metadata:    metadata,
		Replicas:    []*peervault.FileReplica{},
	}, nil
}

// StreamFileOperations streams file operation events
func (s *FileService) StreamFileOperations() (<-chan *peervault.FileOperationEvent, error) {
	// Mock implementation - return a channel that will be closed
	ch := make(chan *peervault.FileOperationEvent)
	close(ch)
	return ch, nil
}
