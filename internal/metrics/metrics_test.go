package metrics

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCounter(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter", map[string]string{"label": "value"})

	assert.Equal(t, "test_counter", counter.name)
	assert.Equal(t, "Test counter", counter.description)
	assert.Equal(t, map[string]string{"label": "value"}, counter.labels)
	assert.Equal(t, int64(0), counter.Get())
}

func TestCounter_Inc(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter", nil)

	counter.Inc()
	assert.Equal(t, int64(1), counter.Get())

	counter.Inc()
	assert.Equal(t, int64(2), counter.Get())
}

func TestCounter_Add(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter", nil)

	counter.Add(5)
	assert.Equal(t, int64(5), counter.Get())

	counter.Add(10)
	assert.Equal(t, int64(15), counter.Get())

	counter.Add(-3)
	assert.Equal(t, int64(12), counter.Get())
}

func TestCounter_ToMetric(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter", map[string]string{"env": "test"})
	counter.Add(42)

	metric := counter.ToMetric()

	assert.Equal(t, "test_counter", metric.Name)
	assert.Equal(t, MetricTypeCounter, metric.Type)
	assert.Equal(t, float64(42), metric.Value)
	assert.Equal(t, "Test counter", metric.Description)
	assert.Equal(t, map[string]string{"env": "test"}, metric.Labels)
	assert.NotZero(t, metric.Timestamp)
}

func TestNewGauge(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge", map[string]string{"unit": "bytes"})

	assert.Equal(t, "test_gauge", gauge.name)
	assert.Equal(t, "Test gauge", gauge.description)
	assert.Equal(t, map[string]string{"unit": "bytes"}, gauge.labels)
	assert.Equal(t, int64(0), gauge.Get())
}

func TestGauge_Set(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge", nil)

	gauge.Set(100)
	assert.Equal(t, int64(100), gauge.Get())

	gauge.Set(50)
	assert.Equal(t, int64(50), gauge.Get())
}

func TestGauge_IncDec(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge", nil)

	gauge.Inc()
	assert.Equal(t, int64(1), gauge.Get())

	gauge.Dec()
	assert.Equal(t, int64(0), gauge.Get())

	gauge.Inc()
	gauge.Inc()
	assert.Equal(t, int64(2), gauge.Get())
}

func TestGauge_AddSub(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge", nil)

	gauge.Add(10)
	assert.Equal(t, int64(10), gauge.Get())

	gauge.Sub(5)
	assert.Equal(t, int64(5), gauge.Get())

	gauge.Add(-3)
	assert.Equal(t, int64(2), gauge.Get())
}

func TestGauge_ToMetric(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge", map[string]string{"type": "memory"})
	gauge.Set(1024)

	metric := gauge.ToMetric()

	assert.Equal(t, "test_gauge", metric.Name)
	assert.Equal(t, MetricTypeGauge, metric.Type)
	assert.Equal(t, float64(1024), metric.Value)
	assert.Equal(t, "Test gauge", metric.Description)
	assert.Equal(t, map[string]string{"type": "memory"}, metric.Labels)
	assert.NotZero(t, metric.Timestamp)
}

func TestNewHistogram(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 5.0}
	histogram := NewHistogram("test_histogram", "Test histogram", buckets, map[string]string{"method": "GET"})

	assert.Equal(t, "test_histogram", histogram.name)
	assert.Equal(t, "Test histogram", histogram.description)
	assert.Equal(t, buckets, histogram.buckets)
	assert.Equal(t, map[string]string{"method": "GET"}, histogram.labels)
	assert.Len(t, histogram.counts, len(buckets)+1) // +1 for +Inf bucket
	assert.Equal(t, int64(0), histogram.GetSum())
	assert.Equal(t, int64(0), histogram.GetTotalCount())
}

func TestHistogram_Observe(t *testing.T) {
	buckets := []float64{1.0, 5.0, 10.0}
	histogram := NewHistogram("test_histogram", "Test histogram", buckets, nil)

	// Observe values
	histogram.Observe(0.5)  // Should go to bucket 0 (≤ 1.0)
	histogram.Observe(3.0)  // Should go to bucket 1 (≤ 5.0)
	histogram.Observe(7.0)  // Should go to bucket 2 (≤ 10.0)
	histogram.Observe(15.0) // Should go to bucket 3 (+Inf)

	assert.Equal(t, int64(1), histogram.GetCount(0))
	assert.Equal(t, int64(1), histogram.GetCount(1))
	assert.Equal(t, int64(1), histogram.GetCount(2))
	assert.Equal(t, int64(1), histogram.GetCount(3))
	assert.Equal(t, int64(4), histogram.GetTotalCount())
	assert.Equal(t, int64(25), histogram.GetSum()) // 0.5 + 3.0 + 7.0 + 15.0
}

