## 1. Repository

- [x] 1.1 Add `ListUnlabelled(ctx context.Context, userID string) ([]*domain.Image, error)` to the `ImageRepository` interface (wherever `ListUnhashed`/`UpdateAILabels` are declared) and implement it in `image_repository.go`, mirroring `ListUnhashed`'s `Unscoped()` + explicit `deleted_at IS NULL` pattern, scoped additionally by `user_id`.
- [x] 1.2 Add an integration test for `ListUnlabelled` covering: returns only the calling user's unlabelled, non-deleted images; excludes another user's unlabelled images; excludes soft-deleted images; returns an empty slice when none match.

## 2. Usecase

- [x] 2.1 Add `BackfillVisionLabels(ctx context.Context, userID string) (int, error)` to `imageUploadUsecase` in `image_upload_usecase.go`: list unlabelled images via the new repo method, enqueue `VisionArgs{ImageID, UserID}` per image in order via the existing `JobEnqueuer`, stop and return the partial count + error on the first enqueue failure, otherwise return the full count.
- [x] 2.2 Add unit tests for `BackfillVisionLabels` per CONVENTIONS.md rules (usecase layer, fakes/mocks for `ImageRepository`/`JobEnqueuer`): all images enqueued successfully; zero unlabelled images; `ListUnlabelled` error is propagated as `(0, err)`; enqueue failure on a middle image returns the correct partial count and the specific wrapped error.

## 3. Handler

- [x] 3.1 Add `BackfillVisionLabels(ctx context.Context, userID string) (int, error)` to the `UploadUsecase` interface in `image_upload.go`.
- [x] 3.2 Add `BackfillVision(c echo.Context) error` to `UploadHandler`: resolve `userID` from `middleware.AuthenticatedUserIDFromContext`, call `uploadUsecase.BackfillVisionLabels`, return `202 Accepted` with `{ "enqueued": <count> }` on success or `500 Internal Server Error` on usecase error.
- [x] 3.3 Register `protected.POST("/me/vision/backfill", uploadHandler.BackfillVision)` in `main.go`.
- [x] 3.4 Add unit tests for `BackfillVision` per CONVENTIONS.md rules (handler layer, fake `UploadUsecase`): authenticated success returns `202` with correct count; usecase error returns `500`; missing authenticated user ID in context returns the existing error path used by other `/me`-style handlers.

## 4. API Client (Bruno)

- [x] 4.1 Add `bruno/backfill-vision.bru` — `POST {{baseUrl}}/me/vision/backfill`, `auth: inherit`, no request body, following the existing `me.bru`/`update-me.bru` format and `seq` numbering.

## 5. Verification

- [x] 5.1 Run `golangci-lint run` from the backend module and fix any issues introduced by this change.
