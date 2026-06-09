### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/image_repository.go` containing only the methods used by `imageUsecase` and the upload transaction callback. Per the conventions, each usecase defines its own interface for its dependencies — trash-related and upload-related methods live on `TrashRepository` and `UploadImageRepository` respectively.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name *string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns non-deleted images for `userID`; fetches `limit + 1` rows; `cursor` applies a keyset filter on the active sort column; images are returned with `Tags` and `ImageFolders` preloaded. `unfiled` true limits to images with no entry in `image_folders` (LEFT JOIN + `IS NULL`, as before). `folderIDs` non-empty filters to images belonging to ANY of the given folders via a correlated `EXISTS` subquery against `image_folders`, so an image matching multiple supplied folder IDs is still returned at most once. `tagIDs` non-empty filters to images associated with ANY of the given tags via a correlated `EXISTS` subquery against `image_tags`, with the same at-most-once guarantee. `mimeTypes` non-empty filters to images whose `mime_type` matches any of the given values via `IN`. `name` non-nil and non-empty filters to images whose `title` contains the value, case-insensitively. All of `unfiled`/`folderIDs`/`tagIDs`/`mimeTypes`/`name` compose via `AND`; no validation rejects contradictory combinations (e.g. `unfiled=true` together with non-empty `folderIDs`) — such combinations simply yield an empty result, the same way an impossible `name`/`tagIDs` combination would (see `GET /images Multi-Value Filter Query Parameters`). `sortField`/`direction` select the ordering: when `sortField` is nil, results order by `created_at DESC, id DESC` (today's default, unchanged); when `sortField` is non-nil, results order by the selected column (with `id` as tiebreaker) in the selected direction (the field's default direction applies when `direction` is nil), and the keyset filter compares against that same column instead of `created_at` (see `image-list-pagination` for the keyset comparison rules). The single-folder, position-ordered, unpaginated "folder contents" query that this method previously served when `folderID` was non-nil is REMOVED from this method — it is now served by a dedicated method described in `folder-image-listing`.
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only; result has `Tags` and `ImageFolders` preloaded
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied scalar fields for the image owned by `userID`; `folder_id` is NOT a valid key in the map (folder assignment is handled by `SetImageFolder`); result has `Tags` and `ImageFolders` preloaded
- `SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error` — see `image-folders` spec for full behaviour
- `SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error` — diffs current memberships against folderIDs and applies deletes/inserts in a transaction
- `MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error` — atomically removes image from fromFolderID and adds to toFolderID
- `UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error` — see `image-folders` spec for full behaviour

`ListStaleUploads` is REMOVED from this interface. Stale upload detection is now handled by `PendingUploadRepository.ListStale`.

`UpdateThumbnailPath` and `UpdateAILabels` are on `UploadImageRepository` (defined in `image_upload_usecase.go`), not `ImageRepository`.

`CountByFolderID` is on `ImageCounter` (defined in `folder_usecase.go`), not `ImageRepository`.

Trash lifecycle methods (`GetDeletedByID`, `SoftDelete`, `Restore`, `ListTrashed`, `ListAllTrashed`, `ListExpiredTrash`, `HardDelete`) are on `TrashRepository` (defined in `trash_repository.go`), not `ImageRepository`.

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

#### Scenario: List defaults to created_at-descending ordering when sort is nil

- **WHEN** `List` is called with a nil `sortField`
- **THEN** results are ordered by `created_at DESC, id DESC`
- **AND** the keyset filter, when a cursor is present, compares `(created_at, id)`

#### Scenario: List honors an explicit sort field

- **WHEN** `List` is called with `sortField = "title"`
- **THEN** results are ordered by `title` (in the requested or field-default direction) with `id` as tiebreaker
- **AND** the keyset filter, when a cursor is present, compares `(title, id)` using the operator matching the sort direction

#### Scenario: List preloads tags and image folders for each image

- **WHEN** `List` is called and images have associated tags and folder memberships
- **THEN** each returned `domain.Image` has its `Tags` and `ImageFolders` slices populated

#### Scenario: List unfiled uses left join

- **WHEN** `List` is called with `unfiled = true`
- **THEN** only images with no row in `image_folders` are returned
- **AND** the query uses a LEFT JOIN on `image_folders` and filters where the join produces no match

#### Scenario: GetByID preloads tags and image folders

- **WHEN** `GetByID` is called for an image that has tags and a folder membership
- **THEN** the returned `domain.Image` has its `Tags` and `ImageFolders` slices populated

#### Scenario: List filters by folderIDs when provided (match-any)

- **WHEN** `List` is called with a non-empty `folderIDs` slice containing two folder UUIDs
- **THEN** only images belonging to at least one of those folders are returned
- **AND** an image belonging to both supplied folders appears exactly once in the results

#### Scenario: List filters by tagIDs when provided (match-any)