func TestHistogram_GetCount_InvalidIndex(t *testing.T) {
	buckets := []float64{1.0, 5.0}
	histogram := NewHistogram("test_histogram", "Test histogram", buckets, nil)

	assert.Equal(t, int64(0), histogram.GetCount(-1))
	assert.Equal(t, int64(0), histogram.GetCount(10)) // Out of bounds
}

func TestHistogram_ToMetric(t *testing.T) {
	buckets := []float64{1.0, 5.0}
	histogram := NewHistogram("test_histogram", "Test histogram", buckets, map[string]string{"service": "api"})
	histogram.Observe(2.0)
	histogram.Observe(8.0)

	metric := histogram.ToMetric()

	assert.Equal(t, "test_histogram", metric.Name)
	assert.Equal(t, MetricTypeHistogram, metric.Type)
	assert.Equal(t, float64(2), metric.Value) // Total count
	assert.Equal(t, "Test histogram", metric.Description)
	assert.Equal(t, map[string]string{"service": "api"}, metric.Labels)
	assert.NotZero(t, metric.Timestamp)
}

func TestNewMetricsRegistry(t *testing.T) {
	registry := NewMetricsRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.metrics)
	assert.Len(t, registry.metrics, 0)
}

func TestMetricsRegistry_RegisterCounter(t *testing.T) {
	registry := NewMetricsRegistry()

	counter := registry.RegisterCounter("requests_total", "Total requests", map[string]string{"method": "GET"})

	assert.NotNil(t, counter)
	assert.Equal(t, "requests_total", counter.name)

	// Verify it's in the registry
	retrieved, exists := registry.GetCounter("requests_total")
	assert.True(t, exists)
	assert.Equal(t, counter, retrieved)
}

func TestMetricsRegistry_RegisterGauge(t *testing.T) {
	registry := NewMetricsRegistry()

	gauge := registry.RegisterGauge("memory_usage", "Memory usage", map[string]string{"unit": "bytes"})

	assert.NotNil(t, gauge)
	assert.Equal(t, "memory_usage", gauge.name)

	// Verify it's in the registry
	retrieved, exists := registry.GetGauge("memory_usage")
	assert.True(t, exists)
	assert.Equal(t, gauge, retrieved)
}

func TestMetricsRegistry_RegisterHistogram(t *testing.T) {
	registry := NewMetricsRegistry()
	buckets := []float64{0.1, 1.0, 10.0}

	histogram := registry.RegisterHistogram("response_time", "Response time", buckets, map[string]string{"endpoint": "/api"})

	assert.NotNil(t, histogram)
	assert.Equal(t, "response_time", histogram.name)

	// Verify it's in the registry
	retrieved, exists := registry.GetHistogram("response_time")
	assert.True(t, exists)
	assert.Equal(t, histogram, retrieved)
}

