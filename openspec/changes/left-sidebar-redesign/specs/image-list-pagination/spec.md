## MODIFIED Requirements

### Requirement: Cursor Encoding and Decoding
The system SHALL encode pagination cursors as base64url strings wrapping a JSON object. For `GET /images`, the cursor JSON SHALL contain `created_at` (RFC3339Nano) and `id` (UUID string). For `GET /images/trash`, the cursor JSON SHALL contain `deleted_at` (RFC3339Nano) and `id` (UUID string). Decoding SHALL return a typed `ImageCursor` struct with an optional `DeletedAt` field. A malformed cursor (invalid base64 or JSON) SHALL be treated as an error by the caller.

#### Scenario: Valid images cursor round-trips correctly
- **WHEN** an `ImageCursor{CreatedAt, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with no loss of precision

#### Scenario: Valid trash cursor round-trips correctly
- **WHEN** an `ImageCursor{DeletedAt, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with `DeletedAt` preserved

#### Scenario: Malformed base64 cursor returns error on decode
- **WHEN** a string that is not valid base64 is passed to the cursor decoder
- **THEN** an error is returned

---

### Requirement: Pagination Types
The system SHALL define the following types in `internal/usecase/`:

```go
type ImageCursor struct {
    CreatedAt time.Time
    DeletedAt *time.Time  // non-nil only for trash cursors
    ID        uuid.UUID
}

type ListImagesParams struct {
    FolderID *uuid.UUID
    Unfiled  bool
    Cursor   *ImageCursor  // nil = first page
    Limit    int           // 0 = use default (50)
}

type ListImagesResult struct {
    Images     []ImageItem
    NextCursor *ImageCursor  // nil = no more pages
}

type ListTrashedParams struct {
    Cursor *ImageCursor  // nil = first page
    Limit  int           // 0 = use default (50)
}

type ListTrashedResult struct {
    Images     []ImageItem
    NextCursor *ImageCursor  // nil = no more pages
}
```

#### Scenario: Zero Limit defaults to 50 for ListImages
- **WHEN** `ListImagesParams.Limit` is 0
- **THEN** the usecase uses a limit of 50

#### Scenario: Zero Limit defaults to 50 for ListTrashed
- **WHEN** `ListTrashedParams.Limit` is 0
- **THEN** the usecase uses a limit of 50

---

### Requirement: ListTrashed sort order
`ListTrashed` SHALL return images sorted by `deleted_at ASC, id ASC` (oldest deleted first). The pagination cursor for trash SHALL be keyed on `deleted_at` and use `(deleted_at, id) > (cursor.deleted_at, cursor.id)` for subsequent pages.

#### Scenario: First page returns oldest deleted images first
- **WHEN** `GET /images/trash` is called without a cursor
- **THEN** the response contains images ordered oldest-deleted-first

#### Scenario: Cursor advances to next page in ascending order
- **WHEN** `GET /images/trash` is called with a valid trash cursor
- **THEN** only images deleted after the cursor's `deleted_at` (or same time with higher id) are returned, maintaining ASC order
