package observability

import (
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

type Telemetry struct {
	Logger *zap.Logger
	Tracer trace.Tracer
	Meter  metric.Meter
}

// NewTelemetry constructs a Telemetry. Nil fields are substituted with noop
// implementations, so constructors and tests can pass nil without nil-guarding.
func NewTelemetry(logger *zap.Logger, tracer trace.Tracer, meter metric.Meter) *Telemetry {
	if logger == nil {
		logger = zap.NewNop()
	}
	if tracer == nil {
		tracer = tracenoop.NewTracerProvider().Tracer("")
	}
	if meter == nil {
		meter = metricnoop.NewMeterProvider().Meter("")
	}
	return &Telemetry{Logger: logger, Tracer: tracer, Meter: meter}
}
