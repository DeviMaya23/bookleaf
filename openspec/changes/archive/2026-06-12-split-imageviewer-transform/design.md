## Context

`ImageViewer.tsx` (207 lines) bundles the pan/zoom/rotate/flip "viewport
transform engine" — state, stale-closure refs, `calcFit`, four effects
(reset-on-image-change, `ResizeObserver`-driven refit, reset-after-rotation,
native wheel listener), and drag-to-pan handlers — together with the
toolbar and viewport rendering. The engine reads/writes `containerRef` for
DOM measurement (`calcFit`) and to attach the native wheel listener, so it
isn't a pure value-in/value-out function — it's a headless hook that owns a
DOM ref, similar in shape to `useDraggable`/`useDroppable` from `dnd-kit`
(already used elsewhere in the codebase).

`globalThis.ResizeObserver` is mocked in `frontend/src/test/setup.ts`: it
fires once on `observe()` with `contentRect = { width: 800 }` (no
`height`), which makes `calcFit` hit its `if (!containerW || !containerH)
return 0.5` branch in tests today. This mock is unaffected by this change.

## Goals / Non-Goals

**Goals:**
- Extract the transform engine into `useImageTransform(image)`, returning
  everything `ImageViewer`'s toolbar and viewport need to render.
- Make wheel-zoom and drag-to-pan — currently untested — testable via a
  minimal render harness around the hook.

**Non-Goals:**
- No toolbar extraction (confirmed out of scope — toolbar stays inline in
  `ImageViewer.tsx`).
- No change to `ImageViewer`'s props, the `imageDetail` query, or the
  Esc-to-close effect — these stay in `ImageViewer.tsx`.
- No behavior change to fit/zoom/rotation/flip/pan math — ported verbatim.

## Decisions

### D1: `useImageTransform` signature

```ts
function useImageTransform(image: Image): {
  containerRef: RefObject<HTMLDivElement>
  transform: string
  zoom: number
  setZoom: (zoom: number) => void
  dragging: boolean
  dragHandlers: {
    onMouseDown: (e: React.MouseEvent) => void
    onMouseMove: (e: React.MouseEvent) => void
    onMouseUp: () => void
    onMouseLeave: () => void
  }
  toggleFlip: () => void
  rotate: () => void
  resetTo1to1: () => void
}
```

- `pan` and `rotation` are not returned individually — neither is displayed
  directly in the toolbar (only baked into the `transform` string), so they
  stay internal to the hook, same as today.
- `zoom`/`setZoom` are returned directly because the toolbar both displays
  the zoom percentage and drives it via the range input.
- `dragHandlers` is a single object spread onto the viewport `<div>`,
  mirroring how `attributes`/`listeners` are spread in `dnd-kit` usages
  elsewhere in this codebase.
- `toggleFlip`, `rotate`, `resetTo1to1` are named mutators (not raw
  setters) for the three toolbar actions that don't need a value passed in.

### D2: Test harness for DOM-ref-dependent behavior

The wheel listener and `ResizeObserver` are attached to `containerRef.current`
inside effects that run on mount — for those effects to do anything in a
test, `containerRef` must already point at a real DOM node *before* the
mount effects run. Bare `renderHook` from `@testing-library/react` doesn't
attach refs to anything, so `useImageTransform.test.ts` will use a small
harness component instead:

```tsx
function Harness({ image }: { image: Image }) {
  const t = useImageTransform(image)
  return (
    <div ref={t.containerRef} data-testid="container" {...t.dragHandlers}>
      <span data-testid="zoom">{t.zoom}</span>
      <span data-testid="transform">{t.transform}</span>
    </div>
  )
}
```

`render(<Harness image={...} />)` attaches the ref during commit (before
effects run), so the wheel listener and `ResizeObserver` observe a real
element. Tests then use `fireEvent.wheel`/`fireEvent.mouseDown` etc. on
`getByTestId('container')` and assert on the exposed `zoom`/`transform`
spans. This is the same "renderHook via a thin wrapper" approach already
implied by `useManualReorder.test.ts`, adapted for the DOM-ref case.

### D3: What moves to hook-level tests vs. stays at `ImageViewer` level

`ImageViewer.test.tsx` already covers rotate/flip/1:1/zoom-slider/fit-on-open
through full render, and these continue to pass unchanged once `ImageViewer`
delegates to the hook (no need to duplicate them at the hook level).
`useImageTransform.test.ts` focuses on what's newly testable and wasn't
covered before:

- wheel-zoom (zoom changes, and zooms toward the cursor position — pan
  adjusts accordingly)
- drag-to-pan (mousedown → mousemove → mouseup changes `pan` via
  `transform`, mouseleave ends the drag)
- reset-on-image-change (changing the `image` prop resets zoom/pan/rotation/
  flip)

### D4: Sequencing

Single hook extraction, one task group — no sequencing decisions needed.

## Risks / Trade-offs

- **[Risk]** The `ResizeObserver` mock only provides `width`, not `height`,
  so `calcFit`'s real fit-calculation branch (`Math.min(containerW / fitW,
  containerH / fitH) * 0.9`) is never exercised by either the existing
  `ImageViewer` tests or the new hook tests — both only reach the `0.5`
  fallback. → **Mitigation**: out of scope for this change (pre-existing
  test-environment limitation, not introduced or worsened by this
  extraction); `calcFit` is ported verbatim.
- **[Trade-off]** `useImageTransform` returning a `dragHandlers` object
  means `ImageViewer` must spread it onto the correct `<div>` — if a future
  edit moves the viewport markup without moving the spread, dragging would
  silently stop working. → **Mitigation**: covered by the drag-to-pan hook
  test (D3) plus `ImageViewer.test.tsx`'s existing render of the viewport.
