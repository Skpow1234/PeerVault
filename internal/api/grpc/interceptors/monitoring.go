package interceptors

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	EnableMetrics    bool
	EnableTracing    bool
	EnableProfiling  bool
	MetricsPrefix    string
	HistogramBuckets []float64
	CounterLabels    []string
	SkipMethods      []string
	SampleRate       float64
	MaxTraceDuration time.Duration
}

// DefaultMonitoringConfig returns the default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		EnableMetrics:    true,
		EnableTracing:    true,
		EnableProfiling:  false,
		MetricsPrefix:    "grpc",
		HistogramBuckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		CounterLabels:    []string{"method", "status_code", "user_id"},
		SkipMethods:      []string{"/peervault.PeerVaultService/HealthCheck"},
		SampleRate:       1.0,
		MaxTraceDuration: 30 * time.Second,
	}
}

// MonitoringInterceptor provides monitoring for gRPC services
type MonitoringInterceptor struct {
	config  *MonitoringConfig
	logger  *slog.Logger
	metrics *Metrics
	tracer  *Tracer
}

// Metrics represents gRPC metrics
type Metrics struct {
	RequestTotal    map[string]int64
	RequestDuration map[string][]float64
	ActiveStreams   int64
	ErrorCount      map[string]int64
	mutex           sync.RWMutex
}

// Tracer represents gRPC tracing
type Tracer struct {
	Traces map[string]*Trace
	mutex  sync.RWMutex
}

// Trace represents a single trace
type Trace struct {
	ID        string
	Method    string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Status    codes.Code
	UserID    string
	Metadata  map[string]string
}

// NewMonitoringInterceptor creates a new monitoring interceptor
func NewMonitoringInterceptor(config *MonitoringConfig, logger *slog.Logger) *MonitoringInterceptor {
	if config == nil {
		config = DefaultMonitoringConfig()
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &MonitoringInterceptor{
		config: config,
		logger: logger,
		metrics: &Metrics{
			RequestTotal:    make(map[string]int64),
			RequestDuration: make(map[string][]float64),
			ErrorCount:      make(map[string]int64),
		},
		tracer: &Tracer{
			Traces: make(map[string]*Trace),
		},
	}
}

// UnaryMonitoringInterceptor returns a unary server interceptor for monitoring
func (mi *MonitoringInterceptor) UnaryMonitoringInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if method should skip monitoring
		if mi.shouldSkipMonitoring(info.FullMethod) {
			return handler(ctx, req)
		}

		start := time.Now()
		traceID := mi.generateTraceID()

		// Start trace if enabled
		if mi.config.EnableTracing {
			mi.startTrace(traceID, info.FullMethod, ctx)
		}

		// Record metrics start
		if mi.config.EnableMetrics {
			mi.recordRequestStart(info.FullMethod)
		}

		// Call handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Record metrics end
		if mi.config.EnableMetrics {
			mi.recordRequestEnd(info.FullMethod, err, duration)
		}

		// End trace if enabled
		if mi.config.EnableTracing {
			mi.endTrace(traceID, err)
		}

		return resp, err
	}
}

// StreamMonitoringInterceptor returns a stream server interceptor for monitoring
func (mi *MonitoringInterceptor) StreamMonitoringInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Check if method should skip monitoring
		if mi.shouldSkipMonitoring(info.FullMethod) {
			return handler(srv, ss)
		}

		start := time.Now()
		traceID := mi.generateTraceID()

		// Start trace if enabled
		if mi.config.EnableTracing {
			mi.startTrace(traceID, info.FullMethod, ss.Context())
		}

		// Record stream start
		if mi.config.EnableMetrics {
			mi.recordStreamStart(info.FullMethod)
		}

		// Wrap stream for monitoring
		wrappedStream := &monitoringServerStream{
			ServerStream: ss,
			interceptor:  mi,
			traceID:      traceID,
			method:       info.FullMethod,
			startTime:    start,
		}

		// Call handler
		err := handler(srv, wrappedStream)

		duration := time.Since(start)

		// Record stream end
		if mi.config.EnableMetrics {
			mi.recordStreamEnd(info.FullMethod, err, duration)
		}

		// End trace if enabled
		if mi.config.EnableTracing {
			mi.endTrace(traceID, err)
		}

		return err
	}
}

// shouldSkipMonitoring checks if the method should skip monitoring
func (mi *MonitoringInterceptor) shouldSkipMonitoring(method string) bool {
	for _, skipMethod := range mi.config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}
	return false
}

// recordRequestStart records the start of a request
func (mi *MonitoringInterceptor) recordRequestStart(method string) {
	mi.metrics.mutex.Lock()
	defer mi.metrics.mutex.Unlock()

	mi.metrics.RequestTotal[method]++
}

// recordRequestEnd records the end of a request
func (mi *MonitoringInterceptor) recordRequestEnd(method string, err error, duration time.Duration) {
	mi.metrics.mutex.Lock()
	defer mi.metrics.mutex.Unlock()

	// Record duration
	durationSeconds := duration.Seconds()
	mi.metrics.RequestDuration[method] = append(mi.metrics.RequestDuration[method], durationSeconds)

	// Record error if present
	if err != nil {
		errorKey := method + "_error"
		mi.metrics.ErrorCount[errorKey]++
	}
}

// recordStreamStart records the start of a stream
func (mi *MonitoringInterceptor) recordStreamStart(method string) {
	mi.metrics.mutex.Lock()
	defer mi.metrics.mutex.Unlock()

	mi.metrics.ActiveStreams++
	mi.metrics.RequestTotal[method]++
}

