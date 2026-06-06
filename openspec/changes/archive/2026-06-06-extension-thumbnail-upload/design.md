## Context

The extension's `handleSave` flow in `background/index.ts` currently performs:
1. `POST /images` → receives `{ upload_url, id }` (ignores `thumbnail_upload_url`)
2. `PUT upload_url` with the original image blob
3. `POST /images/:id/complete`

The backend then runs a `ThumbnailUploadWorker` job asynchronously to generate and store the thumbnail. Until the job completes, `thumbnail_url` is null on the image record and the gallery view polls every second waiting for it.

The `POST /images` response already includes `thumbnail_upload_url` — the extension simply never used it. The web app uses this URL as part of a parallel upload step, and images from the web app are gallery-ready immediately after `complete`.

The extension also contains a `generateThumbnail` function that produces a 60×60 cover-cropped base64 data URL. This is used exclusively for the popup's "Recent Saves" display — it was never uploaded to R2.

## Goals / Non-Goals

**Goals:**
- Extension uploads a proper thumbnail to R2 as part of the save flow
- Saved images are gallery-ready immediately, without relying on the worker
- Popup "Recent Saves" display continues to work, showing the same thumbnail image at 60px via CSS
- The change is self-contained to `background/index.ts`; the backend `HeadObject` safety net remains in place

**Non-Goals:**
- Removing backend backwards-compatibility code (HeadObject check, ThumbnailUploadWorker) — deferred
- Removing the FE gallery polling refetchInterval — deferred
- HEIC support in the extension (not applicable; extension fetches page images which are browser-renderable)

## Decisions

### Decision: Rewrite `generateThumbnail` to return a Blob

The current function produces a 60×60 JPEG via `OffscreenCanvas.convertToBlob()` then converts it to base64. The new function produces a 600px max fit JPEG blob — matching the parameters in `frontend/src/lib/thumbnail.ts` (max 600px, quality 0.9, aspect-ratio preserving).

The base64 conversion for `addRecentSave` becomes a separate step in `handleSave`, done from the same blob after R2 upload.

**Alternatives considered:**
- Keep a separate 60×60 generator for popup and add a new 600px generator for R2. Rejected: two canvas operations for one save; the 600px blob served at 60px CSS size is indistinguishable visually.

### Decision: Upload thumbnail in parallel with the original image

The main image PUT and thumbnail PUT are independent — neither depends on the other's result. They run concurrently via `Promise.all`, the same pattern the web app uses.

**Alternatives considered:**
- Sequential upload (thumbnail after original). Rejected: adds unnecessary latency.

### Decision: Thumbnail upload failure fails the entire save

If the thumbnail PUT fails, `saveImage` throws and `handleSave` shows the error toast — the same outcome as any other upload failure. The image record is not created (complete is never called).

**Alternatives considered:**
- Upload thumbnail as best-effort, proceed to complete even if it fails. Rejected: would silently create images with no thumbnail and no worker fallback once BC is removed.

### Decision: Base64 conversion happens after R2 upload in `handleSave`

The flow in `handleSave`:
1. `saveImage()` returns `{ imageId, thumbnailBlob }`
2. `handleSave` calls `addRecentSave` converting `thumbnailBlob` to base64 inline

This keeps `saveImage` focused on the upload protocol and leaves display concerns to the caller.

## Risks / Trade-offs

- **Chrome service worker eviction mid-upload** → The browser will not evict a service worker during active network I/O. Completing the full 4-step sequence (initiate + 2 PUTs + complete) within a single event handler is the standard pattern for MV3 extensions and is safe. The backend `HeadObject` fallback remains as a catch-all for any unexpected partial upload.

- **Stored base64 size increase** → Recent saves stores up to 5 entries. A 600px JPEG at quality 0.9 is typically 50–200 KB. At 5 entries: ≤ 1 MB. `storage.local` quota is 10 MB (Chrome) / effectively unlimited (Firefox). Acceptable.

- **`createImageBitmap` on fetched blobs** → Works for all formats a browser can render (JPEG, PNG, WebP, GIF, AVIF). The extension fetches images that are already being displayed on the page, so the blob will always be a renderable format.
