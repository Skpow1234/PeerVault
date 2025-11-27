package api

import (
	"time"

	"github.com/Skpow1234/Peervault/internal/api/analytics"
)

// APIAnalytics provides high-level analytics API
type APIAnalytics struct {
	service *analytics.Service
}

// NewAPIAnalytics creates a new API analytics instance
func NewAPIAnalytics(service *analytics.Service) *APIAnalytics {
	return &APIAnalytics{
		service: service,
	}
}

// GetDashboardData returns comprehensive dashboard data
func (a *APIAnalytics) GetDashboardData(period string) (*DashboardData, error) {
	startTime, endTime := analytics.GetTimeWindow(period)
	
	summary, err := a.service.GetAnalyticsSummary()
	if err != nil {
		return nil, err
	}
	
	metrics, err := a.service.GetUsageMetrics(startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	popularity, err := a.service.GetPopularityMetrics(startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	trends, err := a.service.GetUsageTrends(startTime, endTime, "hour")
	if err != nil {
		return nil, err
	}
	
	return &DashboardData{
		Summary:    summary,
		Metrics:    metrics,
		Popularity: popularity,
		Trends:     trends,
		Period:     period,
		GeneratedAt: time.Now(),
	}, nil
}

// GetTopEndpoints returns the most popular endpoints
func (a *APIAnalytics) GetTopEndpoints(limit int, period string) ([]*analytics.EndpointStats, error) {
	startTime, endTime := analytics.GetTimeWindow(period)
	
	metrics, err := a.service.GetUsageMetrics(startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	if len(metrics.TopEndpoints) > limit {
		result := make([]*analytics.EndpointStats, limit)
		for i := 0; i < limit; i++ {
			stats := metrics.TopEndpoints[i]
			result[i] = &stats
		}
		return result, nil
	}
	
	result := make([]*analytics.EndpointStats, len(metrics.TopEndpoints))
	for i := range metrics.TopEndpoints {
		stats := metrics.TopEndpoints[i]
		result[i] = &stats
	}
	return result, nil
}

// GetUserInsights returns insights for a specific user
func (a *APIAnalytics) GetUserInsights(userID string, period string) (*UserInsights, error) {
	startTime, endTime := analytics.GetTimeWindow(period)
	
	behavior, err := a.service.GetUserBehavior(userID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	// Get user's API calls
	query := &analytics.AnalyticsQuery{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     100,
	}
	
	calls, err := a.service.QueryAPICalls(query)
	if err != nil {
		return nil, err
	}
	
	// Calculate additional insights
	insights := &UserInsights{
		UserID:          userID,
		Behavior:        behavior,
		RecentCalls:     calls,
		Period:          period,
		AnalyzedAt:      time.Now(),
	}
	
	// Calculate usage patterns
	if len(calls) > 0 {
		insights.UsagePattern = analyzeUsagePattern(calls)
	}
	
	return insights, nil
}

// GetPerformanceReport generates a performance report
func (a *APIAnalytics) GetPerformanceReport(period string) (*PerformanceReport, error) {
	startTime, endTime := analytics.GetTimeWindow(period)
	
	metrics, err := a.service.GetUsageMetrics(startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	// Get all calls for detailed analysis
	query := &analytics.AnalyticsQuery{
		StartTime: startTime,
		EndTime:   endTime,
	}
	
	calls, err := a.service.QueryAPICalls(query)
	if err != nil {
		return nil, err
	}
	
	report := &PerformanceReport{
		Period:          period,
		TotalRequests:   metrics.TotalRequests,
		SuccessRate:     float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100,
		ErrorRate:       metrics.ErrorRate,
		AverageLatency:  float64(metrics.AverageDuration.Milliseconds()),
		GeneratedAt:     time.Now(),
	}
	
	if len(calls) > 0 {
		report.P50Latency = float64(analytics.CalculatePercentileLatency(calls, 50).Milliseconds())
		report.P95Latency = float64(analytics.CalculatePercentileLatency(calls, 95).Milliseconds())
		report.P99Latency = float64(analytics.CalculatePercentileLatency(calls, 99).Milliseconds())
	}
	
	// Get slowest endpoints
	report.SlowestEndpoints = getSlowestEndpoints(metrics.TopEndpoints, 5)
	
	// Get endpoints with highest error rates
	report.ErrorProneEndpoints = getErrorProneEndpoints(metrics.TopEndpoints, 5)
	
	return report, nil
}

// GetTrendAnalysis analyzes trends over time
func (a *APIAnalytics) GetTrendAnalysis(period string, interval string) (*TrendAnalysis, error) {
	startTime, endTime := analytics.GetTimeWindow(period)
	
	trends, err := a.service.GetUsageTrends(startTime, endTime, interval)
	if err != nil {
		return nil, err
	}
	
	analysis := &TrendAnalysis{
		Period:      period,
		Interval:    interval,
		Trends:      trends,
		GeneratedAt: time.Now(),
	}
	
	// Calculate trend direction
	if len(trends) >= 2 {
		firstPeriod := trends[0]
		lastPeriod := trends[len(trends)-1]
		
		if lastPeriod.RequestCount > firstPeriod.RequestCount {
			analysis.TrendDirection = "increasing"
			if firstPeriod.RequestCount > 0 {
				analysis.TrendPercentage = float64(lastPeriod.RequestCount-firstPeriod.RequestCount) / float64(firstPeriod.RequestCount) * 100
			}
		} else if lastPeriod.RequestCount < firstPeriod.RequestCount {
			analysis.TrendDirection = "decreasing"
			if firstPeriod.RequestCount > 0 {
				analysis.TrendPercentage = float64(firstPeriod.RequestCount-lastPeriod.RequestCount) / float64(firstPeriod.RequestCount) * 100
			}
		} else {
			analysis.TrendDirection = "stable"
		}
	}
	
	return analysis, nil
}

// Helper types

// DashboardData represents comprehensive dashboard data
type DashboardData struct {
	Summary     *analytics.AnalyticsSummary  `json:"summary"`
	Metrics     *analytics.UsageMetrics      `json:"metrics"`
	Popularity  *analytics.PopularityMetrics `json:"popularity"`
	Trends      []*analytics.UsageTrend      `json:"trends"`
	Period      string                       `json:"period"`
	GeneratedAt time.Time                    `json:"generated_at"`
}

// UserInsights represents insights for a specific user
type UserInsights struct {
	UserID       string                        `json:"user_id"`
	Behavior     *analytics.UserBehaviorMetrics `json:"behavior"`
	RecentCalls  []*analytics.APICall          `json:"recent_calls"`
	UsagePattern string                        `json:"usage_pattern"`
	Period       string                        `json:"period"`
	AnalyzedAt   time.Time                     `json:"analyzed_at"`
}

// PerformanceReport represents a performance analysis report
type PerformanceReport struct {
	Period              string                      `json:"period"`
	TotalRequests       int64                       `json:"total_requests"`
	SuccessRate         float64                     `json:"success_rate"`
	ErrorRate           float64                     `json:"error_rate"`
	AverageLatency      float64                     `json:"average_latency_ms"`
	P50Latency          float64                     `json:"p50_latency_ms"`
	P95Latency          float64                     `json:"p95_latency_ms"`
	P99Latency          float64                     `json:"p99_latency_ms"`
	SlowestEndpoints    []*analytics.EndpointStats  `json:"slowest_endpoints"`
	ErrorProneEndpoints []*analytics.EndpointStats  `json:"error_prone_endpoints"`
	GeneratedAt         time.Time                   `json:"generated_at"`
}

// TrendAnalysis represents trend analysis over time
type TrendAnalysis struct {
	Period          string                  `json:"period"`
	Interval        string                  `json:"interval"`
	Trends          []*analytics.UsageTrend `json:"trends"`
	TrendDirection  string                  `json:"trend_direction"`
	TrendPercentage float64                 `json:"trend_percentage"`
	GeneratedAt     time.Time               `json:"generated_at"`
}

// Helper functions

func analyzeUsagePattern(calls []*analytics.APICall) string {
	if len(calls) == 0 {
		return "no data"
	}
	
	// Count calls by hour
	hourCounts := make(map[int]int)
	for _, call := range calls {
		hour := call.Timestamp.Hour()
		hourCounts[hour]++
	}
	
	// Find peak hour
	maxCount := 0
	peakHour := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}
	
	// Determine pattern
	if peakHour >= 9 && peakHour <= 17 {
		return "business hours"
	} else if peakHour >= 18 && peakHour <= 23 {
		return "evening"
	} else {
		return "off-hours"
	}
}

func getSlowestEndpoints(endpoints []analytics.EndpointStats, limit int) []*analytics.EndpointStats {
	// Create a copy and sort by average duration
	sorted := make([]analytics.EndpointStats, len(endpoints))
	copy(sorted, endpoints)
	
	// Simple bubble sort by duration
	for i := 0; i < len(sorted) && i < limit; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].AverageDuration > sorted[i].AverageDuration {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	// Convert to pointers
	result := make([]*analytics.EndpointStats, 0, limit)
	for i := 0; i < len(sorted) && i < limit; i++ {
		stats := sorted[i]
		result = append(result, &stats)
	}
	
	return result
}

func getErrorProneEndpoints(endpoints []analytics.EndpointStats, limit int) []*analytics.EndpointStats {
	// Create a copy and sort by error rate
	sorted := make([]analytics.EndpointStats, len(endpoints))
	copy(sorted, endpoints)
	
	// Simple bubble sort by error rate
	for i := 0; i < len(sorted) && i < limit; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].ErrorRate > sorted[i].ErrorRate {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	// Convert to pointers
	result := make([]*analytics.EndpointStats, 0, limit)
	for i := 0; i < len(sorted) && i < limit; i++ {
		stats := sorted[i]
		result = append(result, &stats)
	}
	
	return result
}

