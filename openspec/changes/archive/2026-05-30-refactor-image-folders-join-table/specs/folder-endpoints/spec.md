## MODIFIED Requirements

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
- `image_count` is the count of non-deleted images with a row in `image_folders` for this folder's ID
- Returns `404 Not Found` if the folder does not exist or belongs to another user

#### Scenario: Authenticated user retrieves their folder

- **WHEN** an authenticated `GET /folders/:id` request is made for a folder owned by the user
- **THEN** the response is `200 OK`
- **AND** the body includes `description` and `image_count`

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

### Requirement: Folder Repository Interface

The system SHALL define a `FolderRepository` interface in the `usecase` package that the SQL repository implements.

Methods required:
- `Create(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `List(ctx, userID string) ([]*domain.Folder, error)`
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Folder, error)`
- `Update(ctx, folder *domain.Folder) (*domain.Folder, error)`
- `DeleteWithCascade(ctx, id uuid.UUID, userID string) error` — in a single transaction: nulls child folders' `parent_id`, then hard-deletes the folder row; the `image_folders` cleanup is handled by `ON DELETE CASCADE` and does NOT require an explicit step
- `FindByName(ctx, userID, name string) (*domain.Folder, error)`
- `CountImagesByFolder(ctx, id uuid.UUID, userID string) (int, error)` — counts non-deleted images with a row in `image_folders` for the given folder; implemented as `Model(&domain.Image{}) + JOIN image_folders WHERE image_folders.folder_id = ? AND images.user_id = ?`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `folderRepository` in the `repository` package implements `usecase.FolderRepository` without compilation errors

#### Scenario: DeleteWithCascade does not explicitly update images table

- **WHEN** `DeleteWithCascade` is called for a folder that has images
- **THEN** the transaction only nulls child folder `parent_id` values and deletes the folder row
- **AND** `image_folders` rows are removed by the database cascade without an explicit UPDATE or DELETE statement in the repository

#### Scenario: CountImagesByFolder excludes soft-deleted images

- **WHEN** `CountImagesByFolder` is called for a folder with 3 images of which 1 is soft-deleted
- **THEN** the count returned is `2`
