package cache

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CacheAnalytics provides analytics and insights for the GraphQL cache
type CacheAnalytics struct {
	queryCache *QueryCache
	metrics    *AnalyticsMetrics
	mu         sync.RWMutex
	logger     *slog.Logger
}

// AnalyticsMetrics holds comprehensive cache analytics
type AnalyticsMetrics struct {
	// Basic metrics
	TotalRequests      int64 `json:"totalRequests"`
	CacheHits          int64 `json:"cacheHits"`
	CacheMisses        int64 `json:"cacheMisses"`
	CacheSets          int64 `json:"cacheSets"`
	CacheDeletes       int64 `json:"cacheDeletes"`
	CacheInvalidations int64 `json:"cacheInvalidations"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"averageResponseTime"`
	MinResponseTime     time.Duration `json:"minResponseTime"`
	MaxResponseTime     time.Duration `json:"maxResponseTime"`
	TotalResponseTime   time.Duration `json:"totalResponseTime"`

	// Hit rate metrics
	HitRate       float64            `json:"hitRate"`
	HitRateByHour map[int]float64    `json:"hitRateByHour"`
	HitRateByDay  map[string]float64 `json:"hitRateByDay"`

	// Query patterns
	TopQueries    []QueryStats   `json:"topQueries"`
	SlowQueries   []QueryStats   `json:"slowQueries"`
	QueryPatterns map[string]int `json:"queryPatterns"`

	// Cache efficiency
	CacheSize        int     `json:"cacheSize"`
	CacheUtilization float64 `json:"cacheUtilization"`
	MemoryUsage      int64   `json:"memoryUsage"`

	// Time-based metrics
	LastReset  time.Time     `json:"lastReset"`
	LastUpdate time.Time     `json:"lastUpdate"`
	Uptime     time.Duration `json:"uptime"`
}

// QueryStats holds statistics for a specific query
type QueryStats struct {
	Query        string                 `json:"query"`
	Hash         string                 `json:"hash"`
	Count        int64                  `json:"count"`
	Hits         int64                  `json:"hits"`
	Misses       int64                  `json:"misses"`
	AverageTime  time.Duration          `json:"averageTime"`
	TotalTime    time.Duration          `json:"totalTime"`
	LastExecuted time.Time              `json:"lastExecuted"`
	HitRate      float64                `json:"hitRate"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
}

// CacheInsight represents an insight about cache performance
type CacheInsight struct {
	Type           InsightType            `json:"type"`
	Severity       Severity               `json:"severity"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// InsightType defines the type of cache insight
type InsightType string

const (
	InsightTypePerformance InsightType = "performance"
	InsightTypeHitRate     InsightType = "hit_rate"
	InsightTypeMemory      InsightType = "memory"
	InsightTypeQuery       InsightType = "query"
	InsightTypeTTL         InsightType = "ttl"
)

// Severity defines the severity of an insight
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// NewCacheAnalytics creates a new cache analytics instance
func NewCacheAnalytics(queryCache *QueryCache, logger *slog.Logger) *CacheAnalytics {
	return &CacheAnalytics{
		queryCache: queryCache,
		metrics: &AnalyticsMetrics{
			HitRateByHour: make(map[int]float64),
			HitRateByDay:  make(map[string]float64),
			QueryPatterns: make(map[string]int),
			LastReset:     time.Now(),
			LastUpdate:    time.Now(),
		},
		logger: logger,
	}
}

// RecordRequest records a cache request
func (ca *CacheAnalytics) RecordRequest(query string, variables map[string]interface{}, hit bool, responseTime time.Duration) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	ca.metrics.TotalRequests++
	ca.metrics.TotalResponseTime += responseTime
	ca.metrics.LastUpdate = time.Now()

	if hit {
		ca.metrics.CacheHits++
	} else {
		ca.metrics.CacheMisses++
	}

	// Update response time metrics
	if ca.metrics.MinResponseTime == 0 || responseTime < ca.metrics.MinResponseTime {
		ca.metrics.MinResponseTime = responseTime
	}
	if responseTime > ca.metrics.MaxResponseTime {
		ca.metrics.MaxResponseTime = responseTime
	}

	// Update hit rate
	if ca.metrics.TotalRequests > 0 {
		ca.metrics.HitRate = float64(ca.metrics.CacheHits) / float64(ca.metrics.TotalRequests)
	}

	// Update hourly hit rate
	hour := time.Now().Hour()
	if ca.metrics.HitRateByHour[hour] == 0 {
		ca.metrics.HitRateByHour[hour] = ca.metrics.HitRate
	} else {
		// Simple moving average
		ca.metrics.HitRateByHour[hour] = (ca.metrics.HitRateByHour[hour] + ca.metrics.HitRate) / 2
	}

	// Update daily hit rate
	day := time.Now().Format("2006-01-02")
	if ca.metrics.HitRateByDay[day] == 0 {
		ca.metrics.HitRateByDay[day] = ca.metrics.HitRate
	} else {
		ca.metrics.HitRateByDay[day] = (ca.metrics.HitRateByDay[day] + ca.metrics.HitRate) / 2
	}

	// Update query patterns
	queryHash := ca.hashQuery(query)
	ca.metrics.QueryPatterns[queryHash]++

	// Update average response time
	ca.metrics.AverageResponseTime = ca.metrics.TotalResponseTime / time.Duration(ca.metrics.TotalRequests)
}

