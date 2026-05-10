## ADDED Requirements

### Requirement: Conditional OTel Initialisation in Server Bootstrap

When `OTEL_ENABLED=true`, `cmd/server/main.go` SHALL initialise `TracerProvider` and `MeterProvider`, register `TraceMiddleware`, `MetricsMiddleware`, the GORM OTel plugin, and the `/metrics` route exactly as today.

When `OTEL_ENABLED` is unset or `false`, `cmd/server/main.go` SHALL skip all of the above and call `observability.NewTelemetry(nil, nil, nil)` directly, routing all instrumentation through no-op providers. No OTel-related env vars are read or validated in this path.

#### Scenario: OTel disabled — server starts without OTEL env vars

- **WHEN** `OTEL_ENABLED` is not set (or set to `false`)
- **AND** `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` are absent
- **THEN** the server starts successfully
- **AND** no tracer provider or meter provider is initialised
- **AND** the `/metrics` endpoint is not registered

#### Scenario: OTel enabled — full provider initialisation

- **WHEN** `OTEL_ENABLED=true` and `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` are set
- **THEN** `TracerProvider` and `MeterProvider` are initialised as before
- **AND** `TraceMiddleware`, `MetricsMiddleware`, and the GORM OTel plugin are registered
- **AND** the `/metrics` endpoint is available

#### Scenario: OTel disabled — GORM plugin not registered

- **WHEN** `OTEL_ENABLED=false`
- **THEN** `db.Use(otelgorm.NewPlugin())` is NOT called
- **AND** no SQL spans or DB metrics are emitted

#### Scenario: OTel disabled — logger remains active, tracer and meter are no-ops

- **WHEN** `OTEL_ENABLED=false`
- **THEN** `tel.Logger` is the real Zap logger and structured log output is emitted to console as normal
- **AND** calls to `tel.Tracer.Start(...)` or `tel.Meter.Int64Counter(...)` complete without error or panic
- **AND** no spans or metrics are exported
