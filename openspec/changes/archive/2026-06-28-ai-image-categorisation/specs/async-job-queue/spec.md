## ADDED Requirements

### Requirement: CategorisationJob

The system SHALL define `CategoriseImageArgs` in `internal/usecase/job_args.go` and `CategorisationWorker` in `internal/worker/categorise.go`.

```go
type CategoriseImageArgs struct {
    ImageID uuid.UUID `json:"image_id"`
    UserID  string    `json:"user_id"`
}

func (CategoriseImageArgs) Kind() string     { return "categorise_image" }
func (CategoriseImageArgs) MaxAttempts() int { return 3 }
```

`CategorisationWorker` SHALL be configured with `river.WorkerDefaults[CategoriseImageArgs]` and SHALL override `NextRetry` to return a fixed 30-second delay from the error time. On each attempt the worker SHALL call `categorisationUsecase.CategoriseImage(ctx, args.UserID, args.ImageID)`.

`CategorisationWorker` SHALL be registered with River in `main.go` and `CategorisationUsecase` SHALL be wired as its dependency (not as dead code).

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
