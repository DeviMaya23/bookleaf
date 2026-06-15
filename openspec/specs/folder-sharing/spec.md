## Purpose
Allow folder owners to generate a public, unauthenticated link that exposes a folder's name, notes, and images (a read-only gallery view) to anyone with the link, and to revoke that link at any time.

## Requirements

### Requirement: FolderShare Domain Type and Migration

The system SHALL define a `domain.FolderShare` type and a `folder_shares` table via migration `000015_create_folder_shares`:

```go
type FolderShare struct {
    ID        uuid.UUID
    FolderID  uuid.UUID
    Token     string
    CreatedAt time.Time
    Folder    Folder
}
```

```sql
CREATE TABLE folder_shares (
    id         UUID PRIMARY KEY,
    folder_id  UUID NOT NULL UNIQUE REFERENCES folders(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

`folder_id` SHALL be unique, enforcing at most one share per folder. Deleting a folder SHALL cascade-delete its `folder_shares` row.

#### Scenario: Deleting a shared folder removes its share row

- **WHEN** a folder that has an associated `folder_shares` row is deleted
- **THEN** the corresponding `folder_shares` row is also removed

#### Scenario: A folder cannot have two share rows

- **WHEN** an attempt is made to insert a second `folder_shares` row for the same `folder_id`
- **THEN** the database rejects the insert due to the unique constraint on `folder_id`

---

### Requirement: FolderShareRepository Interface and Implementation

The `usecase` package SHALL define a `FolderShareRepository` interface:

```go
type FolderShareRepository interface {
    Create(ctx context.Context, folderID uuid.UUID, token string) (*domain.FolderShare, error)
    GetByFolderID(ctx context.Context, folderID uuid.UUID) (*domain.FolderShare, error)
    GetByToken(ctx context.Context, token string) (*domain.FolderShare, error)
    DeleteByFolderID(ctx context.Context, folderID uuid.UUID) error
}
```

`GetByFolderID` and `GetByToken` SHALL return `gorm.ErrRecordNotFound` when no row matches. `GetByToken` SHALL preload the associated `Folder`. `DeleteByFolderID` SHALL not error when no row exists for the given folder.

The system SHALL implement this interface in `repository/folder_share_repository.go` using GORM.

#### Scenario: GetByFolderID returns not-found for an unshared folder

- **WHEN** `GetByFolderID` is called with a folder ID that has no `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: GetByToken preloads the folder

- **WHEN** `GetByToken` is called with a valid token
- **THEN** the returned `FolderShare.Folder` is populated with the associated folder's `Name`, `Description`, and `UserID`

#### Scenario: GetByToken returns not-found for an unknown token

- **WHEN** `GetByToken` is called with a token that does not exist
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: DeleteByFolderID is idempotent

- **WHEN** `DeleteByFolderID` is called for a folder with no existing `folder_shares` row
- **THEN** it returns no error

---

### Requirement: Share-Scoped Narrow Repository Interfaces

The `usecase` package SHALL define two narrow interfaces consumed only by `ShareUsecase`, distinct from the existing `FolderRepository` and `FolderImageRepository` interfaces:

```go
type ShareFolderRepository interface {
    GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error)
}

type ShareImageRepository interface {
    ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
}
```

These SHALL be satisfied implicitly by the existing `folderRepository` and `imageRepository` implementations without modification.

#### Scenario: Existing repositories satisfy the new interfaces

- **WHEN** the Go package is compiled
- **THEN** `folderRepository` implements `usecase.ShareFolderRepository`
- **AND** `imageRepository` implements `usecase.ShareImageRepository`
- **WITHOUT** any changes to `repository/folder_repository.go` or `repository/image_repository.go`

---

### Requirement: ShareUsecase — CreateShare

`ShareUsecase` SHALL define:

```go
CreateShare(ctx context.Context, folderID uuid.UUID, userID string) (token string, created bool, err error)
```

Behavior:
1. Verify the folder exists and is owned by `userID` via `ShareFolderRepository.GetByID`. Return `gorm.ErrRecordNotFound` if not.
2. Call `FolderShareRepository.GetByFolderID(folderID)`. If a row exists, return its `token` with `created = false`.
3. Otherwise generate a token (16 random bytes via `crypto/rand`, base64 `RawURLEncoding`) and call `FolderShareRepository.Create(folderID, token)`. Return the new token with `created = true`.
4. If `Create` fails due to a unique-constraint violation on `folder_id` (a concurrent request created the row first), fall back to `FolderShareRepository.GetByFolderID(folderID)` and return that token with `created = false`.

