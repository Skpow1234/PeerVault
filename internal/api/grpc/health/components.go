package health

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// NewHealthAggregator creates a new health aggregator
func NewHealthAggregator(config *HealthConfig, logger *slog.Logger) *HealthAggregator {
	return &HealthAggregator{
		config: config,
		logger: logger,
	}
}

// AggregateHealth aggregates health results from multiple components
func (ha *HealthAggregator) AggregateHealth(results []*HealthResult) *HealthResult {
	ha.mutex.Lock()
	defer ha.mutex.Unlock()

	if len(results) == 0 {
		return &HealthResult{
			Component: "aggregated",
			Status:    HealthStatusUnknown,
			Message:   "No components to aggregate",
			Timestamp: time.Now(),
		}
	}

	overallStatus := HealthStatusHealthy
	healthyCount := 0
	totalDuration := time.Duration(0)
	aggregatedMetrics := make(map[string]interface{})
	aggregatedDetails := make(map[string]interface{})

	for _, result := range results {
		totalDuration += result.Duration

		if result.Status == HealthStatusHealthy {
			healthyCount++
		} else if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}

		// Aggregate metrics
		for key, value := range result.Metrics {
			aggregatedMetrics[key] = value
		}

		// Aggregate details
		for key, value := range result.Details {
			aggregatedDetails[key] = value
		}
	}

	healthPercentage := float64(healthyCount) / float64(len(results)) * 100

	return &HealthResult{
		Component: "aggregated",
		Status:    overallStatus,
		Message:   fmt.Sprintf("Health percentage: %.2f%%", healthPercentage),
		Timestamp: time.Now(),
		Duration:  totalDuration,
		Metrics: map[string]interface{}{
			"healthy_components": healthyCount,
			"total_components":   len(results),
			"health_percentage":  healthPercentage,
		},
		Details: aggregatedDetails,
	}
}

// NewHealthMetrics creates a new health metrics collector
func NewHealthMetrics(config *HealthConfig, logger *slog.Logger) *HealthMetrics {
	return &HealthMetrics{
		config:      config,
		logger:      logger,
		metrics:     make(map[string]interface{}),
		lastUpdated: time.Now(),
	}
}

// UpdateComponentMetrics updates metrics for a component
func (hm *HealthMetrics) UpdateComponentMetrics(component string, result *HealthResult) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	componentMetrics := map[string]interface{}{
		"status":        result.Status,
		"duration":      result.Duration,
		"timestamp":     result.Timestamp,
		"check_count":   1,     // This would be incremented in a real implementation
		"failure_count": 0,     // This would be incremented based on status
		"success_rate":  100.0, // This would be calculated
	}

	hm.metrics[component] = componentMetrics
	hm.lastUpdated = time.Now()
}

// CollectMetrics collects metrics from all components
func (hm *HealthMetrics) CollectMetrics(checkers map[string]*ComponentHealthChecker) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	totalChecks := int64(0)
	totalFailures := int64(0)
	totalDuration := time.Duration(0)

	for component, checker := range checkers {
		checker.mutex.RLock()
		checkCount := checker.CheckCount
		failureCount := checker.FailureCount
		lastResult := checker.LastResult
		checker.mutex.RUnlock()

		totalChecks += checkCount
		totalFailures += failureCount
		totalDuration += lastResult.Duration

		successRate := float64(0)
		if checkCount > 0 {
			successRate = float64(checkCount-failureCount) / float64(checkCount) * 100
		}

		componentMetrics := map[string]interface{}{
			"check_count":   checkCount,
			"failure_count": failureCount,
			"success_rate":  successRate,
			"last_duration": lastResult.Duration,
			"last_status":   lastResult.Status,
			"last_check":    lastResult.Timestamp,
		}

		hm.metrics[component] = componentMetrics
	}

	// Update overall metrics
	overallSuccessRate := float64(0)
	if totalChecks > 0 {
		overallSuccessRate = float64(totalChecks-totalFailures) / float64(totalChecks) * 100
	}

	hm.metrics["overall"] = map[string]interface{}{
		"total_checks":         totalChecks,
		"total_failures":       totalFailures,
		"overall_success_rate": overallSuccessRate,
		"total_duration":       totalDuration,
		"last_updated":         time.Now(),
	}

	hm.lastUpdated = time.Now()
}

