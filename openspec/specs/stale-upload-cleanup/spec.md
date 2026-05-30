## ADDED Requirements

### Requirement: CompleteUpload commits pending upload to images table

The system SHALL commit a pending upload to the `images` table when `CompleteUpload` is called successfully.

`CompleteUpload` SHALL:
1. Fetch the `PendingUpload` record via `pendingUploadRepo.GetByID(ctx, id, userID)` — return error if not found
2. Generate the thumbnail and extract image dimensions and file size (`prepareThumbnail`)
3. Execute a single DB transaction:
   a. `imageRepo.Create` with all fields from the `PendingUpload` plus server-derived metadata (width, height, file_size, thumbnail_path)
   b. If `pendingUpload.FolderID` is non-nil: call `imageRepo.SetImageFolder(ctx, image.ID, pendingUpload.FolderID)`
   c. `pendingUploadRepo.Delete(ctx, id)` — hard-delete the pending row
4. Return the committed image result

If any step fails, the transaction is rolled back and the `pending_uploads` row survives; the stale cleaner will remove it after the threshold.

#### Scenario: Successful CompleteUpload commits image and removes pending row

- **WHEN** `CompleteUpload` is called with a valid pending upload ID and matching user ID
- **THEN** a row is inserted into `images` with all metadata from `pending_uploads` plus thumbnail, width, height, and file_size
- **AND** the corresponding `pending_uploads` row is deleted
- **AND** if `pending_uploads.folder_id` was non-nil, a row exists in `image_folders` for the new image

#### Scenario: CompleteUpload on non-existent pending upload returns error

- **WHEN** `CompleteUpload` is called with an ID that does not exist in `pending_uploads` for the given user
- **THEN** the operation returns an error
- **AND** no row is inserted into `images`

#### Scenario: CompleteUpload transaction is atomic

- **WHEN** `CompleteUpload` fails during the transaction (e.g. DB error on INSERT)
- **THEN** no row is inserted into `images`
- **AND** the `pending_uploads` row remains and will be cleaned up by the stale cleaner

---

### Requirement: CleanupStaleUploads removes abandoned pending upload records

The system SHALL provide a `CleanupStaleUploads(ctx context.Context, threshold time.Duration)` method on `imageUsecase` that identifies and removes `pending_uploads` records where the upload was never completed.

A record is considered stale when:
- `created_at < now() - threshold`

For each stale record, the method SHALL:
1. Attempt to delete the R2 object at `r2_path` (best-effort; log a warning on failure but continue)
2. Call `pendingUploadRepo.Delete(ctx, id)` to hard-delete the `pending_uploads` row

#### Scenario: Stale pending uploads are deleted and R2 objects are removed

- **WHEN** `CleanupStaleUploads` runs and finds `pending_uploads` rows older than the threshold
- **THEN** the R2 object at each record's `r2_path` is deleted
- **AND** each `pending_uploads` row is hard-deleted

#### Scenario: No stale records results in no-op

- **WHEN** `CleanupStaleUploads` runs and finds no `pending_uploads` rows matching the stale criteria
- **THEN** no records are modified and no R2 deletes are attempted

---

### Requirement: Background goroutine runs cleanup on a ticker

The system SHALL start a background goroutine in `cmd/server/main.go` after dependency wiring that calls `imageUsecase.CleanupStaleUploads` on a 10-minute interval with a 30-minute stale threshold.

#### Scenario: Goroutine starts with the server

- **WHEN** the server starts
- **THEN** a goroutine is running that invokes `CleanupStaleUploads` every 10 minutes

#### Scenario: Cleanup goroutine does not block server startup

- **WHEN** the server starts
- **THEN** the goroutine is launched asynchronously and `e.Start()` is called immediately after
