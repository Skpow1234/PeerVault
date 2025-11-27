package analytics

import (
	"time"
)

// APICall represents a single API call record
type APICall struct {
	ID            string            `json:"id"`
	Timestamp     time.Time         `json:"timestamp"`
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	StatusCode    int               `json:"status_code"`
	Duration      time.Duration     `json:"duration"`
	UserID        string            `json:"user_id,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	IPAddress     string            `json:"ip_address,omitempty"`
	RequestSize   int64             `json:"request_size"`
	ResponseSize  int64             `json:"response_size"`
	ErrorMessage  string            `json:"error_message,omitempty"`
	APIVersion    string            `json:"api_version,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	QueryParams   map[string]string `json:"query_params,omitempty"`
	Endpoint      string            `json:"endpoint"`
}

// UsageMetrics represents aggregated usage metrics
type UsageMetrics struct {
	TotalRequests      int64                    `json:"total_requests"`
	SuccessfulRequests int64                    `json:"successful_requests"`
	FailedRequests     int64                    `json:"failed_requests"`
	AverageDuration    time.Duration            `json:"average_duration"`
	TotalDataIn        int64                    `json:"total_data_in"`
	TotalDataOut       int64                    `json:"total_data_out"`
	RequestsByMethod   map[string]int64         `json:"requests_by_method"`
	RequestsByPath     map[string]int64         `json:"requests_by_path"`
	RequestsByStatus   map[int]int64            `json:"requests_by_status"`
	TopEndpoints       []EndpointStats          `json:"top_endpoints"`
	ErrorRate          float64                  `json:"error_rate"`
	TimeWindow         TimeWindow               `json:"time_window"`
}

// EndpointStats represents statistics for a specific endpoint
type EndpointStats struct {
	Path               string        `json:"path"`
	Method             string        `json:"method"`
	TotalCalls         int64         `json:"total_calls"`
	SuccessfulCalls    int64         `json:"successful_calls"`
	FailedCalls        int64         `json:"failed_calls"`
	AverageDuration    time.Duration `json:"average_duration"`
	MinDuration        time.Duration `json:"min_duration"`
	MaxDuration        time.Duration `json:"max_duration"`
	TotalDataIn        int64         `json:"total_data_in"`
	TotalDataOut       int64         `json:"total_data_out"`
	ErrorRate          float64       `json:"error_rate"`
	LastCalled         time.Time     `json:"last_called"`
}

// UserBehaviorMetrics represents user behavior analysis
type UserBehaviorMetrics struct {
	UserID             string                   `json:"user_id"`
	TotalRequests      int64                    `json:"total_requests"`
	UniqueEndpoints    int                      `json:"unique_endpoints"`
	AverageRequestRate float64                  `json:"average_request_rate"` // requests per hour
	PeakHour           int                      `json:"peak_hour"`            // hour of day (0-23)
	FavoriteEndpoints  []EndpointStats          `json:"favorite_endpoints"`
	UserAgents         map[string]int64         `json:"user_agents"`
	IPAddresses        map[string]int64         `json:"ip_addresses"`
	FirstSeen          time.Time                `json:"first_seen"`
	LastSeen           time.Time                `json:"last_seen"`
	ErrorCount         int64                    `json:"error_count"`
}

// UsageTrend represents usage trends over time
type UsageTrend struct {
	Timestamp      time.Time `json:"timestamp"`
	RequestCount   int64     `json:"request_count"`
	SuccessCount   int64     `json:"success_count"`
	ErrorCount     int64     `json:"error_count"`
	AverageDuration float64  `json:"average_duration_ms"`
	DataIn         int64     `json:"data_in"`
	DataOut        int64     `json:"data_out"`
}

// PopularityMetrics represents API popularity metrics
type PopularityMetrics struct {
	TopEndpoints     []EndpointPopularity `json:"top_endpoints"`
	TrendingUp       []EndpointPopularity `json:"trending_up"`
	TrendingDown     []EndpointPopularity `json:"trending_down"`
	MostActiveUsers  []UserActivity       `json:"most_active_users"`
	TimeWindow       TimeWindow           `json:"time_window"`
}

// EndpointPopularity represents popularity metrics for an endpoint
type EndpointPopularity struct {
	Endpoint       string  `json:"endpoint"`
	CallCount      int64   `json:"call_count"`
	UniqueUsers    int     `json:"unique_users"`
	GrowthRate     float64 `json:"growth_rate"` // percentage change
	TrendScore     float64 `json:"trend_score"`
	LastDayCount   int64   `json:"last_day_count"`
	PreviousDayCount int64 `json:"previous_day_count"`
}

// UserActivity represents user activity metrics
type UserActivity struct {
	UserID         string    `json:"user_id"`
	RequestCount   int64     `json:"request_count"`
	EndpointCount  int       `json:"endpoint_count"`
	LastActivity   time.Time `json:"last_activity"`
}

// TimeWindow represents a time range for analytics
type TimeWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnalyticsQuery represents a query for analytics data
type AnalyticsQuery struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Endpoint     string    `json:"endpoint,omitempty"`
	Method       string    `json:"method,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	StatusCode   int       `json:"status_code,omitempty"`
	MinDuration  time.Duration `json:"min_duration,omitempty"`
	MaxDuration  time.Duration `json:"max_duration,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
	GroupBy      string    `json:"group_by,omitempty"` // "hour", "day", "week", "month"
}

// AnalyticsSummary represents a high-level summary of analytics
type AnalyticsSummary struct {
	OverallMetrics    UsageMetrics          `json:"overall_metrics"`
	TopEndpoints      []EndpointStats       `json:"top_endpoints"`
	RecentTrends      []*UsageTrend         `json:"recent_trends"`
	ActiveUsers       int                   `json:"active_users"`
	NewUsersToday     int                   `json:"new_users_today"`
	SystemHealth      SystemHealth          `json:"system_health"`
	GeneratedAt       time.Time             `json:"generated_at"`
}

// SystemHealth represents the health status based on analytics
type SystemHealth struct {
	Status          string  `json:"status"` // "healthy", "degraded", "unhealthy"
	ErrorRate       float64 `json:"error_rate"`
	AverageLatency  float64 `json:"average_latency_ms"`
	RequestRate     float64 `json:"request_rate"` // requests per second
	Warnings        []string `json:"warnings,omitempty"`
}

// TimeSeriesPoint represents a single point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Values    map[string]interface{} `json:"values"`
}

// AnalyticsFilter represents filters for analytics queries
type AnalyticsFilter struct {
	Paths       []string  `json:"paths,omitempty"`
	Methods     []string  `json:"methods,omitempty"`
	StatusCodes []int     `json:"status_codes,omitempty"`
	UserIDs     []string  `json:"user_ids,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
}

