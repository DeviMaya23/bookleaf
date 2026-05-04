## Why

The local dev stack uses Jaeger for tracing and writes logs to stdout, but there is no log aggregation backend — making it impossible to navigate from a trace to its correlated logs or vice versa. The `trace_id` field is already present in every log line via `LoggerFromContext`, but it has nowhere to land.

## What Changes

- Replace the `jaeger` Docker Compose service with **Tempo** (Grafana's trace backend); Tempo accepts the same OTLP gRPC protocol, so the app-level exporter code changes minimally
- Add **Loki** to Docker Compose as the log aggregation backend
- Add **Promtail** to Docker Compose to collect container stdout and ship structured logs to Loki
- Switch `LOG_FORMAT` from `console` to `json` in the app service so Loki can parse individual fields (including `trace_id`)
- Replace the `jaeger` exporter case in `NewTracerProvider` with `tempo`; both use the same OTLP gRPC transport so the underlying exporter code is identical, only the case name and default endpoint env var change
- Update Grafana provisioning: replace the Jaeger datasource with Tempo; add a Loki datasource; configure bidirectional trace↔log linking so a trace in Tempo links to its Loki logs and a Loki log line's `trace_id` links back to the Tempo trace
- Add `promtail/promtail.yml` and `loki/loki.yml` config files to the repository

## Capabilities

### New Capabilities

- `log-aggregation`: Loki + Promtail in the local dev stack — structured JSON logs from all containers are collected by Promtail and stored in Loki, queryable by field in Grafana Explore

### Modified Capabilities

- `local-dev-stack`: Jaeger replaced with Tempo; Loki, Promtail added; `LOG_FORMAT` changed to `json`; Grafana provisioning updated to Tempo + Loki datasources with trace-log correlation
- `observability-tracing`: `"jaeger"` exporter case replaced with `"tempo"`; Jaeger references in scenarios updated to Tempo/Grafana
- `app-config`: `LOG_FORMAT` default changes from `console` to `json`; `OTEL_EXPORTER` accepted value `"jaeger"` replaced with `"tempo"`

## Impact

- `docker-compose.yml` — service additions and replacements
- `internal/observability/tracing.go` — new `tempo` case
- `grafana/provisioning/datasources/` — replace `jaeger.yml` with `tempo.yml`; add `loki.yml`
- New files: `promtail/promtail.yml`, `loki/loki.yml`
- `.env.example` — update `OTEL_EXPORTER` example value
