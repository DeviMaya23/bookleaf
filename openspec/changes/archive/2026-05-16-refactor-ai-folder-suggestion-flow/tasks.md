## 1. Simplify Vision Flow and Remove FolderSuggestion

- [x] 1.1 Delete the `FolderSuggestion` struct from `internal/usecase/image_usecase.go`
- [x] 1.2 Replace `FolderSuggestion *FolderSuggestion` with `SuggestedFolderName *string` on `CompleteUploadResult`
- [x] 1.3 Update `runVisionFlow` return type to `(*string, string)` and remove the `FolderRepository.FindByName` call — return `&topLabel.Description` as the suggestion
- [x] 1.4 Update the `CompleteUpload` usecase method to assign `result.SuggestedFolderName` from `runVisionFlow`

## 2. Add AcceptSuggestion Usecase Method

- [x] 2.1 Add `AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error` to the `ImageUsecase` interface
- [x] 2.2 Implement `AcceptSuggestion` on `imageUsecase`: get image by ID+userID, call `FindByName`, create folder if not found, update image with resolved folder ID
- [x] 2.3 Write unit tests for `AcceptSuggestion`: success with existing folder, success with new folder created, image not found error

## 3. Update CompleteUpload Handler Response

- [x] 3.1 Remove `completeUploadFolderSuggestionResponse` struct from `internal/handler/image.go`
- [x] 3.2 Update `completeUploadResponse` to replace `FolderSuggestion *completeUploadFolderSuggestionResponse` with `SuggestedFolderName *string \`json:"suggested_folder_name,omitempty"\``
- [x] 3.3 Update `CompleteUpload` handler to map `result.SuggestedFolderName` directly onto the response
- [x] 3.4 Write unit tests for `CompleteUpload` handler: response includes `suggested_folder_name` string, response has null when no suggestion

## 4. Add AcceptSuggestion Handler and Route

- [x] 4.1 Add `AcceptSuggestion` handler method to `ImageHandler` in `internal/handler/image.go` — parse `:id`, bind body, call usecase, return `204`
- [x] 4.2 Register `POST /images/:id/accept-suggestion` on the protected group in `main.go`
- [x] 4.3 Write unit tests for `AcceptSuggestion` handler: success returns 204, missing body field returns 400, image not found returns 404
