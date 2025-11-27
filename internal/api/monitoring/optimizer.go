package monitoring

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Optimizer generates optimization recommendations
type Optimizer struct {
	config    *MonitoringConfig
	collector *Collector
}

// NewOptimizer creates a new optimizer
func NewOptimizer(config *MonitoringConfig, collector *Collector) *Optimizer {
	return &Optimizer{
		config:    config,
		collector: collector,
	}
}

// GenerateRecommendations analyzes performance and generates recommendations
func (o *Optimizer) GenerateRecommendations() []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation
	
	snapshot := o.collector.GetPerformanceSnapshot(nil)
	
	// Analyze overall performance
	recommendations = append(recommendations, o.analyzeResponseTimes(snapshot)...)
	recommendations = append(recommendations, o.analyzeErrorRates(snapshot)...)
	recommendations = append(recommendations, o.analyzeThroughput(snapshot)...)
	recommendations = append(recommendations, o.analyzeSystemLoad(snapshot)...)
	recommendations = append(recommendations, o.analyzeEndpointPatterns(snapshot)...)
	
	return recommendations
}

// analyzeResponseTimes generates recommendations for slow response times
func (o *Optimizer) analyzeResponseTimes(snapshot PerformanceSnapshot) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation
	threshold := o.config.RecommendationThresholds.SlowResponseTimeMs
	
	for _, ep := range snapshot.EndpointStats {
		avgMs := float64(ep.ResponseTime.Mean.Milliseconds())
		
		if avgMs > threshold {
			priority := o.getPriorityForLatency(avgMs, threshold)
			
			// Caching recommendation
			if avgMs > threshold*2 {
				recommendations = append(recommendations, OptimizationRecommendation{
					ID:       uuid.New().String(),
					Type:     RecommendationTypeCaching,
					Priority: priority,
					Title:    fmt.Sprintf("Implement caching for %s", ep.Endpoint),
					Description: fmt.Sprintf(
						"Endpoint %s has high average response time (%.2fms). "+
							"Implementing caching could reduce this significantly.",
						ep.Endpoint, avgMs,
					),
					Endpoint:       ep.Endpoint,
					ImpactEstimate: fmt.Sprintf("Reduce latency by 50-80%% (%.2fms → %.2fms)", avgMs, avgMs*0.3),
					Effort:         "Medium",
					Actions: []string{
						"Identify cacheable data in the endpoint handler",
						"Implement Redis or in-memory cache",
						"Set appropriate cache TTL based on data volatility",
						"Add cache invalidation logic for data updates",
					},
					Rationale: fmt.Sprintf(
						"Current P95 is %.2fms (%.0f%% above threshold). Caching could provide substantial improvements.",
						float64(ep.ResponseTime.P95.Milliseconds()),
						(avgMs-threshold)/threshold*100,
					),
					Metrics: map[string]float64{
						"current_avg_ms": avgMs,
						"threshold_ms":   threshold,
						"p95_ms":         float64(ep.ResponseTime.P95.Milliseconds()),
						"p99_ms":         float64(ep.ResponseTime.P99.Milliseconds()),
					},
					GeneratedAt: time.Now(),
				})
			}
			
			// Code optimization recommendation
			recommendations = append(recommendations, OptimizationRecommendation{
				ID:       uuid.New().String(),
				Type:     RecommendationTypeCodeOptimization,
				Priority: priority,
				Title:    fmt.Sprintf("Optimize code execution for %s", ep.Endpoint),
				Description: fmt.Sprintf(
					"The endpoint %s could benefit from code optimization to reduce response time from %.2fms.",
					ep.Endpoint, avgMs,
				),
				Endpoint:       ep.Endpoint,
				ImpactEstimate: fmt.Sprintf("Reduce latency by 20-40%% (%.2fms → %.2fms)", avgMs, avgMs*0.7),
				Effort:         "High",
				Actions: []string{
					"Profile the endpoint to identify bottlenecks",
					"Optimize database queries (use EXPLAIN ANALYZE)",
					"Reduce unnecessary data processing",
					"Consider async processing for non-critical operations",
					"Optimize loops and data structures",
				},
				Rationale: "Response time analysis indicates significant execution time. Profiling could reveal optimization opportunities.",
				Metrics: map[string]float64{
					"current_avg_ms": avgMs,
					"std_dev_ms":     float64(ep.ResponseTime.StandardDeviation.Milliseconds()),
				},
				GeneratedAt: time.Now(),
			})
		}
	}
	
	return recommendations
}

