package services

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/proto/peervault"
)

// SystemService provides system-related operations
type SystemService struct {
	startTime time.Time
}

// NewSystemService creates a new system service instance
func NewSystemService() *SystemService {
	return &SystemService{
		startTime: time.Now(),
	}
}

// GetSystemInfo retrieves system information
func (s *SystemService) GetSystemInfo() (*peervault.SystemInfoResponse, error) {
	return &peervault.SystemInfoResponse{
		Version:        "1.0.0",
		UptimeSeconds:  int64(time.Since(s.startTime).Seconds()),
		StartTime:      timestamppb.New(s.startTime),
		StorageUsed:    1024 * 1024 * 100, // 100MB
		StorageTotal:   1024 * 1024 * 1024, // 1GB
		PeerCount:      3,
		FileCount:      15,
	}, nil
}

// GetMetrics retrieves system metrics
func (s *SystemService) GetMetrics() (*peervault.MetricsResponse, error) {
	return &peervault.MetricsResponse{
		RequestsTotal:       1250,
		RequestsPerMinute:   12.5,
		ActiveConnections:   8,
		StorageUsagePercent: 9.8,
		LastUpdated:         timestamppb.Now(),
	}, nil
}

// HealthCheck performs a health check
func (s *SystemService) HealthCheck() (*peervault.HealthResponse, error) {
	return &peervault.HealthResponse{
		Status:    "healthy",
		Timestamp: timestamppb.Now(),
		Version:   "1.0.0",
	}, nil
}

// StreamSystemEvents streams system events
func (s *SystemService) StreamSystemEvents() (<-chan *peervault.SystemEvent, error) {
	// Mock implementation - return a channel that will be closed
	ch := make(chan *peervault.SystemEvent)
	close(ch)
	return ch, nil
}
