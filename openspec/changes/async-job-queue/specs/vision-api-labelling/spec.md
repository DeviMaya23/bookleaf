## ADDED Requirements

### Requirement: ProcessVisionLabelling Usecase Method

The system SHALL add `ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error` to `imageUploadUsecase`. This is the method called by `VisionWorker` on each attempt.

The method SHALL:
1. Fetch the user record via `UserRepository.GetByID(ctx, userID)`; return nil (not an error) if `vision_enabled` is false or if `visionService` is nil
2. Fetch the image record via `ImageRepository.GetByID(ctx, imageID, userID)` to get `r2_path`
3. Call `StorageService.GetObject(ctx, r2Path)` to fetch the image bytes
4. Call `VisionService.AnnotateImage` with a 5-second context timeout
5. Marshal the returned labels as JSON and call `ImageRepository.UpdateAILabels(ctx, imageID, labelsJSON)`

Failures at steps 1–5 (except the early-return nil cases) SHALL return a non-nil error so River can retry. If Vision returns zero labels, the method SHALL return nil without calling `UpdateAILabels`.

#### Scenario: Vision enabled — labels saved successfully

- **WHEN** `ProcessVisionLabelling` is called for a user with `vision_enabled = true` and Vision returns labels
- **THEN** `image.ai_labels` is updated with the serialised label array
- **AND** nil is returned

#### Scenario: vision_enabled false — returns nil without Vision call

- **WHEN** `ProcessVisionLabelling` is called and the user's `vision_enabled` is false
- **THEN** the Vision API is not called
- **AND** nil is returned

#### Scenario: Vision API error returns error for retry

- **WHEN** `VisionService.AnnotateImage` returns an error
- **THEN** `ProcessVisionLabelling` returns a non-nil error

---

### Requirement: Vision Flow is Executed Asynchronously via River

The system SHALL NOT call vision labelling synchronously inside `CompleteUpload`. The `runVisionFlow` method SHALL be removed. Vision labelling SHALL be triggered exclusively via the `VisionJob` River job inserted by `CompleteUpload` after the image DB write.

#### Scenario: CompleteUpload does not call Vision API synchronously

- **WHEN** `CompleteUpload` is called
- **THEN** no Vision API call is made within the request's execution path
- **AND** a `VisionArgs` River job is inserted for asynchronous processing