// analyzeErrorRates generates recommendations for high error rates
func (o *Optimizer) analyzeErrorRates(snapshot PerformanceSnapshot) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation
	threshold := o.config.RecommendationThresholds.HighErrorRatePercent
	
	for _, ep := range snapshot.EndpointStats {
		if ep.ErrorRate > threshold {
			recommendations = append(recommendations, OptimizationRecommendation{
				ID:       uuid.New().String(),
				Type:     RecommendationTypeCodeOptimization,
				Priority: RecommendationPriorityHigh,
				Title:    fmt.Sprintf("Investigate errors in %s", ep.Endpoint),
				Description: fmt.Sprintf(
					"Endpoint %s has a high error rate (%.2f%%). This requires immediate attention.",
					ep.Endpoint, ep.ErrorRate,
				),
				Endpoint:       ep.Endpoint,
				ImpactEstimate: "Improve reliability and user experience significantly",
				Effort:         "High",
				Actions: []string{
					"Review error logs for common error patterns",
					"Add better error handling and validation",
					"Implement circuit breakers for external dependencies",
					"Add retry logic with exponential backoff",
					"Improve input validation",
					"Add monitoring alerts for error spikes",
				},
				Rationale: fmt.Sprintf(
					"Error rate of %.2f%% is %.0f%% above the threshold. Errors impact user experience and may indicate bugs.",
					ep.ErrorRate, (ep.ErrorRate-threshold)/threshold*100,
				),
				Metrics: map[string]float64{
					"error_rate":   ep.ErrorRate,
					"threshold":    threshold,
					"success_rate": ep.SuccessRate,
				},
				GeneratedAt: time.Now(),
			})
		}
	}
	
	return recommendations
}

// analyzeThroughput generates recommendations for throughput issues
func (o *Optimizer) analyzeThroughput(snapshot PerformanceSnapshot) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation
	
	// Check if throughput is significantly below peak
	for _, ep := range snapshot.EndpointStats {
		if ep.Throughput.PeakRPS > 0 {
			currentRPS := ep.Throughput.RequestsPerSecond
			dropPercent := (ep.Throughput.PeakRPS - currentRPS) / ep.Throughput.PeakRPS * 100
			
			if dropPercent > 50 && ep.Saturation > 70 {
				recommendations = append(recommendations, OptimizationRecommendation{
					ID:       uuid.New().String(),
					Type:     RecommendationTypeScaling,
					Priority: RecommendationPriorityMedium,
					Title:    fmt.Sprintf("Scale capacity for %s", ep.Endpoint),
					Description: fmt.Sprintf(
						"Endpoint %s is experiencing reduced throughput (%.2f RPS vs peak of %.2f RPS) "+
							"while saturation is high (%.2f%%).",
						ep.Endpoint, currentRPS, ep.Throughput.PeakRPS, ep.Saturation,
					),
					Endpoint:       ep.Endpoint,
					ImpactEstimate: "Restore throughput to peak levels and reduce saturation",
					Effort:         "Medium",
					Actions: []string{
						"Scale horizontally by adding more instances",
						"Implement load balancing",
						"Consider auto-scaling based on metrics",
						"Optimize resource allocation",
					},
					Rationale: fmt.Sprintf(
						"Throughput has dropped %.2f%% from peak while saturation remains high. This indicates capacity constraints.",
						dropPercent,
					),
					Metrics: map[string]float64{
						"current_rps": currentRPS,
						"peak_rps":    ep.Throughput.PeakRPS,
						"saturation":  ep.Saturation,
						"drop_percent": dropPercent,
					},
					GeneratedAt: time.Now(),
				})
			}
		}
	}
	
	return recommendations
}

