## Context

`usePostUploadFeedback` currently handles two concerns: polling `GET /images/:id` every 2s until `thumbnail_url` is non-null, and checking `suggested_folder_name` once after a fixed delay. Both are driven by a single `pendingFeedbackImageId` state in `AppLayout`, which is set only from single-file upload paths (DnD and modal). Batch uploads never set this state, so batch-uploaded images are stuck on placeholder until manual reload.

## Goals / Non-Goals

**Goals:**
- Thumbnail refresh works for all upload paths (single and batch) without upload-specific wiring
- Vision suggestion check is a self-contained hook with a direct call-site trigger and predictable retry behavior
- `AppLayout` holds no upload-feedback state

**Non-Goals:**
- Vision suggestions for batch uploads (explicitly excluded)
- Backend changes of any kind

## Decisions

### Gallery owns thumbnail refresh via refetchInterval

`useInfiniteQuery` accepts a `refetchInterval` callback that receives the current query state. The computed interval is `2000` when any image in the flat-mapped pages has `thumbnail_url === null`, and `false` otherwise. React Query disables polling automatically when `false` is returned.

This was chosen over continuing to poll from a hook because the gallery already holds the full image list and is the natural owner of "are there pending thumbnails right now." Any upload path that invalidates `['images']` — including batch — gets coverage for free.

### useVisionSuggestion exposes a checkVision function instead of watching state

The previous pattern (hook watches `imageId` state via `useEffect`) existed because both thumbnail polling and vision check shared the same trigger. With thumbnail polling removed, the indirection through state has no benefit. A function exposed from the hook lets callers (`handleMainDrop`, `UploadModal.onUploadSuccess`) invoke the check directly at the point where the image ID is already in hand.

Timer IDs are stored in refs so the cleanup function returned from `checkVision` (or a `useEffect` teardown) can cancel pending timers on unmount.

### Two-attempt retry: 1s then 3s total

A single fixed delay is a timing guess. Two attempts give the vision job a reasonable window while keeping total wait time under 3s — beyond which a folder suggestion toast feels disconnected from the upload action. If both checks return null, an error toast is shown. This is the same failure behavior as the current implementation.

### useVisionSuggestion fetches /me internally

`AppLayout` currently holds the `GET /me` query solely to pass `vision_enabled` to `usePostUploadFeedback`. Moving the fetch into `useVisionSuggestion` makes the hook fully self-contained and removes the last reason for AppLayout to know about the me endpoint. React Query's cache means the request is shared if anything else queries `['me']` in the future.

## Risks / Trade-offs

- **refetchInterval fires on all image list queries** — the interval runs whenever the query is active, not only post-upload. If a user has many images with null thumbnails for other reasons (e.g., a failed job), polling continues indefinitely. Mitigation: this matches the intent — poll until resolved — and the backend thumbnail job is expected to complete eventually.
- **checkVision timers survive navigation if AppLayout stays mounted** — if the user navigates between folders while a vision check is in flight, the toast still fires. This is acceptable and matches current behavior.