- **WHEN** `List` is called with a non-empty `tagIDs` slice containing two tag UUIDs
- **THEN** only images associated with at least one of those tags are returned
- **AND** an image associated with both supplied tags appears exactly once in the results

#### Scenario: List filters by mimeTypes when provided (match-any)

- **WHEN** `List` is called with a non-empty `mimeTypes` slice, e.g. `["image/jpeg", "image/png"]`
- **THEN** only images whose `mime_type` equals one of the supplied values are returned

#### Scenario: List filters by name

- **WHEN** `List` is called with a non-nil, non-empty `name`
- **THEN** the query includes a case-insensitive substring match (`ILIKE '%<name>%'`) against `images.title`
- **AND** this filter composes with any `unfiled`, `folderIDs`, `tagIDs`, `mimeTypes`, sort, and cursor conditions already present

---

### Requirement: Trash Repository Interface

The system SHALL define a `TrashRepository` interface in `internal/usecase/trash_repository.go` containing only the methods used by `trashUsecase`. Per the conventions, `trashUsecase` defines its own interface for its repository dependency rather than depending on the broader `ImageRepository`.

Methods:
- `GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only
- `GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `SoftDelete(ctx context.Context, id uuid.UUID, userID string) error`
- `Restore(ctx context.Context, id uuid.UUID, userID string) error`
- `ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns soft-deleted images ordered by `(deleted_at ASC, id ASC)`; fetches `limit + 1` rows; `cursor` nil means first page
- `ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error)` — returns all soft-deleted images for the user with no pagination
- `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `HardDelete(ctx context.Context, id uuid.UUID, userID string) error`

The concrete `*imageRepository` in `internal/repository/` satisfies both `ImageRepository` and `TrashRepository`.

#### Scenario: TrashRepository is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.TrashRepository` without compilation errors

---

### Requirement: Image Routes Wiring

The system SHALL register image routes on the protected Echo group in `main.go`.

Routes:
- `POST /images`
- `POST /images/:id/complete`
- `POST /images/:id/accept-suggestion`
- `GET /images`
- `GET /images/:id`
- `GET /images/in-folder/:id`
- `PATCH /images/:id/position`
- `PATCH /images/:id`
- `DELETE /images/:id`
- `GET /images/trash`
- `POST /images/:id/restore`

`GET /images/in-folder/:id` SHALL be registered such that it does not collide with `GET /images/:id` — its three-segment path (`/images/in-folder/:id`) is distinguishable from the two-segment `/images/:id` by Echo's router regardless of registration order, but it SHALL be registered alongside the other `/images/*` routes for readability. See `folder-image-listing` for its handler behaviour.

#### Scenario: Image routes are registered under auth middleware

- **WHEN** the server starts
- **THEN** all `/images` routes require a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`

#### Scenario: GET /images/in-folder/:id does not collide with GET /images/:id

- **WHEN** a request is made to `GET /images/in-folder/<folder-uuid>`
- **THEN** it is routed to the folder-image-listing handler, not to `GetImage` with `in-folder` interpreted as an image ID

---

### Requirement: PATCH /images/:id/position Route

The system SHALL register `PATCH /images/:id/position` on the protected router, handled by `imageHandler.UpdateImagePosition`.

This route SHALL be registered before `PATCH /images/:id` to avoid Echo treating `position` as an `:id` segment.

#### Scenario: Route is reachable

- **WHEN** `PATCH /images/:id/position` is called with a valid token and body
- **THEN** the request is routed to `UpdateImagePosition` handler (not `UpdateImage`)

---

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to image querying and editing operations. The `CompleteUpload` method SHALL return a result struct alongside the error.

```go
CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error)
AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error
```

`InitiateUpload` SHALL accept a `description *string` parameter:

```go
InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*UploadInitResult, error)
```

`UpdateImageParams` SHALL include `Description`, `SourceURL`, and `Tags` fields:

```go
type UpdateImageParams struct {
    Title       *string
    FolderID    **uuid.UUID
    Description *string
    SourceURL   **string
    Tags        *[]uuid.UUID  // nil = no change; non-nil (including empty slice) = replace tag set
}
```

`CompleteUploadResult` is defined in `internal/usecase/`. The `SuggestedFolderName` and `Warning` fields are removed; vision labelling is now asynchronous:

```go
type CompleteUploadResult struct {
    ImageID uuid.UUID
}
```

`ListImagesParams` SHALL include `Unfiled`, `FolderIDs`, `TagIDs`, `MIMETypes`, `Name`, `Sort`, and `Direction` fields:

```go
type ListImagesParams struct {
    Unfiled   bool
    FolderIDs []uuid.UUID
    TagIDs    []uuid.UUID
    MIMETypes []string
    Name      *string
    Sort      *string
    Direction *string
    Cursor    *ImageCursor
    Limit     int
}
```

The single-value `FolderID *uuid.UUID` and `TagID *uuid.UUID` fields are REMOVED from `ListImagesParams`. There is no longer a folder-view mode reachable through `ListImages`/`ListImagesParams` — fetching a single folder's contents in custom order is handled by a dedicated method described in `folder-image-listing`.

`Name`, when non-nil and non-empty, filters results to images whose title contains the value, case-insensitively.

`FolderIDs`, `TagIDs`, and `MIMETypes`, when non-empty, filter results to images matching ANY of the supplied values for that dimension (match-any); see `GET /images Multi-Value Filter Query Parameters` for the full contract.

`Sort`, when non-nil, selects the ordering column (`created_at` or `title`); `Direction`, when non-nil, selects `asc` or `desc`. Both are validated against an allow-list by the handler before reaching the usecase (see `GET /images sort and direction query parameters`) — the usecase passes them through to the repository unchanged, performing no additional validation or defaulting of its own. When `Sort` is nil, ordering follows the default described in the `ImageRepository.List` requirement (`created_at DESC`).

`ListImages` SHALL use the paginated signature:

```go
ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error)
```

`UpdateImage` SHALL return `*ImageItem`:

```go
UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*ImageItem, error)
```

`ImageItem` is defined in `internal/usecase/`:

```go
type ImageItem struct {
    Image        *domain.Image
    ThumbnailURL *string
}
```

`ListImagesResult.Images` SHALL be `[]ImageItem`.

`ImageDetail` SHALL include a `ThumbnailURL *string` field alongside `ImageURL`:

```go
type ImageDetail struct {
    Image        *domain.Image
    ImageURL     string
    ThumbnailURL *string
}
```

All other method signatures are unchanged.

When `UpdateImageParams.Tags` is non-nil, `UpdateImage` SHALL call `tagRepo.ReplaceImageTags` with the image ID and the given tag IDs after the scalar field update.

Trash lifecycle methods (`SoftDelete`, `ListTrashed`, `Restore`, `DeleteFromTrash`, `EmptyTrash`, `PurgeExpiredTrash`) are NOT part of `ImageUsecase`. They belong to `TrashUsecase` (defined in `internal/usecase/trash_usecase.go`).

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors

#### Scenario: ImageUsecase interface contains no trash methods

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` in `internal/usecase/` does not define or implement `SoftDelete`, `ListTrashed`, `Restore`, `DeleteFromTrash`, `EmptyTrash`, or `PurgeExpiredTrash`

