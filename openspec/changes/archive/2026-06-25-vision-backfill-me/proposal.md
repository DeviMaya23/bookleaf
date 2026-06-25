## Why

Vision labelling currently only runs once, at upload time (`CompleteUpload` enqueues a `VisionArgs` job). If a user uploads an image while `vision_enabled` is `false`, or vision labelling otherwise never ran for an image, that image's `ai_labels` stays `null` forever — there is no existing path to retroactively label it. We need a manually-triggered way to catch these images up, without building any automatic/background sweep.

## What Changes

- Add `POST /me/vision/backfill`: for the authenticated caller, finds all of their images with `ai_labels IS NULL` (excluding deleted images) and enqueues the existing `VisionArgs` River job for each.
- Add `ImageRepository.ListUnlabelled(ctx, userID) ([]*domain.Image, error)` — mirrors the existing `ListUnhashed` pattern (`Unscoped().Where("ai_labels IS NULL AND deleted_at IS NULL")`), scoped additionally by `user_id`.
- Add `imageUploadUsecase.BackfillVisionLabels(ctx, userID string) error` — lists unlabelled images via the new repo method and enqueues `VisionArgs{ImageID, UserID}` for each via the existing `JobEnqueuer`.
- No new River job type, no new worker, no batching/limit (this is a one-shot, on-demand call, not a periodic job — enqueueing is cheap, so all unlabelled images are enqueued in a single call).
- No new auth pattern — the endpoint is scoped to the calling user's own ID (from the existing JWT-derived `userID`), same as every other `/me` route.
- Double-enqueueing (e.g. calling the endpoint twice before prior jobs finish) is explicitly not guarded against — `ProcessVisionLabelling` is idempotent enough that a duplicate run just re-labels the image.

## Capabilities

### New Capabilities
- `vision-backfill-me`: `POST /me/vision/backfill` endpoint, the `ListUnlabelled` repository method, and the `BackfillVisionLabels` usecase method that together let the authenticated user re-trigger vision labelling for all of their currently-unlabelled images.

### Modified Capabilities
(none — existing `vision-api-labelling` and `me-endpoint` requirements are unchanged; this change only adds a new entry point that reuses the existing `VisionArgs` job and worker as-is)

## Impact

- **Backend**: new handler route under the existing protected `/me` group (`backend/internal/handler/`), new usecase method on `imageUploadUsecase` (`backend/internal/usecase/image_upload_usecase.go`), new repository method on `ImageRepository` (`backend/internal/repository/image_repository.go` + interface).
- **No frontend or extension changes** — no user-facing trigger, per requirements.
- **No new dependencies, no new job/worker types, no new auth/middleware.**
