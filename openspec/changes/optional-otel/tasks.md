## 1. Config Changes

- [x] 1.1 Add `OTELEnabled bool` field to `ObsConfig` in `backend/internal/config/config.go`
- [x] 1.2 Change `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` from `requireEnv` to `envWithDefault(..., "")` in `loadFromEnv`
- [x] 1.3 After loading `OTELEnabled`, validate that `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` are non-empty only when `OTELEnabled` is `true`, returning a descriptive error if either is missing

## 2. Config Unit Tests

- [x] 2.1 Add test: `OTEL_ENABLED` unset → `cfg.Obs.OTELEnabled` is `false` and no error
- [x] 2.2 Add test: `OTEL_ENABLED=false` with `OTEL_EXPORTER` absent → no error, `cfg.Obs.OTELExporter` is `""`
- [x] 2.3 Add test: `OTEL_ENABLED=true` with `OTEL_EXPORTER` absent → error naming missing variable
- [x] 2.4 Add test: `OTEL_ENABLED=true` with `OTEL_METRICS_EXPORTER` absent → error naming missing variable
- [x] 2.5 Add test: `OTEL_ENABLED=true` with both exporter vars set → success, fields populated correctly

## 3. Server Bootstrap

- [x] 3.1 Wrap `NewTracerProvider`, `NewMeterProvider`, `TraceMiddleware`, `MetricsMiddleware`, `db.Use(otelgorm.NewPlugin())`, and the `/metrics` route registration in `if cfg.Obs.OTELEnabled { ... }` in `backend/cmd/server/main.go`
- [x] 3.2 Move `observability.NewTelemetry(...)` outside the block: call `NewTelemetry(logger, otel.Tracer("bookleaf"), otel.Meter("bookleaf"))` inside the block (where real providers are set), and `NewTelemetry(nil, nil, nil)` in the else branch
