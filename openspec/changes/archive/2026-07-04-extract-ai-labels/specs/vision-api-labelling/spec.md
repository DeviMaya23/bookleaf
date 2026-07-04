## MODIFIED Requirements

### Requirement: Image AI Label Persistence

After a successful Vision API call, the system SHALL serialise all returned labels as JSON and persist them via `UpdateLabels` on `ImageRepository`, which atomically writes both `images.ai_labels` (raw JSONB backup) and normalised rows into `image_labels`.

`UpdateLabels(ctx context.Context, id uuid.UUID, rawJSON json.RawMessage, labels []domain.ImageLabel) error` — updates `ai_labels` and replaces `image_labels` rows for the given image atomically; no ownership check (called internally).

All labels returned by Vision SHALL be stored regardless of score, to preserve the full result for future use.

#### Scenario: All labels stored regardless of score

- **WHEN** Vision API returns 5 labels with varying scores
- **THEN** all 5 labels are written to both `images.ai_labels` and `image_labels`

#### Scenario: Repository error is propagated

- **WHEN** `UpdateLabels` returns a database error
- **THEN** the error is returned to the caller

---

### Requirement: ProcessVisionLabelling Usecase Method

The system SHALL implement `ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error` on `imageUploadUsecase`. This is the method called by `VisionWorker` on each attempt.

The method SHALL:
1. Fetch the user record via `UserRepository.GetByID(ctx, userID)`; return nil (not an error) if `vision_enabled` is false or if `visionService` is nil
2. Fetch the image record via `ImageRepository.GetByID(ctx, imageID, userID)` to get `thumbnail_path`
3. Call `StorageService.GetObject(ctx, thumbnailPath)` to fetch the thumbnail bytes; return a non-nil error if `thumbnail_path` is nil
4. Call `VisionService.AnnotateImage` with a 5-second context timeout
5. Marshal the returned labels as JSON (empty array `[]` if Vision returned zero labels), convert `[]domain.Label` to `[]domain.ImageLabel`, and call `ImageRepository.UpdateLabels(ctx, imageID, labelsJSON, imageLabels)`
6. If `user.AICategorisationEnabled` is true, enqueue a `CategoriseImageArgs` job via the `enqueuer`

Failures at steps 1–6 (except the early-return nil cases at step 1) SHALL return a non-nil error so River can retry. `UpdateLabels` SHALL always be called after a successful Vision API call, even when zero labels are returned.

#### Scenario: Vision enabled — labels saved successfully

- **WHEN** `ProcessVisionLabelling` is called for a user with `vision_enabled = true` and Vision returns labels
- **THEN** `image.ai_labels` is updated and `image_labels` rows are inserted with the label data
- **AND** nil is returned

#### Scenario: Vision enabled — zero labels returned

- **WHEN** `ProcessVisionLabelling` is called and Vision returns zero labels
- **THEN** `image.ai_labels` is set to an empty JSON array `[]` and any existing `image_labels` rows for that image are removed
- **AND** nil is returned

#### Scenario: vision_enabled false — returns nil without Vision call

- **WHEN** `ProcessVisionLabelling` is called and the user's `vision_enabled` is false
- **THEN** the Vision API is not called
- **AND** `UpdateLabels` is not called
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
