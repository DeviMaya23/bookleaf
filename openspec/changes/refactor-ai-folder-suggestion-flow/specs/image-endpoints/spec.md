## MODIFIED Requirements

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations, including `UpdateImage`. The `CompleteUpload` method SHALL return a result struct alongside the error.

```go
CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error)
AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error
```

`InitiateUpload` SHALL accept a `description *string` parameter:

```go
InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*UploadInitResult, error)
```

`UpdateImageParams` SHALL include a `Description` field:

```go
type UpdateImageParams struct {
    Title       *string
    FolderID    **uuid.UUID
    Description *string
}
```

`CompleteUploadResult` is defined in `internal/usecase/`. The `FolderSuggestion` struct is removed; the result carries a plain string field instead:

```go
type CompleteUploadResult struct {
    ImageID              uuid.UUID
    SuggestedFolderName  *string
    Warning              string
}
```

`ListImages` and `ListTrashed` SHALL use the paginated signatures:

```go
ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error)
ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error)
```

All other method signatures are unchanged.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors

---

### Requirement: Image Routes Wiring

The system SHALL register image routes on the protected Echo group in `main.go`.

Routes:
- `POST /images`
- `POST /images/:id/complete`
- `POST /images/:id/accept-suggestion`
- `GET /images`
- `GET /images/:id`
- `PATCH /images/:id`
- `DELETE /images/:id`
- `GET /images/trash`
- `POST /images/:id/restore`

#### Scenario: Image routes are registered under auth middleware

- **WHEN** the server starts
- **THEN** all `/images` routes require a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`

---

### Requirement: CompleteUpload Response Body

The `POST /images/:id/complete` handler SHALL return `200 OK` with a JSON body on success.

Response shape:
```json
{
  "image_id": "<uuid>",
  "suggested_folder_name": "<string | null>",
  "warning": "<string>"
}
```

- `image_id` SHALL always be present
- `suggested_folder_name` SHALL be `null` when the user does not have `vision_enabled`, when the Vision API returns no labels, or when Vision is not configured
- `warning` SHALL be omitted from the response when empty (`omitempty`)

#### Scenario: Vision enabled and suggestion resolved

- **WHEN** `CompleteUpload` succeeds and Vision returns at least one label
- **THEN** the response is `200 OK`
- **AND** `suggested_folder_name` is the top label description string
- **AND** `warning` is absent from the response body

#### Scenario: Vision enabled but API call fails

- **WHEN** `CompleteUpload` succeeds but the Vision API returns an error
- **THEN** the response is still `200 OK`
- **AND** `suggested_folder_name` is `null`
- **AND** `warning` is a non-empty string describing the failure

#### Scenario: Vision not enabled

- **WHEN** the image owner has `vision_enabled = false`
- **THEN** the response is `200 OK`
- **AND** `suggested_folder_name` is `null`
- **AND** `warning` is absent
