## Why

The current `images.folder_id` column allows an image to belong to at most one folder, which blocks future work on manual ordering and multi-folder support. Extracting folder membership into a join table (`image_folders`) decouples the relationship and adds a `position` field needed as a prerequisite for drag-and-drop manual ordering.

## What Changes

- New `image_folders` join table with `(image_id, folder_id, position TEXT)` and a composite primary key; both FKs use `ON DELETE CASCADE`
- `folder_id` column dropped from `images` table; existing data backfilled into `image_folders` with initial positions derived from `created_at` order within each folder
- New `ImageFolder` domain struct; `Image` struct loses `FolderID` and `Folder` fields, gains `ImageFolders []ImageFolder`
- New `ImageRepository.SetImageFolder` method replaces folder assignment in `Update` calls across `InitiateUpload`, `AcceptSuggestion`, and `UpdateImage`
- All folder-filtered image queries rewritten to `JOIN image_folders` starting from `Model(&domain.Image{})` so GORM's soft-delete scope is applied automatically
- `DELETE /folders/:id` cascade behaviour simplified: explicit `UPDATE images SET folder_id = NULL` step removed; FK cascade handles cleanup
- API response shape unchanged — `folder_id` in responses still returns a single UUID (first entry from `ImageFolders` slice)

## Capabilities

### New Capabilities

- `image-folders`: The `image_folders` join table, its domain struct, migration, indexes, and `SetImageFolder` repository method

### Modified Capabilities

- `image-domain`: `Image` struct loses `FolderID`/`Folder` fields, gains `ImageFolders []ImageFolder`; `ImageFolder` struct introduced
- `image-endpoints`: `List`, `GetByID`, `GetDeletedByID` preload `ImageFolders`; `CountByFolderID` queries via join; `SetImageFolder` called by `InitiateUpload`, `AcceptSuggestion`, `UpdateImage`; `toImageResponse` reads folder from `ImageFolders[0]`
- `folder-endpoints`: `CountImagesByFolder` and `DeleteWithCascade` in folder repository updated to work without `images.folder_id`

## Impact

- **Migration**: new file `000010` — creates `image_folders`, backfills, drops `images.folder_id`
- **Backend files**: `domain/image.go`, `repository/image_repository.go`, `repository/folder_repository.go`, `usecase/image_repository.go`, `usecase/image_usecase.go`, `handler/image.go`
- **Tests**: unit tests for usecase and handler; integration tests for both image and folder repositories
- **API contract**: no breaking changes — `folder_id` field remains in all response shapes
- **Bruno**: no request shape changes; no updates needed
