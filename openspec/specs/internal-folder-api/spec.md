## Purpose

Expose a set of internal (service-to-service) HTTP endpoints for querying folder sharing state by user or folder ID, protected by a shared secret rather than Kinde auth.

## Requirements

### Requirement: Internal API Secret Config

The system SHALL require an `INTERNAL_API_SECRET` env var. `config.Config` SHALL expose it as `InternalAPISecret string`. If the value is empty, `config.Load` SHALL return an error.

#### Scenario: Missing env var fails startup

- **WHEN** the server starts without `INTERNAL_API_SECRET` set
- **THEN** config loading returns a non-nil error containing `"INTERNAL_API_SECRET"`

---

### Requirement: InternalSecretMiddleware

The `handler/middleware` package SHALL expose:

```go
func NewInternalSecretMiddleware(secret string) echo.MiddlewareFunc
```

The middleware SHALL read the `X-Bookleaf-Internal-Secret` request header. If the header is absent or does not equal `secret`, it SHALL return `401 Unauthorized` and not call the next handler. If it matches, it SHALL call the next handler unchanged.

#### Scenario: Missing secret header returns 401

- **WHEN** a request to an internal route is made without the `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`
- **AND** the downstream handler is not called

#### Scenario: Wrong secret value returns 401

- **WHEN** a request to an internal route is made with an incorrect `X-Bookleaf-Internal-Secret` value
- **THEN** the response is `401 Unauthorized`

#### Scenario: Correct secret passes through

- **WHEN** a request to an internal route is made with the correct `X-Bookleaf-Internal-Secret` value
- **THEN** the downstream handler is called and its response is returned unchanged

---

### Requirement: InternalShareUsecase — GetPublicFoldersByUser

`shareUsecase` SHALL implement:

```go
type FolderShareSummary struct {
    FolderID   uuid.UUID
    Token      string
    FolderName string
}

GetPublicFoldersByUser(ctx context.Context, userID string) ([]FolderShareSummary, error)
```

It SHALL call `FolderShareRepository.ListByUserID(ctx, userID)` and map each result to a `FolderShareSummary`, including `FolderName` from the item. If no rows exist, it SHALL return an empty slice and no error.

#### Scenario: Returns all public folders for a user

- **WHEN** `GetPublicFoldersByUser` is called for a user who has two shared folders
- **THEN** it returns a slice of two `FolderShareSummary` values with the correct `FolderID`, `Token`, and `FolderName` for each

#### Scenario: Returns empty slice when user has no public folders

- **WHEN** `GetPublicFoldersByUser` is called for a user with no `folder_shares` rows
- **THEN** it returns an empty slice and no error

---

### Requirement: InternalShareUsecase — GetSharedFolderByFolderID

`shareUsecase` SHALL implement:

```go
GetSharedFolderByFolderID(ctx context.Context, folderID uuid.UUID) (*SharedFolder, error)
```

Behavior mirrors `GetSharedFolder` (token-based) exactly, except the entry point is folder ID:
1. Call `FolderShareRepository.GetByFolderIDWithFolder(ctx, folderID)`. Return `gorm.ErrRecordNotFound` if the folder has no share.
2. Call `ShareImageRepository.ListByFolder(ctx, share.Folder.UserID, share.FolderID, nil, nil)`.
3. For each image generate presigned `FullResURL`, `DownloadURL`, and `ThumbnailURL` using the existing `presignedGetTTL`, identical to `GetSharedFolder`.
4. Return `SharedFolder{Name: share.Folder.Name, Notes: share.Folder.Description, Images: ...}`.

#### Scenario: Returns folder contents for a shared folder

- **WHEN** `GetSharedFolderByFolderID` is called with a folder ID that has a share record
- **THEN** it returns a `SharedFolder` with the correct name, notes, and images with presigned URLs

#### Scenario: Unshared folder returns not-found

- **WHEN** `GetSharedFolderByFolderID` is called with a folder ID that has no `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

---

### Requirement: InternalShareUsecase — CheckFolderPublicStatus

`shareUsecase` SHALL implement:

```go
CheckFolderPublicStatus(ctx context.Context, folderID uuid.UUID) (token string, err error)
```

It SHALL call `FolderShareRepository.GetByFolderID(ctx, folderID)`. If found, return the share's token. If `gorm.ErrRecordNotFound`, return `"", gorm.ErrRecordNotFound`.

#### Scenario: Public folder returns token

- **WHEN** `CheckFolderPublicStatus` is called for a folder that has a `folder_shares` row
- **THEN** it returns the share's token and no error

#### Scenario: Private folder returns not-found

- **WHEN** `CheckFolderPublicStatus` is called for a folder that has no `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

---

### Requirement: InternalHandler

The `handler` package SHALL define `InternalHandler` in `handler/internal.go`:

```go
type InternalShareUsecase interface {
    GetPublicFoldersByUser(ctx context.Context, userID string) ([]usecase.FolderShareSummary, error)
    GetSharedFolderByFolderID(ctx context.Context, folderID uuid.UUID) (*usecase.SharedFolder, error)
    CheckFolderPublicStatus(ctx context.Context, folderID uuid.UUID) (string, error)
}

type InternalHandler struct { ... }

func NewInternalHandler(shareUsecase InternalShareUsecase, tel *observability.Telemetry) *InternalHandler
```

