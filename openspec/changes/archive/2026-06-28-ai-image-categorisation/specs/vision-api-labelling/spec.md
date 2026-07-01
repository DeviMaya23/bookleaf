## MODIFIED Requirements

### Requirement: ProcessVisionLabelling Usecase Method

The system SHALL implement `ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error` on `imageUploadUsecase`. This is the method called by `VisionWorker` on each attempt.

The method SHALL:
1. Fetch the user record via `UserRepository.GetByID(ctx, userID)`; return nil (not an error) if `vision_enabled` is false or if `visionService` is nil
2. Fetch the image record via `ImageRepository.GetByID(ctx, imageID, userID)` to get `thumbnail_path`
3. Call `StorageService.GetObject(ctx, thumbnailPath)` to fetch the thumbnail bytes; return a non-nil error if `thumbnail_path` is nil
4. Call `VisionService.AnnotateImage` with a 5-second context timeout
5. Marshal the returned labels as JSON (empty array `[]` if Vision returned zero labels) and call `ImageRepository.UpdateAILabels(ctx, imageID, labelsJSON)`
6. If `user.AICategorisationEnabled` is true, enqueue a `CategoriseImageArgs` job via the `enqueuer`

Failures at steps 1–6 (except the early-return nil cases at step 1) SHALL return a non-nil error so River can retry. `UpdateAILabels` SHALL always be called after a successful Vision API call, even when zero labels are returned.

#### Scenario: Vision enabled — labels saved successfully

- **WHEN** `ProcessVisionLabelling` is called for a user with `vision_enabled = true` and Vision returns labels
- **THEN** `image.ai_labels` is updated with the serialised label array
- **AND** nil is returned

#### Scenario: Vision enabled — zero labels returned

- **WHEN** `ProcessVisionLabelling` is called and Vision returns zero labels
- **THEN** `image.ai_labels` is updated with an empty JSON array `[]`
- **AND** nil is returned

#### Scenario: vision_enabled false — returns nil without Vision call

- **WHEN** `ProcessVisionLabelling` is called and the user's `vision_enabled` is false
- **THEN** the Vision API is not called
- **AND** `UpdateAILabels` is not called
- **AND** nil is returned

#### Scenario: Vision API error returns error for retry

- **WHEN** `VisionService.AnnotateImage` returns an error
- **THEN** `ProcessVisionLabelling` returns a non-nil error

#### Scenario: ai_categorisation_enabled true — categorise job enqueued after labels saved

- **WHEN** `ProcessVisionLabelling` completes successfully and the user's `ai_categorisation_enabled` is true
- **THEN** a `CategoriseImageArgs` job is inserted with the image ID and user ID
- **AND** nil is returned

#### Scenario: ai_categorisation_enabled false — categorise job not enqueued

- **WHEN** `ProcessVisionLabelling` completes successfully and the user's `ai_categorisation_enabled` is false
- **THEN** no `CategoriseImageArgs` job is inserted
- **AND** nil is returned
