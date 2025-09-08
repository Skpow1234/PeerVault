package analytics

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// QueryAnalytics tracks and analyzes GraphQL query performance
type QueryAnalytics struct {
	queries   map[string]*QueryMetrics
	mu        sync.RWMutex
	logger    *slog.Logger
	startTime time.Time
}

// QueryMetrics holds metrics for a specific query
type QueryMetrics struct {
	Query         string                 `json:"query"`
	Hash          string                 `json:"hash"`
	Count         int64                  `json:"count"`
	TotalTime     time.Duration          `json:"totalTime"`
	AverageTime   time.Duration          `json:"averageTime"`
	MinTime       time.Duration          `json:"minTime"`
	MaxTime       time.Duration          `json:"maxTime"`
	ErrorCount    int64                  `json:"errorCount"`
	LastExecuted  time.Time              `json:"lastExecuted"`
	FirstExecuted time.Time              `json:"firstExecuted"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Complexity    int                    `json:"complexity"`
	Depth         int                    `json:"depth"`
	FieldCount    int                    `json:"fieldCount"`
	Source        string                 `json:"source,omitempty"`
	UserAgent     string                 `json:"userAgent,omitempty"`
	IPAddress     string                 `json:"ipAddress,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// QueryExecution represents a single query execution
type QueryExecution struct {
	Query      string                 `json:"query"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	StartTime  time.Time              `json:"startTime"`
	EndTime    time.Time              `json:"endTime"`
	Duration   time.Duration          `json:"duration"`
	Error      error                  `json:"error,omitempty"`
	Complexity int                    `json:"complexity"`
	Depth      int                    `json:"depth"`
	FieldCount int                    `json:"fieldCount"`
	Source     string                 `json:"source,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	IPAddress  string                 `json:"ipAddress,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// QueryAnalyticsConfig holds configuration for query analytics
type QueryAnalyticsConfig struct {
	EnableTracking     bool          `json:"enableTracking"`
	MaxQueries         int           `json:"maxQueries"`
	RetentionPeriod    time.Duration `json:"retentionPeriod"`
	EnableComplexity   bool          `json:"enableComplexity"`
	EnableDepth        bool          `json:"enableDepth"`
	EnableFieldCount   bool          `json:"enableFieldCount"`
	EnableUserTracking bool          `json:"enableUserTracking"`
	EnableIPTracking   bool          `json:"enableIPTracking"`
}

// DefaultQueryAnalyticsConfig returns the default configuration
func DefaultQueryAnalyticsConfig() *QueryAnalyticsConfig {
	return &QueryAnalyticsConfig{
		EnableTracking:     true,
		MaxQueries:         10000,
		RetentionPeriod:    24 * time.Hour,
		EnableComplexity:   true,
		EnableDepth:        true,
		EnableFieldCount:   true,
		EnableUserTracking: false,
		EnableIPTracking:   false,
	}
}

// NewQueryAnalytics creates a new query analytics instance
func NewQueryAnalytics(logger *slog.Logger) *QueryAnalytics {
	return &QueryAnalytics{
		queries:   make(map[string]*QueryMetrics),
		logger:    logger,
		startTime: time.Now(),
	}
}

// RecordQuery records a query execution
func (qa *QueryAnalytics) RecordQuery(execution *QueryExecution) {
	qa.mu.Lock()
	defer qa.mu.Unlock()

	// Generate query hash
	hash := qa.generateQueryHash(execution.Query, execution.Variables)

	// Get or create metrics
	metrics, exists := qa.queries[hash]
	if !exists {
		metrics = &QueryMetrics{
			Query:         execution.Query,
			Hash:          hash,
			FirstExecuted: execution.StartTime,
			Variables:     execution.Variables,
			Complexity:    execution.Complexity,
			Depth:         execution.Depth,
			FieldCount:    execution.FieldCount,
			Source:        execution.Source,
			UserAgent:     execution.UserAgent,
			IPAddress:     execution.IPAddress,
			Metadata:      make(map[string]interface{}),
		}
		qa.queries[hash] = metrics
	}

	// Update metrics
	metrics.Count++
	metrics.TotalTime += execution.Duration
	metrics.AverageTime = metrics.TotalTime / time.Duration(metrics.Count)
	metrics.LastExecuted = execution.StartTime

	// Update min/max times
	if metrics.MinTime == 0 || execution.Duration < metrics.MinTime {
		metrics.MinTime = execution.Duration
	}
	if execution.Duration > metrics.MaxTime {
		metrics.MaxTime = execution.Duration
	}

	// Count errors
	if execution.Error != nil {
		metrics.ErrorCount++
	}

	// Update metadata
	if execution.Metadata != nil {
		for k, v := range execution.Metadata {
			metrics.Metadata[k] = v
		}
	}

	qa.logger.Debug("Recorded query execution",
		"hash", hash,
		"duration", execution.Duration,
		"complexity", execution.Complexity)
}

// GetQueryMetrics returns metrics for a specific query
func (qa *QueryAnalytics) GetQueryMetrics(hash string) (*QueryMetrics, bool) {
	qa.mu.RLock()
	defer qa.mu.RUnlock()
	metrics, exists := qa.queries[hash]
	return metrics, exists
}

// GetTopQueries returns the most frequently executed queries
func (qa *QueryAnalytics) GetTopQueries(limit int) []*QueryMetrics {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	var queries []*QueryMetrics
	for _, metrics := range qa.queries {
		queries = append(queries, metrics)
	}

	// Sort by count (descending)
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Count > queries[j].Count
	})

	if limit > 0 && limit < len(queries) {
		queries = queries[:limit]
	}

	return queries
}

