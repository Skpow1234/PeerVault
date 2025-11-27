package monitoring

import (
	"context"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"
)

// Collector collects and aggregates performance metrics
type Collector struct {
	metrics      []PerformanceMetric
	mu           sync.RWMutex
	config       *MonitoringConfig
	windowStart  time.Time
	activeReqs   int64
	activeMu     sync.RWMutex
}

// NewCollector creates a new performance metrics collector
func NewCollector(config *MonitoringConfig) *Collector {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Collector{
		metrics:     make([]PerformanceMetric, 0, 10000),
		config:      config,
		windowStart: time.Now(),
	}
}

// RecordMetric records a performance metric
func (c *Collector) RecordMetric(metric PerformanceMetric) {
	if !c.config.Enabled {
		return
	}
	
	// Sample based on sample rate
	if c.config.SampleRate < 1.0 && !c.shouldSample() {
		return
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Add metric
	c.metrics = append(c.metrics, metric)
	
	// Keep only metrics within the window
	c.pruneOldMetrics()
}

// IncrementActiveRequests increments the active request counter
func (c *Collector) IncrementActiveRequests() {
	c.activeMu.Lock()
	defer c.activeMu.Unlock()
	c.activeReqs++
}

// DecrementActiveRequests decrements the active request counter
func (c *Collector) DecrementActiveRequests() {
	c.activeMu.Lock()
	defer c.activeMu.Unlock()
	if c.activeReqs > 0 {
		c.activeReqs--
	}
}

// GetActiveRequests returns the current number of active requests
func (c *Collector) GetActiveRequests() int64 {
	c.activeMu.RLock()
	defer c.activeMu.RUnlock()
	return c.activeReqs
}

// GetResponseTimeStats calculates response time statistics
func (c *Collector) GetResponseTimeStats(endpoint, method string) ResponseTimeStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	var times []time.Duration
	for _, m := range c.metrics {
		if (endpoint == "" || m.Endpoint == endpoint) &&
			(method == "" || m.Method == method) {
			times = append(times, m.ResponseTime)
		}
	}
	
	if len(times) == 0 {
		return ResponseTimeStats{
			Endpoint: endpoint,
			Method:   method,
		}
	}
	
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	
	stats := ResponseTimeStats{
		Endpoint:    endpoint,
		Method:      method,
		Count:       int64(len(times)),
		Min:         times[0],
		Max:         times[len(times)-1],
		Mean:        calculateMean(times),
		Median:      times[len(times)/2],
		P95:         times[int(float64(len(times))*0.95)],
		P99:         times[int(float64(len(times))*0.99)],
		LastUpdated: time.Now(),
	}
	
	stats.StandardDeviation = calculateStdDev(times, stats.Mean)
	
	return stats
}

// GetThroughputStats calculates throughput statistics
func (c *Collector) GetThroughputStats(endpoint, method string) ThroughputStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	now := time.Now()
	window := c.config.WindowSize
	
	var count int64
	var totalBytes int64
	var peakRPS float64
	var peakTime time.Time
	
	// Count requests in time buckets for peak detection
	buckets := make(map[int64]int)
	
	for _, m := range c.metrics {
		if (endpoint == "" || m.Endpoint == endpoint) &&
			(method == "" || m.Method == method) {
			count++
			totalBytes += m.ResponseSize
			
			// Track peak RPS (per second buckets)
			bucket := m.Timestamp.Unix()
			buckets[bucket]++
			if float64(buckets[bucket]) > peakRPS {
				peakRPS = float64(buckets[bucket])
				peakTime = time.Unix(bucket, 0)
			}
		}
	}
	
	duration := window.Seconds()
	if duration == 0 {
		duration = 1
	}
	
	return ThroughputStats{
		Endpoint:          endpoint,
		Method:            method,
		RequestsPerSecond: float64(count) / duration,
		RequestsPerMinute: float64(count) / duration * 60,
		RequestsPerHour:   float64(count) / duration * 3600,
		PeakRPS:           peakRPS,
		PeakTime:          peakTime,
		BytesPerSecond:    float64(totalBytes) / duration,
		LastUpdated:       now,
	}
}

