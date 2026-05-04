## ADDED Requirements

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
