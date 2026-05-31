## Purpose

Defines the `PATCH /images/:id/position` endpoint and the `UpdateImagePosition` usecase method for reordering images within a folder using fractional index (fracdex) positioning.

## Requirements

### Requirement: PATCH /images/:id/position Endpoint

The system SHALL expose `PATCH /images/:id/position` on the protected router. This endpoint updates the fractional index position of an image within a specific folder.

Request body:
```json
{
  "folder_id": "uuid",
  "position": "a1V"
}
```

- `folder_id` (required): the folder in which the position is being set
- `position` (required): a non-empty fracdex key string computed by the caller

The handler SHALL:
1. Parse and validate `:id` as a UUID; return `400` if invalid
2. Validate that `folder_id` is a non-empty valid UUID; return `400` if missing or invalid
3. Validate that `position` is a non-empty string; return `400` if empty
4. Resolve the authenticated `userID` from context
5. Delegate to `imageUsecase.UpdateImagePosition(ctx, imageID, userID, folderID, position)`
6. Return `204 No Content` on success

The handler SHALL NOT validate the fracdex key format beyond requiring it to be non-empty. Key correctness is the caller's responsibility.

#### Scenario: Valid reorder request returns 204

- **WHEN** an authenticated `PATCH /images/:id/position` request is sent with a valid `folder_id` and non-empty `position`
- **THEN** the response is `204 No Content`
- **AND** `image_folders.position` is updated for that `(image_id, folder_id)` pair

#### Scenario: Missing position field returns 400

- **WHEN** `PATCH /images/:id/position` is sent with an empty or absent `position`
- **THEN** the response is `400 Bad Request`

#### Scenario: Missing folder_id returns 400

- **WHEN** `PATCH /images/:id/position` is sent without a valid `folder_id`
- **THEN** the response is `400 Bad Request`

#### Scenario: Image not in specified folder returns 404

- **WHEN** `PATCH /images/:id/position` is sent with a `folder_id` for which the image has no membership
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request returns 401

- **WHEN** `PATCH /images/:id/position` is called without a valid auth token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: UpdateImagePosition Usecase Method

The `ImageUsecase` interface SHALL add:

```go
UpdateImagePosition(ctx context.Context, imageID uuid.UUID, userID string, folderID uuid.UUID, position string) error
```

The implementation SHALL:
1. Verify the image exists and belongs to `userID` via `imageRepo.GetByID`; return a not-found error if not
2. Call `imageRepo.UpdateImageFolderPosition(ctx, imageID, folderID, position)`
3. Return any repository error directly

No server-side position computation is performed. The caller supplies the final fracdex key.

#### Scenario: Successful position update

- **WHEN** `UpdateImagePosition` is called with a valid image ID owned by the user, a folder the image belongs to, and a non-empty position string
- **THEN** the repository position is updated and no error is returned

#### Scenario: Image not found returns error

- **WHEN** `UpdateImagePosition` is called with an image ID that does not exist or belongs to another user
- **THEN** a not-found error is returned and no update occurs
