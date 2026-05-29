## Context

The existing upload flow is a single-file modal (`UploadModal.tsx`) backed by a three-step API sequence: `POST /images` → `PUT` to R2 presigned URL → `POST /images/:id/complete`. The backend is not changing. The batch flow reuses those same three API functions for every file in the batch, orchestrated entirely on the frontend.

Current entry point: a `+ Image` button in `AppLayout` that opens `UploadModal` directly. Current drag-and-drop: drops the first file only.

## Goals / Non-Goals

**Goals:**
- Allow uploading up to 20 files in a single action via a new `BatchUploadModal`
- Run up to 3 uploads concurrently; queue the rest
- Validate file count (≤ 20) and file size (≤ 50 MB) before any requests fire
- Silently drop oversized files with a per-file warning; continue uploading the rest
- Auto-retry each failed file once; surface permanent failures with a manual retry button
- Add uploaded images to the gallery as each file completes (progressive invalidation)
- Replace the `+ Image` button with a split/dropdown (single upload / batch upload)
- Route multi-file drags on the app surface to the batch modal

**Non-Goals:**
- Backend changes of any kind
- Per-file metadata (notes, source URL) in batch mode
- Real XHR upload progress percentages (indeterminate spinners are sufficient for now)
- Folder suggestions in batch mode (the `suggested_folder_name` response field is ignored)

## Decisions

### D1 — Separate modal, not an extended UploadModal

`BatchUploadModal.tsx` is a new component. `UploadModal.tsx` is untouched.

**Why**: The two flows have fundamentally different state shapes (single file + rich metadata vs. file list + queue). Sharing the component would require mode-switching logic that increases complexity and regression risk with no user-facing benefit.

**Alternative considered**: A single modal that detects count at selection time. Rejected because the file picker lives inside the current modal, making clean detection awkward, and because it couples two independent features.

### D2 — Per-file state machine

Each file in the batch is tracked with a status:

```
OVERSIZED   — failed pre-upload size validation, never queued
PENDING     — queued, waiting for a concurrency slot
UPLOADING   — one of the 3 concurrent slots
SUCCESS     — all three steps completed
FAILED      — first attempt failed, will auto-retry once
FAILED_FINAL — second attempt failed, manual retry available
```

Retry count is tracked per-file. Manual retry resets a `FAILED_FINAL` file to `PENDING` and re-enters the queue — no further auto-retry beyond that manual attempt.

### D3 — Concurrency via a running-count gate

Three uploads run in parallel. The queue is managed with a simple in-flight counter rather than a library. When a file finishes (success or permanent failure), the next `PENDING` file is started if the counter is below 3.

**Why**: A counter is simple, predictable, and matches the stated requirement. No dependency needed.

### D4 — Vision API suppression by ignoring the response field

The batch modal calls `completeUpload` as-is. The `suggested_folder_name` field in the response is ignored — no suggestion UI is shown. The Vision API may still execute server-side; suppression here means the UX doesn't surface it.

**Why**: Backend is untouched. The suggestion flow only makes sense for a single deliberate upload, not a batch import.

### D5 — File count cap: reject all, show error

If more than 20 files are selected or dropped, no files are accepted and an inline error is shown. The user must re-select.

**Why**: Silently trimming to 20 risks confusing the user about which files were dropped. An explicit rejection is easier to reason about.

### D6 — Oversized files: drop individually, continue with the rest

Files exceeding 50 MB are filtered out before queuing. They appear in the file list with an "Too large" status badge. Valid files proceed to upload without waiting for user confirmation.

**Why**: One oversized file should not block a 19-file batch. Users organising their own files should not be penalised for a single outlier.

### D7 — Entry point: split/dropdown button

The `+ Image` button in `AppLayout` becomes a dropdown with two items: "Upload image" (opens existing `UploadModal`) and "Upload multiple images" (opens `BatchUploadModal`). Both modals are mounted in `AppLayout`.

**Why**: Keeps both paths equally discoverable at the top-level surface. Avoids burying batch upload inside the single-upload modal.

### D8 — Multi-file drag routes to batch modal with files pre-loaded

`handleMainDrop` in `AppLayout` checks `files.length`. If > 1, the batch modal opens with those files pre-populated. If 1, the existing single-file auto-upload path runs unchanged.

### D9 — Progressive gallery invalidation

`queryClient.invalidateQueries({ queryKey: ['images'] })` is called per-file as each upload completes successfully — not batched at the end. Images appear in the gallery as they land.

### D10 — Modal stays open after batch completes

The batch modal does not auto-close. The user closes it manually after reviewing the final state (especially useful when there are failures to retry).

## Risks / Trade-offs

- **Vision API still runs server-side per file**: 20 batch uploads = 20 Vision API calls. This is acceptable for now but may become a cost concern at scale. → Mitigation: a backend flag can be added later to skip Vision for batch; the frontend already ignores the result.
- **No real upload progress %**: Indeterminate spinners give no sense of a large file's progress. → Acceptable for V1; XHR-based progress can be added later by replacing `fetch` in `putToR2` with an XHR wrapper.
- **Manual retry has no auto-retry**: After two failed attempts, a user clicking Retry gets one more try with no auto-retry safety net. → Acceptable; the user is explicitly intervening and has already seen two failures.

## Open Questions

None — all decisions resolved in explore session.
