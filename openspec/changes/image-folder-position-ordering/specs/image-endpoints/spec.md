## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — when `folderID` is non-nil: returns all non-deleted images for that folder ordered by `image_folders.position ASC`; `cursor` and `limit` are ignored; images are returned with `Tags` and `ImageFolders` preloaded. When `folderID` is nil: returns non-deleted images ordered by `(created_at DESC, id DESC)`; fetches `limit + 1` rows; `cursor` applies a keyset filter; `unfiled` true limits to images with no entry in `image_folders`; `tagID` non-nil filters by tag.
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only; result has `Tags` and `ImageFolders` preloaded
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only; result has `ImageFolders` preloaded
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `UpdateAILabels(ctx, id uuid.UUID, labels json.RawMessage) error`
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied scalar fields for the image owned by `userID`; `folder_id` is NOT a valid key in the map (folder assignment is handled by `SetImageFolder`); result has `Tags` and `ImageFolders` preloaded
- `SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error` — see `image-folders` spec for full behaviour
- `UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error` — see `image-folders` spec for full behaviour
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns soft-deleted images ordered by `(deleted_at ASC, id ASC)`; fetches `limit + 1` rows; `cursor` nil means first page
- `CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)` — counts non-deleted images with a row in `image_folders` for the given folder
- `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `HardDelete(ctx context.Context, id uuid.UUID, userID string) error`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

#### Scenario: List orders by position for folder views

- **WHEN** `List` is called with a non-nil `folderID`
- **THEN** results are ordered by `image_folders.position ASC`
- **AND** cursor and limit parameters have no effect

#### Scenario: List preloads tags and image folders for each image

- **WHEN** `List` is called and images have associated tags and folder memberships
- **THEN** each returned `domain.Image` has its `Tags` and `ImageFolders` slices populated

---

### Requirement: GET /images endpoint — image response shape

The per-item `imageResponse` shape for `GET /images` SHALL include a `position` field:

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string|null",
  "mime_type": "string",
  "source_url": "string|null",
  "folder_id": "uuid|null",
  "position": "string|null",
  "thumbnail_url": "string|null",
  "width": "int|null",
  "height": "int|null",
  "file_size": "int|null",
  "tags": [{ "id": "uuid", "name": "string" }],
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

- `position` SHALL be populated from `ImageFolders[0].Position` when the image has a folder membership
- `position` SHALL be `null` when the image has no folder membership (all/unfiled views)

The `toImageResponse` function in `internal/handler/image.go` SHALL set `Position` from the first `ImageFolder` entry when present.

#### Scenario: Folder view response includes position

- **WHEN** `GET /images?folder_id=<id>` is called
- **THEN** each image in the response has a non-null `position` field containing its fracdex key

#### Scenario: Non-folder view response has null position

- **WHEN** `GET /images` is called without `folder_id` (all or unfiled view)
- **THEN** `position` is `null` for images with no folder membership

---

### Requirement: GET /images and GET /images/:id — folder_id in Response

The `folder_id` field in `imageResponse` and `imageDetailResponse` SHALL be populated from the image's `ImageFolders` slice. If `ImageFolders` is non-empty, `folder_id` SHALL be the `FolderID` of the first entry. If `ImageFolders` is empty, `folder_id` SHALL be `null`.

The `position` field follows the same pattern: populated from `ImageFolders[0].Position` if present, `null` otherwise.

#### Scenario: Image response includes folder_id from ImageFolders

- **WHEN** an image has a folder membership
- **THEN** the response includes a non-null `folder_id` matching the folder the image is in

#### Scenario: Image response has null folder_id for unfiled image

- **WHEN** an image has no folder membership
- **THEN** `folder_id` in the response is `null`

#### Scenario: Image response includes position from ImageFolders

- **WHEN** an image has a folder membership
- **THEN** the response includes a non-null `position` matching the fracdex key in `image_folders`

---

### Requirement: PATCH /images/:id/position Route

The system SHALL register `PATCH /images/:id/position` on the protected router, handled by `imageHandler.UpdateImagePosition`.

This route SHALL be registered before `PATCH /images/:id` to avoid Echo treating `position` as an `:id` segment.

#### Scenario: Route is reachable

- **WHEN** `PATCH /images/:id/position` is called with a valid token and body
- **THEN** the request is routed to `UpdateImagePosition` handler (not `UpdateImage`)
