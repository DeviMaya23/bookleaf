## Why

The frontend is adding a multi-select mode to the gallery (separate, future proposal) whose two core bulk actions are "add selection to a folder" and "move selection to trash." Doing this with today's per-image endpoints would force the frontend to either loop N sequential `POST /images/:id/move-folder` / `DELETE /images/:id` calls, or fetch each image's current `folder_ids` first to do a full-replace via `PATCH /images/:id` (`SyncImageFolders`). Both push operation-specific looping and round-trip logic onto the client. Two bulk endpoints let the client send one request per bulk action and let the backend own the per-item iteration, validation, and partial-failure handling.

## What Changes

- Add `POST /images/bulk/add-to-folder` — accepts a list of image IDs and one folder ID; adds each image to the folder (insert into `image_folders`, no-op if the membership row already exists — same idempotent semantics as the existing single-image move-folder path).
- Add `POST /images/bulk/trash` — accepts a list of image IDs; soft-deletes each one (reuses existing single-image soft-delete logic, looped per ID).
- Both endpoints process each image ID independently: a failure on one image (not found, not owned by the caller, or — for trash — already trashed) is logged server-side and skipped; it does not fail the whole request. `folder_id` itself (for add-to-folder) is validated up front and a missing/unowned folder fails the entire request with 404.
- Both endpoints return `200 OK` with `{ "succeeded_count": <n> }`. No per-item failure detail is returned to the caller.
- `image_ids` are validated as well-formed UUIDs; a malformed ID fails the request with 400. No cap on batch size.
- This is a backend-only change. No frontend code consumes these endpoints yet — that lands in a follow-up proposal for the gallery multi-select UI.

## Capabilities

### New Capabilities
- `image-bulk-folder-add`: Defines the `POST /images/bulk/add-to-folder` endpoint and its usecase/repository support for adding many images to one folder in a single request.
- `image-bulk-trash`: Defines the `POST /images/bulk/trash` endpoint and its usecase support for soft-deleting many images in a single request.

### Modified Capabilities
(none — existing single-image endpoints and their requirements are unchanged)

## Impact

- **Routes**: Two new routes on the existing `protected` group in `backend/cmd/server/main.go`, alongside the current `/images/:id/...` routes: `POST /images/bulk/add-to-folder`, `POST /images/bulk/trash`.
- **Usecase layer**: New methods on `ImageUsecase` (bulk add-to-folder) and `TrashUsecase` (bulk trash), each looping per-image-ID calls to existing or near-existing repository primitives.
- **Repository layer**: Likely one new repository method for idempotent single-row insert into `image_folders` (distinct from `SyncImageFolders`'s full-replace), used per image ID in the bulk add-to-folder loop. Trash reuses the existing soft-delete repository method.
- **No DB schema changes** — uses the existing `image_folders` table and `images.deleted_at` column.
- **No frontend changes** in this proposal.
