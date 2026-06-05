## Why

After upload, images with pending thumbnails show a placeholder indefinitely — they only update after a manual page reload or navigation. The current thumbnail polling in `usePostUploadFeedback` is upload-path-specific and doesn't cover batch uploads. Additionally, `usePostUploadFeedback` couples thumbnail polling and vision suggestion under a single state-mediated trigger, creating unnecessary complexity that makes both concerns harder to reason about independently.

## What Changes

- `ImageGrid` gains a `refetchInterval` on its `useInfiniteQuery` that polls every 2s while any loaded image has `thumbnail_url === null`, and stops automatically once all thumbnails resolve — covering all upload paths without upload-specific wiring
- `usePostUploadFeedback` is deleted entirely
- A new `useVisionSuggestion` hook replaces it, exposing a `checkVision(imageId)` function that callers invoke directly at upload success — no intermediary state required; internally retries at 1s then 3s total before showing an error toast
- `AppLayout` drops `pendingFeedbackImageId` state and the `GET /me` query; vision is triggered directly from `handleMainDrop` and `UploadModal.onUploadSuccess`
- Vision suggestions remain scoped to single-file uploads (modal and DnD); batch upload is explicitly excluded

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `fe-gallery-view`: adds a requirement — the gallery SHALL self-poll while any loaded image has a null thumbnail
- `fe-async-job-feedback`: removes the thumbnail polling requirement (ownership moves to the gallery); updates the vision suggestion check to a two-attempt retry scoped to single-file uploads only

## Impact

- `frontend/src/components/ImageGrid.tsx` — `useInfiniteQuery` gains `refetchInterval`
- `frontend/src/hooks/usePostUploadFeedback.ts` — deleted
- `frontend/src/hooks/useVisionSuggestion.ts` — new file
- `frontend/src/components/AppLayout.tsx` — `pendingFeedbackImageId` state removed; `getMe` / `['me']` query removed; `usePostUploadFeedback` replaced by `useVisionSuggestion`; `checkVision` wired to DnD and modal upload success
- No backend changes, no API changes, no new dependencies
