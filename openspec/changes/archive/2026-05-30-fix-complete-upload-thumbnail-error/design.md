## Context

`CompleteUpload` fetches the uploaded image from R2, generates a thumbnail synchronously, then marks the image as `is_uploaded = true`. If thumbnail generation fails, the current code sets a `Warning` string and returns `(result, nil)` — a `200 OK` that the frontend treats as success. The image record remains with `is_uploaded = false` and is eventually purged by the stale cleanup job.

The frontend already handles HTTP errors correctly: `UploadModal` has an `onError` toast, and `BatchUploadModal` has a `catch` block that retries then marks `FAILED_FINAL`.

## Goals / Non-Goals

**Goals:**
- `CompleteUpload` returns a real error when thumbnail generation fails
- Both upload flows surface failure correctly through their existing error paths
- Specs reflect the corrected behavior

**Non-Goals:**
- Retrying thumbnail generation server-side
- Frontend UI changes
- Investigating the root cause of the specific PNG failure (tracked separately)

## Decisions

**Return the error directly from `CompleteUpload`**

When `prepareThumbnail` fails, return the error instead of swallowing it. The HTTP handler already maps usecase errors to `500 Internal Server Error`. No handler changes needed.

Alternative considered: return a structured 4xx error. Rejected — thumbnail failure is a server-side processing problem, not a client mistake. `500` is correct.

**Remove `Warning` from the thumbnail failure path**

The `Warning` field on `CompleteUploadResult` is only meaningful for vision labelling failures (which are non-blocking). Thumbnail failure is now an error, so `Warning` is no longer set in that branch. The field and its vision usage are retained unchanged.

**No `is_uploaded` fallback**

Thumbnail is mandatory. An image without a thumbnail MUST NOT be marked as uploaded. The stale cleanup removing it is the intended recovery path.

## Risks / Trade-offs

- **Retry on full re-upload**: `BatchUploadModal` retries by re-running `initiateUpload → putToR2 → completeUpload`, which re-uploads the file. If thumbnail failure is systemic (e.g. corrupt file), the retry will fail again and land on `FAILED_FINAL`. Acceptable — retry is one attempt, not a loop.
- **`Warning` field becomes partially unused**: It remains in `CompleteUploadResult` for vision failures. Worth revisiting if vision is ever removed, but not worth cleaning up now.
