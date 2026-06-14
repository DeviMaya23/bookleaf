## 1. Backend: FolderImageRepository interface

- [x] 1.1 Rename the `ImageCounter` interface in `internal/usecase/folder_usecase.go` to `FolderImageRepository` and add `ListByFolder(ctx, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)`
- [x] 1.2 Update `folderUsecase` field type and `NewFolderUsecase` signature/wiring in `cmd/server/main.go` to the renamed interface
- [x] 1.3 Update any existing fake/test double implementing the old `ImageCounter` interface to satisfy `FolderImageRepository`

## 2. Backend: ExportFolder usecase

- [x] 2.1 Add `store usecase.StorageService` as a constructor dependency to `folderUsecase` / `NewFolderUsecase`, and wire the existing `r2Storage` instance into it in `cmd/server/main.go`
- [x] 2.2 Implement `ExportFolder(ctx, folderID uuid.UUID, userID string, w io.Writer) error` on `folderUsecase`: list images via `ListByFolder`, build a `zip.Writer` over `w`, derive entry names from sanitized titles + `downloadFileExtension`, dedup colliding names with ` (1)`, ` (2)`, etc., copy each image's bytes from `store.GetObject`
- [x] 2.3 Add `ExportFolder` to the `FolderUsecase` interface in `internal/handler/folder.go`
- [x] 2.4 Unit tests for `ExportFolder` per `specs/folder-export/spec.md`: entry naming, dedup, title sanitization, empty folder, `ListByFolder` error, `GetObject` error — using a fake `FolderImageRepository` and a value-return spy `StorageService`

## 3. Backend: handler, route, and Bruno

- [x] 3.1 Implement `FolderHandler.ExportFolder`: parse UUID (400), call `GetByID` for ownership (404), derive sanitized archive filename with `export.zip` fallback, set `Content-Type`/`Content-Disposition`, stream via `ExportFolder`
- [x] 3.2 Register `GET /folders/:id/export` on the protected group in `cmd/server/main.go`
- [x] 3.3 Handler unit tests per `specs/folder-export/spec.md`: success headers, 400 invalid UUID, 404 not found/not owned, filename sanitization and empty-name fallback — using a value-return spy `FolderUsecase`
- [x] 3.4 Add `bruno/folders/export-folder.bru` for `GET {{baseUrl}}/folders/{{folderId}}/export`

## 4. Frontend: data layer

- [x] 4.1 Add `FolderDetail` type and `getFolder(getToken, id)` to `frontend/src/lib/folders.ts`, calling `GET /folders/:id`
- [x] 4.2 Add `exportFolder(getToken, id): Promise<Blob>` to `frontend/src/lib/folders.ts`, calling `GET /folders/:id/export`

## 5. Frontend: Export folder button

- [x] 5.1 In `FolderPanelContent`, add a `useQuery(['folder', folder.id], () => getFolder(getToken, folder.id))` for the folder detail/`image_count`
- [x] 5.2 Add an "Export folder" button (sticky footer, mirroring `DownloadButton`'s style/placement) with disabled (when `image_count` is 0 or detail is loading), "Preparing export..." loading, and error-toast states
- [x] 5.3 On click: call `exportFolder`, convert the resulting `Blob` to an object URL, trigger download via a temporary `<a download="<folder name>.zip">`, then revoke the object URL

## 6. Verification

- [x] 6.1 Run `golangci-lint run` in `backend/` and fix any issues
- [x] 6.2 Run `npm run build` and `npm run lint` in `frontend/` and fix any issues
