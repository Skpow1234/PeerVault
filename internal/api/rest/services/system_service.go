package services

import (
	"context"
	"github.com/Skpow1234/Peervault/internal/api/rest/types"
)

// SystemService defines the interface for system operations
type SystemService interface {
	// GetSystemInfo retrieves system information
	GetSystemInfo(ctx context.Context) (*types.SystemInfo, error)
	
	// GetMetrics retrieves system metrics
	GetMetrics(ctx context.Context) (*types.Metrics, error)
	
	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) (bool, error)
}
