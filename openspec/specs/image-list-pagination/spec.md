# image-list-pagination

## Purpose

Defines cursor-based keyset pagination for image listings — cursor encoding/decoding, pagination types, and sort-aware ordering for `GET /images` and `GET /images/trash`.

## Requirements

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

---

### Requirement: Pagination Types

The system SHALL define the following types in `internal/usecase/`:

```go
type ImageCursor struct {
    CreatedAt time.Time
    Title     *string     // non-nil only when the active sort orders by title
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

---

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

---

### Requirement: ImageUsecase ListImages Signature

The `ImageUsecase` interface SHALL update `ListImages` to:

```go
ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error)
```

#### Scenario: Interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `ImageUsecase` without compilation errors

---

### Requirement: ImageUsecase ListTrashed Signature

The `ImageUsecase` interface SHALL update `ListTrashed` to:

```go
ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error)
```

#### Scenario: Interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `ImageUsecase` without compilation errors

---

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

---

### Requirement: ImageRepository List Signature

The `ImageRepository` interface SHALL update `List` to:

```go
List(ctx context.Context, userID string, folderID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)
```

The SQL implementation SHALL:
- Order by `(created_at DESC, id DESC)`
- Apply a keyset filter `(created_at, id) < (cursor.CreatedAt, cursor.ID)` when a cursor is provided
- Fetch `limit + 1` rows to detect next-page existence (the caller is responsible for trimming and cursor extraction)

#### Scenario: First page — no cursor

- **WHEN** `List` is called with `cursor = nil` and `limit = 50`
- **THEN** the query returns up to 51 rows ordered by `created_at DESC, id DESC` with no keyset filter

#### Scenario: Subsequent page — cursor provided

- **WHEN** `List` is called with a cursor encoding `(T, id)`
- **THEN** the query returns rows where `(created_at, id) < (T, id)` ordered by `created_at DESC, id DESC`

#### Scenario: Interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

---

### Requirement: ImageRepository ListTrashed Signature

The `ImageRepository` interface SHALL update `ListTrashed` to:

```go
ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)
```

The SQL implementation SHALL:
- Filter by `deleted_at IS NOT NULL` and `user_id`
- Order by `(created_at DESC, id DESC)`
- Apply a keyset filter `(created_at, id) < (cursor.CreatedAt, cursor.ID)` when a cursor is provided
- Fetch `limit + 1` rows to detect next-page existence

#### Scenario: First page — no cursor

- **WHEN** `ListTrashed` is called with `cursor = nil` and `limit = 50`
- **THEN** the query returns up to 51 soft-deleted rows ordered by `created_at DESC, id DESC` with no keyset filter

#### Scenario: Subsequent page — cursor provided

- **WHEN** `ListTrashed` is called with a cursor encoding `(T, id)`
- **THEN** the query returns soft-deleted rows where `(created_at, id) < (T, id)` ordered by `created_at DESC, id DESC`

#### Scenario: Interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

---

### Requirement: GET /images Pagination Query Parameters

The `GET /images` handler SHALL accept:

| Parameter  | Type   | Default | Max | Description                                          |
|------------|--------|---------|-----|------------------------------------------------------|
| `limit`    | int    | 50      | 200 | Page size (silently clamped, not rejected)            |
| `cursor`   | string | —       | —   | Opaque cursor from a previous response               |
| `folder_id`| uuid   | —       | —   | When present, bypasses cursor/limit (returns all)    |

When `folder_id` is provided, `cursor` and `limit` are ignored entirely. The handler SHALL NOT attempt to parse a cursor and SHALL NOT apply any limit. All images in the folder are returned in a single response.

When `folder_id` is absent (all or unfiled views), cursor/limit behaviour is unchanged: an unparseable `cursor` value SHALL return `400 Bad Request`.

#### Scenario: Folder view ignores cursor and limit

- **WHEN** `GET /images?folder_id=<id>&cursor=<any>&limit=<any>` is called
- **THEN** all images in the folder are returned regardless of cursor or limit values
- **AND** `next_cursor` in the response is `null`

#### Scenario: Request with no pagination params uses defaults (non-folder view)

- **WHEN** `GET /images` is called with no `limit`, `cursor`, or `folder_id` params
- **THEN** up to 50 images are returned

#### Scenario: Request with explicit limit (non-folder view)

- **WHEN** `GET /images?limit=10` is called without `folder_id`
- **THEN** up to 10 images are returned

#### Scenario: Limit above 200 is silently clamped (non-folder view)

- **WHEN** `GET /images?limit=500` is called without `folder_id`
- **THEN** up to 200 images are returned and no error is returned

#### Scenario: Invalid cursor returns 400 (non-folder view)

- **WHEN** `GET /images?cursor=notvalidbase64!!!` is called without `folder_id`
- **THEN** the response is `400 Bad Request`

---

### Requirement: GET /images/trash Pagination Query Parameters

The `GET /images/trash` handler SHALL accept the same pagination parameters as `GET /images`:

| Parameter | Type   | Default | Max | Description |
|-----------|--------|---------|-----|-------------|
| `limit`   | int    | 50      | 200 | Page size (silently clamped, not rejected) |
| `cursor`  | string | —       | —   | Opaque cursor from a previous response |

An unparseable `cursor` value SHALL return `400 Bad Request`.

#### Scenario: Request with no pagination params uses defaults

- **WHEN** `GET /images/trash` is called with no `limit` or `cursor` params
- **THEN** up to 50 trashed images are returned

#### Scenario: Request with explicit limit

- **WHEN** `GET /images/trash?limit=10` is called
- **THEN** up to 10 trashed images are returned

#### Scenario: Limit above 200 is silently clamped

- **WHEN** `GET /images/trash?limit=500` is called
- **THEN** up to 200 trashed images are returned and no error is returned

#### Scenario: Invalid cursor returns 400

- **WHEN** `GET /images/trash?cursor=notvalidbase64!!!` is called
- **THEN** the response is `400 Bad Request`

---

### Requirement: GET /images Response Envelope

The `GET /images` response SHALL change from a plain array to a paginated envelope:

```json
{
  "images": [ /* array of imageResponse objects */ ],
  "next_cursor": "<opaque string | null>"
}
```

- `next_cursor` SHALL be `null` when the current page is the last page
- `next_cursor` SHALL be a non-empty opaque string when more results exist

#### Scenario: Response includes next_cursor when more pages exist

- **WHEN** `GET /images` is called and the total matching images exceed the requested limit
- **THEN** the response body contains a non-null `next_cursor`
- **AND** `images` contains exactly `limit` items

#### Scenario: next_cursor is null on the last page

- **WHEN** `GET /images` is called and all matching images fit within the limit
- **THEN** `next_cursor` is `null` in the response body
- **AND** `images` contains all matching items

#### Scenario: Cursor from one response yields the next page

- **WHEN** the `next_cursor` from a first `GET /images` response is passed as `cursor` in a second request
- **THEN** the second response contains the next set of images in descending `created_at` order with no overlap with the first page

---

### Requirement: GET /images Response Envelope — folder view

When `folder_id` is provided, the `GET /images` response envelope SHALL always have `next_cursor: null`:

```json
{
  "images": [ /* all images in folder, ordered by position ASC */ ],
  "next_cursor": null
}
```

#### Scenario: Folder view always returns null next_cursor

- **WHEN** `GET /images?folder_id=<id>` is called regardless of how many images are in the folder
- **THEN** `next_cursor` is `null` in the response body
- **AND** `images` contains all non-deleted images in that folder ordered by `image_folders.position ASC`

---

### Requirement: GET /images/trash Response Envelope

The `GET /images/trash` response SHALL change from a plain array to a paginated envelope with the same shape as `GET /images`:

```json
{
  "images": [ /* array of imageResponse objects */ ],
  "next_cursor": "<opaque string | null>"
}
```

- `next_cursor` SHALL be `null` when the current page is the last page
- `next_cursor` SHALL be a non-empty opaque string when more results exist

#### Scenario: Response includes next_cursor when more trashed images exist

- **WHEN** `GET /images/trash` is called and the total trashed images exceed the requested limit
- **THEN** the response body contains a non-null `next_cursor`
- **AND** `images` contains exactly `limit` items

#### Scenario: next_cursor is null on the last page of trash

- **WHEN** `GET /images/trash` is called and all trashed images fit within the limit
- **THEN** `next_cursor` is `null` in the response body
- **AND** `images` contains all trashed items

#### Scenario: Cursor from trash response yields the next trash page

- **WHEN** the `next_cursor` from a first `GET /images/trash` response is passed as `cursor` in a second request
- **THEN** the second response contains the next set of trashed images in descending `created_at` order with no overlap with the first page
