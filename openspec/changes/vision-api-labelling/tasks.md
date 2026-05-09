## 1. Config

- [ ] 1.1 Add `VisionConfig` struct with `APIKey string` to `internal/config/config.go`
- [ ] 1.2 Add `Vision VisionConfig` field to `Config` struct
- [ ] 1.3 Load `GOOGLE_VISION_API_KEY` via `envWithDefault` in `loadFromEnv`

## 2. Vision Package

- [ ] 2.1 Create `internal/vision/vision.go` with `Label` struct and `VisionService` interface
- [ ] 2.2 Implement `visionClient` HTTP client (base64 encode, call `v1/images:annotate`, parse response, sort by score desc)
- [ ] 2.3 Add `NewVisionClient(apiKey string) VisionService` constructor

## 3. FolderRepository

- [ ] 3.1 Add `FindByName(ctx, userID, name string) (*domain.Folder, error)` to `FolderRepository` interface in `internal/usecase/folder_repository.go`
- [ ] 3.2 Implement `FindByName` in `internal/repository/folder_repository.go` using `ILIKE`; return `nil, nil` when not found

## 4. ImageRepository

- [ ] 4.1 Add `UpdateAILabels(ctx, id uuid.UUID, labels json.RawMessage) error` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [ ] 4.2 Implement `UpdateAILabels` in `internal/repository/image_repository.go`

## 5. Usecase — Result Types and Signature

- [ ] 5.1 Add `FolderSuggestion` and `CompleteUploadResult` structs to `internal/usecase/image_usecase.go`
- [ ] 5.2 Update `ImageUsecase` interface: change `CompleteUpload` return type to `(*CompleteUploadResult, error)`
- [ ] 5.3 Add `VisionService` and `FolderRepository` fields to `imageUsecase` struct; update `NewImageUsecase` constructor

## 6. Usecase — CompleteUpload Restructure

- [ ] 6.1 Move `StorageService.GetObject` and `ThumbnailService.Generate` out of `generateThumbnail` goroutine into the synchronous `CompleteUpload` body; buffer thumbnail as `[]byte`
- [ ] 6.2 On GetObject or Generate failure, log the error, set `Warning` on result, and return without launching the goroutine
- [ ] 6.3 Update goroutine to accept pre-generated thumbnail bytes and perform only `PutObject` + `UpdateThumbnailPath`

## 7. Usecase — Vision and Folder Suggestion

- [ ] 7.1 After thumbnail generation, fetch the user record via `UserRepository.GetByID` and check `VisionEnabled`
- [ ] 7.2 If enabled and `VisionService` is non-nil, call `AnnotateImage` with a 5-second timeout context
- [ ] 7.3 On Vision failure, log the error and set `Warning` on result; continue without labels
- [ ] 7.4 On success, call `UpdateAILabels` to persist all returned labels as JSON
- [ ] 7.5 If labels are non-empty, call `FolderRepository.FindByName` with the top label description and resolve `FolderSuggestion`

## 8. Handler

- [ ] 8.1 Update `CompleteUpload` handler to consume `*CompleteUploadResult` from the usecase
- [ ] 8.2 Define `completeUploadResponse` struct with `ImageID`, `FolderSuggestion` (nullable), and `Warning` (omitempty)
- [ ] 8.3 Change handler response from `204 NoContent` to `200 OK` with the response body

## 9. Wiring

- [ ] 9.1 In `main.go`, construct `VisionClient` from `cfg.Vision.APIKey` if non-empty (nil otherwise)
- [ ] 9.2 Pass `VisionService`, `FolderRepository`, and `UserRepository` into `NewImageUsecase`

## 10. Unit Tests

- [ ] 10.1 `imageUsecase.CompleteUpload` — success: vision enabled, folder matched, returns correct result
- [ ] 10.2 `imageUsecase.CompleteUpload` — vision disabled: `FolderSuggestion` is nil, no Vision call made
- [ ] 10.3 `imageUsecase.CompleteUpload` — Vision API failure: returns 200 result with `Warning` set
- [ ] 10.4 `imageUsecase.CompleteUpload` — GetObject failure: warning set, goroutine not launched
- [ ] 10.5 `ImageHandler.CompleteUpload` — success: response body contains `image_id` and `folder_suggestion`
- [ ] 10.6 `ImageHandler.CompleteUpload` — warning present: response body contains `warning` field
