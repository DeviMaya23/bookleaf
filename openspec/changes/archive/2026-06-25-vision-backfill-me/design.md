## Context

Vision labelling has exactly one entry point today: `CompleteUpload` enqueues a `VisionArgs{ImageID, UserID}` River job, processed by `VisionWorker` → `imageUploadUsecase.ProcessVisionLabelling`. That method already gates on `user.VisionEnabled` and is safe to call multiple times for the same image (it just re-fetches, re-annotates, and overwrites `ai_labels`).

There is a directly analogous precedent: `BackfillPhash` lists images via `ImageHashRepository.ListUnhashed` and processes them in a bounded batch, run periodically. This change reuses that *shape* (list-unprocessed → process) but not the periodic-job *mechanism* — see Decisions.

## Goals / Non-Goals

**Goals:**
- Let the authenticated caller re-trigger vision labelling for every one of their own images that has never been labelled (`ai_labels IS NULL`).
- Reuse the existing `VisionArgs` job and `VisionWorker` untouched.
- Keep the change backend-only, with no new architectural surface (no new job kind, no new auth pattern, no new route group).

**Non-Goals:**
- No periodic/automatic backfill — this is operator-triggered only, never fired by a timer.
- No user-facing UI trigger.
- No protection against duplicate enqueueing if the endpoint is called more than once before prior jobs finish (acceptable per proposal — `ProcessVisionLabelling` is safe to re-run).
- No pagination/batching of the unlabelled-image list — enqueueing is a cheap DB insert per image, so the whole set is processed in one call.

## Decisions

**1. New endpoint lives on `UploadHandler`, not `MeHandler`, despite the `/me` path prefix.**
`UploadHandler` already wraps `UploadUsecase`, backed by the same `imageUploadUsecase` instance that will own `BackfillVisionLabels`. `MeHandler` wraps a separate `UserUsecase`/`AccountUsecase` pair with no relationship to image data. Adding the method to `UploadUsecase` and routing `POST /me/vision/backfill` to `uploadHandler.BackfillVision` keeps the handler/usecase pairing 1:1, same as `AcceptSuggestion` (which is already registered under `/images/...` on `uploadHandler` despite living conceptually elsewhere). Echo's router doesn't require path prefix to match handler struct name.
- Alternative considered: add a narrow `VisionBackfillUsecase` interface to `MeHandler` instead. Rejected — it would require threading the `imageUploadUsecase` instance into `MeHandler`'s constructor purely for one method, duplicating a dependency `UploadHandler` already has.

**2. `ListUnlabelled` mirrors `ListUnhashed` exactly.**
```go
func (r *imageRepository) ListUnlabelled(ctx context.Context, userID string) ([]*domain.Image, error) {
    var images []*domain.Image
    err := r.db.WithContext(ctx).
        Unscoped().
        Where("user_id = ? AND ai_labels IS NULL AND deleted_at IS NULL", userID).
        Find(&images).Error
    ...
}
```
`Unscoped()` is required because GORM's default scope already excludes soft-deleted rows via `deleted_at`, but the existing codebase convention (`ListUnhashed`) is to `Unscoped()` + filter `deleted_at IS NULL` explicitly. Matching it keeps the two backfill-style queries consistent rather than introducing a second style.

**3. `BackfillVisionLabels` enqueues sequentially and fails fast.**
```go
func (u *imageUploadUsecase) BackfillVisionLabels(ctx context.Context, userID string) (int, error) {
    images, err := u.imageRepo.ListUnlabelled(ctx, userID)
    if err != nil { return 0, fmt.Errorf("list unlabelled images: %w", err) }

    for i, img := range images {
        if err := u.enqueuer.Insert(ctx, VisionArgs{ImageID: img.ID, UserID: userID}); err != nil {
            return i, fmt.Errorf("enqueue vision labelling: %w", err)
        }
    }
    return len(images), nil
}
```
No transaction wraps the loop — each `Insert` is an independent River job insert, same as the single-image call in `CompleteUpload`. If enqueueing fails partway, images already enqueued stay enqueued (harmless); the caller gets an error and can retry the whole endpoint, which will simply skip images that still have `ai_labels IS NULL`... wait — already-enqueued-but-not-yet-processed images still have `ai_labels IS NULL`, so a retry would re-enqueue them too. This is accepted under "no double-enqueue guard."

**4. Response: `202 Accepted` with the enqueued count.**
```json
{ "enqueued": 7 }
```
202 rather than 200/204 signals that labelling itself happens asynchronously, not within this request. The count gives the caller (you, manually testing) immediate feedback without needing to inspect the job queue.

## Risks / Trade-offs

- **[Risk] A user with a very large unlabelled backlog triggers many job inserts in one request.** → Accepted per proposal — no batching. If this becomes a real problem, batching can be added later without changing the endpoint's external contract.
- **[Risk] Partial enqueue on mid-loop DB error leaves some images enqueued, some not.** → Mitigation: none needed — re-calling the endpoint is safe and will simply enqueue the remainder (and harmlessly re-enqueue any still-pending ones, per Non-Goals).
- **[Risk] Vision API quota/cost spike if a user has hundreds of unlabelled images.** → Out of scope for this change; `ProcessVisionLabelling`'s existing per-job behavior (including the 3-attempt River retry policy) is unchanged.
