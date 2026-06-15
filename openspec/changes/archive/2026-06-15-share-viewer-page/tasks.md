## 1. Backend: SharedImage width/height

- [x] 1.1 In `backend/internal/usecase/share_usecase.go`, add `Width *int` and `Height *int` to `SharedImage`, and set them from `img.Width`/`img.Height` in `GetSharedFolder`.
- [x] 1.2 In `backend/internal/handler/share.go`, add `width *int` / `height *int` (JSON `width`/`height`) to `sharedImageResponse`, and map them from `shared.Images[i].Width`/`Height` in `GetSharedFolder`.

## 2. Frontend: share API client

- [x] 2.1 In `frontend/src/lib/share.ts`, add `getSharedFolder(token: string)` calling `GET /share/:token` via `apiFetch(path, () => Promise.resolve(undefined))` (no auth), throwing on non-OK; add `SharedFolderResponse`/`SharedImage` types matching the response shape (`folder: { name, notes }`, `images: { title, thumbnail_url, full_res_url, width, height }[]`).

## 3. Route and page shell

- [x] 3.1 In `frontend/src/App.tsx`, add `<Route path="/share/:token" element={<SharePage />} />` inside the `PublicThemeLock` route group.
- [x] 3.2 Create `frontend/src/features/share-viewer/components/SharePage.tsx`: read `:token` via `useParams`, `useQuery(['share', token], () => getSharedFolder(token), { retry: false })`, and branch on loading / error / success.
- [x] 3.3 Loading state: centered spinner (`Loader2`, matching `ImageGrid`'s loading pattern), no topbar/grid/panel.
- [x] 3.4 Error state (query `error` set, e.g. 404): centered `AlertCircle` (`text-destructive`) + "This link is invalid or has expired" + "Shared via Bookleaf" branding link to `/`, no topbar/grid/panel.
- [x] 3.5 Success state shell: `h-screen flex flex-col` with a topbar (`"Bookleaf"` wordmark + separator + folder name) and a `flex-1 flex` body containing the image area and `SharedFolderPanel`.

## 4. Masonry grid integration

- [x] 4.1 Create `frontend/src/features/share-viewer/lib/toGalleryImage.ts`: maps a `SharedImage` + index into the `Image` shape from `lib/images.ts` (per design.md decision 1), with neutral placeholders for unused fields and `id: String(index)`.
- [x] 4.2 In `SharePage`'s image area, track `containerWidth` via the same `ResizeObserver` pattern as `ImageGrid`, and render `MasonryLayout` with the adapted images, `dropIndicatorId={null}`, and a share-page `renderCard`.
- [x] 4.3 Empty-folder state: when `images.length === 0`, render the `ImageIcon` + "No images here yet" empty-state pattern from `ImageGrid` in place of the grid.

## 5. Image card and download button

- [x] 5.1 Create `frontend/src/features/share-viewer/components/SharedImageCard.tsx` (the `renderCard` for `MasonryLayout`): thumbnail + title, styled like `MasonryCardContent`, wrapped in `className="group relative ..."`.
- [x] 5.2 Add a hover-reveal download button: `Button` (`variant="secondary"`, `size="icon-sm"`, `className="absolute bottom-2 right-2 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"`) with a `Download` icon, rendered via Base UI's render prop as `<a href={image.full_res_url} download>`.
- [x] 5.3 Wire `onDoubleClick` on the card to call `SharePage`'s `setLightboxIndex(index)`.

## 6. Lightbox

- [x] 6.1 Create `frontend/src/features/share-viewer/components/Lightbox.tsx`: props `images: SharedImage[]`, `index: number`, `onClose: () => void`, `onNavigate: (dir: 1 | -1) => void`. Renders a fullscreen overlay with the current image's `full_res_url`, a counter (`{index + 1} / {images.length}`), and caption (`title`).
- [x] 6.2 Add close (`X`), prev (`ChevronLeft`), and next (`ChevronRight`) `Button`s (`size="icon-sm"`, `rounded-full`), prev/next disabled/hidden at the first/last image.
- [x] 6.3 Add a `keydown` listener (mirroring `ImageViewer`'s Escape pattern) for `Escape` (close), `ArrowLeft` (prev, if not first), `ArrowRight` (next, if not last).
- [x] 6.4 Clicking the backdrop calls `onClose`; clicking the image or controls calls `stopPropagation` so it does not close.
- [x] 6.5 In `SharePage`, add `lightboxIndex: number | null` state; render `Lightbox` when non-null, with `onNavigate` clamping/moving the index and `onClose` resetting it to `null`.

## 7. Side panel

- [x] 7.1 Create `frontend/src/features/share-viewer/components/SharedFolderPanel.tsx`: read-only sections styled like `FolderPanelContent` (bordered rows, uppercase section labels) — folder name (`<h1>`), image count (`images.length`), and Notes (`folder.notes` text, or "No notes added" if `null`).
- [x] 7.2 Add a footer with an "Export folder" `Button` (icon + label, like `DownloadButton`) and a "Shared via Bookleaf" branding line.
- [x] 7.3 Wire the export button to download `GET /share/:token/export` (mirroring `FolderPanelContent.handleExport`'s blob/anchor pattern, but via an unauthenticated fetch); disable it when `images.length === 0`.

## 8. Tests

- [x] 8.1 `SharePage.test.tsx`: loading state renders spinner only; error/404 renders the invalid-link error state; success with images renders topbar + grid + panel; success with zero images renders empty-state message and disabled export.
- [x] 8.2 `SharedFolderPanel.test.tsx`: renders name/count/notes; renders "No notes added" when `notes` is `null`; export button disabled when `images.length === 0`.
- [x] 8.3 `Lightbox.test.tsx`: opens at the given index; `ArrowRight`/`ArrowLeft` navigate and are no-ops at the bounds; `Escape` and backdrop click call `onClose`; clicking the image does not call `onClose`.
- [x] 8.4 `SharedImageCard.test.tsx` (or covered in `SharePage.test.tsx`): download button is present with `href` equal to `full_res_url` and `download` attribute set; double-click triggers the lightbox-open callback.
- [x] 8.5 In `backend/internal/usecase/share_usecase_test.go`, extend `GetSharedFolder` scenarios: an image with non-nil `Width`/`Height` produces a `SharedImage` with matching `Width`/`Height`; an image with nil `Width`/`Height` produces a `SharedImage` with nil `Width`/`Height`.
- [x] 8.6 In `backend/internal/handler/share_test.go`, extend `GetSharedFolder` scenarios to assert the response JSON includes `width`/`height` per image (set and `null`).

## 9. Verification

- [x] 9.1 Run `golangci-lint run` from `backend/` and fix any issues.
- [x] 9.2 Run `npm run build` and `npm run lint` in `frontend/` and fix any issues.

## 10. Card selection outline and download button position

- [x] 10.1 In `SharedImageCard`, add a local `isSelected` state toggled by clicking the card (not the download button); render a thin outline (`ring-1 ring-primary`) when selected.
- [x] 10.2 Reposition the hover download button so it sits at the bottom-right corner of the image itself (not overlapping the title line below it) — wrap just the image area in `relative` rather than the whole card.
- [x] 10.3 Extend `SharePage.test.tsx` (or a new `SharedImageCard.test.tsx`): clicking a card toggles the selection outline on and off; clicking the download button does not toggle it.
- [x] 10.4 Run `npm run build` and `npm run lint` in `frontend/` and fix any issues.

## 11. Fix theme lock and force real file downloads

- [x] 11.1 In `SharePage.tsx`, apply `bg-background text-foreground` to the loading, error, and success-state root containers, matching the pattern used by `LandingPage`/`SimplePageLayout` under `PublicThemeLock`, so the page stays on the warm theme regardless of the viewer's global theme preference.
- [x] 11.2 Make the "Bookleaf" wordmark in `SharePage`'s topbar a `Link to="/"`, matching the "Shared via Bookleaf" branding link.
- [x] 11.3 In `backend/internal/usecase/share_usecase.go`, add `DownloadURL` to `SharedImage` and populate it via `store.GeneratePresignedDownloadURL(ctx, img.R2Path, filename, presignedGetTTL)` (filename derived from `img.Title` + `downloadFileExtension(img.MIMEType)`), so the download link forces `Content-Disposition: attachment` independent of `FullResURL` (used for inline display).
- [x] 11.4 In `backend/internal/handler/share.go`, add `download_url` to `sharedImageResponse` and map it from `SharedImage.DownloadURL`.
- [x] 11.5 In `frontend/src/lib/share.ts`, add `download_url` to `SharedImage`.
- [x] 11.6 In `SharedImageCard.tsx`, use `download_url` (not `full_res_url`) for the download button's `href`.
- [x] 11.7 Extend `share_usecase_test.go` and `share_test.go` to assert `DownloadURL`/`download_url`; update `SharePage.test.tsx` and `Lightbox.test.tsx` fixtures with `download_url`.
- [x] 11.8 Run `golangci-lint run` from `backend/`, and `npm run build` / `npm run lint` from `frontend/`, fixing any issues.
