## Context

The observability stack is fully wired at the infrastructure level: `NewLogger`, `NewTracerProvider`, `NewMeterProvider`, and the Echo trace/metrics middleware are all in place. However, application layers — handlers, usecases, and the R2 storage client — are uninstrumented. Structured domain events are absent, and distributed traces terminate at the HTTP boundary span with no child spans showing where time is actually spent.

Current state of each layer:
- `internal/handler/` — no spans, no structured logging beyond implicit error returns
- `internal/usecase/` — uses `log.Printf` (stdlib) in a few places; no spans
- `internal/storage/r2.go` — no logging or tracing

## Goals / Non-Goals

**Goals:**
- Introduce a `Telemetry` struct that bundles `Logger`, `Tracer`, and `Meter` for clean injection into application layers
- Add a request lifecycle logging middleware that captures `user_id`, endpoint, method, status code, and duration after Kinde auth resolves the user
- Instrument all handlers, usecases, and R2 storage with child spans (`package.Function` naming) and structured log events per the logging conventions below
- Standardise log field names across all layers

**Non-Goals:**
- Instrumentation of the SQL repository layer (spans and logs for DB queries are out of scope for this change)
- Adding new metrics instruments beyond what the existing `MetricsMiddleware` already records
- Changing the Kinde auth middleware itself (auth rejection logs are added via a thin wrapper, not by modifying the middleware)

## Decisions

### D1 — Telemetry struct for dependency injection

The `Telemetry` struct (`internal/observability/telemetry.go`) bundles `*zap.Logger`, `trace.Tracer`, and `metric.Meter`:

```go
type Telemetry struct {
    Logger *zap.Logger
    Tracer trace.Tracer
    Meter  metric.Meter
}

func NewTelemetry(logger *zap.Logger, tracer trace.Tracer, meter metric.Meter) *Telemetry
```

**Why not pass each primitive individually?** Three separate constructor parameters per layer creates churn whenever a new observability primitive is added. A struct keeps signatures stable.

**Why not use context to thread the logger/tracer?** Context-carried dependencies are invisible at the type level and require type assertions. Explicit struct fields are testable and self-documenting.

**Why not use global OTel accessors (`otel.GetTracerProvider()`)?** Global state is harder to test and couples components to init order. The struct makes injection explicit and replaceable in tests with no-op implementations.

All application layers — `ImageHandler`, `FolderHandler`, `MeHandler`, `ImageUsecase`, `FolderUsecase`, `UserUsecase`, and `r2Storage` — receive `*Telemetry` as a struct field set at construction time. The `HealthHandler` is excluded; it already logs via the existing logger injected in main and is a health probe, not a domain operation.

### D2 — Request lifecycle logging via a dedicated Echo middleware

A new `LoggingMiddleware(tel *observability.Telemetry) echo.MiddlewareFunc` logs one INFO line per request:

Fields: `user_id` (from Kinde context), `method`, `endpoint` (matched route pattern), `status_code`, `duration_ms`, `trace_id` (via `LoggerFromContext`).

**Why middleware, not per-handler logging?** Every handler would need the same boilerplate. Middleware guarantees coverage and a consistent field set without duplicating code across 10+ routes.

**Why after Kinde auth?** The user ID is only available once the JWT is validated. The logging middleware is registered on the protected route group (after auth), not on the global Echo instance. Unauthenticated rejections (401s) are handled separately by the auth middleware.

### D3 — Auth rejection logging in the Kinde middleware

The existing Kinde auth middleware is extended to emit a single WARN log when it rejects a token:

Fields: `event: "auth.token_rejected"`, `reason` (e.g., `"invalid_token"`, `"missing_header"`), `trace_id`.

**Why WARN not ERROR?** A rejected token is an expected adversarial or misconfiguration event, not an internal server error.

Cross-user resource access cannot be reliably detected at the handler layer because the repo layer returns `gorm.ErrRecordNotFound` for both genuine not-found and ownership mismatches. Logging for this case is deferred to a future spec that introduces a distinct error type.

### D4 — Per-layer child span convention

At the start of any handler, usecase, or storage method that does meaningful work:

