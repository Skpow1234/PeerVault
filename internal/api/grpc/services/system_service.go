package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/grpc/types"
)

type SystemService struct {
	logger    *slog.Logger
	startTime time.Time
}

func NewSystemService(logger *slog.Logger) *SystemService {
	return &SystemService{
		logger:    logger,
		startTime: time.Now(),
	}
}

func (s *SystemService) GetSystemInfo(ctx context.Context) (*types.SystemInfoResponse, error) {
	uptime := time.Since(s.startTime)

	return &types.SystemInfoResponse{
		Version:       "1.0.0",
		UptimeSeconds: int64(uptime.Seconds()),
		StartTime:     s.startTime,
		StorageUsed:   1024 * 1024,      // 1MB
		StorageTotal:  10 * 1024 * 1024, // 10MB
		PeerCount:     3,
		FileCount:     5,
	}, nil
}

func (s *SystemService) GetMetrics(ctx context.Context) (*types.MetricsResponse, error) {
	return &types.MetricsResponse{
		RequestsTotal:       100,
		RequestsPerMinute:   10.5,
		ActiveConnections:   5,
		StorageUsagePercent: 10.0,
		LastUpdated:         time.Now(),
	}, nil
}

func (s *SystemService) HealthCheck(ctx context.Context) (*types.HealthResponse, error) {
	return &types.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}, nil
}

func (s *SystemService) StreamSystemEvents(stream interface{}) error {
	// TODO: Implement real-time system event streaming
	s.logger.Info("System event streaming not yet implemented")
	return nil
}
