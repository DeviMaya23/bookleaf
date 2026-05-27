## Purpose

Defines how recently saved image thumbnails are generated, stored, and surfaced in the browser extension popup, including the storage schema and helpers for dark mode and avatar persistence.

---

## Requirements

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
