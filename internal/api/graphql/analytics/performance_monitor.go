package analytics

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// PerformanceMonitor monitors GraphQL server performance
type PerformanceMonitor struct {
	metrics   *PerformanceMetrics
	mu        sync.RWMutex
	logger    *slog.Logger
	startTime time.Time
	config    *PerformanceMonitorConfig
}

// PerformanceMetrics holds comprehensive performance metrics
type PerformanceMetrics struct {
	// Request metrics
	TotalRequests       int64         `json:"totalRequests"`
	SuccessfulRequests  int64         `json:"successfulRequests"`
	FailedRequests      int64         `json:"failedRequests"`
	AverageResponseTime time.Duration `json:"averageResponseTime"`
	MinResponseTime     time.Duration `json:"minResponseTime"`
	MaxResponseTime     time.Duration `json:"maxResponseTime"`
	TotalResponseTime   time.Duration `json:"totalResponseTime"`

	// Throughput metrics
	RequestsPerSecond float64 `json:"requestsPerSecond"`
	PeakRPS           float64 `json:"peakRPS"`
	CurrentRPS        float64 `json:"currentRPS"`

	// Error metrics
	ErrorRate  float64          `json:"errorRate"`
	ErrorCount int64            `json:"errorCount"`
	ErrorTypes map[string]int64 `json:"errorTypes"`

	// System metrics
	MemoryUsage    int64         `json:"memoryUsage"`
	MemoryPeak     int64         `json:"memoryPeak"`
	CPUUsage       float64       `json:"cpuUsage"`
	GoroutineCount int           `json:"goroutineCount"`
	GCPauseTime    time.Duration `json:"gcPauseTime"`

	// Query metrics
	QueryComplexity int `json:"queryComplexity"`
	QueryDepth      int `json:"queryDepth"`
	FieldCount      int `json:"fieldCount"`

	// Time-based metrics
	Uptime      time.Duration `json:"uptime"`
	LastUpdate  time.Time     `json:"lastUpdate"`
	LastRequest time.Time     `json:"lastRequest"`

	// Performance buckets
	ResponseTimeBuckets map[string]int64 `json:"responseTimeBuckets"`
	RequestSizeBuckets  map[string]int64 `json:"requestSizeBuckets"`
}

// PerformanceMonitorConfig holds configuration for performance monitoring
type PerformanceMonitorConfig struct {
	EnableSystemMetrics      bool            `json:"enableSystemMetrics"`
	EnableQueryMetrics       bool            `json:"enableQueryMetrics"`
	EnableErrorTracking      bool            `json:"enableErrorTracking"`
	EnableThroughputTracking bool            `json:"enableThroughputTracking"`
	UpdateInterval           time.Duration   `json:"updateInterval"`
	RetentionPeriod          time.Duration   `json:"retentionPeriod"`
	RPSWindowSize            time.Duration   `json:"rpsWindowSize"`
	ResponseTimeBuckets      []time.Duration `json:"responseTimeBuckets"`
	RequestSizeBuckets       []int64         `json:"requestSizeBuckets"`
}

// DefaultPerformanceMonitorConfig returns the default configuration
func DefaultPerformanceMonitorConfig() *PerformanceMonitorConfig {
	return &PerformanceMonitorConfig{
		EnableSystemMetrics:      true,
		EnableQueryMetrics:       true,
		EnableErrorTracking:      true,
		EnableThroughputTracking: true,
		UpdateInterval:           1 * time.Second,
		RetentionPeriod:          1 * time.Hour,
		RPSWindowSize:            1 * time.Minute,
		ResponseTimeBuckets: []time.Duration{
			1 * time.Millisecond,
			10 * time.Millisecond,
			100 * time.Millisecond,
			1 * time.Second,
			10 * time.Second,
		},
		RequestSizeBuckets: []int64{
			1024,     // 1KB
			10240,    // 10KB
			102400,   // 100KB
			1048576,  // 1MB
			10485760, // 10MB
		},
	}
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *PerformanceMonitorConfig, logger *slog.Logger) *PerformanceMonitor {
	if config == nil {
		config = DefaultPerformanceMonitorConfig()
	}

	pm := &PerformanceMonitor{
		metrics: &PerformanceMetrics{
			ErrorTypes:          make(map[string]int64),
			ResponseTimeBuckets: make(map[string]int64),
			RequestSizeBuckets:  make(map[string]int64),
		},
		logger:    logger,
		startTime: time.Now(),
		config:    config,
	}

	// Start background monitoring
	if config.EnableSystemMetrics {
		go pm.startSystemMonitoring()
	}

	return pm
}

