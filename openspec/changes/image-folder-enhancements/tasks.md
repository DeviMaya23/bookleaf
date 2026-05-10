## 1. Migrations

- [ ] 1.1 Add migration 000006 (`add_image_metadata_fields`) — up: `ALTER TABLE images ADD COLUMN description text, ADD COLUMN width integer, ADD COLUMN height integer, ADD COLUMN file_size bigint`; down: drop those columns
- [ ] 1.2 Add migration 000007 (`add_folder_description`) — up: `ALTER TABLE folders ADD COLUMN description text`; down: `DROP COLUMN description`

## 2. Domain Structs

- [ ] 2.1 Add `Description *string` to `domain.Image` with GORM tag `column:description`
- [ ] 2.2 Add `Width *int`, `Height *int`, `FileSize *int64` to `domain.Image` with correct GORM column tags
- [ ] 2.3 Add `Description *string` to `domain.Folder` with GORM tag `column:description`

## 3. Image Repository

- [ ] 3.1 Add `CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)` to the `ImageRepository` interface in `internal/usecase/image_repository.go`
- [ ] 3.2 Implement `CountByFolderID` on `imageRepository` in `internal/repository/image_repository.go` — query non-deleted images by `folder_id`
- [ ] 3.3 Add a `CountByFolderID` integration test in `internal/repository/image_repository_integration_test.go`

## 4. Image Usecase

- [ ] 4.1 Add `Description *string` to `UpdateImageParams` in `internal/usecase/image_usecase.go`
- [ ] 4.2 Update `ImageUsecase` interface: add `description *string` parameter to `InitiateUpload`
- [ ] 4.3 Update `imageUsecase.InitiateUpload` to accept and persist `description`
- [ ] 4.4 Update `imageUsecase.UpdateImage` to pass `description` through the `fields` map when present
- [ ] 4.5 Extend `prepareThumbnail` to return `([]byte, int, int, int64, error)` — width, height, and file size alongside thumbnail bytes; use `image.DecodeConfig` on a bytes reader of the raw object for dimensions, and `len(rawBytes)` for file size; treat dimension decode failure as non-fatal (log and return zero values)
- [ ] 4.6 In `CompleteUpload`, capture width/height/fileSize from the updated `prepareThumbnail` return and call `imageRepo.Update` to persist them before spawning the thumbnail goroutine

## 5. Image Handler

- [ ] 5.1 Add `Description *string` to `initiateImageUploadRequest` and pass it to `imageUsecase.InitiateUpload`
- [ ] 5.2 Add `Description *string` to `updateImageRequest` and pass it to `UpdateImageParams`
- [ ] 5.3 Add `Description *string`, `Width *int`, `Height *int`, `FileSize *int64` to `imageResponse` and `imageDetailResponse`
- [ ] 5.4 Populate the new fields in all handler response mapping functions

## 6. Folder Domain & Repository

- [ ] 6.1 Add `FolderDetail` struct to `internal/usecase/` with fields `Folder *domain.Folder` and `ImageCount int64`
- [ ] 6.2 Update `FolderUsecase` interface: `Create` and `Update` gain a `description *string` param; `GetByID` returns `(*FolderDetail, error)`
- [ ] 6.3 Update `folderUsecase` constructor to accept `ImageRepository` as a dependency
- [ ] 6.4 Update `folderUsecase.Create` to accept and persist `description`
- [ ] 6.5 Update `folderUsecase.Update` to accept and persist `description`
- [ ] 6.6 Update `folderUsecase.GetByID` to return `*FolderDetail` — call `imageRepo.CountByFolderID` and include the result
- [ ] 6.7 Update `folderRepository.Create` and `folderRepository.Update` SQL calls to handle the new `description` column
- [ ] 6.8 Wire the `imageRepo` into `NewFolderUsecase` in `main.go`

## 7. Folder Handler

- [ ] 7.1 Add `Description *string` to `folderRequest`
- [ ] 7.2 Add `Description *string` to `folderResponse`
- [ ] 7.3 Add a separate `folderDetailResponse` struct that extends `folderResponse` with `ImageCount int64`
- [ ] 7.4 Update `CreateFolder` handler to pass `description` to usecase
- [ ] 7.5 Update `UpdateFolder` handler to pass `description` to usecase
- [ ] 7.6 Update `GetFolder` handler to use `folderDetailResponse` (mapping `FolderDetail.ImageCount` to `image_count`)
- [ ] 7.7 Update response mapping helpers (`toFolderResponse`) accordingly

## 8. Unit Tests

- [ ] 8.1 Update `imageUsecase` unit tests to cover `InitiateUpload` with description and `UpdateImage` with description
- [ ] 8.2 Update `imageUsecase` unit tests to cover `CompleteUpload` persisting width/height/file_size (success) and graceful handling of decode failure
- [ ] 8.3 Update `imageHandler` unit tests to assert new fields in responses
- [ ] 8.4 Update `folderUsecase` unit tests: add mock `ImageRepository`; cover `Create`/`Update` with description; cover `GetByID` returning `FolderDetail` with `ImageCount`
- [ ] 8.5 Update `folderHandler` unit tests to assert `description` in create/update/list responses and `image_count` in get-by-id response
