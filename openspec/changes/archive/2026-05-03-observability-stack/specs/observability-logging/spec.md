## ADDED Requirements

### Requirement: Zap Logger Initialisation

The system SHALL initialise a `*zap.Logger` in `cmd/server/main.go` before any other component. The logger SHALL be constructed by a `NewLogger(format string) (*zap.Logger, error)` function in `internal/observability/logging.go`. When `format` is `"json"`, it SHALL use `zap.NewProduction()` config. When `format` is `"console"` (or any other value), it SHALL use `zap.NewDevelopment()` config. The function SHALL return an error if Zap fails to build.

#### Scenario: Production logger in JSON format

- **WHEN** `LOG_FORMAT=json`
- **THEN** `NewLogger("json")` returns a `*zap.Logger` configured with production JSON encoding
- **AND** log output is valid JSON with `level`, `ts`, `msg` fields

#### Scenario: Development logger in console format

- **WHEN** `LOG_FORMAT=console` (or `LOG_FORMAT` is unset, defaulting to `console`)
- **THEN** `NewLogger("console")` returns a `*zap.Logger` with human-readable console output

### Requirement: Trace ID Field Injection

The system SHALL provide a `LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger` function in `internal/observability/logging.go`. If the context carries an active OTel span with a valid trace ID, the function SHALL return a child logger with a `trace_id` string field appended. If no active span exists or the trace ID is invalid, it SHALL return the base logger unchanged.

#### Scenario: Active span in context

- **WHEN** a context with a valid OTel span is passed to `LoggerFromContext`
- **THEN** the returned logger includes a `trace_id` field matching the span's trace ID in hex format

#### Scenario: No span in context

- **WHEN** a context without an active span is passed to `LoggerFromContext`
- **THEN** the returned logger is the base logger with no `trace_id` field added