```go
ctx, span := tel.Tracer.Start(ctx, "package.Function")
defer span.End()
```

Errors are recorded before returning:

```go
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())
return err
```

**Why `package.Function` naming?** Span names appear in Jaeger's trace list. `handler.InitiateImageUpload` is immediately locatable in the source; a generic `"upload"` is not.

**Why always defer `span.End()`?** Early returns and panics must not leak spans. Deferring guarantees the span is always closed.

### D5 — Standardised log field names

All layers use the same field keys to enable consistent filtering in log aggregators:

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | string | Kinde user ID |
| `image_id` | string | Image UUID |
| `folder_id` | string | Folder UUID |
| `file_size` | int64 | Bytes |
| `mime_type` | string | MIME type string |
| `r2_key` | string | Full R2 object key |
| `duration_ms` | float64 | Operation wall time in milliseconds |
| `method` | string | HTTP method |
| `endpoint` | string | Matched Echo route pattern |
| `status_code` | int | HTTP response status |
| `event` | string | Dot-namespaced event name (e.g. `"r2.upload.started"`) |
| `image_count` | int | Count of images (folder deletion) |
| `error` | string | Error message (on failure events only) |

No image binary data, no presigned URLs, no raw JWT claims are ever logged.

### D6 — Domain log events by layer

**Handler layer** (via `LoggingMiddleware` + auth middleware):
- `auth.token_rejected` (WARN) — invalid/missing token

**Usecase layer:**
- `r2.upload.started` (INFO) — fields: `image_id`, `user_id`, `mime_type`, `file_size`, `r2_key`
- `r2.upload.completed` (INFO) — fields: `image_id`, `user_id`, `duration_ms`
- `thumbnail.job.started` (INFO) — fields: `image_id`, `user_id`
- `thumbnail.job.completed` (INFO) — fields: `image_id`, `user_id`, `duration_ms`
- `thumbnail.job.failed` (ERROR) — fields: `image_id`, `user_id`, `error`
- `image.mutated` (INFO) — fields: `image_id`, `user_id`, `operation` (e.g. `"moved_to_folder"`, `"trashed"`, `"edited"`)
- `folder.mutated` (INFO) — fields: `folder_id`, `user_id`, `operation`, `image_count` (deletion only)

**Storage layer (R2):**
- `r2.presigned_put.success` (INFO) — fields: `image_id`
- `r2.presigned_put.failed` (ERROR) — fields: `image_id`, `error`
- `r2.presigned_get.success` (INFO) — fields: `image_id`
- `r2.presigned_get.failed` (ERROR) — fields: `image_id`, `error`

## Risks / Trade-offs

- **Constructor churn** → Every handler and usecase constructor gains a `*Telemetry` param and `main.go` must wire it. This is mechanical but touches many files simultaneously. Mitigated by the fact that `Telemetry` is a single addition rather than three separate params; future primitives don't require further signature changes.
- **Log volume in production** → INFO logs on every R2 presigned URL generation and every mutation may be high-volume on active accounts. Mitigated by using structured Zap fields (no string interpolation overhead) and the fact that production uses JSON format which can be sampled at the log aggregator layer.
- **No-op Telemetry in tests** → Unit tests pass `observability.NewTelemetry(nil, nil, nil)`; nil fields are substituted with noop implementations inside `NewTelemetry`, so no guard code is needed in constructors or tests.

## Migration Plan

1. Add `internal/observability/telemetry.go` with `Telemetry` struct and `NewTelemetry` (nil-safe)
2. Update `r2Storage` constructor to accept `*Telemetry`; add log events to `GeneratePresignedPutURL`, `GeneratePresignedGetURL`, `PutObject`
3. Update usecase constructors to accept `*Telemetry`; add child spans and domain log events
4. Update handler constructors to accept `*Telemetry`; add child spans and auth event logs
5. Add `LoggingMiddleware` to `internal/observability/echo_middleware.go`; register on the protected route group in `main.go`
6. Extend Kinde middleware to log token rejection events
7. Wire `NewTelemetry(logger, tracer, meter)` in `cmd/server/main.go` after providers are initialised; pass to all constructors
8. No rollback required — adding logging/tracing is purely additive; removing it is a revert
