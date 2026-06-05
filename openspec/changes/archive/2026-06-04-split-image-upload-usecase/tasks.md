## 1. Create `image_upload_usecase.go`

- [x] 1.1 Create `internal/usecase/image_upload_usecase.go` with the `imageUploadUsecase` struct and its constructor `NewImageUploadUsecase`
- [x] 1.2 Define the upload usecase's own interfaces: `ImageRepository` (Create, SetImageFolder, GetByID), `FolderRepository` (GetByID, FindByName, Create), `PendingUploadRepository`, `UserRepository`, `StorageService` — each scoped to only the methods needed by this usecase
- [x] 1.3 Move result types `UploadInitResult` and `CompleteUploadResult` from `image_usecase.go` to `image_upload_usecase.go`
- [x] 1.4 Move constants `uploadURLTTL` and `ThumbnailService`/`VisionService` interface declarations if not already in a separate file
- [x] 1.5 Move methods `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion`, `CleanupStaleUploads` to the new struct
- [x] 1.6 Move private helpers `prepareThumbnail`, `runVisionFlow`, `uploadThumbnail` to the new struct
- [x] 1.7 Move upload OTel metrics (`uploadCount`, `thumbnailDuration`, `thumbnailCount`) and their initialisation to the new constructor

## 2. Update `image_usecase.go`

- [x] 2.1 Remove `pendingUploadRepo`, `thumbnails`, `visionService`, `userRepo` fields from `imageUsecase` struct and constructor
- [x] 2.2 Remove `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion`, `CleanupStaleUploads` methods
- [x] 2.3 Remove private helpers `prepareThumbnail`, `runVisionFlow`, `uploadThumbnail`
- [x] 2.4 Remove `UploadInitResult`, `CompleteUploadResult` type declarations
- [x] 2.5 Remove `uploadURLTTL` constant and `ThumbnailService`/`VisionService` interface declarations if moved
- [x] 2.6 Remove upload metric fields and initialisation from constructor
- [x] 2.7 Verify `image_usecase.go` compiles cleanly with no unused imports

## 3. Create `image_upload_usecase_test.go`

- [x] 3.1 Move all `TestImageUsecase_InitiateUpload_*` tests (blank title, blank mime type, R2 path format, folder found, folder not found, create pending fails) — rename prefix to `TestImageUploadUsecase_`
- [x] 3.2 Move all `TestImageUsecase_CompleteUpload_*` tests (persists metadata, decode failure, sets folder, vision enabled/disabled/fails, upload count success/error) — rename prefix
- [x] 3.3 Move `TestImageUsecase_AcceptSuggestion_ExistingFolder` and `TestImageUsecase_AcceptSuggestion_CreatesFolder` — rename prefix
- [x] 3.4 Move `TestImageUsecase_CleanupStaleUploads_*` tests — rename prefix
- [x] 3.5 Update test helpers and spy/fake construction to use the new `imageUploadUsecase` constructor and its interfaces

## 4. Update `image_usecase_test.go`

- [x] 4.1 Remove all upload-related test functions moved in task 3
- [x] 4.2 Verify remaining tests compile and pass (`go test ./internal/usecase/...`)

## 5. Create `internal/handler/image_upload.go`

- [x] 5.1 Define `UploadUsecase` interface with `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion`
- [x] 5.2 Create `UploadHandler` struct with `uploadUsecase UploadUsecase` and `tel *observability.Telemetry` fields
- [x] 5.3 Move `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion` handler methods from `ImageHandler` to `UploadHandler`

## 6. Update `internal/handler/image.go`

- [x] 6.1 Remove `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion` from the `ImageUsecase` interface
- [x] 6.2 Remove `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion` handler methods from `ImageHandler`
- [x] 6.3 Verify `image.go` compiles cleanly

## 7. Create `internal/handler/image_upload_test.go`

- [x] 7.1 Move `TestImageHandler_InitiateUpload`, `TestImageHandler_InitiateUpload_MalformedJSON`, `TestImageHandler_InitiateUpload_PassesDescription` — rename prefix to `TestUploadHandler_`
- [x] 7.2 Move `TestImageHandler_CompleteUpload` and `TestImageHandler_CompleteUpload_InvalidUUID` — rename prefix
- [x] 7.3 Move `TestImageHandler_AcceptSuggestion`, `TestImageHandler_AcceptSuggestion_BlankFolderName`, `TestImageHandler_AcceptSuggestion_InvalidUUID`, `TestImageHandler_AcceptSuggestion_MalformedJSON` — rename prefix
- [x] 7.4 Update spy construction to use `UploadUsecase` interface and `NewUploadHandler`

## 8. Update `internal/handler/image_test.go`

- [x] 8.1 Remove all upload-related test functions moved in task 7
- [x] 8.2 Verify remaining tests compile and pass (`go test ./internal/handler/...`)

## 9. Wire up in `cmd/server/main.go`

- [x] 9.1 Add `usecase.NewImageUploadUsecase(...)` constructor call with its dependencies
- [x] 9.2 Add `handler.NewUploadHandler(uploadUsecase, tel)` constructor call
- [x] 9.3 Move the three upload routes (`POST /images`, `POST /images/:id/complete`, `POST /images/:id/accept-suggestion`) to `uploadHandler`
- [x] 9.4 Run `go build ./...` and `go test ./...` to confirm everything compiles and passes
