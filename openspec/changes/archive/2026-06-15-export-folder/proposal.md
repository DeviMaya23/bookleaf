## Why

Users can currently only download one image at a time. When a folder contains many images they want to take elsewhere, downloading each one individually is tedious. Bundling a folder's images into a single zip download removes that friction with minimal new infrastructure, since the existing R2 storage and folder/image data model already support it.

## What Changes

- New authenticated backend endpoint `GET /folders/:id/export` that streams a zip archive of the images directly inside a folder (single-level only — child folders are not recursed into).
- The zip is built and streamed synchronously: each image is fetched from R2 via the existing `StorageService.GetObject` and copied into a zip entry on `http.ResponseWriter`, with no temp files and no background job.
- The zip's filename is the sanitized folder name + `.zip` (via `Content-Disposition`).
- Each entry's filename inside the zip is `<image title>.<ext>` (extension via the existing MIME-to-extension mapping); duplicate names within the same export are disambiguated by appending ` (1)`, ` (2)`, etc.
- A failed/aborted R2 fetch mid-stream simply truncates the response — no retry or partial-zip handling.
- New "Export folder" button added to `FolderPanelContent`, shown in the right panel:
  - Disabled when the folder has 0 images.
  - Shows a "Preparing export..." state while the request is in flight.
  - On completion, fetches the zip via an authenticated request, converts the response to a Blob, and triggers a browser download — a different mechanism from the existing presigned-URL-based single-image download, since the backend itself performs the zip work.

**Non-goal**: exporting across all of a user's folders at once. That would involve much larger totals and likely needs an async job (the existing River queue) plus temporary R2 storage and a notification — out of scope here, and this change does not introduce shared abstractions for it.

## Capabilities

### New Capabilities
- `folder-export`: Backend capability for streaming a zip archive of a single folder's images (single-level) via a new authenticated endpoint, reading each image from R2 and writing directly to the HTTP response with collision-safe entry naming.

### Modified Capabilities
- `fe-folder-panel`: Adds an "Export folder" button to `FolderPanelContent`, including its disabled/preparing states and the client-side fetch-blob-download flow.

## Impact

- **Backend**: new handler method (`backend/internal/handler/folder.go`) and route registration in `main.go`; new `FolderUsecase` method to list the folder's images and stream the zip; reuses `StorageService.GetObject` and the existing image-folder listing query; reuses `downloadFileExtension` from `image.go`; adds a small filename-sanitization + collision-dedup helper.
- **Frontend**: `FolderPanelContent.tsx` gains the export button and its states; new helper in `frontend/src/lib/folders.ts` to call the export endpoint and trigger the download.
- **No database schema changes.**
- **No new dependencies** — Go's `archive/zip` is part of the standard library.