// GetEndpointPerformance gets performance stats for a specific endpoint
func (c *Collector) GetEndpointPerformance(endpoint, method string) EndpointPerformance {
	responseTime := c.GetResponseTimeStats(endpoint, method)
	throughput := c.GetThroughputStats(endpoint, method)
	
	// Calculate error rate
	c.mu.RLock()
	var total, errors int64
	for _, m := range c.metrics {
		if (endpoint == "" || m.Endpoint == endpoint) &&
			(method == "" || m.Method == method) {
			total++
			if m.StatusCode >= 400 {
				errors++
			}
		}
	}
	c.mu.RUnlock()
	
	errorRate := 0.0
	successRate := 100.0
	if total > 0 {
		errorRate = float64(errors) / float64(total) * 100
		successRate = 100 - errorRate
	}
	
	// Calculate saturation (simplified - based on response time vs baseline)
	saturation := 0.0
	if responseTime.Mean > 0 {
		baseline := time.Duration(100 * time.Millisecond)
		saturation = math.Min(100, float64(responseTime.Mean)/float64(baseline)*100)
	}
	
	return EndpointPerformance{
		Endpoint:     endpoint,
		Method:       method,
		ResponseTime: responseTime,
		Throughput:   throughput,
		ErrorRate:    errorRate,
		SuccessRate:  successRate,
		Saturation:   saturation,
	}
}

// GetPerformanceSnapshot creates a point-in-time performance snapshot
func (c *Collector) GetPerformanceSnapshot(ctx context.Context) PerformanceSnapshot {
	// Get overall stats
	overallRT := c.GetResponseTimeStats("", "")
	overallTP := c.GetThroughputStats("", "")
	
	// Get per-endpoint stats
	endpoints := c.getUniqueEndpoints()
	endpointStats := make([]EndpointPerformance, 0, len(endpoints))
	for _, ep := range endpoints {
		stats := c.GetEndpointPerformance(ep, "")
		endpointStats = append(endpointStats, stats)
	}
	
	// Sort by response time (slowest first)
	sort.Slice(endpointStats, func(i, j int) bool {
		return endpointStats[i].ResponseTime.Mean > endpointStats[j].ResponseTime.Mean
	})
	
	// Get system load
	systemLoad := c.getSystemLoad()
	
	// Calculate health score
	healthScore := c.calculateHealthScore(overallRT, overallTP, systemLoad)
	
	return PerformanceSnapshot{
		Timestamp:           time.Now(),
		OverallResponseTime: overallRT,
		OverallThroughput:   overallTP,
		EndpointStats:       endpointStats,
		SystemLoad:          systemLoad,
		HealthScore:         healthScore,
	}
}

// GetRealTimeMetrics returns current real-time metrics
func (c *Collector) GetRealTimeMetrics() RealTimeMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	now := time.Now()
	oneMinuteAgo := now.Add(-1 * time.Minute)
	
	var lastMinuteCount int64
	var lastMinuteTotal time.Duration
	var errorCount, successCount int64
	
	for _, m := range c.metrics {
		if m.Timestamp.After(oneMinuteAgo) {
			lastMinuteCount++
			lastMinuteTotal += m.ResponseTime
			if m.StatusCode >= 400 {
				errorCount++
			} else {
				successCount++
			}
		}
	}
	
	avgLatency := 0.0
	if lastMinuteCount > 0 {
		avgLatency = float64(lastMinuteTotal.Milliseconds()) / float64(lastMinuteCount)
	}
	
	return RealTimeMetrics{
		Timestamp:          now,
		CurrentRPS:         float64(lastMinuteCount) / 60.0,
		AverageLatencyMs:   avgLatency,
		ActiveRequests:     int(c.GetActiveRequests()),
		ErrorCount:         errorCount,
		SuccessCount:       successCount,
		LastMinuteRequests: lastMinuteCount,
	}
}

// GetPerformanceTrend analyzes performance trends
func (c *Collector) GetPerformanceTrend(endpoint string, bucketSize time.Duration) PerformanceTrend {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Group metrics into time buckets
	buckets := make(map[int64][]PerformanceMetric)
	for _, m := range c.metrics {
		if endpoint == "" || m.Endpoint == endpoint {
			bucket := m.Timestamp.Unix() / int64(bucketSize.Seconds())
			buckets[bucket] = append(buckets[bucket], m)
		}
	}
	
	// Calculate stats for each bucket
	var dataPoints []TrendDataPoint
	for bucket, metrics := range buckets {
		if len(metrics) == 0 {
			continue
		}
		
		timestamp := time.Unix(bucket*int64(bucketSize.Seconds()), 0)
		
		var totalRT time.Duration
		var errorCount int64
		for _, m := range metrics {
			totalRT += m.ResponseTime
			if m.StatusCode >= 400 {
				errorCount++
			}
		}
		
		avgRT := float64(totalRT.Milliseconds()) / float64(len(metrics))
		throughput := float64(len(metrics)) / bucketSize.Seconds()
		errorRate := float64(errorCount) / float64(len(metrics)) * 100
		
		dataPoints = append(dataPoints, TrendDataPoint{
			Timestamp:    timestamp,
			ResponseTime: avgRT,
			Throughput:   throughput,
			ErrorRate:    errorRate,
		})
	}
	
	// Sort by timestamp
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp)
	})
	
	// Determine trend direction
	trend := TrendDirectionStable
	if len(dataPoints) >= 2 {
		first := dataPoints[0].ResponseTime
		last := dataPoints[len(dataPoints)-1].ResponseTime
		change := (last - first) / first * 100
		
		if change > 10 {
			trend = TrendDirectionDegrading
		} else if change < -10 {
			trend = TrendDirectionImproving
		}
	}
	
	// Calculate volatility (coefficient of variation)
	volatility := calculateVolatility(dataPoints)
	
	return PerformanceTrend{
		Endpoint:   endpoint,
		Period:     c.config.WindowSize.String(),
		DataPoints: dataPoints,
		Trend:      trend,
		Volatility: volatility,
	}
}

