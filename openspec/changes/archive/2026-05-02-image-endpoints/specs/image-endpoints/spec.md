## ADDED Requirements

### Requirement: POST /images — Initiate Upload

The system SHALL expose a `POST /images` endpoint on the protected route group that creates an image record and returns a presigned PUT URL for direct R2 upload.

Request body:
```json
{
  "title": "string (required)",
  "mime_type": "string (required)",
  "source_url": "string (optional)",
  "folder_id": "uuid (optional)"
}
```

Response body (201):
```json
{
  "id": "uuid",
  "upload_url": "string (presigned PUT URL, valid 15 minutes)",
  "r2_path": "string"
}
```

- Backend generates a UUID for the image and the R2 key (`users/{kindeID}/images/{imageID}.{ext}`)
- `title` and `mime_type` are required and MUST NOT be empty
- `thumbnail_path` is null until `POST /images/:id/complete` is called and the goroutine succeeds

#### Scenario: Authenticated user initiates an image upload

- **WHEN** an authenticated `POST /images` request is made with valid `title` and `mime_type`
- **THEN** the response is `201 Created`
- **AND** the body contains a UUID `id`, a presigned PUT `upload_url`, and the `r2_path`
- **AND** the image record is persisted in the database with `thumbnail_path` null

#### Scenario: Request with missing title or mime_type is rejected

- **WHEN** an authenticated `POST /images` request is made with an empty `title` or `mime_type`
- **THEN** the response is `400 Bad Request`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `POST /images` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: POST /images/:id/complete — Notify Upload Complete

The system SHALL expose a `POST /images/:id/complete` endpoint on the protected route group. When called, it SHALL trigger asynchronous thumbnail generation and respond immediately.

Response: `204 No Content`

- The image MUST be owned by the authenticated user
- Returns `404 Not Found` if the image does not exist or belongs to another user
- The endpoint does NOT wait for thumbnail generation to finish

#### Scenario: Upload completion triggers async thumbnail generation

- **WHEN** an authenticated `POST /images/:id/complete` request is made for an existing image
- **THEN** the response is `204 No Content`
- **AND** a goroutine is started to fetch the original, generate a thumbnail, store it, and update `thumbnail_path`

#### Scenario: Image not found or not owned by user

- **WHEN** a `POST /images/:id/complete` request is made for a non-existent or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `POST /images/:id/complete` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: GET /images — Gallery List

The system SHALL expose a `GET /images` endpoint on the protected route group that returns all non-deleted images owned by the authenticated user.

Query parameters:
- `folder_id` (optional UUID) — filter images belonging to a specific folder; omit to return all images regardless of folder

Response body (200):
```json
[{
  "id": "uuid",
  "title": "string",
  "mime_type": "string",
  "source_url": "string|null",
  "folder_id": "uuid|null",
  "thumbnail_url": "string|null",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}]
```

- `thumbnail_url` is a CDN URL (`{R2_PUBLIC_URL}/{thumbnail_path}`) if `thumbnail_path` is set, otherwise null
- Presigned URLs are NOT generated for gallery list items
- Returns an empty array if no images match

#### Scenario: Authenticated user lists all their images

- **WHEN** an authenticated `GET /images` request is made without `folder_id`
- **THEN** the response is `200 OK`
- **AND** the body contains all non-deleted images owned by the user

#### Scenario: Authenticated user filters images by folder

