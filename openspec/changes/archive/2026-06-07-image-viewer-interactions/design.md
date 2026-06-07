## Context

`ImageViewer.tsx` currently renders the full-res image with a static toolbar — buttons are present but wired to nothing. Transform state (`zoom`, `pan`, `rotation`, `flipped`) doesn't exist yet. This change adds all of it, self-contained within `ImageViewer.tsx`.

## Goals / Non-Goals

**Goals:**
- Cursor-centered mouse wheel zoom (5%–800%)
- Drag pan
- 90° CW rotation with fit-reset on each step
- Horizontal flip that composes correctly with rotation
- Two-way zoom slider (wheel updates slider, slider updates zoom)
- 1:1 button: zoom=1.0, pan=(0,0)
- Fit-on-open: viewer calculates and applies fit zoom when it mounts
- Full transform reset when image prop changes

**Non-Goals:**
- Touch / pinch-to-zoom
- Keyboard shortcuts for zoom/rotate/flip
- Persist transforms across sessions
- Prev/next image navigation

## Decisions

### Custom transform implementation — no external library

All pan and zoom logic is implemented manually rather than via `react-zoom-pan-pinch` or similar.

**Why**: RZPP manages its own transform layer. When rotation is applied via a CSS wrapper inside `TransformComponent`, RZPP's pan operates in screen space without knowledge of the inner rotation. This causes drag direction to become semantically incorrect at 90°/270° (dragging right pans the image along its rotated axis, not visually right). Cursor-centered zoom also misbehaves because RZPP's hit-test math assumes the natural image bounding box.

Rolling the implementation keeps all four transforms in a single CSS string where composition is predictable. The total custom code is ~70 lines of state and event handlers — not large enough to justify a dependency with these interaction problems.

**Alternative considered**: `react-zoom-pan-pinch` — rejected due to rotation composition issues described above.

### Single CSS transform string

All transforms are expressed as one `transform` property on the `<img>` element:

```
translate(panX px, panY px) scale(zoom) scaleX(flip ? -1 : 1) rotate(rotationDeg deg)
```

Order is deliberate: `rotate` and `scaleX` are content operations (orient the image), `scale` and `translate` are viewport operations (position and zoom the oriented image). `transform-origin` remains `center center`.

### Image positioned absolutely, centered via offset

The `<img>` is `position: absolute` with `left: 50%; top: 50%; transform: translate(-50%, -50%) ...`. This places the transform origin at the container center, so zoom and pan operate symmetrically from center — no extra offset math needed.

The `translate(-50%, -50%)` prefix is always prepended to the transform string before the user-controlled translate, scale, flip, and rotate.

### Wheel listener via native addEventListener (passive: false)

React synthetic `onWheel` cannot call `preventDefault()` in Chrome (passive by default). The wheel handler is attached via `containerRef.current.addEventListener('wheel', handler, { passive: false })` in a `useEffect`, and removed on cleanup.

### Cursor-centered zoom math

```
const factor = deltaY < 0 ? 1.1 : 1 / 1.1
const newZoom = clamp(zoom * factor, 0.05, 8)

// cursor offset from container center
const cx = clientX - rect.left - rect.width / 2
const cy = clientY - rect.top - rect.height / 2

// shift pan to keep cursor-anchored point fixed
newPanX = cx - (cx - pan.x) * (newZoom / zoom)
newPanY = cy - (cy - pan.y) * (newZoom / zoom)
```

Pan math operates in screen/container coordinates and is rotation-unaware — correct because pan is a screen-space offset regardless of image orientation.

### Fit zoom calculation

```
const isSwapped = rotation % 180 !== 0
const fitW = isSwapped ? image.height : image.width
const fitH = isSwapped ? image.width  : image.height
const fitZoom = Math.min(containerW / fitW, containerH / fitH) * 0.9
```

Called on mount (via `useEffect` after first render when container dimensions are available) and on each rotation step. A `ResizeObserver` on the container recalculates fit if the container resizes (e.g. right panel opens/closes).

### Rotation resets zoom to fit and pan to zero

On each 90° rotation, `zoom` is set to `calcFit()` and `pan` resets to `{ x: 0, y: 0 }`. This prevents the image from going off-screen after dimension swap at 90°/270°, and matches Eagle's behavior.

### Transform reset on image change

A `useEffect` keyed on `image.id` resets all transform state: `zoom` to fit, `pan` to `{ x: 0, y: 0 }`, `rotation` to `0`, `flipped` to `false`.

## Risks / Trade-offs

- [Fit zoom needs container dimensions] → Container is rendered before `calcFit` runs. Using `useEffect` with `containerRef` is safe because the container is always mounted before the effect fires. A `ResizeObserver` handles subsequent resize events.
- [Wheel deltaY units vary across devices/browsers] → `deltaY` is used only for sign (positive = zoom out, negative = zoom in), making it device-agnostic.
- [No touch support] → Out of scope. A follow-up can add pinch-to-zoom using `PointerEvent` or `react-zoom-pan-pinch` if rotation is not a concern on touch devices.

## Open Questions

None.
