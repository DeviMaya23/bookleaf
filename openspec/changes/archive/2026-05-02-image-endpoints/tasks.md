## 1. Config & Domain

- [x] 1.1 Add `R2Config` struct to `internal/config/config.go` — load `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL` as required env vars
- [x] 1.2 Update config tests to cover new R2 required vars
- [x] 1.3 Fix `Image` GORM struct folder FK tag in `internal/domain/image.go` — change `OnDelete:SET NULL` to `OnDelete:RESTRICT`

## 2. Storage Service

- [x] 2.1 Add AWS SDK v2 and `disintegration/imaging` to `go.mod` (`go get`)
- [x] 2.2 Create `internal/storage/storage.go` — define `StorageService` interface with `GeneratePresignedPutURL`, `GeneratePresignedGetURL`, `GetObject`, `PutObject`, `CDNUrl`
- [x] 2.3 Create `internal/storage/r2.go` — implement `StorageService` using AWS SDK v2 S3-compatible client pointed at `https://{R2_ACCOUNT_ID}.r2.cloudflarestorage.com`
- [x] 2.4 Implement `CDNUrl` — return `{R2_PUBLIC_URL}/{key}`
- [x] 2.5 Implement `GeneratePresignedPutURL` — S3 presign PUT with given TTL
- [x] 2.6 Implement `GeneratePresignedGetURL` — S3 presign GET with given TTL
- [x] 2.7 Implement `GetObject` — S3 GetObject returning body reader
- [x] 2.8 Implement `PutObject` — S3 PutObject with content type header
- [x] 2.9 Add helper `MimeTypeToExt(mimeType string) string` — maps `image/jpeg`→`.jpg`, `image/png`→`.png`, `image/webp`→`.webp`, `image/gif`→`.gif`, fallback `.jpg`

## 3. Thumbnail Service

- [x] 3.1 Create `internal/thumbnail/thumbnail.go` — define `ThumbnailService` interface with `Generate(ctx, src io.Reader) (io.Reader, error)`
- [x] 3.2 Implement `Generate` using `disintegration/imaging` — fit within 600x600 preserving aspect ratio, `imaging.Lanczos` filter, encode output as JPEG

## 4. Repository

- [x] 4.1 Create `internal/usecase/image_repository.go` — define `ImageRepository` interface with `Create`, `List`, `GetByID`, `GetDeletedByID`, `UpdateThumbnailPath`, `SoftDelete`, `Restore`, `ListTrashed`
- [x] 4.2 Create `internal/repository/image_repository.go` — implement `ImageRepository`; all queries scoped by `userID` except `UpdateThumbnailPath`
- [x] 4.3 Implement `Create` — insert image row, return created image
- [x] 4.4 Implement `List` — select non-deleted images for a user, optionally filtered by `folder_id`, ordered by `created_at DESC`
- [x] 4.5 Implement `GetByID` — select non-deleted image by `id` and `user_id`; return `gorm.ErrRecordNotFound` on miss
- [x] 4.6 Implement `GetDeletedByID` — use `db.Unscoped()` to select a soft-deleted image by `id` and `user_id`
- [x] 4.7 Implement `UpdateThumbnailPath` — update `thumbnail_path` by `id` (no user scope; called internally)
- [x] 4.8 Implement `SoftDelete` — set `deleted_at` via GORM delete scoped by `id` and `user_id`
- [x] 4.9 Implement `Restore` — use `db.Unscoped()` to clear `deleted_at` by `id` and `user_id`
- [x] 4.10 Implement `ListTrashed` — use `db.Unscoped().Where("deleted_at IS NOT NULL AND user_id = ?")` ordered by `deleted_at DESC`

## 5. Usecase

- [x] 5.1 Create `internal/usecase/image_usecase.go` — define `ImageUsecase` interface and `imageUsecase` struct (depends on `ImageRepository`, `StorageService`, `ThumbnailService`)
- [x] 5.2 Implement `InitiateUpload` — validate title/mime_type, generate UUID, build R2 key, create DB record, call `GeneratePresignedPutURL` (15min TTL), return ID + upload URL + r2_path
- [x] 5.3 Implement `CompleteUpload` — verify ownership via `GetByID`, fire goroutine: fetch original via `GetObject`, call `ThumbnailService.Generate`, `PutObject` thumbnail, `UpdateThumbnailPath`
- [x] 5.4 Implement `ListImages` — delegate to repository `List` with optional folder filter
- [x] 5.5 Implement `GetImage` — delegate to `GetByID`, call `GeneratePresignedGetURL` (24hr TTL), return image + presigned URL
- [x] 5.6 Implement `SoftDelete` — delegate to repository `SoftDelete`, propagate not-found
- [x] 5.7 Implement `ListTrashed` — delegate to repository `ListTrashed`
- [x] 5.8 Implement `Restore` — delegate to repository `GetDeletedByID` (404 if not found/not deleted), then `Restore`

## 6. Handler

- [x] 6.1 Create `internal/handler/image.go` — define `ImageHandler` struct and constructor
- [x] 6.2 Implement `InitiateUpload` handler — parse body, extract userID, call usecase, return 201
- [x] 6.3 Implement `CompleteUpload` handler — parse `:id`, extract userID, call usecase, return 204
- [x] 6.4 Implement `ListImages` handler — extract userID + optional `folder_id` query param, call usecase, return 200
- [x] 6.5 Implement `GetImage` handler — parse `:id`, extract userID, call usecase, return 200 or 404
- [x] 6.6 Implement `SoftDelete` handler — parse `:id`, extract userID, call usecase, return 204 or 404
- [x] 6.7 Implement `ListTrashed` handler — extract userID, call usecase, return 200
- [x] 6.8 Implement `Restore` handler — parse `:id`, extract userID, call usecase, return 200 or 404

## 7. Routing

- [x] 7.1 Wire `StorageService`, `ThumbnailService`, `ImageRepository`, `ImageUsecase`, `ImageHandler` in `cmd/server/main.go`
- [x] 7.2 Register all 7 image routes on the protected route group

## 8. Tests

- [x] 8.1 Create `internal/usecase/image_usecase_test.go` — table-driven unit tests with mocked `ImageRepository`, `StorageService`, `ThumbnailService`; success + failure per method
- [x] 8.2 Create `internal/handler/image_test.go` — table-driven unit tests with mocked `ImageUsecase`; success + failure per handler method
- [x] 8.3 Create `internal/repository/image_repository_integration_test.go` — Testcontainers integration tests; success + failure per repository method
