## Why

When thumbnail generation fails during `CompleteUpload`, the backend currently returns `200 OK` with a `warning` field — but the frontend ignores the warning and shows a success checkmark. The image record stays with `is_uploaded = false` and is silently removed by the stale cleanup job, leaving the user with a phantom upload confirmation and a missing image.

## What Changes

- `CompleteUpload` SHALL return a real error (non-2xx) when `prepareThumbnail` fails, instead of setting `Warning` and returning `nil`
- The `Warning` field on `CompleteUploadResult` is retained for non-thumbnail failures (e.g. vision labelling) but is no longer used for thumbnail failures
- No frontend changes: both upload flows already handle HTTP errors via existing `catch`/`onError` paths

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `image-thumbnail`: thumbnail failure behavior changes — `CompleteUpload` returns an error instead of `200 OK` with a warning
- `image-endpoints`: `POST /images/:id/complete` response changes — thumbnail failure yields a non-2xx response; `warning` field is no longer used for thumbnail failures

## Impact

- `backend/internal/usecase/image_usecase.go` — `CompleteUpload` method
- `backend/internal/usecase/image_usecase_test.go` — update test scenarios for thumbnail failure
- `openspec/specs/image-thumbnail/spec.md` — delta spec required
- `openspec/specs/image-endpoints/spec.md` — delta spec required
