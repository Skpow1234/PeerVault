package tracing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTracingConstants(t *testing.T) {
	// Test span kind constants
	assert.Equal(t, SpanKind("client"), SpanKindClient)
	assert.Equal(t, SpanKind("server"), SpanKindServer)
	assert.Equal(t, SpanKind("producer"), SpanKindProducer)
	assert.Equal(t, SpanKind("consumer"), SpanKindConsumer)
	assert.Equal(t, SpanKind("internal"), SpanKindInternal)

	// Test span status constants
	assert.Equal(t, SpanStatus("ok"), SpanStatusOK)
	assert.Equal(t, SpanStatus("error"), SpanStatusError)
}

func TestSpan(t *testing.T) {
	span := Span{
		TraceID:    "trace123",
		SpanID:     "span456",
		ParentID:   "parent789",
		Name:       "test-span",
		Kind:       SpanKindServer,
		Status:     SpanStatusOK,
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(100 * time.Millisecond),
		Duration:   100 * time.Millisecond,
		Attributes: map[string]interface{}{"key": "value"},
		Events:     []SpanEvent{{Name: "event1", Timestamp: time.Now()}},
		Links:      []SpanLink{{TraceID: "trace999", SpanID: "span888"}},
	}

	assert.Equal(t, TraceID("trace123"), span.TraceID)
	assert.Equal(t, SpanID("span456"), span.SpanID)
	assert.Equal(t, SpanID("parent789"), span.ParentID)
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, SpanKindServer, span.Kind)
	assert.Equal(t, SpanStatusOK, span.Status)
	assert.NotZero(t, span.StartTime)
	assert.NotZero(t, span.EndTime)
	assert.Equal(t, 100*time.Millisecond, span.Duration)
	assert.Len(t, span.Attributes, 1)
	assert.Len(t, span.Events, 1)
	assert.Len(t, span.Links, 1)
}

func TestSpanEvent(t *testing.T) {
	event := SpanEvent{
		Name:       "test-event",
		Timestamp:  time.Now(),
		Attributes: map[string]interface{}{"event_key": "event_value"},
	}

	assert.Equal(t, "test-event", event.Name)
	assert.NotZero(t, event.Timestamp)
	assert.Len(t, event.Attributes, 1)
	assert.Equal(t, "event_value", event.Attributes["event_key"])
}

func TestSpanLink(t *testing.T) {
	link := SpanLink{
		TraceID:    "trace123",
		SpanID:     "span456",
		Attributes: map[string]interface{}{"link_key": "link_value"},
	}

	assert.Equal(t, TraceID("trace123"), link.TraceID)
	assert.Equal(t, SpanID("span456"), link.SpanID)
	assert.Len(t, link.Attributes, 1)
	assert.Equal(t, "link_value", link.Attributes["link_key"])
}

func TestSpanOptions(t *testing.T) {
	span := &Span{}

	// Test WithSpanKind
	WithSpanKind(SpanKindClient)(span)
	assert.Equal(t, SpanKindClient, span.Kind)

	// Test WithParent
	parent := Span{SpanID: "parent123"}
	WithParent(parent)(span)
	assert.Equal(t, SpanID("parent123"), span.ParentID)

	// Test WithAttributes
	attrs := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	WithAttributes(attrs)(span)
	assert.Len(t, span.Attributes, 2)
	assert.Equal(t, "value1", span.Attributes["key1"])
	assert.Equal(t, "value2", span.Attributes["key2"])

	// Test WithAttributes with existing attributes
	WithAttributes(map[string]interface{}{"key3": "value3"})(span)
	assert.Len(t, span.Attributes, 3)
	assert.Equal(t, "value3", span.Attributes["key3"])
}

func TestSimpleTracer_NewSimpleTracer(t *testing.T) {
	tracer := NewSimpleTracer()
	assert.NotNil(t, tracer)
	assert.NotNil(t, tracer.spans)
	assert.Len(t, tracer.spans, 0)
}

func TestSimpleTracer_StartSpan(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Test basic span creation
	ctx, span := tracer.StartSpan(ctx, "test-span")
	assert.NotEmpty(t, span.TraceID)
	assert.NotEmpty(t, span.SpanID)
	assert.Equal(t, "test-span", span.Name)
	assert.Equal(t, SpanKindInternal, span.Kind)
	assert.Equal(t, SpanStatusOK, span.Status)
	assert.NotZero(t, span.StartTime)
	assert.NotNil(t, span.Attributes)
	assert.NotNil(t, span.Events)
	assert.NotNil(t, span.Links)

	// Test span is stored
	storedSpan, exists := tracer.GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.Equal(t, span.SpanID, storedSpan.SpanID)

	// Test span is added to context
	ctxSpan, ok := FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, span.SpanID, ctxSpan.SpanID)
}

