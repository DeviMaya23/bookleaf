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
