## Context

The backend is a Go + Echo HTTP server with no structured logging, no distributed tracing, no metrics, and a trivial `/health` endpoint. As the app moves toward production on GCP, we need the full observability triad (logs, traces, metrics) working identically in both environments, with the only difference being which sink each signal goes to. The two environments are:

- **Local**: Grafana stack (Grafana + Jaeger + Prometheus) running in Docker Compose; app also containerised; PostgreSQL runs natively on the host machine
- **Production (GCP)**: Cloud Logging (structured JSON on stdout), Cloud Trace (spans via GCP exporter), Cloud Monitoring (metrics via GCP exporter)

The key constraint is that application code must not import environment-specific exporters directly — all wiring happens at startup in `main.go`.

## Goals / Non-Goals

**Goals:**
- Structured logging with Zap; format switchable via `LOG_FORMAT` env var (`console` / `json`)
- Distributed tracing via OpenTelemetry SDK; exporter switchable via `OTEL_EXPORTER` env var (`jaeger` / `gcp`)
- Trace ID injected as a field into every log line for log–trace correlation
- Echo middleware propagates trace context; each HTTP request gets a root span
- Metrics via OpenTelemetry Metrics SDK; exporter switchable via `OTEL_METRICS_EXPORTER` env var (`prometheus` / `gcp`); HTTP request rate and latency recorded via metrics middleware
- Docker Compose file for local: app, Grafana, Jaeger, Prometheus; DB via `host.docker.internal`
- `/health` returns structured JSON with `db` and `r2` component probes

**Non-Goals:**
- Log shipping from Docker (Loki) — out of scope for the app itself; Grafana stack reads container stdout
- Baggage propagation or custom span attributes beyond the request span
- Alerting rules

## Decisions

### Decision 1: Logger passed as a dependency, not a global

**Choice**: `*zap.Logger` is constructed in `main.go` and passed to handlers/middleware that need it. No `zap.ReplaceGlobals`.

**Rationale**: Global state makes tests harder to isolate and obscures dependencies. Passing the logger explicitly is consistent with how the existing codebase handles `*gorm.DB` and `storage.StorageService`.

**Alternative considered**: `zap.L()` global — rejected because tests would share logger state and it makes the dependency invisible.

---

### Decision 2: TracerProvider constructed in `main.go`, exporter chosen by env var

**Choice**: `observability/tracing.go` exposes `NewTracerProvider(ctx, exporter string) (*sdktrace.TracerProvider, error)`. It switches on the `OTEL_EXPORTER` value and returns a provider with the appropriate exporter wired in. The provider is registered as the global OTel provider.

**Rationale**: The switch lives in one place. Adding a third exporter (e.g., OTLP to a generic collector) means adding one case. Application code only calls `otel.Tracer(...)` — it never imports a specific exporter package.

**Alternative considered**: Separate packages per exporter — rejected as over-engineered for two environments.

---

### Decision 3: Trace ID injected via a Zap logger wrapper, not a separate middleware

**Choice**: A helper `LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger` extracts the current span's trace ID and returns a child logger with a `trace_id` field. Handlers call this helper when they want a context-scoped logger.

**Rationale**: Middleware-injected logger-per-request (storing `*zap.Logger` in Echo context) adds coupling. The helper approach is pull-based: code only pays the cost when it asks for a logger.

**Alternative considered**: Storing logger in Echo context via middleware — rejected because it requires handlers to know the context key and creates implicit coupling.

---

### Decision 4: Echo tracing middleware wraps each request in a span

**Choice**: A custom Echo middleware (`observability/echo_middleware.go`) calls `otel.Tracer("bookleaf").Start(r.Context(), routePattern)` before the next handler and ends the span after. It propagates W3C `traceparent` headers from incoming requests.

**Rationale**: All inbound HTTP requests are traced without any handler-level changes. Route pattern (not URL) is used as the span name to avoid high-cardinality span names from path params.

**Alternative considered**: `otelecho` contrib package — viable, but the custom middleware gives control over span naming and avoids pulling in a contrib dependency for minimal logic.

---

### Decision 5: Health check probes DB and R2 with a timeout

**Choice**: `GET /health` runs a `db.WithContext(ctx).Exec("SELECT 1")` and a `store.GeneratePresignedGetURL(ctx, "health-check", 1*time.Second)` (or equivalent cheap R2 call) with a 3-second timeout. Returns `200 OK` with `{"status":"ok","db":"ok","r2":"ok"}` if both pass; returns `200 OK` with per-component `"error"` strings if either fails (does not return a non-2xx so load balancers stay green).

