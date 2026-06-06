## Why

The browser extension still uses the old upload flow where the backend worker generates the thumbnail asynchronously after upload, leaving `thumbnail_url` null until the worker completes. The new presigned-URL thumbnail flow (already in place for the web app) allows the client to generate and upload the thumbnail directly, so the image is gallery-ready immediately after save.

## What Changes

- The extension background service worker generates a 600×600px max JPEG thumbnail from the fetched image blob before completing the upload
- The extension uploads the thumbnail blob directly to `thumbnail_upload_url` (returned by `POST /images`) as a parallel step alongside the original image upload
- The extension reuses the same thumbnail blob (converted to base64) for local popup "Recent Saves" display, replacing the current 60×60 crop
- The save flow becomes a 4-step sequence: initiate → parallel PUT (image + thumbnail) → complete

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `extension-save-image`: upload flow gains a thumbnail generation and upload step; thumbnail parameters (size, quality) change to match the web app standard

## Impact

- `extensions/src/background/index.ts`: `generateThumbnail` rewritten to return a Blob at 600px max; `saveImage` gains a parallel thumbnail PUT; `handleSave` converts the blob to base64 for `addRecentSave`
- No backend changes (the `thumbnail_upload_url` field is already returned by `POST /images`; the `HeadObject` fallback in `CompleteUpload` continues to serve as a safety net)
- No frontend changes
