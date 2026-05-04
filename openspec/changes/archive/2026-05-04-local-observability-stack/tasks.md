## 1. Config Files

- [x] 1.1 Create `tempo/tempo.yml` — single-binary mode, OTLP gRPC receiver on port `4317`, local filesystem storage
- [x] 1.2 Create `loki/loki.yml` — single-process mode, filesystem storage, listen on port `3100`
- [x] 1.3 Create `promtail/promtail.yml` — Docker socket discovery, JSON pipeline stage to extract `trace_id`, push to `http://loki:3100/loki/api/v1/push`

## 2. Grafana Provisioning

- [x] 2.1 Delete `grafana/provisioning/datasources/jaeger.yml`
- [x] 2.2 Create `grafana/provisioning/datasources/tempo.yml` — Tempo datasource at `http://tempo:3200` with `tracesToLogsV2` block pointing at Loki datasource, using `trace_id` as correlation field
- [x] 2.3 Create `grafana/provisioning/datasources/loki.yml` — Loki datasource at `http://loki:3100` with `derivedFields` block matching `trace_id` and linking to Tempo datasource

## 3. Docker Compose

- [x] 3.1 Remove `jaeger` service; remove `OTEL_JAEGER_ENDPOINT` from app env
- [x] 3.2 Add `tempo` service — image `grafana/tempo:latest`, mount `./tempo/tempo.yml`, expose port `4317`
- [x] 3.3 Add `loki` service — image `grafana/loki:latest`, mount `./loki/loki.yml`, expose port `3100`
- [x] 3.4 Add `promtail` service — image `grafana/promtail:latest`, mount `./promtail/promtail.yml` and `/var/run/docker.sock` (read-only), depends on `loki`
- [x] 3.5 Update `app` service: change `OTEL_EXPORTER=tempo`, add `OTEL_TEMPO_ENDPOINT=tempo:4317`, change `LOG_FORMAT=json`; update `depends_on` to `tempo`, `prometheus`, `loki`

## 4. Application Code

- [x] 4.1 In `internal/observability/tracing.go`, replace the `"jaeger"` case with `"tempo"` — read endpoint from `OTEL_TEMPO_ENDPOINT` (default `localhost:4317`); same `otlptracegrpc` exporter code
- [x] 4.2 In `internal/observability/tracing.go`, add a `resource.NewWithAttributes` call setting `service.name=bookleaf` and pass it to `sdktrace.WithResource` on the provider (already added in a prior fix — verify it is present and uses the correct attribute key `"service.name"`)

## 5. Config Default

- [x] 5.1 In `internal/config/config.go`, change the `LOG_FORMAT` default from `"console"` to `"json"` in `envWithDefault("LOG_FORMAT", "console")`

## 6. Env Example

- [x] 6.1 In `.env.example`, change `OTEL_EXPORTER=jaeger` to `OTEL_EXPORTER=tempo`; replace `OTEL_JAEGER_ENDPOINT` with `OTEL_TEMPO_ENDPOINT=localhost:4317`
