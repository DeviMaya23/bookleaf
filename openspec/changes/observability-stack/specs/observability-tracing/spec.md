## ADDED Requirements

### Requirement: TracerProvider Initialisation

The system SHALL provide a `NewTracerProvider(ctx context.Context, exporter string) (*sdktrace.TracerProvider, error)` function in `internal/observability/tracing.go`. It SHALL switch on the `exporter` argument:

- `"jaeger"` — creates an OTLP gRPC exporter pointed at `OTEL_JAEGER_ENDPOINT` (default `localhost:4317`)
- `"gcp"` — creates a GCP Cloud Trace exporter using Application Default Credentials and `GOOGLE_CLOUD_PROJECT` env var

The function SHALL register the constructed provider as the global OTel `TracerProvider` via `otel.SetTracerProvider`. It SHALL also set the global `TextMapPropagator` to W3C TraceContext + Baggage. The caller is responsible for calling `provider.Shutdown(ctx)` on application exit.

#### Scenario: Jaeger exporter selected

- **WHEN** `OTEL_EXPORTER=jaeger`
- **THEN** `NewTracerProvider` returns a provider that exports spans to the configured Jaeger OTLP endpoint

#### Scenario: GCP exporter selected

- **WHEN** `OTEL_EXPORTER=gcp`
- **THEN** `NewTracerProvider` returns a provider that exports spans to GCP Cloud Trace

#### Scenario: Unknown exporter

- **WHEN** `OTEL_EXPORTER` is set to an unrecognised value
- **THEN** `NewTracerProvider` returns an error naming the unrecognised value

### Requirement: Echo Request Tracing Middleware

The system SHALL provide an Echo middleware `TraceMiddleware(tracer trace.Tracer) echo.MiddlewareFunc` in `internal/observability/echo_middleware.go`. The middleware SHALL:

1. Extract W3C `traceparent`/`tracestate` headers from the inbound request using the global propagator
2. Start a new span named after the matched route pattern (e.g., `GET /images/:id`)
3. Store the span context on the request context so downstream code can access it
4. End the span after the handler returns, recording the HTTP status code as a span attribute

#### Scenario: Inbound request creates a root span

- **WHEN** an HTTP request arrives without a `traceparent` header
- **THEN** the middleware creates a new root span
- **AND** the span name matches the route pattern

#### Scenario: Inbound request continues a remote trace

- **WHEN** an HTTP request arrives with a valid `traceparent` header
- **THEN** the middleware creates a child span whose parent is the remote trace context

#### Scenario: Span records HTTP status code

- **WHEN** the handler returns a `4xx` or `5xx` response
- **THEN** the span has an `http.status_code` attribute matching the response code
