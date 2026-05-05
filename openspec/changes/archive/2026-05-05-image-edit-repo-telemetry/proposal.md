## Why

Two gaps remain in the image feature and the observability stack. First, there is no way to rename an image or move it to a different folder after upload — the only mutation available is soft-delete. Second, SQL queries are invisible in traces and metrics: every span tree ends at the usecase layer, with no visibility into how long individual queries take or whether they error.

## What Changes

- Add `PATCH /images/:id` endpoint accepting `title` and/or `folder_id`; either field is optional in the request body; the image binary is never modified
- Add `UpdateImage` method to `ImageUsecase`, `ImageRepository`, and `imageRepository`
- Emit `event: "image.mutated"` log with `operation: "moved_to_folder"` only when `folder_id` changes; successful title-only edits produce no log event
- Integrate `go-gorm/opentelemetry` GORM plugin to instrument all SQL queries with traces, metrics (query duration histograms, error counts), and logging automatically
- Bridge GORM's logger interface to the existing `*zap.Logger` so SQL slow queries and errors appear in the structured log output

## Capabilities

### New Capabilities

- `image-edit`: `PATCH /images/:id` endpoint, usecase method, and repository method for updating image metadata (title, folder_id)

### Modified Capabilities

- `image-endpoints`: adds `PATCH /images/:id` route requirement and repository `Update` method
- `observability-logging`: adds `image.mutated` log event with `operation: "moved_to_folder"` from `ImageUsecase.UpdateImage` when folder changes; title-only updates produce no domain event log
- `observability-tracing`: adds SQL layer instrumentation via GORM plugin; spans for all DB queries appear as children of the calling usecase or handler span
- `observability-metrics`: GORM plugin emits DB query duration histograms and error counters; these are queryable in Prometheus alongside existing HTTP metrics

## Impact

- `internal/domain/image.go` — no struct changes required
- `internal/usecase/image_usecase.go` — new `UpdateImage` method
- `internal/repository/image_repository.go` — new `Update` method
- `internal/handler/image.go` — new `UpdateImage` handler
- `cmd/server/main.go` — register GORM plugin and new route
- New dependency: `github.com/go-gorm/opentelemetry`
- `internal/repository/` — shared GORM plugin setup; zap-to-GORM logger bridge
