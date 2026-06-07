## Why

The image viewer shell (landed in `image-viewer-shell`) renders the full-res image but all transform controls are static — the toolbar buttons do nothing. This change wires the interactions: pan, zoom, rotate, and flip, turning the viewer into a usable inspection tool.

## What Changes

- **Mouse wheel** zooms in/out centered on the cursor position (Eagle-style cursor-anchored zoom)
- **Drag** pans the image freely within the canvas
- **Rotate button** rotates 90° clockwise per click (cycles 0 → 90 → 180 → 270 → 0); resets zoom to fit and pan to center on each rotation
- **Flip button** toggles horizontal flip; combines correctly with any rotation
- **Zoom slider** controls zoom level (5%–800%) and is kept in sync with wheel zoom
- **1:1 button** resets zoom to 100% and clears pan
- **Fit zoom on open** — viewer opens with image scaled to fit the canvas (not at 100%)
- All transforms are view-only; the stored image is never modified
- All transforms reset when the viewer is closed and reopened for any image

## Capabilities

### New Capabilities

- `fe-image-viewer-interactions`: Pan, zoom, rotate, and flip interactions for the image viewer. Covers cursor-centered wheel zoom, drag pan, 90° rotation with fit-reset, horizontal flip, zoom slider sync, 1:1 reset, and fit-on-open behavior.

### Modified Capabilities

- `fe-image-viewer`: Toolbar controls transition from non-functional shell to fully wired interactions.

## Impact

- `frontend/src/components/ImageViewer.tsx` — all changes confined here; adds transform state and event handlers
- No new dependencies
- No backend changes
