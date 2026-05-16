## Why

There is currently no way to upload images from the frontend. Users have no mechanism to add images to their library or into a specific folder — implementing the upload flow is the core missing piece of the app.

## What Changes

- Add a **+ Image** button in the top-right of the main content area.
- Clicking the button opens an **Upload Modal** with:
  - A drag-and-drop zone (also accepts a file picker click); single image only; accepted types: JPEG, PNG, GIF, WEBP.
  - A **Title** field (optional); placeholder shows the selected filename; field is not auto-filled.
  - A **Submit** button. On submit, title falls back to filename if left blank.
- File validation runs on both drop and file-picker selection; invalid types are rejected immediately.
- Submit calls `POST /images` (initiate upload) then `PUT` to the presigned R2 URL, then `POST /images/:id/complete`.
- The `folder_id` sent to the API is derived from the current URL (`/folders/:folderId`), or omitted when on root (`/`).
- **Success path**: show a success toast, close modal, refresh image list.
- **Folder suggestion path**: if `complete` returns a non-null `suggested_folder_name`, replace the upload form with a suggestion view inside the modal — user can Accept (calls `POST /images/:id/accept-suggestion`) or Ignore. Either action closes the modal with a success toast.
- **Failure path**: show an error toast, keep modal open.
- Add `sonner` as the toast library (not currently installed); wire the `<Toaster>` into the app root.
- Add new API functions: `initiateUpload`, `putToR2`, `completeUpload`, `acceptSuggestion` to `lib/images.ts`.

## Capabilities

### New Capabilities
- `fe-image-upload-flow`: Upload modal (drop zone, file picker, title field, submission flow, folder suggestion view) and toast notifications.

### Modified Capabilities
<!-- none -->

## Impact

- **New component**: `UploadModal.tsx`
- **Modified**: `AppLayout.tsx` — adds the + Image button and wires modal open state
- **Modified**: `lib/images.ts` — adds upload API functions
- **Modified**: `main.tsx` or `App.tsx` — adds `<Toaster />` from `sonner`
- **New dependency**: `sonner` (toast library)
- **No backend changes** required