func TestMetricsRegistry_GetNonExistent(t *testing.T) {
	registry := NewMetricsRegistry()

	counter, exists := registry.GetCounter("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, counter)

	gauge, exists := registry.GetGauge("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, gauge)

	histogram, exists := registry.GetHistogram("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, histogram)
}

func TestMetricsRegistry_GetAllMetrics(t *testing.T) {
	registry := NewMetricsRegistry()

	// Register different types of metrics
	counter := registry.RegisterCounter("requests", "Request count", nil)
	gauge := registry.RegisterGauge("active_users", "Active users", nil)
	histogram := registry.RegisterHistogram("latency", "Request latency", []float64{0.1, 1.0}, nil)

	// Modify values
	counter.Add(5)
	gauge.Set(10)
	histogram.Observe(0.5)

	// Get all metrics
	metrics := registry.GetAllMetrics()
	assert.Len(t, metrics, 3)

	// Verify each metric
	foundCounter, foundGauge, foundHistogram := false, false, false
	for _, metric := range metrics {
		switch metric.Name {
		case "requests":
			assert.Equal(t, MetricTypeCounter, metric.Type)
			assert.Equal(t, float64(5), metric.Value)
			foundCounter = true
		case "active_users":
			assert.Equal(t, MetricTypeGauge, metric.Type)
			assert.Equal(t, float64(10), metric.Value)
			foundGauge = true
		case "latency":
			assert.Equal(t, MetricTypeHistogram, metric.Type)
			assert.Equal(t, float64(1), metric.Value) // Total count
			foundHistogram = true
		}
		assert.NotZero(t, metric.Timestamp)
	}

	assert.True(t, foundCounter, "Counter metric not found")
	assert.True(t, foundGauge, "Gauge metric not found")
	assert.True(t, foundHistogram, "Histogram metric not found")
}

func TestMetricsCollector(t *testing.T) {
	registry := NewMetricsRegistry()
	collector := NewMetricsCollector(registry, 100*time.Millisecond)

	assert.NotNil(t, collector)
	assert.Equal(t, registry, collector.registry)
	assert.Equal(t, 100*time.Millisecond, collector.interval)
	assert.NotNil(t, collector.ctx)
	assert.NotNil(t, collector.cancel)
}

func TestGlobalRegistry(t *testing.T) {
	// Test global functions
	counter := RegisterCounter("global_requests", "Global requests", map[string]string{"global": "true"})
	assert.NotNil(t, counter)

	counter.Add(10)

	retrieved, exists := GetCounter("global_requests")
	assert.True(t, exists)
	assert.Equal(t, int64(10), retrieved.Get())

	gauge := RegisterGauge("global_memory", "Global memory", nil)
	assert.NotNil(t, gauge)

	gauge.Set(1024)

	retrievedGauge, exists := GetGauge("global_memory")
	assert.True(t, exists)
	assert.Equal(t, int64(1024), retrievedGauge.Get())

	histogram := RegisterHistogram("global_latency", "Global latency", []float64{0.1, 1.0}, nil)
	assert.NotNil(t, histogram)

	histogram.Observe(0.5)

	retrievedHistogram, exists := GetHistogram("global_latency")
	assert.True(t, exists)
	assert.Equal(t, int64(1), retrievedHistogram.GetTotalCount())
}

func TestMetricType_Constants(t *testing.T) {
	assert.Equal(t, MetricType("counter"), MetricTypeCounter)
	assert.Equal(t, MetricType("gauge"), MetricTypeGauge)
	assert.Equal(t, MetricType("histogram"), MetricTypeHistogram)
	assert.Equal(t, MetricType("summary"), MetricTypeSummary)
}

func TestConcurrentAccess(t *testing.T) {
	counter := NewCounter("concurrent_counter", "Concurrent counter", nil)
	gauge := NewGauge("concurrent_gauge", "Concurrent gauge", nil)
	histogram := NewHistogram("concurrent_histogram", "Concurrent histogram", []float64{1.0, 10.0}, nil)

	var wg sync.WaitGroup

	// Test concurrent counter operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				counter.Inc()
			}
		}()
	}

	// Test concurrent gauge operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				gauge.Inc()
			}
		}()
	}

	// Test concurrent histogram operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				histogram.Observe(float64(j % 20))
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(1000), counter.Get())             // 10 goroutines * 100 increments
	assert.Equal(t, int64(1000), gauge.Get())               // 10 goroutines * 100 increments
	assert.Equal(t, int64(1000), histogram.GetTotalCount()) // 10 goroutines * 100 observations
}

func TestHistogram_BucketBoundaries(t *testing.T) {
	buckets := []float64{1.0, 5.0, 10.0}
	histogram := NewHistogram("test_histogram", "Test histogram", buckets, nil)

	testCases := []struct {
		value    float64
		expected int
	}{
		{0.5, 0},  // <= 1.0
		{1.0, 0},  // <= 1.0
		{2.0, 1},  // <= 5.0
		{5.0, 1},  // <= 5.0
		{7.0, 2},  // <= 10.0
		{10.0, 2}, // <= 10.0
		{15.0, 3}, // > 10.0 (inf bucket)
	}

	for _, tc := range testCases {
		histogram.Observe(tc.value)
		assert.Equal(t, int64(1), histogram.GetCount(tc.expected), "Value %f should go to bucket %d", tc.value, tc.expected)

		// Reset for next test
		histogram = NewHistogram("test_histogram", "Test histogram", buckets, nil)
	}
}

func TestMetricsRegistry_OverwriteMetrics(t *testing.T) {
	registry := NewMetricsRegistry()

	// Register a counter
	counter1 := registry.RegisterCounter("test_metric", "First counter", map[string]string{"version": "1"})
	counter1.Add(5)

	// Register another counter with the same name (should overwrite)
	counter2 := registry.RegisterCounter("test_metric", "Second counter", map[string]string{"version": "2"})
	counter2.Add(10)

	// Should only have one metric with the new values
	retrieved, exists := registry.GetCounter("test_metric")
	assert.True(t, exists)
	assert.Equal(t, int64(10), retrieved.Get())
	assert.Equal(t, "Second counter", retrieved.description)
	assert.Equal(t, map[string]string{"version": "2"}, retrieved.labels)

	metrics := registry.GetAllMetrics()
	assert.Len(t, metrics, 1)
	assert.Equal(t, float64(10), metrics[0].Value)
}
