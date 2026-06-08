## MODIFIED Requirements

### Requirement: Cursor Encoding and Decoding

The system SHALL encode pagination cursors as base64url strings wrapping a JSON object. For `GET /images`, the cursor JSON SHALL contain the value of the column the active sort orders by — `created_at` (RFC3339Nano) when sorting by creation time, or `title` (string) when sorting by title — plus `id` (UUID string). For `GET /images/trash`, the cursor JSON SHALL contain `deleted_at` (RFC3339Nano) and `id` (UUID string). Decoding SHALL return a typed `ImageCursor` struct with optional `Title` and `DeletedAt` fields, each populated only in the contexts that use them. A malformed cursor (invalid base64 or JSON) SHALL be treated as an error by the caller.

The cursor payload carries no marker indicating which field it was built on. The caller is responsible for resending the same `sort`/`direction` query parameters on every subsequent page request — exactly the same implicit contract that already governs `folder_id`/`tag_id`/`name` (the cursor has no opinion about which filters or ordering produced the page; the request's parameters are the source of truth). The repository reads whichever cursor field corresponds to the current request's active sort column.

#### Scenario: Valid images cursor round-trips correctly when sorted by created_at

- **WHEN** an `ImageCursor{CreatedAt, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with no loss of precision

#### Scenario: Valid images cursor round-trips correctly when sorted by title

- **WHEN** an `ImageCursor{Title, ID}` is encoded then decoded
- **THEN** the decoded struct equals the original with `Title` preserved

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
    Title     *string     // non-nil only when the active sort orders by title
    DeletedAt *time.Time  // non-nil only for trash cursors
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

#### Scenario: Limit above 200 is clamped for ListImages

- **WHEN** `ListImagesParams.Limit` exceeds 200
- **THEN** the usecase caps it at 200

#### Scenario: Zero Limit defaults to 50 for ListTrashed

- **WHEN** `ListTrashedParams.Limit` is 0
- **THEN** the usecase uses a limit of 50

#### Scenario: Limit above 200 is clamped for ListTrashed

- **WHEN** `ListTrashedParams.Limit` exceeds 200
- **THEN** the usecase caps it at 200

---

## ADDED Requirements

### Requirement: Sort-aware keyset comparison for GET /images

When `folderID` is nil, the `ORDER BY` clause and the keyset `WHERE` comparison SHALL both be derived from the active sort field and direction, generalizing the previously hardcoded `created_at DESC` behaviour:

| Active sort  | Direction | `ORDER BY`                 | Keyset `WHERE` (when a cursor is present)              |
|--------------|-----------|----------------------------|---------------------------------------------------------|
| `created_at` | `desc`    | `created_at DESC, id DESC` | `(created_at, id) < (cursor.created_at, cursor.id)`     |
| `created_at` | `asc`     | `created_at ASC, id ASC`   | `(created_at, id) > (cursor.created_at, cursor.id)`     |
| `title`      | `asc`     | `title ASC, id ASC`        | `(title, id) > (cursor.title, cursor.id)`               |
| `title`      | `desc`    | `title DESC, id DESC`      | `(title, id) < (cursor.title, cursor.id)`               |

The comparison operator SHALL be `>` for ascending order and `<` for descending order. `id` SHALL always serve as the tiebreaker column, ordered in the same direction as the primary sort column — generalizing the existing `created_at`/`id` tiebreak so that pagination remains stable across pages even when multiple rows share the same value in the primary column (e.g. duplicate titles).

When `sortField` is nil (the default), the table's `created_at`/`desc` row applies — identical to today's behaviour.

#### Scenario: Ascending sort uses greater-than keyset comparison

- **WHEN** `List` is called with `folderID = nil`, the active sort is `title` ascending, and a cursor encoding `(T, id)`
- **THEN** the query returns rows where `(title, id) > (T, id)` ordered by `title ASC, id ASC`

#### Scenario: Descending sort uses less-than keyset comparison

- **WHEN** `List` is called with `folderID = nil`, the active sort is `created_at` descending, and a cursor encoding `(T, id)`
- **THEN** the query returns rows where `(created_at, id) < (T, id)` ordered by `created_at DESC, id DESC`

#### Scenario: id tiebreaks alongside the primary column in matching direction

- **WHEN** two or more images share the same value in the active sort column
- **THEN** they are ordered relative to one another by `id` in the same direction (`ASC`/`DESC`) as the primary column
- **AND** a keyset cursor built from the page boundary correctly excludes every already-seen `(value, id)` pair from the next page, with no duplicates or gaps

#### Scenario: Default sort behaves exactly as before this change

- **WHEN** `List` is called with `folderID = nil` and no explicit sort field
- **THEN** the `ORDER BY` and keyset comparison are `created_at DESC, id DESC` and `(created_at, id) < (cursor.created_at, cursor.id)` respectively — unchanged from prior behaviour
