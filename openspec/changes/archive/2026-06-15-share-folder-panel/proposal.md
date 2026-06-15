## Why

The backend already supports creating, fetching, and deleting a public share link for a folder (`POST/GET/DELETE /folders/:id/share`), but there's no way for users to enable sharing or get the link from the UI. Folder owners need a simple way to turn sharing on/off for a folder and copy the resulting link.

## What Changes

- Add a "Share folder" section to `FolderPanelContent`, placed between the Notes section and the Export footer, using the same bordered-row style as Notes.
- On mount, fetch the current share state via `GET /folders/:id/share` — a 404 means sharing is off (not an error).
- A switch toggles sharing on/off:
  - Turning on calls `POST /folders/:id/share` (idempotent — returns the existing token if one already exists) and reveals a read-only, truncated input showing `window.location.origin + /share/:token`, with an adjacent copy-to-clipboard icon button.
  - Turning off opens a confirm dialog warning that the existing link will stop working and that re-enabling will generate a new link. Confirming calls `DELETE /folders/:id/share`.
- New `frontend/src/lib/share.ts` wrapper functions for the three authenticated endpoints (get/create/delete folder share).

**Non-goal**: the public share viewer page (`/share/:token`) is not part of this change — the copied link will not resolve to a page yet.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `fe-folder-panel`: `FolderPanelContent` gains a "Share folder" section with a share-state query, an on/off switch, a revealed share-link field with copy-to-clipboard, and a confirm dialog for disabling.

## Impact

- `frontend/src/features/right-panel/components/FolderPanelContent.tsx`: new Share folder section, state, and handlers.
- `frontend/src/lib/share.ts`: new file with `getFolderShare`, `createFolderShare`, `deleteFolderShare`.
- No backend changes — all three endpoints already exist.
