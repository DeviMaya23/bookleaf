## Why

The `PATCH /images/:id` endpoint accepts `folder_id` as a scalar field — an artifact from when folder was a column on the `images` table. It was never updated for the join table model, so DnD moves only add images to the target folder without removing them from the source folder. The root cause is using an image-property setter for what is actually a relationship operation.

## What Changes

- **BREAKING** `PATCH /images/:id` — replace the `folder_id: uuid|null` field with `folder_ids: uuid[]` (complete folder membership list). The BE diffs the provided list against current memberships: removes departed rows, inserts new rows with fracdex positions, and leaves unchanged rows untouched (preserving their positions).
- New `POST /images/:id/move-folder` endpoint — accepts `{ from_folder_id: uuid|null, to_folder_id: uuid|null }`. Atomically removes the `(imageID, fromFolderID)` row and inserts `(imageID, toFolderID)`. Only the two named folders are touched; all other memberships are unaffected.
- FE `onDragEnd` handler — updated to call `POST /images/:id/move-folder` instead of `PATCH /images/:id` with `folder_id`. The drag item already carries `currentFolderId` so both operands are available at drop time.
- Drop-to-unsorted (DnD) — updated to call `POST /images/:id/move-folder` with `{ from_folder_id: currentFolderId, to_folder_id: null }`.

## Capabilities

### New Capabilities

- `image-folder-move`: `POST /images/:id/move-folder` endpoint — handler, usecase method, repository method; atomically removes from source folder and adds to destination folder without touching other memberships

### Modified Capabilities

- `image-edit`: replace scalar `folder_id` with `folder_ids: uuid[]` complete list on `PATCH /images/:id`; BE performs a position-preserving diff
- `fe-drag-drop-image-to-folder`: update `onDragEnd` to call the new move-folder endpoint instead of patching `folder_id`

## Impact

- `backend/internal/handler/image.go` — new `MoveImageFolder` handler, update `EditImage` handler
- `backend/internal/usecase/image_usecase.go` — new `MoveImageFolder` usecase method, update `UpdateImage` to diff `folder_ids`
- `backend/internal/usecase/image_repository.go` — new `MoveImageFolder` interface method
- `backend/internal/repository/image_repository.go` — new `MoveImageFolder` repository implementation
- `frontend/src/App.tsx` — update `onDragEnd`
- `frontend/src/lib/images.ts` — new `moveImageFolder` API function
- Bruno collection — new `move-folder` request file