#### Scenario: UpdateImage replaces tags when Tags param is non-nil

- **WHEN** `UpdateImage` is called with `Tags` set to a slice of tag UUIDs
- **THEN** the image's tag associations are replaced with that set

#### Scenario: UpdateImage does not touch tags when Tags param is nil

- **WHEN** `UpdateImage` is called without a `Tags` param (nil)
- **THEN** the image's existing tag associations are unchanged

#### Scenario: UpdateImage clears all tags when Tags is an empty slice

- **WHEN** `UpdateImage` is called with `Tags` set to an empty slice
- **THEN** all tag associations for the image are removed

#### Scenario: ListImages passes FolderIDs, TagIDs, and MIMETypes to repository

- **WHEN** `ListImages` is called with non-empty `FolderIDs`, `TagIDs`, and `MIMETypes` params
- **THEN** the repository's `List` is invoked with those exact slices as `folderIDs`/`tagIDs`/`mimeTypes`
- **AND** only images matching the composed filter (AND across dimensions, match-any within each) are returned

#### Scenario: ListImages passes Name to repository

- **WHEN** `ListImages` is called with a non-nil, non-empty `Name` param
- **THEN** the repository is called with that value as the `name` filter
- **AND** only images whose title contains the value, case-insensitively, are returned

#### Scenario: ListImages ignores blank Name

- **WHEN** `ListImages` is called with `Name` pointing to an empty string
- **THEN** the name filter is not applied and results are unaffected

#### Scenario: ListImages passes Sort and Direction through to the repository unchanged

- **WHEN** `ListImages` is called with non-nil `Sort` and `Direction` params
- **THEN** the repository's `List` is invoked with those exact values as `sortField`/`direction`, with no further validation, normalization, or defaulting performed by the usecase

#### Scenario: ListImages passes nil Sort and Direction through as nil

- **WHEN** `ListImages` is called with `Sort` and `Direction` both nil
- **THEN** the repository's `List` is invoked with nil `sortField`/`direction`, preserving the default ordering

---

### Requirement: Thumbnail URL Generation

The usecase SHALL generate presigned GET URLs for thumbnails with a 24h TTL. A private helper `thumbnailURL(ctx context.Context, path *string) *string` on `imageUsecase` SHALL:

- Return `nil` if `path` is `nil`
- Call `store.GeneratePresignedGetURL` with `presignedGetTTL` (24h)
- Return `nil` if presigning fails (non-fatal; thumbnail is cosmetic)

This helper SHALL be called by `ListImages`, `GetImage`, and `UpdateImage` on `imageUsecase`. `trashUsecase` has its own equivalent helper used by `ListTrashed` and `Restore`.

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

