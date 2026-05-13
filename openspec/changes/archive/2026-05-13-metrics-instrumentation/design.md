## Context

Traces and structured logs are in place across HTTP, DB, and R2 layers. The `Telemetry` struct (`internal/observability/telemetry.go`) carries a `metric.Meter` and is already injected into `r2Storage` and `imageUsecase`, but neither uses it for metric instruments. The Echo `MetricsMiddleware` (`echo_middleware.go`) already instruments request duration and active requests but is missing request count and error counters. The `gorm.io/plugin/opentelemetry` package is already in `go.mod` and registered via `db.Use(otelgorm.NewPlugin())` in `main.go` as of the merged `feat/image-edit-repo-telemetry` PR. The custom `gorm_logger.go` zap bridge is also in place.

## Goals / Non-Goals

**Goals:**
- Add request count counter and 4xx/5xx error counters to `MetricsMiddleware`
- otelgorm plugin already registered — DB traces and connection pool metrics already in place
- Add histogram and counter instruments to `r2Storage` for presigned URL generation duration and upload success/failure count
- Add histogram and counter instruments to `imageUsecase` for thumbnail generation duration and success/failure

**Non-Goals:**
- Query duration histograms or error count metrics for DB (covered by trace spans)
- Instrumenting `GetObject` or `PutObject` as separate upload metrics (these are internal transfer operations; the user-visible event is `CompleteUpload`)
- Changing the Prometheus or GCP exporter setup

## Decisions

### 1. Use `github.com/go-gorm/opentelemetry` for DB instrumentation

The `github.com/go-gorm/opentelemetry` plugin will be added as a dependency and registered via `db.Use(otelgorm.NewPlugin())` in `main.go` after `gorm.Open`. It provides two things:

- **Traces** per query with `db.operation.name` and `db.collection.name` as span attributes — query duration and errors are observable via span duration and span error status
- **9 connection pool metrics** (gauges/counters): `go.sql.connections_open`, `go.sql.connections_in_use`, `go.sql.connections_idle`, `go.sql.connections_max_open`, and counters for wait time, wait count, and connections closed by policy

The plugin does not emit query duration histograms or query error counters as OTel metric instruments. Query duration per operation/table and error counts are covered by the trace spans and are out of scope for this change's metric instruments.

### 2. Instrument initialisation at constructor time

Following the existing `MetricsMiddleware` pattern, all instruments (histograms, counters) will be initialised once in the constructor (`NewR2Storage`, `NewImageUsecase`, the GORM plugin constructor). Errors from instrument creation are ignored with `_` because the OTel SDK returns noop instruments on failure rather than panicking — consistent with the existing middleware code.

Instruments will be stored as fields on their respective structs.

### 3. HTTP error counter uses status class attribute

Rather than a separate counter per status code, a single `http.server.request.errors` counter will use a `http.status_class` attribute with values `"4xx"` or `"5xx"`. This keeps cardinality low and makes alerting straightforward (one metric, filter by class).

The counter is incremented inside the existing `defer` block in `MetricsMiddleware` where `statusCode` is already resolved.

### 4. Upload success/failure counter lives in `imageUsecase.CompleteUpload`

The actual file transfer is a direct client-to-R2 presigned PUT that never passes through the server. `CompleteUpload` is the server-side acknowledgement and is the correct place to track upload success/failure count.

### 5. Presigned URL duration measured in `r2Storage`, not usecase

`GeneratePresignedPutURL` and `GeneratePresignedGetURL` are the canonical locations for this latency — they make the actual AWS SDK call. Measuring at the usecase layer would include unrelated logic.

### 6. Thumbnail metrics in `imageUsecase.uploadThumbnail`

This goroutine already calculates `time.Since(start)` for its log line. The histogram record and success/failure counter increment will replace/accompany the existing log — no new timing logic required.

## Risks / Trade-offs

- **Query duration/error not in metrics** — query duration and error counts are observable only via traces, not dashboards or alerting rules that query metric instruments. Accepted trade-off for this change.
- **Goroutine context loss for thumbnail metrics** — `uploadThumbnail` uses `context.Background()` rather than the request context, so thumbnail metrics will not carry trace context. This is pre-existing behaviour; do not change it here.

## Migration Plan

1. Extend `MetricsMiddleware` with new instruments — additive, existing instruments unchanged
2. Add instruments to `r2Storage` and `imageUsecase` constructors and record in relevant methods — additive

No migration or rollback strategy needed; all changes are purely additive metric instrumentation.

## Open Questions

- Should presigned GET URL generation be instrumented separately from presigned PUT? Both are included in scope for consistency.
