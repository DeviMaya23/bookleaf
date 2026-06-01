## 1. Dependencies

- [x] 1.1 Add `@dnd-kit/sortable` and `fractional-indexing` to `frontend/package.json` and run `npm install`

## 2. Backend — Add position to list response

- [x] 2.1 Add `FolderPosition *string` to `ImageItem` struct in `backend/internal/usecase/image_usecase.go`
- [x] 2.2 In `ListImages` usecase, when `params.FolderID != nil`, iterate the image's `ImageFolders` to find the matching entry and set `FolderPosition`
- [x] 2.3 Add `Position *string json:"position"` to `imageResponse` struct in `backend/internal/handler/image.go`
- [x] 2.4 Update `toImageResponse` to map `item.FolderPosition` to `response.Position`

## 3. API and Utilities

- [x] 3.1 Add `position: string | null` to `Image` interface in `frontend/src/lib/images.ts`
- [x] 3.2 Add `updateImagePosition(getToken, imageId, folderId, position)` to `frontend/src/lib/images.ts` — calls `PATCH /images/:id/position` with `{ folder_id, position }`
- [x] 3.3 Create `frontend/src/lib/fracdex.ts` — re-export `generateKeyBetween` from `fractional-indexing` as `KeyBetween`

## 4. MasonryLayout Component

- [x] 4.1 Create `frontend/src/components/MasonryLayout.tsx` — accepts `images[]`, `containerWidth`, renders round-robin columns (`item[i] → column[i % numCols]`)
- [x] 4.2 Compute `numCols = Math.max(1, Math.floor(containerWidth / 220))` and `colWidth = (containerWidth - gap * (numCols - 1)) / numCols`
- [x] 4.3 Derive image card height as `colWidth / (image.width / image.height)`, fall back to square when dimensions are null
- [x] 4.4 Render title below thumbnail, truncated to one line

## 5. ImageGrid Refactor

- [x] 5.1 Add `layoutMode: 'masonry'` prop to `ImageGrid` (default `'masonry'`); wire `ResizeObserver` on the container ref to track `containerWidth`
- [x] 5.2 Replace `columns-*` CSS with `MasonryLayout` rendered with observed `containerWidth`
- [x] 5.3 Replace `useDraggable` on `ImageCard` with `useSortable` from `@dnd-kit/sortable`; keep `disabled` when `isTrash` or `view.type !== 'folder'`
- [x] 5.4 Add `SortableContext` (with `rectSortingStrategy`) wrapping all image cards inside `ImageGrid`
- [x] 5.5 Add `DragOverlay` rendering a semi-transparent copy of the active image card during drag
- [x] 5.6 Add `orderedImages` state initialised from `allImages`; reset when query data changes (view switch)

## 6. Reorder Logic

- [x] 6.1 Implement `onDragEnd` handler — call `arrayMove` to update `orderedImages` optimistically, compute new fracdex key with `KeyBetween(prev?.position ?? null, next?.position ?? null)`, call `updateImagePosition` mutation
- [x] 6.2 On mutation error: revert `orderedImages` to pre-drag state and show `toast.error('Failed to save order')`

## 7. Tests

- [x] 7.1 Unit test `MasonryLayout` — success: correct column count for given container width; failure: falls back to 1 column when container width is 0
- [x] 7.2 Unit test `onDragEnd` logic — success: correct fracdex key computed for a middle-position move; failure: `orderedImages` reverts and error toast shown on API failure

## 8. Bruno

- [x] 8.1 Bruno file for `PATCH /images/:id/position` already exists at `bruno/images/update-image-position.bru`
