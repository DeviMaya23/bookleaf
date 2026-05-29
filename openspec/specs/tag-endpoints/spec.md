## Purpose

Defines the Tag usecase interface, handler, route wiring, and Bruno request files for CRUD operations on user-owned tags.

## Requirements

### Requirement: Tag Usecase Interface

The system SHALL define a `TagUsecase` interface in `internal/usecase/tag_usecase.go` and a concrete implementation.

```go
var (
    ErrInvalidTagName  = errors.New("tag name is required")
    ErrDuplicateTagName = errors.New("tag name already exists")
)

type TagUsecase interface {
    Create(ctx context.Context, userID string, name string) (*domain.Tag, error)
    List(ctx context.Context, userID string) ([]*domain.Tag, error)
    Update(ctx context.Context, id uuid.UUID, userID string, name string) (*domain.Tag, error)
    Delete(ctx context.Context, id uuid.UUID, userID string) error
}
```

`Create` SHALL return `ErrInvalidTagName` if `name` is empty or whitespace-only.
`Create` SHALL return `ErrDuplicateTagName` if a tag with the same name already exists for the user (detected from a unique constraint violation).
`Update` SHALL return `ErrInvalidTagName` if `name` is empty or whitespace-only.
`Update` SHALL return `ErrDuplicateTagName` if the new name conflicts with an existing tag for the user.

#### Scenario: Create returns error for empty name

- **WHEN** `Create` is called with a whitespace-only name
- **THEN** `ErrInvalidTagName` is returned

#### Scenario: Create returns error for duplicate name

- **GIVEN** a tag named "nature" already exists for the user
- **WHEN** `Create` is called with name "nature"
- **THEN** `ErrDuplicateTagName` is returned

#### Scenario: Create success

- **WHEN** `Create` is called with a valid unique name
- **THEN** a new `Tag` record is returned with a non-nil UUID and the given name

#### Scenario: Update returns error for duplicate name

- **GIVEN** a tag named "travel" already exists for the user
- **WHEN** `Update` is called on a different tag with name "travel"
- **THEN** `ErrDuplicateTagName` is returned

### Requirement: Tag Handler

The system SHALL define a `TagHandler` in `internal/handler/tag.go` with four methods wired to the tag routes.

Request/response types:

```go
type createTagRequest struct {
    Name string `json:"name"`
}

type updateTagRequest struct {
    Name string `json:"name"`
}

type tagResponse struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

Handler methods:
- `CreateTag` — binds `createTagRequest`, calls `tagUsecase.Create`, returns `201 Created` with `tagResponse`
- `ListTags` — calls `tagUsecase.List`, returns `200 OK` with `[]tagResponse`
- `UpdateTag` — parses `:id`, binds `updateTagRequest`, calls `tagUsecase.Update`, returns `200 OK` with `tagResponse`
- `DeleteTag` — parses `:id`, calls `tagUsecase.Delete`, returns `204 No Content`

Error mapping:
- `ErrInvalidTagName` → `400 Bad Request`
- `ErrDuplicateTagName` → `409 Conflict`
- `gorm.ErrRecordNotFound` → `404 Not Found`
- All other errors → `500 Internal Server Error`

Each handler method SHALL start a tracing span via `tel.Tracer.Start`.

#### Scenario: CreateTag returns 201 with created tag

- **WHEN** `POST /tags` is called with `{"name": "landscape"}`
- **THEN** the response is `201 Created` with the tag's `id`, `name`, `created_at`, `updated_at`

#### Scenario: CreateTag returns 400 for empty name

- **WHEN** `POST /tags` is called with `{"name": ""}`
- **THEN** the response is `400 Bad Request`

#### Scenario: CreateTag returns 409 for duplicate name

- **GIVEN** the user already has a tag named "landscape"
- **WHEN** `POST /tags` is called with `{"name": "landscape"}`
- **THEN** the response is `409 Conflict`

#### Scenario: ListTags returns all user tags

- **WHEN** `GET /tags` is called
- **THEN** the response is `200 OK` with an array of the authenticated user's tags

#### Scenario: UpdateTag returns 200 with renamed tag

- **WHEN** `PUT /tags/:id` is called with a new name
- **THEN** the response is `200 OK` with the updated tag

#### Scenario: UpdateTag returns 404 for unknown tag

- **WHEN** `PUT /tags/:id` is called with an ID that does not belong to the user
- **THEN** the response is `404 Not Found`

#### Scenario: DeleteTag returns 204

- **WHEN** `DELETE /tags/:id` is called for an existing tag
- **THEN** the response is `204 No Content`
- **AND** the tag's associations in `image_tags` are also removed

### Requirement: Tag Routes Wiring

The system SHALL register tag routes on the protected Echo group in `cmd/server/main.go`.

Routes:
- `POST /tags` → `tagHandler.CreateTag`
- `GET /tags` → `tagHandler.ListTags`
- `PUT /tags/:id` → `tagHandler.UpdateTag`
- `DELETE /tags/:id` → `tagHandler.DeleteTag`

#### Scenario: Tag routes require authentication

- **WHEN** any `/tags` route is called without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

### Requirement: Bruno Files for Tag Endpoints

The system SHALL include Bruno request files for each tag endpoint under `bruno/`.

Files:
- `bruno/tags/create-tag.bru`
- `bruno/tags/list-tags.bru`
- `bruno/tags/update-tag.bru`
- `bruno/tags/delete-tag.bru`

#### Scenario: Bruno files exist for all tag endpoints

- **WHEN** the bruno collection is opened
- **THEN** all four tag request files are present and reference the correct method and path
