## 1. Dependency

- [x] 1.1 Add `github.com/rocicorp/fracdex` to `backend/go.mod` via `go get`

## 2. Repository Layer

- [x] 2.1 Update `SetImageFolder` in `image_repository.go`: replace `MAX(position::int)` with `fracdex.KeyBetween(maxPosition, "")` where `maxPosition` is fetched as the lexicographic MAX of the position column (empty string if no rows)
- [x] 2.2 Add `UpdateImageFolderPosition(ctx, imageID, folderID, position)` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 2.3 Implement `UpdateImageFolderPosition` in `image_repository.go`: targeted `UPDATE image_folders SET position = ? WHERE image_id = ? AND folder_id = ?`, return wrapped `ErrRecordNotFound` if no rows affected
- [x] 2.4 Update `List` in `image_repository.go`: when `folderID` is non-nil, JOIN `image_folders` and `ORDER BY image_folders.position ASC` with no cursor filter and no limit; keep existing `created_at DESC` path for nil folderID
- [x] 2.5 Write integration tests for `UpdateImageFolderPosition` (row exists → updated; row missing → ErrRecordNotFound)
- [x] 2.6 Write integration tests for updated `List` (folder view returns all images ordered by position; all-view still paginates by created_at)

## 3. Usecase Layer

- [x] 3.1 Add `UpdateImagePosition(ctx, imageID, userID, folderID, position)` to `ImageUsecase` interface in `internal/usecase/image_usecase.go`
- [x] 3.2 Implement `UpdateImagePosition`: verify image ownership via `GetByID`, delegate to `imageRepo.UpdateImageFolderPosition`, propagate errors
- [x] 3.3 Update `ListImages` usecase: when `params.FolderID` is non-nil, pass `nil` cursor and `0` limit to `imageRepo.List` and set `NextCursor` to nil in the result
- [x] 3.4 Write unit tests for `UpdateImagePosition` (success path; image not found)
- [x] 3.5 Write unit tests for `ListImages` with folder (no cursor/limit applied; NextCursor is nil)

## 4. Handler Layer

- [x] 4.1 Add `position *string` to `imageResponse` struct in `internal/handler/image.go`
- [x] 4.2 Update `toImageResponse`: populate `Position` from `ImageFolders[0].Position` when `ImageFolders` is non-empty
- [x] 4.3 Update `ListImages` handler: skip cursor parsing when `folder_id` query param is present
- [x] 4.4 Add `UpdateImagePosition` handler in `internal/handler/image.go`: parse `:id`, bind and validate `{ folder_id, position }` body, call usecase, return `204 No Content`
- [x] 4.5 Write unit tests for `UpdateImagePosition` handler (success → 204; missing position → 400; missing folder_id → 400; image not found → 404)
- [x] 4.6 Write unit tests for `ListImages` handler folder path (cursor param ignored; next_cursor null in response)

## 5. Router

- [x] 5.1 Register `PATCH /images/:id/position` route in `cmd/server/main.go` before `PATCH /images/:id`
- [x] 5.2 Create Bruno file for `PATCH /images/:id/position` endpoint

## 6. Position Rebalance Script

- [x] 6.1 Create `cmd/migrate-positions/main.go`: connect via `DATABASE_URL`, iterate `image_folders` grouped by `folder_id` ordered by `position ASC`, generate fracdex keys with `fracdex.KeyBetween`, update rows in a per-folder transaction
