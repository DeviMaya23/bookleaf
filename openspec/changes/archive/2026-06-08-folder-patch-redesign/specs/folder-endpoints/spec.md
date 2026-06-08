## MODIFIED Requirements

### Requirement: PATCH /folders/:id — Update Folder

The system SHALL expose a `PATCH /folders/:id` endpoint on the protected route group that partially updates a folder's `name`, `parent_id`, and/or `description`. Only fields present in the request body are modified — omitted fields are left unchanged. This replaces the prior `PUT /folders/:id` full-replace contract, which silently nulled any field absent from the request body.

Request body (all fields optional; presence, not just value, is significant):
```json
{ "name": "string (optional)", "parent_id": "uuid|null (optional)", "description": "string|null (optional)" }
```

Response body (200): updated folder in the same shape as `GET /folders` list item (with `description`, without `image_count`).

- The folder MUST be owned by the authenticated user
- If `name` is present in the request, it MUST NOT be empty or whitespace-only; if `name` is absent, the folder's existing `name` is preserved unchanged
- If `parent_id` is present in the request (including explicit `null`), the folder's `parent_id` is updated to that value; if `parent_id` is absent, the folder's existing `parent_id` is preserved unchanged
- If `parent_id` is present as a non-null value, the referenced parent folder MUST be owned by the same user
- If `description` is present in the request (including explicit `null`), the folder's `description` is updated to that value — `null` clears it; if `description` is absent, the folder's existing `description` is preserved unchanged

#### Scenario: Updating only the name preserves parent and description

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "name": "<new name>" }` for a folder that has a non-null `parent_id` and `description`
- **THEN** the response is `200 OK`
- **AND** the folder's `name` is updated
- **AND** the folder's `parent_id` and `description` are unchanged

#### Scenario: Updating only the parent preserves name and description

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "parent_id": "<uuid>" }` (or `{ "parent_id": null }`) for a folder that has a non-empty `name` and `description`
- **THEN** the response is `200 OK`
- **AND** the folder's `parent_id` is updated to the provided value
- **AND** the folder's `name` and `description` are unchanged

#### Scenario: Updating only the description preserves name and parent

- **WHEN** an authenticated `PATCH /folders/:id` request is made with only `{ "description": "<text>" }` (or `{ "description": null }`) for a folder that has a non-empty `name` and a non-null `parent_id`
- **THEN** the response is `200 OK`
- **AND** the folder's `description` is updated to the provided value (or cleared, if `null`)
- **AND** the folder's `name` and `parent_id` are unchanged

#### Scenario: Authenticated user updates multiple fields at once

- **WHEN** an authenticated `PATCH /folders/:id` request is made with `{ "name": "<new name>", "description": "<text>" }`
- **THEN** the response is `200 OK`
- **AND** the body contains the folder with both `name` and `description` updated
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

### Requirement: Folder Repository Interface

The system SHALL define a `FolderRepository` interface in the `usecase` package that the SQL repository implements, following the same pattern as `UserRepository`.

Methods required:
- `Create(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Folder, error)`
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Folder, error)` — performs a selective column update using only the keys present in `fields` (e.g. `Model(&domain.Folder{}).Where("id = ? AND user_id = ?", id, userID).Updates(fields)`), returns `gorm.ErrRecordNotFound` wrapped when no row matches, and re-fetches the updated folder on success. This mirrors `imageRepository.Update` and replaces the prior full-row-replace implementation that overwrote `name`, `parent_id`, and `description` unconditionally.
- `DeleteWithCascade(ctx, id uuid.UUID, userID string) error` — in a single transaction: nulls child folders' `parent_id`, then hard-deletes the folder row; the `image_folders` cleanup is handled by `ON DELETE CASCADE` and does NOT require an explicit step
- `FindByName(ctx, userID, name string) (*domain.Folder, error)`
- `CountImagesByFolder(ctx, id uuid.UUID, userID string) (int, error)` — counts non-deleted images with a row in `image_folders` for the given folder; implemented as `Model(&domain.Image{}) + JOIN image_folders WHERE image_folders.folder_id = ? AND images.user_id = ?`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `folderRepository` in the `repository` package implements `usecase.FolderRepository` without compilation errors

#### Scenario: Update modifies only the supplied fields

- **WHEN** `Update` is called with `fields` containing only `{"name": "<new name>"}` for a folder that has a non-null `parent_id` and `description`
- **THEN** only the `name` column is written to the database
- **AND** the folder's `parent_id` and `description` columns retain their prior values

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
}
```

Interface methods:
- `Create(ctx, userID, name string, parentID *uuid.UUID, description *string) (*domain.Folder, error)`
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*FolderDetail, error)`
- `Update(ctx, id uuid.UUID, userID string, params UpdateFolderParams) (*domain.Folder, error)` — validates that `Name`, if non-nil, is non-blank; builds a selective field map containing only the params that are non-nil and passes it to `folderRepo.Update`; replaces the prior `Update(ctx, id, userID, name string, parentID *uuid.UUID, description *string)` signature, which always wrote all three fields
- `Delete(ctx, id uuid.UUID, userID string) error`

`folderUsecase` SHALL receive an `ImageRepository` as a constructor dependency so `GetByID` can call `imageRepo.CountByFolderID`.

#### Scenario: Update validates a present-but-blank name

- **WHEN** `Update` is called with `params.Name` pointing to an empty or whitespace-only string
- **THEN** the usecase returns `ErrInvalidFolderName`
- **AND** the repository's `Update` is not called

#### Scenario: Update passes through only provided fields

- **WHEN** `Update` is called with `params.Name` set and `params.ParentID`/`params.Description` left `nil`
- **THEN** the usecase calls `folderRepo.Update` with a fields map containing only the `name` key
