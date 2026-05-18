## Why

The app has no way to view or edit image metadata without leaving the gallery. A right-side panel surfaces key metadata inline — and exposes the download action — without disrupting the browsing flow.

## What Changes

- **New right panel** (320px) opens when a user clicks an image card; shows thumbnail, title, notes, source URL (editable), detail grid, and a Download image button
- **Thumbnail in panel is clickable** and opens the existing lightbox for full-resolution viewing
- **Gallery layout** changes from fixed-column grid to Pinterest-style masonry (natural aspect ratios, no card border)
- **PATCH /images/:id** gains a `source_url` patchable field
- **Download image button** calls `GET /images/:id/download` and triggers a browser download; filename is the image title (already handled by BE)
- Selected image state lifts from `ImageGrid` to `AppLayout` to allow the panel to co-exist at layout level
- Colour palette strip and tags are out of scope (not yet implemented in BE)

## Capabilities

### New Capabilities

- `fe-right-panel`: Right panel component — metadata display (title, notes, source URL, details), editable source URL saved via PATCH, Download image button

### Modified Capabilities

- `image-edit`: Add `source_url` as a patchable field on `PATCH /images/:id`
- `fe-gallery-view`: Replace fixed-column grid with masonry layout; image card click now opens the right panel instead of the lightbox
- `fe-image-lightbox`: Lightbox trigger changes from image card click to thumbnail click inside the right panel

## Impact

- **Backend**: `backend/internal/handler/image.go`, `backend/internal/usecase/image_usecase.go`, `backend/internal/repository/image_repository.go`, handler/usecase unit tests, Bruno file
- **Frontend**: `AppLayout.tsx` (state lift + panel slot), `ImageGrid.tsx` (masonry layout, click handler), new `RightPanel.tsx`, `src/lib/images.ts` (add `updateImage`, `downloadImage` functions)
- **No new dependencies expected** — masonry via CSS `column-count`, panel layout via existing Tailwind/shadcn primitives
