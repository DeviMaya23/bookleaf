## Why

The multi-folder backend for images is complete — images can belong to multiple folders and the PATCH endpoint syncs them — but the frontend still treats folder as a single, read-only field in the right panel. Users have no way to assign or remove folders from an image through the UI.

## What Changes

- `GET /images/:id` response is updated to return `folder_ids: []uuid` (array) instead of the singular `folder_id: *uuid`
- The `imageDetailResponse` and `imageResponse` handler structs are updated accordingly
- The `Image` and `ImageDetail` FE types gain a `folder_ids: string[]` field (replacing `folder_id`)
- A new `FolderInput` component is added — a combobox multi-select from existing folders, no inline creation
- `RightPanel` replaces the static folder name display with the editable `FolderInput`
- On folder change, `RightPanel` calls `PATCH /images/:id` with `{ folder_ids: [...] }`, same pattern as tags

## Capabilities

### New Capabilities

- `fe-image-folder-panel`: FolderInput component and right panel folder editing integration

### Modified Capabilities

- `fe-right-panel`: adds editable folder section (replaces static folder name in details grid)
- `image-edit`: GET /images/:id response shape changes — `folder_id` becomes `folder_ids[]`

## Impact

- **Backend**: `imageDetailResponse` and `imageResponse` structs in `handler/image.go`, `toImageResponse` helper, `GetImage` handler
- **Frontend**: `Image` type in `lib/images.ts`, `RightPanel.tsx`, new `FolderInput.tsx` component
- **Existing folder display**: the static "Folder" row in the details grid is replaced by the new editable input