// GetSlowestQueries returns the slowest queries
func (qa *QueryAnalytics) GetSlowestQueries(limit int) []*QueryMetrics {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	var queries []*QueryMetrics
	for _, metrics := range qa.queries {
		queries = append(queries, metrics)
	}

	// Sort by average time (descending)
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].AverageTime > queries[j].AverageTime
	})

	if limit > 0 && limit < len(queries) {
		queries = queries[:limit]
	}

	return queries
}

// GetMostComplexQueries returns the most complex queries
func (qa *QueryAnalytics) GetMostComplexQueries(limit int) []*QueryMetrics {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	var queries []*QueryMetrics
	for _, metrics := range qa.queries {
		queries = append(queries, metrics)
	}

	// Sort by complexity (descending)
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Complexity > queries[j].Complexity
	})

	if limit > 0 && limit < len(queries) {
		queries = queries[:limit]
	}

	return queries
}

// GetQueriesWithErrors returns queries that have errors
func (qa *QueryAnalytics) GetQueriesWithErrors(limit int) []*QueryMetrics {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	var queries []*QueryMetrics
	for _, metrics := range qa.queries {
		if metrics.ErrorCount > 0 {
			queries = append(queries, metrics)
		}
	}

	// Sort by error count (descending)
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].ErrorCount > queries[j].ErrorCount
	})

	if limit > 0 && limit < len(queries) {
		queries = queries[:limit]
	}

	return queries
}

// GetQueryStats returns overall query statistics
func (qa *QueryAnalytics) GetQueryStats() map[string]interface{} {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	stats := map[string]interface{}{
		"totalQueries":    len(qa.queries),
		"totalExecutions": int64(0),
		"totalErrors":     int64(0),
		"totalTime":       time.Duration(0),
		"averageTime":     time.Duration(0),
		"uptime":          time.Since(qa.startTime),
	}

	var totalExecutions, totalErrors int64
	var totalTime time.Duration

	for _, metrics := range qa.queries {
		totalExecutions += metrics.Count
		totalErrors += metrics.ErrorCount
		totalTime += metrics.TotalTime
	}

	stats["totalExecutions"] = totalExecutions
	stats["totalErrors"] = totalErrors
	stats["totalTime"] = totalTime

	if totalExecutions > 0 {
		stats["averageTime"] = totalTime / time.Duration(totalExecutions)
		stats["errorRate"] = float64(totalErrors) / float64(totalExecutions)
	}

	return stats
}

// GetQueryTrends returns query trends over time
func (qa *QueryAnalytics) GetQueryTrends(period time.Duration) map[string]interface{} {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	now := time.Now()
	startTime := now.Add(-period)

	trends := map[string]interface{}{
		"period":    period,
		"startTime": startTime,
		"endTime":   now,
		"queries":   make(map[string]interface{}),
	}

	// Group queries by time buckets
	timeBuckets := make(map[string]int64)
	executionBuckets := make(map[string]int64)
	errorBuckets := make(map[string]int64)

	for _, metrics := range qa.queries {
		// Check if query was executed in the period
		if metrics.LastExecuted.After(startTime) {
			bucket := metrics.LastExecuted.Truncate(time.Hour).Format("2006-01-02T15:00:00Z")
			timeBuckets[bucket]++
			executionBuckets[bucket] += metrics.Count
			errorBuckets[bucket] += metrics.ErrorCount
		}
	}

	trends["timeBuckets"] = timeBuckets
	trends["executionBuckets"] = executionBuckets
	trends["errorBuckets"] = errorBuckets

	return trends
}

