## ADDED Requirements

### Requirement: Thumbnail generation after successful save

After `POST /images/:id/complete` returns a successful response in `background/index.ts`, the background service worker SHALL generate a 60×60 JPEG thumbnail from the image blob already held in memory. Thumbnail generation SHALL NOT block the save result or the Chrome notification — it runs after both are dispatched, wrapped in its own try/catch so any failure is silent (logged only).

Thumbnail generation steps:
1. Call `createImageBitmap(blob)` to decode the image
2. Create a 60×60 `OffscreenCanvas`
3. Draw the image scaled to cover the canvas, centred (cover crop)
4. Call `canvas.convertToBlob({ type: 'image/jpeg', quality: 0.7 })`
5. Read the blob as `ArrayBuffer` and convert to a base64 data URL using chunked `btoa` (chunk size: 8192 bytes)
6. Call `addRecentSave({ imageId, title, dataUrl, savedAt: Date.now() })`

#### Scenario: Thumbnail is stored after a successful save

- **WHEN** the 3-step upload sequence completes without error
- **THEN** a 60×60 JPEG thumbnail is generated from the image blob
- **AND** a new entry is prepended to `recentSaves` in `chrome.storage.local`

#### Scenario: Thumbnail failure does not affect the save result

- **WHEN** `createImageBitmap` or `OffscreenCanvas.convertToBlob` throws
- **THEN** the Chrome success notification is still shown
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

The popup SHALL read `recentSaves` from `chrome.storage.local` on mount. If the array is non-empty, it SHALL render up to 5 square thumbnails in a horizontal strip. Each thumbnail SHALL use the stored `dataUrl` as the `src` of an `<img>` element. If the array is empty or absent, the empty state is shown instead (see `extension-popup-ui` spec).

#### Scenario: Thumbnail strip renders stored thumbnails

- **WHEN** the popup opens and `recentSaves` contains 3 entries
- **THEN** 3 square thumbnail images are rendered in the strip

#### Scenario: Empty state shown when no recent saves

- **WHEN** the popup opens and `recentSaves` is empty
- **THEN** the empty state message is shown and no thumbnail strip is rendered
