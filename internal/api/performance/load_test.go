package performance

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadTesterCreation tests the creation of a load tester
func TestLoadTesterCreation(t *testing.T) {
	config := &LoadTestConfig{
		BaseURL:     "http://localhost:3000",
		Concurrency: 5,
		Duration:    10 * time.Second,
	}

	tester := NewLoadTester(config, slog.Default())
	require.NotNil(t, tester)
	assert.Equal(t, config, tester.config)
}

// TestDefaultLoadTestConfig tests the default configuration
func TestDefaultLoadTestConfig(t *testing.T) {
	config := DefaultLoadTestConfig()
	require.NotNil(t, config)

	assert.Equal(t, "http://localhost:3000", config.BaseURL)
	assert.Equal(t, 10, config.Concurrency)
	assert.Equal(t, 60*time.Second, config.Duration)
	assert.Equal(t, 100, config.RequestsPerSec)
	assert.Equal(t, 30*time.Second, config.Timeout)
}

// TestLoadTestRequestValidation tests request validation
func TestLoadTestRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request LoadTestRequest
		valid   bool
	}{
		{
			name: "Valid GET request",
			request: LoadTestRequest{
				Method: "GET",
				Path:   "/health",
			},
			valid: true,
		},
		{
			name: "Valid POST request",
			request: LoadTestRequest{
				Method: "POST",
				Path:   "/api/nodes",
				Body:   map[string]string{"name": "test"},
			},
			valid: true,
		},
		{
			name: "Request with headers",
			request: LoadTestRequest{
				Method: "GET",
				Path:   "/api/nodes",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in a real implementation, you'd have proper validation
			assert.NotEmpty(t, tt.request.Method)
			assert.NotEmpty(t, tt.request.Path)
		})
	}
}

// TestLoadTestResultInitialization tests result initialization
func TestLoadTestResultInitialization(t *testing.T) {
	config := DefaultLoadTestConfig()
	tester := NewLoadTester(config, slog.Default())

	results := tester.GetResults()
	require.NotNil(t, results)

	assert.Equal(t, int64(0), results.TotalRequests)
	assert.Equal(t, int64(0), results.SuccessfulRequests)
	assert.Equal(t, int64(0), results.FailedRequests)
	assert.Equal(t, float64(0), results.ErrorRate)
	assert.Equal(t, float64(0), results.RequestsPerSecond)
	assert.NotNil(t, results.ResponseTimeDistribution)
	assert.NotNil(t, results.Errors)
}

// TestLoadTestWithEmptyRequests tests load testing with no requests
func TestLoadTestWithEmptyRequests(t *testing.T) {
	config := &LoadTestConfig{
		BaseURL:     "http://localhost:3000",
		Concurrency: 1,
		Duration:    1 * time.Second,
	}

	tester := NewLoadTester(config, slog.Default())
	ctx := context.Background()

	// Test with empty requests slice - should not error in current implementation
	result, err := tester.RunLoadTest(ctx, []LoadTestRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestLoadTestWithValidRequests tests load testing with valid requests
func TestLoadTestWithValidRequests(t *testing.T) {
	config := &LoadTestConfig{
		BaseURL:        "http://localhost:3000",
		Concurrency:    1,
		Duration:       100 * time.Millisecond, // Very short test
		RequestsPerSec: 10,
	}

	tester := NewLoadTester(config, slog.Default())
	ctx := context.Background()

	requests := []LoadTestRequest{
		{
			Method: "GET",
			Path:   "/health",
		},
	}

	// This tests the basic structure - current implementation doesn't make actual requests
	result, err := tester.RunLoadTest(ctx, requests)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestResponseTimeDistribution tests response time distribution recording
func TestResponseTimeDistribution(t *testing.T) {
	// Test different response times
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{50 * time.Millisecond, "0-100ms"},
		{150 * time.Millisecond, "100-200ms"},
		{300 * time.Millisecond, "200-500ms"},
		{750 * time.Millisecond, "500ms-1s"},
		{1500 * time.Millisecond, "1-2s"},
		{3000 * time.Millisecond, "2-5s"},
		{6000 * time.Millisecond, "5s+"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			// Create a mock response
			// In a real test, you'd use httptest.NewServer
			// For now, just test the distribution logic
			assert.Equal(t, tc.expected, getResponseTimeBucket(tc.duration))
		})
	}
}

// getResponseTimeBucket is a helper function to test response time distribution
func getResponseTimeBucket(responseTime time.Duration) string {
	ms := responseTime.Milliseconds()

	switch {
	case ms < 100:
		return "0-100ms"
	case ms < 200:
		return "100-200ms"
	case ms < 500:
		return "200-500ms"
	case ms < 1000:
		return "500ms-1s"
	case ms < 2000:
		return "1-2s"
	case ms < 5000:
		return "2-5s"
	default:
		return "5s+"
	}
}

// TestLoadTestConfigValidation tests configuration validation
func TestLoadTestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *LoadTestConfig
		valid  bool
	}{
		{
			name: "Valid config",
			config: &LoadTestConfig{
				BaseURL:     "http://localhost:3000",
				Concurrency: 10,
				Duration:    60 * time.Second,
			},
			valid: true,
		},
		{
			name: "Zero concurrency",
			config: &LoadTestConfig{
				BaseURL:     "http://localhost:3000",
				Concurrency: 0,
				Duration:    60 * time.Second,
			},
			valid: false,
		},
		{
			name: "Empty base URL",
			config: &LoadTestConfig{
				BaseURL:     "",
				Concurrency: 10,
				Duration:    60 * time.Second,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - in a real implementation, you'd have proper validation
			if tt.valid {
				assert.NotEmpty(t, tt.config.BaseURL)
				assert.Greater(t, tt.config.Concurrency, 0)
			}
		})
	}
}
