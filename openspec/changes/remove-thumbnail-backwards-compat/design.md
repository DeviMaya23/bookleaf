## Context

All upload paths (web app and browser extension) now generate and upload thumbnails client-side via presigned URL before calling `POST /images/:id/complete`. The backend infrastructure that existed as a fallback for the old extension flow is now unreachable:

- `CompleteUpload` calls `HeadObject` to detect whether a thumbnail exists, then conditionally enqueues `ThumbnailUploadArgs` if it does not. With all clients uploading thumbnails, the "not found" branch is never taken.
- `ThumbnailUploadWorker` processes enqueued jobs by fetching the original image from R2, running `ThumbnailService.Generate`, and writing the result back. No new jobs of this kind will ever be enqueued.
- `ImageGrid` polls every 1 second while any loaded image has `thumbnail_url === null`. Since thumbnails are now set at insert time, `thumbnail_url` will never be null for newly uploaded images.

## Goals / Non-Goals

**Goals:**
- Delete all code that exists solely to support the old async thumbnail path
- `CompleteUpload` sets `thumbnail_path` unconditionally — no conditional logic, no R2 check
- Gallery query has no polling — `refetchInterval` option removed entirely
- Clean up orphaned River jobs in the database before deploying

**Non-Goals:**
- Changing the thumbnail format, size, or quality (done in previous change)
- Changing `InitiateUpload` (the presigned URL response shape stays the same)
- Adding any new thumbnail capability

## Decisions

### Decision: Set `thumbnail_path` unconditionally in `CompleteUpload`

After removing the `HeadObject` check, `CompleteUpload` computes the thumbnail key (same formula as `InitiateUpload`: `users/{userID}/thumbnails/{imageID}.jpg`) and assigns it to `img.ThumbnailPath` unconditionally. The image is always created with `thumbnail_path` set.

**Alternatives considered:**
- Keep a lightweight existence check as a defensive guard. Rejected: adds an extra R2 API call on every upload, and there is no longer a code path that would fail to upload the thumbnail — the contract is that `complete` is only called after both PUTs succeed.

### Decision: Remove `refetchInterval` option entirely (not set to `false`)

Removing the option entirely is cleaner than leaving `refetchInterval: false` — the default is already `false`, and no option is clearer intent than an explicit false value.

### Decision: One-time SQL to discard orphaned jobs before deploying (option B)

Any `thumbnail_upload` jobs already in the River queue will become permanently stuck once the worker is unregistered. Running `DELETE FROM river_jobs WHERE kind = 'thumbnail_upload'` before the deploy discards them cleanly. These jobs are for images that were already either processed successfully or abandoned — no data loss occurs.

**Alternatives considered:**
- Keep the worker registered for a grace period. Rejected: extends the scope of this change across multiple deploys for no practical benefit since the jobs can't process images that aren't affected.
- Add a catch-all discard worker. Rejected: adds new code to remove code, and the SQL is simpler and immediate.

### Decision: Run `go mod tidy` to remove `disintegration/imaging`

`pkg/thumbnail/thumbnail.go` is the only file importing `disintegration/imaging`. Deleting it makes the import dangling. `go mod tidy` removes it from `go.mod` and `go.sum`.

## Risks / Trade-offs

- **Images with existing `thumbnail_url = null` in the database** → These are pre-existing images uploaded via the old extension before this change. They will remain without a thumbnail permanently. No new images will be affected. Acceptable — these are a small, bounded set of legacy records.

- **`thumbnail_path` is set but the object doesn't exist in R2** → This would only happen if a client calls `complete` without uploading the thumbnail first. That would be a bug in the caller, not a normal state. The existing `InitiateUpload`/`CompleteUpload` contract assumes the client honors the presigned URLs it's given.

## Migration Plan

1. Run SQL before deploying the new backend: `DELETE FROM river_jobs WHERE kind = 'thumbnail_upload'`
2. Deploy backend (worker unregistered, `CompleteUpload` simplified, `HeadObject` removed)
3. Deploy frontend (refetchInterval removed)
4. Extension is already deployed

Rollback: if the backend deploy needs to be reverted, the worker and HeadObject check can be restored from git. The deleted River jobs are harmless to lose — they cannot be meaningfully replayed without re-uploading the original images.
