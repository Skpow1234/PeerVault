package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a single metric
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description"`
}

// Counter represents a counter metric
type Counter struct {
	name        string
	value       int64
	labels      map[string]string
	description string
	mu          sync.RWMutex
}

// NewCounter creates a new counter metric
func NewCounter(name, description string, labels map[string]string) *Counter {
	return &Counter{
		name:        name,
		labels:      labels,
		description: description,
	}
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add adds the given value to the counter
func (c *Counter) Add(delta int64) {
	atomic.AddInt64(&c.value, delta)
}

// Get returns the current counter value
func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.value)
}

// ToMetric converts the counter to a Metric
func (c *Counter) ToMetric() Metric {
	return Metric{
		Name:        c.name,
		Type:        MetricTypeCounter,
		Value:       float64(c.Get()),
		Labels:      c.labels,
		Timestamp:   time.Now(),
		Description: c.description,
	}
}

// Gauge represents a gauge metric
type Gauge struct {
	name        string
	value       int64
	labels      map[string]string
	description string
	mu          sync.RWMutex
}

// NewGauge creates a new gauge metric
func NewGauge(name, description string, labels map[string]string) *Gauge {
	return &Gauge{
		name:        name,
		labels:      labels,
		description: description,
	}
}

// Set sets the gauge value
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Inc increments the gauge by 1
func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

// Dec decrements the gauge by 1
func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

// Add adds the given value to the gauge
func (g *Gauge) Add(delta int64) {
	atomic.AddInt64(&g.value, delta)
}

// Sub subtracts the given value from the gauge
func (g *Gauge) Sub(delta int64) {
	atomic.AddInt64(&g.value, -delta)
}

// Get returns the current gauge value
func (g *Gauge) Get() int64 {
	return atomic.LoadInt64(&g.value)
}

// ToMetric converts the gauge to a Metric
func (g *Gauge) ToMetric() Metric {
	return Metric{
		Name:        g.name,
		Type:        MetricTypeGauge,
		Value:       float64(g.Get()),
		Labels:      g.labels,
		Timestamp:   time.Now(),
		Description: g.description,
	}
}

// Histogram represents a histogram metric
type Histogram struct {
	name        string
	buckets     []float64
	counts      []int64
	sum         int64
	labels      map[string]string
	description string
	mu          sync.RWMutex
}

// NewHistogram creates a new histogram metric
func NewHistogram(name, description string, buckets []float64, labels map[string]string) *Histogram {
	return &Histogram{
		name:        name,
		buckets:     buckets,
		counts:      make([]int64, len(buckets)+1), // +1 for the +Inf bucket
		labels:      labels,
		description: description,
	}
}

// Observe records a value in the histogram
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	atomic.AddInt64(&h.sum, int64(value))

	// Find the appropriate bucket
	bucketIndex := len(h.buckets) // Default to +Inf bucket
	for i, bucket := range h.buckets {
		if value <= bucket {
			bucketIndex = i
			break
		}
	}

	atomic.AddInt64(&h.counts[bucketIndex], 1)
}

// GetCount returns the count for a specific bucket
func (h *Histogram) GetCount(bucketIndex int) int64 {
	if bucketIndex < 0 || bucketIndex >= len(h.counts) {
		return 0
	}
	return atomic.LoadInt64(&h.counts[bucketIndex])
}

// GetSum returns the sum of all observed values
func (h *Histogram) GetSum() int64 {
	return atomic.LoadInt64(&h.sum)
}

// GetTotalCount returns the total count of observations
func (h *Histogram) GetTotalCount() int64 {
	var total int64
	for i := range h.counts {
		total += atomic.LoadInt64(&h.counts[i])
	}
	return total
}

// ToMetric converts the histogram to a Metric
func (h *Histogram) ToMetric() Metric {
	return Metric{
		Name:        h.name,
		Type:        MetricTypeHistogram,
		Value:       float64(h.GetTotalCount()),
		Labels:      h.labels,
		Timestamp:   time.Now(),
		Description: h.description,
	}
}