**Rationale**: Returning 200 even on component failure keeps the process in rotation — a DB blip should not cause a full restart. The structured body lets uptime monitors parse component status separately.

**Alternative considered**: Return 503 on any failure — rejected because transient DB hiccups would trigger unnecessary restarts in Cloud Run.

---

### Decision 6: Docker Compose targets local development only

**Choice**: `docker-compose.yml` at the repo root defines four services: `app` (built from `./backend`), `grafana`, `jaeger` (all-in-one), `prometheus`. The `app` service sets `OTEL_EXPORTER=jaeger`, `OTEL_METRICS_EXPORTER=prometheus`, `LOG_FORMAT=console`, and uses `host.docker.internal:5432` for `DATABASE_URL`. No DB container.

**Rationale**: DB running natively avoids a second Postgres instance and matches the user's current workflow. `host.docker.internal` is available on Docker Desktop for Mac/Windows; Linux needs `--add-host=host-gateway`.

**Alternative considered**: DB in Docker Compose — rejected per user requirement.

---

### Decision 7: MeterProvider follows the same factory pattern as TracerProvider

**Choice**: `internal/observability/metrics.go` exposes `NewMeterProvider(exporter string) (*sdkmetric.MeterProvider, error)`. It switches on `OTEL_METRICS_EXPORTER`:
- `"prometheus"` — creates a Prometheus exporter (`go.opentelemetry.io/otel/exporters/prometheus`) and registers a `/metrics` HTTP handler on a dedicated Echo route (outside the protected group)
- `"gcp"` — creates a GCP Cloud Monitoring push exporter with a periodic reader (60s interval)

The provider is registered as the global OTel `MeterProvider`. The caller defers `provider.Shutdown(ctx)`.

**Rationale**: Same pattern as `NewTracerProvider` — one factory, all environment-specific code in one switch, application code only calls `otel.Meter(...)`. Consistent, easy to extend.

**Alternative considered**: Using `OTEL_METRICS_EXPORTER` as the standard OTel SDK env var — the SDK auto-configures exporters via this var in some languages, but Go's SDK does not auto-configure; explicit factory is required.

---

### Decision 8: HTTP metrics recorded via a dedicated metrics middleware

**Choice**: A separate `MetricsMiddleware(meter metric.Meter) echo.MiddlewareFunc` in `internal/observability/echo_middleware.go` records two instruments per request:
- `http.server.request.duration` — histogram (milliseconds), attributes: `http.method`, `http.route`, `http.status_code`
- `http.server.active_requests` — updown counter, incremented on request start and decremented on completion

**Rationale**: Keeping metrics middleware separate from trace middleware means each can be registered independently. Using OTel semantic convention attribute names ensures compatibility with Grafana dashboards built on OTel conventions.

**Alternative considered**: Recording metrics inside the trace middleware — rejected because it couples two concerns and makes each harder to test in isolation.

## Risks / Trade-offs

- **GCP Cloud Trace and Cloud Monitoring exporters require ADC** — Application Default Credentials must be configured in production (`GOOGLE_APPLICATION_CREDENTIALS` or Workload Identity). The app will start but export silently fails without them. → Mitigation: document in README; exporters log errors on failed export.
- **Prometheus `/metrics` endpoint is unauthenticated** — Exposes internal counters without auth. Acceptable for an internal service; can be locked down at the load balancer level if needed. → Mitigation: register the route but note it should not be publicly routed in prod.
- **`host.docker.internal` on Linux** — Not available by default; requires `extra_hosts: ["host.docker.internal:host-gateway"]` in the Compose service. → Mitigation: add it to the `app` service definition unconditionally (Docker Desktop ignores it on Mac/Windows).
- **R2 health check costs a presign call** — Presigning is local computation, no network call, so it is effectively free. → No mitigation needed.
- **OTel SDK version churn** — The OTel Go SDK has had frequent API breaks. Pin to a specific minor version and update deliberately. → Mitigation: explicit version pins in `go.mod`.

## Open Questions

- Should Grafana be pre-configured with a Jaeger datasource, a Prometheus datasource, and default dashboards, or just datasources? *(Lean towards datasources only to reduce provisioning complexity; dashboards can be imported manually.)*
- `OTEL_EXPORTER=gcp` requires the `cloud.google.com/go/trace` exporter — confirm GCP project ID env var naming (`GOOGLE_CLOUD_PROJECT` is standard).
- Should `http.server.active_requests` updown counter be included, or just the request duration histogram? *(Lean towards including it — low cost, high value for detecting concurrency spikes.)*
