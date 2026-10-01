## 1. Backend — Repository

- [x] 1.1 Add `GetManyByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]*domain.Image, error)` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 1.2 Implement `GetManyByIDs` in `internal/repository/image_repository.go` — single `WHERE id IN (...) AND user_id = ? AND deleted_at IS NULL` query with Tags and ImageFolders preloaded

## 2. Backend — Shared Zip Helper

- [x] 2.1 Extract `writeImagesToZip(ctx context.Context, images []*domain.Image, store StorageService, w io.Writer) error` as a package-level function in the `usecase` package (new file `internal/usecase/zip.go`)
- [x] 2.2 Refactor `folderUsecase.ExportFolder` to delegate the zip loop to `writeImagesToZip`

## 3. Backend — BulkExportImages Usecase

- [x] 3.1 Add `BulkExportImages(ctx context.Context, userID uuid.UUID, imageIDs []uuid.UUID, w io.Writer) error` to `ImageUsecase` interface
- [x] 3.2 Implement `BulkExportImages` in `imageUsecase`: call `GetManyByIDs` then `writeImagesToZip`
- [x] 3.3 Write unit tests for `imageUsecase.BulkExportImages` (zip contents for owned images, unowned IDs dropped, error from `GetManyByIDs` propagated)

## 4. Backend — Handler & Route

- [x] 4.1 Add `BulkExport` handler to `ImageHandler` in `internal/handler/image.go` — parse `image_ids`, set zip headers, stream via `BulkExportImages`
- [x] 4.2 Register `POST /images/bulk/export` in `cmd/server/main.go` alongside existing `/images/bulk/` routes
- [x] 4.3 Write unit tests for `ImageHandler.BulkExport` (200 with correct headers, 400 for malformed UUID)
- [x] 4.4 Create Bruno file for `POST /images/bulk/export`

## 5. Frontend — Lib Function

- [x] 5.1 Add `bulkExportImages(getToken: GetToken, imageIds: string[]): Promise<Blob>` to `src/lib/images.ts`

## 6. Frontend — SelectionPanelBody Revamp

- [x] 6.1 Add `onExitSelectMode` and `onDownloadZip` props to `SelectionPanelBody`; forward `onExitSelectMode` from `RightPanel.deriveContent`
- [x] 6.2 Replace folder section: swap `SelectionFolderPicker` for `FolderInput` with local chip state and a disabled-when-empty Apply button; pass `allFolders` from a `useQuery` call already in `SelectionPanelBody`
- [x] 6.3 Add Apply click handler that calls `onAddToFolder` once per selected folder chip, then clears chips
- [x] 6.4 Add × icon button inline with the "N selected" count in the header row
- [x] 6.5 Add "Download as ZIP" button with loading state; call `onDownloadZip` on click; show error toast on failure
- [x] 6.6 Delete `SelectionFolderPicker.tsx`

## 7. Frontend — AppLayout Download Handler

- [x] 7.1 Add `handleDownloadSelection` in `AppLayout`: call `bulkExportImages`, create blob URL, trigger `<a>` download with filename `bookleaf-export.zip`, show error toast on failure
- [x] 7.2 Pass `onDownloadZip: handleDownloadSelection` through `PanelContent.selection` type and into `SelectionPanelBody`

## 8. Linting & Build

- [x] 8.1 Run `golangci-lint run` and fix any issues
- [x] 8.2 Run `npm run build` and `npm run lint` and fix any issues
