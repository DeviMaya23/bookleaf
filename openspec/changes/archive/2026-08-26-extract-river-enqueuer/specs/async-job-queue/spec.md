## MODIFIED Requirements

### Requirement: Worker Package Structure

The system SHALL house all River workers in `internal/worker/`, organized into domain-grouped files:

- `image.go` — `VisionWorker`, `CategorisationWorker`, `BackfillPhashWorker`, and their Args types
- `account_deletion.go` — `AccountWipeWorker`, `BookletUserDeletionWorker`, `AccountWipeReconcileWorker`, `PurgedAccountSweepWorker`, and their Args types
- `cleanup.go` — `CleanupStaleUploadsWorker`, `TrashPurgeWorker`, `R2DeleteWorker`, and their Args types

Each worker SHALL implement `river.Worker[T]` for an Args type defined in the same file. Workers SHALL depend on narrow usecase interfaces (not repositories), following the same handler → usecase dependency convention. `Kind()` and `MaxAttempts()` SHALL remain on Args types — they are River infrastructure concerns and belong in the worker package.

#### Scenario: Worker package compiles against usecase interfaces

- **WHEN** the Go package is compiled
- **THEN** all workers in `internal/worker/` satisfy `river.Worker[T]` for their respective args type without compilation errors

#### Scenario: Args types are defined in the worker package

- **WHEN** the Go package is compiled
- **THEN** no job Args types exist in the `usecase` package

---

### Requirement: CategorisationJob

The system SHALL define `CategoriseImageArgs` in `internal/worker/image.go` and `CategorisationWorker` in the same file.

```go
type CategoriseImageArgs struct {
    ImageID uuid.UUID `json:"image_id"`
    UserID  string    `json:"user_id"`
}

func (CategoriseImageArgs) Kind() string     { return "categorise_image" }
func (CategoriseImageArgs) MaxAttempts() int { return 3 }
```

`CategorisationWorker` SHALL be configured with `river.WorkerDefaults[CategoriseImageArgs]` and SHALL override `NextRetry` to return a fixed 30-second delay from the error time. On each attempt the worker SHALL call `categorisationUsecase.CategoriseImage(ctx, args.UserID, args.ImageID)`.

`CategorisationWorker` SHALL be registered with River in `main.go` and `CategorisationUsecase` SHALL be wired as its dependency.

#### Scenario: Job is registered and accepted by River

- **WHEN** the server starts
- **THEN** River accepts insertions of `CategoriseImageArgs` jobs without error

#### Scenario: Retry uses fixed 30s delay

- **WHEN** `CategoriseImage` returns a non-nil error
- **THEN** the next attempt is scheduled approximately 30 seconds later
- **AND** the job is discarded after 3 total failed attempts

#### Scenario: Worker calls CategoriseImage with correct args

- **WHEN** a `CategoriseImageArgs` job is processed
- **THEN** `categorisationUsecase.CategoriseImage` is called with the job's `UserID` and `ImageID`

## ADDED Requirements

### Requirement: Per-Usecase Job Enqueuer Interfaces

Each usecase that enqueues jobs SHALL declare its own narrow enqueuer interface, following the same pattern as repository interfaces. No usecase SHALL reference `river`, `JobArgs`, or any job queue type directly.

The interfaces SHALL be:

```go
// account_usecase.go
type accountJobEnqueuer interface {
    EnqueueAccountWipe(ctx context.Context, userID string) error
    EnqueueAccountWipeUnique(ctx context.Context, userID string) error
    EnqueueBookletUserDeletion(ctx context.Context, userID string) error
    EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}

// image_upload_usecase.go
type imageUploadJobEnqueuer interface {
    EnqueueVision(ctx context.Context, imageID uuid.UUID, userID string) error
    EnqueueCategoriseImage(ctx context.Context, imageID uuid.UUID, userID string) error
}

// trash_usecase.go
type trashJobEnqueuer interface {
    EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}
```

#### Scenario: Usecase package has no River imports

- **WHEN** the `usecase` package is compiled
- **THEN** it contains no imports of `github.com/riverqueue/river` or any River sub-package

#### Scenario: Enqueuer interface is satisfied by the River adapter

- **WHEN** the server starts
- **THEN** the concrete `riverEnqueuer` in `cmd/server/enqueuer.go` satisfies all three per-usecase enqueuer interfaces

---

### Requirement: River Adapter File

The system SHALL define the concrete River job enqueuer in `cmd/server/enqueuer.go` as a single struct in `package main`. The struct SHALL implement all per-usecase enqueuer interfaces. Each method SHALL construct the appropriate Args type from the `worker` package and call `river.Client.Insert` with the correct `InsertOpts`.

`EnqueueAccountWipeUnique` SHALL use `river.UniqueOpts` scoped to `ByArgs: true` across all non-terminal job states (available, scheduled, pending, running, retryable, discarded), matching the previous `InsertUnique` behavior.

The deferred init pattern — setting `enqueuer.client` after `river.NewClient` to break the init cycle — SHALL be preserved.

#### Scenario: Adapter file is separate from main.go

- **WHEN** the `cmd/server/` directory is listed
- **THEN** `enqueuer.go` exists as a separate file from `main.go`
- **AND** `main.go` contains no `riverEnqueuer` type definition or method implementations

#### Scenario: Unique enqueue deduplicates pending wipe jobs

- **WHEN** `EnqueueAccountWipeUnique` is called for a `userID` that already has a pending `AccountWipeArgs` job
- **THEN** no duplicate job is inserted
- **AND** the call returns nil

## REMOVED Requirements

### Requirement: Generic JobEnqueuer Interface

**Reason**: Replaced by per-usecase typed enqueuer interfaces. The generic interface leaked River's `JobArgs` shape (`Kind()`, `MaxAttempts()`) into the usecase layer and required a type assertion in the adapter to cast `usecase.JobArgs` to `river.JobArgs`.

**Migration**: Each usecase that previously accepted `JobEnqueuer` now accepts its own narrow interface. The concrete River adapter in `cmd/server/enqueuer.go` implements all of them.