func TestSimpleTracer_StartSpan_WithOptions(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Test span with options
	attrs := map[string]interface{}{"key": "value"}
	_, span := tracer.StartSpan(ctx, "test-span",
		WithSpanKind(SpanKindServer),
		WithAttributes(attrs),
	)

	assert.Equal(t, SpanKindServer, span.Kind)
	assert.Len(t, span.Attributes, 1)
	assert.Equal(t, "value", span.Attributes["key"])
}

func TestSimpleTracer_StartSpan_WithParent(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Create parent span
	ctx, parentSpan := tracer.StartSpan(ctx, "parent-span")

	// Create child span
	_, childSpan := tracer.StartSpan(ctx, "child-span", WithParent(parentSpan))

	assert.Equal(t, parentSpan.SpanID, childSpan.ParentID)
	assert.Equal(t, parentSpan.TraceID, childSpan.TraceID)
}

func TestSimpleTracer_Finish(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Start span
	_, span := tracer.StartSpan(ctx, "test-span")
	startTime := span.StartTime

	// Wait a bit to ensure duration is measurable
	time.Sleep(10 * time.Millisecond)

	// Finish span
	tracer.Finish(span)

	// Check that span is updated
	storedSpan, exists := tracer.GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.NotZero(t, storedSpan.EndTime)
	assert.NotZero(t, storedSpan.Duration)
	assert.True(t, storedSpan.EndTime.After(startTime))
}

func TestSimpleTracer_Inject(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Create span
	_, span := tracer.StartSpan(ctx, "test-span")

	// Inject span
	headers, err := tracer.Inject(ctx, span)
	assert.NoError(t, err)
	assert.Len(t, headers, 2)
	assert.Equal(t, string(span.TraceID), headers["trace-id"])
	assert.Equal(t, string(span.SpanID), headers["span-id"])

	// Test with parent span
	ctx, childSpan := tracer.StartSpan(ctx, "child-span", WithParent(span))
	headers, err = tracer.Inject(ctx, childSpan)
	assert.NoError(t, err)
	assert.Len(t, headers, 3)
	assert.Equal(t, string(childSpan.TraceID), headers["trace-id"])
	assert.Equal(t, string(childSpan.SpanID), headers["span-id"])
	assert.Equal(t, string(childSpan.ParentID), headers["parent-id"])
}

func TestSimpleTracer_Extract(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Test valid headers
	headers := map[string]string{
		"trace-id":  "trace123",
		"span-id":   "span456",
		"parent-id": "parent789",
	}

	span, err := tracer.Extract(ctx, headers)
	assert.NoError(t, err)
	assert.Equal(t, TraceID("trace123"), span.TraceID)
	assert.Equal(t, SpanID("span456"), span.SpanID)
	assert.Equal(t, SpanID("parent789"), span.ParentID)
	assert.NotNil(t, span.Attributes)
	assert.NotNil(t, span.Events)
	assert.NotNil(t, span.Links)

	// Test headers without parent-id
	headers = map[string]string{
		"trace-id": "trace123",
		"span-id":  "span456",
	}

	span, err = tracer.Extract(ctx, headers)
	assert.NoError(t, err)
	assert.Equal(t, TraceID("trace123"), span.TraceID)
	assert.Equal(t, SpanID("span456"), span.SpanID)
	assert.Empty(t, span.ParentID)

	// Test missing trace-id
	headers = map[string]string{
		"span-id": "span456",
	}

	_, err = tracer.Extract(ctx, headers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trace-id header not found")

	// Test missing span-id
	headers = map[string]string{
		"trace-id": "trace123",
	}

	_, err = tracer.Extract(ctx, headers)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "span-id header not found")
}

func TestSimpleTracer_GetSpan(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Test getting non-existent span
	_, exists := tracer.GetSpan("non-existent")
	assert.False(t, exists)

	// Test getting existing span
	_, span := tracer.StartSpan(ctx, "test-span")
	storedSpan, exists := tracer.GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.Equal(t, span.SpanID, storedSpan.SpanID)
}

