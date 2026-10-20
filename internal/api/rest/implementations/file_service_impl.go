package implementations

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Skpow1234/Shugo/internal/api/rest/services"
	"github.com/Skpow1234/Shugo/internal/api/rest/types"
	"github.com/Skpow1234/Shugo/internal/app/fileserver"
)

// FileServiceImpl serves REST file operations backed by a Shugo fileserver.
type FileServiceImpl struct {
	server *fileserver.Server
}

// NewFileService creates a FileService. If server is nil, mutating ops fail closed.
func NewFileService(server *fileserver.Server) services.FileService {
	return &FileServiceImpl{server: server}
}

func (s *FileServiceImpl) ListFiles(ctx context.Context) ([]types.File, error) {
	_ = ctx
	if s.server == nil {
		return []types.File{}, nil
	}
	infos := s.server.ListFiles()
	out := make([]types.File, 0, len(infos))
	for _, info := range infos {
		out = append(out, fileInfoToREST(info, s.server.NodeID()))
	}
	return out, nil
}

func (s *FileServiceImpl) GetFile(ctx context.Context, key string) (*types.File, error) {
	_ = ctx
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	info, ok := s.server.GetFileInfo(key)
	if !ok || !s.server.Has(key) {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	f := fileInfoToREST(*info, s.server.NodeID())
	return &f, nil
}

func (s *FileServiceImpl) UploadFile(ctx context.Context, name string, data []byte, contentType string, metadata map[string]string) (*types.File, error) {
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	if name == "" {
		name = fmt.Sprintf("file_%d", time.Now().UnixNano())
	}
	key := name
	if metadata == nil {
		metadata = map[string]string{}
	}
	metadata["name"] = name

	if err := s.server.StoreWithMeta(ctx, key, bytes.NewReader(data), contentType, metadata); err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}
	info, ok := s.server.GetFileInfo(key)
	if !ok {
		return nil, fmt.Errorf("store succeeded but metadata missing for %s", key)
	}
	f := fileInfoToREST(*info, s.server.NodeID())
	return &f, nil
}

func (s *FileServiceImpl) DeleteFile(ctx context.Context, key string) error {
	if s.server == nil {
		return fmt.Errorf("fileserver not configured")
	}
	return s.server.Delete(ctx, key)
}

func (s *FileServiceImpl) UpdateFileMetadata(ctx context.Context, key string, metadata map[string]string) (*types.File, error) {
	_ = ctx
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	info, err := s.server.UpdateFileMetadata(key, metadata)
	if err != nil {
		return nil, err
	}
	f := fileInfoToREST(*info, s.server.NodeID())
	return &f, nil
}

// ReadFileContent returns decrypted file bytes for a logical key.
func (s *FileServiceImpl) ReadFileContent(ctx context.Context, key string) ([]byte, error) {
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	r, err := s.server.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if closer, ok := r.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}
	return io.ReadAll(r)
}

func fileInfoToREST(info fileserver.FileInfo, nodeID string) types.File {
	return types.File{
		Key:         info.Key,
		Name:        info.Name,
		Size:        info.Size,
		ContentType: info.ContentType,
		Hash:        info.Hash,
		CreatedAt:   info.CreatedAt,
		UpdatedAt:   info.UpdatedAt,
		Metadata:    info.Metadata,
		Replicas: []types.FileReplica{
			{PeerID: nodeID, Status: "active", CreatedAt: info.CreatedAt},
		},
	}
}
