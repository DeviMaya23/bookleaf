## Why

The current sidebar only shows a flat folder list with a single "Unsorted" system entry, making it impossible to navigate to "All images" or "Trash", and offering no visual hierarchy for nested folders. This update brings the sidebar in line with the new design and unlocks folder nesting and trash management that the backend already supports.

## What Changes

- Add three pinned system entries to the sidebar: **All** (all images), **Unsorted** (unfiled images), and **Trash** (soft-deleted images, de-emphasized)
- Replace the flat folder list with a recursive tree that respects `parent_id`, with per-node expand/collapse
- Add **New subfolder** to the folder context menu, passing `parent_id` to `POST /folders`
- Add a **Trash view** with images sorted oldest-deleted-first; context menu shows **Restore** instead of Delete
- Add `/unsorted` and `/trash` routes; `/` becomes the All-images view
- Change `ListTrashed` pagination to sort by `deleted_at ASC` (requires cursor key change on backend)

## Capabilities

### New Capabilities
- `fe-sidebar-nav`: Left sidebar navigation — system folder rows (All, Unsorted, Trash), nested folder tree with expand/collapse, New subfolder action
- `fe-trash-view`: Frontend trash view — lists soft-deleted images via `GET /images/trash`, sorted oldest-deleted-first, with per-image Restore action

### Modified Capabilities
- `app-shell`: Sidebar structure requirements change — Unsorted is no longer the sole system entry; All, Unsorted, and Trash replace it; folder list renders as a tree
- `folder-management`: New subfolder creation requirement — `POST /folders` now accepts `parent_id`; context menu gains "New subfolder" item
- `image-list-pagination`: Trash cursor key changes — `ListTrashed` sorts by `deleted_at ASC` so the cursor must carry `deleted_at` instead of `created_at`

## Impact

- **Frontend routes**: `App.tsx` gains `/unsorted` and `/trash`; `/` semantics change from "unfiled" to "all"
- **Frontend components**: `FolderSidebar.tsx` (rewrite), `AppLayout.tsx` (view discriminator), `ImageGrid.tsx` (trash mode)
- **Frontend libs**: `images.ts` (new `getTrashedImages`, `restoreImage`; updated `getImages`), `folders.ts` (optional `parent_id` on `createFolder`)
- **Backend**: `image_pagination.go`, `image_repository.go`, `image_usecase.go` — cursor struct and `ListTrashed` query change
- **No API contract changes** for `GET /images` or `GET /folders` — existing endpoints are unchanged; only sort order and cursor encoding differ for the trash endpoint
