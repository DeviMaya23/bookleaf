# observability-tracing Specification

## Purpose
Defines OpenTelemetry distributed tracing requirements including TracerProvider initialisation and Echo request tracing middleware.

## Requirements

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
### Requirement: Per-Layer Child Span Creation

Every handler method, usecase method, and storage method that performs meaningful work SHALL create a child span at its entry point. The span SHALL be created from the `Tracer` field of the injected `*observability.Telemetry`:

```go
ctx, span := tel.Tracer.Start(ctx, "package.Function")
defer span.End()
```

`span.End()` SHALL always be deferred immediately after `Start` so that early returns and panics do not leak open spans.

The updated `ctx` carrying the child span SHALL be passed to all downstream calls (usecase calls from handlers, repository and storage calls from usecases) so that the trace tree is correctly linked.

Layers in scope:
- `internal/handler/` — all public handler methods
- `internal/usecase/` — all public usecase methods
- `internal/storage/r2.go` — `GeneratePresignedPutURL`, `GeneratePresignedGetURL`, `PutObject`, `GetObject`

SQL repository methods are out of scope for this change.

#### Scenario: Handler creates a child span linked to the HTTP trace

- **WHEN** a handler method is invoked and the request context already carries a span from `TraceMiddleware`
- **THEN** `tel.Tracer.Start(ctx, ...)` creates a child span whose parent is the HTTP boundary span
- **AND** the child span appears nested under the route span in Jaeger

#### Scenario: Usecase span is a child of the handler span

- **WHEN** a handler calls a usecase method, passing the span-carrying `ctx`
- **THEN** the usecase span appears nested under the handler span in the trace

#### Scenario: Deferred End is called on early return

- **WHEN** a handler or usecase method returns early due to a validation error before completing all work
- **THEN** `span.End()` is still called (via defer)
- **AND** the span appears in Jaeger with a short duration reflecting the early exit

### Requirement: Span Naming Convention

Spans SHALL be named using the format `package.Function`, where:
- `package` is the Go package name (e.g. `handler`, `usecase`, `storage`)
- `Function` is the exported method name on the type (e.g. `InitiateUpload`, `ListImages`)

Examples:
- `handler.InitiateImageUpload`
- `usecase.CompleteUpload`
- `storage.GeneratePresignedPutURL`

Generic names (`"upload"`, `"handler"`, `"db"`) SHALL NOT be used.

#### Scenario: Span name is locatable in source

- **WHEN** a span named `usecase.CompleteUpload` appears in Jaeger
- **THEN** a developer can locate the span origin by searching for `"usecase.CompleteUpload"` in the codebase

### Requirement: Error Recording on Spans

When a method returns a non-nil error, the span SHALL be marked with the error before returning:

```go
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
    return err
}
```

Both `RecordError` and `SetStatus` SHALL be called together. Calling only one is not sufficient — `RecordError` adds the error as a span event, while `SetStatus` marks the span as failed in the trace UI.

Business-logic errors (e.g. `ErrNotFound`, `ErrUnauthorized`) SHALL be recorded on the span in the same way as infrastructure errors.

#### Scenario: Infrastructure error is recorded on span

- **WHEN** a usecase calls a repository and the repository returns an error
- **THEN** `span.RecordError(err)` and `span.SetStatus(codes.Error, err.Error())` are called on the usecase span
- **AND** the span appears as an error span in Jaeger

#### Scenario: Business logic error is recorded on span

- **WHEN** a handler detects an unauthorised access attempt and returns an error
- **THEN** the handler span is marked as an error span with the reason recorded

### Requirement: Trace Context Propagation Through Layers

The `ctx` value returned by `tel.Tracer.Start` SHALL be threaded through all downstream calls within the same request. Handlers SHALL pass the span-enriched `ctx` to usecase calls; usecases SHALL pass it to repository and storage calls.

This propagation is additive to the W3C `traceparent` propagation already implemented in `TraceMiddleware` — the middleware seeds the root span, and each layer adds child spans by using the `ctx` it receives.

#### Scenario: Full trace shows all layers for a single request

- **WHEN** `POST /images` is called end-to-end
- **THEN** the Jaeger trace for that request contains a root HTTP span from `TraceMiddleware`, a child `handler.InitiateImageUpload` span, a child `usecase.InitiateUpload` span, and a child `storage.GeneratePresignedPutURL` span
- **AND** all spans share the same trace ID
