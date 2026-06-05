## Purpose

The system uses River (a Postgres-backed job queue) to execute background work asynchronously — including thumbnail generation, vision labelling, and periodic maintenance jobs. This spec covers the River infrastructure setup and the worker package structure.

---

## Requirements

### Requirement: River Job Queue Infrastructure

The system SHALL initialize a River client backed by the existing Postgres database. River SHALL use a dedicated `*pgxpool.Pool` opened from the same `config.DB.URL` as GORM — separate from GORM's own pool. The pool SHALL be sized at a maximum of 3 connections.

The River DB schema SHALL be installed as migration `000012` using River's published SQL. The migration SHALL be applied via the existing `golang-migrate` flow before the binary starts.

The River client SHALL be stored on the `server` struct and stopped in `server.shutdown()` before the process exits.

#### Scenario: River client starts with the server

- **WHEN** the server starts
- **THEN** the River client is running and accepting job insertions

#### Scenario: River client stops gracefully on shutdown

- **WHEN** the server receives a shutdown signal
- **THEN** `riverClient.Stop(ctx)` is called within the shutdown context deadline

---

### Requirement: Worker Package Structure

The system SHALL house all River workers in `internal/worker/`. Each worker SHALL implement `river.Worker[T]` for a specific args type. Workers SHALL depend on narrow usecase interfaces (not repositories), following the same handler → usecase dependency convention.

#### Scenario: Worker package compiles against usecase interfaces

- **WHEN** the Go package is compiled
- **THEN** all workers in `internal/worker/` satisfy `river.Worker[T]` for their respective args type without compilation errors

---

### Requirement: ThumbnailUploadJob

The system SHALL define `ThumbnailUploadArgs` and `ThumbnailUploadWorker` in `internal/worker/`.

```go
type ThumbnailUploadArgs struct {
    ImageID      uuid.UUID `json:"image_id"`
    UserID       string    `json:"user_id"`
    R2Path       string    `json:"r2_path"`
    ThumbnailKey string    `json:"thumbnail_key"`
}

func (ThumbnailUploadArgs) Kind() string { return "thumbnail_upload" }
```

The worker SHALL be configured with `river.WorkerDefaults[ThumbnailUploadArgs]` overriding `MaxAttempts` to 5. On each attempt the worker SHALL call `uploadUsecase.ProcessThumbnailUpload(ctx, args.ImageID, args.R2Path, args.ThumbnailKey)`. River's default exponential backoff applies between retries.

`CompleteUpload` SHALL insert a `ThumbnailUploadArgs` job (with default `ScheduledAt: now`) immediately after the image DB write succeeds.

#### Scenario: Job fires immediately after CompleteUpload

- **WHEN** `CompleteUpload` commits the image to the DB
- **THEN** a `ThumbnailUploadArgs` job with `Kind() == "thumbnail_upload"` is inserted in the same request
- **AND** the job is picked up within the next River poll interval

#### Scenario: Job retries on failure up to 5 attempts

- **WHEN** `ProcessThumbnailUpload` returns a non-nil error
- **THEN** River schedules a retry with exponential backoff
- **AND** the job is discarded after 5 total failed attempts

---

### Requirement: VisionJob

The system SHALL define `VisionArgs` and `VisionWorker` in `internal/worker/`.

```go
type VisionArgs struct {
    ImageID uuid.UUID `json:"image_id"`
    UserID  string    `json:"user_id"`
    R2Path  string    `json:"r2_path"`
}

func (VisionArgs) Kind() string { return "vision_labelling" }
```

The worker SHALL be configured with `river.WorkerDefaults[VisionArgs]` overriding `MaxAttempts` to 3. The worker SHALL implement `NextAttemptScheduledAt` to return a fixed 10-second delay from the error time, regardless of attempt number. On each attempt the worker SHALL call `uploadUsecase.ProcessVisionLabelling(ctx, args.ImageID, args.UserID)`.

`CompleteUpload` SHALL insert a `VisionArgs` job (with default `ScheduledAt: now`) immediately after the image DB write succeeds, alongside the thumbnail job.

#### Scenario: Job fires immediately after CompleteUpload

- **WHEN** `CompleteUpload` commits the image to the DB
- **THEN** a `VisionArgs` job with `Kind() == "vision_labelling"` is inserted in the same request

#### Scenario: Retry uses fixed 10s delay

- **WHEN** `ProcessVisionLabelling` returns a non-nil error
- **THEN** the next attempt is scheduled approximately 10 seconds later
- **AND** the job is discarded after 3 total failed attempts

---

### Requirement: Periodic Jobs Replace Ticker Workers

The system SHALL register two River periodic jobs, replacing the ticker goroutines previously in `main.go`:

- Every 10 minutes: calls `uploadUsecase.CleanupStaleUploads(ctx, 30*time.Minute)`
- Every 24 hours: calls `imageUsecase.PurgeExpiredTrash(ctx, 30*24*time.Hour)`

The `compositeImageWorker` struct, `imageWorkerUsecase` interface, and `startWorkers` function SHALL be removed from `main.go`.

#### Scenario: Periodic jobs are registered at startup

- **WHEN** the server starts
- **THEN** the River client has two periodic jobs registered
- **AND** `main.go` contains no ticker goroutines for these tasks
