## MODIFIED Requirements

### Requirement: Thumbnail generation after successful save

After the 4-step upload sequence completes in `background/index.ts`, the background script SHALL store a recent save entry using the thumbnail blob generated during upload (if any).

Thumbnail generation happens **before** `POST /images/:id/complete` as part of the upload sequence, not as a post-save side effect. See `extension-save-image` for the full upload flow.

When `OffscreenCanvas` is available, thumbnail generation steps are:
1. Call `createImageBitmap(blob)` to decode the image
2. Compute scaled dimensions: fit within 600×600 pixels while preserving the original aspect ratio (`scale = Math.min(1, 600 / Math.max(width, height))`)
3. Create an `OffscreenCanvas` at the scaled dimensions
4. Draw the image scaled to fit
5. Call `canvas.convertToBlob({ type: 'image/jpeg', quality: 0.9 })`
6. Return the `Blob`

After save completes, `handleSave` SHALL convert the thumbnail blob to a base64 data URL using chunked `btoa` (chunk size: 8192 bytes) and call `addRecentSave({ imageId, title, dataUrl, savedAt: Date.now() })`.

When `OffscreenCanvas` is NOT available, no thumbnail blob is generated. The thumbnail PUT is skipped during upload. After save completes, `handleSave` SHALL call `addRecentSave({ imageId, title, dataUrl: "", savedAt: Date.now() })`.

Thumbnail generation failure (thrown by `createImageBitmap` or `convertToBlob`) SHALL propagate and fail the entire save — it is not wrapped in a silent try/catch.

#### Scenario: Thumbnail is stored after a successful save when OffscreenCanvas is available

- **WHEN** the 4-step upload sequence completes without error and `OffscreenCanvas` is available
- **THEN** a JPEG thumbnail blob (max 600px, aspect-ratio preserving) is generated from the image blob
- **AND** the thumbnail blob is uploaded to R2 via `thumbnail_upload_url`
- **AND** a new entry is prepended to `recentSaves` in extension storage with a non-empty `dataUrl`

#### Scenario: Save entry stored with empty dataUrl when OffscreenCanvas is not available

- **WHEN** the 4-step upload sequence completes without error and `OffscreenCanvas` is not available
- **THEN** a success toast is shown
- **AND** `addRecentSave` is called with `dataUrl: ""`
- **AND** the entry is prepended to `recentSaves` in extension storage

#### Scenario: Thumbnail generation failure fails the save

- **WHEN** `createImageBitmap` or `OffscreenCanvas.convertToBlob` throws during the upload sequence
- **THEN** the save fails
- **AND** an error toast is shown
- **AND** no entry is added to `recentSaves`

#### Scenario: Thumbnail is not stored when save fails

- **WHEN** any step in the 4-step upload sequence throws
- **THEN** no thumbnail generation is attempted for `addRecentSave`
- **AND** no entry is added to `recentSaves`
