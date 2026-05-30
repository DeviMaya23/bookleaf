## Context

Images currently carry a `folder_id UUID` column — a direct nullable FK to `folders`. This works for single-folder membership but cannot support multiple folders per image and provides no ordering hook. The `image_folders` join table replaces this column, carrying an additional `position TEXT` field for future fractional-index-based manual ordering. All existing data is preserved via a backfill migration.

## Goals / Non-Goals

**Goals:**
- Replace `images.folder_id` with an `image_folders` join table
- Preserve all existing folder membership data via migration backfill
- Keep the public API contract unchanged (`folder_id` still present in all responses)
- Ensure GORM soft-delete scope applies automatically to all folder-filtered image queries
- Lay the groundwork for manual ordering (position field exists; ordering by it is a future feature)

**Non-Goals:**
- Exposing multiple folder membership in the API (future feature)
- Implementing manual drag-and-drop ordering (future feature)
- Changing any FE-facing field names or response shapes

## Decisions

### Join table FK cascade strategy

Both FKs on `image_folders` use `ON DELETE CASCADE`:
- `image_folders.image_id` → when an image is hard-deleted, its join rows are cleaned up automatically
- `image_folders.folder_id` → when a folder is deleted, its join rows are cleaned up automatically

**Why CASCADE over RESTRICT + explicit nulling**: The current `DeleteWithCascade` in the folder repository explicitly sets `images.folder_id = NULL` before deleting the folder because the FK was `ON DELETE SET NULL`. With the join table, "unfile" semantics are expressed as row deletion rather than NULL assignment, which CASCADE handles. This removes a step from `DeleteWithCascade` and eliminates the need to coordinate the image update inside the transaction.

Soft-delete does not touch `image_folders` — only hard-delete triggers the cascade. This is intentional: restoring a soft-deleted image retains its folder membership. If the folder was hard-deleted while the image was in trash, the cascade already cleaned up the join row, so the restored image lands as unfiled.

### Query base: always start from `Model(&domain.Image{})`

All folder-filtered image queries (list by folder, count by folder, unfiled filter) use `Model(&domain.Image{})` as the GORM base and JOIN into `image_folders`, never the reverse.

**Why**: GORM's soft-delete scope (`deleted_at IS NULL`) is model-bound. Starting from `image_folders` directly loses this automatic filter and requires manual `deleted_at IS NULL` assertions — a maintenance hazard that will silently include trashed images if forgotten on a new query.

### `SetImageFolder` as the single write path for folder assignment

A new `ImageRepository.SetImageFolder(ctx, imageID UUID, folderID *UUID) error` method handles all folder assignment:
- `folderID == nil` → `DELETE FROM image_folders WHERE image_id = ?` (unfile)
- `folderID != nil` → `INSERT INTO image_folders ... ON CONFLICT (image_id, folder_id) DO UPDATE SET position = EXCLUDED.position`

The scalar `Update` method's `fields map[string]any` no longer accepts `folder_id` as a key. Call sites (`InitiateUpload`, `AcceptSuggestion`, `UpdateImage`) are updated to call `SetImageFolder` separately after the scalar update.

**Why separate method**: Mixing folder assignment into the generic `Update` map is what introduced the current tight coupling. A named method makes intent explicit, is independently testable, and is the correct abstraction boundary for what will become a richer operation when manual ordering is implemented.

### Position initial value on backfill

Initial positions are generated as `ROW_NUMBER() OVER (PARTITION BY folder_id ORDER BY created_at ASC)::TEXT`. This produces integer strings (`"1"`, `"2"`, ...) which sort correctly lexicographically within small sets. When manual ordering is implemented, positions will be rewritten using a proper fractional indexing scheme (e.g., zero-padded or LexoRank-style strings). The backfill value is a placeholder; the field's purpose is to have a stable column in place for the future ordering feature.

### API response: single `folder_id` from first `ImageFolders` entry

`toImageResponse` reads `image.ImageFolders[0].FolderID` if the slice is non-empty, else returns nil. The response field name and type (`folder_id *uuid.UUID`) are unchanged.

**Why not `folder_ids []uuid.UUID`**: No FE feature currently uses multiple folders. A breaking change here adds FE work for zero user-visible benefit. The multi-folder API shape will be introduced alongside the feature that needs it.

## Risks / Trade-offs

- **Position string sort correctness** → simple integer strings sort lexicographically correctly only up to 9 items (`"10"` < `"2"` lexicographically). Mitigation: acceptable for a placeholder; when manual ordering is shipped, positions will be rewritten with a proper scheme.
- **`SetImageFolder` position on new inserts** → for now, new folder assignments append to the end using `(SELECT COALESCE(MAX(position::int), 0) + 1 FROM image_folders WHERE folder_id = ?)::TEXT`. This is a placeholder and breaks under concurrent inserts, but is safe for single-user sequential operations at current scale.
- **Preload adds a query per list call** → `Preload("ImageFolders")` adds one extra IN-query per list/get call. Acceptable; GORM batches preloads so it is not N+1.

## Migration Plan

Migration `000010` runs three steps in order:

1. `CREATE TABLE image_folders` with both FKs as `ON DELETE CASCADE` and the three indexes
2. `INSERT INTO image_folders` backfill from `images WHERE folder_id IS NOT NULL` using `ROW_NUMBER()` for position
3. `ALTER TABLE images DROP COLUMN folder_id`

Down migration reverses:
1. `ALTER TABLE images ADD COLUMN folder_id UUID REFERENCES folders(id) ON DELETE SET NULL`
2. Backfill `images.folder_id` from `image_folders` (one row per image, lowest position wins)
3. `DROP TABLE image_folders`

No application downtime is required beyond the migration run itself. The migration is not zero-downtime (the column drop is destructive); deploy with the app stopped or behind a maintenance window.