// GetMetrics returns current metrics
func (hm *HealthMetrics) GetMetrics() map[string]interface{} {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	metrics := make(map[string]interface{})
	for key, value := range hm.metrics {
		metrics[key] = value
	}

	metrics["last_updated"] = hm.lastUpdated
	return metrics
}

// NewHealthTracer creates a new health tracer
func NewHealthTracer(config *HealthConfig, logger *slog.Logger) *HealthTracer {
	return &HealthTracer{
		config: config,
		logger: logger,
		traces: make(map[string]*HealthTrace),
	}
}

// RecordTrace records a health check trace
func (ht *HealthTracer) RecordTrace(component string, result *HealthResult) {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	traceID := fmt.Sprintf("%s_%d", component, time.Now().UnixNano())
	trace := &HealthTrace{
		ID:        traceID,
		Component: component,
		StartTime: result.Timestamp.Add(-result.Duration),
		EndTime:   result.Timestamp,
		Duration:  result.Duration,
		Status:    result.Status,
		Details: map[string]interface{}{
			"message": result.Message,
			"metrics": result.Metrics,
		},
	}

	if result.Error != nil {
		trace.Details["error"] = result.Error.Error()
	}

	ht.traces[traceID] = trace
}

// GetTraces returns current traces
func (ht *HealthTracer) GetTraces() map[string]*HealthTrace {
	ht.mutex.RLock()
	defer ht.mutex.RUnlock()

	traces := make(map[string]*HealthTrace)
	for traceID, trace := range ht.traces {
		traces[traceID] = trace
	}

	return traces
}

// CleanupOldTraces removes old traces
func (ht *HealthTracer) CleanupOldTraces() {
	ht.mutex.Lock()
	defer ht.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Keep traces for 24 hours
	for traceID, trace := range ht.traces {
		if trace.EndTime.Before(cutoff) {
			delete(ht.traces, traceID)
		}
	}
}

// NewHealthProfiler creates a new health profiler
func NewHealthProfiler(config *HealthConfig, logger *slog.Logger) *HealthProfiler {
	return &HealthProfiler{
		config:   config,
		logger:   logger,
		profiles: make(map[string]*HealthProfile),
	}
}

// UpdateProfile updates the profile for a component
func (hp *HealthProfiler) UpdateProfile(component string, result *HealthResult) {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	profile, exists := hp.profiles[component]
	if !exists {
		profile = &HealthProfile{
			Component:   component,
			CheckCount:  0,
			MinDuration: result.Duration,
			MaxDuration: result.Duration,
			LastUpdated: time.Now(),
		}
		hp.profiles[component] = profile
	}

	profile.CheckCount++
	profile.TotalDuration += result.Duration
	profile.AvgDuration = profile.TotalDuration / time.Duration(profile.CheckCount)
	profile.LastUpdated = time.Now()

	if result.Duration < profile.MinDuration {
		profile.MinDuration = result.Duration
	}
	if result.Duration > profile.MaxDuration {
		profile.MaxDuration = result.Duration
	}
}

// GetProfiles returns current profiles
func (hp *HealthProfiler) GetProfiles() map[string]*HealthProfile {
	hp.mutex.RLock()
	defer hp.mutex.RUnlock()

	profiles := make(map[string]*HealthProfile)
	for component, profile := range hp.profiles {
		profiles[component] = profile
	}

	return profiles
}

// CleanupOldProfiles removes old profiles
func (hp *HealthProfiler) CleanupOldProfiles() {
	hp.mutex.Lock()
	defer hp.mutex.Unlock()

	cutoff := time.Now().Add(-7 * 24 * time.Hour) // Keep profiles for 7 days
	for component, profile := range hp.profiles {
		if profile.LastUpdated.Before(cutoff) {
			delete(hp.profiles, component)
		}
	}
}

