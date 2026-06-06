## 1. Frontend

- [x] 1.1 Remove the `refetchInterval` option from `useInfiniteQuery` in `ImageGrid.tsx`

## 2. Backend — usecase

- [x] 2.1 Remove `ThumbnailService` interface and `thumbnails` field from `imageUploadUsecase`; remove `thumbnailDuration` and `thumbnailCount` metrics and their initialization
- [x] 2.2 Remove `ThumbnailService` parameter from `NewImageUploadUsecase` constructor
- [x] 2.3 Simplify `CompleteUpload`: remove `HeadObject` call and conditional worker enqueue; set `img.ThumbnailPath = &thumbnailKey` unconditionally
- [x] 2.4 Remove `ProcessThumbnailUpload` method from `imageUploadUsecase`
- [x] 2.5 Remove `ThumbnailUploadArgs` from `job_args.go`

## 3. Backend — storage

- [x] 3.1 Remove `HeadObject` from the `StorageService` interface in `image_storage.go`
- [x] 3.2 Remove `HeadObject` method from `r2Storage` in `r2.go`

## 4. Backend — worker and service

- [x] 4.1 Delete `backend/internal/worker/thumbnail.go`
- [x] 4.2 Delete `backend/pkg/thumbnail/thumbnail.go`
- [x] 4.3 Remove `thumbnailService` instantiation and `ThumbnailUploadWorker` registration from `cmd/server/main.go`; remove the `thumbnail` package import

## 5. Backend — tests

- [x] 5.1 Remove `mockThumbnailService` struct and `Generate` method from `image_upload_usecase_test.go`
- [x] 5.2 Remove `HeadObject` method, `headObjectFound`, `headCalls`, and `headObjectErr` fields from `mockStorageService` in `image_usecase_test.go`
- [x] 5.3 Update `newImageUploadUsecase` helper in `image_upload_usecase_test.go` to drop the `thumbnails ThumbnailService` parameter; update all call sites
- [x] 5.4 Delete `TestImageUploadUsecase_CompleteUpload_ThumbnailPresent_SetsThumbnailPathAndSkipsWorker` and `TestImageUploadUsecase_CompleteUpload_ThumbnailAbsent_LeavesThumbnailPathNilAndEnqueuesWorker`
- [x] 5.5 Delete `TestImageUploadUsecase_ProcessThumbnailUpload_Success`, `TestImageUploadUsecase_ProcessThumbnailUpload_FetchFails`, and `TestImageUploadUsecase_ProcessThumbnailUpload_UploadFails`
- [x] 5.6 Add `TestImageUploadUsecase_CompleteUpload_AlwaysSetsThumbnailPath`: assert `ThumbnailPath` is non-nil and correctly formatted, and `enqueuer.insertCalls == 1` (vision only)
- [x] 5.7 Update existing `CompleteUpload` tests that used `headObjectFound` or checked `enqueuer.insertCalls` for two jobs

## 6. Dependency cleanup

- [x] 6.1 Run `go mod tidy` to remove `disintegration/imaging` from `go.mod` and `go.sum`
- [x] 6.2 Verify `go build ./...` and `go test ./...` pass cleanly
