## 1. AppLayout entry point

- [x] 1.1 Replace the `+ Image` button in `AppLayout` with a split/dropdown button offering "Upload image" and "Upload multiple images"
- [x] 1.2 Add `batchUploadOpen` state to `AppLayout` alongside the existing `uploadOpen` state
- [x] 1.3 Update `handleMainDrop` to check `files.length`: route multi-file drops to the batch modal (set pre-loaded files + open batch modal), keep single-file path unchanged
- [x] 1.4 Mount `<BatchUploadModal>` in `AppLayout` JSX, passing `open`, `onOpenChange`, `folderId`, and `initialFiles` props

## 2. BatchUploadModal — scaffold and file selection

- [x] 2.1 Create `frontend/src/components/BatchUploadModal.tsx` with modal shell (Dialog, header, close handler)
- [x] 2.2 Add a drop zone that accepts multiple files via drag-and-drop and a hidden `<input type="file" multiple>` for the picker
- [x] 2.3 Accept an `initialFiles?: File[]` prop and pre-populate the file list when provided (for the multi-file drag entry path)

## 3. BatchUploadModal — validation

- [x] 3.1 Implement file count validation: if total selected files exceed 20, show inline error and reject all files
- [x] 3.2 Implement per-file type validation (JPEG, PNG, GIF, WEBP): mark invalid files with `UNSUPPORTED` status and a "Unsupported type" badge
- [x] 3.3 Implement per-file size validation (≤ 50 MB): mark oversized files with `OVERSIZED` status and a "Too large" badge; valid files in the same selection are unaffected

## 4. BatchUploadModal — upload queue and concurrency

- [x] 4.1 Define the per-file state type: `PENDING | UPLOADING | SUCCESS | FAILED | FAILED_FINAL | OVERSIZED | UNSUPPORTED`
- [x] 4.2 Implement the concurrency gate: track an in-flight counter, run at most 3 uploads simultaneously, start the next `PENDING` file when a slot frees
- [x] 4.3 Implement the three-step upload sequence per file: `initiateUpload` → `putToR2` → `completeUpload` (reuse existing functions from `images.ts`); set `folder_id` from current folder prop, title from filename without extension; ignore `suggested_folder_name` in response
- [x] 4.4 Implement auto-retry: on first failure, transition to `FAILED` and retry once automatically; on second failure, transition to `FAILED_FINAL`
- [x] 4.5 Implement manual retry: clicking "Retry" on a `FAILED_FINAL` file resets it to `PENDING` and re-enters the queue (one attempt, no further auto-retry)
- [x] 4.6 On each file's success, call `queryClient.invalidateQueries({ queryKey: ['images'] })` immediately (progressive gallery update)

## 5. BatchUploadModal — per-file UI

- [x] 5.1 Render the file list: each row shows filename, status indicator, and (for `FAILED_FINAL`) a Retry button
- [x] 5.2 Show an indeterminate spinner for `UPLOADING` files and a neutral pending indicator for `PENDING` files
- [x] 5.3 Show a success icon for `SUCCESS` files
- [x] 5.4 Show an error indicator for `FAILED_FINAL` files alongside the Retry button
- [x] 5.5 Show the appropriate badge for `OVERSIZED` ("Too large") and `UNSUPPORTED` ("Unsupported type") files

## 6. Tests

- [x] 6.1 Write unit tests for `BatchUploadModal`: success scenario (all files upload successfully, gallery invalidated per file) and failure scenario (one file fails twice, shows Retry button)
- [x] 6.2 Write unit tests for validation logic: count cap rejection and per-file oversized/unsupported filtering
