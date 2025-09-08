package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// TraceID represents a unique trace identifier
type TraceID string

// SpanID represents a unique span identifier
type SpanID string

// SpanKind represents the kind of span
type SpanKind string

const (
	SpanKindClient   SpanKind = "client"
	SpanKindServer   SpanKind = "server"
	SpanKindProducer SpanKind = "producer"
	SpanKindConsumer SpanKind = "consumer"
	SpanKindInternal SpanKind = "internal"
)

// SpanStatus represents the status of a span
type SpanStatus string

const (
	SpanStatusOK    SpanStatus = "ok"
	SpanStatusError SpanStatus = "error"
)

// Span represents a span in a trace
type Span struct {
	TraceID    TraceID                `json:"trace_id"`
	SpanID     SpanID                 `json:"span_id"`
	ParentID   SpanID                 `json:"parent_id,omitempty"`
	Name       string                 `json:"name"`
	Kind       SpanKind               `json:"kind"`
	Status     SpanStatus             `json:"status"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   time.Duration          `json:"duration"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Events     []SpanEvent            `json:"events,omitempty"`
	Links      []SpanLink             `json:"links,omitempty"`
}

// SpanEvent represents an event within a span
type SpanEvent struct {
	Name       string                 `json:"name"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// SpanLink represents a link to another span
type SpanLink struct {
	TraceID    TraceID                `json:"trace_id"`
	SpanID     SpanID                 `json:"span_id"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Tracer interface for creating and managing spans
type Tracer interface {
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)
	Finish(span Span)
	Inject(ctx context.Context, span Span) (map[string]string, error)
	Extract(ctx context.Context, headers map[string]string) (Span, error)
}

// SpanOption represents an option for creating a span
type SpanOption func(*Span)

// WithSpanKind sets the span kind
func WithSpanKind(kind SpanKind) SpanOption {
	return func(s *Span) {
		s.Kind = kind
	}
}

// WithParent sets the parent span
func WithParent(parent Span) SpanOption {
	return func(s *Span) {
		s.ParentID = parent.SpanID
	}
}

// WithAttributes sets span attributes
func WithAttributes(attrs map[string]interface{}) SpanOption {
	return func(s *Span) {
		if s.Attributes == nil {
			s.Attributes = make(map[string]interface{})
		}
		for k, v := range attrs {
			s.Attributes[k] = v
		}
	}
}

// SimpleTracer implements a simple in-memory tracer
type SimpleTracer struct {
	spans map[SpanID]Span
	mu    sync.RWMutex
}

// NewSimpleTracer creates a new simple tracer
func NewSimpleTracer() *SimpleTracer {
	return &SimpleTracer{
		spans: make(map[SpanID]Span),
	}
}

// StartSpan starts a new span
func (st *SimpleTracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	span := Span{
		TraceID:    generateTraceID(),
		SpanID:     generateSpanID(),
		Name:       name,
		Kind:       SpanKindInternal,
		Status:     SpanStatusOK,
		StartTime:  time.Now(),
		Attributes: make(map[string]interface{}),
		Events:     make([]SpanEvent, 0),
		Links:      make([]SpanLink, 0),
	}

	// Apply options
	for _, opt := range opts {
		opt(&span)
	}

	// Store span
	st.mu.Lock()
	st.spans[span.SpanID] = span
	st.mu.Unlock()

	// Add span to context
	ctx = context.WithValue(ctx, contextKey("span"), span)

	return ctx, span
}

// Finish finishes a span
func (st *SimpleTracer) Finish(span Span) {
	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	st.mu.Lock()
	st.spans[span.SpanID] = span
	st.mu.Unlock()
}

// Inject injects span context into headers
func (st *SimpleTracer) Inject(ctx context.Context, span Span) (map[string]string, error) {
	headers := map[string]string{
		"trace-id": string(span.TraceID),
		"span-id":  string(span.SpanID),
	}

	if span.ParentID != "" {
		headers["parent-id"] = string(span.ParentID)
	}

	return headers, nil
}

// Extract extracts span context from headers
func (st *SimpleTracer) Extract(ctx context.Context, headers map[string]string) (Span, error) {
	traceID, ok := headers["trace-id"]
	if !ok {
		return Span{}, fmt.Errorf("trace-id header not found")
	}

	spanID, ok := headers["span-id"]
	if !ok {
		return Span{}, fmt.Errorf("span-id header not found")
	}

	span := Span{
		TraceID:    TraceID(traceID),
		SpanID:     SpanID(spanID),
		Attributes: make(map[string]interface{}),
		Events:     make([]SpanEvent, 0),
		Links:      make([]SpanLink, 0),
	}

	if parentID, ok := headers["parent-id"]; ok {
		span.ParentID = SpanID(parentID)
	}

	return span, nil
}