When `folder_id` is provided, the usecase SHALL look up the folder by ID scoped to the authenticated user via `folderRepo.GetByID`. If the folder is not found, `folder_id` SHALL be stored as `nil` in the `pending_uploads` record. No error SHALL be returned to the caller in this case.

The usecase SHALL create a row in `pending_uploads` (not `images`) with all initiation-time metadata including `folder_id`. No row is written to `image_folders` at this point; folder assignment is deferred to `CompleteUpload`.

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

Response body (201): `id`, `upload_url`, `r2_path`. The `id` in the response is the `pending_uploads.id`, which will become `images.id` upon commit.

#### Scenario: Upload initiated with a valid folder_id

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that exists and belongs to the user
- **THEN** the response is `201 Created`
- **AND** a row exists in `pending_uploads` with the provided `folder_id`
- **AND** no row exists yet in `image_folders` (folder assignment is deferred to CompleteUpload)

#### Scenario: Upload initiated with a folder_id that does not exist

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that does not exist (or belongs to another user)
- **THEN** the response is `201 Created`
- **AND** the `pending_uploads` row has `folder_id = NULL`

#### Scenario: Upload initiated with null or omitted folder_id

- **WHEN** an authenticated `POST /images` request omits `folder_id` or sets it to `null`
- **THEN** the response is `201 Created`
- **AND** the `pending_uploads` row has `folder_id = NULL`

#### Scenario: Upload initiated with description

- **WHEN** an authenticated `POST /images` request includes a non-empty `description`
- **THEN** the response is `201 Created`
- **AND** the `pending_uploads` row has the supplied `description` value persisted

#### Scenario: Upload initiated without description

- **WHEN** an authenticated `POST /images` request omits `description`
- **THEN** the response is `201 Created`
- **AND** the `pending_uploads` row has `description` as NULL

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
  "folder_ids": ["uuid"],
  "thumbnail_url": "string|null",
  "width": "integer|null",
  "height": "integer|null",
  "file_size": "integer|null",
  "tags": [{ "id": "uuid", "name": "string" }],
  "position": "string|null",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

- `folder_ids` SHALL be a non-null array of UUIDs — empty (`[]`) when the image has no folder memberships, populated with all folder IDs the image belongs to otherwise
- `position` SHALL always be `null` in `GET /images` and `GET /images/:id` results. `GET /images` no longer has a folder-scoped, position-ordered mode — that query is served exclusively by `GET /images/in-folder/:id` (see `folder-image-listing`), whose response populates `position` from `image_folders.position` for the queried folder
- `GET /images/:id` (`imageDetailResponse`) follows the same shape and includes an additional `image_url` field; `position` is always `null` in the detail response (no folder context)

The `ImageItem` struct in `internal/usecase/image_usecase.go` retains its `FolderPosition *string` field, but `ListImages` SHALL always leave it `nil` — the gallery query has no single-folder context to derive a position from. The dedicated folder-listing usecase method described in `folder-image-listing` is the only populator of `FolderPosition`. The `toImageResponse` function in `internal/handler/image.go` SHALL continue to map `item.FolderPosition` to `Position` on the response struct (which will be `nil`/`null` for all `GET /images`/`GET /images/:id` results).

#### Scenario: Image list response returns paginated envelope

- **WHEN** an authenticated `GET /images` request is made
- **THEN** the response is an object with an `images` array and a `next_cursor` field
- **AND** each item in `images` includes a `folder_ids` array (never null)

#### Scenario: GET /images always returns null position

- **WHEN** an authenticated `GET /images` request is made, with or without `folder_ids`/`tag_ids`/`mime_types`/`unfiled`/`name` filters
- **THEN** every image in the response has `position: null`

#### Scenario: Image detail response includes folder_ids array

- **WHEN** an authenticated `GET /images/:id` request is made for an existing image
- **THEN** the response includes a `folder_ids` field containing all folder UUIDs the image belongs to

#### Scenario: folder_ids is empty for an unfiled image

- **WHEN** `GET /images/:id` is called for an image with no folder memberships
- **THEN** `folder_ids` in the response is an empty array `[]`

#### Scenario: folder_ids contains all memberships for a multi-folder image

- **WHEN** `GET /images/:id` is called for an image belonging to two folders
- **THEN** `folder_ids` contains both folder UUIDs

---

### Requirement: GET /images and GET /images/:id — folder_ids in Response

The `folder_ids` field in `imageResponse` and `imageDetailResponse` SHALL be populated from the image's `ImageFolders` slice by iterating all entries and collecting their `FolderID` values. The field SHALL never be null — an image with no folder memberships returns an empty array `[]`.

The `folder_id` (singular) field is removed from `imageResponse` and `imageDetailResponse`. The `position` field remains in the response shape but is always `null` for `GET /images` and `GET /images/:id` — position is only populated by `GET /images/in-folder/:id` (see `folder-image-listing`). The `ImageRepository` interface, `toImageResponse` helper, and all related handler structs SHALL reflect this change.

