## Why

The current lightbox is a bare Dialog that shows the full-res image with no controls — not useful for close inspection of reference images. This change replaces it with a proper in-app viewer that occupies the main panel, giving images the full available space while keeping the metadata panel accessible alongside.

## What Changes

- **Double-clicking** an image card in the gallery opens the image viewer, replacing the gallery grid in the main content area
- The `ImageViewer` component renders with a dark toolbar (back, flip, rotate, zoom controls — static shell in this phase) and the full-res image centered in the remaining space
- The right panel remains visible alongside the viewer unchanged
- The old Dialog-based lightbox is removed from `RightPanel`; the thumbnail is no longer clickable
- Pressing `Esc` or clicking the back button in the toolbar returns to the gallery

## Capabilities

### New Capabilities

- `fe-image-viewer`: Full-res image viewer that replaces the gallery grid in the main panel. Displays the image with a toolbar (back, rotate, flip, zoom controls as non-functional shell) and a filename/dimensions badge. Pan, zoom, rotate, and flip interactions are deferred to a follow-up change.

### Modified Capabilities

- `fe-image-lightbox`: All existing requirements are replaced. The Dialog-based lightbox triggered by thumbnail click is removed. The thumbnail in `RightPanel` is no longer interactive.

## Impact

- `frontend/src/components/ImageViewer.tsx` — new component
- `frontend/src/components/ImageGrid.tsx` — add `onImageDoubleClick` callback prop
- `frontend/src/components/RightPanel.tsx` — remove Dialog lightbox and thumbnail click handler
- `frontend/src/components/AppLayout.tsx` — add viewer state (`viewerOpen`, `viewerImageId`), wire `onImageDoubleClick` from `ImageGrid`, conditionally render `ImageViewer` vs `ImageGrid` in `<main>`
- No backend changes
- No new dependencies
