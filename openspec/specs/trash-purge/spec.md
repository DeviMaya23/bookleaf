## ADDED Requirements

### Requirement: PurgeExpiredTrash permanently removes old trashed images

The system SHALL provide a `PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error` method on `ImageUsecase` that permanently removes images that have been soft-deleted for longer than the given threshold.

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

### Requirement: ListExpiredTrash queries images past the retention window

The system SHALL provide a `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)` method on `ImageRepository` that returns all images where `deleted_at IS NOT NULL AND deleted_at < olderThan`.

#### Scenario: Returns images past retention window

- **WHEN** `ListExpiredTrash` is called with a cutoff time
- **THEN** it returns all image records with `deleted_at` set and older than the cutoff
- **AND** records with `deleted_at` newer than the cutoff are not returned

#### Scenario: Returns empty slice when no records match

- **WHEN** `ListExpiredTrash` is called and no images are past the cutoff
- **THEN** it returns an empty slice and no error

### Requirement: HardDelete permanently removes an image record

The system SHALL provide a `HardDelete(ctx context.Context, id uuid.UUID, userID string) error` method on `ImageRepository` that permanently removes the image record using an unscoped delete, scoped to the given `userID`.

#### Scenario: Hard delete removes the record permanently

- **WHEN** `HardDelete` is called with a valid image ID and user ID
- **THEN** the image record is removed from the database and cannot be restored

#### Scenario: Hard delete on non-existent record returns error

- **WHEN** `HardDelete` is called with an image ID that does not match the given user ID
- **THEN** the operation returns an error

### Requirement: Background goroutine runs purge on a daily ticker

The system SHALL run `PurgeExpiredTrash` automatically every 24 hours from a goroutine started at server startup, using a 30-day retention threshold.

#### Scenario: Goroutine starts with the server

- **WHEN** the server starts
- **THEN** a goroutine is running that invokes `PurgeExpiredTrash` every 24 hours with a 30-day threshold

#### Scenario: Purge goroutine does not block server startup

- **WHEN** the server starts
- **THEN** the goroutine is launched asynchronously and `e.Start()` is called immediately after
