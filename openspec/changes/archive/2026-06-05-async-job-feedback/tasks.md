## 1. Backend: ProcessVisionLabelling always writes ai_labels

- [x] 1.1 Update `ProcessVisionLabelling` in `image_upload_usecase.go`: remove the early-return when Vision returns zero labels; always call `UpdateAILabels` with the marshalled labels (empty `[]` if zero labels returned)
- [x] 1.2 Update the existing unit test for `ProcessVisionLabelling`: add a scenario asserting that when Vision returns zero labels, `UpdateAILabels` is called with an empty JSON array and nil is returned

## 2. Backend: suggested_folder_name on GET /images/:id

- [x] 2.1 Add `SuggestedFolderName *string` to `usecase.ImageDetail` in `image_usecase.go`
- [x] 2.2 Populate `SuggestedFolderName` in the `GetImage` usecase: nil if `AILabels` is null or an empty array; the `Description` of the first label otherwise
- [x] 2.3 Add `SuggestedFolderName *string json:"suggested_folder_name"` to `imageDetailResponse` in `handler/image.go` and map it from `result.SuggestedFolderName`
- [x] 2.4 Add unit tests for `GetImage` usecase covering: `AILabels` null → `SuggestedFolderName` nil; `AILabels` empty array → nil; `AILabels` with labels → top label `Description`
- [x] 2.5 Update `GetImage` handler unit test happy path to assert `suggested_folder_name` is present in the response body
- [x] 2.6 Update `bruno/images/get-image.bru` to include `suggested_folder_name` in the example response

## 3. Frontend: Remove dead suggestion view from UploadModal

- [x] 3.1 Remove `SuggestionState` interface, `suggestion` state, `setSuggestion`, `acceptMutation`, and `handleIgnore` from `UploadModal.tsx`
- [x] 3.2 Remove the suggestion JSX branch (the `{suggestion ? (...) : (...)}` conditional) — keep only the upload form branch
- [x] 3.3 Update `uploadMutation.onSuccess` to always show a success toast and close the modal (remove the `result.suggested_folder_name` check)
- [x] 3.4 Remove `suggested_folder_name` from `CompleteUploadResult` in `frontend/src/lib/images.ts`
- [x] 3.5 Update `UploadModal` tests: remove any suggestion view test cases; add a scenario asserting the modal closes and a success toast is shown on successful upload

## 4. Frontend: usePostUploadFeedback hook

- [x] 4.1 Create `frontend/src/hooks/usePostUploadFeedback.ts`; the hook accepts `imageId: string | null` and `visionEnabled: boolean` and internally uses `useKindeAuth` and `useQueryClient`
- [x] 4.2 Implement thumbnail polling: when `imageId` is non-null, poll `GET /images/:id` every 2s; on `thumbnail_url` non-null, patch the image in the matching React Query list cache entries and stop; stop silently after 30s
- [x] 4.3 Implement vision check: when `imageId` is non-null and `visionEnabled` is true, call `GET /images/:id` once after a 2–3s delay; show a Sonner suggestion toast with Accept / Ignore actions if `suggested_folder_name` is present, otherwise show a "Couldn't get folder suggestion" error toast
- [x] 4.4 Accept action on the suggestion toast calls `acceptSuggestion` and invalidates `['images']` and `['folders']` query caches on success
- [x] 4.5 Wire `usePostUploadFeedback` into `UploadModal`: pass `result.image_id` and the user's `vision_enabled` (from the `/me` query) to trigger the hook after `uploadMutation` succeeds
- [x] 4.6 Wire `usePostUploadFeedback` trigger into `AppLayout.handleMainDrop`: call `setPendingFeedbackImageId(imageDetail.id)` after a successful drag-and-drop upload so the hook fires for the DnD path as well as the modal path
