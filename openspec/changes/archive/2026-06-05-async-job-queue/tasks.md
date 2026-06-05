## 1. Dependencies and Migration

- [x] 1.1 Add `riverqueue/river`, `riverqueue/river/riverdriver/riverpgxv5`, and `jackc/pgx/v5` (promote from indirect) to `go.mod` via `go get`
- [x] 1.2 Create `backend/migration/000012_river_schema.up.sql` with River's published schema SQL
- [x] 1.3 Create `backend/migration/000012_river_schema.down.sql` that drops all River tables

## 2. Worker Package and Job Args

- [x] 2.1 Create `internal/worker/` package with `thumbnail.go` defining `ThumbnailUploadArgs` (with `Kind()`) and `ThumbnailUploadWorker` struct
- [x] 2.2 Create `internal/worker/vision.go` defining `VisionArgs` (with `Kind()`) and `VisionWorker` struct with `NextAttemptScheduledAt` returning a fixed 10s delay
- [x] 2.3 Create `internal/worker/periodic.go` defining `CleanupStaleUploadsWorker` and `TrashPurgeWorker` for use as periodic job handlers

## 3. Usecase Changes

- [x] 3.1 Add `ProcessThumbnailUpload(ctx context.Context, imageID uuid.UUID, r2Path, thumbnailKey string) error` to `imageUploadUsecase` (replaces the `uploadThumbnail` goroutine body)
- [x] 3.2 Add `ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error` to `imageUploadUsecase` (replaces the `runVisionFlow` body; returns nil early if vision disabled)
- [x] 3.3 Remove the `uploadThumbnail` goroutine method and `runVisionFlow` method from `imageUploadUsecase`
- [x] 3.4 Update `CompleteUpload` to insert `ThumbnailUploadArgs` and `VisionArgs` River jobs after the DB transaction, replacing `go u.uploadThumbnail(...)` and `u.runVisionFlow(...)`
- [x] 3.5 Simplify `CompleteUploadResult` to `{ ImageID uuid.UUID }` — remove `SuggestedFolderName` and `Warning` fields

## 4. Handler Changes

- [x] 4.1 Update `UploadHandler.CompleteUpload` response to return only `image_id`; remove `suggested_folder_name` and `warning` from `completeUploadResponse`

## 5. Worker Implementations

- [x] 5.1 Implement `ThumbnailUploadWorker.Work` to call `uploadUsecase.ProcessThumbnailUpload` with max 5 attempts
- [x] 5.2 Implement `VisionWorker.Work` to call `uploadUsecase.ProcessVisionLabelling` with max 3 attempts and fixed 10s retry delay
- [x] 5.3 Implement periodic worker `Work` methods that call `uploadUsecase.CleanupStaleUploads` and `imageUsecase.PurgeExpiredTrash` respectively

## 6. Wire Up in main.go

- [x] 6.1 Open a dedicated `*pgxpool.Pool` from `cfg.DB.URL` (max 3 connections) for River
- [x] 6.2 Initialise the River client with all workers registered (`ThumbnailUploadWorker`, `VisionWorker`, `CleanupStaleUploadsWorker`, `TrashPurgeWorker`) and both periodic jobs configured
- [x] 6.3 Store the River client on the `server` struct; add `riverClient.Stop(ctx)` to `server.shutdown()`
- [x] 6.4 Remove `compositeImageWorker`, `imageWorkerUsecase` interface, and `startWorkers` from `main.go`
- [x] 6.5 Remove the `imageWorker` field from the `server` struct

## 7. Tests

- [x] 7.1 Update `imageUploadUsecase` unit tests: remove scenarios for `runVisionFlow` and `uploadThumbnail`; add scenarios for `ProcessThumbnailUpload` (success, any-step failure) and `ProcessVisionLabelling` (success, vision disabled early return, vision API error)
- [x] 7.2 Update `UploadHandler` unit tests: update `CompleteUpload` handler test to assert response contains only `image_id` and that the usecase is called correctly (no `SuggestedFolderName` assertion)

## 8. Bruno File

- [x] 8.1 Update the existing `complete-upload` Bruno request file to reflect the simplified response shape (remove `suggested_folder_name` and `warning` from the example response)
