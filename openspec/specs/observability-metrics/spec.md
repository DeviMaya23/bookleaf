# observability-metrics Specification

## Purpose
Defines OpenTelemetry metrics requirements including MeterProvider initialisation, HTTP metrics middleware, and the Prometheus metrics endpoint.

## Requirements

### Requirement: MeterProvider Initialisation

The system SHALL provide a `NewMeterProvider(exporter string) (*sdkmetric.MeterProvider, error)` function in `internal/observability/metrics.go`. It SHALL switch on the `exporter` argument:

- `"prometheus"` — creates a Prometheus exporter via `go.opentelemetry.io/otel/exporters/prometheus` and wires it into a `sdkmetric.MeterProvider` with a default reader. Returns the provider and the `promhttp.Handler()` that the caller registers as the `/metrics` route.
- `"gcp"` — creates a GCP Cloud Monitoring push exporter (`github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric`) with a 60-second periodic reader using ADC and `GOOGLE_CLOUD_PROJECT`.

The function SHALL register the constructed provider as the global OTel `MeterProvider` via `otel.SetMeterProvider`. The caller is responsible for calling `provider.Shutdown(ctx)` on application exit. For the `"prometheus"` case, the function SHALL additionally return the `http.Handler` to be mounted at `/metrics`. For the `"gcp"` case it returns `nil` for the handler.

#### Scenario: Prometheus exporter selected

- **WHEN** `OTEL_METRICS_EXPORTER=prometheus`
- **THEN** `NewMeterProvider` returns a provider backed by the Prometheus exporter
- **AND** returns an `http.Handler` suitable for mounting at `/metrics`

#### Scenario: GCP Cloud Monitoring exporter selected

- **WHEN** `OTEL_METRICS_EXPORTER=gcp`
- **THEN** `NewMeterProvider` returns a provider that pushes metrics to GCP Cloud Monitoring every 60 seconds

#### Scenario: Unknown metrics exporter

- **WHEN** `OTEL_METRICS_EXPORTER` is set to an unrecognised value
- **THEN** `NewMeterProvider` returns an error naming the unrecognised value

### Requirement: HTTP Metrics Middleware

The system SHALL provide a `MetricsMiddleware(meter metric.Meter) echo.MiddlewareFunc` in `internal/observability/echo_middleware.go`. The middleware SHALL record the following instruments on every HTTP request:

- `http.server.request.duration` — Float64 histogram (unit: `ms`); attributes: `http.request.method`, `http.route`, `http.response.status_code`
- `http.server.active_requests` — Int64 updown counter; incremented when a request starts, decremented when the handler returns; attributes: `http.request.method`, `http.route`

Attribute values SHALL use the matched Echo route pattern (e.g., `/images/:id`) not the raw URL path, to prevent high-cardinality label explosion.

#### Scenario: Successful request is measured

- **WHEN** an HTTP request completes successfully
- **THEN** `http.server.request.duration` is recorded with the correct method, route, and status code attributes
- **AND** `http.server.active_requests` returns to its pre-request value after the handler returns

#### Scenario: Failed request is measured

- **WHEN** a handler returns a `4xx` or `5xx` status code
- **THEN** `http.server.request.duration` is recorded with the actual status code as the `http.response.status_code` attribute

### Requirement: Prometheus Metrics Endpoint

When `OTEL_METRICS_EXPORTER=prometheus`, the system SHALL expose a `GET /metrics` endpoint outside the protected route group. The endpoint SHALL serve the Prometheus text exposition format produced by `promhttp.Handler()`.

#### Scenario: Metrics endpoint accessible without auth

- **WHEN** `GET /metrics` is called without an Authorization header
- **THEN** the response is `200 OK` with `Content-Type: text/plain; version=0.0.4`
- **AND** the body contains OTel-instrumented counters and histograms in Prometheus format