// GetQueryInsights returns insights about query performance
func (qa *QueryAnalytics) GetQueryInsights() []QueryInsight {
	qa.mu.RLock()
	defer qa.mu.RUnlock()

	var insights []QueryInsight

	// Analyze query performance
	stats := qa.GetQueryStats()
	errorRate := stats["errorRate"].(float64)

	// High error rate insight
	if errorRate > 0.1 { // 10% error rate
		insights = append(insights, QueryInsight{
			Type:           InsightTypeErrorRate,
			Severity:       SeverityWarning,
			Title:          "High Error Rate",
			Description:    fmt.Sprintf("Query error rate is %.2f%%, which is above the recommended 10%%", errorRate*100),
			Recommendation: "Review error logs and optimize problematic queries",
			Metadata: map[string]interface{}{
				"errorRate": errorRate,
				"threshold": 0.1,
			},
		})
	}

	// Slow queries insight
	slowQueries := qa.GetSlowestQueries(5)
	if len(slowQueries) > 0 && slowQueries[0].AverageTime > 1*time.Second {
		insights = append(insights, QueryInsight{
			Type:           InsightTypePerformance,
			Severity:       SeverityWarning,
			Title:          "Slow Queries Detected",
			Description:    "Some queries are taking longer than 1 second on average",
			Recommendation: "Optimize slow queries or implement caching",
			Metadata: map[string]interface{}{
				"slowestQuery": slowQueries[0].Query,
				"averageTime":  slowQueries[0].AverageTime,
			},
		})
	}

	// Complex queries insight
	complexQueries := qa.GetMostComplexQueries(5)
	if len(complexQueries) > 0 && complexQueries[0].Complexity > 1000 {
		insights = append(insights, QueryInsight{
			Type:           InsightTypeComplexity,
			Severity:       SeverityInfo,
			Title:          "Complex Queries Detected",
			Description:    "Some queries have high complexity scores (>1000)",
			Recommendation: "Consider breaking down complex queries or implementing query depth limits",
			Metadata: map[string]interface{}{
				"mostComplexQuery": complexQueries[0].Query,
				"complexity":       complexQueries[0].Complexity,
			},
		})
	}

	// High frequency queries insight
	topQueries := qa.GetTopQueries(5)
	if len(topQueries) > 0 && topQueries[0].Count > 1000 {
		insights = append(insights, QueryInsight{
			Type:           InsightTypeFrequency,
			Severity:       SeverityInfo,
			Title:          "High Frequency Queries",
			Description:    "Some queries are executed very frequently",
			Recommendation: "Consider implementing caching for frequently executed queries",
			Metadata: map[string]interface{}{
				"mostFrequentQuery": topQueries[0].Query,
				"executionCount":    topQueries[0].Count,
			},
		})
	}

	return insights
}

// generateQueryHash generates a hash for a query and its variables
func (qa *QueryAnalytics) generateQueryHash(query string, variables map[string]interface{}) string {
	// Create a deterministic hash from query and variables
	hashData := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	hashBytes, _ := json.Marshal(hashData)
	return fmt.Sprintf("%x", hashBytes)
}

// Cleanup removes old query data
func (qa *QueryAnalytics) Cleanup(retentionPeriod time.Duration) {
	qa.mu.Lock()
	defer qa.mu.Unlock()

	cutoff := time.Now().Add(-retentionPeriod)
	var removed int

	for hash, metrics := range qa.queries {
		if metrics.LastExecuted.Before(cutoff) {
			delete(qa.queries, hash)
			removed++
		}
	}

	if removed > 0 {
		qa.logger.Info("Cleaned up old query data", "removed", removed, "retentionPeriod", retentionPeriod)
	}
}

// Reset clears all query analytics data
func (qa *QueryAnalytics) Reset() {
	qa.mu.Lock()
	defer qa.mu.Unlock()

	qa.queries = make(map[string]*QueryMetrics)
	qa.startTime = time.Now()

	qa.logger.Info("Query analytics data reset")
}

// QueryInsight represents an insight about query performance
type QueryInsight struct {
	Type           InsightType            `json:"type"`
	Severity       Severity               `json:"severity"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Recommendation string                 `json:"recommendation"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
