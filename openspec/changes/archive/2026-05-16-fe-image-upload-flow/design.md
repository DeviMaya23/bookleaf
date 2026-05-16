## Context

The frontend currently has no image upload capability. The backend already exposes the 3-step upload contract: `POST /images` (initiate, returns presigned PUT URL + image ID), `PUT <presigned-url>` (upload bytes directly to R2), `POST /images/:id/complete` (finalise, may return a folder suggestion). The `folder_id` is already in scope via React Router's `useParams` on `AppLayout`. There is no toast library installed; shadcn uses `sonner` as its recommended toast primitive.

## Goals / Non-Goals

**Goals:**
- Implement the full 3-step upload flow from the UI
- Provide drop zone + file picker with type validation (JPEG, PNG, GIF, WEBP)
- Show a folder suggestion view inline in the modal when the API returns one
- Provide success/error toast feedback
- Keep `folder_id` automatically tied to the current URL context

**Non-Goals:**
- Multi-file upload
- Upload progress bar (the R2 PUT is treated as a black box; no streaming progress)
- Drag-and-drop outside the modal
- Editing or previewing the image after selection

## Decisions

**`UploadModal` as a self-contained component.**
All upload state (selected file, title input, loading, suggestion view) lives inside `UploadModal`. `AppLayout` only owns the `open` boolean. This keeps the modal independently testable and avoids leaking upload state into the layout layer.

**3-step upload executed sequentially inside a single `useMutation`.**
`initiateUpload → putToR2 → completeUpload` are chained in one async function passed to `useMutation`. This gives a single loading/error state to drive the UI. An alternative (3 separate mutations) would complicate error handling and intermediate-state UI with no benefit at this scale.

**`sonner` for toasts.**
It's the shadcn-recommended toast library, has a minimal API, and integrates trivially (`<Toaster />` in root + `toast()` call). No wrapper needed. Alternatives like `react-hot-toast` are equivalent but less idiomatic with the existing shadcn setup.

**Title falls back to filename at submit time, not on input.**
The field is left blank; the placeholder shows the filename. On submit, if `title.trim() === ''` the file's base name (without extension) is used. This avoids confusing auto-fill behaviour and keeps the field clearly optional.

**`folder_id` derived from `useParams` at submit time.**
`AppLayout` already reads `folderId` from `useParams`. The modal receives it as a prop. If `folderId` is undefined (root route), `folder_id` is omitted from the initiate request — the backend will store null. No special handling needed.

**File validation at selection, not only at submit.**
Both the drop handler and the file-picker `onChange` check the MIME type immediately and reject invalid files with an inline error. This gives faster feedback and prevents submitting a bad file.

**Suggestion view replaces the form inside the same modal.**
When `completeUpload` returns a non-null `suggested_folder_name`, a state flag switches the modal body to the suggestion view (Accept / Ignore buttons). There is no navigation or second dialog — keeping it in one modal avoids focus-trap and accessibility complexity.

## Risks / Trade-offs

- **R2 PUT failure**: The image record is already created in the DB before the PUT. If the PUT fails, the record exists but is empty. → Acceptable for now; the backend does not yet expose a cleanup/cancel endpoint, and this edge case is low probability in practice.
- **Large file upload blocking the UI**: The `fetch` PUT to R2 is not streamed; very large files will block until complete. → No progress indicator is in scope; acceptable for typical image sizes.
- **`sonner` adds a dependency**: Small (~5 kB gzipped), well-maintained, and already idiomatic with shadcn — risk is negligible.
