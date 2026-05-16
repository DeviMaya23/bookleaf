## ADDED Requirements

### Requirement: CompleteUpload marks image as uploaded

The system SHALL set `is_uploaded = true` on the Image record when `CompleteUpload` is called successfully.

#### Scenario: Successful CompleteUpload updates is_uploaded

- **WHEN** `CompleteUpload` is called with a valid image ID and user ID
- **THEN** the image record's `is_uploaded` field is set to `true` in the database

#### Scenario: CompleteUpload on non-existent image does not update is_uploaded

- **WHEN** `CompleteUpload` is called with an image ID that does not exist for the given user
- **THEN** the operation returns an error and no `is_uploaded` update is performed

### Requirement: CleanupStaleUploads removes abandoned upload records

The system SHALL provide a `CleanupStaleUploads(ctx context.Context, threshold time.Duration)` method on `imageUsecase` that identifies and removes Image records where the upload was never completed.

A record is considered stale when:
- `is_uploaded = false`
- `created_at < now() - threshold`

For each stale record, the method SHALL:
1. Attempt to delete the R2 object at `r2_path` (best-effort; log a warning on failure but continue)
2. Soft-delete the Image record

#### Scenario: Stale records are soft-deleted and R2 objects are removed

- **WHEN** `CleanupStaleUploads` runs and finds Image records with `is_uploaded = false` older than the threshold
- **THEN** the R2 object at each record's `r2_path` is deleted
- **AND** each Image record is soft-deleted (GORM `DeletedAt` is set)

#### Scenario: No stale records results in no-op

- **WHEN** `CleanupStaleUploads` runs and finds no Image records matching the stale criteria
- **THEN** no records are modified and no R2 deletes are attempted

### Requirement: Background goroutine runs cleanup on a ticker

The system SHALL start a background goroutine in `cmd/server/main.go` after dependency wiring that calls `imageUsecase.CleanupStaleUploads` on a 10-minute interval with a 30-minute stale threshold.

#### Scenario: Goroutine starts with the server

- **WHEN** the server starts
- **THEN** a goroutine is running that invokes `CleanupStaleUploads` every 10 minutes

#### Scenario: Cleanup goroutine does not block server startup

- **WHEN** the server starts
- **THEN** the goroutine is launched asynchronously and `e.Start()` is called immediately after
