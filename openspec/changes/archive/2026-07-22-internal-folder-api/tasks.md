## 1. Config

- [x] 1.1 Add `InternalAPISecret string` field to `config.Config` in `internal/platform/config/config.go`
- [x] 1.2 Load `INTERNAL_API_SECRET` via `requireEnv` in `loadFromEnv` and set it on the struct

## 2. Repository

- [x] 2.1 Add `GetByFolderIDWithFolder` and `ListByUserID` to the `FolderShareRepository` interface in `internal/usecase/folder_share_repository.go`
- [x] 2.2 Implement `GetByFolderIDWithFolder` in `internal/repository/folder_share_repository.go` (Preload `"Folder"`, query by `folder_id`)
- [x] 2.3 Implement `ListByUserID` in `internal/repository/folder_share_repository.go` (JOIN `folder_shares` with `folders` on `folder_id`, filter `folders.user_id = ?`)

## 3. Usecase

- [x] 3.1 Add `FolderShareSummary` struct to `internal/usecase/share_usecase.go`
- [x] 3.2 Implement `GetPublicFoldersByUser(ctx, userID string) ([]FolderShareSummary, error)` on `shareUsecase`
- [x] 3.3 Implement `GetSharedFolderByFolderID(ctx, folderID uuid.UUID) (*SharedFolder, error)` on `shareUsecase` (mirrors `GetSharedFolder`, uses `GetByFolderIDWithFolder`)
- [x] 3.4 Implement `CheckFolderPublicStatus(ctx, folderID uuid.UUID) (string, error)` on `shareUsecase`

## 4. Middleware

- [x] 4.1 Create `internal/handler/middleware/internal_secret.go` with `NewInternalSecretMiddleware(secret string) echo.MiddlewareFunc`
- [x] 4.2 Write unit tests for `NewInternalSecretMiddleware` in `internal_secret_test.go`: absent header → 401, wrong value → 401, correct value → next called

## 5. Handler

- [x] 5.1 Create `internal/handler/internal.go` with `InternalShareUsecase` interface, `InternalHandler` struct, and `NewInternalHandler` constructor
- [x] 5.2 Implement `ListPublicFolders` handler (reads `:user_id` as string, calls `GetPublicFoldersByUser`, returns `200` with `{"folder_list": [...]}`)
- [x] 5.3 Implement `GetFolderContents` handler (parses `:folder_id` as UUID, calls `GetSharedFolderByFolderID`, maps to shared folder response shape, `404` on not-found)
- [x] 5.4 Implement `CheckFolderStatus` handler (parses `:folder_id` as UUID, calls `CheckFolderPublicStatus`, returns `{"token": "..."}` or `404`)
- [x] 5.5 Write unit tests for `InternalHandler` in `internal_test.go`:
  - `ListPublicFolders`: 200 with populated list, 200 with empty list
  - `GetFolderContents`: 200 with correct shape, 404 on not-found, 400 on invalid UUID
  - `CheckFolderStatus`: 200 with token, 404 on not-found, 400 on invalid UUID

## 6. Usecase Tests

- [x] 6.1 Write unit tests for `GetPublicFoldersByUser`: returns summaries for user with shares, returns empty slice for user with none
- [x] 6.2 Write unit tests for `GetSharedFolderByFolderID`: returns folder contents with presigned URLs for a shared folder, returns `gorm.ErrRecordNotFound` for unshared folder
- [x] 6.3 Write unit tests for `CheckFolderPublicStatus`: returns token for public folder, returns `gorm.ErrRecordNotFound` for private folder

## 7. Route Registration

- [x] 7.1 Wire up `internalHandler := httphandler.NewInternalHandler(shareUsecase, tel)` in `initApp` in `cmd/server/main.go`
- [x] 7.2 Register `/internal` group with `NewInternalSecretMiddleware(cfg.InternalAPISecret)` and the three routes

## 8. Bruno

- [x] 8.1 Create `bruno/internal/` directory with a `list-public-folders.bru` request (`GET /internal/users/:user_id/public-folders`, includes `X-Bookleaf-Internal-Secret` header)
- [x] 8.2 Create `bruno/internal/get-folder-contents.bru` (`GET /internal/folders/:folder_id/contents`, includes `X-Bookleaf-Internal-Secret` header)
- [x] 8.3 Create `bruno/internal/check-folder-status.bru` (`GET /internal/folders/:folder_id/status`, includes `X-Bookleaf-Internal-Secret` header)

## 9. Lint

- [x] 9.1 Run `golangci-lint run ./...` from `backend/` and fix any issues

## 10. Amendment — ListPublicFolders returns folder name

- [x] 10.1 Add `FolderShareListItem` struct to `usecase/folder_share_repository.go`; change `ListByUserID` return type from `[]*domain.FolderShare` to `[]*FolderShareListItem`
- [x] 10.2 Update `ListByUserID` in `repository/folder_share_repository.go` to use explicit `SELECT folder_shares.folder_id, folder_shares.token, folders.name AS folder_name` with `Table("folder_shares")` and `Scan` into `[]*usecase.FolderShareListItem`
- [x] 10.3 Add `FolderName string` to `FolderShareSummary` in `share_usecase.go`; map `item.FolderName` in `GetPublicFoldersByUser`
- [x] 10.4 Add `FolderName string \`json:"folder_name"\`` to `folderShareSummaryResponse` in `handler/internal.go`; update the mapping loop to set `FolderName`
- [x] 10.5 Update `fakeFolderShareRepo.ListByUserID` in `fakes_test.go` to return `[]*FolderShareListItem` (map from stored `domain.FolderShare` including `Folder.Name`)
- [x] 10.6 Update `GetPublicFoldersByUser` usecase tests: seed `Folder.Name` in test shares; assert `FolderName` in returned summaries
- [x] 10.7 Update `ListPublicFolders` handler tests: set `FolderName` in spy summaries; assert `folder_name` in response body
- [x] 10.8 Run `golangci-lint run ./...` from `backend/` and fix any issues
