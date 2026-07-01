## Requirements

### Requirement: ListUnlabelled Repository Method

The system SHALL add `ListUnlabelled(ctx context.Context, userID string) ([]*domain.Image, error)` to `ImageRepository`. The query SHALL select images belonging to `userID` where `ai_labels IS NULL` and `deleted_at IS NULL`, bypassing GORM's default soft-delete scope (`Unscoped()`) and filtering `deleted_at` explicitly, consistent with the existing `ListUnhashed` method.

#### Scenario: Returns only the user's unlabelled, non-deleted images

- **WHEN** `ListUnlabelled` is called for a user with three images: one with `ai_labels` populated, one with `ai_labels IS NULL`, and one with `ai_labels IS NULL` but soft-deleted
- **THEN** only the single non-deleted image with `ai_labels IS NULL` is returned

#### Scenario: Does not return another user's unlabelled images

- **WHEN** `ListUnlabelled` is called for user A, and user B has images with `ai_labels IS NULL`
- **THEN** user B's images are not included in the result

#### Scenario: No unlabelled images returns an empty slice

- **WHEN** `ListUnlabelled` is called for a user whose images all have non-null `ai_labels`
- **THEN** an empty slice and a nil error are returned

### Requirement: BackfillVisionLabels Usecase Method

The system SHALL add `BackfillVisionLabels(ctx context.Context, userID string) (int, error)` to `imageUploadUsecase`. The method SHALL:

1. Fetch the user's unlabelled images via `ImageRepository.ListUnlabelled(ctx, userID)`.
2. For each returned image, enqueue a `VisionArgs{ImageID, UserID}` River job via the existing `JobEnqueuer.Insert`, in list order.
3. Return the number of jobs successfully enqueued.

If `ListUnlabelled` returns an error, the method SHALL return `(0, err)`. If `JobEnqueuer.Insert` fails for one of the images, the method SHALL stop enqueueing further images and return the count enqueued so far along with a non-nil error. No transaction wraps the enqueue loop — jobs already inserted before a failure remain enqueued.

This method does not filter by `vision_enabled` itself; it relies on `ProcessVisionLabelling`'s existing early-return behavior when `vision_enabled` is false.

#### Scenario: All unlabelled images enqueued successfully

- **WHEN** `BackfillVisionLabels` is called for a user with 3 unlabelled images
- **THEN** a `VisionArgs` job is enqueued for each of the 3 images
- **AND** `(3, nil)` is returned

#### Scenario: No unlabelled images

- **WHEN** `BackfillVisionLabels` is called for a user with zero unlabelled images
- **THEN** no jobs are enqueued
- **AND** `(0, nil)` is returned

#### Scenario: Enqueue failure stops further enqueueing

- **WHEN** `BackfillVisionLabels` is called for a user with 3 unlabelled images and the job enqueuer fails on the 2nd image
- **THEN** exactly 1 job has been enqueued (for the 1st image)
- **AND** a non-nil error is returned along with a count of `1`

### Requirement: POST /me/vision/backfill Endpoint

The system SHALL expose a `POST /me/vision/backfill` endpoint in the protected route group, requiring a valid JWT. The endpoint SHALL call `BackfillVisionLabels` with the authenticated caller's user ID (from the JWT-derived context, identical to every other `/me` route) — it SHALL NOT accept a user ID from the request path, query, or body.

On success, the response SHALL be `202 Accepted` with body:
```json
{ "enqueued": <number> }
```

The endpoint SHALL NOT filter or validate based on the caller's `vision_enabled` setting — images are still enqueued, and `ProcessVisionLabelling` decides at job-processing time whether to call the Vision API.

#### Scenario: Authenticated user backfills unlabelled images

- **WHEN** an authenticated request is made to `POST /me/vision/backfill` and the caller has 5 images with `ai_labels IS NULL`
- **THEN** the response is `202 Accepted`
- **AND** the response body is `{ "enqueued": 5 }`
- **AND** a `VisionArgs` River job has been inserted for each of the 5 images

#### Scenario: Authenticated user with no unlabelled images

- **WHEN** an authenticated request is made to `POST /me/vision/backfill` and the caller has zero images with `ai_labels IS NULL`
- **THEN** the response is `202 Accepted`
- **AND** the response body is `{ "enqueued": 0 }`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `POST /me/vision/backfill` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no jobs are enqueued

#### Scenario: Usecase failure returns server error

- **WHEN** an authenticated request is made to `POST /me/vision/backfill` and `BackfillVisionLabels` returns a non-nil error
- **THEN** the response is `500 Internal Server Error`
