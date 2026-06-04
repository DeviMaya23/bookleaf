package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func newTestProcessor(ratio float64) (*filteringProcessor, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	return &filteringProcessor{delegate: recorder, ratio: ratio}, recorder
}

// startWithTraceID starts a span whose trace ID is inherited from a remote parent
// with the given trace ID, ensuring the filteringProcessor sees a predictable hash.
func startWithTraceID(t *testing.T, proc *filteringProcessor, traceID trace.TraceID) {
	t.Helper()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(proc),
	)
	remoteCtx := trace.ContextWithRemoteSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     trace.SpanID{0x01},
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	}))
	_, span := tp.Tracer("test").Start(remoteCtx, "span")
	span.End()
}

func TestFilteringProcessor_ErrorSpanAlwaysForwarded(t *testing.T) {
	proc, recorder := newTestProcessor(0.0) // ratio=0: nothing sampled by ratio

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(proc),
	)
	_, span := tp.Tracer("test").Start(context.Background(), "error-span")
	span.SetStatus(codes.Error, "something went wrong")
	span.End()

	assert.Len(t, recorder.Ended(), 1, "error span should always be forwarded regardless of ratio")
}

func TestFilteringProcessor_NonErrorSpanDroppedOutsideRatio(t *testing.T) {
	proc, recorder := newTestProcessor(0.0) // ratio=0: drop everything non-error

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(proc),
	)
	_, span := tp.Tracer("test").Start(context.Background(), "ok-span")
	span.End()

	assert.Len(t, recorder.Ended(), 0, "non-error span should be dropped when ratio=0")
}

func TestFilteringProcessor_NonErrorSpanForwardedWithinRatio(t *testing.T) {
	// lower 8 bytes = 0x0000000000000001 → hash = 1 → 1/MaxUint64 ≈ 0 < 0.5 → forwarded
	lowHashTraceID := trace.TraceID{0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	proc, recorder := newTestProcessor(0.5)

	startWithTraceID(t, proc, lowHashTraceID)

	assert.Len(t, recorder.Ended(), 1, "span with hash below ratio should be forwarded")
}

func TestFilteringProcessor_NonErrorSpanDroppedByHash(t *testing.T) {
	// lower 8 bytes = 0xFFFFFFFFFFFFFFFF → hash = MaxUint64 → ~1.0, not < 0.5 → dropped
	highHashTraceID := trace.TraceID{0x01, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	proc, recorder := newTestProcessor(0.5)

	startWithTraceID(t, proc, highHashTraceID)

	assert.Len(t, recorder.Ended(), 0, "span with hash above ratio should be dropped")
}

func TestFilteringProcessor_ConsistentDecisionAcrossSameTraceID(t *testing.T) {
	proc, recorder := newTestProcessor(1.0)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(proc),
	)
	ctx, root := tp.Tracer("test").Start(context.Background(), "root")
	_, child := tp.Tracer("test").Start(ctx, "child")
	child.End()
	root.End()

	ended := recorder.Ended()
	assert.Len(t, ended, 2)
	assert.Equal(t, ended[0].SpanContext().TraceID(), ended[1].SpanContext().TraceID(),
		"both spans should share the same trace ID")
}
