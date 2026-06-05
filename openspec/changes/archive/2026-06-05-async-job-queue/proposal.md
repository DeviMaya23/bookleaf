## Why

Background work in the upload flow is either a fire-and-forget goroutine with no retry (thumbnail upload) or blocking the HTTP response unnecessarily (vision labelling). The ticker-based workers in `main.go` have no persistence, no retry, and their interface declarations don't belong there.

## What Changes

- **NEW**: `internal/worker/` package housing typed River job workers
- **NEW**: River client initialised in `main.go`; River DB migration added
- **MODIFIED**: `CompleteUpload` enqueues a `ThumbnailUploadJob` and a `VisionJob` instead of launching a goroutine and calling `runVisionFlow` synchronously
- **MODIFIED**: `CleanupStaleUploads` and `PurgeExpiredTrash` become River periodic jobs; ticker goroutines removed from `main.go`
- **REMOVED**: `compositeImageWorker`, `imageWorkerUsecase` interface, and `startWorkers` from `main.go`
- **BREAKING**: `CompleteUpload` response no longer returns `suggested_folder_name` or `warning`; `CompleteUploadResult` is simplified to just `ImageID`

## Capabilities

### New Capabilities

- `async-job-queue`: River-based job queue backed by Postgres. Covers River client setup, worker registration, job argument types, and the `internal/worker/` package structure.

### Modified Capabilities

- `image-thumbnail`: Async thumbnail storage requirement changes from a fire-and-forget goroutine to a River job with 5 attempts and immediate first execution.
- `vision-api-labelling`: Vision flow requirement changes from synchronous execution inside `CompleteUpload` to a River job with 3 attempts and ~10s retry delay. `CompleteUpload` no longer returns `suggested_folder_name` or `warning`.
- `stale-upload-cleanup`: Background cleanup requirement changes from a ticker goroutine in `main.go` to a River periodic job running every 10 minutes.
- `trash-purge`: Background purge requirement changes from a ticker goroutine in `main.go` to a River periodic job running every 24 hours.

## Impact

- **Backend**: `internal/usecase/image_upload_usecase.go`, `cmd/server/main.go`, new `internal/worker/` package, new DB migration
- **Dependencies**: `riverqueue/river`, `riverqueue/river/riverdriver/riverpgxv5`, `jackc/pgx/v5`
- **Database**: River schema migration required before deployment
- **API**: `POST /images/:id/complete` response shape changes (breaking for FE, acceptable — AI labelling not in production)
