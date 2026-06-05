## Context

The current upload completion path (`POST /images/:id/complete`) fetches the original image from R2 and decodes it in Go twice: once as a synchronous preflight (`thumbnails.Generate`) that gates the HTTP response, and once inside the `ThumbnailUploadWorker`. Go's `image` package only has decoders registered for JPEG and PNG — adding WebP/AVIF/HEIC without CGO is not possible, and CGO is incompatible with the Alpine-based Dockerfile and Cloud Run deployment.

The fix moves thumbnail generation to the client. The browser natively decodes all modern formats. The thumbnail (always JPEG) is PUT directly to R2 before `/complete` is called, removing the need for server-side image decoding entirely.

## Goals / Non-Goals

**Goals:**
- Accept WebP and AVIF originals, stored as-is in R2
- Accept HEIC on Safari, converting to JPEG before upload (original HEIC is never stored)
- Remove server-side image decoding from the upload completion path
- Maintain backward compatibility with the browser extension (worker fallback)

**Non-Goals:**
- Updating the browser extension to use the new thumbnail upload URL (separate proposal)
- Supporting HEIC on non-Safari browsers
- Server-side AVIF or HEIC decoding
- Changing thumbnail dimensions or encoding parameters

## Decisions

### 1. Two presigned URLs from InitiateUpload

`POST /images` will return both `upload_url` (original) and `thumbnail_upload_url` (presigned PUT for `users/{userID}/thumbnails/{imageID}.jpg`, content type `image/jpeg`). The thumbnail R2 path is deterministic at initiation time (imageID is already generated), so issuing both URLs in one round trip is straightforward.

**Alternative considered:** A separate `POST /images/:id/thumbnail-url` endpoint. Rejected — extra round trip with no benefit; the path is always known at initiate time.

### 2. HEAD check in CompleteUpload to conditionally skip the worker

`CompleteUpload` performs a `StorageService.HeadObject` on the thumbnail R2 path. If the object exists, `thumbnail_path` is set at image creation time and the `ThumbnailUploadWorker` job is not enqueued. If the object is absent (extension flow, or FE failure), `thumbnail_path` remains null and the worker is enqueued as before.

This preserves extension functionality without any protocol change. The extension's existing flow (`POST /images` → PUT original → `POST /complete`) continues to work; it simply doesn't upload a thumbnail, so the worker handles it.

**Alternative considered:** A `thumbnail_uploaded: bool` flag in the `/complete` request body. Rejected — HEAD check is self-verifying and requires no client cooperation; a flag could lie.

**Alternative considered:** Remove the worker entirely and accept extension breakage. Rejected — the user explicitly requires extension saves to keep working until the extension proposal ships.

### 3. Remove the synchronous thumbnail preflight from CompleteUpload

The current spec (`image-thumbnail`) requires a synchronous `ThumbnailService.Generate` call inside `CompleteUpload` that gates the HTTP response — if it fails, the whole upload fails. This preflight is removed. Its only purpose was early failure detection; with client-generated thumbnails, the server never decodes the original at all.

The `extractImageMetadata` call (for width/height via `stdimage.DecodeConfig`) remains but is already a soft failure — unknown formats yield null dimensions, which is acceptable.

### 4. HEIC → JPEG conversion on the client (Safari only)

HEIC files are converted to JPEG (93% quality) via canvas before upload. The stored original is JPEG; the MIME type sent to the backend is `image/jpeg`. The backend never sees `image/heic`. HEIC is excluded from the accepted-types list on non-Safari browsers by checking `navigator.userAgent` for Safari at runtime.

**Why JPEG over WebP for conversion:** The stated goal is cross-device availability. JPEG is truly universal; WebP still breaks in some email clients and older Android gallery apps.

**Alternative considered:** Store HEIC and convert via a separate transcoding service. Rejected — introduces infrastructure complexity; client conversion is free and produces the same result.

### 5. WebP and AVIF stored as-is

These formats have near-universal modern browser support (~97% and ~90% respectively). The browser can decode them natively for thumbnail generation. Originals are uploaded unchanged, consistent with the principle of not tampering with high-res images.

### 6. Thumbnail canvas sizing

The client generates thumbnails at a maximum of 600×600px, preserving aspect ratio — matching the existing server-side spec. Format is always `image/jpeg` at quality 0.9.

## Risks / Trade-offs

**[Memory pressure on large images]** → A 48MP RAW-equivalent image decoded to canvas is ~180MB of raw pixel data. In practice, HEIC from iPhone cameras is 12MP (~46MB), which is manageable on modern Safari. Extreme cases (high-end camera apps) could cause OOM on iOS. Mitigation: draw to canvas at a capped intermediate size (e.g., 4096px on the longest side) before encoding if the source exceeds a threshold.

**[Thumbnail missing if FE fails mid-upload]** → If the client uploads the original but fails to upload the thumbnail before calling `/complete`, the HEAD check will miss and fall back to the worker. Worker only succeeds for JPEG and PNG (current decoders). For WebP/AVIF, the worker would fail and the image would have no thumbnail. Mitigation: FE should upload thumbnail before calling `/complete`; if thumbnail upload fails, the full upload should be aborted rather than completed.

**[Safari HEIC detection via user agent]** → UA sniffing is fragile. A more reliable check is `document.createElement('canvas').toDataURL('image/heic')` to detect HEIC encode support, or attempting `createImageBitmap` on a 1-byte blob. However, these are async and more complex. UA check is pragmatic for now.

**[Worker still decodes for extension saves]** → The worker still calls `ThumbnailService.Generate` for extension-saved images (JPEG/PNG only). This is unchanged behaviour and not a regression. Extension format support is constrained to what the worker can handle until the extension proposal ships.

## Migration Plan

No database schema changes required. The `thumbnail_path` column already exists and is nullable. No new migrations.

Deployment is a standard rolling deploy — the new `thumbnail_upload_url` field in the initiate response is additive. Old clients that don't use it will trigger the worker fallback transparently.
