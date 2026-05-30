## 1. Migration

- [ ] 1.1 Write migration `000010_create_image_folders.up.sql`: CREATE TABLE image_folders with both FKs as ON DELETE CASCADE, three indexes, backfill from images.folder_id with ROW_NUMBER() positions, then DROP COLUMN folder_id from images
- [ ] 1.2 Write migration `000010_create_image_folders.down.sql`: add folder_id back to images, backfill from image_folders (lowest position per image), drop image_folders table

## 2. Domain Model

- [ ] 2.1 Add `ImageFolder` struct to `internal/domain/image.go` with `ImageID`, `FolderID`, `Position` fields and GORM tags
- [ ] 2.2 Remove `FolderID *uuid.UUID` field and `Folder *Folder` association from `Image` struct; add `ImageFolders []ImageFolder` with `gorm:"foreignKey:ImageID"` tag

## 3. Repository Interface

- [ ] 3.1 Add `SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [ ] 3.2 Update `List` signature comment in interface: document that `unfiled` now filters via LEFT JOIN on `image_folders` and that results include `ImageFolders` preloaded
- [ ] 3.3 Update `CountByFolderID` comment in interface: document that it queries via JOIN on `image_folders` using `Model(&domain.Image{})` as base

## 4. Image Repository Implementation

- [ ] 4.1 Implement `SetImageFolder` in `internal/repository/image_repository.go`: delete row when `folderID == nil`; INSERT with computed position when non-nil; upsert on conflict
- [ ] 4.2 Update `List`: replace `WHERE images.folder_id = ?` with `JOIN image_folders ON image_folders.image_id = images.id AND image_folders.folder_id = ?`; replace unfiled `WHERE images.folder_id IS NULL` with `LEFT JOIN image_folders ... WHERE image_folders.image_id IS NULL`; add `Preload("ImageFolders")`
- [ ] 4.3 Update `GetByID`: add `Preload("ImageFolders")`
- [ ] 4.4 Update `GetDeletedByID`: add `Preload("ImageFolders")`
- [ ] 4.5 Update `CountByFolderID`: replace direct `WHERE folder_id = ?` on images with `Model(&domain.Image{}).Joins("JOIN image_folders ON image_folders.image_id = images.id").Where("image_folders.folder_id = ?", folderID).Count(&count)`
- [ ] 4.6 Remove any `folder_id` key handling from `Update` (it should never appear in the fields map after this change; no code change needed if it was never explicitly validated, but verify callers are clean)

## 5. Folder Repository Implementation

- [ ] 5.1 Update `DeleteWithCascade` in `internal/repository/folder_repository.go`: remove the `UPDATE images SET folder_id = NULL` step from the transaction (cascade handles it); keep the child folder `parent_id` nulling step
- [ ] 5.2 Update `CountImagesByFolder`: replace `Table("images").Where("folder_id = ? AND user_id = ?", ...)` with `Model(&domain.Image{}).Joins("JOIN image_folders ON image_folders.image_id = images.id").Where("image_folders.folder_id = ? AND images.user_id = ?", id, userID).Count(&count)`

## 6. Usecase Layer

- [ ] 6.1 Update `InitiateUpload` in `image_usecase.go`: remove `FolderID: folderID` from the `domain.Image{}` struct literal; after `imageRepo.Create`, call `imageRepo.SetImageFolder(ctx, created.ID, folderID)` if `folderID != nil`
- [ ] 6.2 Update `AcceptSuggestion`: replace `imageRepo.Update(ctx, imageID, userID, map[string]any{"folder_id": folder.ID})` with `imageRepo.SetImageFolder(ctx, imageID, &folder.ID)`
- [ ] 6.3 Update `UpdateImage`: remove `fields["folder_id"] = *params.FolderID` from the scalar fields map; after `imageRepo.Update`, add `if params.FolderID != nil { imageRepo.SetImageFolder(ctx, id, *params.FolderID) }`
- [ ] 6.4 Update folder-changed logging in `UpdateImage`: derive old folder from `existing.ImageFolders` (first entry if non-empty) instead of `existing.FolderID`

## 7. Handler Layer

- [ ] 7.1 Update `toImageResponse` in `internal/handler/image.go`: replace `FolderID: item.Image.FolderID` with `FolderID: firstFolderID(item.Image.ImageFolders)` using a local helper that returns `&imageFolders[0].FolderID` if non-empty, else `nil`

## 8. Unit Tests — Usecase

- [ ] 8.1 Update `image_usecase_test.go`: replace any mock expectations on `imageRepo.Update` with `folder_id` in the map with expectations on `imageRepo.SetImageFolder` for `InitiateUpload`, `AcceptSuggestion`, and `UpdateImage`
- [ ] 8.2 Add mock for `SetImageFolder` to the mock `ImageRepository` used in usecase tests (success and failure scenarios)
- [ ] 8.3 Verify each affected usecase method (InitiateUpload, AcceptSuggestion, UpdateImage) has one success and one failure scenario for the folder assignment path

## 9. Unit Tests — Handler

- [ ] 9.1 Update `image_test.go`: update any assertions that check `folder_id` in the response to set up `ImageFolders` on the mock image struct rather than `FolderID`
- [ ] 9.2 Add scenario: `toImageResponse` returns null `folder_id` when `ImageFolders` is empty
- [ ] 9.3 Add scenario: `toImageResponse` returns correct `folder_id` when `ImageFolders` has one entry

## 10. Integration Tests — Image Repository

- [ ] 10.1 Update `image_repository_integration_test.go`: update any assertions on `image.FolderID` to check `image.ImageFolders` instead
- [ ] 10.2 Add integration test for `SetImageFolder`: success (assign folder, verify row in image_folders), failure (invalid image_id)
- [ ] 10.3 Add integration test for `List` with folder filter using the new join path
- [ ] 10.4 Add integration test for `List` with `unfiled = true` using the new LEFT JOIN path
- [ ] 10.5 Add integration test for `CountByFolderID` verifying soft-deleted images are excluded

## 11. Integration Tests — Folder Repository

- [ ] 11.1 Update `folder_repository_integration_test.go`: update `DeleteWithCascade` test to assert that `image_folders` rows are removed (rather than asserting `images.folder_id` is NULL)
- [ ] 11.2 Update `CountImagesByFolder` test to use `image_folders` rows for setup and verify soft-deleted images are excluded

## 12. Bruno

- [ ] 12.1 Verify `bruno/images/update-image.bru` request shape is unchanged (folder_id still sent as uuid or null — no edits needed; confirm only)