// Built-in health check functions

// DatabaseHealthCheck provides a health check for database connections
func DatabaseHealthCheck(ctx context.Context) (*HealthResult, error) {
	// This is a mock implementation
	// In a real implementation, you would check database connectivity
	start := time.Now()

	// Simulate database check
	time.Sleep(10 * time.Millisecond)

	return &HealthResult{
		Component: "database",
		Status:    HealthStatusHealthy,
		Message:   "Database connection healthy",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Metrics: map[string]interface{}{
			"connection_count": 10,
			"active_queries":   5,
			"pool_size":        20,
		},
		Details: map[string]interface{}{
			"database_type": "postgresql",
			"version":       "13.4",
		},
	}, nil
}

// CacheHealthCheck provides a health check for cache systems
func CacheHealthCheck(ctx context.Context) (*HealthResult, error) {
	// This is a mock implementation
	start := time.Now()

	// Simulate cache check
	time.Sleep(5 * time.Millisecond)

	return &HealthResult{
		Component: "cache",
		Status:    HealthStatusHealthy,
		Message:   "Cache system healthy",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Metrics: map[string]interface{}{
			"hit_rate":     0.95,
			"miss_rate":    0.05,
			"memory_usage": "512MB",
		},
		Details: map[string]interface{}{
			"cache_type": "redis",
			"version":    "6.2.6",
		},
	}, nil
}

// StorageHealthCheck provides a health check for storage systems
func StorageHealthCheck(ctx context.Context) (*HealthResult, error) {
	// This is a mock implementation
	start := time.Now()

	// Simulate storage check
	time.Sleep(15 * time.Millisecond)

	return &HealthResult{
		Component: "storage",
		Status:    HealthStatusHealthy,
		Message:   "Storage system healthy",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Metrics: map[string]interface{}{
			"disk_usage":    "75%",
			"free_space":    "25GB",
			"io_operations": 1000,
		},
		Details: map[string]interface{}{
			"storage_type": "local",
			"filesystem":   "ext4",
		},
	}, nil
}

// NetworkHealthCheck provides a health check for network connectivity
func NetworkHealthCheck(ctx context.Context) (*HealthResult, error) {
	// This is a mock implementation
	start := time.Now()

	// Simulate network check
	time.Sleep(20 * time.Millisecond)

	return &HealthResult{
		Component: "network",
		Status:    HealthStatusHealthy,
		Message:   "Network connectivity healthy",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Metrics: map[string]interface{}{
			"latency_ms":     5.2,
			"packet_loss":    0.0,
			"bandwidth_mbps": 1000,
		},
		Details: map[string]interface{}{
			"interface":  "eth0",
			"ip_address": "192.168.1.100",
		},
	}, nil
}

// ServiceHealthCheck provides a health check for external services
func ServiceHealthCheck(serviceName, endpoint string) func(context.Context) (*HealthResult, error) {
	return func(ctx context.Context) (*HealthResult, error) {
		// This is a mock implementation
		start := time.Now()

		// Simulate service check
		time.Sleep(30 * time.Millisecond)

		return &HealthResult{
			Component: serviceName,
			Status:    HealthStatusHealthy,
			Message:   fmt.Sprintf("Service %s healthy", serviceName),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Metrics: map[string]interface{}{
				"response_time_ms": 30.0,
				"status_code":      200,
			},
			Details: map[string]interface{}{
				"endpoint":     endpoint,
				"service_type": "http",
			},
		}, nil
	}
}

// CustomHealthCheck creates a custom health check function
func CustomHealthCheck(component string, checkFunc func(context.Context) error) func(context.Context) (*HealthResult, error) {
	return func(ctx context.Context) (*HealthResult, error) {
		start := time.Now()

		err := checkFunc(ctx)
		status := HealthStatusHealthy
		message := "Component healthy"

		if err != nil {
			status = HealthStatusUnhealthy
			message = err.Error()
		}

		return &HealthResult{
			Component: component,
			Status:    status,
			Message:   message,
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     err,
		}, nil
	}
}