// GetSpan retrieves a span by ID
func (st *SimpleTracer) GetSpan(spanID SpanID) (Span, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	span, exists := st.spans[spanID]
	return span, exists
}

// GetSpansByTraceID retrieves all spans for a trace
func (st *SimpleTracer) GetSpansByTraceID(traceID TraceID) []Span {
	st.mu.RLock()
	defer st.mu.RUnlock()

	var spans []Span
	for _, span := range st.spans {
		if span.TraceID == traceID {
			spans = append(spans, span)
		}
	}

	return spans
}

// GetAllSpans retrieves all spans
func (st *SimpleTracer) GetAllSpans() []Span {
	st.mu.RLock()
	defer st.mu.RUnlock()

	spans := make([]Span, 0, len(st.spans))
	for _, span := range st.spans {
		spans = append(spans, span)
	}

	return spans
}

// Clear clears all spans
func (st *SimpleTracer) Clear() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.spans = make(map[SpanID]Span)
}

// SpanContext represents span context in a context
type SpanContext struct {
	Span Span
}

// FromContext extracts span from context
func FromContext(ctx context.Context) (Span, bool) {
	span, ok := ctx.Value(contextKey("span")).(Span)
	return span, ok
}

// WithSpan adds span to context
func WithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, contextKey("span"), span)
}

// StartSpanFromContext starts a new span from existing context
func StartSpanFromContext(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	// Get parent span if exists
	if parentSpan, ok := FromContext(ctx); ok {
		opts = append(opts, WithParent(parentSpan))
	}

	return GlobalTracer.StartSpan(ctx, name, opts...)
}

// FinishSpanFromContext finishes a span from context
func FinishSpanFromContext(ctx context.Context) {
	if span, ok := FromContext(ctx); ok {
		GlobalTracer.Finish(span)
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs map[string]interface{}) {
	if span, ok := FromContext(ctx); ok {
		event := SpanEvent{
			Name:       name,
			Timestamp:  time.Now(),
			Attributes: attrs,
		}

		span.Events = append(span.Events, event)

		// Update the span in the tracer's storage
		if tracer, ok := GlobalTracer.(*SimpleTracer); ok {
			tracer.mu.Lock()
			if storedSpan, exists := tracer.spans[span.SpanID]; exists {
				storedSpan.Events = append(storedSpan.Events, event)
				tracer.spans[span.SpanID] = storedSpan
			}
			tracer.mu.Unlock()
		}
	}
}

// SetSpanStatus sets the status of the current span
func SetSpanStatus(ctx context.Context, status SpanStatus) {
	if span, ok := FromContext(ctx); ok {
		// Update the span in the tracer's storage
		if tracer, ok := GlobalTracer.(*SimpleTracer); ok {
			tracer.mu.Lock()
			if storedSpan, exists := tracer.spans[span.SpanID]; exists {
				storedSpan.Status = status
				tracer.spans[span.SpanID] = storedSpan
			}
			tracer.mu.Unlock()
		}
	}
}

// SetSpanAttribute sets an attribute on the current span
func SetSpanAttribute(ctx context.Context, key string, value interface{}) {
	if span, ok := FromContext(ctx); ok {
		if span.Attributes == nil {
			span.Attributes = make(map[string]interface{})
		}
		span.Attributes[key] = value

		// Update the span in the tracer's storage
		if tracer, ok := GlobalTracer.(*SimpleTracer); ok {
			tracer.mu.Lock()
			if storedSpan, exists := tracer.spans[span.SpanID]; exists {
				if storedSpan.Attributes == nil {
					storedSpan.Attributes = make(map[string]interface{})
				}
				storedSpan.Attributes[key] = value
				tracer.spans[span.SpanID] = storedSpan
			}
			tracer.mu.Unlock()
		}
	}
}

// Global tracer instance
var GlobalTracer Tracer = NewSimpleTracer()

// Utility functions for generating IDs
func generateTraceID() TraceID {
	return TraceID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func generateSpanID() SpanID {
	return SpanID(fmt.Sprintf("%d", time.Now().UnixNano()))
}
