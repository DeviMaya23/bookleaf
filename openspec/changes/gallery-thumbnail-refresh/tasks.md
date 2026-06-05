## 1. Gallery Thumbnail Self-Poll

- [x] 1.1 Add `refetchInterval` to `useInfiniteQuery` in `ImageGrid.tsx` — return `2000` when any flat-mapped image has `thumbnail_url === null`, `false` otherwise

## 2. useVisionSuggestion Hook

- [x] 2.1 Create `frontend/src/hooks/useVisionSuggestion.ts` — hook fetches `GET /me` internally via `useQuery` with `staleTime: Infinity`; returns `{ checkVision: (imageId: string) => void }`
- [x] 2.2 Implement retry logic in `checkVision`: schedule first check at 1s; if `suggested_folder_name` is null, schedule second check 2s later (3s total from call); cancel pending timers on unmount
- [x] 2.3 Implement suggestion toast: message `Suggested folder: "<name>"` with Accept action (`POST /images/:id/accept-suggestion`, invalidate `['images']` and `['folders']`) and Ignore action (dismiss, no API call)
- [x] 2.4 Implement error toast `"Couldn't get folder suggestion"` on null after both attempts or on any thrown error
- [x] 2.5 Guard `checkVision` — return immediately without scheduling if `vision_enabled` is false

## 3. AppLayout Wiring

- [x] 3.1 Remove `pendingFeedbackImageId` state and its setter
- [x] 3.2 Remove `getMe` import, the `['me']` `useQuery` call, and the `me` data reference
- [x] 3.3 Remove `usePostUploadFeedback` import and call
- [x] 3.4 Add `useVisionSuggestion`; call `checkVision(imageDetail.id)` in `handleMainDrop` after successful single-file upload
- [x] 3.5 Pass `onUploadSuccess={checkVision}` to `UploadModal`

## 4. Cleanup

- [x] 4.1 Delete `frontend/src/hooks/usePostUploadFeedback.ts`
- [x] 4.2 Verify no remaining imports reference `usePostUploadFeedback`
