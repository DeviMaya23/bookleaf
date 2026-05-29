## Why

Users currently have to upload images one at a time, making it tedious to add a batch of images to the gallery. A multi-file upload flow reduces friction for bulk imports without requiring any backend changes.

## What Changes

- The `+ Image` toolbar button becomes a split/dropdown with two options: "Upload image" (existing single-file modal, untouched) and "Upload multiple images" (new batch modal)
- A new `BatchUploadModal` component handles multi-file selection, validation, queued parallel uploads, per-file progress, and per-file error recovery
- Dragging multiple files onto the app surface opens the batch modal with those files pre-loaded (previously only the first file was picked up)
- Vision API folder suggestions are suppressed for batch uploads
- File metadata (notes, source URL) is not collected for batch uploads — filename is used as title, matching the existing single-upload default behaviour

## Capabilities

### New Capabilities

- `fe-batch-upload`: Multi-file upload modal with concurrent queue, per-file progress, partial failure handling, and manual retry

### Modified Capabilities

- `fe-image-upload-flow`: Entry point changes (split button replaces direct button) and multi-file drag handling added to `AppLayout`

## Impact

- `frontend/src/components/AppLayout.tsx` — split button, batch modal state, multi-file drag branch
- `frontend/src/components/BatchUploadModal.tsx` — new file
- `frontend/src/lib/images.ts` — no changes (reused as-is)
- `frontend/src/components/UploadModal.tsx` — no changes
- `frontend/src/lib/dragHandlers.ts` — no changes