// analyzeSystemLoad generates recommendations for system resource issues
func (o *Optimizer) analyzeSystemLoad(snapshot PerformanceSnapshot) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation
	
	// High memory usage
	if snapshot.SystemLoad.Memory > o.config.RecommendationThresholds.HighMemoryPercent {
		recommendations = append(recommendations, OptimizationRecommendation{
			ID:       uuid.New().String(),
			Type:     RecommendationTypeConfiguration,
			Priority: RecommendationPriorityHigh,
			Title:    "Optimize memory usage",
			Description: fmt.Sprintf(
				"System memory usage is high (%.2f%%). This could lead to performance degradation or crashes.",
				snapshot.SystemLoad.Memory,
			),
			ImpactEstimate: "Prevent OOM errors and improve stability",
			Effort:         "Medium",
			Actions: []string{
				"Profile memory usage to identify leaks",
				"Implement connection pooling with limits",
				"Add memory-aware caching with eviction policies",
				"Optimize data structures and reduce allocations",
				"Consider increasing system memory if needed",
			},
			Rationale: "High memory usage can lead to garbage collection pressure, swapping, and system instability.",
			Metrics: map[string]float64{
				"memory_percent": snapshot.SystemLoad.Memory,
				"threshold":      o.config.RecommendationThresholds.HighMemoryPercent,
			},
			GeneratedAt: time.Now(),
		})
	}
	
	// High goroutine count
	if snapshot.SystemLoad.Goroutines > 5000 {
		recommendations = append(recommendations, OptimizationRecommendation{
			ID:       uuid.New().String(),
			Type:     RecommendationTypeCodeOptimization,
			Priority: RecommendationPriorityHigh,
			Title:    "Investigate goroutine leak",
			Description: fmt.Sprintf(
				"Very high goroutine count (%d). This may indicate a goroutine leak.",
				snapshot.SystemLoad.Goroutines,
			),
			ImpactEstimate: "Prevent resource exhaustion and improve stability",
			Effort:         "High",
			Actions: []string{
				"Use pprof to identify goroutine sources",
				"Review code for missing channel closes",
				"Ensure all goroutines have proper termination conditions",
				"Add context cancellation where appropriate",
				"Implement goroutine pools for recurring tasks",
			},
			Rationale: "Excessive goroutines consume memory and scheduler resources, degrading overall performance.",
			Metrics: map[string]float64{
				"goroutines": float64(snapshot.SystemLoad.Goroutines),
			},
			GeneratedAt: time.Now(),
		})
	}
	
	return recommendations
}

// analyzeEndpointPatterns looks for patterns that suggest optimizations
func (o *Optimizer) analyzeEndpointPatterns(snapshot PerformanceSnapshot) []OptimizationRecommendation {
	var recommendations []OptimizationRecommendation
	
	// Find endpoints with high variability (high std dev)
	for _, ep := range snapshot.EndpointStats {
		stdDevMs := float64(ep.ResponseTime.StandardDeviation.Milliseconds())
		meanMs := float64(ep.ResponseTime.Mean.Milliseconds())
		
		if meanMs > 0 {
			cv := (stdDevMs / meanMs) * 100 // Coefficient of variation
			
			if cv > 50 { // High variability
				recommendations = append(recommendations, OptimizationRecommendation{
					ID:       uuid.New().String(),
					Type:     RecommendationTypeCodeOptimization,
					Priority: RecommendationPriorityMedium,
					Title:    fmt.Sprintf("Reduce response time variability for %s", ep.Endpoint),
					Description: fmt.Sprintf(
						"Endpoint %s has high response time variability (CV: %.2f%%). "+
							"This indicates inconsistent performance.",
						ep.Endpoint, cv,
					),
					Endpoint:       ep.Endpoint,
					ImpactEstimate: "More predictable performance and better user experience",
					Effort:         "Medium",
					Actions: []string{
						"Identify causes of slow requests (slow queries, external API calls)",
						"Implement timeouts for external dependencies",
						"Add request queuing to smooth out spikes",
						"Consider background processing for variable-time operations",
					},
					Rationale: "High variability suggests some requests are significantly slower than others, indicating potential optimization targets.",
					Metrics: map[string]float64{
						"std_dev_ms":          stdDevMs,
						"mean_ms":             meanMs,
						"coefficient_variation": cv,
						"min_ms":              float64(ep.ResponseTime.Min.Milliseconds()),
						"max_ms":              float64(ep.ResponseTime.Max.Milliseconds()),
					},
					GeneratedAt: time.Now(),
				})
			}
		}
	}
	
	return recommendations
}

// getPriorityForLatency determines recommendation priority based on latency
func (o *Optimizer) getPriorityForLatency(current, threshold float64) RecommendationPriority {
	ratio := current / threshold
	
	if ratio > 3 {
		return RecommendationPriorityHigh
	} else if ratio > 2 {
		return RecommendationPriorityMedium
	}
	return RecommendationPriorityLow
}

