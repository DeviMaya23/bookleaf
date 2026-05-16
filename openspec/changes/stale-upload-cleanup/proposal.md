## Why

When `InitiateUpload` is called, an Image record is created in the database and a presigned PUT URL is returned to the frontend. If the frontend never uploads to R2 or never calls `CompleteUpload`, the DB record persists indefinitely as a ghost entry with no corresponding file — polluting the image table and misrepresenting storage state.

## What Changes

- Add `is_uploaded bool DEFAULT false` column to the `images` table via a new DB migration
- `CompleteUpload` usecase sets `is_uploaded = true` when an upload is confirmed
- New `CleanupStaleUploads` method on `imageUsecase` soft-deletes image records where `is_uploaded = false` and `created_at < now() - 30 minutes`, and deletes the corresponding R2 object if one exists
- Background goroutine in `cmd/server/main.go` runs `CleanupStaleUploads` on a 10-minute ticker

## Capabilities

### New Capabilities

- `stale-upload-cleanup`: Periodic background cleanup of image records where the presigned upload was never completed

### Modified Capabilities

- `image-domain`: `is_uploaded` field added to the Image model and its DB migration

## Impact

- **Database**: New `is_uploaded` boolean column on `images` table; new migration file required
- **API**: No endpoint changes; `CompleteUpload` behaviour is extended (not broken)
- **Storage**: Goroutine calls R2 delete — requires `storageService` to be passed into or accessible from the cleanup path
- **Startup**: `cmd/server/main.go` starts a goroutine after wiring dependencies; no new dependencies introduced