Methods:
- `ListPublicFolders(c echo.Context) error` — reads `:user_id` path param (string, no UUID parsing), calls `GetPublicFoldersByUser`, returns `200` with `{"folder_list": [{"folder_id": "...", "token": "...", "folder_name": "..."}, ...]}`. Empty list returns `200` with `{"folder_list": []}`.
- `GetFolderContents(c echo.Context) error` — parses `:folder_id` as UUID (`400` on failure), calls `GetSharedFolderByFolderID`, maps to the same response shape as `GET /share/:token` (`404` on `gorm.ErrRecordNotFound`).
- `CheckFolderStatus(c echo.Context) error` — parses `:folder_id` as UUID (`400` on failure), calls `CheckFolderPublicStatus`, returns `200` with `{"token": "..."}` or `404` on not-found.

#### Scenario: ListPublicFolders returns folder list

- **WHEN** `GET /internal/users/:user_id/public-folders` is called with a valid secret and a user who has public folders
- **THEN** the response is `200 OK` with `{"folder_list": [{"folder_id": "...", "token": "...", "folder_name": "..."}, ...]}`

#### Scenario: ListPublicFolders returns empty list

- **WHEN** `GET /internal/users/:user_id/public-folders` is called for a user with no public folders
- **THEN** the response is `200 OK` with `{"folder_list": []}`

#### Scenario: GetFolderContents returns shared folder shape

- **WHEN** `GET /internal/folders/:folder_id/contents` is called for a shared folder
- **THEN** the response is `200 OK` with the same JSON shape as `GET /share/:token`

#### Scenario: GetFolderContents for unshared folder returns 404

- **WHEN** `GET /internal/folders/:folder_id/contents` is called for a folder with no share record
- **THEN** the response is `404 Not Found`

#### Scenario: GetFolderContents invalid UUID returns 400

- **WHEN** `GET /internal/folders/:folder_id/contents` is called with a non-UUID `:folder_id`
- **THEN** the response is `400 Bad Request`

#### Scenario: CheckFolderStatus returns token for public folder

- **WHEN** `GET /internal/folders/:folder_id/status` is called for a public folder
- **THEN** the response is `200 OK` with `{"token": "..."}`

#### Scenario: CheckFolderStatus returns 404 for private or nonexistent folder

- **WHEN** `GET /internal/folders/:folder_id/status` is called for a folder with no share record
- **THEN** the response is `404 Not Found`

#### Scenario: CheckFolderStatus invalid UUID returns 400

- **WHEN** `GET /internal/folders/:folder_id/status` is called with a non-UUID `:folder_id`
- **THEN** the response is `400 Bad Request`

---

### Requirement: Internal Route Registration

`main.go` SHALL register an `/internal` Echo group with `NewInternalSecretMiddleware(cfg.InternalAPISecret)` applied, containing:

- `GET /internal/users/:user_id/public-folders` → `internalHandler.ListPublicFolders`
- `GET /internal/folders/:folder_id/contents` → `internalHandler.GetFolderContents`
- `GET /internal/folders/:folder_id/status` → `internalHandler.CheckFolderStatus`

The group SHALL NOT have the `AuthMiddleware` or `MaintenanceMiddleware` applied.

#### Scenario: Internal routes do not require Kinde auth

- **WHEN** any internal route is called with the correct `X-Bookleaf-Internal-Secret` and no `Authorization` header
- **THEN** the request is processed and a non-401 response is returned

#### Scenario: Internal routes are blocked without the secret

- **WHEN** any internal route is called without the `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`

---

### Requirement: InternalHandler Unit Tests

The system SHALL have unit tests for `InternalHandler` covering all three handler methods, using a value-return spy `InternalShareUsecase`.

Tests SHALL cover:
- `ListPublicFolders`: `200` with populated list, `200` with empty list
- `GetFolderContents`: `200` with correct response shape, `404` on not-found, `400` on invalid UUID
- `CheckFolderStatus`: `200` with token, `404` on not-found, `400` on invalid UUID

#### Scenario: ListPublicFolders unit test asserts response shape

- **WHEN** the spy returns two `FolderShareSummary` values (each with `FolderID`, `Token`, and `FolderName`)
- **THEN** the test asserts the response body contains `{"folder_list": [...]}` with both entries, each including `folder_name`

#### Scenario: GetFolderContents unit test asserts not-found mapping

- **WHEN** the spy `GetSharedFolderByFolderID` returns `gorm.ErrRecordNotFound`
- **THEN** the test asserts the response status is `404`

#### Scenario: CheckFolderStatus unit test asserts token in response

- **WHEN** the spy `CheckFolderPublicStatus` returns a token string
- **THEN** the test asserts the response body contains `{"token": "<that token>"}`

---

### Requirement: InternalSecretMiddleware Unit Tests

The system SHALL have unit tests for `NewInternalSecretMiddleware` covering:
- Absent header → `401`
- Wrong value → `401`
- Correct value → next handler called, response passed through

#### Scenario: Middleware unit test asserts 401 on missing header

- **WHEN** the middleware is applied and a request is made without `X-Bookleaf-Internal-Secret`
- **THEN** the test asserts the response status is `401` and the downstream handler is not called

#### Scenario: Middleware unit test asserts passthrough on correct secret

- **WHEN** the middleware is applied and a request is made with the correct header value
- **THEN** the test asserts the downstream handler is called exactly once

---

### Requirement: Bruno Collection for Internal Endpoints

A Bruno collection file SHALL be created at `bruno/internal/` with requests for all three internal endpoints, including the `X-Bookleaf-Internal-Secret` header.

#### Scenario: Bruno files exist for all three internal endpoints

- **WHEN** a developer opens the Bruno collection
- **THEN** they can find and run requests for list public folders, get folder contents, and check folder status
