package monitoring

import (
	"time"
)

// PerformanceMetric represents a single performance measurement
type PerformanceMetric struct {
	Timestamp       time.Time     `json:"timestamp"`
	Endpoint        string        `json:"endpoint"`
	Method          string        `json:"method"`
	ResponseTime    time.Duration `json:"response_time"`
	StatusCode      int           `json:"status_code"`
	RequestSize     int64         `json:"request_size"`
	ResponseSize    int64         `json:"response_size"`
	Throughput      float64       `json:"throughput"` // requests per second
}

// ResponseTimeStats represents response time statistics
type ResponseTimeStats struct {
	Endpoint         string        `json:"endpoint"`
	Method           string        `json:"method"`
	Count            int64         `json:"count"`
	Min              time.Duration `json:"min"`
	Max              time.Duration `json:"max"`
	Mean             time.Duration `json:"mean"`
	Median           time.Duration `json:"median"`
	P95              time.Duration `json:"p95"`
	P99              time.Duration `json:"p99"`
	StandardDeviation time.Duration `json:"std_dev"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// ThroughputStats represents throughput statistics
type ThroughputStats struct {
	Endpoint            string    `json:"endpoint"`
	Method              string    `json:"method"`
	RequestsPerSecond   float64   `json:"requests_per_second"`
	RequestsPerMinute   float64   `json:"requests_per_minute"`
	RequestsPerHour     float64   `json:"requests_per_hour"`
	PeakRPS             float64   `json:"peak_rps"`
	PeakTime            time.Time `json:"peak_time"`
	BytesPerSecond      float64   `json:"bytes_per_second"`
	LastUpdated         time.Time `json:"last_updated"`
}

// PerformanceSnapshot represents a point-in-time performance snapshot
type PerformanceSnapshot struct {
	Timestamp             time.Time           `json:"timestamp"`
	OverallResponseTime   ResponseTimeStats   `json:"overall_response_time"`
	OverallThroughput     ThroughputStats     `json:"overall_throughput"`
	EndpointStats         []EndpointPerformance `json:"endpoint_stats"`
	SystemLoad            SystemLoad          `json:"system_load"`
	ActiveAlerts          []PerformanceAlert  `json:"active_alerts"`
	HealthScore           float64             `json:"health_score"` // 0-100
}

// EndpointPerformance represents performance data for a specific endpoint
type EndpointPerformance struct {
	Endpoint         string            `json:"endpoint"`
	Method           string            `json:"method"`
	ResponseTime     ResponseTimeStats `json:"response_time"`
	Throughput       ThroughputStats   `json:"throughput"`
	ErrorRate        float64           `json:"error_rate"`
	SuccessRate      float64           `json:"success_rate"`
	Saturation       float64           `json:"saturation"` // 0-100, how close to capacity
}

// SystemLoad represents overall system load metrics
type SystemLoad struct {
	CPU              float64   `json:"cpu_percent"`
	Memory           float64   `json:"memory_percent"`
	Goroutines       int       `json:"goroutines"`
	ActiveConnections int      `json:"active_connections"`
	QueueDepth       int       `json:"queue_depth"`
	Timestamp        time.Time `json:"timestamp"`
}

// PerformanceAlert represents a performance-related alert
type PerformanceAlert struct {
	ID          string             `json:"id"`
	Type        AlertType          `json:"type"`
	Severity    AlertSeverity      `json:"severity"`
	Endpoint    string             `json:"endpoint,omitempty"`
	Message     string             `json:"message"`
	Value       float64            `json:"value"`
	Threshold   float64            `json:"threshold"`
	TriggeredAt time.Time          `json:"triggered_at"`
	Status      AlertStatus        `json:"status"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
}

// AlertType defines the type of performance alert
type AlertType string

const (
	AlertTypeResponseTime    AlertType = "response_time"
	AlertTypeThroughput      AlertType = "throughput"
	AlertTypeErrorRate       AlertType = "error_rate"
	AlertTypeSaturation      AlertType = "saturation"
	AlertTypeSystemLoad      AlertType = "system_load"
	AlertTypeAnomalous       AlertType = "anomalous"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus defines the status of an alert
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusSilenced  AlertStatus = "silenced"
)

