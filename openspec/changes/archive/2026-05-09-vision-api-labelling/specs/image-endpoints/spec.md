## ADDED Requirements

### Requirement: CompleteUpload Response Body

The `POST /images/:id/complete` handler SHALL return `200 OK` with a JSON body on success.

Response shape:
```json
{
  "image_id": "<uuid>",
  "folder_suggestion": {
    "folder_id": "<uuid | null>",
    "folder_name": "<string>",
    "is_new": true
  },
  "warning": "<string>"
}
```

- `image_id` SHALL always be present
- `folder_suggestion` SHALL be `null` when the user does not have `vision_enabled`, when the Vision API returns no labels, or when Vision is not configured
- `warning` SHALL be omitted from the response when empty (`omitempty`)

#### Scenario: Vision enabled and folder matched

- **WHEN** `CompleteUpload` succeeds and a folder suggestion is resolved
- **THEN** the response is `200 OK`
- **AND** `folder_suggestion.folder_id` is the matched folder's UUID
- **AND** `folder_suggestion.is_new` is `false`
- **AND** `warning` is absent from the response body

#### Scenario: Vision enabled but API call fails

- **WHEN** `CompleteUpload` succeeds but the Vision API returns an error
- **THEN** the response is still `200 OK`
- **AND** `folder_suggestion` is `null`
- **AND** `warning` is a non-empty string describing the failure

#### Scenario: Vision not enabled

- **WHEN** the image owner has `vision_enabled = false`
- **THEN** the response is `200 OK`
- **AND** `folder_suggestion` is `null`
- **AND** `warning` is absent

---

## MODIFIED Requirements

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations, including `UpdateImage`. The `CompleteUpload` method SHALL return a result struct alongside the error.

```go
CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error)
```

Where `CompleteUploadResult` is defined in `internal/usecase/`:
```go
type FolderSuggestion struct {
    FolderID   *uuid.UUID
    FolderName string
    IsNew      bool
}

type CompleteUploadResult struct {
    ImageID          uuid.UUID
    FolderSuggestion *FolderSuggestion
    Warning          string
}
```

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors
