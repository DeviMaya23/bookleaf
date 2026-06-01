## Context

The gallery currently renders images using CSS `column-count`, which fills columns top-down. This means DOM index does not match visual position, making drag-to-reorder with dnd-kit impossible to wire correctly. The backend already stores per-folder positions as fracdex keys (`image_folders.position`) and exposes `PATCH /images/:id/position` — the frontend just hasn't used it yet.

The folder view also already returns images sorted `ASC` by position with no pagination cursor, so the full ordered list is always available.

## Goals / Non-Goals

**Goals:**
- Replace CSS columns masonry with explicit column-assignment masonry where `item[i] → column[i % N]`
- Drag-to-reorder within folder views, with positions persisted via `PATCH /images/:id/position`
- Column count adapts to container width via `ResizeObserver`
- `layoutMode` prop seam on `ImageGrid` for future `justified` and `grid` modes

**Non-Goals:**
- Manual ordering on unsorted or all-images views
- Implementing justified or grid layout modes now
- True masonry (shortest-column-first) — intentionally avoided because it breaks flat array order

## Decisions

### 1. Round-robin column assignment over true masonry

True masonry places each item in the shortest column. This is visually optimal but breaks the invariant that flat array index maps to visual position — item at index 3 might end up in column 0 row 2 or column 2 row 1 depending on heights above it.

Round-robin (`item[i] → column[i % N]`) means the visual column of an item is always deterministic from its array index. DnD reorders the flat array; column assignment follows automatically. The columns will have unequal heights — this is acceptable and matches Eagle's "waterfall" look.

### 2. `MasonryLayout` as a standalone component

`ImageGrid` selects which layout component to render based on `layoutMode`. Each layout is an isolated component receiving `images[]` and a `containerWidth`. When `justified` or `grid` are added, they slot in without touching `ImageGrid`'s DnD or data-fetching logic.

### 3. Column count from `ResizeObserver` + target column width

A `TARGET_COL_WIDTH = 220` constant (px) determines the base column width. Column count = `Math.max(1, Math.floor(containerWidth / TARGET_COL_WIDTH))`. The actual column width = `(containerWidth - gaps) / numCols`.

`ResizeObserver` watches the gallery container ref. When the right panel opens/closes and the container shrinks or grows, column count recomputes automatically.

### 4. `@dnd-kit/sortable` on a flat `orderedImages` array

`ImageGrid` owns an `orderedImages` state initialised from the query result. `SortableContext` wraps all items with `strategy: rectSortingStrategy`. On `DragEndEvent`, `arrayMove` reorders the local state (optimistic), then a mutation calls `PATCH /images/:id/position`.

The `useDraggable` calls currently on `ImageCard` are replaced by `useSortable`. The existing `DndContext` in `AppLayout` that handles image-to-folder dragging is separate — folder drop targets remain unaffected.

### 5. Fracdex key computation on drag end

When item at `oldIndex` is dropped at `newIndex`, the new position key is:

```
prevKey = orderedImages[newIndex - 1]?.position ?? ''
nextKey = orderedImages[newIndex + 1]?.position ?? null
newKey  = KeyBetween(prevKey, nextKey)   // fractional-indexing
```

Edge cases: first position → `KeyBetween('', firstItem.position)`, last position → `KeyBetween(lastItem.position, null)`.

Only the moved item gets a new key. Other items retain their existing positions.

### 6. Optimistic update with rollback

Local `orderedImages` state updates immediately on drag end (good UX, no flicker). If the API call fails, a `toast.error` is shown and `orderedImages` is reset to the pre-drag state.

Reordering is disabled (sortable `disabled` prop) on non-folder views (`view.type !== 'folder'`), matching the same `disabled` pattern already used for the trash view.

## Risks / Trade-offs

**Concurrent reorder from two clients** → Last write wins. Position keys are independent per item — only the moved item's key is sent. Acceptable for a single-user app.

**Images with empty position (`''`)** → New images get a fracdex key assigned at add time (backend already does this). Legacy images without a key would sort to the top. Mitigated by the backend's `COALESCE(MAX(position), '')` default — all folder memberships created after the position feature was introduced have keys. Treat `''` as before all others; sorting still works.

**Column count changes on resize during drag** → A mid-drag resize is disorienting. Acceptable edge case — users rarely resize during a drag. `ResizeObserver` recomputes after drag ends anyway.

**`orderedImages` state diverges from server after failed update** → Rollback to pre-drag state on error. Subsequent successful reorders will re-sync from the fresh optimistic state.

## Open Questions

- Should `TARGET_COL_WIDTH` be a user-configurable setting (like Eagle's zoom slider) or a fixed constant for now? → Fixed constant for now; this can become a setting when the layout picker is built.
