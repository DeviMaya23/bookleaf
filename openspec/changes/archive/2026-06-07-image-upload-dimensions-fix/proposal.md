## Why

Image `width`/`height` are sometimes saved as `NULL` because the backend derives them with Go's stdlib `image.DecodeConfig`, which only understands JPEG/PNG. Formats like AVIF, WebP, GIF, and HEIC silently fail to decode — the same CGO-library gap that recently drove thumbnail generation to move client-side. The frontend and extension already decode every uploaded image (via `createImageBitmap`) to build thumbnails, so they have reliable `width`/`height`/`file_size` values in hand; the backend should stop guessing and accept these as part of completing the upload.

## What Changes

- **BREAKING**: `POST /images/:id/complete` now accepts an optional JSON body with `width`, `height`, and `file_size` (integers) supplied by the client, instead of deriving them server-side from the uploaded bytes.
- Remove `extractImageMetadata` (and the now-unused `image.DecodeConfig`/`image/jpeg`/`image/png` stdlib decoding) from `imageUploadUsecase.CompleteUpload`.
- Implausible client-supplied values (`<= 0`, or absent) SHALL be stored as `NULL` — the upload still completes successfully; bad metadata never blocks a finished upload. This mirrors today's "decode failed → NULL" degradation, which the image viewer already handles gracefully.
- Frontend (`UploadModal.tsx`, `BatchUploadModal.tsx`, `dragHandlers.ts` / `lib/thumbnail.ts`): capture `width`/`height` from the `createImageBitmap` decode already performed for thumbnail generation, capture `file_size` from the source `Blob`, and send all three on `completeUpload`. All three upload entry points (single upload, batch upload, and drag-and-drop) share the same `generateThumbnail` decode and `completeUpload` call, so each is updated identically — leaving any one of them behind would mean those uploads permanently persist `NULL` dimensions now that the backend no longer derives them server-side.
- Extension (`background/index.ts`): same adjustment — capture `width`/`height` from its own `createImageBitmap` decode and `file_size` from the blob, and send them on the `complete` call. When `OffscreenCanvas` is unavailable (no thumbnail/decode happens), omit dimensions — `file_size` is still sent from `blob.size`.
- No database migration: `images.width`, `images.height`, and `images.file_size` are already nullable columns; only the source of truth for populating them changes.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `image-endpoints`: `CompleteUpload` request contract changes — the usecase now accepts and persists client-supplied `width`/`height`/`file_size` (sanitized to `NULL` when implausible) instead of decoding the uploaded bytes server-side.
- `image-domain`: `Width`, `Height`, and `FileSize` on `domain.Image` are now populated from client-supplied values at upload completion rather than derived server-side.
- `fe-image-upload-flow`: the upload flow now computes and transmits `width`/`height`/`file_size` alongside the existing `completeUpload` call.
- `fe-batch-upload`: each file's upload sequence now computes and transmits `width`/`height`/`file_size` alongside its `completeUpload` call.
- `fe-drag-drop-file-upload`: the drop/auto-upload sequence now computes and transmits `width`/`height`/`file_size` alongside its `completeUpload` call.
- `extension-save-image`: the save flow now computes and transmits `width`/`height`/`file_size` alongside the existing `complete` call, with graceful omission of dimensions when `OffscreenCanvas` is unavailable.

## Impact

- **Backend**: `internal/usecase/image_upload_usecase.go` (`CompleteUpload`, `extractImageMetadata` removal), `internal/handler/image_upload.go` (request DTO for `/images/:id/complete`), Bruno collection update for the `complete` endpoint.
- **Frontend**: `frontend/src/components/UploadModal.tsx`, `frontend/src/components/BatchUploadModal.tsx`, `frontend/src/lib/dragHandlers.ts`, `frontend/src/lib/thumbnail.ts`, `frontend/src/lib/images.ts` (`completeUpload` signature/payload) — all three upload entry points share the same `generateThumbnail`/`completeUpload` plumbing and are updated together.
- **Extension**: `extensions/src/background/index.ts` (`generateThumbnail`, `saveImage`).
- **No schema migration** — `width`, `height`, `file_size` columns already exist and are nullable.
