## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns non-deleted images ordered by `(created_at DESC, id DESC)`; fetches `limit + 1` rows so the caller can detect next-page existence; `folderID` nil means no folder filter; `tagID` nil means no tag filter; `unfiled` true limits to images with no entry in `image_folders`; images are returned with their `Tags` and `ImageFolders` preloaded
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only; result has `Tags` and `ImageFolders` preloaded
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only; result has `ImageFolders` preloaded
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `UpdateAILabels(ctx, id uuid.UUID, labels json.RawMessage) error`
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied scalar fields for the image owned by `userID`; `folder_id` is NOT a valid key in the map (folder assignment is handled by `SetImageFolder`); result has `Tags` and `ImageFolders` preloaded
- `SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error` — see `image-folders` spec for full behaviour
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns soft-deleted images ordered by `(deleted_at ASC, id ASC)`; fetches `limit + 1` rows; `cursor` nil means first page
- `CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)` — counts non-deleted images with a row in `image_folders` for the given folder; implemented as `Model(&domain.Image{}) + JOIN image_folders WHERE image_folders.folder_id = ?`
- `ListStaleUploads(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `HardDelete(ctx context.Context, id uuid.UUID, userID string) error`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

#### Scenario: List preloads tags and image folders for each image

- **WHEN** `List` is called and images have associated tags and folder memberships
- **THEN** each returned `domain.Image` has its `Tags` and `ImageFolders` slices populated

#### Scenario: List filters by folder via join

- **WHEN** `List` is called with a non-nil `folderID`
- **THEN** only images with a row in `image_folders` for that folder are returned
- **AND** the query uses `Model(&domain.Image{})` as the base so soft-deleted images are excluded automatically

#### Scenario: List unfiled uses left join

- **WHEN** `List` is called with `unfiled = true`
- **THEN** only images with no row in `image_folders` are returned
- **AND** the query uses a LEFT JOIN on `image_folders` and filters where the join produces no match

#### Scenario: GetByID preloads tags and image folders

- **WHEN** `GetByID` is called for an image that has tags and a folder membership
- **THEN** the returned `domain.Image` has its `Tags` and `ImageFolders` slices populated

#### Scenario: CountByFolderID excludes soft-deleted images

- **WHEN** a folder contains 3 images of which one is soft-deleted
- **THEN** `CountByFolderID` returns `2`

---

### Requirement: POST /images — Initiate Upload Request and Response

The `POST /images` handler SHALL accept an optional `description` field and an optional `folder_id` field in the request body.

When `folder_id` is provided, the usecase SHALL look up the folder by ID scoped to the authenticated user via `folderRepo.GetByID`. If the folder is not found, the image SHALL be created with no folder assignment. No error SHALL be returned to the caller in this case.

After the image record is created, if `folderID` is non-nil, the usecase SHALL call `imageRepo.SetImageFolder(ctx, created.ID, folderID)` to record the membership in `image_folders`.

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
- **AND** a row exists in `image_folders` for the created image and the provided folder

#### Scenario: Upload initiated with a folder_id that does not exist

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that does not exist (or belongs to another user)
- **THEN** the response is `201 Created`
- **AND** the created image has no row in `image_folders`

#### Scenario: Upload initiated with null or omitted folder_id

- **WHEN** an authenticated `POST /images` request omits `folder_id` or sets it to `null`
- **THEN** the response is `201 Created`
- **AND** the created image has no row in `image_folders`

---

### Requirement: PATCH /images/:id — UpdateImage Folder Assignment

When `PATCH /images/:id` includes a `folder_id` field, the usecase SHALL call `imageRepo.SetImageFolder` after completing the scalar field update, rather than including `folder_id` in the `fields` map passed to `imageRepo.Update`.

- If `params.FolderID` is nil (field absent from request): no call to `SetImageFolder`; existing folder membership is unchanged
- If `params.FolderID` is `&nil` (field present as JSON `null`): call `SetImageFolder(ctx, id, nil)` to remove folder membership
- If `params.FolderID` is `&&folderUUID` (field present as a UUID): call `SetImageFolder(ctx, id, &folderUUID)` to assign the folder

#### Scenario: PATCH with folder_id assigns folder via SetImageFolder

- **WHEN** `PATCH /images/:id` is sent with a valid `folder_id`
- **THEN** `imageRepo.SetImageFolder` is called with the image ID and the provided folder UUID
- **AND** a row exists in `image_folders` for the image and folder after the update

#### Scenario: PATCH with folder_id null removes folder membership

- **WHEN** `PATCH /images/:id` is sent with `"folder_id": null`
- **THEN** `imageRepo.SetImageFolder` is called with `folderID = nil`
- **AND** the image has no row in `image_folders` after the update

#### Scenario: PATCH without folder_id leaves folder membership unchanged

- **WHEN** `PATCH /images/:id` is sent without a `folder_id` field
- **THEN** `imageRepo.SetImageFolder` is NOT called
- **AND** the image's existing folder membership is unchanged

---

### Requirement: AcceptSuggestion Uses SetImageFolder

The `AcceptSuggestion` usecase method SHALL call `imageRepo.SetImageFolder(ctx, imageID, &folder.ID)` to assign the suggested folder, instead of using `imageRepo.Update` with a `folder_id` key.

#### Scenario: AcceptSuggestion assigns folder via SetImageFolder

- **WHEN** `AcceptSuggestion` is called with a valid suggestion
- **THEN** a row is inserted into `image_folders` for the image and the resolved folder
- **AND** `imageRepo.Update` is NOT called with a `folder_id` key

---

### Requirement: GET /images and GET /images/:id — folder_id in Response

The `folder_id` field in `imageResponse` and `imageDetailResponse` SHALL be populated from the image's `ImageFolders` slice. If `ImageFolders` is non-empty, `folder_id` SHALL be the `FolderID` of the first entry. If `ImageFolders` is empty, `folder_id` SHALL be `null`.

`toImageResponse` SHALL read `item.Image.ImageFolders[0].FolderID` when the slice is non-empty.

The response field name, type, and nullability are unchanged from the current API contract.

#### Scenario: Image response includes folder_id from ImageFolders

- **WHEN** `GET /images` or `GET /images/:id` is called for an image with a folder membership
- **THEN** the response includes a non-null `folder_id` matching the folder the image is in

#### Scenario: Image response has null folder_id for unfiled image

- **WHEN** `GET /images` or `GET /images/:id` is called for an image with no folder membership
- **THEN** `folder_id` in the response is `null`
