## Why

Folder owners can now generate a public share link for a folder (`feat/share-folder-panel`), but the link doesn't resolve to anything — `/share/:token` has no route or page. Recipients need a read-only page that shows the shared folder's images, lets them view full-resolution images, download individual images, and export the whole folder.

## What Changes

- Add a public route `/share/:token` (outside `AuthGuard`, locked to the "warm" theme) rendering a new `SharePage`.
- On mount, call `GET /share/:token`:
  - **Valid token, has images**: render a topbar (Bookleaf wordmark + folder name), a masonry image grid (reusing `MasonryLayout`/`masonry.ts`), and a new read-only `SharedFolderPanel` side panel (folder name, image count, notes, export button, "Shared via Bookleaf" branding).
  - **Valid token, empty folder**: masonry area shows the existing gallery empty-state pattern ("No images here yet"); the export button in `SharedFolderPanel` is disabled.
  - **Invalid/expired token (404)**: collapse to a centered error state (icon + message + "Shared via Bookleaf" branding) with no topbar/grid/panel chrome.
- Each image card gets a hover-reveal download button linking to the image's `download_url` (a presigned URL with `Content-Disposition: attachment`, so the browser actually downloads rather than navigating to/displaying the image).
- Double-clicking an image card opens a new `Lightbox` overlay: full-resolution image, prev/next navigation (by array position), Escape/click-outside to close, image counter and caption.
- The export button in `SharedFolderPanel` calls the existing `GET /share/:token/export` to download the folder as a zip.
- **Backend**: add nullable `width`/`height` (`*int`) to `SharedImage` and `sharedImageResponse` (the `GET /share/:token` response), sourced from `domain.Image.Width`/`Height`, so the frontend masonry layout can compute aspect-ratio-based card heights without additional requests.
- **Backend**: add `download_url` (`string`) to `SharedImage` and `sharedImageResponse`, a presigned download URL (`Content-Disposition: attachment`) generated via the existing `StorageService.GeneratePresignedDownloadURL`, for the per-card download button.
- New `frontend/src/lib/share.ts` addition: `getSharedFolder(token)` — a public, unauthenticated wrapper around `GET /share/:token`.

**Non-goal**: editing folder name/notes, toggling/copying the share link, and any other owner-facing controls — those remain in `FolderPanelContent` and are not part of this read-only page.

## Capabilities

### New Capabilities
- `fe-share-viewer`: the public `/share/:token` page — route and theme, data loading and its three states (populated, empty, invalid token), masonry grid reuse with per-card download button, lightbox with keyboard/click navigation, and the read-only side panel.

### Modified Capabilities
- `folder-sharing`: `GetSharedFolder` / `SharedImage` / `sharedImageResponse` (`GET /share/:token`) gain `Width`/`Height` (`*int`, nullable) per image, sourced from `domain.Image`, and a `DownloadURL`/`download_url` (`string`) per image, a presigned download URL with `Content-Disposition: attachment`.

## Impact

- `backend/internal/usecase/share_usecase.go`: `SharedImage` struct gains `Width`/`Height` and `DownloadURL` fields; `GetSharedFolder` populates them from `domain.Image.Width`/`Height` and `StorageService.GeneratePresignedDownloadURL`.
- `backend/internal/handler/share.go`: `sharedImageResponse` gains `width`/`height` (`*int`) and `download_url` (`string`) fields.
- `frontend/src/lib/share.ts`: new `getSharedFolder(token)`.
- `frontend/src/App.tsx`: new `/share/:token` route under the "warm" theme lock.
- New `frontend/src/features/share-viewer/` feature directory: `SharePage`, `SharedFolderPanel`, `Lightbox`, and a share-page image card (the `MasonryLayout` `renderCard` for this page, including the hover download button).
- `frontend/src/features/gallery/components/MasonryLayout.tsx` / `masonry.ts`: reused as-is for layout; may need its prop typing loosened from `images: Image[]` to a minimal shape (`{ width, height, ... }`) so the share page's image objects (a different shape than `lib/images.ts`'s `Image`) can be passed in — to be resolved in design.md.