// OptimizationRecommendation represents a performance optimization suggestion
type OptimizationRecommendation struct {
	ID              string                    `json:"id"`
	Type            RecommendationType        `json:"type"`
	Priority        RecommendationPriority    `json:"priority"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	Endpoint        string                    `json:"endpoint,omitempty"`
	ImpactEstimate  string                    `json:"impact_estimate"`
	Effort          string                    `json:"effort"`
	Actions         []string                  `json:"actions"`
	Rationale       string                    `json:"rationale"`
	Metrics         map[string]float64        `json:"metrics"`
	GeneratedAt     time.Time                 `json:"generated_at"`
}

// RecommendationType defines the type of optimization recommendation
type RecommendationType string

const (
	RecommendationTypeCaching         RecommendationType = "caching"
	RecommendationTypeIndexing        RecommendationType = "indexing"
	RecommendationTypeQueryOptimization RecommendationType = "query_optimization"
	RecommendationTypeRateLimiting    RecommendationType = "rate_limiting"
	RecommendationTypeLoadBalancing   RecommendationType = "load_balancing"
	RecommendationTypeCodeOptimization RecommendationType = "code_optimization"
	RecommendationTypeScaling         RecommendationType = "scaling"
	RecommendationTypeConfiguration   RecommendationType = "configuration"
)

// RecommendationPriority defines the priority of a recommendation
type RecommendationPriority string

const (
	RecommendationPriorityHigh   RecommendationPriority = "high"
	RecommendationPriorityMedium RecommendationPriority = "medium"
	RecommendationPriorityLow    RecommendationPriority = "low"
)

// PerformanceTrend represents performance trends over time
type PerformanceTrend struct {
	Endpoint          string              `json:"endpoint"`
	Period            string              `json:"period"`
	DataPoints        []TrendDataPoint    `json:"data_points"`
	Trend             TrendDirection      `json:"trend"`
	Volatility        float64             `json:"volatility"`
	Forecast          []TrendDataPoint    `json:"forecast,omitempty"`
}

// TrendDataPoint represents a single point in a trend
type TrendDataPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime float64   `json:"response_time_ms"`
	Throughput   float64   `json:"throughput_rps"`
	ErrorRate    float64   `json:"error_rate"`
}

// TrendDirection indicates the direction of a trend
type TrendDirection string

const (
	TrendDirectionImproving  TrendDirection = "improving"
	TrendDirectionStable     TrendDirection = "stable"
	TrendDirectionDegrading  TrendDirection = "degrading"
)

// PerformanceReport represents a comprehensive performance report
type PerformanceReport struct {
	GeneratedAt        time.Time                      `json:"generated_at"`
	Period             string                         `json:"period"`
	Summary            PerformanceSummary             `json:"summary"`
	EndpointPerformance []EndpointPerformance         `json:"endpoint_performance"`
	Trends             []PerformanceTrend             `json:"trends"`
	Alerts             []PerformanceAlert             `json:"alerts"`
	Recommendations    []OptimizationRecommendation   `json:"recommendations"`
	Comparisons        *PerformanceComparison         `json:"comparisons,omitempty"`
}

// PerformanceSummary provides a high-level performance summary
type PerformanceSummary struct {
	TotalRequests        int64         `json:"total_requests"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	P95ResponseTime      time.Duration `json:"p95_response_time"`
	P99ResponseTime      time.Duration `json:"p99_response_time"`
	Throughput           float64       `json:"throughput_rps"`
	ErrorRate            float64       `json:"error_rate"`
	HealthScore          float64       `json:"health_score"`
	PerformanceGrade     string        `json:"performance_grade"` // A, B, C, D, F
}

// PerformanceComparison compares performance across time periods
type PerformanceComparison struct {
	CurrentPeriod  PerformanceSummary `json:"current_period"`
	PreviousPeriod PerformanceSummary `json:"previous_period"`
	ChangePercent  map[string]float64 `json:"change_percent"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled                  bool          `json:"enabled"`
	SampleRate               float64       `json:"sample_rate"` // 0.0 to 1.0
	WindowSize               time.Duration `json:"window_size"`
	AlertThresholds          AlertThresholds `json:"alert_thresholds"`
	RecommendationThresholds RecommendationThresholds `json:"recommendation_thresholds"`
}

// AlertThresholds defines thresholds for triggering alerts
type AlertThresholds struct {
	ResponseTimeP95Ms     float64 `json:"response_time_p95_ms"`
	ResponseTimeP99Ms     float64 `json:"response_time_p99_ms"`
	ErrorRatePercent      float64 `json:"error_rate_percent"`
	ThroughputDropPercent float64 `json:"throughput_drop_percent"`
	SaturationPercent     float64 `json:"saturation_percent"`
	CPUPercent            float64 `json:"cpu_percent"`
	MemoryPercent         float64 `json:"memory_percent"`
}

// RecommendationThresholds defines thresholds for generating recommendations
type RecommendationThresholds struct {
	SlowResponseTimeMs    float64 `json:"slow_response_time_ms"`
	HighErrorRatePercent  float64 `json:"high_error_rate_percent"`
	LowCacheHitPercent    float64 `json:"low_cache_hit_percent"`
	HighCPUPercent        float64 `json:"high_cpu_percent"`
	HighMemoryPercent     float64 `json:"high_memory_percent"`
}

// RealTimeMetrics represents real-time performance metrics
type RealTimeMetrics struct {
	Timestamp         time.Time `json:"timestamp"`
	CurrentRPS        float64   `json:"current_rps"`
	AverageLatencyMs  float64   `json:"average_latency_ms"`
	ActiveRequests    int       `json:"active_requests"`
	ErrorCount        int64     `json:"error_count"`
	SuccessCount      int64     `json:"success_count"`
	LastMinuteRequests int64    `json:"last_minute_requests"`
}

// PerformanceQuery represents a query for performance data
type PerformanceQuery struct {
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Endpoint     string    `json:"endpoint,omitempty"`
	Method       string    `json:"method,omitempty"`
	MinLatency   time.Duration `json:"min_latency,omitempty"`
	MaxLatency   time.Duration `json:"max_latency,omitempty"`
	GroupBy      string    `json:"group_by,omitempty"` // endpoint, method, time
	Aggregation  string    `json:"aggregation,omitempty"` // avg, min, max, p95, p99
}

