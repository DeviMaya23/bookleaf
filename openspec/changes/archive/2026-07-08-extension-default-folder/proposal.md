## Why

Images saved from the browser extension always land in Unsorted, with no way to direct them to a specific folder. Users who organise their library into folders have to manually move every saved image after the fact.

## What Changes

- Extension settings panel gains a **"Save destination"** section (above Hotkeys) with a native `<select>` dropdown listing the user's folders plus a "None (Unsorted)" default option.
- The selected folder ID is persisted in `browser.storage.local`.
- All save paths in the extension (single right-click/drag/snip/video-frame and batch picker) read the stored folder ID and pass it to `POST /images` as `folder_id`.
- Folder list is fetched from `GET /folders` each time the Settings panel mounts.
- If the stored folder no longer exists, the backend already silently falls back to Unsorted — no special handling needed.

## Capabilities

### New Capabilities

- `extension-default-folder`: Settings UI and storage for the default save-destination folder, plus wiring of `folder_id` through all extension save paths.

### Modified Capabilities

- `extension-save-image`: Save flow now accepts and forwards an optional `folder_id` to the API.
- `extension-image-picker`: Batch-picker saves now read and forward the default folder ID.

## Impact

- **`extensions/src/lib/storage.ts`** — new key + getter/setter for default folder ID.
- **`extensions/src/lib/api.ts`** — new `getFolders()` function calling `GET /folders`.
- **`extensions/src/background/index.ts`** — `saveImage`, `persistImage`, `handleSave`, `handlePickerSaveMessage` updated to accept and forward `folderId`.
- **`extensions/src/popup/Settings.tsx`** — new "Save destination" section with folder dropdown.
- **Backend** — no changes; `POST /images` already accepts `folder_id` and gracefully falls back on invalid values.