func TestSimpleTracer_GetSpansByTraceID(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Create spans with same trace ID (parent-child relationship)
	ctx, span1 := tracer.StartSpan(ctx, "span1")
	ctx, span2 := tracer.StartSpan(ctx, "span2", WithParent(span1))

	// Create span with different trace ID
	_, span3 := tracer.StartSpan(ctx, "span3")

	// Wait a bit to ensure spans are stored
	time.Sleep(10 * time.Millisecond)

	// Get spans by trace ID
	spans := tracer.GetSpansByTraceID(span1.TraceID)
	assert.Len(t, spans, 2)

	// Check that both spans are returned
	spanIDs := make(map[SpanID]bool)
	for _, span := range spans {
		spanIDs[span.SpanID] = true
	}
	assert.True(t, spanIDs[span1.SpanID])
	assert.True(t, spanIDs[span2.SpanID])
	assert.False(t, spanIDs[span3.SpanID])
}

func TestSimpleTracer_GetAllSpans(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Create multiple spans
	ctx, span1 := tracer.StartSpan(ctx, "span1")
	_, span2 := tracer.StartSpan(ctx, "span2")
	_, span3 := tracer.StartSpan(ctx, "span3")

	// Wait a bit to ensure spans are stored
	time.Sleep(10 * time.Millisecond)

	// Get all spans
	spans := tracer.GetAllSpans()
	assert.Len(t, spans, 3)

	// Check that all spans are returned
	spanIDs := make(map[SpanID]bool)
	for _, span := range spans {
		spanIDs[span.SpanID] = true
	}
	assert.True(t, spanIDs[span1.SpanID])
	assert.True(t, spanIDs[span2.SpanID])
	assert.True(t, spanIDs[span3.SpanID])
}

func TestSimpleTracer_Clear(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Create some spans
	ctx, span1 := tracer.StartSpan(ctx, "span1")
	_, span2 := tracer.StartSpan(ctx, "span2")

	// Wait a bit to ensure spans are stored
	time.Sleep(10 * time.Millisecond)

	// Verify spans exist
	assert.Len(t, tracer.GetAllSpans(), 2)

	// Clear spans
	tracer.Clear()

	// Verify spans are cleared
	assert.Len(t, tracer.GetAllSpans(), 0)
	_, exists := tracer.GetSpan(span1.SpanID)
	assert.False(t, exists)
	_, exists = tracer.GetSpan(span2.SpanID)
	assert.False(t, exists)
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()

	// Test context without span
	_, ok := FromContext(ctx)
	assert.False(t, ok)

	// Test context with span
	tracer := NewSimpleTracer()
	ctx, span := tracer.StartSpan(ctx, "test-span")

	ctxSpan, ok := FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, span.SpanID, ctxSpan.SpanID)
}

func TestWithSpan(t *testing.T) {
	ctx := context.Background()
	span := Span{
		SpanID: "test-span-id",
		Name:   "test-span",
	}

	// Add span to context
	ctx = WithSpan(ctx, span)

	// Verify span is in context
	ctxSpan, ok := FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, span.SpanID, ctxSpan.SpanID)
}

func TestStartSpanFromContext(t *testing.T) {
	ctx := context.Background()

	// Test starting span without parent
	ctx, span1 := StartSpanFromContext(ctx, "span1")
	assert.NotEmpty(t, span1.SpanID)
	assert.Empty(t, span1.ParentID)

	// Test starting span with parent
	_, span2 := StartSpanFromContext(ctx, "span2")
	assert.NotEmpty(t, span2.SpanID)
	assert.Equal(t, span1.SpanID, span2.ParentID)
	assert.Equal(t, span1.TraceID, span2.TraceID)
}

func TestFinishSpanFromContext(t *testing.T) {
	ctx := context.Background()

	// Test finishing span without span in context
	FinishSpanFromContext(ctx) // Should not panic

	// Test finishing span with span in context
	ctx, span := StartSpanFromContext(ctx, "test-span")
	startTime := span.StartTime

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Finish span
	FinishSpanFromContext(ctx)

	// Verify span is finished
	storedSpan, exists := GlobalTracer.(*SimpleTracer).GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.NotZero(t, storedSpan.EndTime)
	assert.NotZero(t, storedSpan.Duration)
	assert.True(t, storedSpan.EndTime.After(startTime))
}

