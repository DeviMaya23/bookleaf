## Why

The selection panel supports bulk folder assignment but has no equivalent for tags, forcing users to tag images one at a time via the individual image panel. Separately, `BulkTrash` loops through N individual `SoftDelete` calls instead of issuing a single atomic DB update.

## What Changes

- **New**: `POST /images/bulk/tag` endpoint — appends a set of tags to many images in one atomic transaction (`ON CONFLICT DO NOTHING`)
- **New**: `AppendImageTagsBulk` repository method and `BulkAddTags` usecase method backing the endpoint above
- **Improved**: `BulkTrash` usecase replaces its per-image `SoftDelete` loop with a single `BulkSoftDelete(ids, userID)` repository call (`UPDATE images SET deleted_at = NOW() WHERE id IN (?) AND user_id = ?`)
- **Extended**: Selection panel adds a "Add tags" section below "Add to folder", reusing the existing `TagInput` chip component. A single Apply button fires both the folder and tag bulk requests concurrently. The `onAddToFolder`-per-folder prop is replaced by a unified `onApply({ folderIds, tagIds })` prop.

## Capabilities

### New Capabilities

- `image-bulk-tag`: `POST /images/bulk/tag` — appends tag IDs to many images atomically; idempotent via `ON CONFLICT DO NOTHING`

### Modified Capabilities

- `fe-selection-panel`: Adds a tag chip picker section and unifies Apply to cover both folders and tags in one action
- `image-bulk-trash`: Internal improvement — replace per-image loop with a single atomic `UPDATE` via a new `BulkSoftDelete` repo method; spec-level semantics (skip unowned, return succeeded_count) are unchanged

## Impact

- **Backend**
  - `backend/internal/repository/image_repository.go`: new `BulkSoftDelete` method
  - `backend/internal/usecase/trash_usecase.go`: `BulkTrash` updated to call `BulkSoftDelete`
  - `backend/internal/repository/tag_repository.go`: new `AppendImageTagsBulk` method
  - `backend/internal/usecase/image_usecase.go` (or `tag_usecase.go`): new `BulkAddTags` usecase method
  - `backend/internal/handler/image.go`: new `BulkAddTags` handler + route registered in `main.go`
  - `backend/internal/usecase/tag_repository.go` (interface): adds `AppendImageTagsBulk`
- **Frontend**
  - `frontend/src/features/right-panel/components/SelectionPanelBody.tsx`: adds tag section, unified Apply
  - `frontend/src/features/right-panel/components/RightPanel.tsx`: prop shape update
  - `frontend/src/app-shell/AppLayout.tsx`: wires `onApply`, calls new `bulkAddTags` lib function
  - `frontend/src/lib/images.ts` (or `lib/tags.ts`): new `bulkAddTags` function for `POST /images/bulk/tag`
- **Bruno**: new `POST /images/bulk/tag` request file