// RecordRequest records a request execution
func (pm *PerformanceMonitor) RecordRequest(duration time.Duration, success bool, errorType string, requestSize int64, queryComplexity int, queryDepth int, fieldCount int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	pm.metrics.TotalRequests++
	pm.metrics.TotalResponseTime += duration
	pm.metrics.LastRequest = now
	pm.metrics.LastUpdate = now

	// Update response time metrics
	if pm.metrics.MinResponseTime == 0 || duration < pm.metrics.MinResponseTime {
		pm.metrics.MinResponseTime = duration
	}
	if duration > pm.metrics.MaxResponseTime {
		pm.metrics.MaxResponseTime = duration
	}

	// Update success/failure counts
	if success {
		pm.metrics.SuccessfulRequests++
	} else {
		pm.metrics.FailedRequests++
		pm.metrics.ErrorCount++
		if pm.config.EnableErrorTracking && errorType != "" {
			pm.metrics.ErrorTypes[errorType]++
		}
	}

	// Update query metrics
	if pm.config.EnableQueryMetrics {
		pm.metrics.QueryComplexity = queryComplexity
		pm.metrics.QueryDepth = queryDepth
		pm.metrics.FieldCount = fieldCount
	}

	// Update response time buckets
	if pm.config.EnableThroughputTracking {
		pm.updateResponseTimeBuckets(duration)
		pm.updateRequestSizeBuckets(requestSize)
	}

	// Update average response time
	if pm.metrics.TotalRequests > 0 {
		pm.metrics.AverageResponseTime = pm.metrics.TotalResponseTime / time.Duration(pm.metrics.TotalRequests)
	}

	// Update error rate
	if pm.metrics.TotalRequests > 0 {
		pm.metrics.ErrorRate = float64(pm.metrics.FailedRequests) / float64(pm.metrics.TotalRequests)
	}

	// Update uptime
	pm.metrics.Uptime = time.Since(pm.startTime)
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Update system metrics
	if pm.config.EnableSystemMetrics {
		pm.updateSystemMetrics()
	}

	// Update throughput metrics
	if pm.config.EnableThroughputTracking {
		pm.updateThroughputMetrics()
	}

	// Create a copy to avoid race conditions
	metrics := *pm.metrics
	return &metrics
}

// GetPerformanceReport returns a comprehensive performance report
func (pm *PerformanceMonitor) GetPerformanceReport() map[string]interface{} {
	metrics := pm.GetMetrics()

	report := map[string]interface{}{
		"metrics":         metrics,
		"insights":        pm.generatePerformanceInsights(),
		"recommendations": pm.generateRecommendations(),
		"generatedAt":     time.Now(),
	}

	return report
}

// GetPerformanceInsights returns insights about performance
func (pm *PerformanceMonitor) GetPerformanceInsights() []PerformanceInsight {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var insights []PerformanceInsight

	// High error rate insight
	if pm.metrics.ErrorRate > 0.05 { // 5% error rate
		insights = append(insights, PerformanceInsight{
			Type:           InsightTypeErrorRate,
			Severity:       SeverityWarning,
			Title:          "High Error Rate",
			Description:    fmt.Sprintf("Error rate is %.2f%%, which is above the recommended 5%%", pm.metrics.ErrorRate*100),
			Recommendation: "Investigate and fix the root cause of errors",
			Metadata: map[string]interface{}{
				"errorRate": pm.metrics.ErrorRate,
				"threshold": 0.05,
			},
		})
	}

	// Slow response time insight
	if pm.metrics.AverageResponseTime > 500*time.Millisecond {
		insights = append(insights, PerformanceInsight{
			Type:           InsightTypeResponseTime,
			Severity:       SeverityWarning,
			Title:          "Slow Response Time",
			Description:    fmt.Sprintf("Average response time is %v, which is above the recommended 500ms", pm.metrics.AverageResponseTime),
			Recommendation: "Optimize queries, implement caching, or scale resources",
			Metadata: map[string]interface{}{
				"averageResponseTime": pm.metrics.AverageResponseTime,
				"threshold":           500 * time.Millisecond,
			},
		})
	}

	// High memory usage insight
	if pm.metrics.MemoryUsage > 1024*1024*1024 { // 1GB
		insights = append(insights, PerformanceInsight{
			Type:           InsightTypeMemory,
			Severity:       SeverityWarning,
			Title:          "High Memory Usage",
			Description:    fmt.Sprintf("Memory usage is %d MB, which is above the recommended 1GB", pm.metrics.MemoryUsage/(1024*1024)),
			Recommendation: "Monitor memory usage and consider optimizing queries or increasing memory",
			Metadata: map[string]interface{}{
				"memoryUsage": pm.metrics.MemoryUsage,
				"threshold":   1024 * 1024 * 1024,
			},
		})
	}

	// High CPU usage insight
	if pm.metrics.CPUUsage > 80.0 {
		insights = append(insights, PerformanceInsight{
			Type:           InsightTypeCPU,
			Severity:       SeverityWarning,
			Title:          "High CPU Usage",
			Description:    fmt.Sprintf("CPU usage is %.2f%%, which is above the recommended 80%%", pm.metrics.CPUUsage),
			Recommendation: "Optimize queries, implement caching, or scale resources",
			Metadata: map[string]interface{}{
				"cpuUsage":  pm.metrics.CPUUsage,
				"threshold": 80.0,
			},
		})
	}

	// High goroutine count insight
	if pm.metrics.GoroutineCount > 1000 {
		insights = append(insights, PerformanceInsight{
			Type:           InsightTypeGoroutines,
			Severity:       SeverityWarning,
			Title:          "High Goroutine Count",
			Description:    fmt.Sprintf("Goroutine count is %d, which is above the recommended 1000", pm.metrics.GoroutineCount),
			Recommendation: "Investigate goroutine leaks or optimize concurrent operations",
			Metadata: map[string]interface{}{
				"goroutineCount": pm.metrics.GoroutineCount,
				"threshold":      1000,
			},
		})
	}

	return insights
}

