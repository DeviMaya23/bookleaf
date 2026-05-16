## 1. Repository

- [x] 1.1 Add `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 1.2 Add `HardDelete(ctx context.Context, id uuid.UUID, userID string) error` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 1.3 Implement `ListExpiredTrash` in `internal/repository/image_repository.go` — queries `Unscoped()` where `deleted_at IS NOT NULL AND deleted_at < olderThan`
- [x] 1.4 Implement `HardDelete` in `internal/repository/image_repository.go` — `Unscoped().Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Image{})`

## 2. Usecase

- [x] 2.1 Add `PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error` to `ImageUsecase` interface in `internal/usecase/image_usecase.go`
- [x] 2.2 Implement `PurgeExpiredTrash`: call `ListExpiredTrash`, then for each record attempt `store.DeleteObject(r2_path)` (log warn on error), attempt `store.DeleteObject(thumbnail_path)` if not nil (log warn on error), then call `imageRepo.HardDelete`

## 3. Server Startup

- [x] 3.1 Add goroutine to `cmd/server/main.go` alongside the existing stale-upload cleanup goroutine — `time.NewTicker(24 * time.Hour)`, calls `imageUsecase.PurgeExpiredTrash(ctx, 30*24*time.Hour)` on each tick, logs errors at warn level

## 4. Unit Tests

- [x] 4.1 Add unit test for `PurgeExpiredTrash` success scenario — mock `ListExpiredTrash` returning expired records with and without thumbnails, verify `DeleteObject` is called for `r2_path` and `thumbnail_path` where present, and `HardDelete` is called for each
- [x] 4.2 Add unit test for `PurgeExpiredTrash` failure scenario — mock `ListExpiredTrash` returning an error, verify no deletes are attempted
- [x] 4.3 Update handler and folder usecase test mocks to add `PurgeExpiredTrash` and `HardDelete` stub methods to satisfy updated interfaces
