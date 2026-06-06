## ADDED Requirements

### Requirement: ListAllTrashed retrieves all trashed images for a user without pagination

The system SHALL provide a `ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error)` method on `ImageRepository` that returns all soft-deleted image records for the given user, with no cursor or limit. It SHALL return an empty slice (not an error) when the user has no trashed images.

#### Scenario: Returns all trashed images for the user

- **WHEN** `ListAllTrashed` is called for a user with trashed images
- **THEN** it returns all soft-deleted image records belonging to that user

#### Scenario: Returns empty slice when user has no trashed images

- **WHEN** `ListAllTrashed` is called for a user with no trashed images
- **THEN** it returns an empty slice and no error

---

### Requirement: DeleteFromTrash permanently deletes a single trashed image

The system SHALL provide a `DeleteFromTrash(ctx context.Context, id uuid.UUID, userID string) error` method on `ImageUsecase` that permanently removes a single soft-deleted image owned by the given user.

The method SHALL:
1. Fetch the image via the existing `GetDeletedByID` on `ImageRepository`; return a not-found error if it does not exist in trash
2. Attempt to delete the R2 object at `r2_path` (best-effort; log warn on failure, continue)
3. Attempt to delete the R2 object at `thumbnail_path` if not nil (best-effort; log warn on failure, continue)
4. Hard-delete the DB record

#### Scenario: Successfully deletes a trashed image

- **WHEN** `DeleteFromTrash` is called with a valid ID belonging to a trashed image owned by the user
- **THEN** the R2 object is deleted, the thumbnail R2 object is deleted (if present), and the DB record is permanently removed

#### Scenario: Returns not-found when image is not in trash

- **WHEN** `DeleteFromTrash` is called with an ID that is not in trash or belongs to another user
- **THEN** it returns a not-found error and performs no deletions

#### Scenario: R2 delete failure does not block hard delete

- **WHEN** `DeleteFromTrash` is called and the R2 object deletion fails
- **THEN** the error is logged at warn level and the DB record is still hard-deleted

---

### Requirement: EmptyTrash permanently deletes all trashed images for a user

The system SHALL provide an `EmptyTrash(ctx context.Context, userID string) error` method on `ImageUsecase` that permanently removes all soft-deleted images owned by the given user.

The method SHALL:
1. Fetch all trashed images via `ListAllTrashed`
2. For each image: delete R2 object (best-effort), delete thumbnail if present (best-effort), hard-delete DB record
3. Return nil if the user has no trashed images (no-op)

#### Scenario: Successfully empties all trashed images

- **WHEN** `EmptyTrash` is called for a user with trashed images
- **THEN** all R2 objects are deleted and all trashed DB records are permanently removed

#### Scenario: No trashed images results in no-op

- **WHEN** `EmptyTrash` is called for a user with no trashed images
- **THEN** no deletions are performed and no error is returned

#### Scenario: R2 delete failure does not block remaining deletions

- **WHEN** `EmptyTrash` is called and the R2 delete for one image fails
- **THEN** the error is logged at warn level and the loop continues, deleting remaining images

---

### Requirement: DELETE /images/trash/:id endpoint permanently deletes a single trashed image

The system SHALL expose `DELETE /images/trash/:id` as an authenticated endpoint that calls `DeleteFromTrash` on `ImageUsecase` and returns `204 No Content` on success.

It SHALL return `404 Not Found` if the image is not in the authenticated user's trash.

#### Scenario: Returns 204 on successful deletion

- **WHEN** an authenticated user sends `DELETE /images/trash/:id` with a valid trashed image ID
- **THEN** the response status is `204 No Content` and the image is permanently deleted

#### Scenario: Returns 404 when image is not in trash

- **WHEN** an authenticated user sends `DELETE /images/trash/:id` with an ID not in their trash
- **THEN** the response status is `404 Not Found`

#### Scenario: Returns 400 when ID is not a valid UUID

- **WHEN** an authenticated user sends `DELETE /images/trash/:id` with a malformed ID
- **THEN** the response status is `400 Bad Request`

---

### Requirement: DELETE /images/trash endpoint permanently deletes all trashed images for the user

The system SHALL expose `DELETE /images/trash` as an authenticated endpoint that calls `EmptyTrash` on `ImageUsecase` and returns `204 No Content` on success, including when the user has no trashed images.

#### Scenario: Returns 204 after emptying trash

- **WHEN** an authenticated user sends `DELETE /images/trash` and has trashed images
- **THEN** the response status is `204 No Content` and all trashed images are permanently deleted

#### Scenario: Returns 204 when trash is already empty

- **WHEN** an authenticated user sends `DELETE /images/trash` and has no trashed images
- **THEN** the response status is `204 No Content` and no deletions are performed
