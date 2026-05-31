## Context

The multi-folder backend is complete. `PATCH /images/:id` accepts `folder_ids[]` and syncs memberships via `SyncImageFolders`. However, `GET /images/:id` still returns a singular `folder_id *uuid.UUID` built by `firstFolderID()` (takes `ImageFolders[0]`). The FE `Image` type mirrors this with `folder_id: string | null`, and the right panel displays the folder name as a static string.

## Goals / Non-Goals

**Goals:**
- Return all folder memberships from `GET /images/:id` (and list endpoints)
- Let users add/remove folders from the right panel
- Pattern mirrors the existing tag editing flow

**Non-Goals:**
- Inline folder creation (folders are structural; creation belongs in the sidebar)
- Changing how folder-filtered gallery views work (`?folder_id=` query param on list)
- Position management via the folder input (positions are managed by drag-and-drop)

## Decisions

### 1. Replace `folder_id` with `folder_ids` in the response

`imageDetailResponse` and `imageResponse` will drop `folder_id *uuid.UUID` and `position *string` and replace them with `folder_ids []uuid.UUID`. The `toImageResponse` helper is updated accordingly; `firstFolderID` and `firstFolderPosition` helpers are removed.

**Why not keep both?** Keeping `folder_id` as a compat alias would perpetuate the misconception that an image has one folder. The FE is the only consumer, and this is a single-team project — a clean break is cheaper than dual fields.

**FE type change**: `folder_id: string | null` becomes `folder_ids: string[]` on `Image`. Any FE code referencing `image.folder_id` (gallery filtering uses `?folder_id=` query param, not the image object field) is audited and updated.

### 2. FolderInput: combobox multi-select, no creation

`FolderInput` is a new component similar in structure to `TagInput` but simpler:
- Renders current folder assignments as removable pill chips
- A text input filters existing folders by name (dropdown)
- Selecting from the dropdown adds the folder; ✕ removes it
- No "commit raw value" path — only pre-existing folders can be selected
- `onChange(folders: { id: string; name: string }[])` callback — component is unaware of the API

**Why no creation?** Accidental folder creation is hard to clean up; folders appear in sidebar nav. The dedicated folder creation flow (sidebar) is the right entry point.

### 3. Right panel wires FolderInput the same way as TagInput

`RightPanel` will:
1. Maintain `folders` local state initialised from `image.folder_ids` (resolved to `{ id, name }[]` via the cached `['folders']` query)
2. Reset on `image.id` change
3. On `FolderInput` `onChange` → call `PATCH /images/:id` with `{ folder_ids: [...] }`
4. Show success/error toast
5. Invalidate `['images']` on success

The details grid row that currently shows the static folder name is removed; the new `FolderInput` section sits between Source URL and Tags, consistent with the Tags section placement pattern.

## Risks / Trade-offs

- **List response shape change**: `imageResponse` (used by `GET /images`) also loses `folder_id` in favour of `folder_ids`. Any code that reads `image.folder_id` from list responses must be updated. Risk is low — it's a FE-only consumer — but requires a complete audit.
- **Extension**: The Chrome extension may also use the `Image` type. It should be audited, though it likely doesn't display folder info.
