## ADDED Requirements

### Requirement: POST /folders — Create Folder

The system SHALL expose a `POST /folders` endpoint on the protected route group that creates a new folder owned by the authenticated user.

Request body:
```json
{ "name": "string (required)", "parent_id": "uuid (optional)" }
```

Response body (201):
```json
{ "id": "uuid", "name": "string", "parent_id": "uuid|null", "created_at": "timestamp", "updated_at": "timestamp" }
```

- `parent_id` in the request is optional; omitting it creates a root-level folder
- If `parent_id` is provided, the referenced folder MUST be owned by the authenticated user
- `name` is required and MUST NOT be empty

#### Scenario: Authenticated user creates a root folder

- **WHEN** an authenticated `POST /folders` request is made with a valid `name` and no `parent_id`
- **THEN** the response is `201 Created`
- **AND** the body contains the new folder with `parent_id` as null

#### Scenario: Authenticated user creates a nested folder

- **WHEN** an authenticated `POST /folders` request is made with a valid `name` and a `parent_id` that belongs to the same user
- **THEN** the response is `201 Created`
- **AND** the body contains the new folder with the given `parent_id`

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
[{ "id": "uuid", "name": "string", "parent_id": "uuid|null", "created_at": "timestamp", "updated_at": "timestamp" }]
```

- Returns a flat list of all folders for the user (no nesting in the response)
- Returns an empty array if the user has no folders

#### Scenario: Authenticated user lists their folders

- **WHEN** an authenticated `GET /folders` request is made
- **THEN** the response is `200 OK`
- **AND** the body is an array of all folders owned by the authenticated user

#### Scenario: User with no folders receives empty array

- **WHEN** an authenticated `GET /folders` request is made by a user with no folders
- **THEN** the response is `200 OK`
- **AND** the body is an empty array

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /folders` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: GET /folders/:id — Get Folder

The system SHALL expose a `GET /folders/:id` endpoint on the protected route group that returns a single folder by ID.

Response body (200): same shape as a single item from the list response.

- The folder MUST be owned by the authenticated user
- Returns `404 Not Found` if the folder does not exist or belongs to another user

#### Scenario: Authenticated user retrieves their folder

- **WHEN** an authenticated `GET /folders/:id` request is made with a valid folder ID owned by the user
- **THEN** the response is `200 OK`
- **AND** the body contains the folder data

#### Scenario: Folder not found or not owned by user

- **WHEN** an authenticated `GET /folders/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /folders/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: PUT /folders/:id — Update Folder

The system SHALL expose a `PUT /folders/:id` endpoint on the protected route group that updates a folder's `name` and/or `parent_id`.

Request body:
```json
{ "name": "string (required)", "parent_id": "uuid|null (optional)" }
```

Response body (200): updated folder in the same shape as GET.

- The folder MUST be owned by the authenticated user
- If `parent_id` is provided, the referenced parent folder MUST be owned by the same user
- `name` is required and MUST NOT be empty

#### Scenario: Authenticated user updates folder name

- **WHEN** an authenticated `PUT /folders/:id` request is made with a new valid `name`
- **THEN** the response is `200 OK`
- **AND** the body contains the folder with the updated name

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

### Requirement: DELETE /folders/:id — Delete Folder

The system SHALL expose a `DELETE /folders/:id` endpoint on the protected route group that permanently deletes a folder and handles cascading side effects.

Response: `204 No Content` on success.

Cascade behaviour on delete (all steps executed in a single transaction):
1. Child folders (folders whose `parent_id` equals the deleted folder's ID) SHALL have their `parent_id` set to `null` via an explicit `UPDATE` in the repository
2. Images whose `folder_id` equals the deleted folder's ID SHALL have their `folder_id` set to `null` via an explicit `UPDATE` in the repository
3. The folder row is then permanently deleted

Both FK constraints (`folders.parent_id` and `images.folder_id`) are `ON DELETE RESTRICT` — the DB will reject the delete if the application skips either cascade step.

- The folder MUST be owned by the authenticated user
- Returns `404 Not Found` if the folder does not exist or belongs to another user
- There is no soft delete; the row is permanently removed

#### Scenario: Authenticated user deletes a folder with no children

- **WHEN** an authenticated `DELETE /folders/:id` request is made for a folder with no child folders
- **THEN** the response is `204 No Content`
- **AND** the folder row is permanently removed from the database

#### Scenario: Deleting a folder orphans its child folders to root

- **WHEN** an authenticated `DELETE /folders/:id` request is made for a folder that has child folders
- **THEN** the response is `204 No Content`
- **AND** those child folders have their `parent_id` set to `null`
- **AND** the deleted folder row is permanently removed

#### Scenario: Folder not found or not owned by user

- **WHEN** an authenticated `DELETE /folders/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `DELETE /folders/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: Folder Repository Interface

The system SHALL define a `FolderRepository` interface in the `usecase` package that the SQL repository implements, following the same pattern as `UserRepository`.

Methods required:
- `Create(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Folder, error)`
- `Update(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `DeleteWithCascade(ctx, id uuid.UUID, userID string) error` — in a single transaction: nulls child folders' `parent_id`, nulls images' `folder_id`, then hard-deletes the folder row

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `folderRepository` in the `repository` package implements `usecase.FolderRepository` without compilation errors

---

### Requirement: Folder Usecase Interface

The system SHALL define a `FolderUsecase` interface in the `usecase` package with methods corresponding to the five CRUD operations.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `folderUsecase` implements `usecase.FolderUsecase` without compilation errors

---

### Requirement: Folder Routes Wiring

The system SHALL register folder routes on the protected Echo group in `main.go`, using the same auth middleware already applied to `/me`.

Routes:
- `POST /folders`
- `GET /folders`
- `GET /folders/:id`
- `PUT /folders/:id`
- `DELETE /folders/:id`

#### Scenario: Folder routes are registered under auth middleware

- **WHEN** the server starts
- **THEN** all `/folders` routes require a valid Kinde Bearer token
- **AND** unauthenticated requests to any `/folders` route return `401 Unauthorized`

---

### Requirement: Folder Usecase Unit Tests

The system SHALL have unit tests for `folderUsecase` covering each method with a mocked `FolderRepository`. Each method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: Usecase unit tests cover the happy path

- **WHEN** the usecase method is called with valid inputs and the mock repository returns successfully
- **THEN** the test asserts the correct result is returned without error

#### Scenario: Usecase unit tests cover repository failure

- **WHEN** the usecase method is called and the mock repository returns an error
- **THEN** the test asserts the error is propagated

---

### Requirement: Folder Handler Unit Tests

The system SHALL have unit tests for `FolderHandler` covering each handler method with a mocked `FolderUsecase`. Each handler method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: Handler unit tests cover the happy path

- **WHEN** the handler is called with a valid request and the mock usecase returns successfully
- **THEN** the test asserts the correct HTTP status code and response body

#### Scenario: Handler unit tests cover usecase failure

- **WHEN** the handler is called and the mock usecase returns an error
- **THEN** the test asserts the appropriate HTTP error status code is returned

---

### Requirement: Folder Repository Integration Tests

The system SHALL have integration tests for `folderRepository` using Testcontainers (following the existing pattern in `internal/repository/`). Each repository method SHALL be tested against a real database. Unit tests SHALL NOT be written for the SQL repository.

#### Scenario: Repository integration tests exercise each method against a real database

- **WHEN** the integration test suite runs with a live PostgreSQL container
- **THEN** each `FolderRepository` method (`Create`, `List`, `GetByID`, `Update`, `DeleteWithCascade`) is exercised with at least one success scenario and one failure scenario
