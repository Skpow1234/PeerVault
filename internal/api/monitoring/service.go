package monitoring

import (
	"context"
	"log/slog"
	"time"
)

// Service provides the main performance monitoring service
type Service struct {
	collector *Collector
	alerter   *Alerter
	optimizer *Optimizer
	config    *MonitoringConfig
	logger    *slog.Logger
}

// NewService creates a new monitoring service
func NewService(config *MonitoringConfig, logger *slog.Logger) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	
	collector := NewCollector(config)
	alerter := NewAlerter(config, collector, logger)
	optimizer := NewOptimizer(config, collector)
	
	service := &Service{
		collector: collector,
		alerter:   alerter,
		optimizer: optimizer,
		config:    config,
		logger:    logger,
	}
	
	// Start background monitoring
	if config.Enabled {
		go service.startBackgroundMonitoring()
	}
	
	return service
}

// RecordMetric records a performance metric
func (s *Service) RecordMetric(metric PerformanceMetric) {
	s.collector.RecordMetric(metric)
}

// GetPerformanceSnapshot gets the current performance snapshot
func (s *Service) GetPerformanceSnapshot(ctx context.Context) PerformanceSnapshot {
	snapshot := s.collector.GetPerformanceSnapshot(ctx)
	snapshot.ActiveAlerts = s.alerter.GetActiveAlerts()
	return snapshot
}

// GetRealTimeMetrics gets real-time performance metrics
func (s *Service) GetRealTimeMetrics() RealTimeMetrics {
	return s.collector.GetRealTimeMetrics()
}

// GetResponseTimeStats gets response time statistics
func (s *Service) GetResponseTimeStats(endpoint, method string) ResponseTimeStats {
	return s.collector.GetResponseTimeStats(endpoint, method)
}

// GetThroughputStats gets throughput statistics
func (s *Service) GetThroughputStats(endpoint, method string) ThroughputStats {
	return s.collector.GetThroughputStats(endpoint, method)
}

// GetEndpointPerformance gets performance stats for an endpoint
func (s *Service) GetEndpointPerformance(endpoint, method string) EndpointPerformance {
	return s.collector.GetEndpointPerformance(endpoint, method)
}

// GetPerformanceTrend gets performance trend data
func (s *Service) GetPerformanceTrend(endpoint string, bucketSize time.Duration) PerformanceTrend {
	return s.collector.GetPerformanceTrend(endpoint, bucketSize)
}

// GetPerformanceReport generates a comprehensive performance report
func (s *Service) GetPerformanceReport(period string) PerformanceReport {
	snapshot := s.GetPerformanceSnapshot(context.Background())
	
	// Calculate summary
	summary := PerformanceSummary{
		TotalRequests:       snapshot.OverallResponseTime.Count,
		AverageResponseTime: snapshot.OverallResponseTime.Mean,
		P95ResponseTime:     snapshot.OverallResponseTime.P95,
		P99ResponseTime:     snapshot.OverallResponseTime.P99,
		Throughput:          snapshot.OverallThroughput.RequestsPerSecond,
		HealthScore:         snapshot.HealthScore,
		PerformanceGrade:    calculateGrade(snapshot.HealthScore),
	}
	
	// Get trends for each endpoint
	var trends []PerformanceTrend
	for _, ep := range snapshot.EndpointStats {
		trend := s.GetPerformanceTrend(ep.Endpoint, 1*time.Minute)
		trends = append(trends, trend)
	}
	
	// Get recommendations
	recommendations := s.GenerateRecommendations()
	
	return PerformanceReport{
		GeneratedAt:         time.Now(),
		Period:              period,
		Summary:             summary,
		EndpointPerformance: snapshot.EndpointStats,
		Trends:              trends,
		Alerts:              snapshot.ActiveAlerts,
		Recommendations:     recommendations,
	}
}

// GetActiveAlerts gets all active performance alerts
func (s *Service) GetActiveAlerts() []PerformanceAlert {
	return s.alerter.GetActiveAlerts()
}

// GetAllAlerts gets all alerts (active, resolved, silenced)
func (s *Service) GetAllAlerts() []PerformanceAlert {
	return s.alerter.GetAllAlerts()
}

// SilenceAlert silences a specific alert
func (s *Service) SilenceAlert(alertID string) error {
	return s.alerter.SilenceAlert(alertID)
}

// ResolveAlert manually resolves an alert
func (s *Service) ResolveAlert(alertID string) error {
	return s.alerter.ResolveAlert(alertID)
}

// RegisterAlertCallback registers a callback for alert notifications
func (s *Service) RegisterAlertCallback(callback AlertCallback) {
	s.alerter.RegisterCallback(callback)
}

// GenerateRecommendations generates optimization recommendations
func (s *Service) GenerateRecommendations() []OptimizationRecommendation {
	return s.optimizer.GenerateRecommendations()
}

// IncrementActiveRequests increments active request counter
func (s *Service) IncrementActiveRequests() {
	s.collector.IncrementActiveRequests()
}

// DecrementActiveRequests decrements active request counter
func (s *Service) DecrementActiveRequests() {
	s.collector.DecrementActiveRequests()
}

// startBackgroundMonitoring starts background monitoring tasks
func (s *Service) startBackgroundMonitoring() {
	// Start alert monitoring
	go s.alerter.StartMonitoring(30 * time.Second)
	
	// Periodically log performance summary
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		snapshot := s.GetPerformanceSnapshot(context.Background())
		
		s.logger.Info("Performance summary",
			"health_score", snapshot.HealthScore,
			"avg_response_time_ms", snapshot.OverallResponseTime.Mean.Milliseconds(),
			"p95_response_time_ms", snapshot.OverallResponseTime.P95.Milliseconds(),
			"throughput_rps", snapshot.OverallThroughput.RequestsPerSecond,
			"active_alerts", len(snapshot.ActiveAlerts),
			"goroutines", snapshot.SystemLoad.Goroutines,
			"memory_percent", snapshot.SystemLoad.Memory,
		)
	}
}

// Helper functions

func calculateGrade(healthScore float64) string {
	switch {
	case healthScore >= 90:
		return "A"
	case healthScore >= 80:
		return "B"
	case healthScore >= 70:
		return "C"
	case healthScore >= 60:
		return "D"
	default:
		return "F"
	}
}

