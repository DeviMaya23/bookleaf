## 1. Repository layer

- [x] 1.1 Add `AddImageToFolder(ctx context.Context, imageID, folderID uuid.UUID) error` to the `ImageRepository` interface in `internal/usecase/image_repository.go`, documented as idempotent (no-op if the membership row already exists)
- [x] 1.2 Implement `AddImageToFolder` in `internal/repository/image_repository.go` — compute the next fracdex position via `MAX(position)` for the folder (same approach as `SetImageFolder`), then insert with `clause.OnConflict{Columns: []clause.Column{{Name: "image_id"}, {Name: "folder_id"}}, DoNothing: true}`
- [x] 1.3 Add `FilterOwnedImageIDs(ctx context.Context, ids []uuid.UUID, userID string) ([]uuid.UUID, error)` to `ImageRepository` — returns the subset of `ids` that exist and belong to `userID` (used by both bulk usecases to pre-filter before the per-item loop); exclude soft-deleted rows for the add-to-folder case via the default GORM scope
- [x] 1.4 Implement `FilterOwnedImageIDs` in `internal/repository/image_repository.go`
- [x] 1.5 Add integration tests for `AddImageToFolder` (insert when absent, no-op when already present, position appended after existing max) in `internal/repository/image_repository_integration_test.go`
- [x] 1.6 Add integration tests for `FilterOwnedImageIDs` (returns only owned/existing ids, empty input, all unowned) in `internal/repository/image_repository_integration_test.go`

## 2. Usecase layer

- [x] 2.1 Add `BulkAddToFolder(ctx context.Context, userID string, imageIDs []uuid.UUID, folderID uuid.UUID) (int, error)` to `ImageUsecase` interface in `internal/handler/image.go`
- [x] 2.2 Implement `BulkAddToFolder` in `internal/usecase/image_usecase.go`: validate folder via existing `folderRepo.GetByID` (404-equivalent error if missing/unowned), call `FilterOwnedImageIDs`, loop the valid subset calling `AddImageToFolder` per id — on a per-id error, log via `observability.LoggerFromContext` and continue, otherwise increment the success count; return the final count
- [x] 2.3 Add `BulkTrash(ctx context.Context, userID string, imageIDs []uuid.UUID) (int, error)` to `TrashUsecase` interface in `internal/handler/trash.go`
- [x] 2.4 Implement `BulkTrash` in `internal/usecase/trash_usecase.go`: call `imageRepo.FilterOwnedImageIDs`, loop the valid subset calling the existing `SoftDelete` per id — on a per-id error (including already-trashed, which `SoftDelete` already reports as not-found), log and continue, otherwise increment the success count; return the final count
- [x] 2.5 Unit test `BulkAddToFolder`: all succeed; folder not found/unowned (returns error, no images processed); one image unowned (others still succeed, count excludes it); one image already in folder (counts as success, no error)
- [x] 2.6 Unit test `BulkTrash`: all succeed; one image already trashed (counts excluded, others succeed); one image unowned (counts excluded, others succeed); all images invalid (returns 0, no error)

## 3. Handler layer

- [x] 3.1 Add `BulkAddToFolder` handler method in `internal/handler/image.go`: parse + validate request body (`image_ids []string`, `folder_id string`), return 400 on malformed UUIDs, call usecase, return `200 {"succeeded_count": n}` on success, map folder-not-found error to 404
- [x] 3.2 Add `BulkTrash` handler method in `internal/handler/trash.go`: parse + validate request body (`image_ids []string`), return 400 on malformed UUIDs, call usecase, return `200 {"succeeded_count": n}`
- [x] 3.3 Register routes in `backend/cmd/server/main.go`: `protected.POST("/images/bulk/add-to-folder", imageHandler.BulkAddToFolder)` and `protected.POST("/images/bulk/trash", trashHandler.BulkTrash)`
- [x] 3.4 Unit test `BulkAddToFolder` handler: valid request returns 200 + count; malformed image UUID returns 400; usecase folder-not-found error maps to 404
- [x] 3.5 Unit test `BulkTrash` handler: valid request returns 200 + count; malformed image UUID returns 400

## 4. Mocks

- [x] 4.1 Add `AddImageToFolder` and `FilterOwnedImageIDs` stub methods to the mock `ImageRepository` used in `internal/usecase/image_usecase_test.go` (and any other test file implementing that interface)
- [x] 4.2 Add `FilterOwnedImageIDs` stub method to the mock repository used in `internal/usecase/trash_usecase_test.go`
- [x] 4.3 Add `BulkAddToFolder` and `BulkTrash` stub methods to any mock implementations of `ImageUsecase`/`TrashUsecase` used in handler tests

## 5. API documentation

- [x] 5.1 Add `bruno/images/bulk-add-to-folder.bru` request file (POST, sample `image_ids` + `folder_id` body)
- [x] 5.2 Add `bruno/images/bulk-trash.bru` request file (POST, sample `image_ids` body)

## 6. Verification

- [x] 6.1 Run `golangci-lint run` from `backend/` and fix any issues
- [x] 6.2 Run the full Go test suite (unit + integration) and confirm it passes
