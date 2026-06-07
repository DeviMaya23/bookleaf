## Context

`images.width`, `images.height`, and `images.file_size` are nullable columns populated today by `imageUploadUsecase.CompleteUpload` via `extractImageMetadata`, which fetches the uploaded bytes from R2 and runs Go stdlib `image.DecodeConfig` (with `image/jpeg` and `image/png` blank-imported). Any format outside that pair — AVIF, WebP, GIF, HEIC — fails to decode silently, and `width`/`height` are persisted as `NULL`. `file_size` is unaffected (it's just `len(bytes)`).

This is the same CGO-without-libvips constraint that recently moved thumbnail generation to the client (`2026-06-06-remove-thumbnail-backwards-compat`). Both the web frontend (`frontend/src/lib/thumbnail.ts`) and the browser extension (`extensions/src/background/index.ts`) already call `createImageBitmap()` to build a thumbnail — which means both already decode every uploaded image and have `width`/`height` on the resulting `ImageBitmap`. Today those values are used only to scale the thumbnail canvas and then discarded.

The upload sequence (both FE and extension) is: `POST /images` (initiate, creates a `pending_uploads` row + presigned URLs) → generate thumbnail (decodes via `createImageBitmap`) → PUT original + thumbnail to R2 → `POST /images/:id/complete` (finalizes the `domain.Image` row from the `pending_uploads` row, deletes the pending row).

## Goals / Non-Goals

**Goals:**
- Persist correct `width`/`height`/`file_size` for every image format the app accepts, including those the backend cannot decode.
- Remove the backend's now-redundant, format-limited decode path entirely (no dual-source-of-truth, no fallback to keep in sync).
- Keep the change to existing nullable columns — no schema migration.

**Non-Goals:**
- Backfilling `width`/`height`/`file_size` for images already stored with `NULL` values (future uploads only — `ImageViewer.tsx` already degrades gracefully when these are missing).
- Changing where `mime_type` is supplied (`InitiateUpload` keeps it — it's structurally required there to build the presigned PUT URL's `Content-Type`).
- Adding any new decode step on the client — FE and extension already decode the image for thumbnail generation; this change only captures and forwards values that already exist in memory.

## Decisions

### 1. Carry `width`/`height`/`file_size` on `CompleteUpload`, not `InitiateUpload`

`InitiateUpload` creates a row in `pending_uploads` — an ephemeral "promise of an upload" that is read once and deleted by `CompleteUpload`. `mime_type` lives there because the server has an immediate, structural need for it: it cannot generate a presigned PUT URL with the correct `Content-Type` without it. `width`/`height`/`file_size` have no such need at initiation time — moreover, they describe *the delivered bytes*, which don't exist yet (from the server's perspective) when `InitiateUpload` runs. Recording them on the "promise" entity would mean asserting facts about goods that haven't arrived.

Putting them on `CompleteUpload` means filling the *same* `domain.Image.Width/Height/FileSize` fields that already exist, just sourced from the request body instead of `extractImageMetadata`. **No new columns, no new tables, no migration** — and `extractImageMetadata` and its stdlib decode imports can be deleted outright.

Alternative considered: add `width`/`height`/`file_size` to `pending_uploads` and thread them through at initiation. Rejected — it requires a migration to a table whose entire lifecycle is "insert → read once → delete," purely to shuttle values that the client could just as easily hand over at completion, after it has actually decoded the bytes it's about to upload.

### 2. Implausible values are sanitized to `NULL`, never reject the completion

By the time `CompleteUpload` runs, the original and thumbnail have already been PUT to R2 and the `pending_uploads` row exists. Rejecting the call over a malformed `width: -1` would strand an already-uploaded file and a dangling pending row over what is fundamentally cosmetic metadata (a display badge and zoom-fit calculation in `ImageViewer.tsx`, which already defaults to `0.5` zoom when dimensions are missing).

So the usecase treats client-supplied `width`/`height`/`file_size` as *optional, best-effort* values: positive integers are persisted; anything else (`<= 0`, missing/absent) is stored as `NULL`, exactly mirroring today's "decode failed → NULL" outcome. The upload always completes successfully when the R2 writes and DB transaction succeed — bad metadata is never a reason to fail an otherwise-successful upload.

Alternative considered: validate and reject with a 4xx if values look implausible. Rejected — there's no clean recovery path for the client at that point (the bytes are already in R2; re-running `CompleteUpload` with corrected values is the only option, and the failure mode being guarded against — a client sending nonsense — degrades to exactly the same `NULL` outcome as today's unsupported-format path anyway).

### 3. No fallback decode path on the backend

`extractImageMetadata` is deleted rather than kept as a server-side fallback/verification. A fallback would mean two sources of truth for the same fields with no clear precedence rule, plus keeping format-specific decode imports alive for a code path that — per decision 2 — can never improve on "trust the client or store NULL." Simpler to delete it: one source of truth (the client that already decoded the image), one degradation path (`NULL`).

### 4. Extension: omit dimensions (not zeros) when `OffscreenCanvas` is unavailable

The extension only calls `createImageBitmap` inside `generateThumbnail`, which itself only runs `if (typeof OffscreenCanvas !== "undefined")`. In environments without `OffscreenCanvas`, no decode happens — there is no bitmap to read `width`/`height` from. In that case the extension sends `file_size` (from `blob.size`, no decode required) but omits `width`/`height` from the `complete` payload, which the backend treats as absent → `NULL`. This is a pre-existing degradation path (today these images get `NULL` dimensions from a failed server-side decode); the only thing that changes is *why* they're `NULL`.

## Risks / Trade-offs

- **[Risk]** A misbehaving or malicious client could send arbitrary `width`/`height`/`file_size`. → **Mitigation**: these fields are purely descriptive (display + zoom-fit UX), not used for authorization, storage accounting, or any security-relevant decision; sanitization to positive-integers-or-NULL (decision 2) is the appropriate level of trust for cosmetic metadata. `file_size` in particular is already trivially knowable to the client (`blob.size` / the bytes it's uploading).
- **[Risk]** Removing `extractImageMetadata` means the backend permanently loses the ability to independently verify these fields. → **Mitigation**: it could already only verify JPEG/PNG (a minority of supported formats going forward), and verification was never enforced (decode failure was already silently tolerated). Nothing of practical value is lost.
- **[Trade-off]** Pre-existing images with `NULL` dimensions remain `NULL` (no backfill). → Accepted per Non-Goals; `ImageViewer.tsx` already degrades gracefully, and a backfill would require re-fetching and decoding every historical image — disproportionate to the problem.

## Migration Plan

No database migration. Deploy order matters only in the loose sense that this is a contract addition (new optional request fields), not a removal:
1. Ship the backend change first — `CompleteUpload` accepts optional `width`/`height`/`file_size` in the request body, falling back to `NULL` when absent (which is exactly today's behavior for any client that doesn't send them).
2. Ship FE and extension changes to start sending the values.

This ordering means an old client talking to a new backend behaves exactly as it does today (fields absent → `NULL`), and there's no window where a new client sends fields a not-yet-updated backend doesn't understand and rejects (it simply ignores unknown JSON fields). Rollback is symmetric — reverting either side independently is safe.

## Open Questions

None — placement, validation behavior, and fallback strategy were resolved during exploration (see `proposal.md` Why/What Changes).
