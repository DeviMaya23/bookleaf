### Requirement: HTTP Metrics Middleware

The system SHALL provide a `MetricsMiddleware(meter metric.Meter) echo.MiddlewareFunc` in `internal/observability/echo_middleware.go`. The middleware SHALL record the following instruments on every HTTP request:

- `http.server.request.duration` — Float64 histogram (unit: `ms`); attributes: `http.request.method`, `http.route`, `http.response.status_code`
- `http.server.active_requests` — Int64 updown counter; incremented when a request starts, decremented when the handler returns; attributes: `http.request.method`, `http.route`
- `http.server.request.count` — Int64 counter; incremented once per completed request; attributes: `http.request.method`, `http.route`, `http.response.status_code`
- `http.server.request.errors` — Int64 counter; incremented only for `4xx` and `5xx` responses; attributes: `http.request.method`, `http.route`, `http.status_class` (`"4xx"` or `"5xx"`)

Attribute values SHALL use the matched Echo route pattern (e.g., `/images/:id`) not the raw URL path, to prevent high-cardinality label explosion. All four instruments SHALL be initialised once at middleware construction time.

#### Scenario: Successful request is measured

- **WHEN** an HTTP request completes successfully
- **THEN** `http.server.request.duration` is recorded with the correct method, route, and status code attributes
- **AND** `http.server.active_requests` returns to its pre-request value after the handler returns
- **AND** `http.server.request.count` is incremented by 1

#### Scenario: Failed request is measured

- **WHEN** a handler returns a `4xx` or `5xx` status code
- **THEN** `http.server.request.duration` is recorded with the actual status code as the `http.response.status_code` attribute
- **AND** `http.server.request.count` is incremented by 1
- **AND** `http.server.request.errors` is incremented with the appropriate `http.status_class`

### Requirement: GORM SQL Query Metrics via Plugin

The `go-gorm/opentelemetry` plugin registered at startup SHALL also emit OTel metrics for all SQL queries. No additional configuration beyond `db.Use(otelgorm.NewPlugin())` is required — the plugin emits metrics automatically using the global `MeterProvider`.

The plugin SHALL emit:
- **DB query duration histogram** — records execution time per SQL operation and table; attributes: SQL operation type (e.g. `SELECT`, `INSERT`, `UPDATE`, `DELETE`), table name
- **DB error counter** — incremented when a GORM call returns an error; attributes: SQL operation type, table name

Metric cardinality SHALL be bounded. Raw query text and parameter values SHALL NOT appear as metric attributes.

When `OTEL_METRICS_EXPORTER=prometheus`, these DB metrics SHALL be queryable in Prometheus alongside the existing HTTP metrics via the `/metrics` endpoint.

#### Scenario: DB query duration is recorded per operation

- **WHEN** a GORM query executes
- **THEN** a histogram observation is recorded with the query duration
- **AND** the attributes include the SQL operation type and table name

#### Scenario: DB error increments error counter

- **WHEN** a GORM call returns an error
- **THEN** the DB error counter is incremented with the relevant operation type and table name attributes

#### Scenario: DB metrics are queryable in Prometheus

- **WHEN** `OTEL_METRICS_EXPORTER=prometheus` and at least one GORM query has executed
- **THEN** `GET /metrics` returns DB-related metric families (e.g. query duration histograms) alongside HTTP metrics
