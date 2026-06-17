## ADDED Requirements

### Requirement: POST /folders — Create Folder

The system SHALL expose a `POST /folders` endpoint on the protected route group that creates a new folder owned by the authenticated user.

Request body:
```json
{ "name": "string (required)", "parent_id": "uuid (optional)", "description": "string (optional)", "icon": "string (optional)" }
```

Response body (201):
```json
{ "id": "uuid", "name": "string", "description": "string|null", "icon": "string|null", "parent_id": "uuid|null", "created_at": "timestamp", "updated_at": "timestamp" }
```

- `parent_id` in the request is optional; omitting it creates a root-level folder
- If `parent_id` is provided, the referenced folder MUST be owned by the authenticated user
- `name` is required and MUST NOT be empty
- `description` is optional; omitting it stores NULL
- `icon` is optional; omitting it stores NULL (the default icon is used). If provided, it MUST be a key in the folder icon allowlist.

#### Scenario: Authenticated user creates a folder with description

- **WHEN** an authenticated `POST /folders` request is made with a valid `name` and a `description`
- **THEN** the response is `201 Created`
- **AND** the body contains the new folder with the supplied `description`

#### Scenario: Authenticated user creates a folder without description

- **WHEN** an authenticated `POST /folders` request omits `description`
- **THEN** the response is `201 Created`
- **AND** `description` in the body is `null`

#### Scenario: Authenticated user creates a folder with an icon

- **WHEN** an authenticated `POST /folders` request is made with a valid `name` and an allowlisted `icon`
- **THEN** the response is `201 Created`
- **AND** the body contains the new folder with the supplied `icon`

#### Scenario: Authenticated user creates a folder without an icon

- **WHEN** an authenticated `POST /folders` request omits `icon`
- **THEN** the response is `201 Created`
- **AND** `icon` in the body is `null`

#### Scenario: Request with a non-allowlisted icon is rejected

- **WHEN** an authenticated `POST /folders` request is made with an `icon` value not in the allowlist
- **THEN** the response is `400 Bad Request`
- **AND** no folder is created

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
[{ "id": "uuid", "name": "string", "description": "string|null", "icon": "string|null", "parent_id": "uuid|null", "created_at": "timestamp", "updated_at": "timestamp" }]
```

- Returns a flat list of all folders for the user (no nesting in the response)
- Returns an empty array if the user has no folders

#### Scenario: Authenticated user lists their folders

- **WHEN** an authenticated `GET /folders` request is made
- **THEN** the response is `200 OK`
- **AND** each folder object includes `description` and `icon` fields (null when not set)

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
  "icon": "string|null",
  "parent_id": "uuid|null",
  "image_count": "integer",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

- The folder MUST be owned by the authenticated user
- `image_count` is the count of non-deleted images with a row in `image_folders` for this folder's ID
- Returns `404 Not Found` if the folder does not exist or belongs to another user

#### Scenario: Authenticated user retrieves their folder

- **WHEN** an authenticated `GET /folders/:id` request is made for a folder owned by the user
- **THEN** the response is `200 OK`
- **AND** the body includes `description`, `icon`, and `image_count`

#### Scenario: image_count reflects non-deleted images only

- **WHEN** a folder has 3 images in `image_folders`, one of which is soft-deleted in `images`
- **THEN** `GET /folders/:id` returns `image_count: 2`

#### Scenario: Folder not found or not owned by user

- **WHEN** an authenticated `GET /folders/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /folders/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: PATCH /folders/:id — Update Folder

The system SHALL expose a `PATCH /folders/:id` endpoint on the protected route group that partially updates a folder's `name`, `parent_id`, `description`, and/or `icon`. Only fields present in the request body are modified — omitted fields are left unchanged. This replaces the prior `PUT /folders/:id` full-replace contract, which silently nulled any field absent from the request body.

Request body (all fields optional; presence, not just value, is significant):
```json
{ "name": "string (optional)", "parent_id": "uuid|null (optional)", "description": "string|null (optional)", "icon": "string|null (optional)" }
```