`toImageResponse` SHALL populate `FolderIDs []uuid.UUID` from `item.Image.ImageFolders` (all entries, not just index 0). The `firstFolderID` and `firstFolderPosition` helpers are removed.

#### Scenario: Image response includes folder_ids from ImageFolders

- **WHEN** an image has one or more folder memberships
- **THEN** the response includes a non-empty `folder_ids` array containing all folder UUIDs

#### Scenario: Image response has empty folder_ids for unfiled image

- **WHEN** an image has no folder membership
- **THEN** `folder_ids` in the response is `[]`

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

### Requirement: DELETE /images/:id — Soft Delete

The system SHALL expose a `DELETE /images/:id` endpoint on the protected route group that soft-deletes an image by setting `deleted_at`.

Response: `204 No Content`

- The image MUST be owned by the authenticated user
- Returns `404 Not Found` if the image does not exist or belongs to another user
- The image is NOT removed from R2; only the `deleted_at` timestamp is set

#### Scenario: Authenticated user soft-deletes an image

- **WHEN** an authenticated `DELETE /images/:id` request is made for an owned non-deleted image
- **THEN** the response is `204 No Content`
- **AND** the image has `deleted_at` set in the database
- **AND** the image no longer appears in `GET /images` results

#### Scenario: Image not found or not owned by user

- **WHEN** an authenticated `DELETE /images/:id` request is made for a non-existent or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `DELETE /images/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: POST /images/:id/restore — Restore from Trash

The system SHALL expose a `POST /images/:id/restore` endpoint on the protected route group that restores a soft-deleted image by clearing `deleted_at`.

Response body (200): the restored image in the same `imageResponse` shape as `GET /images` list items, including a presigned `thumbnail_url` (24h TTL).

- The image MUST be soft-deleted and owned by the authenticated user
- Returns `404 Not Found` if the image does not exist, is not soft-deleted, or belongs to another user

#### Scenario: Authenticated user restores a trashed image

- **WHEN** an authenticated `POST /images/:id/restore` request is made for a soft-deleted owned image
- **THEN** the response is `200 OK`
- **AND** `deleted_at` is cleared in the database
- **AND** the response body includes the restored image with all `imageResponse` fields
- **AND** the image appears again in `GET /images` results

#### Scenario: Image not found, not deleted, or not owned

- **WHEN** a `POST /images/:id/restore` request is made for an image that is not soft-deleted or does not belong to the user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `POST /images/:id/restore` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

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

### Requirement: Image Response Types Include Tags

The system SHALL include tag data in image response types returned by the image handler.

```go
type tagResponse struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
}

type imageResponse struct {
    ID           uuid.UUID     `json:"id"`
    Title        string        `json:"title"`
    Description  *string       `json:"description"`
    MIMEType     string        `json:"mime_type"`
    SourceURL    *string       `json:"source_url"`
    FolderIDs    []uuid.UUID   `json:"folder_ids"`
    Position     *string       `json:"position"`
    ThumbnailURL *string       `json:"thumbnail_url"`
    Width        *int          `json:"width"`
    Height       *int          `json:"height"`
    FileSize     *int64        `json:"file_size"`
    Tags         []tagResponse `json:"tags"`
    CreatedAt    time.Time     `json:"created_at"`
    UpdatedAt    time.Time     `json:"updated_at"`
}

