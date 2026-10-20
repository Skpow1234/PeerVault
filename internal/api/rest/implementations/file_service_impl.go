package implementations

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/rest/services"
	"github.com/Skpow1234/Peervault/internal/api/rest/types"
	"github.com/Skpow1234/Peervault/internal/app/fileserver"
)

// FileServiceImpl serves REST file operations backed by a PeerVault fileserver.
type FileServiceImpl struct {
	server *fileserver.Server

	mu    sync.RWMutex
	index map[string]*types.File
}

// NewFileService creates a FileService. If server is nil, operations that need
// storage return errors (no more silent mock success for uploads).
func NewFileService(server *fileserver.Server) services.FileService {
	return &FileServiceImpl{
		server: server,
		index:  make(map[string]*types.File),
	}
}

func (s *FileServiceImpl) ListFiles(ctx context.Context) ([]types.File, error) {
	_ = ctx
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]types.File, 0, len(s.index))
	for _, f := range s.index {
		cp := *f
		if f.Metadata != nil {
			cp.Metadata = copyMap(f.Metadata)
		}
		out = append(out, cp)
	}
	return out, nil
}

func (s *FileServiceImpl) GetFile(ctx context.Context, key string) (*types.File, error) {
	_ = ctx
	s.mu.RLock()
	rec, ok := s.index[key]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	if s.server != nil && !s.server.Has(key) {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	cp := *rec
	if rec.Metadata != nil {
		cp.Metadata = copyMap(rec.Metadata)
	}
	return &cp, nil
}

func (s *FileServiceImpl) UploadFile(ctx context.Context, name string, data []byte, contentType string, metadata map[string]string) (*types.File, error) {
	if s.server == nil {
		return nil, fmt.Errorf("fileserver not configured")
	}
	if name == "" {
		name = fmt.Sprintf("file_%d", time.Now().UnixNano())
	}
	key := name
	hash := fmt.Sprintf("%x", sha256.Sum256(data))

	if err := s.server.Store(ctx, key, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	now := time.Now().UTC()
	rec := &types.File{
		Key:         key,
		Name:        name,
		Size:        int64(len(data)),
		ContentType: contentType,
		Hash:        hash,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    copyMap(metadata),
		Replicas: []types.FileReplica{
			{PeerID: s.server.NodeID(), Status: "active", CreatedAt: now},
		},
	}

	s.mu.Lock()
	s.index[key] = rec
	s.mu.Unlock()

	cp := *rec
	if rec.Metadata != nil {
		cp.Metadata = copyMap(rec.Metadata)
	}
	return &cp, nil
}

func (s *FileServiceImpl) DeleteFile(ctx context.Context, key string) error {
	if s.server == nil {
		return fmt.Errorf("fileserver not configured")
	}
	if err := s.server.Delete(ctx, key); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.index, key)
	s.mu.Unlock()
	return nil
}

func (s *FileServiceImpl) UpdateFileMetadata(ctx context.Context, key string, metadata map[string]string) (*types.File, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, ok := s.index[key]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", key)
	}
	if s.server != nil && !s.server.Has(key) {
		return nil, fmt.Errorf("file not found: %s", key)
	}

	rec.Metadata = copyMap(metadata)
	rec.UpdatedAt = time.Now().UTC()

	cp := *rec
	if rec.Metadata != nil {
		cp.Metadata = copyMap(rec.Metadata)
	}
	return &cp, nil
}

// ReadFileContent returns decrypted file bytes for a logical key (helper for download endpoints).
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

func copyMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
