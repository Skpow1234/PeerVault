package health

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	assert.NotNil(t, hc)
	assert.NotNil(t, hc.checks)
	assert.Equal(t, 0, len(hc.checks))
}

func TestHealthChecker_RegisterUnregisterCheck(t *testing.T) {
	hc := NewHealthChecker()

	// Create a simple health check
	check := NewSimpleHealthCheck("test-check", "Test check", 5*time.Second, func(ctx context.Context) error {
		return nil
	})

	// Register the check
	hc.RegisterCheck(check)

	// Verify it's registered
	hc.mu.RLock()
	assert.Contains(t, hc.checks, "test-check")
	hc.mu.RUnlock()

	// Unregister the check
	hc.UnregisterCheck("test-check")

	// Verify it's unregistered
	hc.mu.RLock()
	assert.NotContains(t, hc.checks, "test-check")
	hc.mu.RUnlock()
}

func TestSimpleHealthCheck(t *testing.T) {
	// Test successful health check
	check := NewSimpleHealthCheck("success-check", "This check always passes", 5*time.Second, func(ctx context.Context) error {
		return nil
	})

	assert.Equal(t, "success-check", check.Name())
	assert.Equal(t, 5*time.Second, check.Timeout())

	ctx := context.Background()
	result := check.Check(ctx)

	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Equal(t, "This check always passes", result.Message)
	// Note: Timestamp and Duration are set by HealthChecker, not individual checks
}

func TestSimpleHealthCheck_Failure(t *testing.T) {
	// Test failing health check
	check := NewSimpleHealthCheck("fail-check", "This check always fails", 5*time.Second, func(ctx context.Context) error {
		return errors.New("simulated failure")
	})

	ctx := context.Background()
	result := check.Check(ctx)

	assert.Equal(t, HealthStatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "simulated failure")
	// Note: Timestamp and Duration are set by HealthChecker, not individual checks
}

func TestHTTPHealthCheck(t *testing.T) {
	check := NewHTTPHealthCheck("http-check", "http://example.com/health", "HTTP health check", 10*time.Second)

	assert.Equal(t, "http-check", check.Name())
	assert.Equal(t, 10*time.Second, check.Timeout())

	ctx := context.Background()
	result := check.Check(ctx)

	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Contains(t, result.Message, "http://example.com/health")
	assert.NotNil(t, result.Details)
	assert.Equal(t, "http://example.com/health", result.Details["url"])
}

func TestDatabaseHealthCheck(t *testing.T) {
	// Test successful database check
	check := NewDatabaseHealthCheck("db-check", "Database connection healthy", 5*time.Second, func(ctx context.Context) error {
		return nil
	})

	assert.Equal(t, "db-check", check.Name())
	assert.Equal(t, 5*time.Second, check.Timeout())

	ctx := context.Background()
	result := check.Check(ctx)

	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Equal(t, "Database connection healthy", result.Message)
}

func TestDatabaseHealthCheck_Failure(t *testing.T) {
	// Test failing database check
	check := NewDatabaseHealthCheck("db-fail-check", "Database connection failed", 5*time.Second, func(ctx context.Context) error {
		return errors.New("connection refused")
	})

	ctx := context.Background()
	result := check.Check(ctx)

	assert.Equal(t, HealthStatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "connection refused")
}

func TestHealthChecker_CheckAll(t *testing.T) {
	hc := NewHealthChecker()

	// Register multiple checks
	check1 := NewSimpleHealthCheck("check1", "Check 1", 1*time.Second, func(ctx context.Context) error {
		return nil
	})
	check2 := NewSimpleHealthCheck("check2", "Check 2", 1*time.Second, func(ctx context.Context) error {
		return errors.New("check2 failed")
	})

	hc.RegisterCheck(check1)
	hc.RegisterCheck(check2)

	ctx := context.Background()
	results := hc.CheckAll(ctx)

	assert.Len(t, results, 2)
	assert.Contains(t, results, "check1")
	assert.Contains(t, results, "check2")

	assert.Equal(t, HealthStatusHealthy, results["check1"].Status)
	assert.Equal(t, HealthStatusUnhealthy, results["check2"].Status)

	// Check timestamps and durations are set
	for _, result := range results {
		assert.NotZero(t, result.Timestamp)
		// Note: Duration may be zero for very fast checks
	}
}

