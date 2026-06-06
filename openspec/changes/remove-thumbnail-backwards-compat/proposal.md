## Why

With the browser extension now uploading thumbnails directly via presigned URL (same as the web app), the thumbnail worker fallback path is dead code. All upload paths are now client-side, making the async worker infrastructure, the gallery polling mechanism, and the HeadObject existence check unnecessary.

## What Changes

- **Remove** `refetchInterval` polling in `ImageGrid` — the gallery no longer needs to poll for `thumbnail_url === null`; all uploads now set the thumbnail synchronously
- **Remove** `HeadObject` check in `CompleteUpload` — thumbnail path is now set unconditionally (always present, always uploaded by client before `complete` is called)
- **Remove** conditional `ThumbnailUploadArgs` enqueue in `CompleteUpload`
- **Remove** `ThumbnailUploadWorker` and its registration
- **Remove** `ThumbnailUploadArgs` job type
- **Remove** `ProcessThumbnailUpload` usecase method
- **Remove** `ThumbnailService` interface and `imagingThumbnailService` implementation
- **Remove** thumbnail metrics (`thumbnailDuration`, `thumbnailCount`) and the `thumbnails` field from `imageUploadUsecase`
- **Remove** `HeadObject` from `StorageService` interface and `r2Storage` implementation
- **Remove** `disintegration/imaging` dependency (no longer used after thumbnail service is gone)
- **Deployment**: run `DELETE FROM river_jobs WHERE kind = 'thumbnail_upload'` to discard any orphaned jobs enqueued before this release

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `image-thumbnail`: removes ThumbnailService, Thumbnail Generation, Async Thumbnail Storage, and ProcessThumbnailUpload requirements; `CompleteUpload` now sets `thumbnail_path` unconditionally
- `fe-gallery-view`: removes the "Gallery self-polls while any image has a pending thumbnail" requirement

## Impact

- `frontend/src/components/ImageGrid.tsx`: remove `refetchInterval` option from `useInfiniteQuery`
- `backend/internal/usecase/image_upload_usecase.go`: simplify `CompleteUpload`, remove `ProcessThumbnailUpload`, remove `ThumbnailService` interface and struct fields, remove thumbnail metrics
- `backend/internal/usecase/image_storage.go`: remove `HeadObject` from `StorageService` interface
- `backend/internal/storage/r2.go`: remove `HeadObject` method
- `backend/internal/usecase/job_args.go`: remove `ThumbnailUploadArgs`
- `backend/internal/worker/thumbnail.go`: delete file
- `backend/pkg/thumbnail/thumbnail.go`: delete file
- `backend/cmd/server/main.go`: remove `thumbnailService` instantiation and `ThumbnailUploadWorker` registration
- `backend/internal/usecase/image_upload_usecase_test.go`: remove `mockThumbnailService`, update all `NewImageUploadUsecase` calls, delete three `ProcessThumbnailUpload` tests and two `CompleteUpload` scenario tests, add one replacement test asserting thumbnail path is always set
- `backend/internal/usecase/image_usecase_test.go`: remove `HeadObject` from `mockStorageService`
- `go.mod` / `go.sum`: remove `disintegration/imaging`
