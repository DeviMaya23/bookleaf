## 1. Storage Layer

- [x] 1.1 Add `GeneratePresignedDownloadURL(ctx context.Context, key, filename string, ttl time.Duration) (string, error)` to the `StorageService` interface in `internal/storage/`
- [x] 1.2 Implement `GeneratePresignedDownloadURL` on the R2 concrete struct, setting `ResponseContentDisposition` to `attachment; filename="<filename>"` on the presign request
- [x] 1.3 Update the `StorageService` mock (used in usecase unit tests) to include `GeneratePresignedDownloadURL`

## 2. Usecase Layer

- [x] 2.1 Add `DownloadImage(ctx context.Context, id uuid.UUID, userID string) (string, error)` to the `ImageUsecase` interface in `internal/usecase/`
- [x] 2.2 Implement `DownloadImage` on the concrete `imageUsecase`: fetch image via `GetByID`, derive filename from title + MIME type, call `store.GeneratePresignedDownloadURL` with 5-minute TTL, return the URL
- [x] 2.3 Write unit tests for `DownloadImage`: success scenario and failure scenario (image not found)

## 3. Handler Layer

- [x] 3.1 Add `DownloadImage` handler method to the image handler struct: parse `:id` UUID, call usecase, return `200 OK` with `{ "download_url": "..." }` or appropriate error status
- [x] 3.2 Write unit tests for the `DownloadImage` handler: success scenario and failure scenario (not found)

## 4. Route Registration

- [x] 4.1 Register `GET /images/:id/download` on the protected Echo group in `main.go`
