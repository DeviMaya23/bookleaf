## MODIFIED Requirements

### Requirement: Trash Repository Interface

The system SHALL define a `TrashRepository` interface in `internal/usecase/trash_repository.go` containing only the methods used by `trashUsecase`. Per the conventions, `trashUsecase` defines its own interface for its repository dependency rather than depending on the broader `ImageRepository`.

Methods:
- `GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only
- `GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `SoftDelete(ctx context.Context, id uuid.UUID, userID string) error`
- `Restore(ctx context.Context, id uuid.UUID, userID string) error`
- `ListTrashed(ctx context.Context, userID string, name *string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns soft-deleted images; when `sortField` is nil, ordered by `(deleted_at DESC, id DESC)` (the caller — the handler — is expected to substitute `"deleted_at"` explicitly rather than relying on this method's own nil-field fallback, but the fallback exists for direct/test callers); when non-nil, ordered by the selected column (`deleted_at`, `created_at`, or `title`, with `id` as tiebreaker) in the selected direction, and the keyset filter compares against that same column (see `image-list-pagination` for the keyset comparison rules); fetches `limit + 1` rows; `cursor` nil means first page
- `ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error)` — returns all soft-deleted images for the user with no pagination
- `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `HardDelete(ctx context.Context, id uuid.UUID, userID string) error`

The concrete `*imageRepository` in `internal/repository/` satisfies both `ImageRepository` and `TrashRepository`.

#### Scenario: TrashRepository is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.TrashRepository` without compilation errors

#### Scenario: ListTrashed defaults to deleted_at-descending ordering when sort is nil

- **WHEN** `ListTrashed` is called with a nil `sortField`
- **THEN** results are ordered by `deleted_at DESC, id DESC`

#### Scenario: ListTrashed honors an explicit title sort field

- **WHEN** `ListTrashed` is called with `sortField = "title"`
- **THEN** results are ordered by `title ASC, id ASC` (the field's default direction)
- **AND** the keyset filter, when a cursor is present, compares `(title, id)` using the operator matching the sort direction

#### Scenario: ListTrashed honors an explicit created_at sort field

- **WHEN** `ListTrashed` is called with `sortField = "created_at"`
- **THEN** results are ordered by `created_at DESC, id DESC` (the field's default direction)
- **AND** the keyset filter, when a cursor is present, compares `(created_at, id)` using the operator matching the sort direction

## ADDED Requirements

### Requirement: GET /images/trash sort and direction query parameters

The `GET /images/trash` handler SHALL accept optional `sort` and `direction` query parameters that select the ordering of returned trashed images. This allow-list and default are distinct from `GET /images`'s — `deleted_at` is a valid value here (it has no meaning on `GET /images`, where it is always null) and is the default when `sort` is absent.

| `sort` value  | Meaning                        | Default `direction` when omitted |
|---------------|--------------------------------|-----------------------------------|
| (absent)      | Defaults to `deleted_at` (see below) | n/a — `direction` has no effect on its own |
| `deleted_at`  | Order by deletion time          | `desc` (most recently deleted first) |
| `created_at`  | Order by creation time         | `desc` (newest first)             |
| `title`       | Order alphabetically by title  | `asc` (A → Z)                     |

Rules:
- `sort`, when present and non-empty, SHALL be validated against the allow-list above (`deleted_at`, `created_at`, `title`); any other value SHALL cause the handler to return `400 Bad Request`
- When `sort` is absent or empty, the handler SHALL substitute `"deleted_at"` as the effective sort field before passing it to `trashUsecase.ListTrashed` — this is a deliberate, endpoint-specific default decided in this handler, not a generic "nil means default" fallback shared with `GET /images`
- `direction`, when present and non-empty, SHALL be validated against `{asc, desc}`; any other value SHALL cause the handler to return `400 Bad Request`
- When `direction` is absent or empty, the effective sort field's default direction (per the table above) is used
- The FE's Trash sort control only ever sends `deleted_at` or `title` as explicit `sort` values (see `fe-gallery-sort`); `created_at` remains a valid, allow-listed value for direct API/bruno use even though no FE code path sends it for Trash
- The resolved values SHALL be passed to `trashUsecase.ListTrashed` via `ListTrashedParams.Sort`/`ListTrashedParams.Direction` as `*string`

#### Scenario: Explicit deleted_at sort with explicit direction

- **WHEN** `GET /images/trash?sort=deleted_at&direction=asc` is called
- **THEN** results are ordered by `deleted_at ASC, id ASC`

#### Scenario: Explicit title sort with explicit direction

- **WHEN** `GET /images/trash?sort=title&direction=desc` is called
- **THEN** results are ordered by `title DESC, id DESC`

#### Scenario: Explicit sort field without direction uses the field's default direction

- **WHEN** `GET /images/trash?sort=title` is called without a `direction` parameter
- **THEN** results are ordered by `title ASC, id ASC`

#### Scenario: Invalid sort value returns 400

- **WHEN** `GET /images/trash?sort=file_size` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: Invalid direction value returns 400

- **WHEN** `GET /images/trash?sort=title&direction=descending` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: Omitted sort and direction default to deleted_at descending

- **WHEN** `GET /images/trash` is called without `sort` or `direction`
- **THEN** results are ordered by `deleted_at DESC, id DESC`
- **AND** the response cursor remains an opaque string requiring no client-side changes

#### Scenario: Explicit created_at sort is accepted for direct API use

- **WHEN** `GET /images/trash?sort=created_at` is called
- **THEN** the response is `200 OK`
- **AND** results are ordered by `created_at DESC, id DESC` (the field's default direction)