func TestAddSpanEvent(t *testing.T) {
	ctx := context.Background()

	// Test adding event without span in context
	AddSpanEvent(ctx, "test-event", nil) // Should not panic

	// Test adding event with span in context
	ctx, span := StartSpanFromContext(ctx, "test-span")
	attrs := map[string]interface{}{"key": "value"}

	AddSpanEvent(ctx, "test-event", attrs)

	// Verify event is added
	storedSpan, exists := GlobalTracer.(*SimpleTracer).GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.Len(t, storedSpan.Events, 1)
	assert.Equal(t, "test-event", storedSpan.Events[0].Name)
	assert.Len(t, storedSpan.Events[0].Attributes, 1)
	assert.Equal(t, "value", storedSpan.Events[0].Attributes["key"])
}

func TestSetSpanStatus(t *testing.T) {
	ctx := context.Background()

	// Test setting status without span in context
	SetSpanStatus(ctx, SpanStatusError) // Should not panic

	// Test setting status with span in context
	ctx, span := StartSpanFromContext(ctx, "test-span")

	SetSpanStatus(ctx, SpanStatusError)

	// Verify status is set
	storedSpan, exists := GlobalTracer.(*SimpleTracer).GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.Equal(t, SpanStatusError, storedSpan.Status)
}

func TestSetSpanAttribute(t *testing.T) {
	ctx := context.Background()

	// Test setting attribute without span in context
	SetSpanAttribute(ctx, "key", "value") // Should not panic

	// Test setting attribute with span in context
	ctx, span := StartSpanFromContext(ctx, "test-span")

	SetSpanAttribute(ctx, "key1", "value1")
	SetSpanAttribute(ctx, "key2", "value2")

	// Verify attributes are set
	storedSpan, exists := GlobalTracer.(*SimpleTracer).GetSpan(span.SpanID)
	assert.True(t, exists)
	assert.Len(t, storedSpan.Attributes, 2)
	assert.Equal(t, "value1", storedSpan.Attributes["key1"])
	assert.Equal(t, "value2", storedSpan.Attributes["key2"])
}

func TestGlobalTracer(t *testing.T) {
	// Test that global tracer is initialized
	assert.NotNil(t, GlobalTracer)

	// Test that it's a SimpleTracer
	_, ok := GlobalTracer.(*SimpleTracer)
	assert.True(t, ok)
}

func TestGenerateTraceID(t *testing.T) {
	id1 := generateTraceID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateTraceID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestGenerateSpanID(t *testing.T) {
	id1 := generateSpanID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateSpanID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestSimpleTracer_ConcurrentAccess(t *testing.T) {
	tracer := NewSimpleTracer()

	// Test concurrent span creation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			_, span := tracer.StartSpan(context.Background(), "concurrent-span")
			assert.NotEmpty(t, span.SpanID)

			// Test concurrent access to GetSpan
			storedSpan, exists := tracer.GetSpan(span.SpanID)
			assert.True(t, exists)
			assert.Equal(t, span.SpanID, storedSpan.SpanID)

			// Test concurrent access to GetAllSpans
			spans := tracer.GetAllSpans()
			assert.GreaterOrEqual(t, len(spans), 1)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Wait a bit to ensure all spans are stored
	time.Sleep(10 * time.Millisecond)

	// Verify all spans were created
	spans := tracer.GetAllSpans()
	assert.Len(t, spans, 10)
}

func TestSimpleTracer_EdgeCases(t *testing.T) {
	tracer := NewSimpleTracer()
	ctx := context.Background()

	// Test empty span name
	ctx, span := tracer.StartSpan(ctx, "")
	assert.Empty(t, span.Name)

	// Test span with empty attributes
	ctx, span = tracer.StartSpan(ctx, "test-span", WithAttributes(map[string]interface{}{}))
	assert.NotNil(t, span.Attributes)
	assert.Len(t, span.Attributes, 0)

	// Test getting span with empty ID
	_, exists := tracer.GetSpan("")
	assert.False(t, exists)

	// Test getting spans by empty trace ID
	spans := tracer.GetSpansByTraceID("")
	assert.Len(t, spans, 0)

	// Test extract with empty headers
	_, err := tracer.Extract(ctx, map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trace-id header not found")
}

func TestSpanContext(t *testing.T) {
	span := Span{
		SpanID: "test-span-id",
		Name:   "test-span",
	}

	spanCtx := SpanContext{Span: span}
	assert.Equal(t, span.SpanID, spanCtx.Span.SpanID)
	assert.Equal(t, span.Name, spanCtx.Span.Name)
}