Response body (200): updated folder in the same shape as `GET /folders` list item (with `description` and `icon`, without `image_count`).

- The folder MUST be owned by the authenticated user
- If `name` is present in the request, it MUST NOT be empty or whitespace-only; if `name` is absent, the folder's existing `name` is preserved unchanged
- If `parent_id` is present in the request (including explicit `null`), the folder's `parent_id` is updated to that value; if `parent_id` is absent, the folder's existing `parent_id` is preserved unchanged
- If `parent_id` is present as a non-null value, the referenced parent folder MUST be owned by the same user
- If `description` is present in the request (including explicit `null`), the folder's `description` is updated to that value — `null` clears it; if `description` is absent, the folder's existing `description` is preserved unchanged
- If `icon` is present in the request (including explicit `null`), the folder's `icon` is updated to that value — `null` resets it to the default icon; if `icon` is absent, the folder's existing `icon` is preserved unchanged. If `icon` is present as a non-null value, it MUST be a key in the folder icon allowlist.

#### Scenario: Updating only the name preserves parent, description, and icon

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "name": "<new name>" }` for a folder that has a non-null `parent_id`, `description`, and `icon`
- **THEN** the response is `200 OK`
- **AND** the folder's `name` is updated
- **AND** the folder's `parent_id`, `description`, and `icon` are unchanged

#### Scenario: Updating only the parent preserves name, description, and icon

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "parent_id": "<uuid>" }` (or `{ "parent_id": null }`) for a folder that has a non-empty `name`, `description`, and `icon`
- **THEN** the response is `200 OK`
- **AND** the folder's `parent_id` is updated to the provided value
- **AND** the folder's `name`, `description`, and `icon` are unchanged

#### Scenario: Updating only the description preserves name, parent, and icon

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "description": "<text>" }` (or `{ "description": null }`) for a folder that has a non-empty `name`, a non-null `parent_id`, and a non-null `icon`
- **THEN** the response is `200 OK`
- **AND** the folder's `description` is updated to the provided value (or cleared, if `null`)
- **AND** the folder's `name`, `parent_id`, and `icon` are unchanged

#### Scenario: Updating only the icon preserves name, parent, and description

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "icon": "<allowlisted key>" }` (or `{ "icon": null }`) for a folder that has a non-empty `name`, a non-null `parent_id`, and a non-null `description`
- **THEN** the response is `200 OK`
- **AND** the folder's `icon` is updated to the provided value (or reset to default, if `null`)
- **AND** the folder's `name`, `parent_id`, and `description` are unchanged

#### Scenario: Updating to a non-allowlisted icon is rejected

- **WHEN** an authenticated `PATCH /folders/:id` request is made with `icon` present and set to a value not in the allowlist
- **THEN** the response is `400 Bad Request`
- **AND** no field on the folder is modified

#### Scenario: Authenticated user updates multiple fields at once

- **WHEN** an authenticated `PATCH /folders/:id` request is made with `{ "name": "<new name>", "description": "<text>", "icon": "<allowlisted key>" }`
- **THEN** the response is `200 OK`
- **AND** the body contains the folder with `name`, `description`, and `icon` all updated
- **AND** the folder's `parent_id` is unchanged

#### Scenario: Folder not found or not owned by user

- **WHEN** an authenticated `PATCH /folders/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Request with a blank name is rejected

- **WHEN** an authenticated `PATCH /folders/:id` request is made with `name` present but empty or whitespace-only
- **THEN** the response is `400 Bad Request`
- **AND** no field on the folder is modified

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `PATCH /folders/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: DELETE /folders/:id — Delete Folder

The system SHALL expose a `DELETE /folders/:id` endpoint on the protected route group that permanently deletes a folder and handles cascading side effects.

Response: `204 No Content` on success.

