### Requirement: Cursor Encoding and Decoding

The system SHALL encode pagination cursors as base64url strings wrapping a JSON object with `created_at` (RFC3339Nano) and `id` (UUID string) fields. Decoding SHALL return a typed `ImageCursor` struct. A malformed cursor (invalid base64 or JSON) SHALL be treated as an error by the caller. The same encoding is shared between `GET /images` and `GET /images/trash`.

#### Scenario: Valid cursor round-trips correctly

- **WHEN** an `ImageCursor{CreatedAt, ID}` is encoded then decoded
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
    ID        uuid.UUID
}

type ListImagesParams struct {
    FolderID *uuid.UUID
    Cursor   *ImageCursor  // nil = first page
    Limit    int           // 0 = use default (50)
}

type ListImagesResult struct {
    Images     []*domain.Image
    NextCursor *ImageCursor  // nil = no more pages
}

type ListTrashedParams struct {
    Cursor *ImageCursor  // nil = first page
    Limit  int           // 0 = use default (50)
}

type ListTrashedResult struct {
    Images     []*domain.Image
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

| Parameter | Type   | Default | Max | Description |
|-----------|--------|---------|-----|-------------|
| `limit`   | int    | 50      | 200 | Page size (silently clamped, not rejected) |
| `cursor`  | string | —       | —   | Opaque cursor from a previous response |

An unparseable `cursor` value SHALL return `400 Bad Request`.

#### Scenario: Request with no pagination params uses defaults

- **WHEN** `GET /images` is called with no `limit` or `cursor` params
- **THEN** up to 50 images are returned

#### Scenario: Request with explicit limit

- **WHEN** `GET /images?limit=10` is called
- **THEN** up to 10 images are returned

#### Scenario: Limit above 200 is silently clamped

- **WHEN** `GET /images?limit=500` is called
- **THEN** up to 200 images are returned and no error is returned

#### Scenario: Invalid cursor returns 400

- **WHEN** `GET /images?cursor=notvalidbase64!!!` is called
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