// startSystemMonitoring starts background system monitoring
func (pm *PerformanceMonitor) startSystemMonitoring() {
	ticker := time.NewTicker(pm.config.UpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		pm.updateSystemMetrics()
	}
}

// updateSystemMetrics updates system-related metrics
func (pm *PerformanceMonitor) updateSystemMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Update memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	pm.metrics.MemoryUsage = int64(m.Alloc)
	if pm.metrics.MemoryUsage > pm.metrics.MemoryPeak {
		pm.metrics.MemoryPeak = pm.metrics.MemoryUsage
	}

	// Update goroutine count
	pm.metrics.GoroutineCount = runtime.NumGoroutine()

	// Update GC pause time
	pm.metrics.GCPauseTime = time.Duration(m.PauseTotalNs)

	// Update CPU usage (simplified)
	pm.metrics.CPUUsage = pm.calculateCPUUsage()
}

// updateThroughputMetrics updates throughput-related metrics
func (pm *PerformanceMonitor) updateThroughputMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Calculate requests per second
	if pm.metrics.Uptime > 0 {
		pm.metrics.RequestsPerSecond = float64(pm.metrics.TotalRequests) / pm.metrics.Uptime.Seconds()
	}

	// Update peak RPS
	if pm.metrics.RequestsPerSecond > pm.metrics.PeakRPS {
		pm.metrics.PeakRPS = pm.metrics.RequestsPerSecond
	}

	// Update current RPS (simplified)
	pm.metrics.CurrentRPS = pm.metrics.RequestsPerSecond
}

// updateResponseTimeBuckets updates response time distribution buckets
func (pm *PerformanceMonitor) updateResponseTimeBuckets(duration time.Duration) {
	for _, bucket := range pm.config.ResponseTimeBuckets {
		if duration <= bucket {
			bucketKey := fmt.Sprintf("<=%v", bucket)
			pm.metrics.ResponseTimeBuckets[bucketKey]++
			break
		}
	}
}

// updateRequestSizeBuckets updates request size distribution buckets
func (pm *PerformanceMonitor) updateRequestSizeBuckets(size int64) {
	for _, bucket := range pm.config.RequestSizeBuckets {
		if size <= bucket {
			bucketKey := fmt.Sprintf("<=%d", bucket)
			pm.metrics.RequestSizeBuckets[bucketKey]++
			break
		}
	}
}

// calculateCPUUsage calculates CPU usage (simplified implementation)
func (pm *PerformanceMonitor) calculateCPUUsage() float64 {
	// This is a simplified CPU usage calculation
	// In a real implementation, you would use more sophisticated methods
	return float64(runtime.NumGoroutine()) * 0.1 // Simplified calculation
}

// generatePerformanceInsights generates performance insights
func (pm *PerformanceMonitor) generatePerformanceInsights() []PerformanceInsight {
	return pm.GetPerformanceInsights()
}

// generateRecommendations generates performance recommendations
func (pm *PerformanceMonitor) generateRecommendations() []string {
	var recommendations []string

	metrics := pm.metrics

	// Response time recommendations
	if metrics.AverageResponseTime > 500*time.Millisecond {
		recommendations = append(recommendations, "Consider implementing query caching to reduce response times")
		recommendations = append(recommendations, "Optimize complex queries that are taking too long")
	}

	// Error rate recommendations
	if metrics.ErrorRate > 0.05 {
		recommendations = append(recommendations, "Investigate and fix the root cause of errors")
		recommendations = append(recommendations, "Implement better error handling and validation")
	}

	// Memory recommendations
	if metrics.MemoryUsage > 1024*1024*1024 {
		recommendations = append(recommendations, "Monitor memory usage and consider optimizing queries")
		recommendations = append(recommendations, "Implement memory-efficient data structures")
	}

	// Throughput recommendations
	if metrics.RequestsPerSecond > 1000 {
		recommendations = append(recommendations, "Consider implementing rate limiting")
		recommendations = append(recommendations, "Scale resources to handle high throughput")
	}

	return recommendations
}

// Reset resets all performance metrics
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.metrics = &PerformanceMetrics{
		ErrorTypes:          make(map[string]int64),
		ResponseTimeBuckets: make(map[string]int64),
		RequestSizeBuckets:  make(map[string]int64),
	}
	pm.startTime = time.Now()

	pm.logger.Info("Performance metrics reset")
}

// PerformanceInsight represents an insight about performance
type PerformanceInsight struct {
	Type           InsightType            `json:"type"`
	Severity       Severity               `json:"severity"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
