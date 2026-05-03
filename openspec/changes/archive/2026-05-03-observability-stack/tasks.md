## 1. Config

- [x] 1.1 Add `ObsConfig` struct to `internal/config/config.go` with `OTELExporter string` (required), `OTELMetricsExporter string` (required), and `LogFormat string` (optional, default `"console"`)
- [x] 1.2 Wire `Obs ObsConfig` into `Config` struct and load from env via `config.Load()`
- [x] 1.3 Update config unit tests to cover `OTEL_EXPORTER` required validation, `OTEL_METRICS_EXPORTER` required validation, and `LOG_FORMAT` default

## 2. Logging

- [x] 2.1 Add `go.uber.org/zap` to `go.mod`
- [x] 2.2 Create `internal/observability/logging.go` — `NewLogger(format string) (*zap.Logger, error)` switching on `"json"` (production) vs `"console"` (development)
- [x] 2.3 Implement `LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger` — extracts OTel trace ID from context and returns a child logger with a `trace_id` field; returns base logger unchanged if no active span

## 3. Tracing

- [x] 3.1 Add OTel trace dependencies to `go.mod`: `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk`, `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`, GCP Cloud Trace exporter `github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace`
- [x] 3.2 Create `internal/observability/tracing.go` — `NewTracerProvider(ctx context.Context, exporter string) (*sdktrace.TracerProvider, error)` switching on `"jaeger"` and `"gcp"`; registers global TracerProvider and W3C propagator
- [x] 3.3 Implement `"jaeger"` case — OTLP gRPC exporter to `OTEL_JAEGER_ENDPOINT` (default `localhost:4317`)
- [x] 3.4 Implement `"gcp"` case — GCP Cloud Trace exporter using ADC and `GOOGLE_CLOUD_PROJECT`
- [x] 3.5 Return an error for unrecognised exporter values

## 4. Metrics

- [x] 4.1 Add OTel metrics dependencies to `go.mod`: `go.opentelemetry.io/otel/sdk/metric`, `go.opentelemetry.io/otel/exporters/prometheus`, GCP Cloud Monitoring exporter `github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric`
- [x] 4.2 Create `internal/observability/metrics.go` — `NewMeterProvider(exporter string) (*sdkmetric.MeterProvider, http.Handler, error)` switching on `"prometheus"` and `"gcp"`; registers global MeterProvider
- [x] 4.3 Implement `"prometheus"` case — Prometheus exporter with default reader; return `promhttp.Handler()` as the second return value
- [x] 4.4 Implement `"gcp"` case — Cloud Monitoring push exporter with 60-second periodic reader using ADC and `GOOGLE_CLOUD_PROJECT`; return `nil` handler
- [x] 4.5 Return an error for unrecognised metrics exporter values

## 5. Echo Middleware

- [x] 5.1 Create `internal/observability/echo_middleware.go` — `TraceMiddleware(tracer trace.Tracer) echo.MiddlewareFunc`
- [x] 5.2 Extract incoming W3C `traceparent`/`tracestate` headers using the global propagator
- [x] 5.3 Start a span named after the matched Echo route pattern; store span context on the request context
- [x] 5.4 Record `http.status_code` as a span attribute after the handler returns; end the span
- [x] 5.5 Add `MetricsMiddleware(meter metric.Meter) echo.MiddlewareFunc` to the same file
- [x] 5.6 Implement `http.server.request.duration` Float64 histogram — record elapsed milliseconds with `http.request.method`, `http.route`, `http.response.status_code` attributes
- [x] 5.7 Implement `http.server.active_requests` Int64 updown counter — increment on request start, decrement after handler returns

## 6. Health Handler

- [x] 6.1 Create `internal/handler/health.go` — `HealthHandler` struct accepting `*gorm.DB` and `storage.StorageService`
- [x] 6.2 Add `Ping(ctx context.Context) error` to `StorageService` interface and implement as `HeadBucket` in `internal/storage/r2.go`
- [x] 6.3 Implement `GET /health` handler: probe DB with `SELECT 1` and R2 via `Ping`, both under a 3-second context deadline
- [x] 6.4 Return `200 OK` with `{"status":"ok"|"degraded","db":"ok"|<error>,"r2":"ok"|<error>}` in all cases
- [x] 6.5 Write unit tests for `HealthHandler` covering all-healthy, DB failure, and R2 failure scenarios

## 7. Main Wiring

- [x] 7.1 In `cmd/server/main.go`, initialise logger via `observability.NewLogger(cfg.Obs.LogFormat)` and defer `logger.Sync()`
- [x] 7.2 Initialise tracer provider via `observability.NewTracerProvider(ctx, cfg.Obs.OTELExporter)` and defer `provider.Shutdown(ctx)`
- [x] 7.3 Initialise meter provider via `observability.NewMeterProvider(cfg.Obs.OTELMetricsExporter)` and defer `provider.Shutdown(ctx)`
- [x] 7.4 Register `TraceMiddleware` and `MetricsMiddleware` on the Echo instance (before route groups)
- [x] 7.5 If the metrics handler is non-nil (Prometheus case), register `GET /metrics` on Echo outside the protected group
- [x] 7.6 Wire `HealthHandler` with `*gorm.DB` and `storage.StorageService`; register `GET /health` outside the protected group

## 8. Docker Compose & Provisioning

- [x] 8.1 Create `docker-compose.yml` at repo root with `app`, `jaeger` (`jaegertracing/all-in-one:latest`), `prometheus` (`prom/prometheus:latest`), and `grafana` (`grafana/grafana:latest`) services
- [x] 8.2 Configure `app` service: build from `./backend`, env vars from `.env` file plus `OTEL_EXPORTER=jaeger`/`OTEL_METRICS_EXPORTER=prometheus`/`LOG_FORMAT=console`/`OTEL_JAEGER_ENDPOINT=jaeger:4317`, port `8080`, `extra_hosts: ["host.docker.internal:host-gateway"]`
- [x] 8.3 Configure `jaeger` service: expose ports `16686` (UI) and `4317` (OTLP gRPC)
- [x] 8.4 Configure `prometheus` service: mount `./prometheus/prometheus.yml`; expose port `9090`
- [x] 8.5 Configure `grafana` service: expose port `3000`; mount `./grafana/provisioning` as a volume
- [x] 8.6 Create `prometheus/prometheus.yml` with a scrape job targeting `app:8080/metrics`
- [x] 8.7 Create `grafana/provisioning/datasources/jaeger.yml` with Jaeger datasource at `http://jaeger:16686`
- [x] 8.8 Create `grafana/provisioning/datasources/prometheus.yml` with Prometheus datasource at `http://prometheus:9090`
