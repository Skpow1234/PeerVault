package health

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Skpow1234/Peervault/proto/peervault"
)

// HealthConfig represents health check configuration
type HealthConfig struct {
	CheckInterval   time.Duration
	CheckTimeout    time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	EnableMetrics   bool
	EnableTracing   bool
	EnableProfiling bool
	MetricsInterval time.Duration
	TraceInterval   time.Duration
	ProfileInterval time.Duration
}

// DefaultHealthConfig returns the default health check configuration
func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		CheckInterval:   30 * time.Second,
		CheckTimeout:    5 * time.Second,
		MaxRetries:      3,
		RetryDelay:      time.Second,
		EnableMetrics:   true,
		EnableTracing:   true,
		EnableProfiling: false,
		MetricsInterval: 60 * time.Second,
		TraceInterval:   300 * time.Second,
		ProfileInterval: 600 * time.Second,
	}
}

// AdvancedHealthChecker provides comprehensive health checking
type AdvancedHealthChecker struct {
	config       *HealthConfig
	logger       *slog.Logger
	checkers     map[string]*ComponentHealthChecker
	checkersMux  sync.RWMutex
	dependencies map[string][]string
	depsMux      sync.RWMutex
	aggregator   *HealthAggregator
	metrics      *HealthMetrics
	tracer       *HealthTracer
	profiler     *HealthProfiler
	stopChan     chan struct{}
}

// ComponentHealthChecker represents a health checker for a specific component
type ComponentHealthChecker struct {
	Component    string
	Config       *HealthConfig
	Logger       *slog.Logger
	CheckFunc    func(context.Context) (*HealthResult, error)
	Dependencies []string
	LastCheck    time.Time
	LastResult   *HealthResult
	CheckCount   int64
	FailureCount int64
	mutex        sync.RWMutex
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Component string
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Duration  time.Duration
	Metrics   map[string]interface{}
	Details   map[string]interface{}
	Error     error
}

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthAggregator aggregates health results from multiple components
type HealthAggregator struct {
	config *HealthConfig
	logger *slog.Logger
	mutex  sync.RWMutex
}

// HealthMetrics collects health-related metrics
type HealthMetrics struct {
	config      *HealthConfig
	logger      *slog.Logger
	metrics     map[string]interface{}
	lastUpdated time.Time
	mutex       sync.RWMutex
}

// HealthTracer traces health check operations
type HealthTracer struct {
	config *HealthConfig
	logger *slog.Logger
	traces map[string]*HealthTrace
	mutex  sync.RWMutex
}

// HealthTrace represents a health check trace
type HealthTrace struct {
	ID        string
	Component string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Status    HealthStatus
	Details   map[string]interface{}
}

// HealthProfiler profiles health check performance
type HealthProfiler struct {
	config   *HealthConfig
	logger   *slog.Logger
	profiles map[string]*HealthProfile
	mutex    sync.RWMutex
}

// HealthProfile represents a health check profile
type HealthProfile struct {
	Component     string
	CheckCount    int64
	TotalDuration time.Duration
	AvgDuration   time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	LastUpdated   time.Time
}

