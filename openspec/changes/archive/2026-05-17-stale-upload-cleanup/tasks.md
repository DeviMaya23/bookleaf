## 1. Database Migration

- [x] 1.1 Create migration `000008_add_image_is_uploaded.up.sql` — adds `is_uploaded boolean NOT NULL DEFAULT false` to `images` table
- [x] 1.2 Create migration `000008_add_image_is_uploaded.down.sql` — drops `is_uploaded` column from `images` table

## 2. Domain

- [x] 2.1 Add `IsUploaded bool` field to `Image` struct in `internal/domain/image.go` with GORM tag `column:is_uploaded;not null;default:false`

## 3. Storage

- [x] 3.1 Add `DeleteObject(ctx context.Context, key string) error` to `StorageService` interface in `internal/storage/storage.go`
- [x] 3.2 Implement `DeleteObject` on `r2Storage` in `internal/storage/r2.go` using the S3 `DeleteObject` API call

## 4. Repository

- [x] 4.1 Add `ListStaleUploads(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)` to `ImageRepository` interface in `internal/usecase/image_repository.go`
- [x] 4.2 Implement `ListStaleUploads` in `internal/repository/image_repository.go` — queries `WHERE is_uploaded = false AND created_at < olderThan AND deleted_at IS NULL`

## 5. Usecase

- [x] 5.1 Update `CompleteUpload` in `internal/usecase/image_usecase.go` to set `is_uploaded = true` on the image record (via `Update` with `fields["is_uploaded"] = true`)
- [x] 5.2 Add `CleanupStaleUploads(ctx context.Context, threshold time.Duration) error` to `ImageUsecase` interface in `internal/usecase/image_usecase.go`
- [x] 5.3 Implement `CleanupStaleUploads`: call `ListStaleUploads`, then for each record attempt `store.DeleteObject(r2_path)` (log warn on error), then call `imageRepo.SoftDelete` using the record's own `UserID`

## 6. Server Startup

- [x] 6.1 Add goroutine to `cmd/server/main.go` after `imageUsecase` is wired — `time.NewTicker(10 * time.Minute)`, calls `imageUsecase.CleanupStaleUploads(ctx, 30*time.Minute)` on each tick, logs errors at warn level

## 7. Unit Tests

- [x] 7.1 Add unit test for `CleanupStaleUploads` success scenario — mock `ListStaleUploads` returning stale records, verify `DeleteObject` and `SoftDelete` are called for each
- [x] 7.2 Add unit test for `CleanupStaleUploads` failure scenario — mock `ListStaleUploads` returning an error, verify no deletes are attempted
- [x] 7.3 Update `CompleteUpload` unit test to assert `is_uploaded = true` is included in the update fields
