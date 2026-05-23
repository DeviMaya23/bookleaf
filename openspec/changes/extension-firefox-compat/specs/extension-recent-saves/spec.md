## MODIFIED Requirements

### Requirement: Thumbnail generation after successful save

After `POST /images/:id/complete` returns a successful response in `background/index.ts`, the background script SHALL attempt to generate a 60×60 JPEG thumbnail from the image blob only when `OffscreenCanvas` is available in the current execution context (`typeof OffscreenCanvas !== "undefined"`). Thumbnail generation SHALL NOT block the save result or the browser notification — it runs after both are dispatched, wrapped in its own try/catch so any failure is silent (logged only).

When `OffscreenCanvas` is available, thumbnail generation steps are:
1. Call `createImageBitmap(blob)` to decode the image
2. Create a 60×60 `OffscreenCanvas`
3. Draw the image scaled to cover the canvas, centred (cover crop)
4. Call `canvas.convertToBlob({ type: 'image/jpeg', quality: 0.7 })`
5. Read the blob as `ArrayBuffer` and convert to a base64 data URL using chunked `btoa` (chunk size: 8192 bytes)
6. Call `addRecentSave({ imageId, title, dataUrl, savedAt: Date.now() })`

When `OffscreenCanvas` is NOT available (Firefox), the script SHALL call `addRecentSave({ imageId, title, dataUrl: "", savedAt: Date.now() })` immediately after the save completes, without attempting thumbnail generation.

#### Scenario: Thumbnail is stored after a successful save on Chrome

- **WHEN** the 3-step upload sequence completes without error and `OffscreenCanvas` is available
- **THEN** a 60×60 JPEG thumbnail is generated from the image blob
- **AND** a new entry is prepended to `recentSaves` in extension storage with a non-empty `dataUrl`

#### Scenario: Save entry stored with empty dataUrl on Firefox

- **WHEN** the 3-step upload sequence completes without error and `OffscreenCanvas` is not available
- **THEN** a success notification is shown
- **AND** `addRecentSave` is called with `dataUrl: ""`
- **AND** the entry is prepended to `recentSaves` in extension storage

#### Scenario: Thumbnail failure does not affect the save result

- **WHEN** `createImageBitmap` or `OffscreenCanvas.convertToBlob` throws
- **THEN** the browser success notification is still shown
- **AND** no entry is added to `recentSaves`
- **AND** the error is logged to the console

#### Scenario: Thumbnail is not generated when save fails

- **WHEN** any step in the 3-step upload sequence throws
- **THEN** no thumbnail generation is attempted
- **AND** no entry is added to `recentSaves`

### Requirement: Recently saved thumbnail strip in popup

The popup SHALL read `recentSaves` from extension storage on mount. If the array is non-empty, it SHALL render up to 5 entries in a horizontal strip. For each entry:
- If `dataUrl` is a non-empty string, render an `<img>` element with `src={dataUrl}`.
- If `dataUrl` is an empty string, render a placeholder box with the same dimensions (flex: 1, aspectRatio: 1, borderRadius: 7) using the current theme's `divider` colour as the background. No `<img>` element is rendered for placeholder entries.

If the array is empty or absent, the empty state is shown instead.

#### Scenario: Thumbnail strip renders stored thumbnails

- **WHEN** the popup opens and `recentSaves` contains 3 entries with non-empty `dataUrl`
- **THEN** 3 square thumbnail images are rendered in the strip

#### Scenario: Placeholder rendered for entry with empty dataUrl

- **WHEN** the popup opens and a `recentSave` entry has `dataUrl: ""`
- **THEN** a themed placeholder box is rendered in place of an `<img>` for that entry
- **AND** the placeholder uses the current theme's divider colour as its background

#### Scenario: Empty state shown when no recent saves

- **WHEN** the popup opens and `recentSaves` is empty
- **THEN** the empty state message is shown and no thumbnail strip is rendered
