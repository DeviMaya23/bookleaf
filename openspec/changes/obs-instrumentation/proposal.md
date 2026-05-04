## Why

The observability stack (providers, middleware, exporters) is wired up but the application code is dark — handlers, usecases, and repositories emit no structured logs and create no child spans. Without instrumentation, traces show only the HTTP boundary span and logs contain no domain context, making it impossible to diagnose latency, auth failures, or R2 issues in production.

## What Changes

- Introduce a `Telemetry` struct (`internal/observability/telemetry.go`) that bundles `*zap.Logger`, `trace.Tracer`, and `metric.Meter` into a single dependency passed to handlers, usecases, and repositories that need observability
- Define structured logging events for: request lifecycle, auth events (new user creation, token rejection, unauthorised access), R2 operations (upload start/complete, presigned URL generation), thumbnail jobs (started, completed/failed), and folder/image mutations
- Define per-layer child span conventions: span created at the start of each handler, usecase, and repository function; named `package.Function`; errors recorded on span; `span.End()` always deferred

## Capabilities

### New Capabilities

- `obs-telemetry`: The `Telemetry` struct definition, constructor (`NewTelemetry`), and the contract for how it is passed to application layers

### Modified Capabilities

- `observability-logging`: Add domain event logging requirements — specific fields, log levels, and events that each layer must emit (beyond the existing logger initialisation and trace ID injection)
- `observability-tracing`: Add per-layer child span requirements — naming convention (`package.Function`), mandatory `defer span.End()`, error recording on span, and propagation from the HTTP middleware already in place

## Impact

- `internal/observability/` — new `telemetry.go` file; `logging.go` and `tracing.go` gain no new functions but their contracts expand
- `internal/handler/` — all handlers receive `Telemetry` and must create child spans + emit lifecycle/auth log events
- `internal/usecase/` — all usecases receive `Telemetry` and must create child spans + emit domain log events (R2, thumbnail, mutations)
- `internal/storage/` (R2) — R2 operations emit structured log events via the logger passed through context or `Telemetry`
- `cmd/server/main.go` — constructs `Telemetry` after providers are initialised and wires it into all layers
- No API surface changes; no new environment variables; no breaking changes to existing interfaces