#### Scenario: First share creates a new token

- **WHEN** `CreateShare` is called for a folder owned by the user with no existing share
- **THEN** it returns a newly generated token and `created = true`

#### Scenario: Repeat share returns the existing token

- **WHEN** `CreateShare` is called for a folder that already has a `folder_shares` row
- **THEN** it returns the existing token unchanged and `created = false`
- **AND** `FolderShareRepository.Create` is not called

#### Scenario: Concurrent create falls back to the winning row

- **WHEN** `FolderShareRepository.Create` returns a unique-constraint violation error
- **THEN** `CreateShare` calls `GetByFolderID` and returns that token with `created = false`
- **AND** does not propagate the constraint-violation error

#### Scenario: Folder not owned by user

- **WHEN** `CreateShare` is called with a folder ID that does not exist or belongs to another user
- **THEN** it returns `gorm.ErrRecordNotFound`
- **AND** no `folder_shares` row is created

---

### Requirement: ShareUsecase — GetShare

`ShareUsecase` SHALL define:

```go
GetShare(ctx context.Context, folderID uuid.UUID, userID string) (token string, err error)
```

Behavior:
1. Verify the folder exists and is owned by `userID` via `ShareFolderRepository.GetByID`. Return `gorm.ErrRecordNotFound` if not.
2. Call `FolderShareRepository.GetByFolderID(folderID)` and return its token, or `gorm.ErrRecordNotFound` if no share exists.

#### Scenario: Returns token for a shared folder

- **WHEN** `GetShare` is called for a folder owned by the user that has a `folder_shares` row
- **THEN** it returns that row's token

#### Scenario: Returns not-found for an unshared folder

- **WHEN** `GetShare` is called for a folder owned by the user with no `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: Folder not owned by user

- **WHEN** `GetShare` is called with a folder ID that does not exist or belongs to another user
- **THEN** it returns `gorm.ErrRecordNotFound`

---

### Requirement: ShareUsecase — DeleteShare

`ShareUsecase` SHALL define:

```go
DeleteShare(ctx context.Context, folderID uuid.UUID, userID string) error
```

Behavior:
1. Verify the folder exists and is owned by `userID` via `ShareFolderRepository.GetByID`. Return `gorm.ErrRecordNotFound` if not.
2. Call `FolderShareRepository.DeleteByFolderID(folderID)`. This SHALL succeed (no error) whether or not a share row existed.

#### Scenario: Revokes an existing share

- **WHEN** `DeleteShare` is called for a folder owned by the user that has a `folder_shares` row
- **THEN** the row is deleted and no error is returned

#### Scenario: Deleting a non-existent share is a no-op

- **WHEN** `DeleteShare` is called for a folder owned by the user with no `folder_shares` row
- **THEN** no error is returned

#### Scenario: Folder not owned by user

- **WHEN** `DeleteShare` is called with a folder ID that does not exist or belongs to another user
- **THEN** it returns `gorm.ErrRecordNotFound`
- **AND** no `folder_shares` row is deleted

---

### Requirement: ShareUsecase — GetSharedFolder (public read)

`ShareUsecase` SHALL define:

```go
type SharedImage struct {
    Title        string
    ThumbnailURL *string
    FullResURL   string
    DownloadURL  string
    Width        *int
    Height       *int
}

type SharedFolder struct {
    Name   string
    Notes  *string
    Images []SharedImage
}

