package api

import (
	"time"

	"github.com/Skpow1234/Peervault/internal/api/monitoring"
)

// MonitoringAPI provides high-level monitoring functionality
type MonitoringAPI struct {
	service *monitoring.Service
}

// NewMonitoringAPI creates a new monitoring API instance
func NewMonitoringAPI(service *monitoring.Service) *MonitoringAPI {
	return &MonitoringAPI{
		service: service,
	}
}

// GetHealthStatus returns simplified health status
func (m *MonitoringAPI) GetHealthStatus() *HealthStatus {
	snapshot := m.service.GetPerformanceSnapshot(nil)
	
	status := &HealthStatus{
		Score:     snapshot.HealthScore,
		Grade:     calculateGrade(snapshot.HealthScore),
		Status:    getStatus(snapshot.HealthScore),
		Timestamp: time.Now(),
		Metrics: HealthMetrics{
			AvgResponseTimeMs: float64(snapshot.OverallResponseTime.Mean.Milliseconds()),
			P95ResponseTimeMs: float64(snapshot.OverallResponseTime.P95.Milliseconds()),
			ThroughputRPS:     snapshot.OverallThroughput.RequestsPerSecond,
			ActiveAlerts:      len(snapshot.ActiveAlerts),
			MemoryPercent:     snapshot.SystemLoad.Memory,
		},
	}
	
	// Add warnings
	if snapshot.HealthScore < 60 {
		status.Warnings = append(status.Warnings, "Performance degraded")
	}
	if len(snapshot.ActiveAlerts) > 0 {
		status.Warnings = append(status.Warnings, "Active performance alerts")
	}
	
	return status
}

// GetPerformanceInsights provides actionable performance insights
func (m *MonitoringAPI) GetPerformanceInsights() *PerformanceInsights {
	snapshot := m.service.GetPerformanceSnapshot(nil)
	recommendations := m.service.GenerateRecommendations()
	
	// Find slowest endpoints
	slowest := make([]EndpointInsight, 0, 5)
	for i := 0; i < len(snapshot.EndpointStats) && i < 5; i++ {
		ep := snapshot.EndpointStats[i]
		slowest = append(slowest, EndpointInsight{
			Endpoint:          ep.Endpoint,
			AvgResponseTimeMs: float64(ep.ResponseTime.Mean.Milliseconds()),
			P95ResponseTimeMs: float64(ep.ResponseTime.P95.Milliseconds()),
			ThroughputRPS:     ep.Throughput.RequestsPerSecond,
			ErrorRate:         ep.ErrorRate,
			Recommendation:    getSlowestEndpointRecommendation(ep),
		})
	}
	
	// Find most error-prone endpoints
	errorProne := make([]EndpointInsight, 0)
	for _, ep := range snapshot.EndpointStats {
		if ep.ErrorRate > 5.0 {
			errorProne = append(errorProne, EndpointInsight{
				Endpoint:          ep.Endpoint,
				AvgResponseTimeMs: float64(ep.ResponseTime.Mean.Milliseconds()),
				ErrorRate:         ep.ErrorRate,
				Recommendation:    "Investigate and fix errors",
			})
		}
	}
	
	// Calculate improvement potential
	potential := calculateImprovementPotential(snapshot, recommendations)
	
	return &PerformanceInsights{
		SlowestEndpoints:       slowest,
		ErrorProneEndpoints:    errorProne,
		TopRecommendations:     getTopRecommendations(recommendations, 5),
		ImprovementPotential:   potential,
		CurrentHealthScore:     snapshot.HealthScore,
		PotentialHealthScore:   snapshot.HealthScore + potential.ScoreImprovement,
		GeneratedAt:            time.Now(),
	}
}

