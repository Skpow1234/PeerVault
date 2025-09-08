package analytics

// InsightType defines the type of performance insight
type InsightType string

const (
	InsightTypeErrorRate       InsightType = "error_rate"
	InsightTypeResponseTime    InsightType = "response_time"
	InsightTypeThroughput      InsightType = "throughput"
	InsightTypeMemoryUsage     InsightType = "memory_usage"
	InsightTypeCPUUsage        InsightType = "cpu_usage"
	InsightTypeCacheHitRate    InsightType = "cache_hit_rate"
	InsightTypeQueryComplexity InsightType = "query_complexity"
	InsightTypePerformance     InsightType = "performance"
	InsightTypeComplexity      InsightType = "complexity"
	InsightTypeFrequency       InsightType = "frequency"
	InsightTypeMemory          InsightType = "memory"
	InsightTypeCPU             InsightType = "cpu"
	InsightTypeGoroutines      InsightType = "goroutines"
)

// Severity defines the severity level of an insight
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)
