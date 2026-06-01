package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func NewLogger(format string) (*zap.Logger, error) {
	if format == "json" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// LoggerFromContext returns a child logger with a trace_id field if the context
// carries an active OTel span. Returns the base logger unchanged if no span is present.
func LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().HasTraceID() {
		return base
	}
	return base.With(zap.String("trace_id", span.SpanContext().TraceID().String()))
}