GetSharedFolder(ctx context.Context, token string) (*SharedFolder, error)
```

Behavior:
1. Call `FolderShareRepository.GetByToken(token)`. Return `gorm.ErrRecordNotFound` if the token is unknown.
2. Call `ShareImageRepository.ListByFolder(ctx, share.Folder.UserID, share.FolderID, nil, nil)` to get the folder's direct images, ordered by `image_folders.position` (same default ordering as the owner's folder view).
3. For each image, generate `ThumbnailURL` (nil if `ThumbnailPath` is nil) and `FullResURL` via `StorageService.GeneratePresignedGetURL` with the existing `presignedGetTTL`. Generate `DownloadURL` via `StorageService.GeneratePresignedDownloadURL(ctx, img.R2Path, filename, presignedGetTTL)`, where `filename` is `img.Title` plus the extension from `downloadFileExtension(img.MIMEType)`. Set `Width` and `Height` directly from `domain.Image.Width` and `domain.Image.Height`.
4. Return `SharedFolder{Name: share.Folder.Name, Notes: share.Folder.Description, Images: ...}`.

#### Scenario: Returns folder name, notes, and ordered images

- **WHEN** `GetSharedFolder` is called with a valid token for a folder with a description and multiple images
- **THEN** it returns the folder's `Name` and `Notes` (from `Description`)
- **AND** `Images` are ordered the same as `image_folders.position`
- **AND** each image has a non-empty `FullResURL` and a non-empty `DownloadURL`

#### Scenario: Image without a thumbnail has a nil ThumbnailURL

- **WHEN** `GetSharedFolder` is called for a folder containing an image with `ThumbnailPath == nil`
- **THEN** that image's `SharedImage.ThumbnailURL` is `nil`

#### Scenario: Image dimensions are passed through

- **WHEN** `GetSharedFolder` is called for a folder containing an image with non-nil `domain.Image.Width` and `Height`
- **THEN** that image's `SharedImage.Width` and `SharedImage.Height` equal the source image's `Width` and `Height`

#### Scenario: Image with no recorded dimensions has nil Width and Height

- **WHEN** `GetSharedFolder` is called for a folder containing an image with `Width == nil` and `Height == nil`
- **THEN** that image's `SharedImage.Width` and `SharedImage.Height` are `nil`

#### Scenario: Unknown token returns not-found

- **WHEN** `GetSharedFolder` is called with a token that does not match any `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: Folder with no images returns an empty list

- **WHEN** `GetSharedFolder` is called for a shared folder with zero direct images
- **THEN** it returns `SharedFolder.Images` as an empty slice and no error

---

### Requirement: Owner-Facing Share Endpoints

The system SHALL expose the following authenticated routes on the `protected` Echo group, handled by `ShareHandler`:

- `POST /folders/:id/share` — calls `ShareUsecase.CreateShare`. Returns `201 Created` with `{"token": "..."}` if a new share was created, or `200 OK` with `{"token": "..."}` if a share already existed.
- `GET /folders/:id/share` — calls `ShareUsecase.GetShare`. Returns `200 OK` with `{"token": "..."}`, or `404 Not Found` if the folder is not shared.
- `DELETE /folders/:id/share` — calls `ShareUsecase.DeleteShare`. Returns `204 No Content`.

All three SHALL: extract `userID` from the Kinde JWT context, parse `:id` as a UUID (`400 Bad Request` on failure), and return `404 Not Found` when `gorm.ErrRecordNotFound` is returned by the usecase (folder not found or not owned).

#### Scenario: Create share for an unshared folder

- **WHEN** an authenticated owner calls `POST /folders/:id/share` for a folder with no existing share
- **THEN** the response is `201 Created` with a JSON body containing a `token` field

#### Scenario: Create share is idempotent

- **WHEN** an authenticated owner calls `POST /folders/:id/share` for a folder that is already shared
- **THEN** the response is `200 OK` with the same `token` as before

#### Scenario: Get share for a shared folder

- **WHEN** an authenticated owner calls `GET /folders/:id/share` for a shared folder
- **THEN** the response is `200 OK` with the folder's `token`

#### Scenario: Get share for an unshared folder returns 404

- **WHEN** an authenticated owner calls `GET /folders/:id/share` for a folder that has never been shared
- **THEN** the response is `404 Not Found`

#### Scenario: Delete share revokes access

- **WHEN** an authenticated owner calls `DELETE /folders/:id/share` for a shared folder
- **THEN** the response is `204 No Content`
- **AND** a subsequent `GET /folders/:id/share` returns `404 Not Found`

#### Scenario: Invalid UUID returns 400

- **WHEN** any of the three endpoints is called with a non-UUID `:id`
- **THEN** the response is `400 Bad Request`

#### Scenario: Folder not owned by user returns 404

- **WHEN** any of the three endpoints is called for a folder ID that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request returns 401

- **WHEN** any of the three endpoints is called without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: Public Shared Folder Endpoint

The system SHALL expose `GET /share/:token` as a public (unauthenticated) route, registered outside the `protected` group but still behind the global CORS and recovery middleware. The handler SHALL call `ShareUsecase.GetSharedFolder(ctx, token)` and respond:

- `200 OK` with body:
  ```json
  {
    "folder": { "name": "...", "notes": "...|null" },
    "images": [
      { "title": "...", "thumbnail_url": "...|null", "full_res_url": "...", "download_url": "...", "width": 0, "height": 0 }
    ]
  }
  ```
  `width` and `height` SHALL be `null` when the corresponding `SharedImage.Width`/`Height` is `nil`. `full_res_url` SHALL be a presigned URL suitable for inline display (e.g. `<img src>`), while `download_url` SHALL be a presigned URL with `Content-Disposition: attachment; filename="<image title and extension>"` set, suitable for triggering a file download.
