package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/internal/app/fileserver"
	"github.com/Skpow1234/Peervault/proto/peervault"
)

// FileService provides file-related operations backed by fileserver when set.
type FileService struct {
	server *fileserver.Server
	// fallback in-memory store when server is nil (tests / degraded mode)
	files map[string][]byte
}

// NewFileService creates a file service. Pass a fileserver for real storage.
func NewFileService(server *fileserver.Server) *FileService {
	return &FileService{
		server: server,
		files:  make(map[string][]byte),
	}
}

// UploadFile uploads a file and returns file metadata
func (s *FileService) UploadFile(fileKey string, data []byte) (*peervault.FileResponse, error) {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.server.StoreWithMeta(ctx, fileKey, bytes.NewReader(data), "application/octet-stream", nil); err != nil {
			return nil, err
		}
		info, ok := s.server.GetFileInfo(fileKey)
		if !ok {
			return nil, fmt.Errorf("upload missing metadata for %s", fileKey)
		}
		return fileInfoToProto(info, s.server.NodeID()), nil
	}

	s.files[fileKey] = data
	return &peervault.FileResponse{
		Key:         fileKey,
		Name:        fileKey,
		Size:        int64(len(data)),
		ContentType: "application/octet-stream",
		Hash:        "",
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		Metadata:    map[string]string{},
		Replicas:    []*peervault.FileReplica{},
	}, nil
}

// DownloadFile downloads a file by key
func (s *FileService) DownloadFile(key string) ([]byte, error) {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		r, err := s.server.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if closer, ok := r.(io.Closer); ok {
			defer func() { _ = closer.Close() }()
		}
		return io.ReadAll(r)
	}
	data, exists := s.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	return data, nil
}

// ListFiles lists files with pagination and filtering
func (s *FileService) ListFiles(page, pageSize int, filter string) (*peervault.ListFilesResponse, error) {
	_ = filter
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	var files []*peervault.FileResponse
	if s.server != nil {
		infos := s.server.ListFiles()
		files = make([]*peervault.FileResponse, 0, len(infos))
		for i := range infos {
			files = append(files, fileInfoToProto(&infos[i], s.server.NodeID()))
		}
	} else {
		files = make([]*peervault.FileResponse, 0, len(s.files))
		for k, data := range s.files {
			files = append(files, &peervault.FileResponse{
				Key:  k,
				Name: k,
				Size: int64(len(data)),
			})
		}
	}

	start := (page - 1) * pageSize
	if start > len(files) {
		start = len(files)
	}
	end := start + pageSize
	if end > len(files) {
		end = len(files)
	}

	return &peervault.ListFilesResponse{
		Files:    files[start:end],
		Total:    int32(len(files)),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// GetFile retrieves file metadata by key
func (s *FileService) GetFile(key string) (*peervault.FileResponse, error) {
	if s.server != nil {
		info, ok := s.server.GetFileInfo(key)
		if !ok || !s.server.Has(key) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return fileInfoToProto(info, s.server.NodeID()), nil
	}
	data, exists := s.files[key]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	return &peervault.FileResponse{
		Key:  key,
		Name: key,
		Size: int64(len(data)),
	}, nil
}

// DeleteFile deletes a file by key
func (s *FileService) DeleteFile(key string) (bool, error) {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.server.Delete(ctx, key); err != nil {
			return false, err
		}
		return true, nil
	}
	if _, exists := s.files[key]; !exists {
		return false, fmt.Errorf("file not found: %s", key)
	}
	delete(s.files, key)
	return true, nil
}

// UpdateFileMetadata updates file metadata
func (s *FileService) UpdateFileMetadata(key string, metadata map[string]string) (*peervault.FileResponse, error) {
	if s.server != nil {
		info, err := s.server.UpdateFileMetadata(key, metadata)
		if err != nil {
			return nil, err
		}
		return fileInfoToProto(info, s.server.NodeID()), nil
	}
	return s.GetFile(key)
}

// StreamFileOperations streams file operation events
func (s *FileService) StreamFileOperations() (<-chan *peervault.FileOperationEvent, error) {
	ch := make(chan *peervault.FileOperationEvent)
	close(ch)
	return ch, nil
}

func fileInfoToProto(info *fileserver.FileInfo, nodeID string) *peervault.FileResponse {
	return &peervault.FileResponse{
		Key:         info.Key,
		Name:        info.Name,
		Size:        info.Size,
		ContentType: info.ContentType,
		Hash:        info.Hash,
		CreatedAt:   timestamppb.New(info.CreatedAt),
		UpdatedAt:   timestamppb.New(info.UpdatedAt),
		Metadata:    info.Metadata,
		Replicas: []*peervault.FileReplica{
			{PeerId: nodeID, Status: "active", CreatedAt: timestamppb.New(info.CreatedAt)},
		},
	}
}