// RecordCacheSet records a cache set operation
func (ca *CacheAnalytics) RecordCacheSet() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.metrics.CacheSets++
	ca.metrics.LastUpdate = time.Now()
}

// RecordCacheDelete records a cache delete operation
func (ca *CacheAnalytics) RecordCacheDelete() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.metrics.CacheDeletes++
	ca.metrics.LastUpdate = time.Now()
}

// RecordInvalidation records a cache invalidation
func (ca *CacheAnalytics) RecordInvalidation() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.metrics.CacheInvalidations++
	ca.metrics.LastUpdate = time.Now()
}

// GetMetrics returns current analytics metrics
func (ca *CacheAnalytics) GetMetrics() *AnalyticsMetrics {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	// Update cache size and utilization
	cacheStats := ca.queryCache.GetCacheStats()
	ca.metrics.CacheSize = cacheStats["size"].(int)
	ca.metrics.CacheUtilization = ca.calculateCacheUtilization()

	// Update uptime
	ca.metrics.Uptime = time.Since(ca.metrics.LastReset)

	// Create a copy to avoid race conditions
	metrics := *ca.metrics
	return &metrics
}

// GetTopQueries returns the most frequently executed queries
func (ca *CacheAnalytics) GetTopQueries(limit int) []QueryStats {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	// This is a simplified implementation
	// In a real implementation, you would track individual query statistics
	return ca.metrics.TopQueries
}

// GetSlowQueries returns the slowest queries
func (ca *CacheAnalytics) GetSlowQueries(limit int) []QueryStats {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	// This is a simplified implementation
	// In a real implementation, you would track individual query performance
	return ca.metrics.SlowQueries
}

// GenerateInsights generates insights about cache performance
func (ca *CacheAnalytics) GenerateInsights() []CacheInsight {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	var insights []CacheInsight

	// Hit rate insights
	if ca.metrics.HitRate < 0.5 {
		insights = append(insights, CacheInsight{
			Type:           InsightTypeHitRate,
			Severity:       SeverityWarning,
			Title:          "Low Cache Hit Rate",
			Description:    fmt.Sprintf("Cache hit rate is %.2f%%, which is below the recommended 50%%", ca.metrics.HitRate*100),
			Recommendation: "Consider increasing TTL values or adding more warmup queries",
			Timestamp:      time.Now(),
			Metadata: map[string]interface{}{
				"hitRate":   ca.metrics.HitRate,
				"threshold": 0.5,
			},
		})
	}

	// Performance insights
	if ca.metrics.AverageResponseTime > 100*time.Millisecond {
		insights = append(insights, CacheInsight{
			Type:           InsightTypePerformance,
			Severity:       SeverityWarning,
			Title:          "High Average Response Time",
			Description:    fmt.Sprintf("Average response time is %v, which is above the recommended 100ms", ca.metrics.AverageResponseTime),
			Recommendation: "Consider optimizing queries or increasing cache TTL",
			Timestamp:      time.Now(),
			Metadata: map[string]interface{}{
				"averageResponseTime": ca.metrics.AverageResponseTime,
				"threshold":           100 * time.Millisecond,
			},
		})
	}

	// Memory insights
	if ca.metrics.CacheUtilization > 0.9 {
		insights = append(insights, CacheInsight{
			Type:           InsightTypeMemory,
			Severity:       SeverityCritical,
			Title:          "High Cache Utilization",
			Description:    fmt.Sprintf("Cache utilization is %.2f%%, which is above the recommended 90%%", ca.metrics.CacheUtilization*100),
			Recommendation: "Consider increasing cache size or implementing cache eviction policies",
			Timestamp:      time.Now(),
			Metadata: map[string]interface{}{
				"cacheUtilization": ca.metrics.CacheUtilization,
				"threshold":        0.9,
			},
		})
	}

	// Query pattern insights
	if len(ca.metrics.QueryPatterns) > 100 {
		insights = append(insights, CacheInsight{
			Type:           InsightTypeQuery,
			Severity:       SeverityInfo,
			Title:          "High Query Diversity",
			Description:    fmt.Sprintf("Cache contains %d unique query patterns", len(ca.metrics.QueryPatterns)),
			Recommendation: "Consider query normalization or pattern-based caching",
			Timestamp:      time.Now(),
			Metadata: map[string]interface{}{
				"uniqueQueries": len(ca.metrics.QueryPatterns),
			},
		})
	}

	return insights
}

// GetPerformanceReport generates a comprehensive performance report
func (ca *CacheAnalytics) GetPerformanceReport() map[string]interface{} {
	metrics := ca.GetMetrics()
	insights := ca.GenerateInsights()

	// Calculate additional metrics
	efficiency := ca.calculateCacheEfficiency()
	trends := ca.calculateTrends()

	return map[string]interface{}{
		"metrics":     metrics,
		"insights":    insights,
		"efficiency":  efficiency,
		"trends":      trends,
		"generatedAt": time.Now(),
	}
}

