## MODIFIED Requirements

### Requirement: PurgeExpiredTrash permanently removes old trashed images

The system SHALL provide a `PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error` method on `TrashUsecase` (not `ImageUsecase`) that permanently removes images that have been soft-deleted for longer than the given threshold.

For each expired record, the method SHALL:
1. Attempt to delete the R2 object at `r2_path` (best-effort; log warn on failure, continue)
2. Attempt to delete the R2 object at `thumbnail_path` if it is not nil (best-effort; log warn on failure, continue)
3. Hard-delete the DB record (permanent removal, not soft delete)

#### Scenario: Expired trashed images are purged

- **WHEN** `PurgeExpiredTrash` runs and finds images with `deleted_at` older than the threshold
- **THEN** the R2 object at `r2_path` is deleted for each image
- **AND** the R2 object at `thumbnail_path` is deleted for each image that has a thumbnail
- **AND** each image record is permanently removed from the database

#### Scenario: No expired records results in no-op

- **WHEN** `PurgeExpiredTrash` runs and no images have been trashed longer than the threshold
- **THEN** no records are modified and no R2 deletes are attempted

#### Scenario: R2 delete failure does not block hard delete

- **WHEN** `PurgeExpiredTrash` runs and the R2 delete for a record fails
- **THEN** the error is logged at warn level
- **AND** the DB record is still hard-deleted

---

### Requirement: Background periodic job runs purge every 24 hours

The system SHALL run `PurgeExpiredTrash` as a River periodic job firing every 24 hours with a 30-day retention threshold. `TrashPurgeWorker` in `internal/worker/periodic.go` SHALL depend on `trashUsecase` (the new `TrashUsecase` implementation), not `imageUsecase`.

#### Scenario: Purge fires on schedule via River using trashUsecase

- **WHEN** the server is running
- **THEN** `trashUsecase.PurgeExpiredTrash(ctx, 30*24*time.Hour)` is invoked approximately every 24 hours by River's periodic job scheduler

#### Scenario: main.go registers TrashPurgeWorker with trashUsecase

- **WHEN** the Go package is compiled
- **THEN** `river.AddWorker` for `TrashPurgeWorker` receives the `trashUsecase` instance, not `imageUsecase`
