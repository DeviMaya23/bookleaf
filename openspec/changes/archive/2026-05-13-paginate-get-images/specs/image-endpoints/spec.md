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

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations. `ListImages` and `ListTrashed` SHALL use the paginated signatures:

```go
ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error)
ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error)
```

All other method signatures are unchanged.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors

---

### Requirement: GET /images and GET /images/:id — Response Shape

The `GET /images` endpoint SHALL return a paginated envelope (see `image-list-pagination` spec). The per-item `imageResponse` shape is unchanged:

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

`GET /images/:id` response (`imageDetailResponse`) is unchanged.

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
