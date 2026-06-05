## Why

After moving thumbnail generation and vision labelling into River async jobs, the frontend has no awareness of when those jobs complete — newly uploaded images show broken thumbnail links until the user reloads, and folder suggestions from vision are silently dropped and never reach the user.

## What Changes

- `ProcessVisionLabelling` always writes `ai_labels` (even `[]` when no labels found or vision is disabled), making `null` a reliable "job has not run yet" signal
- `GET /images/:id` response gains `suggested_folder_name *string`, computed from the top-scoring AI label when `ai_labels` is non-null and non-empty
- Image cards show a neutral placeholder when `thumbnail_url` is null, instead of a broken image link
- After upload, the frontend polls `GET /images/:id` every 2s for up to 30s; when `thumbnail_url` becomes non-null it patches the image in the query cache and the card flips to the real thumbnail
- After upload, the frontend schedules a single vision check at 2–3s; if `suggested_folder_name` is present a toast is shown with Accept / Ignore actions; if absent (and the user has vision enabled) a brief "Couldn't get folder suggestion" error toast is shown
- The inline folder suggestion view inside `UploadModal` is removed; the modal now closes immediately on successful upload
- Dead `suggested_folder_name` field and related `SuggestionState` type are removed from the frontend `CompleteUploadResult` type and `UploadModal`

## Capabilities

### New Capabilities

- `fe-async-job-feedback`: Post-upload thumbnail polling and vision suggestion check — hooks, cache patching, toast actions

### Modified Capabilities

- `fe-image-upload-flow`: Folder suggestion moves from an inline modal view to a toast notification; modal closes immediately on upload success
- `vision-api-labelling`: `ProcessVisionLabelling` must always write `ai_labels`; `GET /images/:id` exposes `suggested_folder_name`

## Impact

- `backend/internal/usecase/image_upload_usecase.go` — `ProcessVisionLabelling` always writes `ai_labels`
- `backend/internal/usecase/image_usecase.go` — `GetImage` result includes `suggested_folder_name`
- `backend/internal/handler/image.go` — `imageDetailResponse` gains `suggested_folder_name`
- `frontend/src/components/MasonryLayout.tsx` (or card component) — null thumbnail placeholder
- `frontend/src/components/UploadModal.tsx` — remove suggestion view, `SuggestionState`, `acceptMutation`; modal closes on success
- `frontend/src/lib/images.ts` — remove `suggested_folder_name` from `CompleteUploadResult`
- New `frontend/src/hooks/usePostUploadFeedback.ts` (or equivalent) — polling + vision check logic
