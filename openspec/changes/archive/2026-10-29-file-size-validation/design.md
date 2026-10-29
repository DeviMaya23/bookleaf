# Design

## Context

See `proposal.md — Why` for motivation.

The upload flow is a two-step client-driven protocol: `POST /images` (initiate) returns a presigned PUT URL; the client PUTs directly to R2; then `POST /images/:id/complete` finalises. The backend never sees the file bytes — only metadata the client reports. The presigned URL today signs only `Content-Type`, leaving no storage-level enforcement on body size.

Five upload surfaces exist across three layers:
- **Frontend**: `UploadModal` (single file), `BatchUploadModal` (multi-file), AppLayout auto-upload (single file drop), clipboard paste → `UploadModal`
- **Extension**: `persistImage()` — shared entry point for all extension pathways (context menu, drag, snip, video frame, image picker)
- **Backend**: `InitiateUpload` and `CompleteUpload` usecases

`BatchUploadModal` already has a 50 MB check in code (inline constant `MAX_SIZE_BYTES`) but the other surfaces do not.

## Goals / Non-Goals

**Goals:**
- Enforce 50 MB per image at the backend (both initiation and completion)
- Sign `Content-Length` into the presigned PUT URL so R2 rejects size mismatches
- Propagate client-side rejection to all upload surfaces with appropriate UX (inline error or toast)
- Consolidate the existing `BatchUploadModal` constant into the shared `upload.ts` module

**Non-Goals:**
- Per-user or per-plan configurable limits (hardcoded constant only)
- Thumbnail URL ContentLength signing (thumbnail size is unknown at initiate time; thumbnails are always small)
- Rate limiting or total storage quotas

## Decisions

### 1. file_size moves from CompleteUpload to InitiateUpload (and stays at complete too)

**Decision**: Clients send `file_size` in `POST /images` body. Backend validates ≤ 50 MB and refuses to issue a presigned URL otherwise. `CompleteUpload` also validates as a second gate.

**Why**: Signing `Content-Length` into the presigned PUT URL requires the exact size at URL-generation time, so the size must be known at initiation. Moving it earlier does not lose anything — all clients already have the full blob in memory before calling initiate (post-conversion for HEIC).

**Alternative considered**: Keep `file_size` only at complete, validate there. Rejected — this issues a presigned URL before validating, allows the client to PUT the full blob to R2 before the backend can reject it. An oversized object lands in storage even if the complete call fails; the stale-upload cleaner handles orphans but it's wasteful.

---

### 2. ContentLength signed in presigned PUT, not presigned POST

**Decision**: Pass `contentLength` to `GeneratePresignedPutURL`; the AWS SDK includes it in `X-Amz-SignedHeaders`; R2 enforces exact match.

**Why**: Presigned POST with `content-length-range` is the S3-native max-size mechanism, but it requires a protocol change (multipart form POST vs raw PUT) across every upload client. Signing `Content-Length` in the existing presigned PUT achieves equivalent enforcement — a client cannot swap in a larger body without breaking the signature — and requires no client protocol changes.

**Constraint**: This enforces exact size match, not a range. A client that declares 1 MB but tries to PUT 60 MB will be rejected by R2 signature validation. A client that declares the real size (which all honest clients do) is unaffected.

**Thumbnail URL**: ContentLength is NOT signed for the thumbnail presigned URL. The thumbnail is generated client-side after the initiate call, so its size is unknown at URL generation time. Thumbnails are always small (600px max, JPEG 0.9); there is no meaningful attack surface here.

---

### 3. Shared constant in frontend `upload.ts`

**Decision**: Export `MAX_FILE_SIZE_BYTES = 52_428_800` from `frontend/src/lib/upload.ts`. Extend `FileValidationError` with `'oversized'`. Update `validateImageFile()` to check size.

**Why**: `BatchUploadModal` already has a working 50 MB check with an inline constant. Extracting it to `upload.ts` gives all surfaces (single modal, auto-upload, batch) a single source of truth. `validateImageFile()` is already the shared validation entry point used by `UploadModal`, `BatchUploadModal`, and `dragHandlers.ts` — adding size there means all three inherit it automatically without per-surface changes.

**`BatchUploadModal` impact**: Its `makeBatchFile()` currently checks `file.size > MAX_SIZE_BYTES` independently of `validateImageFile()`. After this change, `validateImageFile()` returns `'oversized'` for oversized files, and `makeBatchFile()` maps that to `'OVERSIZED'` status — same observable behaviour, no inline constant.

---

### 4. Extension validates in `persistImage()`, not per-pathway

**Decision**: Single size check in `persistImage()` before the upload sequence.

**Why**: All five extension save pathways (context menu, drag, snip, video frame, image picker) converge on `persistImage()`. A single guard there covers all paths without touching each pathway's handler. The `saveImage()` call within `persistImage()` already holds the blob.

## Risks / Trade-offs

**Client can lie about file_size at initiation** → The presigned URL is signed for the declared size. If a client declares 1 MB, gets a URL, and tries to PUT 60 MB, R2 rejects it (Content-Length mismatch). The backend validate-at-initiate gate is the primary check; the signed URL is the R2-level enforcement; the validate-at-complete gate catches any residual mismatch. Coordinated misuse requires a valid auth token and is an unlikely attack vector for this application.

**HEIC → JPEG conversion changes file size** → The client converts HEIC to JPEG before calling initiate (`upload.ts` does this first). The `file_size` sent is the post-conversion blob size, which is what gets PUT to R2. No mismatch is possible.

**Breaking API change** → `POST /images` gains a required `file_size` field. Any client not sending it receives a 400. All known clients (frontend, extension) are updated in this change. No public API; no external consumers.

## Migration Plan

1. Deploy backend changes first (handler + usecase + storage interface). At this point `POST /images` starts requiring `file_size`. Existing clients without the field will get 400s on new uploads.
2. Deploy frontend and extension changes immediately after (ideally same deploy window). Both are updated to send `file_size`.
3. No database migration required. No existing records are affected.
4. Rollback: revert backend to accept `file_size` as optional (treat absent/zero as unvalidated, skip ContentLength signing). Frontend/extension rollback is optional — sending `file_size` to an older backend that ignores it is harmless.
