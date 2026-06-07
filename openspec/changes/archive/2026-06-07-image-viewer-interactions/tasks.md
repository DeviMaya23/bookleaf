## 1. Transform state

- [x] 1.1 Add transform state to `ImageViewer`: `zoom: number`, `pan: { x: number; y: number }`, `rotation: 0 | 90 | 180 | 270`, `flipped: boolean`; initialise zoom to `0.5` as placeholder (fit calculation comes next)
- [x] 1.2 Add `containerRef: RefObject<HTMLDivElement>` on the canvas div and a `calcFit` function: `Math.min(containerW / fitW, containerH / fitH) * 0.9` where `fitW`/`fitH` swap when `rotation % 180 !== 0`
- [x] 1.3 Add `useEffect` (no deps) that runs after first render to set `zoom` to `calcFit()` — covers fit-on-open
- [x] 1.4 Add `useEffect` keyed on `image.id` that resets all four transforms: `zoom = calcFit()`, `pan = {x:0, y:0}`, `rotation = 0`, `flipped = false`
- [x] 1.5 Add `ResizeObserver` on `containerRef` (in a `useEffect`) that calls `setZoom(calcFit())` and resets pan when container dimensions change

## 2. Image transform application

- [x] 2.1 Change `<img>` from `className="max-w-full max-h-full object-contain"` to `position: absolute` with `left: 50%; top: 50%`; apply CSS transform string: `translate(-50%, -50%) translate(${pan.x}px, ${pan.y}px) scale(${zoom}) scaleX(${flipped ? -1 : 1}) rotate(${rotation}deg)`; remove `object-contain` (image renders at natural size, zoom controls scaling)

## 3. Pan

- [x] 3.1 Add drag pan state: `dragging: boolean`, `dragOrigin: useRef<{x,y}|null>`, `panSnapshot: useRef<{x,y}>`
- [x] 3.2 Add `onMouseDown` on canvas: set `dragging = true`, record `dragOrigin` and snapshot current `pan`
- [x] 3.3 Add `onMouseMove` on canvas: if dragging, `setPan({ x: snapshot.x + (e.clientX - origin.x), y: snapshot.y + (e.clientY - origin.y) })`
- [x] 3.4 Add `onMouseUp` and `onMouseLeave` on canvas: set `dragging = false`, clear `dragOrigin`
- [x] 3.5 Apply `cursor: dragging ? 'grabbing' : 'grab'` to the canvas div

## 4. Zoom (wheel, cursor-centered)

- [x] 4.1 Add native wheel listener via `useEffect` on `containerRef`: `addEventListener('wheel', handler, { passive: false })`; remove on cleanup
- [x] 4.2 In wheel handler: compute `factor = deltaY < 0 ? 1.1 : 1/1.1`, `newZoom = clamp(zoom * factor, 0.05, 8)`
- [x] 4.3 Compute cursor offset from container center: `cx = e.clientX - rect.left - rect.width/2`, `cy = e.clientY - rect.top - rect.height/2`
- [x] 4.4 Adjust pan to keep cursor-anchored point fixed: `newPanX = cx - (cx - pan.x) * (newZoom / zoom)`, `newPanY = cy - (cy - pan.y) * (newZoom / zoom)`; call `setZoom` and `setPan` together

## 5. Toolbar controls

- [x] 5.1 Wire rotate button: `setRotation(r => ((r + 90) % 360) as 0|90|180|270)`, then in a follow-up effect reset `zoom = calcFit()` and `pan = {x:0,y:0}` — use a ref to detect rotation changes and trigger the reset
- [x] 5.2 Wire flip button: `setFlipped(f => !f)`
- [x] 5.3 Wire 1:1 button: `setZoom(1)`, `setPan({x:0,y:0})`
- [x] 5.4 Make zoom slider controlled: `value={Math.round(zoom * 100)}`, `onChange={e => setZoom(Number(e.target.value) / 100)}`
- [x] 5.5 Update zoom percentage label to display `${Math.round(zoom * 100)}%`

## 6. Unit tests

- [x] 6.1 Test: viewer initialises with a non-zero fit zoom (zoom label does not show "0%")
- [x] 6.2 Test: clicking rotate button updates the image's CSS transform to include `rotate(90deg)`
- [x] 6.3 Test: clicking flip button updates the image's CSS transform to include `scaleX(-1)`
- [x] 6.4 Test: clicking 1:1 button sets zoom label to "100%"
- [x] 6.5 Test: dragging the zoom slider updates the zoom percentage label

## 7. Build check

- [x] 7.1 Run `npm run build` from `frontend/` and fix any TypeScript or lint errors
