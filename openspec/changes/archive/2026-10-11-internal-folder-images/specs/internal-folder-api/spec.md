# Spec Delta

## ADDED Requirements

### Requirement: ListFolderImages Usecase

The system SHALL expose an `InternalFolderImagesUsecase` interface on `InternalHandler`:

```go
type InternalFolderImagesUsecase interface {
    ListFolderImages(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]FolderImageThumbnail, error)
}

type FolderImageThumbnail struct {
    ImageID      uuid.UUID
    ThumbnailURL string
}
```

The implementation SHALL verify folder ownership by calling `FolderRepository.GetByID(ctx, folderID, userID)`. If the folder does not exist or is not owned by `userID`, it SHALL return `gorm.ErrRecordNotFound`. It SHALL then call `ImageRepository.ListByFolder` and generate a presigned GET URL for each image whose `ThumbnailPath` is non-nil, using `presignedGetTTL`. Images with a nil `ThumbnailPath` SHALL be excluded from the result.

#### Scenario: Returns thumbnails for owned folder

- **WHEN** `ListFolderImages` is called with a valid userID and a folderID owned by that user
- **THEN** it returns one `FolderImageThumbnail` per image that has a non-nil `ThumbnailPath`, each with a presigned thumbnail URL

#### Scenario: Excludes images without thumbnail path

- **WHEN** a folder contains images where some have a nil `ThumbnailPath`
- **THEN** those images are absent from the returned slice

#### Scenario: Empty folder returns empty slice

- **WHEN** `ListFolderImages` is called for a folder that contains no images
- **THEN** it returns an empty slice and no error

#### Scenario: Unowned folder returns not-found

- **WHEN** `ListFolderImages` is called with a folderID that does not exist or belongs to a different user
- **THEN** it returns `gorm.ErrRecordNotFound`

---

### Requirement: GET /internal/users/:user_id/folders/:folder_id/images

`InternalHandler` SHALL expose a `ListFolderImages(c echo.Context) error` method.

It SHALL read `:user_id` as a Kinde IDP subject, resolve it to an internal UUID via `UserResolver.GetByIDPSubject`, then call `InternalFolderImagesUsecase.ListFolderImages`. On `ErrUserNotFound` it SHALL return `404`. On `gorm.ErrRecordNotFound` from the usecase it SHALL return `404`. On success it SHALL return `200 OK` with body `{ "images": [{ "image_id": "...", "thumbnail_url": "..." }] }`; an empty folder SHALL return `200 OK` with `{ "images": [] }`.

#### Scenario: Returns image thumbnails for an owned folder

- **WHEN** `GET /internal/users/:user_id/folders/:folder_id/images` is called with a valid secret, a known user, and a folder owned by that user
- **THEN** the response is `200 OK` with `{ "images": [...] }` containing one entry per image with a thumbnail

#### Scenario: Empty folder returns 200 with empty array

- **WHEN** the folder exists and is owned by the user but contains no images (or no images have thumbnails)
- **THEN** the response is `200 OK` with `{ "images": [] }`

#### Scenario: Folder not found or not owned returns 404

- **WHEN** the folderID does not exist or belongs to a different user
- **THEN** the response is `404 Not Found`

#### Scenario: Unknown user returns 404

- **WHEN** the `:user_id` IDP subject has no matching user row
- **THEN** the response is `404 Not Found`

#### Scenario: Invalid folder UUID returns 400

- **WHEN** `:folder_id` is not a valid UUID
- **THEN** the response is `400 Bad Request`

---

### Requirement: Internal Route — ListFolderImages

`main.go` SHALL register the route in the existing `/internal` Echo group:

```
GET /internal/users/:user_id/folders/:folder_id/images → internalHandler.ListFolderImages
```

The route SHALL be protected by the existing `NewInternalSecretMiddleware` and SHALL NOT require Kinde auth.

#### Scenario: Route is reachable with correct secret

- **WHEN** `GET /internal/users/:user_id/folders/:folder_id/images` is called with the correct `X-Bookleaf-Internal-Secret`
- **THEN** the request reaches the handler

#### Scenario: Route is blocked without secret

- **WHEN** `GET /internal/users/:user_id/folders/:folder_id/images` is called without the `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`

---

### Requirement: InternalHandler Unit Tests — ListFolderImages

The system SHALL have unit tests for `InternalHandler.ListFolderImages` covering:

- `200` with populated thumbnail list
- `200` with empty list
- `404` when the usecase returns `gorm.ErrRecordNotFound`
- `404` when `UserResolver` returns `ErrUserNotFound`
- `400` on invalid folder UUID

#### Scenario: Unit test asserts thumbnail list shape

- **WHEN** the spy `ListFolderImages` returns two `FolderImageThumbnail` values
- **THEN** the test asserts the response body is `200` with `{ "images": [...] }` containing both entries with `image_id` and `thumbnail_url`

#### Scenario: Unit test asserts 404 on not-found

- **WHEN** the spy `ListFolderImages` returns `gorm.ErrRecordNotFound`
- **THEN** the test asserts the response status is `404`

---

### Requirement: Bruno File — ListFolderImages

A Bruno request file SHALL be created at `bruno/internal/list-folder-images.bru` with the `GET /internal/users/:user_id/folders/:folder_id/images` request and the `X-Bookleaf-Internal-Secret` header.

#### Scenario: Bruno file exists for the new endpoint

- **WHEN** a developer opens the Bruno collection
- **THEN** they can find and run a request for list folder images