// NewAdvancedHealthChecker creates a new advanced health checker
func NewAdvancedHealthChecker(config *HealthConfig, logger *slog.Logger) *AdvancedHealthChecker {
	if config == nil {
		config = DefaultHealthConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &AdvancedHealthChecker{
		config:       config,
		logger:       logger,
		checkers:     make(map[string]*ComponentHealthChecker),
		dependencies: make(map[string][]string),
		aggregator:   NewHealthAggregator(config, logger),
		metrics:      NewHealthMetrics(config, logger),
		tracer:       NewHealthTracer(config, logger),
		profiler:     NewHealthProfiler(config, logger),
		stopChan:     make(chan struct{}),
	}
}

// RegisterComponent registers a health checker for a component
func (ahc *AdvancedHealthChecker) RegisterComponent(component string, checkFunc func(context.Context) (*HealthResult, error), dependencies []string) {
	ahc.checkersMux.Lock()
	defer ahc.checkersMux.Unlock()

	checker := &ComponentHealthChecker{
		Component:    component,
		Config:       ahc.config,
		Logger:       ahc.logger.With("component", component),
		CheckFunc:    checkFunc,
		Dependencies: dependencies,
		LastCheck:    time.Now(),
		LastResult:   &HealthResult{Component: component, Status: HealthStatusUnknown},
	}

	ahc.checkers[component] = checker

	// Register dependencies
	ahc.depsMux.Lock()
	ahc.dependencies[component] = dependencies
	ahc.depsMux.Unlock()

	ahc.logger.Info("Registered health checker for component", "component", component, "dependencies", dependencies)
}

// Start starts the advanced health checker
func (ahc *AdvancedHealthChecker) Start() error {
	ahc.logger.Info("Starting advanced health checker")

	// Start health check loop
	go ahc.healthCheckLoop()

	// Start metrics collection if enabled
	if ahc.config.EnableMetrics {
		go ahc.metricsCollectionLoop()
	}

	// Start tracing if enabled
	if ahc.config.EnableTracing {
		go ahc.tracingLoop()
	}

	// Start profiling if enabled
	if ahc.config.EnableProfiling {
		go ahc.profilingLoop()
	}

	return nil
}

// Stop stops the advanced health checker
func (ahc *AdvancedHealthChecker) Stop() {
	ahc.logger.Info("Stopping advanced health checker")
	close(ahc.stopChan)
}

// healthCheckLoop runs the main health check loop
func (ahc *AdvancedHealthChecker) healthCheckLoop() {
	ticker := time.NewTicker(ahc.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahc.performHealthChecks()
		case <-ahc.stopChan:
			ahc.logger.Info("Health check loop stopped")
			return
		}
	}
}

// performHealthChecks performs health checks for all components
func (ahc *AdvancedHealthChecker) performHealthChecks() {
	ahc.checkersMux.RLock()
	components := make([]string, 0, len(ahc.checkers))
	for component := range ahc.checkers {
		components = append(components, component)
	}
	ahc.checkersMux.RUnlock()

	// Check components in dependency order
	for _, component := range ahc.getDependencyOrder(components) {
		ahc.checkComponent(component)
	}
}

// getDependencyOrder returns components in dependency order
func (ahc *AdvancedHealthChecker) getDependencyOrder(components []string) []string {
	// Simple topological sort - in production, use a proper implementation
	ordered := make([]string, 0, len(components))
	visited := make(map[string]bool)

	var visit func(string)
	visit = func(component string) {
		if visited[component] {
			return
		}
		visited[component] = true

		// Visit dependencies first
		ahc.depsMux.RLock()
		deps := ahc.dependencies[component]
		ahc.depsMux.RUnlock()

		for _, dep := range deps {
			visit(dep)
		}

		ordered = append(ordered, component)
	}

	for _, component := range components {
		visit(component)
	}

	return ordered
}

