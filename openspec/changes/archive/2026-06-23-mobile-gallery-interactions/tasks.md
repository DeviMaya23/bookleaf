## 1. Pointer-capability hook

- [x] 1.1 Create `frontend/src/hooks/useIsCoarsePointer.ts` wrapping `window.matchMedia('(pointer: coarse)')`, re-evaluating on the media query's `change` event.
- [x] 1.2 Write colocated `useIsCoarsePointer.test.ts` covering: initial value reflects `matchMedia` result, and the returned value updates when a `change` event fires.

## 2. Right panel mobile drawer shell

- [x] 2.1 Create `MobileDrawerShell` component (new file under `frontend/src/features/right-panel/components/`): fixed bottom-anchored panel with `translate-y` transform + backdrop, mirroring `FolderSidebar.tsx`'s transform approach on the Y axis. Closes via `onClose` prop on backdrop tap.
- [x] 2.2 Write colocated test for `MobileDrawerShell` covering: renders children, backdrop tap calls `onClose`.
- [x] 2.3 Update `RightPanel.tsx` to call `useIsCoarsePointer()` and branch shell: `MobileDrawerShell` when true, existing `<aside>` sidebar when false. `ImagePanelBody`/`FolderPanelContent` content rendering stays unchanged in both branches.
- [x] 2.4 Update/add tests for `RightPanel.tsx` covering: renders sidebar shell when fine pointer, renders drawer shell when coarse pointer, same content component used in both.

## 3. Gesture remap in AppLayout

- [x] 3.1 Add `lightboxImage` state to `AppLayout.tsx`, alongside existing `selectedImage`/`viewerImage`.
- [x] 3.2 Update the `onImageSelect` callback passed to `ImageGrid`: branch on `useIsCoarsePointer()` — coarse sets `lightboxImage`; fine keeps existing `setSelectedImage`/`setFolderPanelOpen(false)` behavior unchanged.
- [x] 3.3 Update `handleImageDoubleClick` (passed as `onImageDoubleClick`): no-op when `useIsCoarsePointer()` is true; unchanged behavior when false. This is the explicit guard against synthesized double-tap `dblclick` reopening the desktop viewer (per `fe-image-viewer` MODIFIED requirement).
- [x] 3.4 Reset `lightboxImage` to `null` in the existing view-change effect (`AppLayout.tsx:59-63`), alongside `viewerImage`/`selectedImage`.
- [x] 3.5 Update `handleImageDeleted` to also clear `lightboxImage` when the deleted image's id matches, mirroring the existing `viewerImage` check.

## 4. Context menu "View details" item

- [x] 4.1 Add an `onViewDetails?: (image: Image) => void` prop to `ImageCard`/`ImageGrid` (`ImageGrid.tsx`), threaded through to `AppLayout`.
- [x] 4.2 In `ImageCard`'s `ContextMenuContent` (`ImageGrid.tsx:71-92`), add a "View details" item above the existing Delete/Restore items, rendered only when `useIsCoarsePointer()` is true, calling `onViewDetails`.
- [x] 4.3 Wire `onViewDetails` in `AppLayout.tsx` to the same logic the fine-pointer `onImageSelect` branch uses (`setSelectedImage` + `setFolderPanelOpen(false)`), opening the right panel as a bottom drawer.
- [x] 4.4 Update `ImageGrid`/`ImageCard` tests covering: "View details" item present when coarse pointer, absent when fine pointer, calls `onViewDetails` with the correct image.

## 5. Mobile lightbox component

- [x] 5.1 Create `ImageLightbox` component (new file — confirm location: `frontend/src/features/viewer/components/ImageLightbox.tsx` or new `frontend/src/features/lightbox/` folder) implementing: full-viewport overlay, `GET /images/:id` fetch with thumbnail-placeholder-then-full-res-image pattern (reference `ImageViewer.tsx:36-41,98-134` for the pattern; decide extract-vs-duplicate per design.md's note), fit-to-screen `object-fit: contain`, no zoom/rotate/flip/pan.
- [x] 5.2 Add a visible close control that calls `onClose`.
- [x] 5.3 Write colocated `ImageLightbox.test.tsx` covering: shows thumbnail before full-res loads, shows full-res image once loaded, no zoom/rotate/flip controls rendered, close control calls `onClose`.

## 6. Wiring the lightbox into AppLayout

- [x] 6.1 Render `ImageLightbox` in `AppLayout.tsx`'s `<main>` when `lightboxImage` is non-null, as a sibling conditional to the existing `viewerImage`/gallery-grid branch (`AppLayout.tsx:206-212`) — mutually exclusive by construction since only one of `viewerImage`/`lightboxImage` is ever set (per `fe-image-lightbox` MODIFIED requirement).

## 7. Verification

- [x] 7.1 Manually verify on a touch device/emulated coarse pointer: tap opens lightbox, long-press shows context menu with "View details", "View details" opens bottom drawer, double-tap does not open the desktop viewer.
- [x] 7.2 Manually verify on desktop (fine pointer): click opens right panel sidebar exactly as before, double-click opens the desktop viewer exactly as before, no "View details" item in the context menu.
- [x] 7.3 Run `npm run build` in `frontend/` and fix any errors.
- [x] 7.4 Run `npm run lint` in `frontend/` and fix any issues.
