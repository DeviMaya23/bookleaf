## 1. Backend: route and request contract

- [x] 1.1 Change route registration in `cmd/server/main.go` from `protected.PUT("/folders/:id", folderHandler.UpdateFolder)` to `protected.PATCH("/folders/:id", folderHandler.UpdateFolder)`
- [x] 1.2 Replace the `folderRequest` DTO in `internal/handler/folder.go` with a presence-aware `updateFolderRequest` using `json.RawMessage` for `name`, `parent_id`, `description`, mirroring `updateImageRequest` (`internal/handler/image.go:35-41`)
- [x] 1.3 Rewrite `UpdateFolder` to parse each raw field into "absent / explicit null / value", validate that a present `name` is non-blank (mirroring the `title` check at `internal/handler/image.go:291-293`), and build a `usecase.UpdateFolderParams`

## 2. Backend: usecase merge semantics

- [x] 2.1 Define `UpdateFolderParams{ Name *string; ParentID **uuid.UUID; Description **string }` in the `usecase` package, mirroring `UpdateImageParams` (`internal/usecase/image_usecase.go:21-27`)
- [x] 2.2 Change the `FolderUsecase.Update` method signature (in the `FolderUsecase` interface in `internal/handler/folder.go:23` and the `FolderRepository`/usecase implementation) from `Update(ctx, id, userID, name string, parentID *uuid.UUID, description *string)` to `Update(ctx, id uuid.UUID, userID string, params UpdateFolderParams) (*domain.Folder, error)`
- [x] 2.3 Reimplement `folderUsecase.Update` (`internal/usecase/folder_usecase.go:102-126`) to validate `params.Name` only if non-nil, build a `map[string]any` containing only the non-nil params (mirroring `internal/usecase/image_usecase.go:212-221`), and pass it to `folderRepo.Update`

## 3. Backend: repository selective update

- [x] 3.1 Change the `FolderRepository.Update` interface method from `Update(ctx, folder *domain.Folder) (*domain.Folder, error)` to `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Folder, error)`
- [x] 3.2 Rewrite `folderRepository.Update` (`internal/repository/folder_repository.go:69-87`) to drop the pre-fetch and run `Model(&domain.Folder{}).Where("id = ? AND user_id = ?", id, userID).Updates(fields)`, returning a wrapped `gorm.ErrRecordNotFound` when `RowsAffected == 0` and re-fetching via `GetByID` on success — mirroring `imageRepository.Update` (`internal/repository/image_repository.go:302-315`)

## 4. Backend: tests

- [x] 4.1 Rewrite `TestFolderRepository_Update_PersistsFields` (integration test) to assert partial-merge behavior: a field omitted from `fields` is preserved, a field present as explicit `nil`/`null` is cleared, and a field present with a value is overwritten — covering `name`, `parent_id`, and `description` independently
- [x] 4.2 Add/update `folderUsecase.Update` unit tests: rejects a present-but-blank `Name` without calling the repository; passes through a fields map containing only the non-nil params; surfaces repository not-found errors
- [x] 4.3 Add/update `FolderHandler.UpdateFolder` unit tests: correctly distinguishes an absent field, an explicit `null`, and a provided value for `parent_id`/`description`; returns `400` for a present-but-blank `name`; returns `404` for a missing/unowned folder; returns `401` when unauthenticated

## 5. Backend: bruno and lint

- [x] 5.1 Update `bruno/folders/update-folder.bru` to use `PATCH` and a partial-body example (e.g. `{ "name": "..." }` rather than the full three-field body)
- [x] 5.2 Run `golangci-lint run` from the backend module and fix any issues raised

## 6. Frontend: consolidate folder update wrapper

- [x] 6.1 In `frontend/src/lib/folders.ts`, remove `renameFolder`, `moveFolder`, and `updateFolderDetails`, replacing them with a single `updateFolder(getToken, id, params: { name?: string; description?: string | null; parent_id?: string | null })` that issues a `PATCH` request with `JSON.stringify(params)`, mirroring `updateImage`/`UpdateImageParams` (`frontend/src/lib/images.ts:182-198`)

## 7. Frontend: update call sites

- [x] 7.1 `FolderSidebar.tsx:278` — change the rename mutation to call `updateFolder(getToken, id, { name })`
- [x] 7.2 `FolderPanelContent.tsx:31-32` — change the mutation to call `updateFolder(getToken, folder.id, params)`, updating the `Parameters<typeof updateFolderDetails>[2]` type reference to the new wrapper
- [x] 7.3 `dragHandlers.ts:68,73` — change both `moveFolder` calls to `updateFolder(getToken, drag.folderId, { parent_id: drop.folderId })` and `updateFolder(getToken, drag.folderId, { parent_id: null })` respectively, dropping the now-unnecessary `name` field

## 8. Frontend: tests and build

- [x] 8.1 Update `dragHandlers.test.ts` and any other tests referencing `renameFolder`/`moveFolder`/`updateFolderDetails` or asserting full-body `PUT` calls, to assert the new `updateFolder` partial-body calls
- [x] 8.2 Run `npm run build` and fix any issues raised
