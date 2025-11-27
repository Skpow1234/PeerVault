package analytics

import (
	"log/slog"
	"time"
	
	"github.com/google/uuid"
)

// Service represents the analytics service
type Service struct {
	storage Storage
	logger  *slog.Logger
	config  *Config
}

// Config represents analytics service configuration
type Config struct {
	Enabled           bool
	MaxEntries        int
	RetentionDays     int
	CleanupInterval   time.Duration
	EnableUserTracking bool
}

// DefaultConfig returns default analytics configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:           true,
		MaxEntries:        100000,
		RetentionDays:     30,
		CleanupInterval:   24 * time.Hour,
		EnableUserTracking: true,
	}
}

// NewService creates a new analytics service
func NewService(storage Storage, logger *slog.Logger, config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	
	service := &Service{
		storage: storage,
		logger:  logger,
		config:  config,
	}
	
	// Start cleanup routine
	go service.cleanupRoutine()
	
	return service
}

// RecordAPICall records an API call
func (s *Service) RecordAPICall(call *APICall) error {
	if !s.config.Enabled {
		return nil
	}
	
	// Generate ID if not present
	if call.ID == "" {
		call.ID = uuid.New().String()
	}
	
	// Set timestamp if not present
	if call.Timestamp.IsZero() {
		call.Timestamp = time.Now()
	}
	
	// Normalize endpoint (remove query parameters)
	if call.Endpoint == "" {
		call.Endpoint = call.Path
	}
	
	err := s.storage.RecordAPICall(call)
	if err != nil {
		s.logger.Error("Failed to record API call", "error", err)
		return err
	}
	
	return nil
}

// GetUsageMetrics retrieves usage metrics for a time period
func (s *Service) GetUsageMetrics(startTime, endTime time.Time) (*UsageMetrics, error) {
	return s.storage.GetUsageMetrics(startTime, endTime)
}

// GetEndpointStats retrieves statistics for a specific endpoint
func (s *Service) GetEndpointStats(endpoint string, startTime, endTime time.Time) (*EndpointStats, error) {
	return s.storage.GetEndpointStats(endpoint, startTime, endTime)
}

// GetUserBehavior retrieves behavior metrics for a specific user
func (s *Service) GetUserBehavior(userID string, startTime, endTime time.Time) (*UserBehaviorMetrics, error) {
	if !s.config.EnableUserTracking {
		return nil, nil
	}
	return s.storage.GetUserBehavior(userID, startTime, endTime)
}

// GetUsageTrends retrieves usage trends over time
func (s *Service) GetUsageTrends(startTime, endTime time.Time, interval string) ([]*UsageTrend, error) {
	return s.storage.GetUsageTrends(startTime, endTime, interval)
}

// GetPopularityMetrics retrieves popularity metrics
func (s *Service) GetPopularityMetrics(startTime, endTime time.Time) (*PopularityMetrics, error) {
	return s.storage.GetPopularityMetrics(startTime, endTime)
}

// GetAnalyticsSummary retrieves a comprehensive analytics summary
func (s *Service) GetAnalyticsSummary() (*AnalyticsSummary, error) {
	return s.storage.GetAnalyticsSummary()
}

// QueryAPICalls retrieves API calls based on query parameters
func (s *Service) QueryAPICalls(query *AnalyticsQuery) ([]*APICall, error) {
	return s.storage.GetAPICalls(query)
}

// cleanupRoutine periodically removes old analytics data
func (s *Service) cleanupRoutine() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		cutoffTime := time.Now().AddDate(0, 0, -s.config.RetentionDays)
		err := s.storage.CleanupOldData(cutoffTime)
		if err != nil {
			s.logger.Error("Failed to cleanup old analytics data", "error", err)
		} else {
			s.logger.Info("Cleaned up old analytics data", "cutoff_time", cutoffTime)
		}
	}
}

// GetTimeWindow returns a time window for common periods
func GetTimeWindow(period string) (time.Time, time.Time) {
	now := time.Now()
	
	switch period {
	case "hour":
		return now.Add(-time.Hour), now
	case "day", "24h":
		return now.Add(-24 * time.Hour), now
	case "week", "7d":
		return now.Add(-7 * 24 * time.Hour), now
	case "month", "30d":
		return now.Add(-30 * 24 * time.Hour), now
	case "year":
		return now.AddDate(-1, 0, 0), now
	default:
		return now.Add(-24 * time.Hour), now
	}
}

// CalculatePercentileLatency calculates percentile latency from API calls
func CalculatePercentileLatency(calls []*APICall, percentile float64) time.Duration {
	if len(calls) == 0 {
		return 0
	}
	
	// Create a slice of durations
	durations := make([]time.Duration, len(calls))
	for i, call := range calls {
		durations[i] = call.Duration
	}
	
	// Simple bubble sort for durations
	for i := 0; i < len(durations); i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[j] < durations[i] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
	
	// Calculate percentile index
	index := int(float64(len(durations)) * percentile / 100.0)
	if index >= len(durations) {
		index = len(durations) - 1
	}
	
	return durations[index]
}

// GetErrorBreakdown returns a breakdown of errors by status code
func GetErrorBreakdown(calls []*APICall) map[int]int64 {
	breakdown := make(map[int]int64)
	
	for _, call := range calls {
		if call.StatusCode >= 400 {
			breakdown[call.StatusCode]++
		}
	}
	
	return breakdown
}

// GetTopUserAgents returns the most common user agents
func GetTopUserAgents(calls []*APICall, limit int) map[string]int64 {
	userAgents := make(map[string]int64)
	
	for _, call := range calls {
		if call.UserAgent != "" {
			userAgents[call.UserAgent]++
		}
	}
	
	return userAgents
}

