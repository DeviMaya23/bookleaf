## ADDED Requirements

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

