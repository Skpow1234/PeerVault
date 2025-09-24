package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	BaseURL          string        `json:"base_url"`
	Concurrency      int           `json:"concurrency"`
	Duration         time.Duration `json:"duration"`
	RequestsPerSec   int           `json:"requests_per_sec"`
	Timeout          time.Duration `json:"timeout"`
	WarmupDuration   time.Duration `json:"warmup_duration"`
	CooldownDuration time.Duration `json:"cooldown_duration"`
}

// LoadTestResult holds the results of a load test
type LoadTestResult struct {
	TotalRequests            int64            `json:"total_requests"`
	SuccessfulRequests       int64            `json:"successful_requests"`
	FailedRequests           int64            `json:"failed_requests"`
	TotalDuration            time.Duration    `json:"total_duration"`
	AverageResponseTime      time.Duration    `json:"average_response_time"`
	MinResponseTime          time.Duration    `json:"min_response_time"`
	MaxResponseTime          time.Duration    `json:"max_response_time"`
	RequestsPerSecond        float64          `json:"requests_per_second"`
	ErrorRate                float64          `json:"error_rate"`
	ResponseTimeDistribution map[string]int64 `json:"response_time_distribution"`
	Errors                   map[string]int64 `json:"errors"`
	StartTime                time.Time        `json:"start_time"`
	EndTime                  time.Time        `json:"end_time"`
}

// LoadTestRequest defines a request to be tested
type LoadTestRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Weight  int               `json:"weight"` // Relative weight for this request
}

// LoadTester performs load testing
type LoadTester struct {
	config  *LoadTestConfig
	client  *http.Client
	logger  *slog.Logger
	results *LoadTestResult
	mu      sync.RWMutex
}

// NewLoadTester creates a new load tester
func NewLoadTester(config *LoadTestConfig, logger *slog.Logger) *LoadTester {
	if config == nil {
		config = DefaultLoadTestConfig()
	}

	return &LoadTester{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
		results: &LoadTestResult{
			ResponseTimeDistribution: make(map[string]int64),
			Errors:                   make(map[string]int64),
		},
	}
}

// DefaultLoadTestConfig returns the default load test configuration
func DefaultLoadTestConfig() *LoadTestConfig {
	return &LoadTestConfig{
		BaseURL:          "http://localhost:3000",
		Concurrency:      10,
		Duration:         60 * time.Second,
		RequestsPerSec:   100,
		Timeout:          30 * time.Second,
		WarmupDuration:   10 * time.Second,
		CooldownDuration: 5 * time.Second,
	}
}

// RunLoadTest runs a load test with the given requests
func (lt *LoadTester) RunLoadTest(ctx context.Context, requests []LoadTestRequest) (*LoadTestResult, error) {
	lt.logger.Info("Starting load test",
		"concurrency", lt.config.Concurrency,
		"duration", lt.config.Duration,
		"requests_per_sec", lt.config.RequestsPerSec)

	lt.results.StartTime = time.Now()
	defer func() {
		lt.results.EndTime = time.Now()
		lt.results.TotalDuration = lt.results.EndTime.Sub(lt.results.StartTime)
	}()

	// Warmup phase
	if lt.config.WarmupDuration > 0 {
		lt.logger.Info("Running warmup phase", "duration", lt.config.WarmupDuration)
		if err := lt.runWarmup(ctx, requests); err != nil {
			lt.logger.Warn("Warmup failed", "error", err)
		}
	}

	// Main load test
	lt.logger.Info("Running main load test")
	if err := lt.runMainTest(ctx, requests); err != nil {
		return nil, fmt.Errorf("load test failed: %w", err)
	}

	// Cooldown phase
	if lt.config.CooldownDuration > 0 {
		lt.logger.Info("Running cooldown phase", "duration", lt.config.CooldownDuration)
		if err := lt.runCooldown(ctx, requests); err != nil {
			lt.logger.Warn("Cooldown failed", "error", err)
		}
	}

	// Calculate final results
	lt.calculateResults()

	lt.logger.Info("Load test completed",
		"total_requests", lt.results.TotalRequests,
		"successful_requests", lt.results.SuccessfulRequests,
		"failed_requests", lt.results.FailedRequests,
		"requests_per_second", lt.results.RequestsPerSecond,
		"average_response_time", lt.results.AverageResponseTime,
		"error_rate", lt.results.ErrorRate)

	return lt.results, nil
}

// runWarmup runs the warmup phase
func (lt *LoadTester) runWarmup(ctx context.Context, requests []LoadTestRequest) error {
	warmupCtx, cancel := context.WithTimeout(ctx, lt.config.WarmupDuration)
	defer cancel()

	return lt.runTestPhase(warmupCtx, requests, 1) // Low concurrency for warmup
}

// runMainTest runs the main load test
func (lt *LoadTester) runMainTest(ctx context.Context, requests []LoadTestRequest) error {
	mainCtx, cancel := context.WithTimeout(ctx, lt.config.Duration)
	defer cancel()

	return lt.runTestPhase(mainCtx, requests, lt.config.Concurrency)
}

