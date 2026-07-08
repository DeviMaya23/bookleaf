## Context

The extension currently has four save paths — right-click context menu, drag-to-save, snip/video-frame capture, and the batch image picker — all converging on `saveImage()` in `background/index.ts`, which posts to `POST /images` without a `folder_id`. The backend already accepts and validates `folder_id`; invalid or missing values silently fall back to Unsorted.

The Settings panel (`popup/Settings.tsx`) currently manages two persisted preferences (dark mode, drag-to-save) via `browser.storage.local` keys, following a consistent pattern of `get<Pref>` / `set<Pref>` helpers in `storage.ts`.

## Goals / Non-Goals

**Goals:**
- User can select a default save-destination folder from within the extension Settings panel.
- Selection persists across sessions via `browser.storage.local`.
- All four save paths forward the stored folder ID to the API.
- Folder list is fresh on every Settings open (no stale cache).
- Stale or deleted folder IDs are handled entirely by the backend (no extension-side validation).

**Non-Goals:**
- Per-save folder selection (no modal before each save — that is Feature 2, out of scope).
- Folder creation from within the extension.
- Syncing the folder list in the background or on popup open (only on Settings mount).

## Decisions

**D1: Store folder ID only, not the folder name**

Only the `folder_id` UUID is persisted in storage. The folder name is resolved at display time by matching against the freshly fetched list when Settings mounts. This avoids a stale name being shown if the user renames a folder between sessions.

*Alternative considered:* Store `{ id, name }` pair so Settings can display the name without a network call. Rejected — the extra fetch on Settings mount is inexpensive and eliminates any staleness entirely.

**D2: Fetch folders on Settings mount, not on popup open**

`getFolders()` is called inside a `useEffect` in `Settings.tsx` when the component mounts. It is not called on the main popup view.

*Alternative considered:* Fetch on popup open and pass folders down as a prop. Rejected — most users never open Settings; fetching eagerly wastes a network call on every popup open.

**D3: All four save paths read the default folder ID at call time**

`persistImage()` (used by single saves) and `handlePickerSaveMessage()` (used by batch) each call `getDefaultFolderID()` directly before constructing the upload request. No folder ID is threaded through the existing call signatures any higher than necessary.

`handleSave()` accepts an optional `folderId` param so that future callers (e.g. an annotate-before-save flow) can override the default without re-reading storage.

*Alternative considered:* Read folder ID once at extension startup and cache in a module-level variable. Rejected — the user could change the setting while a save is in flight from a background tab; reading at call time is always correct.

**D4: `getFolders()` lives in `api.ts`**

Consistent with `apiFetch` already living there. No new module needed.

## Risks / Trade-offs

**Stale folder list in Settings dropdown** — The list is fetched once on Settings mount. If the user creates a folder in the web app and then immediately opens Settings without closing it, the new folder won't appear. → Acceptable; closing and reopening Settings refreshes the list.

**Race: save fires before Settings change persists** — `setDefaultFolderID()` is async (`browser.storage.local.set`). A save triggered immediately after changing the dropdown could race. → Negligible in practice; the storage write completes in under a millisecond and saves are user-initiated gestures that follow the settings interaction.

**`GET /folders` may fail** — If the API is unreachable when Settings mounts, the dropdown will show only "None (Unsorted)". The stored folder ID is preserved and will be used on the next save attempt. → Acceptable degradation; user sees a partially-loaded UI but no data is lost.
