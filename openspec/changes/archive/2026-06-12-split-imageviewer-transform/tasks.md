## 1. Extract useImageTransform

- [x] 1.1 Create `frontend/src/features/viewer/hooks/useImageTransform.ts`: move `zoom`/`pan`/`rotation`/`flipped`/`dragging` state, the stale-closure refs, `containerRef`, `calcFit`, the four effects (reset-on-image-change, `ResizeObserver` refit, reset-after-rotation, native wheel listener), and the drag-to-pan handlers from `ImageViewer.tsx`, ported verbatim. Return `{ containerRef, transform, zoom, setZoom, dragging, dragHandlers, toggleFlip, rotate, resetTo1to1 }` per design.md D1.
- [x] 1.2 Update `ImageViewer.tsx` to call `useImageTransform(image)`, spread `dragHandlers` onto the viewport `<div>`, attach `containerRef`, and wire the toolbar (zoom slider/label, flip button, rotate button, 1:1 button) and `<img>`'s `transform` style to the hook's return values. Remove the now-unused state/refs/effects/handlers from `ImageViewer.tsx`.
- [x] 1.3 Add `useImageTransform.test.ts` using the harness pattern from design.md D2, covering: wheel-zoom (zoom changes and zooms toward cursor position), drag-to-pan (mousedown → mousemove → mouseup updates `transform`'s pan, mouseleave ends drag), and reset-on-image-change (changing `image` resets zoom/pan/rotation/flip).

## 2. Final checks

- [x] 2.1 Run the existing `ImageViewer.test.tsx` suite and confirm rotate/flip/1:1/zoom-slider/fit-on-open scenarios still pass unchanged through the hook.
- [x] 2.2 Run `npm run build` and fix any resulting issues.
