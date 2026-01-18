# Testing API Analytics

This guide covers testing the API Analytics & Monitoring system.

## Unit Testing

### Test Analytics Storage

```go
package analytics_test

import (
	"testing"
	"time"

	"github.com/Skpow1234/Peervault/internal/api/analytics"
)

func TestMemoryStorage(t *testing.T) {
	storage := analytics.NewMemoryStorage(1000, 30)

	// Test recording API calls
	call := &analytics.APICall{
		ID:         "test-1",
		Timestamp:  time.Now(),
		Method:     "GET",
		Path:       "/api/v1/files",
		StatusCode: 200,
		Duration:   50 * time.Millisecond,
		Endpoint:   "/api/v1/files",
	}

	err := storage.RecordAPICall(call)
	if err != nil {
		t.Fatalf("Failed to record API call: %v", err)
	}

	// Test querying calls
	query := &analytics.AnalyticsQuery{
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now(),
	}

	calls, err := storage.GetAPICalls(query)
	if err != nil {
		t.Fatalf("Failed to get API calls: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}
}

func TestUsageMetrics(t *testing.T) {
	storage := analytics.NewMemoryStorage(1000, 30)
	now := time.Now()

	// Record multiple calls
	for i := 0; i < 100; i++ {
		call := &analytics.APICall{
			ID:         fmt.Sprintf("test-%d", i),
			Timestamp:  now.Add(time.Duration(i) * time.Minute),
			Method:     "GET",
			Path:       "/api/v1/files",
			StatusCode: 200,
			Duration:   time.Duration(40+i) * time.Millisecond,
			Endpoint:   "/api/v1/files",
		}
		storage.RecordAPICall(call)
	}

	// Get metrics
	metrics, err := storage.GetUsageMetrics(now.Add(-2*time.Hour), now.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("Failed to get usage metrics: %v", err)
	}

	if metrics.TotalRequests != 100 {
		t.Errorf("Expected 100 requests, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 100 {
		t.Errorf("Expected 100 successful requests, got %d", metrics.SuccessfulRequests)
	}
}
```

### Test Analytics Service

```go
func TestAnalyticsService(t *testing.T) {
	storage := analytics.NewMemoryStorage(1000, 30)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := analytics.DefaultConfig()

	service := analytics.NewService(storage, logger, config)

	// Test recording
	call := &analytics.APICall{
		Timestamp:  time.Now(),
		Method:     "POST",
		Path:       "/api/v1/files",
		StatusCode: 201,
		Duration:   60 * time.Millisecond,
	}

	err := service.RecordAPICall(call)
	if err != nil {
		t.Fatalf("Failed to record API call: %v", err)
	}

	// Verify ID was generated
	if call.ID == "" {
		t.Error("Expected ID to be generated")
	}

	// Verify endpoint was normalized
	if call.Endpoint == "" {
		t.Error("Expected endpoint to be normalized")
	}
}
```

### Test Middleware

```go
func TestAnalyticsMiddleware(t *testing.T) {
	storage := analytics.NewMemoryStorage(1000, 30)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := analytics.NewService(storage, logger, nil)

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with middleware
	middleware := analytics.Middleware(service, logger)
	wrappedHandler := middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("X-User-ID", "test-user")
	rec := httptest.NewRecorder()

	// Execute request
	wrappedHandler.ServeHTTP(rec, req)

	// Wait for async recording
	time.Sleep(100 * time.Millisecond)

	// Verify recording
	calls, err := storage.GetAPICalls(&analytics.AnalyticsQuery{
		StartTime: time.Now().Add(-1 * time.Minute),
		EndTime:   time.Now(),
	})

	if err != nil {
		t.Fatalf("Failed to get calls: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].UserID != "test-user" {
		t.Errorf("Expected user ID 'test-user', got '%s'", calls[0].UserID)
	}
}
```

## Integration Testing

### Test Analytics Endpoints

Integration endpoint checks run in CI using containerized workflows.

### Load Testing

Load testing runs in CI using containerized workflows.

## Performance Testing

### Test Storage Capacity

```go
func BenchmarkRecordAPICall(b *testing.B) {
	storage := analytics.NewMemoryStorage(100000, 30)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		call := &analytics.APICall{
			ID:         fmt.Sprintf("bench-%d", i),
			Timestamp:  time.Now(),
			Method:     "GET",
			Path:       "/api/v1/test",
			StatusCode: 200,
			Duration:   50 * time.Millisecond,
		}
		storage.RecordAPICall(call)
	}
}

func BenchmarkGetUsageMetrics(b *testing.B) {
	storage := analytics.NewMemoryStorage(100000, 30)
	now := time.Now()

	// Pre-populate with data
	for i := 0; i < 10000; i++ {
		call := &analytics.APICall{
			ID:         fmt.Sprintf("bench-%d", i),
			Timestamp:  now.Add(time.Duration(i) * time.Second),
			Method:     "GET",
			Path:       "/api/v1/test",
			StatusCode: 200,
			Duration:   50 * time.Millisecond,
		}
		storage.RecordAPICall(call)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.GetUsageMetrics(now.Add(-1*time.Hour), now)
	}
}
```

### Stress Test

Stress tests run in CI using containerized workflows.

## Validation Testing

### Verify Data Accuracy

Validation checks run in CI using containerized workflows.

### Test Error Tracking

Error tracking checks run in CI using containerized workflows.

## Cleanup Testing

### Test Data Retention

```go
func TestDataCleanup(t *testing.T) {
	storage := analytics.NewMemoryStorage(1000, 1) // 1 day retention
	now := time.Now()

	// Add old data
	oldCall := &analytics.APICall{
		ID:         "old-1",
		Timestamp:  now.Add(-48 * time.Hour),
		Method:     "GET",
		Path:       "/test",
		StatusCode: 200,
	}
	storage.RecordAPICall(oldCall)

	// Add recent data
	recentCall := &analytics.APICall{
		ID:         "recent-1",
		Timestamp:  now,
		Method:     "GET",
		Path:       "/test",
		StatusCode: 200,
	}
	storage.RecordAPICall(recentCall)

	// Cleanup old data
	cutoff := now.Add(-24 * time.Hour)
	err := storage.CleanupOldData(cutoff)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify only recent data remains
	calls, err := storage.GetAPICalls(&analytics.AnalyticsQuery{
		StartTime: now.Add(-72 * time.Hour),
		EndTime:   now,
	})

	if err != nil {
		t.Fatalf("Failed to get calls: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("Expected 1 call after cleanup, got %d", len(calls))
	}

	if calls[0].ID != "recent-1" {
		t.Errorf("Wrong call retained after cleanup")
	}
}
```

## Continuous Monitoring

### Automated Health Check

Automated monitoring runs in CI or dedicated containerized jobs.

## Test Checklist

- [ ] Unit tests for storage operations
- [ ] Unit tests for service layer
- [ ] Unit tests for middleware
- [ ] Integration tests for all endpoints
- [ ] Load testing with realistic traffic
- [ ] Performance benchmarks
- [ ] Stress testing with concurrent requests
- [ ] Data accuracy validation
- [ ] Error tracking verification
- [ ] Cleanup and retention testing
- [ ] Memory usage profiling
- [ ] Response time monitoring

## Running All Tests

Analytics tests run in CI using containerized workflows.

