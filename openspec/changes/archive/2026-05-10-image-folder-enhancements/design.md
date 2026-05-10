## Context

The image and folder domains are established with their GORM structs, migrations, repository/usecase/handler layers, and unit + integration tests all in place. `CompleteUpload` already fetches the full image binary from R2 via `store.GetObject` inside `prepareThumbnail` — this is the natural insertion point for dimension and size extraction. The folder `GetByID` endpoint currently returns only folder metadata with no awareness of its contents.

## Goals / Non-Goals

**Goals:**
- Add `description` to `Image` and `Folder` — nullable, user-supplied, surfaced on create/update/read
- Add `width`, `height`, `file_size` to `Image` — BE-calculated at `CompleteUpload` time using the bytes already fetched from R2; never set by the client
- Return `image_count` on `GET /folders/:id` — count of non-deleted images belonging to that folder

**Non-Goals:**
- Recalculating dimensions or size for previously uploaded images (no backfill)
- Exposing `image_count` on the flat `GET /folders` list (individual folder detail only)
- Client-supplied dimensions or file size

## Decisions

### 1. Dimension and size extraction inside `prepareThumbnail`

`prepareThumbnail` already reads the full image binary from R2 into memory as `[]byte`. We extend it to also decode image dimensions via Go's stdlib `image` package and capture `len(bytes)` for file size. These values are returned alongside the thumbnail bytes, then persisted via a single `imageRepo.Update` call before the goroutine is spawned.

**Alternative considered**: a separate `store.GetObject` call in `CompleteUpload`. Rejected — redundant network round-trip when the bytes are already in memory.

### 2. Persist dimensions via existing `imageRepo.Update`

The `Update(ctx, id, userID, fields map[string]any)` method already accepts arbitrary field maps. No new repository method is needed — just add `"width"`, `"height"`, `"file_size"` to the fields map.

**Alternative considered**: a dedicated `UpdateDimensions` method on `ImageRepository`. Rejected — the generic `Update` exists precisely for this pattern and keeps the interface lean.

### 3. Folder image count via `ImageRepository.CountByFolderID`

A new `CountByFolderID(ctx, folderID uuid.UUID) (int64, error)` method is added to `ImageRepository`. `folderUsecase.GetByID` receives the `imageRepo` as a constructor dependency and calls it after fetching the folder. The result is wrapped in a new `FolderDetail` return type from the usecase.

**Alternative considered**: adding a count method to `FolderRepository` via a JOIN. Rejected — it couples folder data access to the images table; the existing cross-repo pattern (folder usecase already touches `folderRepo`) is cleaner with explicit repo injection.

**Alternative considered**: a subquery in the existing `folderRepo.GetByID` SQL. Rejected — the repository layer should not orchestrate cross-domain aggregation; that belongs in the usecase.

### 4. Two separate migrations

Migration 000006 adds `description`, `width`, `height`, `file_size` to `images`. Migration 000007 adds `description` to `folders`. Keeping them separate makes each rollback atomic and independent.

### 5. `description` accepted on `InitiateUpload` (POST /images), not only on update

Users may know the description at upload time. It is passed through to `imageRepo.Create` alongside title and MIME type. `UpdateImageParams` gains a `Description *string` field for `PATCH /images/:id`.

## Risks / Trade-offs

- **Dimension decode failure** → If the image MIME type is unsupported by Go's `image` package (e.g. WebP without a registered codec), `image.DecodeConfig` returns an error. Mitigation: treat as non-fatal — log and leave `width`/`height` as NULL rather than failing the whole `CompleteUpload`.
- **File size always available** → `len(bytes)` never fails once the bytes are in memory; no special handling needed.
- **`folderUsecase` gains a new dependency** (`ImageRepository`) → constructor signature changes. Mitigation: update `main.go` wiring; the change is mechanical and localised.
- **`image_count` is a point-in-time snapshot** → it reflects the count at request time, not a maintained counter. Acceptable for this use case.

## Migration Plan

1. Add migration 000006 (`add_image_metadata_fields`) — `ALTER TABLE images ADD COLUMN description text, ADD COLUMN width integer, ADD COLUMN height integer, ADD COLUMN file_size bigint`; down reverses with `DROP COLUMN`
2. Add migration 000007 (`add_folder_description`) — `ALTER TABLE folders ADD COLUMN description text`; down reverses with `DROP COLUMN`
3. Migrations run automatically on server boot via `golang-migrate`; no data backfill required since all new columns are nullable
