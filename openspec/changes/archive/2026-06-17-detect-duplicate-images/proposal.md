## Why

Users frequently upload the same image multiple times — re-saved JPEGs, resized copies, or re-downloads from the same source — polluting their moodboard with near-identical content. Detecting duplicates at upload time lets us surface this before the image lands in the gallery, keeping collections clean without requiring users to manually scan for duplicates.

## What Changes

- The `Image` domain struct gains a `PHash` field (`bit(64)`) to store a perceptual hash computed client-side and sent at upload completion.
- A new DB migration adds a `phash bit(64)` nullable column to the `images` table. Existing images will have `phash = NULL` until the backfill completes; duplicate detection is skipped for un-hashed images.
- The periodic worker gains a backfill sweep: each tick, it fetches a batch of images where `phash IS NULL`, downloads their thumbnails from R2, computes pHash server-side via `goimagehash`, and stores the result. This continues until all existing images are hashed, and self-heals any future gaps.
- `CompleteUpload` usecase and endpoint accept a `phash` value and, after persisting the new image, query the user's existing images for Hamming-distance matches below a threshold. Matching images are returned in the `POST /images/:id/complete` response.
- The FE computes pHash on the thumbnail (already generated as part of the upload flow) before calling `completeUpload`, and passes it in the request body.
- On a successful upload, if the `completeUpload` response includes duplicate matches, a warning toast is shown naming the first matching image.
- In the batch upload modal, files with duplicate matches display an inline duplicate warning alongside their success indicator.

## Capabilities

### New Capabilities

- `image-duplicate-detection`: Backend duplicate detection — `FindDuplicates` repository method, pHash field persistence, usecase query logic, API response shape for returning matches, and periodic worker backfill sweep for existing images.

### Modified Capabilities

- `image-domain`: `Image` struct gains a `PHash *string` field (stored as `bit(64)` in Postgres); new migration adds the column.
- `fe-image-upload-flow`: Upload sequence extended — FE computes pHash on thumbnail and includes it in the `POST /images/:id/complete` request body; response now includes a `duplicates` array; a warning toast is shown when duplicates are detected.
- `fe-batch-upload`: Per-file success handling updated — when `completeUpload` returns duplicate matches for a file, its row in the batch modal shows an inline duplicate warning in addition to the success indicator.

## Impact

**Backend**
- `backend/internal/domain/image.go` — new `PHash` field
- `backend/internal/migrations/` — new migration adding `phash bit(64)` to `images`
- `backend/internal/usecase/image_upload.go` — `CompleteUpload` accepts phash, calls `FindDuplicates`, extends `CompleteUploadResult`
- `backend/internal/usecase/image_upload_usecase_test.go` — new test scenarios
- `backend/internal/handler/image_upload.go` — `completeUploadRequest` and `completeUploadResponse` updated
- `backend/internal/repository/image_repository.go` — new `FindDuplicates` method
- `backend/internal/worker/periodic.go` — new backfill sweep for images with `phash IS NULL`
- New Go dependency: `github.com/corona10/goimagehash` (pHash computation for the periodic backfill)

**Frontend**
- `frontend/src/lib/images.ts` — `completeUpload` sends `phash`, `CompleteUploadResult` gains `duplicates`
- `frontend/src/lib/upload.ts` — `uploadImageFile` computes pHash on the thumbnail blob and passes it through
- `frontend/src/features/upload/components/BatchUploadModal.tsx` — `BatchFile` type and `StatusCell` updated for duplicate state
- New FE dependency: pHash computation library (e.g. `@bouzidanas/phash-js` or lightweight DCT implementation)
