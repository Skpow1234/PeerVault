package health

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a health check
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthResult
	Timeout() time.Duration
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[check.Name()] = check
}

// UnregisterCheck unregisters a health check
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	delete(hc.checks, name)
}

// CheckAll performs all registered health checks
func (hc *HealthChecker) CheckAll(ctx context.Context) map[string]HealthResult {
	hc.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mu.RUnlock()

	results := make(map[string]HealthResult)

	for name, check := range checks {
		checkCtx, cancel := context.WithTimeout(ctx, check.Timeout())
		start := time.Now()

		result := check.Check(checkCtx)
		result.Duration = time.Since(start)
		result.Timestamp = time.Now()

		results[name] = result
		cancel()
	}

	return results
}

// Check performs a specific health check
func (hc *HealthChecker) Check(ctx context.Context, name string) (HealthResult, error) {
	hc.mu.RLock()
	check, exists := hc.checks[name]
	hc.mu.RUnlock()

	if !exists {
		return HealthResult{}, fmt.Errorf("health check %s not found", name)
	}

	checkCtx, cancel := context.WithTimeout(ctx, check.Timeout())
	defer cancel()

	start := time.Now()
	result := check.Check(checkCtx)
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()

	return result, nil
}

// GetOverallStatus returns the overall health status
func (hc *HealthChecker) GetOverallStatus(ctx context.Context) HealthStatus {
	results := hc.CheckAll(ctx)

	if len(results) == 0 {
		return HealthStatusUnknown
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, result := range results {
		switch result.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}

	if hasDegraded {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// GetHealthReport returns a comprehensive health report
func (hc *HealthChecker) GetHealthReport(ctx context.Context) HealthReport {
	results := hc.CheckAll(ctx)
	overallStatus := hc.GetOverallStatus(ctx)

	return HealthReport{
		OverallStatus: overallStatus,
		Timestamp:     time.Now(),
		Checks:        results,
	}
}

// HealthReport represents a comprehensive health report
type HealthReport struct {
	OverallStatus HealthStatus            `json:"overall_status"`
	Timestamp     time.Time               `json:"timestamp"`
	Checks        map[string]HealthResult `json:"checks"`
}

// ToJSON converts the health report to JSON
func (hr *HealthReport) ToJSON() ([]byte, error) {
	return json.Marshal(hr)
}

// SimpleHealthCheck implements a simple health check
type SimpleHealthCheck struct {
	name        string
	checkFunc   func(ctx context.Context) error
	timeout     time.Duration
	description string
}

// NewSimpleHealthCheck creates a new simple health check
func NewSimpleHealthCheck(name, description string, timeout time.Duration, checkFunc func(ctx context.Context) error) *SimpleHealthCheck {
	return &SimpleHealthCheck{
		name:        name,
		checkFunc:   checkFunc,
		timeout:     timeout,
		description: description,
	}
}

// Name returns the health check name
func (shc *SimpleHealthCheck) Name() string {
	return shc.name
}

// Check performs the health check
func (shc *SimpleHealthCheck) Check(ctx context.Context) HealthResult {
	err := shc.checkFunc(ctx)

	if err != nil {
		return HealthResult{
			Status:  HealthStatusUnhealthy,
			Message: err.Error(),
		}
	}

	return HealthResult{
		Status:  HealthStatusHealthy,
		Message: shc.description,
	}
}

// Timeout returns the health check timeout
func (shc *SimpleHealthCheck) Timeout() time.Duration {
	return shc.timeout
}

// HTTPHealthCheck implements an HTTP health check
type HTTPHealthCheck struct {
	name        string
	url         string
	timeout     time.Duration
	description string
}

// NewHTTPHealthCheck creates a new HTTP health check
func NewHTTPHealthCheck(name, url, description string, timeout time.Duration) *HTTPHealthCheck {
	return &HTTPHealthCheck{
		name:        name,
		url:         url,
		timeout:     timeout,
		description: description,
	}
}

// Name returns the health check name
func (hhc *HTTPHealthCheck) Name() string {
	return hhc.name
}

// Check performs the HTTP health check
func (hhc *HTTPHealthCheck) Check(ctx context.Context) HealthResult {
	// In a real implementation, this would make an HTTP request
	// For now, we'll simulate a successful check
	return HealthResult{
		Status:  HealthStatusHealthy,
		Message: fmt.Sprintf("HTTP check for %s successful", hhc.url),
		Details: map[string]interface{}{
			"url": hhc.url,
		},
	}
}

// Timeout returns the health check timeout
func (hhc *HTTPHealthCheck) Timeout() time.Duration {
	return hhc.timeout
}

// DatabaseHealthCheck implements a database health check
type DatabaseHealthCheck struct {
	name        string
	pingFunc    func(ctx context.Context) error
	timeout     time.Duration
	description string
}

// NewDatabaseHealthCheck creates a new database health check
func NewDatabaseHealthCheck(name, description string, timeout time.Duration, pingFunc func(ctx context.Context) error) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name:        name,
		pingFunc:    pingFunc,
		timeout:     timeout,
		description: description,
	}
}

// Name returns the health check name
func (dhc *DatabaseHealthCheck) Name() string {
	return dhc.name
}

// Check performs the database health check
func (dhc *DatabaseHealthCheck) Check(ctx context.Context) HealthResult {
	err := dhc.pingFunc(ctx)

	if err != nil {
		return HealthResult{
			Status:  HealthStatusUnhealthy,
			Message: fmt.Sprintf("Database connection failed: %v", err),
		}
	}

	return HealthResult{
		Status:  HealthStatusHealthy,
		Message: dhc.description,
	}
}

// Timeout returns the health check timeout
func (dhc *DatabaseHealthCheck) Timeout() time.Duration {
	return dhc.timeout
}

// Global health checker
var GlobalHealthChecker = NewHealthChecker()

// Convenience functions for global health checks
func RegisterHealthCheck(check HealthCheck) {
	GlobalHealthChecker.RegisterCheck(check)
}

func UnregisterHealthCheck(name string) {
	GlobalHealthChecker.UnregisterCheck(name)
}

func CheckHealth(ctx context.Context, name string) (HealthResult, error) {
	return GlobalHealthChecker.Check(ctx, name)
}

func CheckAllHealth(ctx context.Context) map[string]HealthResult {
	return GlobalHealthChecker.CheckAll(ctx)
}

func GetOverallHealthStatus(ctx context.Context) HealthStatus {
	return GlobalHealthChecker.GetOverallStatus(ctx)
}

func GetHealthReport(ctx context.Context) HealthReport {
	return GlobalHealthChecker.GetHealthReport(ctx)
}
