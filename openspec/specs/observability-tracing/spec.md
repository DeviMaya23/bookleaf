## ADDED Requirements

### Requirement: GORM SQL Query Tracing via Plugin

The system SHALL register the `go-gorm/opentelemetry` plugin on the `*gorm.DB` instance once at startup via `db.Use(otelgorm.NewPlugin())` in `cmd/server/main.go`, after `gorm.Open`. All repositories share the same `*gorm.DB` and automatically inherit the plugin without per-file changes.

The plugin SHALL create child spans named `gorm.{operation}` (e.g. `gorm.Create`, `gorm.Query`) using the global OTel tracer. These spans SHALL be children of whatever span is in the request context at the time of the GORM call, placing them naturally under the calling usecase span in the trace tree.

No per-repository span creation is required — the plugin handles all SQL-layer spans globally.

#### Scenario: SQL query span appears under usecase span

- **WHEN** a usecase method calls a repository method that executes a SQL query
- **THEN** the trace contains a `gorm.Query` (or similar) child span under the usecase span
- **AND** the span duration reflects the actual query execution time

#### Scenario: GORM plugin is registered at startup and all repositories inherit it

- **WHEN** the server starts and `db.Use(plugin)` is called once
- **THEN** SQL spans are emitted for all subsequent GORM calls across all repositories without any per-repository instrumentation

#### Scenario: GORM error is recorded on the SQL span

- **WHEN** a GORM call returns an error (e.g. record not found, constraint violation)
- **THEN** the `gorm.*` span is marked as an error span with the error recorded

---

### Requirement: Zap Bridge for GORM Logger

The system SHALL include a zap adapter that implements GORM's `logger.Interface` and delegates log output to the application's `*zap.Logger`. The adapter SHALL be passed to `otelgorm.NewPlugin()` as its logger so that GORM diagnostic output flows through the existing structured log pipeline.

Level mapping:
- GORM `Info` → `zap.Debug` (verbose GORM info is noise at INFO level)
- GORM `Warn` → `zap.Warn`
- GORM `Error` → `zap.Error`
- GORM `Trace` (slow query) → `zap.Warn` with `elapsed_ms`, `rows_affected`, and `sql` (parameterised query string, not values)
- GORM `Trace` (error) → `zap.Error` with `elapsed_ms`, `rows_affected`, `sql`, and `error`

The slow-query threshold (above which `Trace` emits at `Warn`) SHALL default to 200ms and SHALL be configurable.

Raw SQL parameter values SHALL NOT appear in log output to prevent accidental credential or PII exposure.

The `Trace` method receives a `context.Context`. The bridge SHALL call `LoggerFromContext(ctx, base)` before logging slow query and error entries so that those log lines carry a `trace_id` field. This ensures Grafana's trace→log correlation button finds the GORM log lines when navigating from a `gorm.*` span in Tempo.

#### Scenario: Slow query emits a Warn log with trace_id

- **WHEN** a GORM query exceeds the slow-query threshold (200ms default)
- **THEN** a WARN log is emitted with `elapsed_ms`, `rows_affected`, and the parameterised `sql` string
- **AND** the log entry includes a `trace_id` field matching the active span's trace ID
- **AND** no parameter values appear in the log entry

#### Scenario: GORM error emits an Error log with trace_id

- **WHEN** GORM encounters a database error during a query
- **THEN** an ERROR log is emitted with `elapsed_ms`, `rows_affected`, `sql`, `error`, and `trace_id`

#### Scenario: GORM slow query log is findable via trace→log correlation

- **WHEN** a slow GORM query occurs within a traced request
- **THEN** clicking the Logs button on the `gorm.*` span in Grafana Tempo opens Loki results containing the slow query log line for that trace ID
