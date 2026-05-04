## Context

The current local dev stack uses Jaeger for traces and stdout for logs. The `trace_id` field is already injected into every structured log line via `LoggerFromContext`, but logs are only visible in `docker compose logs` — there is no aggregation backend to query them by `trace_id`. As a result, trace→log and log→trace navigation is not possible.

Tempo, Loki, and Promtail are the standard Grafana-native backends for traces and logs. They integrate with the existing Grafana service and with each other natively, requiring only datasource provisioning config — no custom plugin or manual wiring.

## Goals / Non-Goals

**Goals:**
- Clickable trace→log navigation in Grafana: open a Tempo trace, click a button, land in Loki filtered to that `trace_id`
- Clickable log→trace navigation: click the `trace_id` value in a Loki log line, jump to the Tempo trace
- All structured log fields (`user_id`, `event`, `image_id`, etc.) queryable in Grafana Explore via Loki
- Minimal changes to application code

**Non-Goals:**
- Production log/trace infrastructure (this is local dev only)
- Alerting or dashboards
- Retaining the `jaeger` exporter case — it is removed entirely

## Decisions

### D1 — Tempo over Jaeger

Tempo accepts OTLP gRPC on port `4317`, the exact same protocol the app already sends. Replacing Jaeger with Tempo requires changing only the docker-compose service definition and the exporter case name in `tracing.go` — the underlying `otlptracegrpc` exporter code is identical.

The key reason to switch is Grafana-native integration: Tempo has a built-in `tracesToLogs` link config that queries Loki by `trace_id` with zero custom setup. Jaeger can be added as a Grafana datasource but doesn't support this native bidirectional linking.

### D2 — Loki + Promtail for log collection

Promtail is configured to scrape Docker container logs via the Docker socket (`/var/run/docker.sock`). It ships logs to Loki with a `container` label derived from the container name.

Loki stores logs and exposes a query API that Grafana queries. No log schema migration or index mapping is needed — Loki indexes labels and queries log content via LogQL.

### D3 — LOG_FORMAT must be json

Promtail ships raw log lines to Loki. With `LOG_FORMAT=console`, each line is a formatted human-readable string; Loki can store it but cannot parse individual fields like `trace_id` or `user_id` as queryable labels.

With `LOG_FORMAT=json`, each line is a JSON object. Promtail uses a `json` pipeline stage to extract `trace_id` (and optionally other fields) as Loki labels, making them filterable in LogQL and usable as the correlation key in Grafana datasource linking.

### D4 — Grafana datasource provisioning for trace↔log linking

Two provisioning files are added under `grafana/provisioning/datasources/`:

- `tempo.yml` — Tempo datasource with a `tracesToLogsV2` block pointing at the Loki datasource; uses `trace_id` as the correlation field
- `loki.yml` — Loki datasource with a `derivedFields` block that matches the `trace_id` JSON field in log lines and creates a link to the Tempo datasource

This is declarative config — no Grafana UI steps required after `docker compose up`.

The existing `jaeger.yml` provisioning file is deleted.

### D5 — Tempo configured in single-binary mode

`tempo/tempo.yml` configures Tempo in single-binary mode with an OTLP gRPC receiver on port `4317` and local filesystem storage. This is standard for local dev — no distributed components, no object storage.

## Risks / Trade-offs

- **Docker socket mount** — Promtail requires read access to `/var/run/docker.sock` to discover containers. This is standard for local dev stacks but should not be replicated in production.
- **Log volume** — all containers' stdout is shipped to Loki, including Prometheus, Grafana, Loki, and Tempo themselves. This is acceptable for local dev but produces noise. Promtail can be filtered to only the `app` container if preferred.
- **Tempo local storage** — Tempo uses filesystem storage in the container; traces are lost when the container is removed. For local dev this is fine.

## Migration Plan

1. Delete `grafana/provisioning/datasources/jaeger.yml`
2. Add `tempo/tempo.yml`, `loki/loki.yml`, `promtail/promtail.yml`
3. Add `grafana/provisioning/datasources/tempo.yml` and `loki.yml`
4. Update `docker-compose.yml`: remove `jaeger`, add `tempo`, `loki`, `promtail`; change app env `OTEL_EXPORTER=tempo`, `OTEL_TEMPO_ENDPOINT=tempo:4317`, `LOG_FORMAT=json`; update `depends_on`
5. Replace `"jaeger"` case with `"tempo"` in `tracing.go`; update default endpoint env var to `OTEL_TEMPO_ENDPOINT`
6. Update `.env.example`

No database migrations. No API changes. Rollback: revert the docker-compose and tracing.go changes.