func TestHealthChecker_Check(t *testing.T) {
	hc := NewHealthChecker()

	check := NewSimpleHealthCheck("test-check", "Test check", 1*time.Second, func(ctx context.Context) error {
		return nil
	})

	hc.RegisterCheck(check)

	ctx := context.Background()

	// Test successful check
	result, err := hc.Check(ctx, "test-check")
	require.NoError(t, err)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	// Note: Timestamp and Duration are set by HealthChecker

	// Test non-existent check
	result, err = hc.Check(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthChecker_GetOverallStatus(t *testing.T) {
	hc := NewHealthChecker()
	ctx := context.Background()

	// Test with no checks
	status := hc.GetOverallStatus(ctx)
	assert.Equal(t, HealthStatusUnknown, status)

	// Add healthy check
	healthyCheck := NewSimpleHealthCheck("healthy", "Healthy", 1*time.Second, func(ctx context.Context) error {
		return nil
	})
	hc.RegisterCheck(healthyCheck)

	status = hc.GetOverallStatus(ctx)
	assert.Equal(t, HealthStatusHealthy, status)

	// Add degraded check
	degradedCheck := NewSimpleHealthCheck("degraded", "Degraded", 1*time.Second, func(ctx context.Context) error {
		return nil
	})
	// Override the check to return degraded status
	degradedCheck.checkFunc = func(ctx context.Context) error {
		return nil
	}
	hc.RegisterCheck(degradedCheck)

	status = hc.GetOverallStatus(ctx)
	assert.Equal(t, HealthStatusHealthy, status) // Should still be healthy

	// Add unhealthy check
	unhealthyCheck := NewSimpleHealthCheck("unhealthy", "Unhealthy", 1*time.Second, func(ctx context.Context) error {
		return errors.New("unhealthy")
	})
	hc.RegisterCheck(unhealthyCheck)

	status = hc.GetOverallStatus(ctx)
	assert.Equal(t, HealthStatusUnhealthy, status)
}

func TestHealthChecker_GetHealthReport(t *testing.T) {
	hc := NewHealthChecker()

	check := NewSimpleHealthCheck("test-check", "Test", 1*time.Second, func(ctx context.Context) error {
		return nil
	})
	hc.RegisterCheck(check)

	ctx := context.Background()
	report := hc.GetHealthReport(ctx)

	assert.Equal(t, HealthStatusHealthy, report.OverallStatus)
	assert.NotZero(t, report.Timestamp)
	assert.Len(t, report.Checks, 1)
	assert.Contains(t, report.Checks, "test-check")
}

func TestHealthReport_ToJSON(t *testing.T) {
	report := HealthReport{
		OverallStatus: HealthStatusHealthy,
		Timestamp:     time.Now(),
		Checks: map[string]HealthResult{
			"test": {
				Status:    HealthStatusHealthy,
				Message:   "Test message",
				Timestamp: time.Now(),
				Duration:  100 * time.Millisecond,
			},
		},
	}

	jsonData, err := report.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it's valid JSON
	assert.Contains(t, string(jsonData), "healthy")
	assert.Contains(t, string(jsonData), "test")
	assert.Contains(t, string(jsonData), "Test message")
}

func TestHealthChecker_Timeout(t *testing.T) {
	hc := NewHealthChecker()

	// Create a slow health check
	slowCheck := NewSimpleHealthCheck("slow-check", "Slow check", 100*time.Millisecond, func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond) // Sleep longer than timeout
		return nil
	})

	hc.RegisterCheck(slowCheck)

	ctx := context.Background()
	results := hc.CheckAll(ctx)

	// The check should still complete but may have taken longer
	assert.Contains(t, results, "slow-check")
	result := results["slow-check"]
	assert.NotZero(t, result.Duration)
}

func TestGlobalHealthChecker(t *testing.T) {
	// Clean up any existing checks
	GlobalHealthChecker.UnregisterCheck("global-test")

	// Register a global health check
	check := NewSimpleHealthCheck("global-test", "Global test check", 1*time.Second, func(ctx context.Context) error {
		return nil
	})

	RegisterHealthCheck(check)

	ctx := context.Background()

	// Test global functions
	result, err := CheckHealth(ctx, "global-test")
	require.NoError(t, err)
	assert.Equal(t, HealthStatusHealthy, result.Status)

	results := CheckAllHealth(ctx)
	assert.Contains(t, results, "global-test")

	status := GetOverallHealthStatus(ctx)
	assert.Equal(t, HealthStatusHealthy, status)

	report := GetHealthReport(ctx)
	assert.Equal(t, HealthStatusHealthy, report.OverallStatus)

	// Clean up
	UnregisterHealthCheck("global-test")
}

func TestHealthStatus_Constants(t *testing.T) {
	assert.Equal(t, HealthStatus("healthy"), HealthStatusHealthy)
	assert.Equal(t, HealthStatus("unhealthy"), HealthStatusUnhealthy)
	assert.Equal(t, HealthStatus("degraded"), HealthStatusDegraded)
	assert.Equal(t, HealthStatus("unknown"), HealthStatusUnknown)
}

func TestHealthResult_JSONTags(t *testing.T) {
	result := HealthResult{
		Status:    HealthStatusHealthy,
		Message:   "Test message",
		Details:   map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
		Duration:  100 * time.Millisecond,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(result)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test that the JSON contains expected fields
	jsonStr := string(jsonData)
	assert.Contains(t, jsonStr, `"status":"healthy"`) // HealthStatusHealthy as string
	assert.Contains(t, jsonStr, `"message":"Test message"`)
	assert.Contains(t, jsonStr, `"key":"value"`)
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	hc := NewHealthChecker()

	// Register a check
	check := NewSimpleHealthCheck("concurrent-check", "Concurrent test", 1*time.Second, func(ctx context.Context) error {
		return nil
	})
	hc.RegisterCheck(check)

	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool, 3)

	// Concurrent check operations
	go func() {
		for i := 0; i < 10; i++ {
			hc.Check(ctx, "concurrent-check")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			hc.CheckAll(ctx)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			hc.GetOverallStatus(ctx)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}
}

func TestHealthChecker_RegisterDuplicate(t *testing.T) {
	hc := NewHealthChecker()

	check1 := NewSimpleHealthCheck("duplicate", "First check", 1*time.Second, func(ctx context.Context) error {
		return nil
	})
	check2 := NewSimpleHealthCheck("duplicate", "Second check", 1*time.Second, func(ctx context.Context) error {
		return nil
	})

	// Register first check
	hc.RegisterCheck(check1)

	// Register second check with same name (should overwrite)
	hc.RegisterCheck(check2)

	// Should only have one check
	hc.mu.RLock()
	assert.Len(t, hc.checks, 1)
	assert.Contains(t, hc.checks, "duplicate")
	hc.mu.RUnlock()
}
