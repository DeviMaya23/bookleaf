## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, folderID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns non-deleted images ordered by `(created_at DESC, id DESC)`; fetches `limit + 1` rows so the caller can detect next-page existence; `folderID` nil means no filter; `cursor` nil means first page
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `UpdateAILabels(ctx, id uuid.UUID, labels json.RawMessage) error`
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied fields for the image owned by `userID`
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns soft-deleted images ordered by `(created_at DESC, id DESC)`; fetches `limit + 1` rows; `cursor` nil means first page
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

`Restore` and `UpdateImage` SHALL return `*ImageItem` instead of `*domain.Image`:

```go
Restore(ctx context.Context, id uuid.UUID, userID string) (*ImageItem, error)
UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*ImageItem, error)
```

`ImageItem` is defined in `internal/usecase/`:

```go
type ImageItem struct {
    Image        *domain.Image
    ThumbnailURL *string
}
```

`ListImagesResult.Images` and `ListTrashedResult.Images` SHALL be `[]ImageItem`.

`ImageDetail` SHALL include a `ThumbnailURL *string` field alongside `ImageURL`:

```go
type ImageDetail struct {
    Image        *domain.Image
    ImageURL     string
    ThumbnailURL *string
}
```

All other method signatures are unchanged.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors

---

### Requirement: Thumbnail URL Generation

The usecase SHALL generate presigned GET URLs for thumbnails with a 24h TTL. A private helper `thumbnailURL(ctx context.Context, path *string) *string` on `imageUsecase` SHALL:

- Return `nil` if `path` is `nil`
- Call `store.GeneratePresignedGetURL` with `presignedGetTTL` (24h)
- Return `nil` if presigning fails (non-fatal; thumbnail is cosmetic)

This helper SHALL be called by `ListImages`, `ListTrashed`, `GetImage`, `Restore`, and `UpdateImage`.

#### Scenario: Thumbnail URL is presigned when thumbnail path exists

- **WHEN** an image has a non-nil `thumbnail_path`
- **THEN** the response includes a non-nil `thumbnail_url` containing a presigned GET URL

#### Scenario: Thumbnail URL is nil when no thumbnail exists

- **WHEN** an image has a nil `thumbnail_path`
- **THEN** `thumbnail_url` in the response is `null`

#### Scenario: Thumbnail URL is nil when presigning fails

- **WHEN** `GeneratePresignedGetURL` returns an error for the thumbnail key
- **THEN** `thumbnail_url` in the response is `null`
- **AND** the overall request succeeds

---

### Requirement: POST /images — Initiate Upload Request and Response

The `POST /images` handler SHALL accept an optional `description` field and an optional `folder_id` field in the request body.

When `folder_id` is provided, the usecase SHALL look up the folder by ID scoped to the authenticated user via `folderRepo.GetByID`. If the folder is not found, the image SHALL be created with `folder_id = null`. No error SHALL be returned to the caller in this case.

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

#### Scenario: Upload initiated with a valid folder_id

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that exists and belongs to the user
- **THEN** the response is `201 Created`
- **AND** the created image record has `folder_id` set to the provided value

#### Scenario: Upload initiated with a folder_id that does not exist

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that does not exist (or belongs to another user)
- **THEN** the response is `201 Created`
- **AND** the created image record has `folder_id` set to `null`

#### Scenario: Upload initiated with null or omitted folder_id

- **WHEN** an authenticated `POST /images` request omits `folder_id` or sets it to `null`
- **THEN** the response is `201 Created`
- **AND** the image record has `folder_id` as NULL

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

The `GET /images` endpoint SHALL return a paginated envelope (see `image-list-pagination` spec). The per-item `imageResponse` shape uses a presigned GET URL (24h TTL) for `thumbnail_url`:

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

`GET /images/:id` response (`imageDetailResponse`) includes a `thumbnail_url` field sourced from `ImageDetail.ThumbnailURL`, also a presigned GET URL (24h TTL).

#### Scenario: Image list response returns paginated envelope

- **WHEN** an authenticated `GET /images` request is made
- **THEN** the response is an object with an `images` array and a `next_cursor` field
- **AND** each item in `images` includes all existing fields (`description`, `width`, `height`, `file_size`, etc.)

#### Scenario: Image detail response is unchanged

- **WHEN** an authenticated `GET /images/:id` request is made for an existing image
- **THEN** the response shape is identical to the pre-pagination `imageDetailResponse`

---

### Requirement: GET /images/trash — Response Shape

The `GET /images/trash` endpoint SHALL return a paginated envelope using the same `imageResponse` shape as `GET /images`:

```json
{
  "images": [ /* array of imageResponse objects */ ],
  "next_cursor": "<opaque string | null>"
}
```

#### Scenario: Trash list response returns paginated envelope

- **WHEN** an authenticated `GET /images/trash` request is made
- **THEN** the response is an object with an `images` array and a `next_cursor` field
- **AND** each item in `images` has the same shape as items returned by `GET /images`

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

---

### Requirement: GET /images unfiled query parameter

The `GET /images` handler SHALL accept an optional `unfiled` boolean query parameter.

| `unfiled` value | Behaviour |
|---|---|
| Absent or `false` | Existing behaviour — no unfoldered filter applied |
| `true` | Returns only images where `folder_id IS NULL`; `folder_id` param is ignored |

`ListImagesParams` SHALL include an `Unfiled bool` field. When `Unfiled = true`, the repository SHALL emit `WHERE folder_id IS NULL` and ignore `FolderID`.

#### Scenario: unfiled=true returns only unfoldered images

- **WHEN** `GET /images?unfiled=true` is called
- **THEN** only images where `folder_id IS NULL` are returned

#### Scenario: unfiled=true ignores folder_id param

- **WHEN** `GET /images?unfiled=true&folder_id=<valid-uuid>` is called
- **THEN** only images where `folder_id IS NULL` are returned
- **AND** the `folder_id` param is not applied as a filter

#### Scenario: unfiled absent or false preserves existing behaviour

- **WHEN** `GET /images` is called without `unfiled` or with `unfiled=false`
- **THEN** existing folder filtering behaviour applies unchanged
