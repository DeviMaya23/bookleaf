## Why

The backend endpoints for permanent deletion (`DELETE /images/trash/:id` and `DELETE /images/trash`) are already implemented but not wired to the frontend. Users currently have no way to permanently delete images or empty their trash from the UI.

## What Changes

- Add "Delete permanently" option to the image card context menu in the trash view, below a separator after "Restore", coloured red — opens a single-image confirmation dialog
- Add a context menu to the Trash sidebar entry with an "Empty trash" option, coloured red — opens a bulk confirmation dialog
- Add `hardDeleteImage` and `emptyTrash` API functions to `lib/images.ts`

## Capabilities

### New Capabilities

- `fe-trash-permanent-delete`: Permanent deletion of individual trashed images and bulk emptying of trash from the frontend

### Modified Capabilities

- `fe-trash-view`: Adds permanent delete entry to the image card context menu in trash view

## Impact

- `frontend/src/lib/images.ts` — two new API functions
- `frontend/src/components/ImageGrid.tsx` — updated trash context menu, new confirmation dialog, new mutation
- `frontend/src/components/FolderSidebar.tsx` — Trash entry wrapped in `ContextMenu`, new confirmation dialog, new mutation
