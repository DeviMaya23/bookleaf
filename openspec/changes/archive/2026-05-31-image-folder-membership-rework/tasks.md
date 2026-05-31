## 1. Repository Layer

- [x] 1.1 Add `SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 1.2 Add `MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error` to `ImageRepository` interface
- [x] 1.3 Implement `SyncImageFolders` in `internal/repository/image_repository.go` (fetch current memberships, diff, delete departed rows, insert new rows with fracdex positions, all in a transaction)
- [x] 1.4 Implement `MoveImageFolder` in `internal/repository/image_repository.go` (delete from-folder row if non-nil, insert to-folder row with fracdex position if non-nil, in a transaction)

## 2. Usecase Layer

- [x] 2.1 Replace `FolderID **uuid.UUID` with `FolderIDs *[]uuid.UUID` in `UpdateImageParams` in `internal/usecase/image_usecase.go`
- [x] 2.2 Update `UpdateImage` usecase to call `SyncImageFolders` when `params.FolderIDs` is non-nil, replacing the existing `SetImageFolder` call
- [x] 2.3 Add `MoveImageFolder(ctx context.Context, imageID uuid.UUID, userID string, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error` to `ImageUsecase` interface
- [x] 2.4 Implement `MoveImageFolder` usecase method (ownership check, no-op if from == to, delegate to repo)

## 3. Handler Layer

- [x] 3.1 Update `EditImage` handler in `internal/handler/image.go` to decode `folder_ids` as `json.RawMessage` (same pattern as `tags`), populating `params.FolderIDs`; remove `folder_id` singular field decoding
- [x] 3.2 Add `MoveImageFolder` handler method in `internal/handler/image.go` — decode body `{ from_folder_id, to_folder_id }`, call usecase, return updated image on 200
- [x] 3.3 Register `POST /images/:id/move-folder` route in the server router

## 4. Unit Tests

- [x] 4.1 Add handler unit tests for `MoveImageFolder` — success (200 OK) and failure (404 when usecase returns ErrRecordNotFound)
- [x] 4.2 Add usecase unit tests for `MoveImageFolder` — success (repo called with correct args) and failure (image not found, repo not called)
- [x] 4.3 Update existing `UpdateImage` handler unit tests to use `folder_ids` instead of `folder_id`
- [x] 4.4 Update existing `UpdateImage` usecase unit tests to use `FolderIDs` instead of `FolderID`

## 5. Bruno

- [x] 5.1 Add `move-image-to-folder.bru` to `bruno/images/` with `POST /images/:id/move-folder` and body `{ from_folder_id, to_folder_id }`
- [x] 5.2 Update `update-image.bru` to use `folder_ids` array instead of `folder_id`

## 6. Frontend

- [x] 6.1 Add `moveImageFolder(getToken, imageId, fromFolderId, toFolderId)` function to `frontend/src/lib/images.ts`
- [x] 6.2 Update `onDragEnd` in `frontend/src/App.tsx` to call `moveImageFolder` with `currentFolderId` (from drag item data) and `targetFolderId` (from drop target)
- [x] 6.3 Update drop-to-unsorted handler to call `moveImageFolder` with `from_folder_id: currentFolderId` and `to_folder_id: null`
- [x] 6.4 Update no-op guard in `onDragEnd` — skip if `currentFolderId == targetFolderId` (including both null)
