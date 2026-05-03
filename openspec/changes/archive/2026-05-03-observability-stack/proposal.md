## Why

The app currently has no structured logging, no distributed tracing, no metrics, and a health endpoint that only signals the process is alive. Adding the full observability triad (logs, traces, metrics) now — before feature complexity grows — gives a consistent foundation for debugging and monitoring across local and production environments without changing application logic when switching between them.

## What Changes

- Add **Zap** as the structured logger; `console` format locally, `json` for production (GCP Cloud Logging parses structured JSON natively)
- Add **OpenTelemetry** tracing with a pluggable exporter interface: **Jaeger** locally (Grafana stack), **GCP Cloud Trace** in production
- Inject trace IDs into every Zap log line so logs and traces correlate in both environments
- Propagate trace context through Echo middleware so all request handlers carry a span
- Add **OpenTelemetry Metrics** SDK with a pluggable exporter: **Prometheus** locally (scraped by Prometheus, visualised in Grafana), **GCP Cloud Monitoring** in production; HTTP request rate and latency recorded automatically via metrics middleware
- Add `OTEL_EXPORTER`, `OTEL_METRICS_EXPORTER`, and `LOG_FORMAT` env vars to drive exporter and format selection
- Add a **Docker Compose** file for local development: app container + Grafana + Jaeger + Prometheus; PostgreSQL runs natively, reached via `host.docker.internal`
- Extend `GET /health` to probe DB and R2 connectivity and return a structured JSON body with per-component status

## Capabilities

### New Capabilities

- `observability-logging`: Zap logger initialisation, pluggable output format (console / JSON), trace ID field injection from context
- `observability-tracing`: OTel SDK bootstrap, `TracerProvider` with pluggable exporters (`jaeger`, `gcp`), Echo middleware to start and finish request spans
- `observability-metrics`: OTel Metrics SDK, `MeterProvider` with pluggable exporters (`prometheus`, `gcp`), Echo middleware to record HTTP request rate and latency
- `local-dev-stack`: Docker Compose file for the app + Grafana + Jaeger + Prometheus; DB access via `host.docker.internal`

### Modified Capabilities

- `app-config`: Add `ObsConfig` sub-struct for `OTEL_EXPORTER` (required), `OTEL_METRICS_EXPORTER` (required), and `LOG_FORMAT` (optional, default `console`)
- `server-bootstrap`: `GET /health` now returns a JSON body with `status`, `db`, and `r2` component checks instead of an empty 200

## Impact

- **New dependencies**: `go.uber.org/zap`, `go.opentelemetry.io/otel` SDK + exporters (`otlptracegrpc`, `go.opentelemetry.io/otel/exporters/prometheus`, GCP Cloud Trace + Cloud Monitoring exporters)
- **`cmd/server/main.go`**: initialise logger, tracer provider, and meter provider before Echo; pass to handlers/middleware
- **`internal/config/config.go`**: three new env vars loaded into `Config`
- **`internal/handler/health.go`** (new): health handler that pings DB and R2
- **`docker-compose.yml`** (new): app, Grafana, Jaeger services
- **No breaking changes** to existing endpoints or database schema
