## Why

The current gallery uses CSS `column-count` masonry, which flows items top-down per column — DOM order does not match visual order, making drag-to-reorder impossible to wire up correctly. Switching to an explicit column-assignment layout unblocks manual ordering as the next feature.

## What Changes

- Replace CSS `column-count` masonry in `ImageGrid` with a masonry layout: explicit round-robin column assignment (`item[i] → column[i % N]`), column count derived from container width and a target thumbnail width via `ResizeObserver`
- Add a `layoutMode` prop to `ImageGrid` (seam for future `justified` and `grid` modes; only `masonry` implemented now)
- Add drag-to-reorder using `@dnd-kit/sortable` on a flat `orderedImageIds` array; reordering is only supported within individual folders (not the unsorted or all-images views)
- Persist reordered positions by calling the existing `PATCH /images/:id/position` endpoint with fracdex keys computed on the client using `fractional-indexing` (Rocicorp, same library as the backend — ensures byte compatibility)
- Add `@dnd-kit/sortable` and `fractional-indexing` as new FE dependencies (`@dnd-kit/modifiers` already present)

## Capabilities

### New Capabilities

- `fe-gallery-masonry`: Masonry layout component — explicit column assignment, `ResizeObserver` for reactive column count, aspect-ratio-derived image heights, title card below each image
- `fe-image-manual-order`: Drag-to-reorder images in the gallery; persists order via `PATCH /images/:id/position` using fracdex keys; only available within individual folders (disabled on unsorted and all-images views)

### Modified Capabilities

- `fe-gallery-view`: The layout requirement changes from CSS `column-count` masonry to the new masonry layout; existing behaviour (pagination, empty state, context menu, right panel) is unchanged
- `image-endpoints`: `GET /images` list response adds a `position` field (nullable string) populated from `image_folders.position` when the query is folder-scoped; null for all other views

## Impact

- `frontend/src/components/ImageGrid.tsx` — primary change
- New `frontend/src/components/MasonryLayout.tsx` — isolated layout component; `ImageGrid` selects between layout components, making future `JustifiedLayout` and `GridLayout` additions straightforward
- New `frontend/src/lib/fracdex.ts` (thin wrapper around `fractional-indexing`)
- `frontend/src/lib/images.ts` — add `updateImagePosition` API call, add `position` field to `Image` interface
- `frontend/package.json` — add `@dnd-kit/sortable`, `fractional-indexing`
- `backend/internal/usecase/image_usecase.go` — add `FolderPosition *string` to `ImageItem`; set it in `ListImages` when folder-scoped
- `backend/internal/handler/image.go` — add `Position *string` to `imageResponse`; map from `item.FolderPosition` in `toImageResponse`
