## Why

The backend currently only supports JPEG and PNG — WebP appears accepted in the frontend but silently fails during upload completion because the Go thumbnail decoder has no WebP support. Expanding format support requires moving thumbnail generation to the frontend to avoid CGO dependencies that are incompatible with the Alpine-based Docker setup and Cloud Run deployment.

## What Changes

- `POST /images` (initiate upload) now returns a second presigned URL (`thumbnail_upload_url`) for the client to PUT a JPEG thumbnail directly to R2
- Frontend generates thumbnails via canvas before calling `/complete`, removing all server-side image decoding from the upload path
- Backend `CompleteUpload` performs a HEAD check on the thumbnail R2 path: if found, pre-sets `thumbnail_path` and skips the thumbnail worker; if not found (extension compatibility), leaves `thumbnail_path` null and enqueues the worker as before
- Accepted formats expanded: WebP, AVIF, and HEIC (Safari only) added to the frontend allowlist
- HEIC is converted to JPEG (93% quality) before upload — the HEIC original is never stored; what gets saved is the converted JPEG
- WebP and AVIF originals are stored as-is; only the thumbnail is JPEG
- Backend `MimeTypeToExt` and `downloadFileExtension` updated to cover WebP and AVIF
- **BREAKING**: Clients that call `/complete` without uploading a thumbnail first will fall back to the existing worker path (backward compat preserved for extension)

## Capabilities

### New Capabilities

- `fe-image-format-support`: Frontend format detection, per-format upload behaviour (store-as-is vs. convert), HEIC-to-JPEG conversion on Safari, AVIF/WebP pass-through

### Modified Capabilities

- `fe-image-upload-flow`: Initiate response now includes `thumbnail_upload_url`; client uploads thumbnail to R2 before calling `/complete`
- `image-thumbnail`: Thumbnail is generated and uploaded by the client; backend HEAD-checks R2 and conditionally skips the thumbnail worker
- `pending-uploads`: Initiate response contract extended with `thumbnail_upload_url`

## Impact

- **Backend**: `image_upload_usecase.go` (InitiateUpload, CompleteUpload), `storage/storage.go` (MimeTypeToExt), `image_usecase.go` (downloadFileExtension), handler response struct
- **Frontend**: `dragHandlers.ts` (ACCEPTED_TYPES, upload flow), new thumbnail generation utility, `images.ts` (initiateUpload response type)
- **Extension**: No changes in this proposal; the existing worker fallback keeps extension saves functional. Extension update is a follow-up.
- **Dependencies**: No new Go dependencies; no CGO; Dockerfile unchanged
