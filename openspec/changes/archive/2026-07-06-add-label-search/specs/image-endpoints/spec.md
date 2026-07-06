## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/image_repository.go` containing only the methods used by `imageUsecase` and the upload transaction callback. Per the conventions, each usecase defines its own interface for its dependencies — trash-related and upload-related methods live on `TrashRepository` and `UploadImageRepository` respectively.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name *string, searchLabels bool, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns non-deleted images for `userID`; fetches `limit + 1` rows; `cursor` applies a keyset filter on the active sort column; images are returned with `Tags` and `ImageFolders` preloaded. `unfiled` true limits to images with no entry in `image_folders` (LEFT JOIN + `IS NULL`, as before). `folderIDs` non-empty filters to images belonging to ANY of the given folders via a correlated `EXISTS` subquery against `image_folders`, so an image matching multiple supplied folder IDs is still returned at most once. `tagIDs` non-empty filters to images associated with ANY of the given tags via a correlated `EXISTS` subquery against `image_tags`, with the same at-most-once guarantee. `mimeTypes` non-empty filters to images whose `mime_type` matches any of the given values via `IN`. `name` non-nil and non-empty filters to images whose `title` contains the value, case-insensitively; when `searchLabels` is also `true`, the filter is widened to `(images.title ILIKE '%<term>%' OR EXISTS (SELECT 1 FROM image_labels WHERE image_id = images.id AND label ILIKE '%<term>%' AND score >= 0.75))`. When `name` is nil or empty, `searchLabels` has no effect. All of `unfiled`/`folderIDs`/`tagIDs`/`mimeTypes`/`name` (including the OR EXISTS expansion) compose via `AND`; no validation rejects contradictory combinations (e.g. `unfiled=true` together with non-empty `folderIDs`) — such combinations simply yield an empty result, the same way an impossible `name`/`tagIDs` combination would (see `GET /images Multi-Value Filter Query Parameters`). `sortField`/`direction` select the ordering: when `sortField` is nil, results order by `created_at DESC, id DESC` (today's default, unchanged); when `sortField` is non-nil, results order by the selected column (with `id` as tiebreaker) in the selected direction (the field's default direction applies when `direction` is nil), and the keyset filter compares against that same column instead of `created_at` (see `image-list-pagination` for the keyset comparison rules). The single-folder, position-ordered, unpaginated "folder contents" query that this method previously served when `folderID` was non-nil is REMOVED from this method — it is now served by a dedicated method described in `folder-image-listing`.
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

- **WHEN** `List` is called with a non-nil, non-empty `name` and `searchLabels = false`
- **THEN** the query includes a case-insensitive substring match (`ILIKE '%<name>%'`) against `images.title` only
- **AND** this filter composes with any `unfiled`, `folderIDs`, `tagIDs`, `mimeTypes`, sort, and cursor conditions already present

#### Scenario: List with searchLabels=true widens name filter to include AI labels above score threshold

- **WHEN** `List` is called with a non-nil, non-empty `name` and `searchLabels = true`
- **THEN** the query matches images whose title contains the name OR that have a label in `image_labels` containing the name case-insensitively with `score >= 0.75`
- **AND** this combined condition composes with other active filters via AND

#### Scenario: List with searchLabels=true and empty name ignores label filter

- **WHEN** `List` is called with `searchLabels = true` but `name` is nil or empty
- **THEN** no label subquery is added — results are identical to calling with `searchLabels = false`

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

`ListImagesParams` SHALL include `Unfiled`, `FolderIDs`, `TagIDs`, `MIMETypes`, `Name`, `SearchLabels`, `Sort`, and `Direction` fields:

```go
type ListImagesParams struct {
    Unfiled      bool
    FolderIDs    []uuid.UUID
    TagIDs       []uuid.UUID
    MIMETypes    []string
    Name         *string
    SearchLabels bool
    Sort         *string
    Direction    *string
    Cursor       *ImageCursor
    Limit        int
}
```

The single-value `FolderID *uuid.UUID` and `TagID *uuid.UUID` fields are REMOVED from `ListImagesParams`. There is no longer a folder-view mode reachable through `ListImages`/`ListImagesParams` — fetching a single folder's contents in custom order is handled by a dedicated method described in `folder-image-listing`.

`Name`, when non-nil and non-empty, filters results to images whose title contains the value, case-insensitively. When `SearchLabels` is also `true`, the filter is widened to include AI label matches (see `image-label-search`).

`SearchLabels`, when `true`, has no effect unless `Name` is also non-nil and non-empty.

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

#### Scenario: ListImages passes SearchLabels to repository

- **WHEN** `ListImages` is called with `SearchLabels = true`
- **THEN** the repository's `List` is invoked with `searchLabels = true`

#### Scenario: ListImages passes Sort and Direction through to the repository unchanged

- **WHEN** `ListImages` is called with non-nil `Sort` and `Direction` params
- **THEN** the repository's `List` is invoked with those exact values as `sortField`/`direction`, with no further validation, normalization, or defaulting performed by the usecase

#### Scenario: ListImages passes nil Sort and Direction through as nil

- **WHEN** `ListImages` is called with `Sort` and `Direction` both nil
- **THEN** the repository's `List` is invoked with nil `sortField`/`direction`, preserving the default ordering
