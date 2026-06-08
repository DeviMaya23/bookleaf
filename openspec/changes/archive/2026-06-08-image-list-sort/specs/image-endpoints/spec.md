## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/image_repository.go` containing only the methods used by `imageUsecase` and the upload transaction callback. Per the conventions, each usecase defines its own interface for its dependencies — trash-related and upload-related methods live on `TrashRepository` and `UploadImageRepository` respectively.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, name *string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — when `folderID` is non-nil: returns all non-deleted images for that folder; `cursor`, `limit`, and `name` are ignored; images are returned with `Tags` and `ImageFolders` preloaded. When `folderID` is nil: returns non-deleted images; fetches `limit + 1` rows; `cursor` applies a keyset filter; `unfiled` true limits to images with no entry in `image_folders`; `tagID` non-nil filters by tag; `name` non-nil and non-empty filters to images whose `title` contains the value, case-insensitively. In both branches, `sortField`/`direction` select the ordering: when `sortField` is nil, the folder branch orders by `image_folders.position ASC` and the non-folder branch orders by `created_at DESC, id DESC` (today's defaults, unchanged); when `sortField` is non-nil, both branches order by the selected column (with `id` as tiebreaker) in the selected direction (the field's default direction applies when `direction` is nil), and the non-folder branch's keyset filter compares against that same column instead of `created_at` (see `image-list-pagination` for the keyset comparison rules)
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

#### Scenario: List defaults to position ordering for folder views when sort is nil

- **WHEN** `List` is called with a non-nil `folderID` and a nil `sortField`
- **THEN** results are ordered by `image_folders.position ASC`
- **AND** cursor and limit parameters have no effect

#### Scenario: List honors an explicit sort field in folder views

- **WHEN** `List` is called with a non-nil `folderID` and a non-nil `sortField` (e.g. `title`)
- **THEN** results are ordered by the selected column and direction instead of `image_folders.position`
- **AND** cursor and limit parameters still have no effect (the branch remains unpaginated)

#### Scenario: List defaults to created_at-descending ordering for non-folder views when sort is nil

- **WHEN** `List` is called with `folderID = nil` and a nil `sortField`
- **THEN** results are ordered by `created_at DESC, id DESC`
- **AND** the keyset filter, when a cursor is present, compares `(created_at, id)`

#### Scenario: List honors an explicit sort field in non-folder views

- **WHEN** `List` is called with `folderID = nil` and `sortField = "title"`
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

#### Scenario: List filters by tagID when provided

- **WHEN** `List` is called with a non-nil `tagID`
- **THEN** only images associated with that tag are returned

#### Scenario: List filters by name in the cursor-paginated branch

- **WHEN** `List` is called with `folderID = nil` and a non-nil, non-empty `name`
- **THEN** the query includes a case-insensitive substring match (`ILIKE '%<name>%'`) against `images.title`
- **AND** this filter composes with any `unfiled`, `tagID`, sort, and cursor conditions already present

#### Scenario: List ignores name in the folder-view branch

- **WHEN** `List` is called with a non-nil `folderID` and a non-nil `name`
- **THEN** the `name` value has no effect on the returned results

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

`ListImagesParams` SHALL include `TagID`, `Name`, `Sort`, and `Direction` fields:

```go
type ListImagesParams struct {
    FolderID  *uuid.UUID
    Unfiled   bool
    TagID     *uuid.UUID
    Name      *string
    Sort      *string
    Direction *string
    Cursor    *ImageCursor
    Limit     int
}
```

`Name`, when non-nil and non-empty, filters results to images whose title contains the value, case-insensitively. It is ignored when `FolderID` is non-nil (folder views are filtered client-side on the frontend; see `fe-gallery-search`).

`Sort`, when non-nil, selects the ordering column (`created_at` or `title`); `Direction`, when non-nil, selects `asc` or `desc`. Both are validated against an allow-list by the handler before reaching the usecase (see `GET /images sort and direction query parameters`) — the usecase passes them through to the repository unchanged, performing no additional validation or defaulting of its own. When `Sort` is nil, ordering follows the per-branch defaults described in the `ImageRepository.List` requirement (`position ASC` for folder views, `created_at DESC` otherwise).

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

#### Scenario: ListImages passes TagID to repository

- **WHEN** `ListImages` is called with a non-nil `TagID` param
- **THEN** only images associated with that tag are returned

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
- **THEN** the repository's `List` is invoked with nil `sortField`/`direction`, preserving the view's existing default ordering

---

## ADDED Requirements

### Requirement: GET /images sort and direction query parameters

The `GET /images` handler SHALL accept optional `sort` and `direction` query parameters that select the ordering of returned images.

| `sort` value  | Meaning                        | Default `direction` when omitted |
|---------------|--------------------------------|-----------------------------------|
| (absent)      | Use the view's existing default ordering (`image_folders.position ASC` for folder views, `created_at DESC` otherwise) | n/a — `direction` has no effect |
| `created_at`  | Order by creation time         | `desc` (newest first)             |
| `title`       | Order alphabetically by title  | `asc` (A → Z)                     |

Rules:
- `sort`, when present and non-empty, SHALL be validated against the allow-list above (`created_at`, `title`); any other value SHALL cause the handler to return `400 Bad Request`
- `direction`, when present and non-empty, SHALL be validated against `{asc, desc}`; any other value SHALL cause the handler to return `400 Bad Request`
- `direction` has no effect when `sort` is absent or empty: it is accepted without validation in that case, mirroring how the folder-view branch already silently ignores unrelated parameters such as `cursor`/`limit`/`name`
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
- **THEN** folder views are ordered by `image_folders.position ASC` and non-folder views by `created_at DESC, id DESC`, exactly as before this change
- **AND** the response cursor remains an opaque string requiring no client-side changes
