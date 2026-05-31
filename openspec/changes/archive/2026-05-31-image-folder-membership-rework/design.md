## Context

The `PATCH /images/:id` endpoint currently accepts `folder_id: uuid|null` — a scalar setter inherited from when folder was a column on `images`. The join table migration (`000010`) made folder a relationship with its own attribute (`position`), but `UpdateImage` was never updated to match. The result: DnD moves call `SetImageFolder(add)` without removing the source folder, leaving images in multiple folders unintentionally.

Two distinct operations are needed:
- **Folder list sync** (side panel UX, future): declare the complete desired set of folder memberships; BE diffs and reconciles, preserving positions of unchanged rows
- **Folder move** (DnD): directed action with an explicit source and destination; only those two rows are touched; no other memberships are affected

## Goals / Non-Goals

**Goals:**
- Fix the DnD bug: dragging A→B removes from A and adds to B, leaving other memberships untouched
- Replace the scalar `folder_id` setter in `PATCH /images/:id` with a complete-list diff (following the tags pattern)
- Add `POST /images/:id/move-folder` as a dedicated atomic move endpoint
- Preserve `SetImageFolder` for the creation path — it is correct there and unchanged

**Non-Goals:**
- Frontend side panel folder management UI (not yet implemented)
- Reordering within a folder (handled separately via `UpdateImagePosition`)
- Changes to how `InitiateUpload` or `CompleteUpload` assign folders

## Decisions

### 1. Folder list sync belongs in the repository layer

The diff (current memberships vs new list) runs inside a single repository method `SyncImageFolders`, wrapped in a transaction.

**Why not the usecase layer:** The usecase would need to fetch current memberships, compute the diff, then call separate remove and add primitives — three operations with no atomicity guarantee between them. If the process crashes mid-diff, the image ends up in a partial state. Keeping it in the repository allows a single transaction.

**Why not replace-all (delete all then insert):** Positions of unchanged memberships must be preserved. Replace-all destroys those positions. The diff explicitly identifies which rows to delete and which to insert, leaving the rest untouched.

### 2. MoveImageFolder as a single atomic repository method

A new `MoveImageFolder(ctx, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error` method wraps both operations in a transaction:
- If `fromFolderID` non-nil: `DELETE WHERE image_id = ? AND folder_id = ?`
- If `toFolderID` non-nil: fracdex-append insert into destination folder

Handles all combinations (move, unfile, file-without-source, no-op). The usecase and handler don't need to know which combination is in play.

**Alternative considered — two separate API calls from the FE (remove then add):** Rejected. Non-atomic; if the first call succeeds and the second fails, the image is left in no folder. A single endpoint that expresses the full intent is safer.

### 3. `UpdateImageParams.FolderID` → `FolderIDs`

`FolderID **uuid.UUID` is replaced by `FolderIDs *[]uuid.UUID`:
- `nil` outer pointer → absent, no change (matches tags pattern)
- `&[]uuid.UUID{}` → empty list, remove all folder memberships
- `&[]uuid.UUID{id1, id2}` → sync to exactly this set

The handler decodes `folder_ids` via `json.RawMessage` using the same absent/null/array logic already in place for `tags`.

### 4. `SetImageFolder` preserved, not replaced

`SetImageFolder` remains the write path for creation (`CompleteUpload`, AI folder suggestion). It was designed for that use case and is correct there. Only the `UpdateImage` path changes.

## Risks / Trade-offs

- **Position for newly added folders appended to end** — when `SyncImageFolders` adds a folder, the image gets the last position in that folder (fracdex append). This is the only sensible default; explicit reordering is a separate concern.
- **`folder_id` field removed from `PATCH /images/:id`** — any existing FE code sending `folder_id` will silently have it ignored (field is simply not decoded). The FE DnD handler is updated in this same change, so there's no window where the old FE hits the new BE. Still, it's a breaking change on the API surface.
- **`SyncImageFolders` fetches current memberships inside the transaction** — adds one SELECT per `PATCH` that includes `folder_ids`. Acceptable; this is a user-initiated edit action.

## Migration Plan

No DB schema changes. All changes are at the application layer.

1. Deploy BE with new `MoveImageFolder` endpoint and updated `PATCH /images/:id` behaviour
2. Deploy FE with updated `onDragEnd` pointing to new endpoint
3. No rollback complexity — old `folder_id` field is simply unrecognised by new BE; if FE is rolled back, DnD reverts to the broken-but-not-crashing previous behaviour