// calculateCacheUtilization calculates cache utilization percentage
func (ca *CacheAnalytics) calculateCacheUtilization() float64 {
	cacheStats := ca.queryCache.GetCacheStats()
	maxSize := cacheStats["config"].(map[string]interface{})["maxCacheSize"].(int)
	currentSize := cacheStats["size"].(int)

	if maxSize == 0 {
		return 0
	}

	return float64(currentSize) / float64(maxSize)
}

// calculateCacheEfficiency calculates cache efficiency metrics
func (ca *CacheAnalytics) calculateCacheEfficiency() map[string]interface{} {
	metrics := ca.metrics

	// Calculate efficiency score (0-100)
	efficiencyScore := float64(0)
	if metrics.TotalRequests > 0 {
		hitRateScore := metrics.HitRate * 40 // 40% weight for hit rate
		responseTimeScore := float64(0)
		if metrics.AverageResponseTime > 0 {
			// Normalize response time (lower is better)
			responseTimeScore = (100 - float64(metrics.AverageResponseTime.Milliseconds())/10) * 0.3
			if responseTimeScore < 0 {
				responseTimeScore = 0
			}
		}
		utilizationScore := (1 - ca.calculateCacheUtilization()) * 30 // 30% weight for utilization
		efficiencyScore = hitRateScore + responseTimeScore + utilizationScore
	}

	return map[string]interface{}{
		"score":             efficiencyScore,
		"hitRateScore":      metrics.HitRate * 40,
		"responseTimeScore": (100 - float64(metrics.AverageResponseTime.Milliseconds())/10) * 0.3,
		"utilizationScore":  (1 - ca.calculateCacheUtilization()) * 30,
	}
}

// calculateTrends calculates performance trends
func (ca *CacheAnalytics) calculateTrends() map[string]interface{} {
	// This is a simplified implementation
	// In a real implementation, you would analyze historical data

	return map[string]interface{}{
		"hitRateTrend":      "stable",
		"responseTimeTrend": "stable",
		"utilizationTrend":  "stable",
		"queryVolumeTrend":  "stable",
	}
}

// ResetMetrics resets all analytics metrics
func (ca *CacheAnalytics) ResetMetrics() {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	ca.metrics = &AnalyticsMetrics{
		HitRateByHour: make(map[int]float64),
		HitRateByDay:  make(map[string]float64),
		QueryPatterns: make(map[string]int),
		LastReset:     time.Now(),
		LastUpdate:    time.Now(),
	}

	ca.logger.Info("Analytics metrics reset")
}

// ExportMetrics exports metrics in various formats
func (ca *CacheAnalytics) ExportMetrics(format string) ([]byte, error) {
	metrics := ca.GetMetrics()

	switch format {
	case "json":
		return json.MarshalIndent(metrics, "", "  ")
	case "csv":
		return ca.exportCSV(metrics), nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportCSV exports metrics in CSV format
func (ca *CacheAnalytics) exportCSV(metrics *AnalyticsMetrics) []byte {
	csv := fmt.Sprintf(`metric,value
totalRequests,%d
cacheHits,%d
cacheMisses,%d
cacheSets,%d
cacheDeletes,%d
cacheInvalidations,%d
hitRate,%.4f
averageResponseTime,%v
minResponseTime,%v
maxResponseTime,%v
cacheSize,%d
cacheUtilization,%.4f
uptime,%v
`,
		metrics.TotalRequests,
		metrics.CacheHits,
		metrics.CacheMisses,
		metrics.CacheSets,
		metrics.CacheDeletes,
		metrics.CacheInvalidations,
		metrics.HitRate,
		metrics.AverageResponseTime,
		metrics.MinResponseTime,
		metrics.MaxResponseTime,
		metrics.CacheSize,
		metrics.CacheUtilization,
		metrics.Uptime,
	)

	return []byte(csv)
}

// hashQuery generates a hash for the query
func (ca *CacheAnalytics) hashQuery(query string) string {
	// This is a simplified implementation
	// In a real implementation, you would use a proper hash function
	return fmt.Sprintf("query_%d", len(query))
}

// GetHourlyStats returns hourly statistics
func (ca *CacheAnalytics) GetHourlyStats() map[int]map[string]interface{} {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	stats := make(map[int]map[string]interface{})
	for hour, hitRate := range ca.metrics.HitRateByHour {
		stats[hour] = map[string]interface{}{
			"hitRate": hitRate,
			"hour":    hour,
		}
	}

	return stats
}

// GetDailyStats returns daily statistics
func (ca *CacheAnalytics) GetDailyStats() map[string]map[string]interface{} {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for day, hitRate := range ca.metrics.HitRateByDay {
		stats[day] = map[string]interface{}{
			"hitRate": hitRate,
			"day":     day,
		}
	}

	return stats
}