- **WHEN** an authenticated `GET /images?folder_id={uuid}` request is made
- **THEN** the response is `200 OK`
- **AND** the body contains only non-deleted images in that folder owned by the user

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /images` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: GET /images/:id — Image Detail

The system SHALL expose a `GET /images/:id` endpoint on the protected route group that returns full image metadata and a fresh presigned GET URL for the full-resolution image.

Response body (200):
```json
{
  "id": "uuid",
  "title": "string",
  "mime_type": "string",
  "source_url": "string|null",
  "folder_id": "uuid|null",
  "thumbnail_url": "string|null",
  "image_url": "string (presigned GET URL, valid 24 hours)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

- `image_url` is generated fresh on every request via `StorageService.GeneratePresignedGetURL` with a 24-hour TTL and is never stored in the database
- The image MUST be owned by the authenticated user
- Returns `404 Not Found` if the image does not exist, is soft-deleted, or belongs to another user

#### Scenario: Authenticated user retrieves image detail

- **WHEN** an authenticated `GET /images/:id` request is made for an owned non-deleted image
- **THEN** the response is `200 OK`
- **AND** the body includes full metadata and a non-empty `image_url`

#### Scenario: Image not found, soft-deleted, or unowned

- **WHEN** an authenticated `GET /images/:id` request is made for a non-existent, soft-deleted, or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /images/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: DELETE /images/:id — Soft Delete

The system SHALL expose a `DELETE /images/:id` endpoint on the protected route group that soft-deletes an image by setting `deleted_at`.

Response: `204 No Content`

- The image MUST be owned by the authenticated user
- Returns `404 Not Found` if the image does not exist or belongs to another user
- The image is NOT removed from R2; only the `deleted_at` timestamp is set

#### Scenario: Authenticated user soft-deletes an image

- **WHEN** an authenticated `DELETE /images/:id` request is made for an owned non-deleted image
- **THEN** the response is `204 No Content`
- **AND** the image has `deleted_at` set in the database
- **AND** the image no longer appears in `GET /images` results

#### Scenario: Image not found or not owned by user

- **WHEN** an authenticated `DELETE /images/:id` request is made for a non-existent or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `DELETE /images/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: GET /images/trash — Trash List

The system SHALL expose a `GET /images/trash` endpoint on the protected route group that returns all soft-deleted images owned by the authenticated user.

Response body (200): same shape as gallery list, but includes only images where `deleted_at IS NOT NULL`.

- Returns an empty array if the user has no soft-deleted images

#### Scenario: Authenticated user lists trashed images

- **WHEN** an authenticated `GET /images/trash` request is made
- **THEN** the response is `200 OK`
- **AND** the body contains only soft-deleted images owned by the user

#### Scenario: Trashed images do not appear in the regular gallery

- **WHEN** an image has been soft-deleted
- **THEN** it does NOT appear in `GET /images` results
- **AND** it DOES appear in `GET /images/trash` results

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /images/trash` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: POST /images/:id/restore — Restore from Trash

The system SHALL expose a `POST /images/:id/restore` endpoint on the protected route group that restores a soft-deleted image by clearing `deleted_at`.

Response body (200): the restored image in the same shape as the gallery list item.

- The image MUST be soft-deleted and owned by the authenticated user
- Returns `404 Not Found` if the image does not exist, is not soft-deleted, or belongs to another user

#### Scenario: Authenticated user restores a trashed image

- **WHEN** an authenticated `POST /images/:id/restore` request is made for a soft-deleted owned image
- **THEN** the response is `200 OK`
- **AND** `deleted_at` is cleared in the database
- **AND** the image appears again in `GET /images` results

#### Scenario: Image not found, not deleted, or not owned

- **WHEN** a `POST /images/:id/restore` request is made for an image that is not soft-deleted or does not belong to the user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `POST /images/:id/restore` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx, userID string, folderID *uuid.UUID) ([]*domain.Image, error)` — returns non-deleted images; `folderID` nil means no filter
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx, userID string) ([]*domain.Image, error)`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

---

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors

---

### Requirement: Image Routes Wiring

The system SHALL register image routes on the protected Echo group in `main.go`.

Routes:
- `POST /images`
- `POST /images/:id/complete`
- `GET /images`
- `GET /images/:id`
- `DELETE /images/:id`
- `GET /images/trash`
- `POST /images/:id/restore`

#### Scenario: Image routes are registered under auth middleware

- **WHEN** the server starts
- **THEN** all `/images` routes require a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`

---

### Requirement: Image Usecase Unit Tests

The system SHALL have unit tests for `imageUsecase` covering each method with mocked `ImageRepository` and `StorageService`. Each method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: Usecase unit tests cover the happy path and failure path

- **WHEN** each usecase method is tested with a valid mock setup
- **THEN** both the success and at least one error case are asserted

---

### Requirement: Image Handler Unit Tests

The system SHALL have unit tests for `ImageHandler` covering each handler method with a mocked `ImageUsecase`. Each handler method SHALL have at minimum one success scenario and one failure scenario.

#### Scenario: Handler unit tests cover HTTP status codes and response shape

- **WHEN** each handler method is tested with a mock usecase
- **THEN** both the success status code and at least one error status code are asserted

---

### Requirement: Image Repository Integration Tests

The system SHALL have integration tests for `imageRepository` using Testcontainers. Each repository method SHALL be tested against a real PostgreSQL database. Unit tests SHALL NOT be written for the SQL repository.

#### Scenario: Repository integration tests exercise each method against a real database

- **WHEN** the integration test suite runs with a live PostgreSQL container
- **THEN** each `ImageRepository` method is exercised with at least one success scenario and one failure scenario