// Helper functions

func (c *Collector) shouldSample() bool {
	// Simple random sampling
	return time.Now().UnixNano()%100 < int64(c.config.SampleRate*100)
}

func (c *Collector) pruneOldMetrics() {
	cutoff := time.Now().Add(-c.config.WindowSize)
	
	// Find first metric within window
	var i int
	for i = 0; i < len(c.metrics); i++ {
		if c.metrics[i].Timestamp.After(cutoff) {
			break
		}
	}
	
	// Remove old metrics
	if i > 0 {
		c.metrics = c.metrics[i:]
	}
}

func (c *Collector) getUniqueEndpoints() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	seen := make(map[string]bool)
	var endpoints []string
	
	for _, m := range c.metrics {
		if !seen[m.Endpoint] {
			seen[m.Endpoint] = true
			endpoints = append(endpoints, m.Endpoint)
		}
	}
	
	return endpoints
}

func (c *Collector) getSystemLoad() SystemLoad {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	
	return SystemLoad{
		CPU:               0, // Would need OS-specific implementation
		Memory:            float64(mem.Alloc) / float64(mem.Sys) * 100,
		Goroutines:        runtime.NumGoroutine(),
		ActiveConnections: int(c.GetActiveRequests()),
		QueueDepth:        0, // Would track queued requests
		Timestamp:         time.Now(),
	}
}

func (c *Collector) calculateHealthScore(rt ResponseTimeStats, tp ThroughputStats, load SystemLoad) float64 {
	score := 100.0
	
	// Deduct for high response times
	if rt.P95 > 500*time.Millisecond {
		score -= 20
	} else if rt.P95 > 200*time.Millisecond {
		score -= 10
	}
	
	// Deduct for low throughput (relative to peak)
	if tp.PeakRPS > 0 && tp.RequestsPerSecond < tp.PeakRPS*0.5 {
		score -= 10
	}
	
	// Deduct for high memory usage
	if load.Memory > 90 {
		score -= 20
	} else if load.Memory > 75 {
		score -= 10
	}
	
	// Deduct for high goroutine count (potential leak)
	if load.Goroutines > 10000 {
		score -= 15
	} else if load.Goroutines > 5000 {
		score -= 5
	}
	
	return math.Max(0, score)
}

// Calculation helpers

func calculateMean(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, t := range times {
		total += t
	}
	
	return total / time.Duration(len(times))
}

func calculateStdDev(times []time.Duration, mean time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	
	var sumSquares float64
	for _, t := range times {
		diff := float64(t - mean)
		sumSquares += diff * diff
	}
	
	variance := sumSquares / float64(len(times))
	return time.Duration(math.Sqrt(variance))
}

func calculateVolatility(dataPoints []TrendDataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}
	
	// Calculate coefficient of variation for response times
	var sum, sumSquares float64
	for _, dp := range dataPoints {
		sum += dp.ResponseTime
		sumSquares += dp.ResponseTime * dp.ResponseTime
	}
	
	n := float64(len(dataPoints))
	mean := sum / n
	variance := (sumSquares / n) - (mean * mean)
	
	if mean == 0 {
		return 0
	}
	
	stdDev := math.Sqrt(variance)
	return (stdDev / mean) * 100 // Coefficient of variation as percentage
}

// DefaultConfig returns default monitoring configuration
func DefaultConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Enabled:    true,
		SampleRate: 1.0, // Sample all requests
		WindowSize: 5 * time.Minute,
		AlertThresholds: AlertThresholds{
			ResponseTimeP95Ms:     500,
			ResponseTimeP99Ms:     1000,
			ErrorRatePercent:      5,
			ThroughputDropPercent: 50,
			SaturationPercent:     80,
			CPUPercent:            80,
			MemoryPercent:         85,
		},
		RecommendationThresholds: RecommendationThresholds{
			SlowResponseTimeMs:   200,
			HighErrorRatePercent: 3,
			LowCacheHitPercent:   50,
			HighCPUPercent:       70,
			HighMemoryPercent:    75,
		},
	}
}

