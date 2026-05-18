## 1. Storage Layer

- [x] 1.1 Remove `CDNUrl` from `StorageService` interface in `internal/storage/storage.go`
- [x] 1.2 Remove `CDNUrl` method and `publicURL` field from `r2Storage` in `internal/storage/r2.go`
- [x] 1.3 Remove `PublicURL` from `R2Config` in `internal/config/config.go` and remove `R2_PUBLIC_URL` loading from `loadFromEnv`
- [x] 1.4 Remove `R2_PUBLIC_URL` from `.env.example`

## 2. Usecase Layer

- [x] 2.1 Add `ImageItem{Image *domain.Image, ThumbnailURL *string}` type to `internal/usecase/image_usecase.go`
- [x] 2.2 Add private `thumbnailURL(ctx context.Context, path *string) *string` helper on `imageUsecase` — returns nil if path is nil, calls `GeneratePresignedGetURL` with `presignedGetTTL`, returns nil on error
- [x] 2.3 Update `ListImages` — call `thumbnailURL` for each image; update `ListImagesResult.Images` to `[]ImageItem`
- [x] 2.4 Update `ListTrashed` — call `thumbnailURL` for each image; update `ListTrashedResult.Images` to `[]ImageItem`
- [x] 2.5 Update `GetImage` — call `thumbnailURL`; add `ThumbnailURL *string` to `ImageDetail`
- [x] 2.6 Update `Restore` — call `thumbnailURL`; change return type from `*domain.Image` to `*ImageItem`; update `ImageUsecase` interface accordingly
- [x] 2.7 Update `UpdateImage` — call `thumbnailURL`; change return type from `*domain.Image` to `*ImageItem`; update `ImageUsecase` interface accordingly

## 3. Handler Layer

- [x] 3.1 Remove `store storage.StorageService` field from `ImageHandler` and its constructor in `internal/handler/image.go`
- [x] 3.2 Convert `toImageResponse` from a method on `ImageHandler` to a package-level function accepting `ImageItem`
- [x] 3.3 Update `ListImages` handler — map `[]ImageItem` using `toImageResponse`
- [x] 3.4 Update `ListTrashed` handler — map `[]ImageItem` using `toImageResponse`
- [x] 3.5 Update `GetImage` handler — use `ThumbnailURL` from `ImageDetail`
- [x] 3.6 Update `Restore` handler — use `*ImageItem` return
- [x] 3.7 Update `UpdateImage` handler — use `*ImageItem` return
- [x] 3.8 Update `NewImageHandler` call in `cmd/server/main.go` — remove `store` argument

## 4. Tests

- [x] 4.1 Update `internal/usecase/image_usecase_test.go` — update `ListImages`, `ListTrashed`, `GetImage`, `Restore`, `UpdateImage` tests to assert on `ImageItem` / `ImageDetail` return types; add mock expectation for `GeneratePresignedGetURL` on thumbnail paths; cover nil thumbnail path (no presign call) and presign failure (nil thumbnail URL returned)
- [x] 4.2 Update `internal/handler/image_test.go` — update mocked usecase return values to use `ImageItem` / updated `ImageDetail`; assert `thumbnail_url` in response JSON is presigned URL or null as appropriate
