## ADDED Requirements

### Requirement: PATCH /images/:id — Edit Image Metadata

The system SHALL expose a `PATCH /images/:id` endpoint on the protected route group that updates the `title` and/or `folder_id` of an existing image. Neither field is required; omitting a field means that field is left unchanged. The image binary (`r2_path`) SHALL NOT be modifiable via this endpoint.

Request body:
```json
{
  "title": "string (optional)",
  "folder_id": "uuid | null (optional)"
}
```

- A `null` `folder_id` in the request body SHALL move the image to the root (clear the folder association).
- An absent `folder_id` field SHALL leave the current `folder_id` unchanged.
- The handler SHALL distinguish absent from null using a presence flag or pointer-of-pointer decoding — not `omitempty` alone.
- `title`, if present, MUST NOT be empty string.
- The image MUST be owned by the authenticated user.
- Returns `404 Not Found` if the image does not exist or belongs to another user.
- Returns `400 Bad Request` if the body is malformed or `title` is an empty string.

Response body (200):
```json
{
  "id": "uuid",
  "title": "string",
  "mime_type": "string",
  "source_url": "string|null",
  "folder_id": "uuid|null",
  "thumbnail_url": "string|null",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

#### Scenario: Title is updated

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"title": "new name"}`
- **THEN** the response is `200 OK`
- **AND** the returned image has `title` set to `"new name"`
- **AND** the image's `folder_id` is unchanged

#### Scenario: Image is moved to a folder

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"folder_id": "<uuid>"}`
- **THEN** the response is `200 OK`
- **AND** the returned image has `folder_id` set to the provided UUID
- **AND** the image's `title` is unchanged

#### Scenario: Image is moved to root with null folder_id

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"folder_id": null}`
- **THEN** the response is `200 OK`
- **AND** the returned image has `folder_id` set to `null`

#### Scenario: Empty title is rejected

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"title": ""}`
- **THEN** the response is `400 Bad Request`

#### Scenario: Image not found or not owned

- **WHEN** an authenticated `PATCH /images/:id` request is made for a non-existent or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `PATCH /images/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: UpdateImage Usecase Method

The `ImageUsecase` interface SHALL include an `UpdateImage(ctx, id uuid.UUID, userID string, params UpdateImageParams) (*domain.Image, error)` method.

`UpdateImageParams` SHALL use pointer fields so the usecase can distinguish absent from provided values:
- `Title *string` — nil means unchanged; non-nil means update to this value
- `FolderID **uuid.UUID` — nil outer pointer means unchanged; non-nil outer pointer with nil inner pointer means clear folder (move to root); non-nil inner pointer means set to that UUID

The usecase SHALL:
1. Fetch the existing image by `id` and `userID`; return `gorm.ErrRecordNotFound` if not found
2. Build a map of only the fields that are non-nil in `params`
3. Delegate to `ImageRepository.Update`
4. Emit `image.mutated / moved_to_folder` log when `FolderID` is present in params AND the new value differs from the existing value (see observability-logging spec)

#### Scenario: Only provided fields are updated

- **WHEN** `UpdateImage` is called with `params.Title = nil` and `params.FolderID` pointing to a UUID
- **THEN** only `folder_id` is written to the database
- **AND** `title` retains its previous value

#### Scenario: Not found returns error

- **WHEN** `UpdateImage` is called for an image that does not exist or is owned by another user
- **THEN** the method returns `gorm.ErrRecordNotFound`

---

### Requirement: UpdateImage Repository Method

The `ImageRepository` interface SHALL include an `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` method.

The repository implementation SHALL:
- Use `db.Model(&Image{}).Where("id = ? AND user_id = ?", id, userID).Updates(fields)` to apply only the supplied fields
- Treat `RowsAffected == 0` as `gorm.ErrRecordNotFound`
- Return the updated image record

`gorm.Save` SHALL NOT be used to avoid overwriting unrelated fields such as `r2_path` and `thumbnail_path`.

#### Scenario: Selective field update does not overwrite unrelated fields

- **WHEN** `Update` is called with `fields = {"title": "new"}` on an image that has a non-null `thumbnail_path`
- **THEN** the database row retains its original `thumbnail_path` value

#### Scenario: Update on non-existent image returns not found

- **WHEN** `Update` is called with an `id` that does not exist for the given `userID`
- **THEN** the method returns `gorm.ErrRecordNotFound`

---

### Requirement: UpdateImage Handler Unit Tests

The system SHALL have unit tests for the `UpdateImage` handler method covering at minimum one success scenario and one failure scenario.

#### Scenario: Handler unit tests cover success and failure

- **WHEN** the `UpdateImage` handler is tested with a mock usecase
- **THEN** the success case asserts `200 OK` with the updated image body
- **AND** the failure case asserts the correct error status code (e.g. `404 Not Found`)

---

### Requirement: UpdateImage Usecase Unit Tests

The system SHALL have unit tests for the `UpdateImage` usecase method covering at minimum one success scenario and one failure scenario.

#### Scenario: Usecase unit tests cover success and failure

- **WHEN** the `UpdateImage` usecase is tested with a mocked repository
- **THEN** the success case asserts the returned image reflects the update
- **AND** the failure case asserts the error is propagated correctly

---

### Requirement: UpdateImage Repository Integration Tests

The system SHALL have integration tests for the `Update` repository method using Testcontainers exercising at minimum one success scenario and one failure scenario.

#### Scenario: Integration test for selective update

- **WHEN** the integration test calls `Update` with a single field map against a real PostgreSQL database
- **THEN** only that field is changed in the row
- **AND** all other fields retain their original values