// MetricsRegistry manages all metrics
type MetricsRegistry struct {
	metrics map[string]interface{}
	mu      sync.RWMutex
}

// NewMetricsRegistry creates a new metrics registry
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]interface{}),
	}
}

// RegisterCounter registers a counter metric
func (mr *MetricsRegistry) RegisterCounter(name, description string, labels map[string]string) *Counter {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	counter := NewCounter(name, description, labels)
	mr.metrics[name] = counter
	return counter
}

// RegisterGauge registers a gauge metric
func (mr *MetricsRegistry) RegisterGauge(name, description string, labels map[string]string) *Gauge {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	gauge := NewGauge(name, description, labels)
	mr.metrics[name] = gauge
	return gauge
}

// RegisterHistogram registers a histogram metric
func (mr *MetricsRegistry) RegisterHistogram(name, description string, buckets []float64, labels map[string]string) *Histogram {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	histogram := NewHistogram(name, description, buckets, labels)
	mr.metrics[name] = histogram
	return histogram
}

// GetCounter retrieves a counter metric
func (mr *MetricsRegistry) GetCounter(name string) (*Counter, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	metric, exists := mr.metrics[name]
	if !exists {
		return nil, false
	}

	counter, ok := metric.(*Counter)
	return counter, ok
}

// GetGauge retrieves a gauge metric
func (mr *MetricsRegistry) GetGauge(name string) (*Gauge, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	metric, exists := mr.metrics[name]
	if !exists {
		return nil, false
	}

	gauge, ok := metric.(*Gauge)
	return gauge, ok
}

// GetHistogram retrieves a histogram metric
func (mr *MetricsRegistry) GetHistogram(name string) (*Histogram, bool) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	metric, exists := mr.metrics[name]
	if !exists {
		return nil, false
	}

	histogram, ok := metric.(*Histogram)
	return histogram, ok
}

// GetAllMetrics returns all registered metrics
func (mr *MetricsRegistry) GetAllMetrics() []Metric {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	var metrics []Metric
	for _, metric := range mr.metrics {
		switch m := metric.(type) {
		case *Counter:
			metrics = append(metrics, m.ToMetric())
		case *Gauge:
			metrics = append(metrics, m.ToMetric())
		case *Histogram:
			metrics = append(metrics, m.ToMetric())
		}
	}

	return metrics
}

// MetricsCollector collects metrics from the registry
type MetricsCollector struct {
	registry *MetricsRegistry
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(registry *MetricsRegistry, interval time.Duration) *MetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsCollector{
		registry: registry,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the metrics collection
func (mc *MetricsCollector) Start() {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.collectMetrics()
		case <-mc.ctx.Done():
			return
		}
	}
}

// Stop stops the metrics collection
func (mc *MetricsCollector) Stop() {
	mc.cancel()
}

// collectMetrics collects all metrics from the registry
func (mc *MetricsCollector) collectMetrics() {
	metrics := mc.registry.GetAllMetrics()
	// In a real implementation, this would send metrics to a monitoring system
	_ = metrics
}

// Global metrics registry
var GlobalRegistry = NewMetricsRegistry()

// Convenience functions for global metrics
func RegisterCounter(name, description string, labels map[string]string) *Counter {
	return GlobalRegistry.RegisterCounter(name, description, labels)
}

func RegisterGauge(name, description string, labels map[string]string) *Gauge {
	return GlobalRegistry.RegisterGauge(name, description, labels)
}

func RegisterHistogram(name, description string, buckets []float64, labels map[string]string) *Histogram {
	return GlobalRegistry.RegisterHistogram(name, description, buckets, labels)
}

func GetCounter(name string) (*Counter, bool) {
	return GlobalRegistry.GetCounter(name)
}

func GetGauge(name string) (*Gauge, bool) {
	return GlobalRegistry.GetGauge(name)
}

func GetHistogram(name string) (*Histogram, bool) {
	return GlobalRegistry.GetHistogram(name)
}
