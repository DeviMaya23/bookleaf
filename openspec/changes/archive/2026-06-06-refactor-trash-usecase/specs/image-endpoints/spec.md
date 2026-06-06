## MODIFIED Requirements

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
    Tags        *[]uuid.UUID
}
```

`CompleteUploadResult` is defined in `internal/usecase/`:

```go
type CompleteUploadResult struct {
    ImageID uuid.UUID
}
```

`ListImagesParams` SHALL include a `TagID` field:

```go
type ListImagesParams struct {
    FolderID *uuid.UUID
    Unfiled  bool
    TagID    *uuid.UUID
    Cursor   *ImageCursor
    Limit    int
}
```

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

Trash lifecycle methods (`SoftDelete`, `ListTrashed`, `Restore`, `DeleteFromTrash`, `EmptyTrash`, `PurgeExpiredTrash`) are NOT part of `ImageUsecase`. They belong to `TrashUsecase` (defined in `internal/usecase/trash_usecase.go`).

#### Scenario: ImageUsecase interface contains no trash methods

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` in `internal/usecase/` does not define or implement `SoftDelete`, `ListTrashed`, `Restore`, `DeleteFromTrash`, `EmptyTrash`, or `PurgeExpiredTrash`

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

## ADDED Requirements

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