Cascade behaviour on delete (all steps executed in a single transaction):
1. Child folders (folders whose `parent_id` equals the deleted folder's ID) SHALL have their `parent_id` set to `null` via an explicit `UPDATE` in the repository
2. The folder row is then permanently deleted
3. Rows in `image_folders` where `folder_id` equals the deleted folder's ID are removed automatically by the database `ON DELETE CASCADE` constraint — no explicit repository step is needed

The FK constraint on `image_folders.folder_id` is `ON DELETE CASCADE`. The constraint on `folders.parent_id` remains `ON DELETE RESTRICT` — step 1 is still required before deletion.

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

#### Scenario: Deleting a folder removes image_folders memberships via cascade

- **WHEN** an authenticated `DELETE /folders/:id` request is made for a folder that has images assigned to it
- **THEN** all rows in `image_folders` referencing the deleted folder are removed
- **AND** those images become unfiled (no folder membership)
- **AND** the image records themselves are not deleted

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
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Folder, error)` — performs a selective column update using only the keys present in `fields` (e.g. `Model(&domain.Folder{}).Where("id = ? AND user_id = ?", id, userID).Updates(fields)`), returns `gorm.ErrRecordNotFound` wrapped when no row matches, and re-fetches the updated folder on success. This mirrors `imageRepository.Update` and replaces the prior full-row-replace implementation that overwrote `name`, `parent_id`, and `description` unconditionally. `fields` MAY include an `icon` key.
- `DeleteWithCascade(ctx, id uuid.UUID, userID string) error` — in a single transaction: nulls child folders' `parent_id`, then hard-deletes the folder row; the `image_folders` cleanup is handled by `ON DELETE CASCADE` and does NOT require an explicit step
- `FindByName(ctx, userID, name string) (*domain.Folder, error)`
- `CountImagesByFolder(ctx, id uuid.UUID, userID string) (int, error)` — counts non-deleted images with a row in `image_folders` for the given folder; implemented as `Model(&domain.Image{}) + JOIN image_folders WHERE image_folders.folder_id = ? AND images.user_id = ?`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `folderRepository` in the `repository` package implements `usecase.FolderRepository` without compilation errors

#### Scenario: Update modifies only the supplied fields

- **WHEN** `Update` is called with `fields` containing only `{"name": "<new name>"}` for a folder that has a non-null `parent_id`, `description`, and `icon`
- **THEN** only the `name` column is written to the database
- **AND** the folder's `parent_id`, `description`, and `icon` columns retain their prior values

#### Scenario: Update writes the icon column when supplied

- **WHEN** `Update` is called with `fields` containing `{"icon": "<allowlisted key>"}`
- **THEN** only the `icon` column is written to the database

#### Scenario: Update returns not-found for a missing or unowned folder

- **WHEN** `Update` is called with an `id`/`userID` pair that matches no row
- **THEN** the call returns an error wrapping `gorm.ErrRecordNotFound`
- **AND** no row is modified

#### Scenario: DeleteWithCascade does not explicitly update images table

- **WHEN** `DeleteWithCascade` is called for a folder that has images
- **THEN** the transaction only nulls child folder `parent_id` values and deletes the folder row
- **AND** `image_folders` rows are removed by the database cascade without an explicit UPDATE or DELETE statement in the repository

#### Scenario: CountImagesByFolder excludes soft-deleted images

- **WHEN** `CountImagesByFolder` is called for a folder with 3 images of which 1 is soft-deleted
- **THEN** the count returned is `2`

---

### Requirement: Folder Usecase Interface

The system SHALL define a `FolderUsecase` interface in the `usecase` package. `GetByID` SHALL return a `FolderDetail` struct that includes the folder and its image count.

```go
type FolderDetail struct {
    Folder     *domain.Folder
    ImageCount int64
}
```

`Update` SHALL accept a params struct that distinguishes "field omitted" from "field explicitly set to null" from "field set to a value", mirroring `usecase.UpdateImageParams`:

```go
type UpdateFolderParams struct {
    Name        *string
    ParentID    **uuid.UUID
    Description **string
    Icon        **string
}
```

Interface methods:
- `Create(ctx, userID, name string, parentID *uuid.UUID, description *string, icon *string) (*domain.Folder, error)` — `icon`, if non-nil, MUST be validated against the folder icon allowlist before being passed to the repository
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*FolderDetail, error)`
- `Update(ctx, id uuid.UUID, userID string, params UpdateFolderParams) (*domain.Folder, error)` — validates that `Name`, if non-nil, is non-blank; validates that `Icon`, if non-nil and pointing to a non-nil value, is a key in the folder icon allowlist; builds a selective field map containing only the params that are non-nil and passes it to `folderRepo.Update`; replaces the prior `Update(ctx, id, userID, name string, parentID *uuid.UUID, description *string)` signature, which always wrote all three fields and did not support `icon`
- `Delete(ctx, id uuid.UUID, userID string) error`

