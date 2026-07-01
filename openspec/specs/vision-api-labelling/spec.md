## Purpose

The system integrates with the Google Cloud Vision API to automatically label images with AI-generated tags. Vision labelling is executed asynchronously via River after an image is committed.

---

## Requirements

### Requirement: VisionService Interface

The system SHALL define a `VisionService` interface in `internal/vision/` with a single `AnnotateImage` method.

```go
type Label struct {
    Description string
    Score       float32
}

type VisionService interface {
    AnnotateImage(ctx context.Context, imageBytes []byte) ([]Label, error)
}
```

Labels SHALL be returned ordered by `Score` descending.

#### Scenario: Interface is satisfied by HTTP client implementation

- **WHEN** the Go package is compiled
- **THEN** the concrete `visionClient` struct satisfies `VisionService` without compilation errors

---

### Requirement: Google Vision HTTP Client

The system SHALL implement `VisionService` as a REST HTTP client that calls the Google Cloud Vision API (`v1/images:annotate`) using an API key. The client SHALL:

- Encode `imageBytes` as a base64 string and request `LABEL_DETECTION` features
- Parse the response and return labels as `[]Label` ordered by score descending
- Respect the context deadline / cancellation passed by the caller

#### Scenario: Labels returned ordered by score

- **WHEN** the Vision API responds with multiple labels at varying scores
- **THEN** `AnnotateImage` returns the labels sorted highest score first

#### Scenario: API error is propagated

- **WHEN** the Vision API returns a non-2xx HTTP status
- **THEN** `AnnotateImage` returns a non-nil error

#### Scenario: Context cancellation is respected

- **WHEN** the context is cancelled before the HTTP response arrives
- **THEN** `AnnotateImage` returns a non-nil error and does not block

---

### Requirement: Image AI Label Persistence

After a successful Vision API call, the system SHALL serialise all returned labels as JSON and persist them to `Image.AILabels` via a new `UpdateAILabels` method on `ImageRepository`.

`UpdateAILabels(ctx context.Context, id uuid.UUID, labels json.RawMessage) error` — updates `ai_labels` for the given image; no ownership check (called internally).

All labels returned by Vision SHALL be stored, regardless of score, to preserve the full result for future use.

#### Scenario: All labels stored regardless of score

- **WHEN** Vision API returns 5 labels with varying scores
- **THEN** all 5 labels are serialised and stored in `Image.AILabels`

#### Scenario: Repository error is propagated

- **WHEN** `UpdateAILabels` returns a database error
- **THEN** the error is returned to the caller

---

### Requirement: FolderRepository FindByName Method

The system SHALL add `FindByName(ctx context.Context, userID, name string) (*domain.Folder, error)` to the `FolderRepository` interface and its SQL implementation.

- The query SHALL be case-insensitive (`ILIKE` or `LOWER()` comparison)
- If no matching folder exists, the method SHALL return `nil, nil` (not `ErrRecordNotFound`)
- Only non-deleted folders SHALL be considered

#### Scenario: Existing folder matched case-insensitively

- **WHEN** the user has a folder named `"Nature"` and `FindByName` is called with `"nature"`
- **THEN** the `Nature` folder is returned

#### Scenario: No matching folder returns nil

- **WHEN** the user has no folder matching the given name
- **THEN** `FindByName` returns `nil, nil`

---

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

---

### Requirement: Vision Flow is Executed Asynchronously via River

The system SHALL NOT call vision labelling synchronously inside `CompleteUpload`. The `runVisionFlow` method SHALL be removed. Vision labelling SHALL be triggered exclusively via the `VisionJob` River job inserted by `CompleteUpload` after the image DB write.

#### Scenario: CompleteUpload does not call Vision API synchronously

- **WHEN** `CompleteUpload` is called
- **THEN** no Vision API call is made within the request's execution path
- **AND** a `VisionArgs` River job is inserted for asynchronous processing

---

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
