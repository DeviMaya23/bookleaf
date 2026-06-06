## 1. Rewrite thumbnail generation

- [x] 1.1 Replace `generateThumbnail` in `background/index.ts` to return a `Blob` instead of a base64 string, using 600px max fit (aspect-ratio preserving), `image/jpeg`, quality 0.9

## 2. Update save flow

- [x] 2.1 Update `saveImage` to destructure `thumbnail_upload_url` from the `POST /images` response
- [x] 2.2 Accept the thumbnail `Blob` as a parameter in `saveImage` and add a parallel `PUT` of the thumbnail blob to `thumbnail_upload_url` alongside the existing image `PUT`
- [x] 2.3 Update `handleSave` to call `generateThumbnail` before `saveImage` (guarded by `typeof OffscreenCanvas !== "undefined"`), pass the blob to `saveImage`, and convert the blob to base64 for `addRecentSave` after save completes

## 3. Unit tests

- [x] 3.1 ~~Update or add unit tests for `generateThumbnail`: assert it returns a `Blob` of type `image/jpeg` with dimensions within 600px bounds~~ — skipped, no test framework in extension; logic mirrors the already-tested FE `generateThumbnail`
