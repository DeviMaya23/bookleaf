## Why

The `folders` table has a `description` column that's never surfaced or editable in the UI, and there's no way to view or edit a folder's title outside the rename dialog. Users need a dedicated place to view and edit folder metadata, and the existing right panel — already a 320px detail surface for images — is a natural fit to extend rather than building a new UI surface.

## What Changes

- Clicking a folder in the sidebar that differs from the currently active folder opens (or updates) the right panel to show that folder's metadata: title and description, both inline-editable
- Re-clicking the currently active folder is a no-op — the panel's current content (whether folder or image metadata) is left untouched
- Title and description are saved via `PUT /folders/:id` on blur, only when the value has changed
- Title cannot be saved as empty; an emptied title reverts to the previous value without a save call. Description may be saved as empty (cleared to `null`)
- The right panel becomes a shared surface across two trigger contexts — image selection and folder selection — each populating it with different content

## Capabilities

### New Capabilities
- `fe-folder-panel`: Folder detail view rendered in the right panel, showing the folder's title and description as inline-editable fields with blur-to-save persistence and validation (title required, description optional)

### Modified Capabilities
- `fe-right-panel`: Gains a new trigger — selecting a folder (distinct from the active one) opens/updates the panel to display that folder's metadata instead of (or in addition to, depending on prior state) image metadata; selecting the already-active folder leaves the panel's current content untouched

## Impact

- Frontend only — no backend changes required (`PUT /folders/:id` already supports `name` and `description` per `folder-endpoints`)
- Affected components: `RightPanel.tsx` (gains folder-content mode), `AppLayout.tsx` (wires folder selection into panel state), `FolderSidebar.tsx` (folder click now also drives panel state, in addition to existing navigation behavior)
- New component for rendering/editing folder metadata within the panel
