# Tasks

## 1. Interface and mock alignment

- [x] 1.1 In `backend/internal/usecase/image_storage.go`: add `GenerateDeterministicPresignedGetURL(ctx context.Context, key string) (string, error)` to `StorageService`; update `PutObject` signature to `PutObject(ctx context.Context, key string, body io.Reader, contentType, cacheControl string) error` — verify `go build ./...` passes with no compile errors
- [x] 1.2 In `backend/internal/usecase/image_usecase_test.go`: update `mockStorageService.PutObject` signature to match the new interface; add a `GenerateDeterministicPresignedGetURL` stub that returns `m.getURL, m.err` — verify `go test ./internal/usecase/...` compiles and existing tests still pass

## 2. Wire thumbnail call sites

- [x] 2.1 In `backend/internal/usecase/image_usecase.go`: switch `thumbnailURL()` to call `u.store.GenerateDeterministicPresignedGetURL(ctx, *path)` (no TTL argument); remove `presignedGetTTL` from that call — verify `TestImageUsecase_GetImageDetail` and thumbnail-related tests still pass
- [x] 2.2 In `backend/internal/usecase/folder_images_usecase.go`: switch the inline `GeneratePresignedGetURL` call to `GenerateDeterministicPresignedGetURL`; update `stubFolderImagesStorage` in `folder_images_usecase_test.go` to override `GenerateDeterministicPresignedGetURL` (returning the same stub URL) and remove the `GeneratePresignedGetURL` override — verify `go test ./internal/usecase/` `-run TestFolderImages` passes
- [x] 2.3 In `backend/internal/usecase/trash_usecase.go`: switch `thumbnailURL()` to `GenerateDeterministicPresignedGetURL` — verify `go test ./internal/usecase/` `-run TestTrash` passes
- [x] 2.4 In `backend/internal/usecase/share_usecase.go`: switch `thumbnailURL()` to `GenerateDeterministicPresignedGetURL`; update `spyShareStorage` in `share_usecase_test.go` to override `GenerateDeterministicPresignedGetURL` instead of `GeneratePresignedGetURL` — verify `go test ./internal/usecase/` `-run TestShare` passes

## 3. Storage unit tests

- [x] 3.1 In `backend/internal/storage/r2_test.go` (create if absent): add unit tests for `GenerateDeterministicPresignedGetURL` using an injected `now` clock — cover: (a) two calls within the same 24 h window return identical URLs, (b) calls at window boundary (just before / just after midnight UTC) return different URLs, (c) expiry embedded in the URL is ≥ 48 h — verify `go test ./internal/storage/...` passes

## 4. Final verification

- [x] 4.1 Run `golangci-lint run ./backend/...` and fix any reported issues
- [x] 4.2 Run `go build ./backend/...` and confirm zero compile errors

## Workflow follow-up

- Archive the change after review.
