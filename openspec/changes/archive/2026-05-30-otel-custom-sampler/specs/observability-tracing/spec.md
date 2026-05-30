## MODIFIED Requirements

### Requirement: TracerProvider Initialisation

The system SHALL provide a `NewTracerProvider(ctx context.Context, exporter string, sampleRatio float64) (*sdktrace.TracerProvider, error)` function in `internal/observability/tracing.go`. It SHALL switch on the `exporter` argument:

- `"tempo"` — creates an OTLP gRPC exporter pointed at `OTEL_TEMPO_ENDPOINT` (default `localhost:4317`)
- `"gcp"` — creates a GCP Cloud Trace exporter using Application Default Credentials and `GOOGLE_CLOUD_PROJECT` env var

The function SHALL register the constructed provider as the global OTel `TracerProvider` via `otel.SetTracerProvider`. It SHALL also set the global `TextMapPropagator` to W3C TraceContext + Baggage. The caller is responsible for calling `provider.Shutdown(ctx)` on application exit.

The provider SHALL be constructed with:
- A resource that sets `service.name` to `"bookleaf"`
- `sdktrace.AlwaysSample()` as the sampler
- A `filteringProcessor` wrapping the batch exporter, configured with `sampleRatio`

#### Scenario: Tempo exporter selected

- **WHEN** `OTEL_EXPORTER=tempo`
- **THEN** `NewTracerProvider` returns a provider that exports spans via OTLP gRPC to the configured Tempo endpoint

#### Scenario: GCP exporter selected

- **WHEN** `OTEL_EXPORTER=gcp`
- **THEN** `NewTracerProvider` returns a provider that exports spans to GCP Cloud Trace

#### Scenario: Unknown exporter

- **WHEN** `OTEL_EXPORTER` is set to an unrecognised value
- **THEN** `NewTracerProvider` returns an error naming the unrecognised value

#### Scenario: Service name is set on traces

- **WHEN** `NewTracerProvider` is called
- **THEN** the provider's resource includes `service.name=bookleaf`

#### Scenario: Ratio sampler is wired

- **WHEN** `NewTracerProvider` is called with `sampleRatio=0.1`
- **THEN** the provider uses `AlwaysSample` as the SDK sampler and a `filteringProcessor` with ratio `0.1`
