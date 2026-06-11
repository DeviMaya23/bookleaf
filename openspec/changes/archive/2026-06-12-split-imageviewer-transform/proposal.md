## Why

`features/viewer/components/ImageViewer.tsx` (207 lines) bundles the
pan/zoom/rotate/flip "viewport transform engine" — state, stale-closure
refs, fit/resize calculations, and wheel/drag event handling — together
with the toolbar and viewport rendering. The transform engine is the
component's core logic and is currently untestable in isolation: every
existing test goes through full render + `userEvent`/`fireEvent`, and the
wheel-zoom and drag-to-pan handlers have no test coverage at all. Lifting
the engine into its own hook lets it be tested directly via `renderHook`
(simulating wheel/mouse events against the hook's `containerRef`) without
mounting the toolbar or `<img>`.

## What Changes

- Extract the pan/zoom/rotate/flip transform engine out of `ImageViewer.tsx`
  into a new `features/viewer/hooks/useImageTransform.ts`: owns `zoom`,
  `pan`, `rotation`, `flipped`, `dragging` state and the stale-closure refs;
  owns `containerRef` and the `calcFit` calculation; owns the four effects
  (reset-on-image-change, `ResizeObserver`-driven refit, reset-after-rotation,
  native wheel listener); owns the drag-to-pan handlers. Returns
  `{ containerRef, transform, zoom, setZoom, dragging, dragHandlers,
  toggleFlip, rotate, resetTo1to1 }`.
- `ImageViewer.tsx` becomes a thin composition: the `imageDetail` query, the
  Esc-to-close effect, and the toolbar/viewport render, wired to
  `useImageTransform`'s return values. The toolbar JSX stays inline in
  `ImageViewer.tsx` (not extracted further).
- Add `useImageTransform.test.ts` covering zoom/pan/rotation/flip state
  transitions, the fit-on-open and reset-on-image-change behavior, and —
  newly testable — wheel-zoom and drag-to-pan, via `renderHook` and
  simulated events against `containerRef.current`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

`frontend-structure` — adds a requirement describing the new
`features/viewer/hooks/useImageTransform.ts` (+ test) location. This is a
zero-functional-change structural extraction; no behavior changes.

## Impact

- `frontend/src/features/viewer/components/ImageViewer.tsx` shrinks —
  loses all transform state, refs, `calcFit`, the four effects, and the
  drag handlers, replaced by a single `useImageTransform(image)` call.
- New files: `frontend/src/features/viewer/hooks/useImageTransform.ts` and
  `frontend/src/features/viewer/hooks/useImageTransform.test.ts`.
- `ImageViewer.test.tsx`'s existing scenarios (rotate, flip, 1:1, zoom
  slider, fit-on-open) continue to pass through `ImageViewer` unchanged;
  new wheel/drag scenarios are added at the hook level instead.
- No change to `ImageViewer`'s props, callers, or any user-visible behavior.
- No backend, API, database, or browser-extension changes. No new
  dependencies.