// checkComponent performs a health check for a specific component
func (ahc *AdvancedHealthChecker) checkComponent(component string) {
	ahc.checkersMux.RLock()
	checker, exists := ahc.checkers[component]
	ahc.checkersMux.RUnlock()

	if !exists {
		return
	}

	// Check dependencies first
	if !ahc.checkDependencies(component) {
		checker.mutex.Lock()
		checker.LastResult = &HealthResult{
			Component: component,
			Status:    HealthStatusDegraded,
			Message:   "Dependencies unhealthy",
			Timestamp: time.Now(),
		}
		checker.mutex.Unlock()
		return
	}

	// Perform health check
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ahc.config.CheckTimeout)
	defer cancel()

	result, err := checker.CheckFunc(ctx)
	if err != nil {
		result = &HealthResult{
			Component: component,
			Status:    HealthStatusUnhealthy,
			Message:   err.Error(),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     err,
		}
	}

	result.Duration = time.Since(start)

	// Update checker
	checker.mutex.Lock()
	checker.LastCheck = time.Now()
	checker.LastResult = result
	checker.CheckCount++
	if result.Status != HealthStatusHealthy {
		checker.FailureCount++
	}
	checker.mutex.Unlock()

	// Update metrics
	if ahc.config.EnableMetrics {
		ahc.metrics.UpdateComponentMetrics(component, result)
	}

	// Update tracer
	if ahc.config.EnableTracing {
		ahc.tracer.RecordTrace(component, result)
	}

	// Update profiler
	if ahc.config.EnableProfiling {
		ahc.profiler.UpdateProfile(component, result)
	}

	ahc.logger.Debug("Health check completed", "component", component, "status", result.Status, "duration", result.Duration)
}

// checkDependencies checks if all dependencies are healthy
func (ahc *AdvancedHealthChecker) checkDependencies(component string) bool {
	ahc.depsMux.RLock()
	deps := ahc.dependencies[component]
	ahc.depsMux.RUnlock()

	for _, dep := range deps {
		ahc.checkersMux.RLock()
		depChecker, exists := ahc.checkers[dep]
		ahc.checkersMux.RUnlock()

		if !exists {
			continue
		}

		depChecker.mutex.RLock()
		status := depChecker.LastResult.Status
		depChecker.mutex.RUnlock()

		if status != HealthStatusHealthy {
			return false
		}
	}

	return true
}

// GetHealthStatus returns the overall health status
func (ahc *AdvancedHealthChecker) GetHealthStatus() *peervault.HealthResponse {
	ahc.checkersMux.RLock()
	defer ahc.checkersMux.RUnlock()

	overallStatus := HealthStatusHealthy
	healthyCount := 0
	totalCount := len(ahc.checkers)

	for _, checker := range ahc.checkers {
		checker.mutex.RLock()
		result := checker.LastResult
		checker.mutex.RUnlock()

		if result.Status == HealthStatusHealthy {
			healthyCount++
		} else if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	// Calculate health percentage
	_ = float64(healthyCount) / float64(totalCount) * 100 // healthPercentage

	return &peervault.HealthResponse{
		Status:    string(overallStatus),
		Timestamp: timestamppb.Now(),
		Version:   "1.0.0",
		// Note: Metadata field not available in current proto definition
	}
}

// GetDetailedHealthStatus returns detailed health status for all components
func (ahc *AdvancedHealthChecker) GetDetailedHealthStatus() map[string]interface{} {
	ahc.checkersMux.RLock()
	defer ahc.checkersMux.RUnlock()

	status := make(map[string]interface{})
	status["timestamp"] = time.Now()
	status["overall_status"] = ahc.getOverallStatus()
	status["components"] = make([]map[string]interface{}, 0)

	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0
	unknownCount := 0

	for _, checker := range ahc.checkers {
		checker.mutex.RLock()
		result := checker.LastResult
		checkCount := checker.CheckCount
		failureCount := checker.FailureCount
		checker.mutex.RUnlock()

		componentStatus := map[string]interface{}{
			"component":     result.Component,
			"status":        result.Status,
			"message":       result.Message,
			"timestamp":     result.Timestamp,
			"duration":      result.Duration,
			"check_count":   checkCount,
			"failure_count": failureCount,
			"success_rate":  float64(checkCount-failureCount) / float64(checkCount) * 100,
			"metrics":       result.Metrics,
			"details":       result.Details,
		}

		if result.Error != nil {
			componentStatus["error"] = result.Error.Error()
		}

		status["components"] = append(status["components"].([]map[string]interface{}), componentStatus)

		switch result.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnknown:
			unknownCount++
		}
	}

	status["summary"] = map[string]interface{}{
		"healthy":   healthyCount,
		"unhealthy": unhealthyCount,
		"degraded":  degradedCount,
		"unknown":   unknownCount,
		"total":     len(ahc.checkers),
	}

	// Add aggregated metrics
	if ahc.config.EnableMetrics {
		status["metrics"] = ahc.metrics.GetMetrics()
	}

	// Add traces
	if ahc.config.EnableTracing {
		status["traces"] = ahc.tracer.GetTraces()
	}

	// Add profiles
	if ahc.config.EnableProfiling {
		status["profiles"] = ahc.profiler.GetProfiles()
	}

	return status
}

