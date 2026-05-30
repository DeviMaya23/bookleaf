## Why

The `images` table currently uses an `is_uploaded` boolean flag to distinguish committed images from in-flight upload records. This creates an invisible invariant — every user-facing query must filter `is_uploaded = true` or it silently leaks uncommitted records — and that invariant is not uniformly enforced today (`GetByID`, `CountByFolderID`, and others omit it). A pending upload is not an image; the boolean flag is a category error in the domain model that produces correctness gaps and will worsen as the codebase grows.

## What Changes

- Introduce a `pending_uploads` table that holds upload-in-progress records until `CompleteUpload` commits them atomically into `images`
- `InitiateUpload` writes to `pending_uploads` instead of `images`; the presigned URL is returned unchanged
- `CompleteUpload` fetches from `pending_uploads`, generates the thumbnail, then executes a transaction: INSERT into `images`, call `SetImageFolder` if a folder was specified, DELETE from `pending_uploads`
- The `is_uploaded` column is dropped from `images`; the `images` table becomes a clean set of committed records requiring no filter
- The stale upload cleaner targets `pending_uploads` instead of `images WHERE is_uploaded = false`
- **BREAKING** (internal only): `ImageRepository.ListStaleUploads` is removed; a new `PendingUploadRepository` provides equivalent stale-listing

## Capabilities

### New Capabilities

- `pending-uploads`: `PendingUpload` domain struct, `pending_uploads` migration, `PendingUploadRepository` interface and SQL implementation

### Modified Capabilities

- `image-domain`: `is_uploaded` field removed from `Image` struct; corresponding DB column dropped via migration
- `image-endpoints`: `InitiateUpload` writes to `pending_uploads`; `CompleteUpload` reads from `pending_uploads` and commits to `images` in a transaction; `List` drops the `is_uploaded = true` filter; `ImageRepository` interface loses `ListStaleUploads`
- `stale-upload-cleanup`: stale detection and deletion now operates on `pending_uploads` rows via `PendingUploadRepository`; R2 cleanup behaviour unchanged

## Impact

- `backend/internal/domain/image.go` — remove `IsUploaded` field
- `backend/internal/domain/` — new `pending_upload.go`
- `backend/internal/usecase/image_repository.go` — remove `ListStaleUploads`
- `backend/internal/usecase/` — new `pending_upload_repository.go` interface
- `backend/internal/repository/image_repository.go` — remove `is_uploaded` filter from `List`, remove `ListStaleUploads`
- `backend/internal/repository/` — new `pending_upload_repository.go` SQL implementation
- `backend/internal/usecase/image_usecase.go` — `InitiateUpload`, `CompleteUpload`, `CleanupStaleUploads` rewritten against new repos
- `backend/migration/` — new up/down migration pair
- Unit tests (usecase + handler) and integration tests updated accordingly
- No HTTP API surface changes; handler layer untouched
