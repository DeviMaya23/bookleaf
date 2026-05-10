## 1. Migrations

- [x] 1.1 Add migration 000006 (`add_image_metadata_fields`) — up: `ALTER TABLE images ADD COLUMN description text, ADD COLUMN width integer, ADD COLUMN height integer, ADD COLUMN file_size bigint`; down: drop those columns
- [x] 1.2 Add migration 000007 (`add_folder_description`) — up: `ALTER TABLE folders ADD COLUMN description text`; down: `DROP COLUMN description`

## 2. Domain Structs

- [x] 2.1 Add `Description *string` to `domain.Image` with GORM tag `column:description`
- [x] 2.2 Add `Width *int`, `Height *int`, `FileSize *int64` to `domain.Image` with correct GORM column tags
- [x] 2.3 Add `Description *string` to `domain.Folder` with GORM tag `column:description`

## 3. Image Repository

- [x] 3.1 Add `CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)` to the `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 3.2 Implement `CountByFolderID` on `imageRepository` in `internal/repository/image_repository.go` — query non-deleted images by `folder_id`
- [x] 3.3 Add a `CountByFolderID` integration test in `internal/repository/image_repository_integration_test.go`

## 4. Image Usecase

- [x] 4.1 Add `Description *string` to `UpdateImageParams` in `internal/usecase/image_usecase.go`
- [x] 4.2 Update `ImageUsecase` interface: add `description *string` parameter to `InitiateUpload`
- [x] 4.3 Update `imageUsecase.InitiateUpload` to accept and persist `description`
- [x] 4.4 Update `imageUsecase.UpdateImage` to pass `description` through the `fields` map when present
- [x] 4.5 Extend `prepareThumbnail` to return `([]byte, int, int, int64, error)` — width, height, and file size alongside thumbnail bytes; use `image.DecodeConfig` on a bytes reader of the raw object for dimensions, and `len(rawBytes)` for file size; treat dimension decode failure as non-fatal (log and return zero values)
- [x] 4.6 In `CompleteUpload`, capture width/height/fileSize from the updated `prepareThumbnail` return and call `imageRepo.Update` to persist them before spawning the thumbnail goroutine

## 5. Image Handler

- [x] 5.1 Add `Description *string` to `initiateImageUploadRequest` and pass it to `imageUsecase.InitiateUpload`
- [x] 5.2 Add `Description *string` to `updateImageRequest` and pass it to `UpdateImageParams`
- [x] 5.3 Add `Description *string`, `Width *int`, `Height *int`, `FileSize *int64` to `imageResponse` and `imageDetailResponse`
- [x] 5.4 Populate the new fields in all handler response mapping functions

## 6. Folder Domain & Repository

- [x] 6.1 Add `FolderDetail` struct to `internal/usecase/` with fields `Folder *domain.Folder` and `ImageCount int64`
- [x] 6.2 Update `FolderUsecase` interface: `Create` and `Update` gain a `description *string` param; `GetByID` returns `(*FolderDetail, error)`
- [x] 6.3 Update `folderUsecase` constructor to accept `ImageRepository` as a dependency
- [x] 6.4 Update `folderUsecase.Create` to accept and persist `description`
- [x] 6.5 Update `folderUsecase.Update` to accept and persist `description`
- [x] 6.6 Update `folderUsecase.GetByID` to return `*FolderDetail` — call `imageRepo.CountByFolderID` and include the result
- [x] 6.7 Update `folderRepository.Create` and `folderRepository.Update` SQL calls to handle the new `description` column
- [x] 6.8 Wire the `imageRepo` into `NewFolderUsecase` in `main.go`

## 7. Folder Handler

- [x] 7.1 Add `Description *string` to `folderRequest`
- [x] 7.2 Add `Description *string` to `folderResponse`
- [x] 7.3 Add a separate `folderDetailResponse` struct that extends `folderResponse` with `ImageCount int64`
- [x] 7.4 Update `CreateFolder` handler to pass `description` to usecase
- [x] 7.5 Update `UpdateFolder` handler to pass `description` to usecase
- [x] 7.6 Update `GetFolder` handler to use `folderDetailResponse` (mapping `FolderDetail.ImageCount` to `image_count`)
- [x] 7.7 Update response mapping helpers (`toFolderResponse`) accordingly

## 8. Unit Tests

- [x] 8.1 Update `imageUsecase` unit tests to cover `InitiateUpload` with description and `UpdateImage` with description
- [x] 8.2 Update `imageUsecase` unit tests to cover `CompleteUpload` persisting width/height/file_size (success) and graceful handling of decode failure
- [x] 8.3 Update `imageHandler` unit tests to assert new fields in responses
- [x] 8.4 Update `folderUsecase` unit tests: add mock `ImageRepository`; cover `Create`/`Update` with description; cover `GetByID` returning `FolderDetail` with `ImageCount`
- [x] 8.5 Update `folderHandler` unit tests to assert `description` in create/update/list responses and `image_count` in get-by-id response