`folderUsecase` SHALL receive an `ImageRepository` as a constructor dependency so `GetByID` can call `imageRepo.CountByFolderID`.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `folderUsecase` implements `usecase.FolderUsecase` without compilation errors

#### Scenario: Update validates a present-but-blank name

- **WHEN** `Update` is called with `params.Name` pointing to an empty or whitespace-only string
- **THEN** the usecase returns `ErrInvalidFolderName`
- **AND** the repository's `Update` is not called

#### Scenario: Update validates a non-allowlisted icon

- **WHEN** `Update` is called with `params.Icon` pointing to a non-nil value not present in the folder icon allowlist
- **THEN** the usecase returns an invalid-icon error
- **AND** the repository's `Update` is not called

#### Scenario: Update passes through only provided fields

- **WHEN** `Update` is called with `params.Name` set and `params.ParentID`/`params.Description`/`params.Icon` left `nil`
- **THEN** the usecase calls `folderRepo.Update` with a fields map containing only the `name` key

---

### Requirement: Folder Routes Wiring

The system SHALL register folder routes on the protected Echo group in `main.go`, using the same auth middleware already applied to `/me`.

Routes:
- `POST /folders`
- `GET /folders`
- `GET /folders/:id`
- `PATCH /folders/:id`
- `DELETE /folders/:id`

#### Scenario: Folder routes are registered under auth middleware

- **WHEN** the server starts
- **THEN** all `/folders` routes require a valid Kinde Bearer token
- **AND** unauthenticated requests to any `/folders` route return `401 Unauthorized`

---

### Requirement: Folder Usecase Unit Tests

The system SHALL have unit tests for `folderUsecase` covering each method with a mocked `FolderRepository` and mocked `ImageRepository`. Each method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: GetByID unit test covers happy path with image count

- **WHEN** the mocked folder repo returns a folder and the mocked image repo returns a count
- **THEN** the test asserts `FolderDetail` contains both the folder and the correct `ImageCount`

#### Scenario: Usecase unit tests cover repository failure

- **WHEN** the usecase method is called and the mock repository returns an error
- **THEN** the test asserts the error is propagated

#### Scenario: Update unit test covers icon allowlist rejection

- **WHEN** `Update` is called with a non-allowlisted `Icon` value
- **THEN** the test asserts an invalid-icon error is returned and the mocked repository's `Update` is not called

---

### Requirement: Folder Handler Unit Tests

The system SHALL have unit tests for `FolderHandler` covering each handler method with a mocked `FolderUsecase`. Each handler method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: GetFolder handler test asserts image_count in response

- **WHEN** the handler is called with a valid request and the mock usecase returns a `FolderDetail` with `ImageCount: 3`
- **THEN** the test asserts the response body includes `"image_count": 3`

#### Scenario: Handler unit tests cover usecase failure

- **WHEN** the handler is called and the mock usecase returns an error
- **THEN** the test asserts the appropriate HTTP error status code is returned

---

### Requirement: Folder Repository Integration Tests

The system SHALL have integration tests for `folderRepository` using Testcontainers (following the existing pattern in `internal/repository/`). Each repository method SHALL be tested against a real database. Unit tests SHALL NOT be written for the SQL repository.

#### Scenario: Repository integration tests exercise each method against a real database

- **WHEN** the integration test suite runs with a live PostgreSQL container
- **THEN** each `FolderRepository` method (`Create`, `List`, `GetByID`, `Update`, `DeleteWithCascade`) is exercised with at least one success scenario and one failure scenario
