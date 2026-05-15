## Why

The current AI labelling flow resolves folder existence eagerly during the Vision pipeline (`runVisionFlow`), coupling folder lookup logic to the upload completion step. Moving folder resolution to an explicit `accept-suggestion` endpoint separates concerns: the upload flow only surfaces a name suggestion, and the client decides when and whether to accept it.

## What Changes

- `FolderSuggestion` struct is removed entirely — `CompleteUploadResult` replaces it with a plain `SuggestedFolderName *string` field
- `POST /images/:id/complete` response replaces the `folder_suggestion` object with a flat `suggested_folder_name` string field (or null)
- **BREAKING**: `folder_suggestion` object removed from complete-upload response
- `FolderRepository.FindByName` call is removed from `runVisionFlow`; the Vision flow only resolves a label name
- New endpoint `POST /images/:id/accept-suggestion` accepts `{ "suggested_folder_name": "Nature" }` and handles the existing-vs-new folder logic server-side (find or create folder, assign to image)

## Capabilities

### New Capabilities

- `accept-suggestion`: New endpoint that accepts a folder name suggestion for an image — finds an existing folder by name (case-insensitive) or creates a new one, then assigns it to the image

### Modified Capabilities

- `vision-api-labelling`: `FolderSuggestion` resolution requirement removed entirely — `FindByName` is no longer called in `runVisionFlow`; Vision flow only surfaces a label name string
- `image-endpoints`: `CompleteUpload` response body changes — `folder_suggestion` object replaced with flat `suggested_folder_name` string (or null)

## Impact

- `internal/usecase/`: `FolderSuggestion` struct deleted, `CompleteUploadResult` gains `SuggestedFolderName *string`, new `AcceptSuggestion` method on `ImageUsecase`
- `internal/vision/`: `runVisionFlow` no longer calls `FolderRepository.FindByName`
- `internal/handler/`: new `AcceptSuggestionHandler`, updated `CompleteUploadHandler` response mapping
- `main.go`: new route `POST /images/:id/accept-suggestion`
- API contract change for `POST /images/:id/complete` response (breaking for clients consuming `folder_id`/`is_new`)
