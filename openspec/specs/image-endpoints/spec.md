## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx, userID string, folderID *uuid.UUID) ([]*domain.Image, error)` — returns non-deleted images; `folderID` nil means no filter
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `UpdateAILabels(ctx, id uuid.UUID, labels json.RawMessage) error`
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied fields for the image owned by `userID`
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx, userID string) ([]*domain.Image, error)`
- `CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)` — counts non-deleted images belonging to the given folder

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

---

### Requirement: Image Routes Wiring

The system SHALL register image routes on the protected Echo group in `main.go`.

Routes:
- `POST /images`
- `POST /images/:id/complete`
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

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations, including `UpdateImage`. The `CompleteUpload` method SHALL return a result struct alongside the error.

```go
CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error)
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

---

### Requirement: POST /images — Initiate Upload Request and Response

The `POST /images` handler SHALL accept an optional `description` field in the request body and persist it on the image record.

Request body:
```json
{
  "title": "string (required)",
  "mime_type": "string (required)",
  "source_url": "string (optional)",
  "folder_id": "uuid (optional)",
  "description": "string (optional)"
}
```

Response body (201): `id`, `upload_url`, `r2_path`.

#### Scenario: Upload initiated with description

- **WHEN** an authenticated `POST /images` request includes a non-empty `description`
- **THEN** the response is `201 Created`
- **AND** the created image record has the supplied `description` value persisted

#### Scenario: Upload initiated without description

- **WHEN** an authenticated `POST /images` request omits `description`
- **THEN** the response is `201 Created`
- **AND** the image record has `description` as NULL

---

### Requirement: GET /images and GET /images/:id — Response Shape

Image list and detail responses SHALL include `description`, `width`, `height`, and `file_size`.

Updated `imageResponse` (list) shape:
```json
{
  "id": "uuid",
  "title": "string",
  "description": "string|null",
  "mime_type": "string",
  "source_url": "string|null",
  "folder_id": "uuid|null",
  "thumbnail_url": "string|null",
  "width": "integer|null",
  "height": "integer|null",
  "file_size": "integer|null",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

`imageDetailResponse` (single) includes the same fields plus `image_url`.

#### Scenario: Image list response includes new fields

- **WHEN** an authenticated `GET /images` request is made
- **THEN** each image object in the response includes `description`, `width`, `height`, and `file_size` (null when not yet populated)

#### Scenario: Image detail response includes new fields

- **WHEN** an authenticated `GET /images/:id` request is made for an existing image
- **THEN** the response includes `description`, `width`, `height`, and `file_size`

---

### Requirement: PATCH /images/:id — Accept Description

The `PATCH /images/:id` handler SHALL accept an optional `description` field and persist it.

Updated request body (all fields optional):
```json
{
  "title": "string",
  "folder_id": "uuid|null",
  "description": "string|null"
}
```

#### Scenario: Update sets description

- **WHEN** an authenticated `PATCH /images/:id` request includes a `description` field
- **THEN** the response is `200 OK`
- **AND** the image record's `description` is updated to the supplied value

#### Scenario: Update clears description

- **WHEN** an authenticated `PATCH /images/:id` request sets `description` to `null`
- **THEN** the response is `200 OK`
- **AND** the image record's `description` is set to NULL

---

### Requirement: CompleteUpload Populates Dimensions and File Size

The `POST /images/:id/complete` usecase SHALL calculate and persist `width`, `height`, and `file_size` from the image bytes fetched from R2.

- `width` and `height` SHALL be decoded using Go's `image.DecodeConfig` on the original image bytes
- `file_size` SHALL be set to the byte length of the fetched object
- If dimension decoding fails (unsupported format), `width` and `height` SHALL be left as NULL and the failure SHALL be logged; `file_size` SHALL still be persisted
- These fields SHALL be persisted via `imageRepo.Update` before the thumbnail goroutine is spawned
- The `CompleteUpload` response body is unchanged

#### Scenario: Dimensions and size extracted successfully

- **WHEN** `CompleteUpload` is called for a JPEG or PNG image
- **THEN** the image record is updated with non-null `width`, `height`, and `file_size`

#### Scenario: Unsupported format — size persisted, dimensions skipped

- **WHEN** `CompleteUpload` is called for an image format not supported by `image.DecodeConfig`
- **THEN** `file_size` is persisted
- **AND** `width` and `height` remain NULL
- **AND** the decode error is logged but `CompleteUpload` does not return an error

---

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
