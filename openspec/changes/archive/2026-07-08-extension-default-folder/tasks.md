## 1. Storage

- [x] 1.1 Add `DEFAULT_FOLDER_KEY` constant and `getDefaultFolder`, `setDefaultFolder`, `clearDefaultFolder` helpers to `extensions/src/lib/storage.ts`

## 2. API Client

- [x] 2.1 Add `getFolders()` function to `extensions/src/lib/api.ts` — calls `GET /folders` via `apiFetch`, returns `Array<{ id: string; name: string }>`

## 3. Background Save Wiring

- [x] 3.1 Add optional `folderId?: string` param to `saveImage()` in `background/index.ts` and include it as `folder_id` in the `POST /images` body when present
- [x] 3.2 Update `persistImage()` to call `getDefaultFolder()` and pass its `id` to `saveImage()`; update the success toast body to say "Added to [name]." when a folder is set, "Added to Unsorted." when not
- [x] 3.3 Update `handleSave()` to accept optional `folderId?: string` and forward it to `persistImage()`
- [x] 3.4 Update `handlePickerSaveMessage()` to call `getDefaultFolder()` once before fan-out and pass the `id` into each `handleSave` call; update the aggregated success toast body to say "X images added to [name]." or "X images added to Unsorted."

## 4. Settings UI

- [x] 4.1 Add a "Save destination" section to `Settings.tsx`: section label styled like "Hotkeys", a row with label "Default folder" and a native `<select>`
- [x] 4.2 On Settings mount, call `getFolders()` and populate the `<select>` with a leading "None (Unsorted)" option followed by fetched folders; disable the select while loading
- [x] 4.3 Initialise the `<select>` value from `getDefaultFolder()` on mount (pre-select stored folder or empty string)
- [x] 4.4 On `<select>` change, call `setDefaultFolder({ id, name })` for a non-empty selection or `clearDefaultFolder()` for empty; handle fetch errors by leaving the select at "None (Unsorted)" without clearing the stored preference

## 5. Quality

- [x] 5.1 Run `npm run build` in `extensions/` and fix any type errors
- [x] 5.2 Run `npm run lint` in `extensions/` and fix any lint issues
