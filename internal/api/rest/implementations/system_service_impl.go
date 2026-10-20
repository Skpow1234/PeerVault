package implementations

import (
	"context"
	"runtime"
	"time"

	"github.com/Skpow1234/Shugo/internal/api/rest/services"
	"github.com/Skpow1234/Shugo/internal/api/rest/types"
	"github.com/Skpow1234/Shugo/internal/app/fileserver"
)

type SystemServiceImpl struct {
	server    *fileserver.Server
	startTime time.Time
}

func NewSystemService(server *fileserver.Server) services.SystemService {
	start := time.Now().UTC()
	if server != nil {
		start = server.StartTime()
	}
	return &SystemServiceImpl{
		server:    server,
		startTime: start,
	}
}

func (s *SystemServiceImpl) GetSystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	_ = ctx
	peerCount := 0
	fileCount := 0
	var storageUsed int64
	if s.server != nil {
		peerCount = s.server.PeerCount()
		files := s.server.ListFiles()
		fileCount = len(files)
		for _, f := range files {
			storageUsed += f.Size
		}
	}
	return &types.SystemInfo{
		Version:      "0.1.0-mvp",
		Uptime:       time.Since(s.startTime),
		StartTime:    s.startTime,
		StorageUsed:  storageUsed,
		StorageTotal: storageUsed,
		PeerCount:    peerCount,
		FileCount:    fileCount,
	}, nil
}

func (s *SystemServiceImpl) GetMetrics(ctx context.Context) (*types.Metrics, error) {
	_ = ctx
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	usage := float64(mem.Alloc) / float64(mem.Sys) * 100
	if mem.Sys == 0 {
		usage = 0
	}
	active := 0
	if s.server != nil {
		active = s.server.PeerCount()
	}
	return &types.Metrics{
		RequestsTotal:     0,
		RequestsPerMinute: 0,
		ActiveConnections: active,
		StorageUsage:      usage,
		LastUpdated:       time.Now().UTC(),
	}, nil
}

func (s *SystemServiceImpl) HealthCheck(ctx context.Context) (bool, error) {
	_ = ctx
	if s.server == nil {
		return true, nil // API-only mode
	}
	return true, nil
}