// GetAlertSummary returns summarized alert information
func (m *MonitoringAPI) GetAlertSummary() *AlertSummary {
	alerts := m.service.GetActiveAlerts()
	
	summary := &AlertSummary{
		TotalAlerts:    len(alerts),
		CriticalCount:  0,
		WarningCount:   0,
		InfoCount:      0,
		Timestamp:      time.Now(),
	}
	
	// Count by severity
	for _, alert := range alerts {
		switch alert.Severity {
		case monitoring.AlertSeverityCritical:
			summary.CriticalCount++
			summary.CriticalAlerts = append(summary.CriticalAlerts, AlertInfo{
				Type:    string(alert.Type),
				Message: alert.Message,
				Value:   alert.Value,
			})
		case monitoring.AlertSeverityWarning:
			summary.WarningCount++
			summary.WarningAlerts = append(summary.WarningAlerts, AlertInfo{
				Type:    string(alert.Type),
				Message: alert.Message,
				Value:   alert.Value,
			})
		case monitoring.AlertSeverityInfo:
			summary.InfoCount++
		}
	}
	
	return summary
}

// GetQuickStats returns quick performance statistics
func (m *MonitoringAPI) GetQuickStats() *QuickStats {
	realtime := m.service.GetRealTimeMetrics()
	snapshot := m.service.GetPerformanceSnapshot(nil)
	
	return &QuickStats{
		CurrentRPS:        realtime.CurrentRPS,
		AvgLatencyMs:      realtime.AverageLatencyMs,
		ActiveRequests:    realtime.ActiveRequests,
		SuccessRate:       calculateSuccessRate(realtime),
		HealthScore:       snapshot.HealthScore,
		ActiveAlerts:      len(snapshot.ActiveAlerts),
		Timestamp:         time.Now(),
	}
}

// Types

