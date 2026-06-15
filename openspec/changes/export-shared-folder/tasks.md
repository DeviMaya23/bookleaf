## 1. ShareUsecase — GetSharedFolderInfo

- [x] 1.1 Add `GetSharedFolderInfo(ctx context.Context, token string) (*domain.Folder, error)` to `ShareUsecase` in `backend/internal/usecase/share_usecase.go`: call `folderShareRepo.GetByToken(ctx, token)`, return `gorm.ErrRecordNotFound` if not found, otherwise return `&share.Folder`.

## 2. ShareHandler — FolderExporter Interface and Export Handler

- [x] 2.1 Add the `FolderExporter` interface to `backend/internal/handler/share.go`: `ExportFolder(ctx context.Context, folderID uuid.UUID, userID string, w io.Writer) error`.
- [x] 2.2 Add a `folderExporter FolderExporter` field to `ShareHandler` and accept it as a new parameter in `NewShareHandler`.
- [x] 2.3 Add `ShareUsecase.GetSharedFolderInfo(ctx context.Context, token string) (*domain.Folder, error)` to the `ShareUsecase` interface in `share.go`.
- [x] 2.4 Implement `ShareHandler.ExportSharedFolder(c echo.Context) error`: extract `:token`, call `GetSharedFolderInfo`, map `gorm.ErrRecordNotFound` to `404`, derive filename via the existing `sanitizeFilename` helper (from `folder.go`), set `Content-Type: application/zip` and `Content-Disposition`, write `200`, then call `folderExporter.ExportFolder(ctx, folder.ID, folder.UserID, c.Response())`, logging any error that occurs after the response has started.

## 3. Routing and Wiring

- [x] 3.1 In `backend/cmd/server/main.go`, pass `folderUsecase` as the `FolderExporter` argument to `NewShareHandler`.
- [x] 3.2 Register `e.GET("/share/:token/export", shareHandler.ExportSharedFolder)` outside the `protected` group, alongside `e.GET("/share/:token", ...)`.

## 4. Bruno

- [x] 4.1 Add `bruno/share/export-shared-folder.bru` for `GET /share/:token/export`, modeled on `bruno/share/get-shared-folder.bru` and `bruno/folders/export-folder.bru`.

## 5. Tests

- [x] 5.1 In `backend/internal/usecase/share_usecase_test.go`, add unit tests for `GetSharedFolderInfo`: returns folder info for a valid token; returns `gorm.ErrRecordNotFound` for an unknown token.
- [x] 5.2 In `backend/internal/handler/share_test.go`, add a value-return spy `FolderExporter` (alongside the existing `ShareUsecase` spy) and unit tests for `ExportSharedFolder`: `200` with `Content-Type: application/zip` and `Content-Disposition: attachment; filename="<folder name>.zip"`; filename sanitization for names with invalid characters and the `export.zip` fallback for names that sanitize to empty; `404` for unknown token with `FolderExporter.ExportFolder` not called.

## 6. Lint

- [x] 6.1 Run `golangci-lint run` from `backend/` and fix any issues introduced by this change.
