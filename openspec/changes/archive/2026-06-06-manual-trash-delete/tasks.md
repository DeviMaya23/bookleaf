## 1. Repository Layer

- [x] 1.1 ~~Add `GetTrashedImage` to `ImageRepository` interface~~ — reuse existing `GetDeletedByID`
- [x] 1.2 Add `ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error)` to `ImageRepository` interface in `usecase/image_repository.go`
- [x] 1.3 ~~Implement `GetTrashedImage` in `repository/image_repository.go`~~ — `GetDeletedByID` already exists at repository/image_repository.go:244
- [x] 1.4 Implement `ListAllTrashed` in `repository/image_repository.go` — unscoped query with `deleted_at IS NOT NULL AND user_id = ?`, no limit/cursor

## 2. Usecase Layer

- [x] 2.1 Add `DeleteFromTrash(ctx context.Context, id uuid.UUID, userID string) error` to `ImageUsecase` interface in `handler/image.go`
- [x] 2.2 Add `EmptyTrash(ctx context.Context, userID string) error` to `ImageUsecase` interface in `handler/image.go`
- [x] 2.3 Implement `DeleteFromTrash` in `usecase/image_usecase.go` — fetch via `GetDeletedByID`, delete R2 object, delete thumbnail (best-effort), hard-delete DB record
- [x] 2.4 Implement `EmptyTrash` in `usecase/image_usecase.go` — fetch all via `ListAllTrashed`, loop with same deletion sequence as `PurgeExpiredTrash`

## 3. Handler Layer

- [x] 3.1 Implement `DeleteFromTrash` handler in `handler/image.go` — parse `:id` UUID, call usecase, return 204; map not-found to 404 and invalid UUID to 400
- [x] 3.2 Implement `EmptyTrash` handler in `handler/image.go` — call usecase, return 204

## 4. Routing

- [x] 4.1 Register `DELETE /images/trash/:id` → `imageHandler.DeleteFromTrash` in `cmd/server/main.go`
- [x] 4.2 Register `DELETE /images/trash` → `imageHandler.EmptyTrash` in `cmd/server/main.go`

## 5. Tests

- [x] 5.1 Write usecase unit tests for `DeleteFromTrash` in `usecase/image_usecase_test.go`:
  - Success: image found in trash, R2 deleted, DB hard-deleted
  - Not found: image not in trash, returns not-found error
- [x] 5.2 Write usecase unit tests for `EmptyTrash` in `usecase/image_usecase_test.go`:
  - Success: trashed images found, all deleted
  - No-op: no trashed images, returns nil
- [x] 5.3 Write handler unit tests for `DeleteFromTrash` in `handler/image_test.go`:
  - 204 on success
  - 404 when usecase returns not-found
  - 400 on malformed UUID
- [x] 5.4 Write handler unit tests for `EmptyTrash` in `handler/image_test.go`:
  - 204 on success

## 6. Bruno

- [x] 6.1 Create `bruno/images/delete-from-trash.bru` for `DELETE /images/trash/:id`
- [x] 6.2 Create `bruno/images/empty-trash.bru` for `DELETE /images/trash`
