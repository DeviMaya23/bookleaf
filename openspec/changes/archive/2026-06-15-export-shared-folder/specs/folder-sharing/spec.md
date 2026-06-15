## ADDED Requirements

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

## MODIFIED Requirements

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