// getOverallStatus determines the overall health status
func (ahc *AdvancedHealthChecker) getOverallStatus() HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, checker := range ahc.checkers {
		checker.mutex.RLock()
		status := checker.LastResult.Status
		checker.mutex.RUnlock()

		switch status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	} else if hasDegraded {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// ForceHealthCheck forces an immediate health check for a component
func (ahc *AdvancedHealthChecker) ForceHealthCheck(component string) error {
	ahc.checkersMux.RLock()
	_, exists := ahc.checkers[component]
	ahc.checkersMux.RUnlock()

	if !exists {
		return fmt.Errorf("health checker not found for component: %s", component)
	}

	ahc.checkComponent(component)
	ahc.logger.Info("Forced health check for component", "component", component)
	return nil
}

// ForceHealthCheckAll forces an immediate health check for all components
func (ahc *AdvancedHealthChecker) ForceHealthCheckAll() {
	ahc.checkersMux.RLock()
	components := make([]string, 0, len(ahc.checkers))
	for component := range ahc.checkers {
		components = append(components, component)
	}
	ahc.checkersMux.RUnlock()

	for _, component := range ahc.getDependencyOrder(components) {
		ahc.checkComponent(component)
	}

	ahc.logger.Info("Forced health check for all components")
}

// GetComponentHealth returns health status for a specific component
func (ahc *AdvancedHealthChecker) GetComponentHealth(component string) (*HealthResult, error) {
	ahc.checkersMux.RLock()
	checker, exists := ahc.checkers[component]
	ahc.checkersMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("health checker not found for component: %s", component)
	}

	checker.mutex.RLock()
	result := checker.LastResult
	checker.mutex.RUnlock()

	return result, nil
}

// GetHealthMetrics returns health metrics
func (ahc *AdvancedHealthChecker) GetHealthMetrics() map[string]interface{} {
	return ahc.metrics.GetMetrics()
}

// GetHealthTraces returns health traces
func (ahc *AdvancedHealthChecker) GetHealthTraces() map[string]*HealthTrace {
	return ahc.tracer.GetTraces()
}

// GetHealthProfiles returns health profiles
func (ahc *AdvancedHealthChecker) GetHealthProfiles() map[string]*HealthProfile {
	return ahc.profiler.GetProfiles()
}

// metricsCollectionLoop runs the metrics collection loop
func (ahc *AdvancedHealthChecker) metricsCollectionLoop() {
	ticker := time.NewTicker(ahc.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahc.metrics.CollectMetrics(ahc.checkers)
		case <-ahc.stopChan:
			ahc.logger.Info("Metrics collection loop stopped")
			return
		}
	}
}

// tracingLoop runs the tracing loop
func (ahc *AdvancedHealthChecker) tracingLoop() {
	ticker := time.NewTicker(ahc.config.TraceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahc.tracer.CleanupOldTraces()
		case <-ahc.stopChan:
			ahc.logger.Info("Tracing loop stopped")
			return
		}
	}
}

// profilingLoop runs the profiling loop
func (ahc *AdvancedHealthChecker) profilingLoop() {
	ticker := time.NewTicker(ahc.config.ProfileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ahc.profiler.CleanupOldProfiles()
		case <-ahc.stopChan:
			ahc.logger.Info("Profiling loop stopped")
			return
		}
	}
}
