package observability

import (
	"context"
	"fmt"
	"os"

	cloudtrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewTracerProvider(ctx context.Context, exporter string) (*sdktrace.TracerProvider, error) {
	var exp sdktrace.SpanExporter
	var err error

	switch exporter {
	case "jaeger":
		endpoint := os.Getenv("OTEL_JAEGER_ENDPOINT")
		if endpoint == "" {
			endpoint = "localhost:4317"
		}
		exp, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("create jaeger otlp exporter: %w", err)
		}

	case "gcp":
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		exp, err = cloudtrace.New(cloudtrace.WithProjectID(projectID))
		if err != nil {
			return nil, fmt.Errorf("create gcp cloud trace exporter: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown trace exporter %q: must be \"jaeger\" or \"gcp\"", exporter)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
