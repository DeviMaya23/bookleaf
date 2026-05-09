## Why

OTel instrumentation is currently always-on and requires `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` to be set, making it impossible to run the server locally or in test environments without a full observability stack. Introducing an `OTEL_ENABLED` flag makes telemetry opt-in so the app works out of the box without an OTel collector.

## What Changes

- `ObsConfig` gains an `OTELEnabled bool` field loaded from `OTEL_ENABLED` (optional, defaults to `false`)
- `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` are only validated as required when `OTEL_ENABLED=true`; they are skipped entirely otherwise
- `cmd/server/main.go` conditionally initialises `TracerProvider` and `MeterProvider` only when OTel is enabled; when disabled, `NewTelemetry` is called with nil arguments so all instrumentation is routed through no-op providers
- `TraceMiddleware`, `MetricsMiddleware`, and the GORM OTel plugin are only registered when OTel is enabled
- The `/metrics` Prometheus endpoint is only registered when OTel is enabled

## Capabilities

### New Capabilities

- `optional-otel`: Controls whether OTel SDK providers and related middleware are initialised based on the `OTEL_ENABLED` env var

### Modified Capabilities

- `app-config`: `ObsConfig` requirement changes — `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` are conditionally required only when `OTEL_ENABLED=true`

## Impact

- `backend/internal/config/config.go` — `ObsConfig` struct and `loadFromEnv` logic
- `backend/internal/config/config_test.go` — new scenarios for OTel-disabled path
- `backend/cmd/server/main.go` — conditional OTel setup block