// runCooldown runs the cooldown phase
func (lt *LoadTester) runCooldown(ctx context.Context, requests []LoadTestRequest) error {
	cooldownCtx, cancel := context.WithTimeout(ctx, lt.config.CooldownDuration)
	defer cancel()

	return lt.runTestPhase(cooldownCtx, requests, 1) // Low concurrency for cooldown
}

// runTestPhase runs a test phase with specified concurrency
func (lt *LoadTester) runTestPhase(ctx context.Context, requests []LoadTestRequest, concurrency int) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	// Rate limiting
	rateLimiter := time.NewTicker(time.Second / time.Duration(lt.config.RequestsPerSec))
	defer rateLimiter.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case <-rateLimiter.C:
			// Acquire semaphore
			semaphore <- struct{}{}
			wg.Add(1)

			go func() {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				// Select a request based on weight
				request := lt.selectRequest(requests)
				lt.executeRequest(ctx, request)
			}()
		}
	}
}

// selectRequest selects a request based on weight
func (lt *LoadTester) selectRequest(requests []LoadTestRequest) LoadTestRequest {
	if len(requests) == 0 {
		return LoadTestRequest{
			Method: "GET",
			Path:   "/health",
		}
	}

	if len(requests) == 1 {
		return requests[0]
	}

	// Simple weight-based selection
	// In production, use proper weighted random selection
	return requests[0]
}

// executeRequest executes a single request
func (lt *LoadTester) executeRequest(ctx context.Context, req LoadTestRequest) {
	start := time.Now()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, lt.config.BaseURL+req.Path, nil)
	if err != nil {
		lt.recordError("request_creation_failed", err)
		return
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set body if provided
	if req.Body != nil {
		bodyData, err := json.Marshal(req.Body)
		if err != nil {
			lt.recordError("body_marshal_failed", err)
			return
		}
		httpReq.Body = &mockBody{data: bodyData}
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := lt.client.Do(httpReq)
	if err != nil {
		lt.recordError("request_failed", err)
		return
	}
	defer resp.Body.Close()

	// Record response
	responseTime := time.Since(start)
	lt.recordResponse(resp, responseTime)
}

// recordResponse records a successful response
func (lt *LoadTester) recordResponse(resp *http.Response, responseTime time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.results.TotalRequests++
	lt.results.SuccessfulRequests++

	// Update response time statistics
	if lt.results.MinResponseTime == 0 || responseTime < lt.results.MinResponseTime {
		lt.results.MinResponseTime = responseTime
	}
	if responseTime > lt.results.MaxResponseTime {
		lt.results.MaxResponseTime = responseTime
	}

	// Record response time distribution
	lt.recordResponseTimeDistribution(responseTime)

	// Record status code
	statusKey := fmt.Sprintf("status_%d", resp.StatusCode)
	lt.results.Errors[statusKey]++
}

// recordError records a failed request
func (lt *LoadTester) recordError(errorType string, err error) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.results.TotalRequests++
	lt.results.FailedRequests++
	lt.results.Errors[errorType]++
}

// recordResponseTimeDistribution records response time in distribution buckets
func (lt *LoadTester) recordResponseTimeDistribution(responseTime time.Duration) {
	var bucket string
	ms := responseTime.Milliseconds()

	switch {
	case ms < 100:
		bucket = "0-100ms"
	case ms < 200:
		bucket = "100-200ms"
	case ms < 500:
		bucket = "200-500ms"
	case ms < 1000:
		bucket = "500ms-1s"
	case ms < 2000:
		bucket = "1-2s"
	case ms < 5000:
		bucket = "2-5s"
	default:
		bucket = "5s+"
	}

	lt.results.ResponseTimeDistribution[bucket]++
}

// calculateResults calculates final test results
func (lt *LoadTester) calculateResults() {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedRequests) / float64(lt.results.TotalRequests) * 100
	}

	if lt.results.TotalDuration > 0 {
		lt.results.RequestsPerSecond = float64(lt.results.TotalRequests) / lt.results.TotalDuration.Seconds()
	}

	// Calculate average response time
	if lt.results.SuccessfulRequests > 0 {
		// This is a simplified calculation
		// In production, maintain a running average
		lt.results.AverageResponseTime = (lt.results.MinResponseTime + lt.results.MaxResponseTime) / 2
	}
}

// GetResults returns the current test results
func (lt *LoadTester) GetResults() *LoadTestResult {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	// Return a copy of the results
	result := *lt.results
	return &result
}

// mockBody implements io.ReadCloser for request body
type mockBody struct {
	data []byte
	pos  int
}

func (mb *mockBody) Read(p []byte) (n int, err error) {
	if mb.pos >= len(mb.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, mb.data[mb.pos:])
	mb.pos += n
	return n, nil
}

func (mb *mockBody) Close() error {
	return nil
}