type HealthStatus struct {
	Score     float64       `json:"score"`
	Grade     string        `json:"grade"`
	Status    string        `json:"status"`
	Metrics   HealthMetrics `json:"metrics"`
	Warnings  []string      `json:"warnings,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

type HealthMetrics struct {
	AvgResponseTimeMs float64 `json:"avg_response_time_ms"`
	P95ResponseTimeMs float64 `json:"p95_response_time_ms"`
	ThroughputRPS     float64 `json:"throughput_rps"`
	ActiveAlerts      int     `json:"active_alerts"`
	MemoryPercent     float64 `json:"memory_percent"`
}

type PerformanceInsights struct {
	SlowestEndpoints      []EndpointInsight                   `json:"slowest_endpoints"`
	ErrorProneEndpoints   []EndpointInsight                   `json:"error_prone_endpoints"`
	TopRecommendations    []monitoring.OptimizationRecommendation `json:"top_recommendations"`
	ImprovementPotential  ImprovementPotential                `json:"improvement_potential"`
	CurrentHealthScore    float64                             `json:"current_health_score"`
	PotentialHealthScore  float64                             `json:"potential_health_score"`
	GeneratedAt           time.Time                           `json:"generated_at"`
}

type EndpointInsight struct {
	Endpoint          string  `json:"endpoint"`
	AvgResponseTimeMs float64 `json:"avg_response_time_ms"`
	P95ResponseTimeMs float64 `json:"p95_response_time_ms,omitempty"`
	ThroughputRPS     float64 `json:"throughput_rps,omitempty"`
	ErrorRate         float64 `json:"error_rate,omitempty"`
	Recommendation    string  `json:"recommendation"`
}

type ImprovementPotential struct {
	LatencyReduction  float64 `json:"latency_reduction_percent"`
	ErrorReduction    float64 `json:"error_reduction_percent"`
	ScoreImprovement  float64 `json:"score_improvement"`
	EstimatedImpact   string  `json:"estimated_impact"`
}

type AlertSummary struct {
	TotalAlerts    int         `json:"total_alerts"`
	CriticalCount  int         `json:"critical_count"`
	WarningCount   int         `json:"warning_count"`
	InfoCount      int         `json:"info_count"`
	CriticalAlerts []AlertInfo `json:"critical_alerts,omitempty"`
	WarningAlerts  []AlertInfo `json:"warning_alerts,omitempty"`
	Timestamp      time.Time   `json:"timestamp"`
}

type AlertInfo struct {
	Type    string  `json:"type"`
	Message string  `json:"message"`
	Value   float64 `json:"value"`
}

type QuickStats struct {
	CurrentRPS     float64   `json:"current_rps"`
	AvgLatencyMs   float64   `json:"avg_latency_ms"`
	ActiveRequests int       `json:"active_requests"`
	SuccessRate    float64   `json:"success_rate"`
	HealthScore    float64   `json:"health_score"`
	ActiveAlerts   int       `json:"active_alerts"`
	Timestamp      time.Time `json:"timestamp"`
}

// Helper functions

func calculateGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func getStatus(score float64) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "good"
	case score >= 60:
		return "fair"
	case score >= 40:
		return "poor"
	default:
		return "critical"
	}
}

func getSlowestEndpointRecommendation(ep monitoring.EndpointPerformance) string {
	avgMs := float64(ep.ResponseTime.Mean.Milliseconds())
	
	if avgMs > 500 {
		return "Critical: Implement caching and optimize queries"
	} else if avgMs > 200 {
		return "Consider caching or code optimization"
	} else if avgMs > 100 {
		return "Minor optimizations possible"
	}
	return "Performance acceptable"
}

func getTopRecommendations(recs []monitoring.OptimizationRecommendation, limit int) []monitoring.OptimizationRecommendation {
	if len(recs) <= limit {
		return recs
	}
	
	// Sort by priority (high first)
	sorted := make([]monitoring.OptimizationRecommendation, len(recs))
	copy(sorted, recs)
	
	// Simple sort by priority
	for i := 0; i < len(sorted) && i < limit; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if getPriorityValue(sorted[j].Priority) > getPriorityValue(sorted[i].Priority) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	return sorted[:limit]
}

func getPriorityValue(priority monitoring.RecommendationPriority) int {
	switch priority {
	case monitoring.RecommendationPriorityHigh:
		return 3
	case monitoring.RecommendationPriorityMedium:
		return 2
	case monitoring.RecommendationPriorityLow:
		return 1
	default:
		return 0
	}
}

func calculateImprovementPotential(snapshot monitoring.PerformanceSnapshot, recommendations []monitoring.OptimizationRecommendation) ImprovementPotential {
	potential := ImprovementPotential{
		EstimatedImpact: "Moderate",
	}
	
	// Estimate latency reduction based on slow endpoints
	slowCount := 0
	for _, ep := range snapshot.EndpointStats {
		if float64(ep.ResponseTime.Mean.Milliseconds()) > 200 {
			slowCount++
		}
	}
	
	if slowCount > 0 {
		potential.LatencyReduction = float64(slowCount) / float64(len(snapshot.EndpointStats)) * 50
	}
	
	// Estimate error reduction
	highErrorCount := 0
	for _, ep := range snapshot.EndpointStats {
		if ep.ErrorRate > 3.0 {
			highErrorCount++
		}
	}
	
	if highErrorCount > 0 {
		potential.ErrorReduction = float64(highErrorCount) / float64(len(snapshot.EndpointStats)) * 60
	}
	
	// Estimate score improvement
	potential.ScoreImprovement = (potential.LatencyReduction + potential.ErrorReduction) / 4
	
	if potential.ScoreImprovement > 15 {
		potential.EstimatedImpact = "High"
	} else if potential.ScoreImprovement > 5 {
		potential.EstimatedImpact = "Moderate"
	} else {
		potential.EstimatedImpact = "Low"
	}
	
	return potential
}

func calculateSuccessRate(metrics monitoring.RealTimeMetrics) float64 {
	total := metrics.SuccessCount + metrics.ErrorCount
	if total == 0 {
		return 100.0
	}
	return float64(metrics.SuccessCount) / float64(total) * 100
}

