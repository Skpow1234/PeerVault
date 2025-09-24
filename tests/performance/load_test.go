package performance

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/performance"
)

func TestLoadTest(t *testing.T) {
	config := &performance.LoadTestConfig{
		BaseURL:          "http://localhost:3000",
		Concurrency:      5,
		Duration:         10 * time.Second,
		RequestsPerSec:   10,
		Timeout:          30 * time.Second,
		WarmupDuration:   2 * time.Second,
		CooldownDuration: 1 * time.Second,
	}

	logger := slog.Default()
	tester := performance.NewLoadTester(config, logger)

	requests := []performance.LoadTestRequest{
		{
			Method: "GET",
			Path:   "/health",
			Weight: 1,
		},
		{
			Method: "GET",
			Path:   "/api/status",
			Weight: 1,
		},
	}

	ctx := context.Background()
	result, err := tester.RunLoadTest(ctx, requests)
	if err != nil {
		t.Fatalf("Load test failed: %v", err)
	}

	// Basic assertions
	if result.TotalRequests == 0 {
		t.Error("Expected at least one request")
	}

	if result.RequestsPerSecond <= 0 {
		t.Error("Expected positive requests per second")
	}

	t.Logf("Load test results:")
	t.Logf("  Total requests: %d", result.TotalRequests)
	t.Logf("  Successful requests: %d", result.SuccessfulRequests)
	t.Logf("  Failed requests: %d", result.FailedRequests)
	t.Logf("  Requests per second: %.2f", result.RequestsPerSecond)
	t.Logf("  Average response time: %v", result.AverageResponseTime)
	t.Logf("  Error rate: %.2f%%", result.ErrorRate)
}

func TestLoadTestWithMockServer(t *testing.T) {
	// This test would use a mock server for more reliable testing
	// For now, we'll skip if no server is available
	t.Skip("Skipping load test with mock server - requires running server")
}
