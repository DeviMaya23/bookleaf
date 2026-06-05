## MODIFIED Requirements

### Requirement: ProcessVisionLabelling Usecase Method

The system SHALL add `ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error` to `imageUploadUsecase`. This is the method called by `VisionWorker` on each attempt.

The method SHALL:
1. Fetch the user record via `UserRepository.GetByID(ctx, userID)`; return nil (not an error) if `vision_enabled` is false or if `visionService` is nil
2. Fetch the image record via `ImageRepository.GetByID(ctx, imageID, userID)` to get `r2_path`
3. Call `StorageService.GetObject(ctx, r2Path)` to fetch the image bytes
4. Call `VisionService.AnnotateImage` with a 5-second context timeout
5. Marshal the returned labels as JSON (empty array `[]` if Vision returned zero labels) and call `ImageRepository.UpdateAILabels(ctx, imageID, labelsJSON)`

Failures at steps 1–5 (except the early-return nil cases at step 1) SHALL return a non-nil error so River can retry. `UpdateAILabels` SHALL always be called after a successful Vision API call, even when zero labels are returned, so that `null` in the database reliably indicates the job has not yet run.

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

---

## ADDED Requirements

### Requirement: suggested_folder_name on GET /images/:id

The `GET /images/:id` response SHALL include a `suggested_folder_name` field. This field SHALL be derived at response time from `Image.AILabels`:

- If `AILabels` is null (vision job has not run yet) or an empty array (job ran, no labels), `suggested_folder_name` SHALL be null.
- If `AILabels` contains one or more labels, `suggested_folder_name` SHALL be the `Description` of the first label (labels are stored ordered by Vision score descending, so the first is the highest-scoring).

No new database column or query is required — the field is computed inline in the `GetImage` usecase or handler.

#### Scenario: ai_labels null — suggested_folder_name is null

- **WHEN** `GET /images/:id` is called and `Image.AILabels` is null
- **THEN** the response includes `"suggested_folder_name": null`

#### Scenario: ai_labels empty — suggested_folder_name is null

- **WHEN** `GET /images/:id` is called and `Image.AILabels` is an empty array
- **THEN** the response includes `"suggested_folder_name": null`

#### Scenario: ai_labels populated — suggested_folder_name is top label

- **WHEN** `GET /images/:id` is called and `Image.AILabels` contains labels
- **THEN** the response includes `"suggested_folder_name"` set to the `Description` of the first label
