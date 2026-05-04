## ADDED Requirements

### Requirement: Telemetry Struct

The system SHALL define a `Telemetry` struct in `internal/observability/telemetry.go` that bundles the three observability primitives into a single injectable dependency:

```go
type Telemetry struct {
    Logger *zap.Logger
    Tracer trace.Tracer
    Meter  metric.Meter
}
```

The package SHALL expose two constructors:

- `NewTelemetry(logger *zap.Logger, tracer trace.Tracer, meter metric.Meter) *Telemetry` — used in `cmd/server/main.go` after providers are initialised
- `NewNoopTelemetry() *Telemetry` — returns a `*Telemetry` backed by `zap.NewNop()`, `noop.NewTracerProvider().Tracer("")`, and `noop.NewMeterProvider().Meter("")`; used in unit tests to avoid nil panics without requiring real providers

#### Scenario: Production Telemetry is constructed from live providers

- **WHEN** `NewTelemetry` is called with a configured Zap logger, a real OTel TracerProvider tracer, and a real OTel MeterProvider meter
- **THEN** the returned `*Telemetry` has non-nil `Logger`, `Tracer`, and `Meter` fields

#### Scenario: Noop Telemetry is safe to use in tests

- **WHEN** `NewNoopTelemetry()` is called
- **THEN** the returned `*Telemetry` has non-nil fields that accept calls without panicking or emitting any output

### Requirement: Telemetry Injection into Application Layers

All application layer constructors that perform observable operations SHALL accept `*observability.Telemetry` as a parameter and store it as a struct field. The affected types are:

- `handler.ImageHandler`
- `handler.FolderHandler`
- `handler.MeHandler`
- `usecase.imageUsecase` (the concrete impl)
- `usecase.folderUsecase` (the concrete impl)
- `usecase.userUsecase` (the concrete impl)
- `storage.r2Storage`

`handler.HealthHandler` is excluded — it is a liveness probe with no domain operations.

`cmd/server/main.go` SHALL construct a single `*Telemetry` instance after providers are initialised and pass it to every affected constructor.

#### Scenario: ImageHandler is constructed with Telemetry

- **WHEN** `handler.NewImageHandler` is called with a valid `*Telemetry`
- **THEN** the handler stores the `*Telemetry` and uses it for spans and log events on every request

#### Scenario: Single Telemetry instance is shared across all layers in main

- **WHEN** `cmd/server/main.go` starts
- **THEN** one `*Telemetry` is constructed after `NewTracerProvider` and `NewMeterProvider` return
- **AND** the same pointer is passed to all handler, usecase, and storage constructors

### Requirement: Request Lifecycle Logging Middleware

The system SHALL provide a `LoggingMiddleware(tel *observability.Telemetry) echo.MiddlewareFunc` in `internal/observability/echo_middleware.go`. The middleware SHALL be registered on the **protected route group** (after the Kinde auth middleware) so that `user_id` is always resolvable from the JWT context.

On each request the middleware SHALL emit one INFO-level structured log entry with the following fields:

| Field | Source |
|---|---|
| `user_id` | Kinde user ID extracted from request context |
| `http.request.method` | HTTP method (OTel semconv) |
| `http.route` | Matched Echo route pattern, e.g. `/images/:id` (OTel semconv) |
| `http.response.status_code` | Response status code (OTel semconv) |
| `duration_ms` | Wall time in milliseconds from request start to handler return |
| `trace_id` | Injected via `LoggerFromContext` |

#### Scenario: Successful request is logged

- **WHEN** an authenticated request completes with `200 OK`
- **THEN** the middleware emits one INFO log with `user_id`, `http.request.method`, `http.route`, `http.response.status_code: 200`, and a positive `duration_ms`

#### Scenario: Error response is logged with correct status code

- **WHEN** a handler returns a `4xx` or `5xx` response
- **THEN** the middleware logs the actual status code in `http.response.status_code`
- **AND** the log level remains INFO (status-based severity is expressed through `http.response.status_code`, not log level)

#### Scenario: Middleware does not log unauthenticated routes

- **WHEN** a request reaches `/health` or `/metrics`
- **THEN** `LoggingMiddleware` does NOT emit a lifecycle log (those routes are outside the protected group)
