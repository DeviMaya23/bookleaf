## Why

People without an account who receive a folder share link can already view and individually save full-res images via `GET /share/:token`, but there's no way to download an entire shared folder at once. Folder owners already have this via `GET /folders/:id/export`; extending it to the public share view lets recipients grab everything in one click without re-implementing the zip logic.

## What Changes

- Add a new public route `GET /share/:token/export` that streams a zip archive of the shared folder's images, mirroring the existing authenticated `GET /folders/:id/export`.
- Add `ShareUsecase.GetSharedFolderInfo(ctx, token) (*domain.Folder, error)` — a thin wrapper around `FolderShareRepository.GetByToken` that returns the shared folder's `ID`, `Name`, and `UserID`, used to derive the archive filename and to drive the export.
- Add a narrow `FolderExporter` interface on `ShareHandler` (`ExportFolder(ctx, folderID uuid.UUID, userID string, w io.Writer) error`), satisfied implicitly by the existing `folderUsecase`. `ShareHandler.ExportSharedFolder` calls `GetSharedFolderInfo` for the filename, then `FolderExporter.ExportFolder(ctx, folder.ID, folder.UserID, w)` to stream the zip — reusing the existing export logic and zip-building code unchanged.
- Add a Bruno request for the new endpoint.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `folder-sharing`: adds a new public endpoint, `GET /share/:token/export`, for downloading a shared folder's images as a zip archive without authentication.

## Impact

- `backend/internal/usecase/share_usecase.go`: new `GetSharedFolderInfo` method.
- `backend/internal/handler/share.go`: new `FolderExporter` interface, new field on `ShareHandler`, new `ExportSharedFolder` handler.
- `backend/cmd/server/main.go`: wire `folderUsecase` into `NewShareHandler`, register `e.GET("/share/:token/export", shareHandler.ExportSharedFolder)`.
- `bruno/share/`: new request file for the export endpoint.
- No changes to `folderUsecase.ExportFolder`, `ShareUsecase.GetSharedFolder`, or any existing endpoint behavior.
