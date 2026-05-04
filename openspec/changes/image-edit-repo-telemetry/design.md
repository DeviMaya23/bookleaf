## Context

Two independent areas of work are bundled here because both are small enough not to warrant separate change cycles. The image edit feature fills a product gap: images can be uploaded and deleted but their metadata is immutable after upload. The GORM observability fills a telemetry gap: the current span tree shows handler → usecase → storage (R2) but drops at the database boundary — SQL query duration and errors are invisible.

Current state:
- No `PATCH /images/:id` route, no `UpdateImage` method anywhere in the stack
- `go.opentelemetry.io/otel` is wired throughout but `gorm.DB` has no registered plugins
- `observability-logging` spec deferred `image.mutated / edited` and `image.mutated / moved_to_folder` because `UpdateImage` did not exist

## Goals / Non-Goals

**Goals:**
- `PATCH /images/:id` that updates `title` and/or `folder_id` in a single request
- `image.mutated / moved_to_folder` log event when `folder_id` changes
- GORM plugin that adds spans, DB metrics, and structured log output for every SQL query without per-repository changes
- Zap bridge so SQL logs (slow queries, errors) flow through the existing `*zap.Logger` at consistent field names

**Non-Goals:**
- Replacing the image binary or `r2_path`
- Moving images between users
- Per-repository manual span creation (the plugin handles this globally)
- Changing the folder domain or folder endpoints

## Decisions

### D1 — PATCH semantics: all metadata fields optional, omit = no change

The request body is `{"title": "...", "folder_id": "..."}` where both fields are optional. A missing field means "do not change this field." A null `folder_id` explicitly moves the image to the root (no folder). This matches standard PATCH convention and avoids forcing clients to re-send unchanged fields.

The usecase receives a dedicated `UpdateImageParams` struct (or pointer fields) so the handler can distinguish "not provided" from "provided as zero value." Using `*string` for title and `**uuid.UUID` (or an explicit sentinel) for folder_id is the standard Go pattern.

Folder ID is nullable in the domain model (`*uuid.UUID`). To allow clients to clear it (move to root), the request field should be modelled as a JSON value that can be `null`, `"<uuid>"`, or absent. The handler unmarshals into a struct with a custom presence flag or uses a `json.RawMessage` wrapper. The simplest approach: use `*uuid.UUID` in the request body where the field being absent in JSON means the pointer is nil (omitempty behaviour from `Bind`).

**Clarification:** a null `folder_id` in the PATCH body clears the folder (moves to root). An absent `folder_id` field leaves it unchanged. This requires the handler to distinguish nil-because-absent from nil-because-null. Use a `bool` presence flag per field, or model using `omitempty`-aware decoding.

### D2 — Conditional `image.mutated / moved_to_folder` log

The log event is emitted only when `folder_id` is present in the request AND differs from the current value. The usecase compares the new folder_id against the existing image record after fetch. Title-only changes produce no domain event log — the span and the single LoggingMiddleware request log are sufficient.

### D3 — go-gorm/opentelemetry plugin registered once at startup

`otelgorm.NewPlugin()` accepts a tracer provider, meter provider, and logger. It is registered once via `db.Use(plugin)` in `main.go` after `gorm.Open`. All repositories share the same `*gorm.DB` and automatically inherit the plugin without any per-file changes.

The plugin creates child spans named `gorm.{operation}` (e.g. `gorm.Create`, `gorm.Query`) using the global tracer. These spans are children of whatever span is in the context at the time of the GORM call, so they slot naturally into the existing `usecase.X → gorm.Y` span tree.

### D4 — Zap bridge for GORM logging

`go-gorm/opentelemetry` accepts a `logger.Interface` (GORM's own interface). The standard GORM logger writes to `io.Writer`; to integrate with zap, a thin adapter struct implements `logger.Interface` and delegates to a `*zap.Logger`:

- `Info` → `zap.Debug` (verbose GORM info is noise at INFO level)
- `Warn` → `zap.Warn`
- `Error` → `zap.Error`
- `Trace` (slow query / error) → `zap.Warn` for slow queries, `zap.Error` for errors; includes `elapsed_ms`, `rows_affected`, `sql` (the parameterised query string, not values)

The adapter is a small internal type in `internal/repository/` or `internal/observability/`. It does not expose raw SQL parameter values to avoid accidental credential leakage.

The slow query threshold (above which Trace emits a Warn) is configurable — default 200ms, matching GORM's default.

### D5 — UpdateImage repository method: selective update with GORM

GORM's `Save` re-saves all fields (problematic for `r2_path`, `thumbnail_path`, etc.). Use `db.Model(&Image{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]any{...})` with only the fields that are explicitly present in the request. `RowsAffected == 0` is treated as `gorm.ErrRecordNotFound`.

## Risks / Trade-offs

- **SQL statement in logs** — the zap bridge logs the parameterised query string for slow queries and errors. Query parameters (values) are NOT logged. However, table/column names and query structure are visible — acceptable for internal observability.
- **GORM plugin metric cardinality** — the plugin emits metrics labelled by SQL operation type and table. Cardinality is bounded and does not include raw query text.
- **Null vs absent folder_id** — the PATCH body semantics (null = clear, absent = unchanged) are easy to misuse from clients. This is documented in the API spec.

## Migration Plan

1. Add `go-gorm/opentelemetry` dependency
2. Implement zap-to-GORM logger bridge in `internal/repository/`
3. Register plugin in `main.go` after `gorm.Open`
4. Add `UpdateImage` down the stack (repo → usecase → handler)
5. Register `PATCH /images/:id` route in `main.go`
6. Add unit tests for handler and usecase; add integration test for repository

No database migrations required — no schema changes.