// recordStreamEnd records the end of a stream
func (mi *MonitoringInterceptor) recordStreamEnd(method string, err error, duration time.Duration) {
	mi.metrics.mutex.Lock()
	defer mi.metrics.mutex.Unlock()

	mi.metrics.ActiveStreams--
	if mi.metrics.ActiveStreams < 0 {
		mi.metrics.ActiveStreams = 0
	}

	// Record duration
	durationSeconds := duration.Seconds()
	mi.metrics.RequestDuration[method] = append(mi.metrics.RequestDuration[method], durationSeconds)

	// Record error if present
	if err != nil {
		errorKey := method + "_error"
		mi.metrics.ErrorCount[errorKey]++
	}
}

// startTrace starts a new trace
func (mi *MonitoringInterceptor) startTrace(traceID, method string, ctx context.Context) {
	mi.tracer.mutex.Lock()
	defer mi.tracer.mutex.Unlock()

	trace := &Trace{
		ID:        traceID,
		Method:    method,
		StartTime: time.Now(),
		Metadata:  make(map[string]string),
	}

	// Add user ID if available
	if userID, ok := ctx.Value("user_id").(string); ok {
		trace.UserID = userID
	}

	mi.tracer.Traces[traceID] = trace
}

// endTrace ends a trace
func (mi *MonitoringInterceptor) endTrace(traceID string, err error) {
	mi.tracer.mutex.Lock()
	defer mi.tracer.mutex.Unlock()

	trace, exists := mi.tracer.Traces[traceID]
	if !exists {
		return
	}

	trace.EndTime = time.Now()
	trace.Duration = trace.EndTime.Sub(trace.StartTime)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			trace.Status = st.Code()
		} else {
			trace.Status = codes.Unknown
		}
	} else {
		trace.Status = codes.OK
	}

	// Clean up old traces
	mi.cleanupOldTraces()
}

// cleanupOldTraces removes traces older than MaxTraceDuration
func (mi *MonitoringInterceptor) cleanupOldTraces() {
	cutoff := time.Now().Add(-mi.config.MaxTraceDuration)
	for traceID, trace := range mi.tracer.Traces {
		if trace.StartTime.Before(cutoff) {
			delete(mi.tracer.Traces, traceID)
		}
	}
}

// generateTraceID generates a unique trace ID
func (mi *MonitoringInterceptor) generateTraceID() string {
	return time.Now().Format("20060102150405") + "_" + time.Now().Format("000000000")
}

// GetMetrics returns current metrics
func (mi *MonitoringInterceptor) GetMetrics() map[string]interface{} {
	mi.metrics.mutex.RLock()
	defer mi.metrics.mutex.RUnlock()

	metrics := make(map[string]interface{})
	metrics["request_total"] = mi.metrics.RequestTotal
	metrics["request_duration"] = mi.metrics.RequestDuration
	metrics["active_streams"] = mi.metrics.ActiveStreams
	metrics["error_count"] = mi.metrics.ErrorCount

	return metrics
}

// GetTraces returns current traces
func (mi *MonitoringInterceptor) GetTraces() map[string]*Trace {
	mi.tracer.mutex.RLock()
	defer mi.tracer.mutex.RUnlock()

	traces := make(map[string]*Trace)
	for traceID, trace := range mi.tracer.Traces {
		traces[traceID] = trace
	}

	return traces
}

// GetTrace returns a specific trace
func (mi *MonitoringInterceptor) GetTrace(traceID string) (*Trace, bool) {
	mi.tracer.mutex.RLock()
	defer mi.tracer.mutex.RUnlock()

	trace, exists := mi.tracer.Traces[traceID]
	return trace, exists
}

// ResetMetrics resets all metrics
func (mi *MonitoringInterceptor) ResetMetrics() {
	mi.metrics.mutex.Lock()
	defer mi.metrics.mutex.Unlock()

	mi.metrics.RequestTotal = make(map[string]int64)
	mi.metrics.RequestDuration = make(map[string][]float64)
	mi.metrics.ActiveStreams = 0
	mi.metrics.ErrorCount = make(map[string]int64)
}

// ResetTraces resets all traces
func (mi *MonitoringInterceptor) ResetTraces() {
	mi.tracer.mutex.Lock()
	defer mi.tracer.mutex.Unlock()

	mi.tracer.Traces = make(map[string]*Trace)
}

// monitoringServerStream wraps a ServerStream for monitoring
type monitoringServerStream struct {
	grpc.ServerStream
	interceptor  *MonitoringInterceptor
	traceID      string
	method       string
	startTime    time.Time
	messageCount int
}

// SendMsg monitors the message being sent
func (mss *monitoringServerStream) SendMsg(m interface{}) error {
	mss.messageCount++

	// Record message metrics
	mss.interceptor.metrics.mutex.Lock()
	messageKey := mss.method + "_sent"
	mss.interceptor.metrics.RequestTotal[messageKey]++
	mss.interceptor.metrics.mutex.Unlock()

	return mss.ServerStream.SendMsg(m)
}

// RecvMsg monitors the message being received
func (mss *monitoringServerStream) RecvMsg(m interface{}) error {
	err := mss.ServerStream.RecvMsg(m)

	if err == nil {
		mss.messageCount++

		// Record message metrics
		mss.interceptor.metrics.mutex.Lock()
		messageKey := mss.method + "_received"
		mss.interceptor.metrics.RequestTotal[messageKey]++
		mss.interceptor.metrics.mutex.Unlock()
	}

	return err
}
