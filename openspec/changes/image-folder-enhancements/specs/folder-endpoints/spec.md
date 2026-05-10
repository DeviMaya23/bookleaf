## MODIFIED Requirements

### Requirement: POST /folders — Create Folder

The system SHALL expose a `POST /folders` endpoint on the protected route group that creates a new folder owned by the authenticated user.

Request body:
```json
{ "name": "string (required)", "parent_id": "uuid (optional)", "description": "string (optional)" }
```

Response body (201):
```json
{ "id": "uuid", "name": "string", "description": "string|null", "parent_id": "uuid|null", "created_at": "timestamp", "updated_at": "timestamp" }
```

- `parent_id` in the request is optional; omitting it creates a root-level folder
- If `parent_id` is provided, the referenced folder MUST be owned by the authenticated user
- `name` is required and MUST NOT be empty
- `description` is optional; omitting it stores NULL

#### Scenario: Authenticated user creates a folder with description

- **WHEN** an authenticated `POST /folders` request is made with a valid `name` and a `description`
- **THEN** the response is `201 Created`
- **AND** the body contains the new folder with the supplied `description`

#### Scenario: Authenticated user creates a folder without description

- **WHEN** an authenticated `POST /folders` request omits `description`
- **THEN** the response is `201 Created`
- **AND** `description` in the body is `null`

#### Scenario: Request with missing name is rejected

- **WHEN** an authenticated `POST /folders` request is made with an empty or missing `name`
- **THEN** the response is `400 Bad Request`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `POST /folders` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: GET /folders — List Folders

The system SHALL expose a `GET /folders` endpoint on the protected route group that returns all folders owned by the authenticated user.

Response body (200):
```json
[{ "id": "uuid", "name": "string", "description": "string|null", "parent_id": "uuid|null", "created_at": "timestamp", "updated_at": "timestamp" }]
```

- Returns a flat list of all folders for the user (no nesting in the response)
- Returns an empty array if the user has no folders

#### Scenario: Authenticated user lists their folders

- **WHEN** an authenticated `GET /folders` request is made
- **THEN** the response is `200 OK`
- **AND** each folder object includes a `description` field (null when not set)

#### Scenario: User with no folders receives empty array

- **WHEN** an authenticated `GET /folders` request is made by a user with no folders
- **THEN** the response is `200 OK`
- **AND** the body is an empty array

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /folders` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: GET /folders/:id — Get Folder

The system SHALL expose a `GET /folders/:id` endpoint on the protected route group that returns a single folder by ID, including its image count.

Response body (200):
```json
{
  "id": "uuid",
  "name": "string",
  "description": "string|null",
  "parent_id": "uuid|null",
  "image_count": "integer",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

- The folder MUST be owned by the authenticated user
- `image_count` is the count of non-deleted images whose `folder_id` matches this folder's ID
- Returns `404 Not Found` if the folder does not exist or belongs to another user

#### Scenario: Authenticated user retrieves their folder

- **WHEN** an authenticated `GET /folders/:id` request is made for a folder owned by the user
- **THEN** the response is `200 OK`
- **AND** the body includes `description` and `image_count`

#### Scenario: image_count reflects non-deleted images only

- **WHEN** a folder has 3 images, one of which is soft-deleted
- **THEN** `GET /folders/:id` returns `image_count: 2`

#### Scenario: Folder not found or not owned by user

- **WHEN** an authenticated `GET /folders/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /folders/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: PUT /folders/:id — Update Folder

The system SHALL expose a `PUT /folders/:id` endpoint on the protected route group that updates a folder's `name`, `parent_id`, and/or `description`.

Request body:
```json
{ "name": "string (required)", "parent_id": "uuid|null (optional)", "description": "string|null (optional)" }
```

Response body (200): updated folder in the same shape as `GET /folders` list item (with `description`, without `image_count`).

- The folder MUST be owned by the authenticated user
- If `parent_id` is provided, the referenced parent folder MUST be owned by the same user
- `name` is required and MUST NOT be empty
- Setting `description` to `null` clears it

#### Scenario: Authenticated user updates folder with description

- **WHEN** an authenticated `PUT /folders/:id` request is made with a new valid `name` and a `description`
- **THEN** the response is `200 OK`
- **AND** the body contains the folder with updated `name` and `description`

#### Scenario: Authenticated user clears description

- **WHEN** an authenticated `PUT /folders/:id` request sets `description` to `null`
- **THEN** the response is `200 OK`
- **AND** the folder's `description` is NULL

#### Scenario: Folder not found or not owned by user

- **WHEN** an authenticated `PUT /folders/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Request with missing name is rejected

- **WHEN** an authenticated `PUT /folders/:id` request is made with an empty or missing `name`
- **THEN** the response is `400 Bad Request`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `PUT /folders/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: Folder Repository Interface

The system SHALL define a `FolderRepository` interface in the `usecase` package that the SQL repository implements.

Methods required:
- `Create(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Folder, error)`
- `Update(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `DeleteWithCascade(ctx, id uuid.UUID, userID string) error`
- `FindByName(ctx, userID, name string) (*domain.Folder, error)`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `folderRepository` in the `repository` package implements `usecase.FolderRepository` without compilation errors

---

### Requirement: Folder Usecase Interface

The system SHALL define a `FolderUsecase` interface in the `usecase` package. `GetByID` SHALL return a `FolderDetail` struct that includes the folder and its image count.

```go
type FolderDetail struct {
    Folder     *domain.Folder
    ImageCount int64
}
```

Interface methods:
- `Create(ctx, userID, name string, parentID *uuid.UUID, description *string) (*domain.Folder, error)`
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*FolderDetail, error)`
- `Update(ctx, id uuid.UUID, userID, name string, parentID *uuid.UUID, description *string) (*domain.Folder, error)`
- `Delete(ctx, id uuid.UUID, userID string) error`

`folderUsecase` SHALL receive an `ImageRepository` as a constructor dependency so `GetByID` can call `imageRepo.CountByFolderID`.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `folderUsecase` implements `usecase.FolderUsecase` without compilation errors

---

### Requirement: Folder Usecase Unit Tests

The system SHALL have unit tests for `folderUsecase` covering each method with a mocked `FolderRepository` and mocked `ImageRepository`. Each method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: GetByID unit test covers happy path with image count

- **WHEN** the mocked folder repo returns a folder and the mocked image repo returns a count
- **THEN** the test asserts `FolderDetail` contains both the folder and the correct `ImageCount`

#### Scenario: Usecase unit tests cover repository failure

- **WHEN** the usecase method is called and the mock repository returns an error
- **THEN** the test asserts the error is propagated

---

### Requirement: Folder Handler Unit Tests

The system SHALL have unit tests for `FolderHandler` covering each handler method with a mocked `FolderUsecase`. Each handler method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: GetFolder handler test asserts image_count in response

- **WHEN** the handler is called with a valid request and the mock usecase returns a `FolderDetail` with `ImageCount: 3`
- **THEN** the test asserts the response body includes `"image_count": 3`

#### Scenario: Handler unit tests cover usecase failure

- **WHEN** the handler is called and the mock usecase returns an error
- **THEN** the test asserts the appropriate HTTP error status code is returned
