package fileserver

import (
	"sync"
	"time"
)

// FileInfo is lightweight metadata tracked for API surfaces (REST/GraphQL/gRPC).
type FileInfo struct {
	Key         string
	Name        string
	Size        int64
	ContentType string
	Hash        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
}

// PeerInfo describes a connected peer for API listing.
type PeerInfo struct {
	ID       string
	Address  string
	Status   string
	LastSeen time.Time
}

type fileRegistry struct {
	mu    sync.RWMutex
	files map[string]*FileInfo
}

func newFileRegistry() *fileRegistry {
	return &fileRegistry{files: make(map[string]*FileInfo)}
}

func (r *fileRegistry) put(info *FileInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *info
	if info.Metadata != nil {
		cp.Metadata = copyStringMap(info.Metadata)
	}
	r.files[info.Key] = &cp
}

func (r *fileRegistry) get(key string) (*FileInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.files[key]
	if !ok {
		return nil, false
	}
	cp := *info
	if info.Metadata != nil {
		cp.Metadata = copyStringMap(info.Metadata)
	}
	return &cp, true
}

func (r *fileRegistry) delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.files, key)
}

func (r *fileRegistry) list() []FileInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]FileInfo, 0, len(r.files))
	for _, info := range r.files {
		cp := *info
		if info.Metadata != nil {
			cp.Metadata = copyStringMap(info.Metadata)
		}
		out = append(out, cp)
	}
	return out
}

func (r *fileRegistry) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.files)
}

func (r *fileRegistry) updateMetadata(key string, metadata map[string]string) (*FileInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, ok := r.files[key]
	if !ok {
		return nil, false
	}
	info.Metadata = copyStringMap(metadata)
	info.UpdatedAt = time.Now().UTC()
	cp := *info
	if info.Metadata != nil {
		cp.Metadata = copyStringMap(info.Metadata)
	}
	return &cp, true
}

func copyStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
