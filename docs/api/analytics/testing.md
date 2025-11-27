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

```bash
#!/bin/bash
set -e

API_BASE="http://localhost:8081/api/v1"
AUTH="Authorization: Bearer demo-token"

echo "Testing Analytics Endpoints..."

# Test summary endpoint
echo "1. Testing /analytics/summary"
curl -s -H "$AUTH" "$API_BASE/analytics/summary" | jq . > /dev/null
echo "✓ Summary endpoint working"

# Test usage metrics
echo "2. Testing /analytics/usage"
curl -s -H "$AUTH" "$API_BASE/analytics/usage?period=24h" | jq . > /dev/null
echo "✓ Usage metrics endpoint working"

# Test trends
echo "3. Testing /analytics/trends"
curl -s -H "$AUTH" "$API_BASE/analytics/trends?period=24h&interval=hour" | jq . > /dev/null
echo "✓ Trends endpoint working"

# Test popularity
echo "4. Testing /analytics/popularity"
curl -s -H "$AUTH" "$API_BASE/analytics/popularity?period=7d" | jq . > /dev/null
echo "✓ Popularity endpoint working"

# Test dashboard
echo "5. Testing /analytics/dashboard"
curl -s -H "$AUTH" "$API_BASE/analytics/dashboard?period=24h" | jq . > /dev/null
echo "✓ Dashboard endpoint working"

echo "All tests passed!"
```

### Load Testing

```bash
#!/bin/bash
# Generate test traffic for analytics

API_BASE="http://localhost:8081/api/v1"
AUTH="Authorization: Bearer demo-token"

echo "Generating test traffic..."

# Generate 100 requests
for i in {1..100}; do
  # Mix of different endpoints and methods
  case $((i % 4)) in
    0)
      curl -s -H "$AUTH" "$API_BASE/files" > /dev/null
      ;;
    1)
      curl -s -H "$AUTH" "$API_BASE/peers" > /dev/null
      ;;
    2)
      curl -s -X POST -H "$AUTH" -H "Content-Type: application/json" \
        -d '{"name":"test"}' "$API_BASE/files" > /dev/null
      ;;
    3)
      curl -s -H "$AUTH" "/health" > /dev/null
      ;;
  esac

  # Random delay
  sleep 0.$((RANDOM % 10))
done

echo "Traffic generated. Check analytics:"
curl -s -H "$AUTH" "$API_BASE/analytics/summary" | \
  jq '{requests: .overall_metrics.total_requests, error_rate: .overall_metrics.error_rate}'
```

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

```bash
#!/bin/bash
# Stress test analytics with concurrent requests

API_BASE="http://localhost:8081/api/v1"
AUTH="Authorization: Bearer demo-token"
CONCURRENT=10
REQUESTS_PER_WORKER=100

echo "Running stress test with $CONCURRENT workers..."

stress_worker() {
  local worker_id=$1
  for i in $(seq 1 $REQUESTS_PER_WORKER); do
    curl -s -H "$AUTH" "$API_BASE/files" > /dev/null
  done
  echo "Worker $worker_id completed"
}

# Start workers
for i in $(seq 1 $CONCURRENT); do
  stress_worker $i &
done

# Wait for all workers
wait

echo "Stress test completed. Total requests: $((CONCURRENT * REQUESTS_PER_WORKER))"
echo "Checking analytics..."

curl -s -H "$AUTH" "$API_BASE/analytics/summary" | \
  jq '{
    requests: .overall_metrics.total_requests,
    error_rate: .overall_metrics.error_rate,
    avg_latency: (.overall_metrics.average_duration / 1000000)
  }'
```

## Validation Testing

### Verify Data Accuracy

```bash
#!/bin/bash
API_BASE="http://localhost:8081/api/v1"
AUTH="Authorization: Bearer demo-token"

echo "Validating analytics accuracy..."

# Make known number of requests
EXPECTED_REQUESTS=50
echo "Making $EXPECTED_REQUESTS requests..."

for i in $(seq 1 $EXPECTED_REQUESTS); do
  curl -s -H "$AUTH" "$API_BASE/files" > /dev/null
done

# Wait for processing
sleep 2

# Check analytics
ACTUAL=$(curl -s -H "$AUTH" "$API_BASE/analytics/usage?period=hour" | \
  jq '.total_requests')

echo "Expected: $EXPECTED_REQUESTS"
echo "Actual: $ACTUAL"

if [ "$ACTUAL" -ge "$EXPECTED_REQUESTS" ]; then
  echo "✓ Validation passed"
else
  echo "✗ Validation failed: expected >= $EXPECTED_REQUESTS, got $ACTUAL"
  exit 1
fi
```

### Test Error Tracking

```bash
#!/bin/bash
API_BASE="http://localhost:8081/api/v1"
AUTH="Authorization: Bearer demo-token"

echo "Testing error tracking..."

# Generate some errors (invalid requests)
for i in {1..10}; do
  curl -s -H "$AUTH" "$API_BASE/nonexistent" > /dev/null
done

# Check error rate
ERROR_RATE=$(curl -s -H "$AUTH" "$API_BASE/analytics/usage?period=hour" | \
  jq '.error_rate')

echo "Error rate: $ERROR_RATE%"

if (( $(echo "$ERROR_RATE > 0" | bc -l) )); then
  echo "✓ Error tracking working"
else
  echo "✗ No errors recorded"
fi
```

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

```bash
#!/bin/bash
# Run this with cron for continuous monitoring

API_BASE="http://localhost:8081/api/v1"
AUTH="Authorization: Bearer demo-token"
ALERT_THRESHOLD=10.0

check_health() {
  local summary=$(curl -s -H "$AUTH" "$API_BASE/analytics/summary")
  
  local status=$(echo $summary | jq -r '.system_health.status')
  local error_rate=$(echo $summary | jq -r '.system_health.error_rate')
  
  echo "[$(date)] Status: $status, Error Rate: $error_rate%"
  
  if [ "$status" != "healthy" ]; then
    echo "ALERT: System status is $status"
    # Send alert
  fi
  
  if (( $(echo "$error_rate > $ALERT_THRESHOLD" | bc -l) )); then
    echo "ALERT: Error rate $error_rate% exceeds threshold $ALERT_THRESHOLD%"
    # Send alert
  fi
}

check_health
```

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

```bash
# Run Go unit tests
go test ./internal/api/analytics/... -v

# Run integration tests
./scripts/test-analytics-integration.sh

# Run performance benchmarks
go test ./internal/api/analytics/... -bench=. -benchmem

# Generate coverage report
go test ./internal/api/analytics/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

