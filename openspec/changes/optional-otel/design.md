## Context

OTel is always initialised at startup. `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` are required env vars, and `NewTracerProvider` / `NewMeterProvider` are unconditionally called in `main.go`. This makes local development and CI environments require a running OTel collector. The `Telemetry` struct already supports no-op construction via `NewTelemetry(nil, nil, nil)`, so the plumbing to support a disabled state exists — it just isn't wired up.

## Goals / Non-Goals

**Goals:**
- `OTEL_ENABLED=true` preserves the current behaviour exactly
- `OTEL_ENABLED` unset or `false` skips all OTel initialisation and routes all instrumentation through no-op providers at zero cost
- OTel-specific env vars (`OTEL_EXPORTER`, `OTEL_METRICS_EXPORTER`) are not validated when OTel is disabled — no startup error if they are absent

**Non-Goals:**
- Changing how any existing instrumentation is written in handlers, usecases, or storage layers
- Supporting per-signal toggling (e.g. traces on, metrics off)
- Removing OTel as a dependency from the binary

## Decisions

### Decision 1: Single `OTEL_ENABLED` bool flag in `ObsConfig`

`ObsConfig` gains `OTELEnabled bool` loaded from `OTEL_ENABLED` via `envWithDefault("OTEL_ENABLED", "false")`. A `"true"` string maps to `true`; anything else is `false`.

`OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` move from `requireEnv` to `envWithDefault(..., "")`. When `OTELEnabled` is true, `loadFromEnv` validates they are non-empty and returns an error if absent. When false, they are ignored regardless of value.

**Alternative considered:** separate `OTEL_TRACING_ENABLED` / `OTEL_METRICS_ENABLED` flags. Rejected — adds complexity with no current need; a single toggle matches the deployment model (observability stack is either present or not).

### Decision 2: Conditional provider init block in `main.go`

`main.go` wraps `NewTracerProvider`, `NewMeterProvider`, `TraceMiddleware`, `MetricsMiddleware`, `db.Use(otelgorm.NewPlugin())`, and the `/metrics` route registration in a single `if cfg.Obs.OTELEnabled { ... }` block. When disabled, `NewTelemetry(nil, nil, nil)` is called directly, which substitutes no-op logger, tracer, and meter internally.

**Alternative considered:** initialise real providers always but point them at a no-op exporter. Rejected — this still requires the exporter env vars and adds unnecessary provider overhead.

### Decision 3: GORM OTel plugin skipped when disabled

`db.Use(otelgorm.NewPlugin())` is guarded by the same `OTELEnabled` check. When disabled the plugin is never registered, so no SQL spans or DB metrics are emitted. This is the correct behaviour — the plugin reads from the global OTel provider, and while a no-op global provider would technically be safe, skipping registration is cleaner and avoids any plugin overhead.

## Risks / Trade-offs

- **Risk: `OTEL_ENABLED` defaults to false, potentially surprising operators who expect the current always-on behaviour** → Mitigation: document the flag in `.env.example`; existing deployments that already set `OTEL_EXPORTER` should also set `OTEL_ENABLED=true`.
- **Risk: `otelgorm` import remains in `main.go` even when disabled** → No mitigation needed; unused-at-runtime imports are fine, and removing the import conditionally would require build tags, which is overkill.

## Migration Plan

1. Deploy with `OTEL_ENABLED=true` on any environment that currently has `OTEL_EXPORTER` set — behaviour is identical to today.
2. Local / CI environments can omit all OTel vars and run without a collector.
3. No rollback complexity — the flag is purely additive.
