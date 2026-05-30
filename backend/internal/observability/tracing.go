package observability

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	cloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type filteringProcessor struct {
	delegate sdktrace.SpanProcessor
	ratio    float64
}

func (p *filteringProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
	p.delegate.OnStart(parent, s)
}

func (p *filteringProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	if s.Status().Code == codes.Error {
		p.delegate.OnEnd(s)
		return
	}
	tid := s.SpanContext().TraceID()
	hash := binary.BigEndian.Uint64(tid[8:])
	if float64(hash)/math.MaxUint64 < p.ratio {
		p.delegate.OnEnd(s)
	}
}

func (p *filteringProcessor) Shutdown(ctx context.Context) error {
	return p.delegate.Shutdown(ctx)
}

func (p *filteringProcessor) ForceFlush(ctx context.Context) error {
	return p.delegate.ForceFlush(ctx)
}

func NewTracerProvider(ctx context.Context, exporter string, sampleRatio float64) (*sdktrace.TracerProvider, error) {
	var exp sdktrace.SpanExporter
	var err error

	switch exporter {
	case "tempo":
		endpoint := os.Getenv("OTEL_TEMPO_ENDPOINT")
		if endpoint == "" {
			endpoint = "localhost:4317"
		}
		exp, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("create tempo otlp exporter: %w", err)
		}

	case "gcp":
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		exp, err = cloudtrace.New(cloudtrace.WithProjectID(projectID))
		if err != nil {
			return nil, fmt.Errorf("create gcp cloud trace exporter: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown trace exporter %q: must be \"tempo\" or \"gcp\"", exporter)
	}

	res := resource.NewWithAttributes(
		resource.Default().SchemaURL(),
		attribute.String("service.name", "bookleaf"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(&filteringProcessor{
			delegate: sdktrace.NewBatchSpanProcessor(exp),
			ratio:    sampleRatio,
		}),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
