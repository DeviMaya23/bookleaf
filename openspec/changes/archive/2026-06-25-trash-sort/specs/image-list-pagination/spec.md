## MODIFIED Requirements

### Requirement: Cursor Encoding and Decoding

The system SHALL encode pagination cursors as base64url strings wrapping a JSON object. For `GET /images`, the cursor JSON SHALL contain the value of the column the active sort orders by — `created_at` (RFC3339Nano) when sorting by creation time, or `title` (string) when sorting by title — plus `id` (UUID string). For `GET /images/trash`, the cursor JSON SHALL contain the value of the column the active sort orders by — `deleted_at` (RFC3339Nano, the default), `created_at` (RFC3339Nano), or `title` (string) — plus `id` (UUID string). Decoding SHALL return a typed `ImageCursor` struct with optional `Title` and `DeletedAt` fields, each populated only in the contexts that use them. A malformed cursor (invalid base64 or JSON) SHALL be treated as an error by the caller.

The cursor payload carries no marker indicating which field it was built on. The caller is responsible for resending the same `sort`/`direction` query parameters on every subsequent page request — exactly the same implicit contract that already governs `folder_id`/`tag_id`/`name` (the cursor has no opinion about which filters or ordering produced the page; the request's parameters are the source of truth). The repository reads whichever cursor field corresponds to the current request's active sort column.

#### Scenario: Valid images cursor round-trips correctly when sorted by created_at

- **WHEN** an `ImageCursor{CreatedAt, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with no loss of precision

#### Scenario: Valid images cursor round-trips correctly when sorted by title

- **WHEN** an `ImageCursor{Title, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with `Title` preserved

#### Scenario: Valid trash cursor round-trips correctly when sorted by deleted_at

- **WHEN** an `ImageCursor{DeletedAt, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with `DeletedAt` preserved

#### Scenario: Valid trash cursor round-trips correctly when sorted by title

- **WHEN** an `ImageCursor{Title, ID}` produced by a trash list request is encoded then decoded
- **THEN** the decoded struct equals the original with `Title` preserved

#### Scenario: Valid trash cursor round-trips correctly when sorted by created_at

- **WHEN** an `ImageCursor{CreatedAt, ID}` produced by a trash list request is encoded then decoded
- **THEN** the decoded struct equals the original with no loss of precision

#### Scenario: Malformed base64 cursor returns error on decode

- **WHEN** a string that is not valid base64 is passed to the cursor decoder
- **THEN** an error is returned

### Requirement: Pagination Types

The system SHALL define the following types in `internal/usecase/`:

```go
type ImageCursor struct {
    CreatedAt time.Time
    Title     *string    // non-nil only when the active sort orders by title
    DeletedAt *time.Time // non-nil only when the active sort orders by deleted_at (trash only)
    ID        uuid.UUID
}

type ListImagesParams struct {
    FolderID  *uuid.UUID
    Unfiled   bool
    TagID     *uuid.UUID
    Name      *string
    Sort      *string       // nil = view's default ordering; validated upstream by the handler
    Direction *string       // nil = sort field's default direction; validated upstream by the handler
    Cursor    *ImageCursor  // nil = first page
    Limit     int           // 0 = use default (50)
}

type ListImagesResult struct {
    Images     []ImageItem
    NextCursor *ImageCursor  // nil = no more pages
}

type ListTrashedParams struct {
    Name      *string
    Sort      *string       // nil = handler defaults to "deleted_at"; validated upstream by the handler
    Direction *string       // nil = sort field's default direction; validated upstream by the handler
    Cursor    *ImageCursor  // nil = first page
    Limit     int           // 0 = use default (50)
}

type ListTrashedResult struct {
    Images     []ImageItem
    NextCursor *ImageCursor  // nil = no more pages
}
```

#### Scenario: Zero Limit defaults to 50 for ListImages

- **WHEN** `ListImagesParams.Limit` is 0
- **THEN** the usecase uses a limit of 50

#### Scenario: Limit above 200 is clamped for ListImages

- **WHEN** `ListImagesParams.Limit` exceeds 200
- **THEN** the usecase caps it at 200

#### Scenario: Zero Limit defaults to 50 for ListTrashed

- **WHEN** `ListTrashedParams.Limit` is 0
- **THEN** the usecase uses a limit of 50

#### Scenario: Limit above 200 is clamped for ListTrashed

- **WHEN** `ListTrashedParams.Limit` exceeds 200
- **THEN** the usecase caps it at 200

#### Scenario: ListTrashed passes Sort and Direction through to the repository unchanged

- **WHEN** `ListTrashed` is called with non-nil `Sort` and `Direction` params
- **THEN** the repository's `ListTrashed` is invoked with those exact values as `sortField`/`direction`, with no further validation, normalization, or defaulting performed by the usecase

#### Scenario: ListTrashed passes nil Sort and Direction through as nil

- **WHEN** `ListTrashed` is called with `Sort` and `Direction` both nil
- **THEN** the repository's `ListTrashed` is invoked with nil `sortField`/`direction`

### Requirement: ListTrashed sort order

`ListTrashed`'s `ORDER BY` clause and pagination-cursor keyset comparison SHALL both be derived from the active sort field and direction via the same `ResolveSort` dispatch used by `GET /images`, which is extended with a `deleted_at` case used only by `ListTrashed`:

| Active sort  | Direction | `ORDER BY`                  | Keyset `WHERE` (when a cursor is present)                |
|--------------|-----------|------------------------------|------------------------------------------------------------|
| `deleted_at` | `desc`    | `deleted_at DESC, id DESC`  | `(deleted_at, id) < (cursor.deleted_at, cursor.id)`         |
| `deleted_at` | `asc`     | `deleted_at ASC, id ASC`    | `(deleted_at, id) > (cursor.deleted_at, cursor.id)`         |
| `created_at` | `desc`    | `created_at DESC, id DESC`   | `(created_at, id) < (cursor.created_at, cursor.id)`         |
| `created_at` | `asc`     | `created_at ASC, id ASC`     | `(created_at, id) > (cursor.created_at, cursor.id)`         |
| `title`      | `asc`     | `title ASC, id ASC`          | `(title, id) > (cursor.title, cursor.id)`                   |
| `title`      | `desc`    | `title DESC, id DESC`        | `(title, id) < (cursor.title, cursor.id)`                   |

When the `sort` query parameter is absent, `GET /images/trash`'s handler defaults the sort field to `deleted_at` (not left nil) before this dispatch runs — see `image-endpoints`. This is a behavior change from `ListTrashed`'s original default (`deleted_at ASC, id ASC`, oldest deleted first): the new default is `deleted_at DESC, id DESC` (newest deleted first). The comparison operator SHALL be `>` for ascending order and `<` for descending order; `id` SHALL always serve as the tiebreaker column, ordered in the same direction as the primary sort column.

#### Scenario: First page returns newest-deleted images first by default

- **WHEN** `GET /images/trash` is called without a cursor, `sort`, or `direction`
- **THEN** the response contains images ordered newest-deleted-first (`deleted_at DESC, id DESC`)

#### Scenario: Cursor advances to next page for the default sort

- **WHEN** `GET /images/trash` is called with a valid cursor and no `sort`/`direction`
- **THEN** only images deleted before the cursor's `deleted_at` (or same time with lower id) are returned, maintaining DESC order

#### Scenario: Explicit title sort orders and paginates by title

- **WHEN** `ListTrashed` is called with the active sort `title` ascending and a cursor encoding `(T, id)`
- **THEN** the query returns rows where `(title, id) > (T, id)` ordered by `title ASC, id ASC`

#### Scenario: Explicit created_at sort orders and paginates by created_at

- **WHEN** `ListTrashed` is called with the active sort `created_at` ascending and a cursor encoding `(T, id)`
- **THEN** the query returns rows where `(created_at, id) > (T, id)` ordered by `created_at ASC, id ASC`

#### Scenario: Explicit deleted_at ascending sort orders and paginates oldest-deleted-first

- **WHEN** `ListTrashed` is called with the active sort `deleted_at` ascending and a cursor encoding `(T, id)`
- **THEN** the query returns rows where `(deleted_at, id) > (T, id)` ordered by `deleted_at ASC, id ASC`