type imageDetailResponse struct {
    // all fields of imageResponse, plus:
    ImageURL string `json:"image_url"`
}
```

`toImageResponse` SHALL map `item.Image.Tags` to `[]tagResponse`. If `Tags` is nil, it SHALL return an empty slice (never `null` in JSON).

#### Scenario: ListImages response includes tags array

- **WHEN** `GET /images` is called and images have associated tags
- **THEN** each image object in the response contains a `tags` array with `id` and `name` for each tag

#### Scenario: ListImages response has empty tags array for untagged images

- **WHEN** `GET /images` is called and an image has no tags
- **THEN** the image object contains `"tags": []`

#### Scenario: GetImage response includes tags

- **WHEN** `GET /images/:id` is called for an image with tags
- **THEN** the response contains a `tags` array with `id` and `name` for each associated tag

---

### Requirement: UpdateImage Accepts Tags Field

The system SHALL accept an optional `tags` field in `PATCH /images/:id` request body to replace the image's tag set.

```go
type updateImageRequest struct {
    Title       *string         `json:"title"`
    Description *string         `json:"description"`
    FolderID    json.RawMessage `json:"folder_id"`
    SourceURL   json.RawMessage `json:"source_url"`
    Tags        json.RawMessage `json:"tags"`
}
```

Parsing rules for `Tags`:
- Field absent or JSON `null` → `params.Tags` is nil (no change to tags)
- `[]` → `params.Tags` is a pointer to an empty slice (clear all tags)
- `["uuid1", "uuid2"]` → `params.Tags` is a pointer to the parsed UUID slice

The handler SHALL return `400 Bad Request` if a tag ID in the array is not a valid UUID.

#### Scenario: PATCH with tags replaces image tag set

- **WHEN** `PATCH /images/:id` is sent with `{"tags": ["<uuid>"]}`
- **THEN** the image's tags are replaced with the specified tag
- **AND** the response body contains the updated tags array

#### Scenario: PATCH without tags does not modify tags

- **WHEN** `PATCH /images/:id` is sent without a `tags` field
- **THEN** the image's existing tags are unchanged

#### Scenario: PATCH with empty tags array clears all tags

- **WHEN** `PATCH /images/:id` is sent with `{"tags": []}`
- **THEN** all tags are removed from the image

#### Scenario: PATCH with invalid tag UUID returns 400

- **WHEN** `PATCH /images/:id` is sent with `{"tags": ["not-a-uuid"]}`
- **THEN** the response is `400 Bad Request`

---

### Requirement: GET /images Multi-Value Filter Query Parameters

The `GET /images` handler SHALL accept optional `folder_ids`, `tag_ids`, and `mime_types` query parameters, each a comma-separated list of values, providing independent match-any filters that compose with each other and with `name`/`unfiled`/sort/pagination via `AND`.

| Parameter | Value type | Matches images that... |
|---|---|---|
| `folder_ids` | comma-separated UUIDs | belong to ANY of the given folders |
| `tag_ids` | comma-separated UUIDs | are associated with ANY of the given tags |
| `mime_types` | comma-separated strings | have a `mime_type` equal to ANY of the given values |

Rules:
- Each parameter is optional; absent or empty means "no filter on that dimension"
- Values are split on `,`; empty segments (e.g. from a trailing comma) SHALL be ignored
- For `folder_ids` and `tag_ids`, each non-empty segment SHALL be parsed as a UUID; if any segment fails to parse, the handler SHALL return `400 Bad Request` (mirroring the existing single-value validation style for `folder_id`/`tag_id`)
- For `mime_types`, segments are passed through as plain strings with no format validation beyond non-emptiness
- The parsed slices are passed to `imageUsecase.ListImages` via `ListImagesParams.FolderIDs`/`TagIDs`/`MIMETypes`; an empty or absent parameter results in a nil/empty slice, which the repository treats as "no filter on that dimension"
- Multiple filters compose via `AND` (an image must satisfy every supplied filter dimension); multiple values within one filter compose via match-any/`OR` (an image must satisfy at least one value in that dimension)
- No cross-filter validation is performed: contradictory combinations (e.g. `unfiled=true&folder_ids=...`) are not rejected and simply produce an empty result (see `GET /images unfiled query parameter`)
- `folder_id` and `tag_id` (singular) are no longer accepted; supplying them has no effect (they are simply unrecognized query parameters)

#### Scenario: GET /images with folder_ids returns images in any of the given folders

- **WHEN** `GET /images?folder_ids=<uuid-a>,<uuid-b>` is called
- **THEN** only images belonging to folder A or folder B (or both) are returned
- **AND** an image belonging to both folders appears exactly once in the results

#### Scenario: GET /images with tag_ids returns images with any of the given tags

- **WHEN** `GET /images?tag_ids=<uuid-a>,<uuid-b>` is called
- **THEN** only images associated with tag A or tag B (or both) are returned
- **AND** an image associated with both tags appears exactly once in the results

#### Scenario: GET /images with mime_types returns images of any of the given types

- **WHEN** `GET /images?mime_types=image/jpeg,image/png` is called
- **THEN** only images whose `mime_type` is `image/jpeg` or `image/png` are returned

#### Scenario: GET /images composes multiple filter dimensions with AND

- **WHEN** `GET /images?folder_ids=<uuid-a>&tag_ids=<uuid-x>,<uuid-y>` is called
- **THEN** only images that belong to folder A AND are associated with tag X or tag Y are returned

#### Scenario: GET /images with invalid UUID in folder_ids or tag_ids returns 400

- **WHEN** `GET /images?folder_ids=not-a-uuid` or `GET /images?tag_ids=<uuid>,not-a-uuid` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: GET /images ignores empty segments in multi-value parameters

- **WHEN** `GET /images?tag_ids=<uuid>,` is called (trailing comma)
- **THEN** the request is treated as if only the single valid UUID were supplied — no `400` is returned for the empty segment

#### Scenario: GET /images without multi-value filter parameters returns unfiltered-by-those-dimensions results

- **WHEN** `GET /images` is called without `folder_ids`, `tag_ids`, or `mime_types`
- **THEN** results are not filtered by folder membership, tags, or MIME type (other active filters still apply)

#### Scenario: GET /images ignores singular folder_id and tag_id parameters

- **WHEN** `GET /images?folder_id=<uuid>` or `GET /images?tag_id=<uuid>` is called
- **THEN** the parameter has no effect on the results (it is not recognized; use `folder_ids`/`tag_ids` instead)

---

### Requirement: CompleteUpload Populates Dimensions and File Size

The `POST /images/:id/complete` handler SHALL accept an optional JSON request body containing client-supplied `width`, `height`, and `file_size` (integers), and the usecase SHALL persist them on the `domain.Image` record being created — instead of deriving them server-side from the uploaded bytes.

- The request body is optional; any of `width`, `height`, `file_size` MAY be omitted
- A supplied value SHALL be persisted only if it is a positive integer (`> 0`); otherwise the corresponding field SHALL be stored as `NULL`
- No image bytes SHALL be fetched from R2 or decoded by the backend to derive these fields (the prior `extractImageMetadata` decode path, including the `image.DecodeConfig`/`image/jpeg`/`image/png` stdlib usage, is removed)
- The `CompleteUpload` response body is unchanged

#### Scenario: Client-supplied dimensions and size are persisted

- **WHEN** `CompleteUpload` is called with a request body of `{ "width": 1920, "height": 1080, "file_size": 245760 }`
- **THEN** the image record is created with `width = 1920`, `height = 1080`, and `file_size = 245760`

#### Scenario: Implausible values are stored as NULL

- **WHEN** `CompleteUpload` is called with a request body containing a non-positive value, e.g. `{ "width": -1, "height": 0, "file_size": 1024 }`
- **THEN** the image record is created with `width = NULL` and `height = NULL`
- **AND** `file_size = 1024` is persisted
- **AND** the upload completes successfully (no error is returned)

#### Scenario: Omitted fields are stored as NULL

- **WHEN** `CompleteUpload` is called with no request body (or a body omitting `width`, `height`, and `file_size`)
- **THEN** the image record is created with `width = NULL`, `height = NULL`, and `file_size = NULL`
- **AND** the upload completes successfully (no error is returned)

---

### Requirement: CompleteUpload Response Body

The `POST /images/:id/complete` handler SHALL return `200 OK` with a JSON body on success.

Response shape:
```json
{
  "image_id": "<uuid>"
}
```

- `image_id` SHALL always be present
- `suggested_folder_name` and `warning` fields are removed; vision labelling is now asynchronous
- If thumbnail generation fails in the synchronous phase, the handler SHALL return a non-2xx error response

`CompleteUploadResult` in `internal/usecase/` SHALL be simplified to:
```go
type CompleteUploadResult struct {
    ImageID uuid.UUID
}
```

#### Scenario: Successful CompleteUpload returns only image_id

- **WHEN** `CompleteUpload` succeeds
- **THEN** the response is `200 OK`
- **AND** the response body contains `image_id`
- **AND** the response body does not contain `suggested_folder_name` or `warning`

#### Scenario: Thumbnail generation failure returns error

- **WHEN** `prepareThumbnail` returns an error during `CompleteUpload`
- **THEN** the response is `500 Internal Server Error`
- **AND** the `pending_uploads` row is not committed to `images`

---

### Requirement: GET /images unfiled query parameter

The `GET /images` handler SHALL accept an optional `unfiled` boolean query parameter.

| `unfiled` value | Behaviour |
|---|---|
| Absent or `false` | No unfiled filter applied; other filters behave as documented |
| `true` | Returns only images with no row in `image_folders`; composes with `folder_ids` and other filters via `AND` exactly like any other filter (see below) |

`ListImagesParams` SHALL include an `Unfiled bool` field. When `Unfiled = true`, the repository SHALL use a LEFT JOIN on `image_folders` and filter for images with no matching row.

`unfiled` and `folder_ids` are independent filters and are NOT mutually exclusive at the parameter level — both are simply passed through and ANDed together like any other filter pair. `unfiled=true` together with a non-empty `folder_ids` is a contradiction (an unfiled image cannot belong to a folder) and SHALL NOT be specially detected, rejected, or have one override the other; it SHALL simply produce an empty result set, the same way any other impossible filter combination would.

#### Scenario: unfiled=true returns only unfoldered images

- **WHEN** `GET /images?unfiled=true` is called
- **THEN** only images with no row in `image_folders` are returned

#### Scenario: unfiled=true combined with folder_ids yields an empty result

- **WHEN** `GET /images?unfiled=true&folder_ids=<valid-uuid>` is called
- **THEN** the response is `200 OK` with an empty `images` array
- **AND** no special validation error is returned — the contradictory filters simply compose to match nothing

#### Scenario: unfiled absent or false preserves existing behaviour

- **WHEN** `GET /images` is called without `unfiled` or with `unfiled=false`
- **THEN** existing filtering behaviour applies unchanged

---

### Requirement: GET /images sort and direction query parameters

The `GET /images` handler SHALL accept optional `sort` and `direction` query parameters that select the ordering of returned images.

| `sort` value  | Meaning                        | Default `direction` when omitted |
|---------------|--------------------------------|-----------------------------------|
| (absent)      | Use the default ordering (`created_at DESC`) | n/a — `direction` has no effect |
| `created_at`  | Order by creation time         | `desc` (newest first)             |
| `title`       | Order alphabetically by title  | `asc` (A → Z)                     |

Rules:
- `sort`, when present and non-empty, SHALL be validated against the allow-list above (`created_at`, `title`); any other value SHALL cause the handler to return `400 Bad Request`
- `direction`, when present and non-empty, SHALL be validated against `{asc, desc}`; any other value SHALL cause the handler to return `400 Bad Request`
- `direction` has no effect when `sort` is absent or empty: it is accepted without validation in that case
- When `sort` is present and valid but `direction` is absent or empty, the field's default direction (per the table above) is used
- The validated values SHALL be passed to `imageUsecase.ListImages` via `ListImagesParams.Sort`/`ListImagesParams.Direction` as `*string`, or `nil` when the corresponding query parameter was absent or empty

#### Scenario: Explicit sort field with explicit direction

- **WHEN** `GET /images?sort=title&direction=desc` is called
- **THEN** results are ordered by `title DESC, id DESC`

#### Scenario: Explicit sort field without direction uses the field's default direction

- **WHEN** `GET /images?sort=title` is called without a `direction` parameter
- **THEN** results are ordered by `title ASC, id ASC`

#### Scenario: Invalid sort value returns 400

- **WHEN** `GET /images?sort=file_size` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: Invalid direction value returns 400

- **WHEN** `GET /images?sort=title&direction=descending` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: Direction without sort is accepted and has no effect

- **WHEN** `GET /images?direction=asc` is called without a `sort` parameter
- **THEN** the response is `200 OK`
- **AND** the view's existing default ordering applies, unaffected by the `direction` value

#### Scenario: Omitted sort and direction preserve existing behaviour

- **WHEN** `GET /images` is called without `sort` or `direction`
- **THEN** results are ordered by `created_at DESC, id DESC`, exactly as before this change
- **AND** the response cursor remains an opaque string requiring no client-side changes

---

### Requirement: Image Usecase Unit Tests

The system SHALL have unit tests for `imageUsecase` in `usecase/image_usecase_test.go` covering each method with mocked `ImageRepository` and `StorageService`. Trash lifecycle methods are covered by `trashUsecase` tests and SHALL NOT appear in `image_usecase_test.go`.

#### Scenario: Usecase unit tests cover the happy path and failure path

- **WHEN** each imageUsecase method is tested with a valid mock setup
- **THEN** both the success and at least one error case are asserted

---

### Requirement: Image Handler Unit Tests

The system SHALL have unit tests for `ImageHandler` in `handler/image_test.go` covering each handler method on `ImageHandler` with a mocked `ImageUsecase`. Trash handler methods are covered by `TrashHandler` tests in `handler/trash_test.go` and SHALL NOT appear in `image_test.go`.

#### Scenario: Handler unit tests cover HTTP status codes and response shape

- **WHEN** each ImageHandler method is tested with a mock usecase
- **THEN** both the success status code and at least one error status code are asserted

---

### Requirement: TrashUsecase interface and TrashHandler defined

The system SHALL define a `TrashUsecase` interface in `internal/usecase/trash_usecase.go` owning the full trash lifecycle: `SoftDelete`, `ListTrashed`, `Restore`, `PurgeExpiredTrash`, `DeleteFromTrash`, `EmptyTrash`, `ProcessR2Delete`.

A `TrashHandler` in `internal/handler/trash.go` SHALL depend on a `TrashUsecase` interface defined locally in that file (following the same pattern as `ImageHandler`). `TrashHandler` SHALL handle all trash lifecycle routes:

- `DELETE /images/:id` → SoftDelete
- `POST /images/:id/restore` → Restore
- `GET /images/trash` → ListTrashed
- `DELETE /images/trash` → EmptyTrash
- `DELETE /images/trash/:id` → DeleteFromTrash

#### Scenario: TrashHandler registered for trash lifecycle routes

- **WHEN** the Go package is compiled and the server starts
- **THEN** `TrashHandler` handles `DELETE /images/:id`, `POST /images/:id/restore`, `GET /images/trash`, `DELETE /images/trash`, and `DELETE /images/trash/:id`
- **AND** `ImageHandler` does not handle any of those routes

#### Scenario: TrashHandler unit tests exist in handler/trash_test.go

- **WHEN** `handler/trash_test.go` is compiled
- **THEN** it covers each `TrashHandler` method with a mocked `TrashUsecase` with at minimum one success and one error scenario per method

---

### Requirement: Image Repository Integration Tests

The system SHALL have integration tests for `imageRepository` using Testcontainers. Each repository method SHALL be tested against a real PostgreSQL database. Unit tests SHALL NOT be written for the SQL repository.

#### Scenario: Repository integration tests exercise each method against a real database

- **WHEN** the integration test suite runs with a live PostgreSQL container
- **THEN** each `ImageRepository` method is exercised with at least one success scenario and one failure scenario
