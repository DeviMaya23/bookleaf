## Why

`PUT /folders/:id` is implemented as a blind full-replace: it fetches the existing row and unconditionally overwrites `name`, `parent_id`, and `description` with whatever the request body contains, so any field the client omits is written as `NULL`/empty. Every real caller (`renameFolder`, `moveFolder`, `updateFolderDetails`) sends a partial body, which means renaming a subfolder silently un-parents it and moving a folder silently clears its description — a live data-integrity bug, not just a UX gap. `PATCH /images/:id` already implements correct partial-update (merge) semantics with a presence-aware contract; folders are the one resource that doesn't follow that house pattern.

## What Changes

- **BREAKING**: `PUT /folders/:id` → `PATCH /folders/:id`, with true partial-merge semantics — only fields present in the request body are modified; omitted fields are left untouched
- Request DTO becomes presence-aware (`json.RawMessage` per optional field, mirroring `updateImageRequest`) so the handler can distinguish "field omitted" from "field explicitly set to null" from "field set to a value"
- Usecase gains optional/nullable params (mirroring `UpdateImageParams`'s `*T`/`**T` convention) with merge-aware validation: `name`, if provided, cannot be blank (matches `title` validation on images); `parent_id` and `description` apply only when present, including explicit-null
- Repository replaces the blanket `Select("name","parent_id","description").Updates(existing)` full overwrite with a selective `map[string]any` update, mirroring `imageRepo.Update`
- **BREAKING**: FE consolidates `renameFolder`, `moveFolder`, and `updateFolderDetails` into a single `updateFolder(getToken, id, { name?, description?, parent_id? })`, mirroring `updateImage`/`UpdateImageParams` — the three-way split existed only as a workaround for the full-replace contract and has no remaining justification once PATCH is real
- All FE call sites (`FolderSidebar.tsx`, `FolderPanelContent.tsx`, `dragHandlers.ts`) updated to the consolidated wrapper and minimal partial-body shapes (e.g. drag-to-move sends `{ parent_id }` only, not `{ name, parent_id }`)
- Bruno collection updated for the new verb and partial-update contract

## Capabilities

### New Capabilities
(none — this redesigns an existing endpoint's contract rather than introducing new capability surface)

### Modified Capabilities
- `folder-endpoints`: rewrite the `PUT /folders/:id — Update Folder` requirement as `PATCH /folders/:id` with partial-merge semantics (resolves the original spec's contradiction of "name required" vs "parent_id optional" under a PUT verb); update the Folder Repository Interface requirement to describe selective-column updates instead of full-row replace
- `fe-folder-panel`: auto-save scenarios for the title/description fields now describe `PATCH /folders/:id` calls through the consolidated `updateFolder` wrapper with single-field partial bodies, instead of `PUT /folders/:id`
- `fe-drag-drop-folder-nesting`: drag-to-move and drag-to-root scenarios now describe `PATCH /folders/:id` called with `{ parent_id: targetFolderId }` / `{ parent_id: null }` only — no longer re-sending `name`
- `folder-management`: rename scenarios now describe `PATCH /folders/:id` called with `{ name }` only

## Impact

- **Backend**: `internal/handler/folder.go` (request DTO + handler), `internal/usecase/folder_usecase.go` (Update signature + merge logic), `internal/repository/folder_repository.go` (selective update), route registration in `cmd/server/main.go` (`PUT` → `PATCH`), folder bruno collection, and the handler/usecase/repository test suites for `UpdateFolder`
- **Frontend**: `frontend/src/lib/folders.ts` (wrapper consolidation), `frontend/src/components/FolderSidebar.tsx`, `frontend/src/components/FolderPanelContent.tsx`, `frontend/src/lib/dragHandlers.ts`, and their associated tests
- **Specs**: delta specs for `folder-endpoints`, `fe-folder-panel`, `fe-drag-drop-folder-nesting`, `folder-management`