- `404 Not Found` if `GetSharedFolder` returns `gorm.ErrRecordNotFound`

#### Scenario: Valid token returns folder and images

- **WHEN** `GET /share/:token` is called with a token for a shared folder
- **THEN** the response is `200 OK`
- **AND** the body contains `folder.name`, `folder.notes`, and an `images` array with `title`, `thumbnail_url`, `full_res_url`, `download_url`, `width`, and `height` per image

#### Scenario: Unknown or revoked token returns 404

- **WHEN** `GET /share/:token` is called with a token that does not exist (including a previously-valid token that has since been revoked via `DELETE /folders/:id/share`)
- **THEN** the response is `404 Not Found`

#### Scenario: No authentication required

- **WHEN** `GET /share/:token` is called without any Authorization header
- **THEN** the request is not rejected by auth middleware

---

### Requirement: ShareUsecase — GetSharedFolderInfo

`ShareUsecase` SHALL define:

```go
GetSharedFolderInfo(ctx context.Context, token string) (*domain.Folder, error)
```

Behavior:
1. Call `FolderShareRepository.GetByToken(token)`. Return `gorm.ErrRecordNotFound` if the token is unknown.
2. Return `share.Folder` (populated with at least `ID`, `Name`, and `UserID` via `GetByToken`'s preload).

#### Scenario: Returns folder info for a valid token

- **WHEN** `GetSharedFolderInfo` is called with a token for a shared folder
- **THEN** it returns a `*domain.Folder` with that folder's `ID`, `Name`, and `UserID`

#### Scenario: Unknown token returns not-found

- **WHEN** `GetSharedFolderInfo` is called with a token that does not match any `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

---

### Requirement: ShareHandler FolderExporter Interface

The `handler` package SHALL define a narrow `FolderExporter` interface consumed only by `ShareHandler`:

```go
type FolderExporter interface {
    ExportFolder(ctx context.Context, folderID uuid.UUID, userID string, w io.Writer) error
}
```

This SHALL be satisfied implicitly by the existing `folderUsecase` (which already implements `ExportFolder` per the `folder-export` capability) without modification. `NewShareHandler` SHALL accept a `FolderExporter` as an additional constructor parameter.

#### Scenario: Existing folder usecase satisfies the new interface

- **WHEN** the Go package is compiled
- **THEN** `folderUsecase` implements `handler.FolderExporter`
- **WITHOUT** any changes to `usecase/folder_usecase.go`

---

### Requirement: Public Shared Folder Export Endpoint

The system SHALL expose `GET /share/:token/export` as a public (unauthenticated) route, registered outside the `protected` group but still behind the global CORS and recovery middleware. `ShareHandler.ExportSharedFolder` SHALL:

1. Extract `:token` from the path.
2. Call `ShareUsecase.GetSharedFolderInfo(ctx, token)`. Return `404 Not Found` if it returns `gorm.ErrRecordNotFound`, before writing any response body.
3. On success, derive the archive filename as `<sanitized folder name>.zip` using the existing `sanitizeFilename` helper (falling back to `"export"` if sanitizing leaves an empty string).
4. Set `Content-Type: application/zip` and `Content-Disposition: attachment; filename="<archive filename>"`, write `200 OK`.
5. Call `FolderExporter.ExportFolder(ctx, folder.ID, folder.UserID, c.Response())` to stream the zip body, where `folder` is the `*domain.Folder` returned by `GetSharedFolderInfo`.
6. If `ExportFolder` returns an error after the response has started, log it server-side; the response is not modified further.

#### Scenario: Valid token streams a zip of the shared folder's images

- **WHEN** `GET /share/:token/export` is called with a token for a shared folder
- **THEN** the response status is `200 OK`
- **AND** the response has `Content-Type: application/zip`
- **AND** the response body is a valid zip archive containing the same images as `GET /share/:token`

#### Scenario: Archive filename derived from folder name

- **WHEN** `GET /share/:token/export` is called for a shared folder named `"Trip / 2024"`
- **THEN** the `Content-Disposition` filename does not contain `/`

#### Scenario: Folder name that sanitizes to empty falls back to a default

- **WHEN** `GET /share/:token/export` is called for a shared folder whose name consists entirely of characters invalid in filenames
- **THEN** the `Content-Disposition` filename is `"export.zip"`

#### Scenario: Unknown or revoked token returns 404

- **WHEN** `GET /share/:token/export` is called with a token that does not exist (including a previously-valid token that has since been revoked via `DELETE /folders/:id/share`)
- **THEN** the response is `404 Not Found`
- **AND** no response body is written

#### Scenario: No authentication required

- **WHEN** `GET /share/:token/export` is called without any Authorization header
- **THEN** the request is not rejected by auth middleware

---

### Requirement: Share Route Registration

The system SHALL register the share routes in `main.go`:
- `protected.POST("/folders/:id/share", ...)`
- `protected.GET("/folders/:id/share", ...)`
- `protected.DELETE("/folders/:id/share", ...)`
- `e.GET("/share/:token", ...)` (outside `protected`)
- `e.GET("/share/:token/export", ...)` (outside `protected`)

#### Scenario: Owner routes require authentication

- **WHEN** the server starts
- **THEN** `POST`, `GET`, and `DELETE /folders/:id/share` each require a valid Kinde Bearer token

#### Scenario: Public routes do not require authentication

- **WHEN** the server starts
- **THEN** `GET /share/:token` and `GET /share/:token/export` do not require a Kinde Bearer token

---

### Requirement: ShareUsecase Unit Tests

The system SHALL have unit tests for `ShareUsecase` using fake `FolderShareRepository`/`ShareFolderRepository`/`ShareImageRepository` (per CONVENTIONS.md fake-vs-spy rules) and a value-return spy `StorageService`, covering:

- `CreateShare`: new share created, idempotent repeat, concurrent-create fallback, not-owned error
- `GetShare`: found, not-shared not-found, not-owned error
- `DeleteShare`: revokes existing, no-op when absent, not-owned error
- `GetSharedFolder`: assembles folder + ordered images with presigned URLs, nil thumbnail handling, unknown token, empty folder
- `GetSharedFolderInfo`: returns folder info for a valid token, unknown token not-found

#### Scenario: Unit test asserts idempotent create does not call Create twice

- **WHEN** the fake `FolderShareRepository` already has a row for the folder
- **THEN** the test asserts `CreateShare` returns that token, `created = false`, and `Create` was not invoked

#### Scenario: Unit test asserts presigned URL assembly

- **WHEN** the fake repositories return a folder and images, and the storage spy returns a fixed presigned URL
- **THEN** the test asserts `GetSharedFolder` returns `SharedFolder.Images` with `FullResURL` set from the spy's return value

#### Scenario: Unit test asserts GetSharedFolderInfo not-found

- **WHEN** the fake `FolderShareRepository.GetByToken` returns `gorm.ErrRecordNotFound`
- **THEN** the test asserts `GetSharedFolderInfo` returns `gorm.ErrRecordNotFound`

---

### Requirement: ShareHandler Unit Tests

The system SHALL have unit tests for `ShareHandler` using value-return spy `ShareUsecase` and `FolderExporter`, covering:

- `POST /folders/:id/share`: `201` on creation, `200` on existing, `400` invalid UUID, `404` not owned, `401` unauthenticated
- `GET /folders/:id/share`: `200` with token, `404` not shared / not owned, `400` invalid UUID, `401` unauthenticated
- `DELETE /folders/:id/share`: `204`, `400` invalid UUID, `404` not owned, `401` unauthenticated
- `GET /share/:token`: `200` with response body shape, `404` for unknown token
- `GET /share/:token/export`: `200` with `Content-Type`/`Content-Disposition` headers and zip body, `404` for unknown token (with `FolderExporter.ExportFolder` not called)

#### Scenario: Handler unit test asserts 201 vs 200 based on `created`

- **WHEN** the spy `ShareUsecase.CreateShare` returns `created = true`
- **THEN** the test asserts the response status is `201`
- **WHEN** the spy returns `created = false`
- **THEN** the test asserts the response status is `200`

#### Scenario: Handler unit test asserts public response body shape

- **WHEN** the spy `ShareUsecase.GetSharedFolder` returns a `SharedFolder` with name, notes, and images
- **THEN** the test asserts the response body matches the documented JSON shape

#### Scenario: Handler unit test asserts export headers and not-found mapping

- **WHEN** the spy `ShareUsecase.GetSharedFolderInfo` returns a folder named `"My Folder"` and the spy `FolderExporter.ExportFolder` succeeds
- **THEN** the test asserts the response status is `200`, `Content-Type` is `application/zip`, and `Content-Disposition` is `attachment; filename="My Folder.zip"`
- **WHEN** the spy `ShareUsecase.GetSharedFolderInfo` returns `gorm.ErrRecordNotFound`
- **THEN** the test asserts the response status is `404` and `FolderExporter.ExportFolder` was not called
