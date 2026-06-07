## MODIFIED Requirements

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/image_repository.go` containing only the methods used by `imageUsecase` and the upload transaction callback. Per the conventions, each usecase defines its own interface for its dependencies — trash-related and upload-related methods live on `TrashRepository` and `UploadImageRepository` respectively.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, name *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — when `folderID` is non-nil: returns all non-deleted images for that folder ordered by `image_folders.position ASC`; `cursor`, `limit`, and `name` are ignored; images are returned with `Tags` and `ImageFolders` preloaded. When `folderID` is nil: returns non-deleted images ordered by `(created_at DESC, id DESC)`; fetches `limit + 1` rows; `cursor` applies a keyset filter; `unfiled` true limits to images with no entry in `image_folders`; `tagID` non-nil filters by tag; `name` non-nil and non-empty filters to images whose `title` contains the value, case-insensitively
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

#### Scenario: List orders by position for folder views

- **WHEN** `List` is called with a non-nil `folderID`
- **THEN** results are ordered by `image_folders.position ASC`
- **AND** cursor and limit parameters have no effect

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
- **AND** this filter composes with any `unfiled`, `tagID`, and cursor conditions already present

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

`ListImagesParams` SHALL include `TagID` and `Name` fields:

```go
type ListImagesParams struct {
    FolderID *uuid.UUID
    Unfiled  bool
    TagID    *uuid.UUID
    Name     *string
    Cursor   *ImageCursor
    Limit    int
}
```

`Name`, when non-nil and non-empty, filters results to images whose title contains the value, case-insensitively. It is ignored when `FolderID` is non-nil (folder views are filtered client-side on the frontend; see `fe-gallery-search`).

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
