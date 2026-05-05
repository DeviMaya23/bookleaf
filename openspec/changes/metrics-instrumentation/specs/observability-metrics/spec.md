## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: DB Connection Pool Metrics via otelgorm

The system SHALL register the `github.com/go-gorm/opentelemetry` plugin on the GORM `*gorm.DB` instance in `cmd/server/main.go` after `gorm.Open`, via `db.Use(otelgorm.NewPlugin())`. The plugin SHALL automatically emit the following connection pool metrics using the global `MeterProvider`:

- `go.sql.connections_open` — observable gauge, total established connections
- `go.sql.connections_in_use` — observable gauge, connections currently in use
- `go.sql.connections_idle` — observable gauge, idle connections
- `go.sql.connections_max_open` — observable gauge, maximum allowed open connections
- `go.sql.connections_wait_count` — observable counter, total wait operations
- `go.sql.connections_wait_duration` — observable counter, total time blocked waiting (nanoseconds)
- `go.sql.connections_closed_max_idle` — observable counter, connections closed due to `SetMaxIdleConns`
- `go.sql.connections_closed_max_idle_time` — observable counter, connections closed due to `SetConnMaxIdleTime`
- `go.sql.connections_closed_max_lifetime` — observable counter, connections closed due to `SetConnMaxLifetime`

The plugin SHALL also produce an OTel trace span per GORM operation with `db.operation.name` and `db.collection.name` attributes.

#### Scenario: DB connection pool metrics are exported

- **WHEN** the application is running and `OTEL_METRICS_EXPORTER=prometheus`
- **THEN** `GET /metrics` includes `go_sql_connections_open`, `go_sql_connections_in_use`, and `go_sql_connections_idle` gauges

#### Scenario: GORM operations produce trace spans

- **WHEN** any GORM database operation executes
- **THEN** a trace span is created with `db.operation.name` set to the SQL operation type and `db.collection.name` set to the table name
