## Purpose

Defines how recently saved image thumbnails are generated, stored, and surfaced in the browser extension popup, including the storage schema and helpers for dark mode and avatar persistence.

---

## Requirements

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

---

### Requirement: Recent saves storage

The storage layer SHALL maintain a `recentSaves` array in `chrome.storage.local`. The array SHALL be capped at 5 entries. When a new entry is added, it SHALL be prepended and the array sliced to the 5 most recent entries.

Each entry shape:
```ts
interface RecentSave {
  imageId: string
  title: string
  dataUrl: string       // data:image/jpeg;base64,...
  savedAt: number       // Date.now() at save time
}
```

`storage.ts` SHALL expose:
- `getRecentSaves(): Promise<RecentSave[]>` — returns the array, or `[]` if absent
- `addRecentSave(entry: RecentSave): Promise<void>` — prepends and slices to 5

#### Scenario: addRecentSave prepends and caps at 5

- **WHEN** `addRecentSave` is called and `recentSaves` already contains 5 entries
- **THEN** the new entry is at index 0
- **AND** the array length remains 5 (oldest entry is dropped)

#### Scenario: getRecentSaves returns empty array when no saves exist

- **WHEN** `getRecentSaves` is called and `bookleaf_recent_saves` is absent from storage
- **THEN** the function resolves with `[]`

---

### Requirement: Dark mode storage helpers

`storage.ts` SHALL expose:
- `getDarkMode(): Promise<boolean>` — returns stored value or `false` if absent
- `setDarkMode(value: boolean): Promise<void>` — persists under `bookleaf_dark_mode`
- `getAvatar(): Promise<string | null>` — returns stored `bookleaf_avatar` URL or `null`
- `setAvatar(url: string): Promise<void>` — persists under `bookleaf_avatar`

#### Scenario: getDarkMode returns false when key is absent

- **WHEN** `getDarkMode` is called and `bookleaf_dark_mode` is absent from storage
- **THEN** the function resolves with `false`

#### Scenario: setDarkMode persists the value

- **WHEN** `setDarkMode(true)` is called
- **THEN** `bookleaf_dark_mode` is set to `true` in `chrome.storage.local`

#### Scenario: getAvatar returns null when key is absent

- **WHEN** `getAvatar` is called and `bookleaf_avatar` is absent from storage
- **THEN** the function resolves with `null`

---

### Requirement: Recently saved thumbnail strip in popup

The popup SHALL read `recentSaves` from extension storage on mount. If the array is non-empty, it SHALL always render exactly 5 slots in a horizontal strip, regardless of how many saves exist. This keeps the strip a consistent width at all fill levels.

For each slot:
- If a `recentSave` entry exists for that slot and its `dataUrl` is a non-empty string, render an `<img>` element with `src={dataUrl}`.
- Otherwise (entry has `dataUrl: ""`, or the slot has no entry yet), render a themed placeholder box with the same dimensions (flex: 1, aspectRatio: 1, borderRadius: 7) using the current theme's `divider` colour as the background.

A "View all" link SHALL appear in the section header when at least one save exists. Clicking it SHALL open the app at the `/all` path in a new tab.

If `recentSaves` is empty or absent, the empty state is shown instead (no strip, no "View all" link).

#### Scenario: Strip always renders 5 slots

- **WHEN** the popup opens and `recentSaves` contains 1 entry with a non-empty `dataUrl`
- **THEN** 1 thumbnail image and 4 placeholder boxes are rendered in the strip

#### Scenario: Placeholder rendered for entry with empty dataUrl

- **WHEN** the popup opens and a `recentSave` entry has `dataUrl: ""`
- **THEN** a themed placeholder box is rendered in place of an `<img>` for that entry
- **AND** the placeholder uses the current theme's divider colour as its background

#### Scenario: View all opens app at /all

- **WHEN** the user clicks "View all" in the recently saved section
- **THEN** a new tab opens to `VITE_APP_URL/all`

#### Scenario: Empty state shown when no recent saves

- **WHEN** the popup opens and `recentSaves` is empty
- **THEN** the empty state message is shown and no thumbnail strip is rendered
