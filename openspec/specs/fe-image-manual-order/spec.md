### Requirement: Ordered image list initialised from API response

When the gallery is in a folder view, `ImageGrid` SHALL maintain an `orderedImages` state initialised from the query result. The API already returns images sorted by `image_folders.position ASC`, so the initial render reflects the persisted order.

#### Scenario: Images displayed in persisted order on load

- **WHEN** the user navigates to a folder
- **THEN** images are displayed in the order returned by `GET /images?folder_id=<id>` (sorted by position ascending)

### Requirement: Drag-to-reorder enabled only in folder views with Manual sort active

Image cards SHALL be wrapped with `useSortable` from `@dnd-kit/sortable`. Pickup itself (`useSortable`'s `disabled` flag) remains gated only on `isTrash`, unchanged — so image cards stay draggable for moving into folders (`fe-drag-drop-image-to-folder`) in every non-trash context, regardless of the active sort.

The *reorder*-specific mechanics, however, SHALL be active only when `view.type === 'folder'` AND the active sort is `Manual` (see `fe-gallery-sort`):
- The grid is wrapped in `SortableContext` (enabling neighbour-aware drag computation and the reorder drop-indicator preview)
- Dropping an image onto another image triggers `sortEndTrigger` and persists the new order via `PATCH /images/:id/position`

When `view.type !== 'folder'`, OR when it is a folder view but an explicit sort (`Date added`/`Name`) is active, these reorder mechanics are inactive — the grid renders as a plain (non-sortable) list, dropping an image onto another image has no reordering effect, and no `position` updates are persisted. In that state, a folder view is behaviorally identical to All/Unsorted for drag-and-drop purposes: image cards remain draggable to move into folders, but not to reorder.

#### Scenario: Drag-to-reorder active in folder view with Manual sort

- **WHEN** the gallery is in a folder view with `Manual` selected as the sort
- **THEN** image cards can be picked up and dropped onto another image card to reorder, persisting the new `position`

#### Scenario: Drag-to-reorder inactive in folder view with an explicit sort active

- **WHEN** the gallery is in a folder view with `Date added` or `Name` selected as the sort
- **THEN** image cards remain draggable (e.g. to move into another folder) but dropping one onto another does not reorder it, and no `position` update is persisted
- **AND** the reorder drop-indicator preview does not appear

#### Scenario: Drag-to-reorder inactive in unsorted view

- **WHEN** the gallery is in the unsorted view
- **THEN** image cards cannot be reordered via drag (though they remain draggable to move into folders)

#### Scenario: Drag-to-reorder inactive in trash view

- **WHEN** the gallery is in the trash view
- **THEN** image cards cannot be dragged at all (`useSortable` is disabled via `isTrash`)

### Requirement: Drag end reorders local state optimistically

On `DragEndEvent`, `ImageGrid` SHALL:
1. Call `arrayMove` to reorder `orderedImages` immediately (optimistic update)
2. Compute a new fracdex key for the moved item using `KeyBetween(prevItem.position, nextItem.position)` from `fractional-indexing`
3. Call `PATCH /images/:id/position` with `{ folder_id, position: newKey }`

`prevItem` is `orderedImages[newIndex - 1]` after the move; its position key is used as the lower bound. `nextItem` is `orderedImages[newIndex + 1]`; its position key is the upper bound. If there is no previous item, the lower bound is `null`. If there is no next item, the upper bound is `null`. Both bounds use `null` (not `''`) as the "no bound" sentinel, matching the `fractional-indexing` API.

#### Scenario: Item moved to middle — fracdex key computed between neighbours

- **WHEN** the user drags image at index 4 to index 1
- **THEN** the new position key is `KeyBetween(orderedImages[0].position, orderedImages[2].position)`
- **AND** `PATCH /images/<id>/position` is called with that key

#### Scenario: Item moved to first position

- **WHEN** the user drags an image to index 0
- **THEN** the new position key is `KeyBetween(null, orderedImages[1].position)`

#### Scenario: Item moved to last position

- **WHEN** the user drags an image to the last index
- **THEN** the new position key is `KeyBetween(orderedImages[secondToLast].position, null)`

### Requirement: Failed reorder rolls back and shows error toast

If `PATCH /images/:id/position` returns an error, `ImageGrid` SHALL revert `orderedImages` to the state before the drag and show a `toast.error('Failed to save order')`.

#### Scenario: API failure reverts order

- **WHEN** the user reorders an image and the API call fails
- **THEN** the gallery reverts to the pre-drag order
- **AND** a toast error "Failed to save order" is shown

### Requirement: DragOverlay shows ghost of dragged card

During a sort drag, the existing image-drag overlay (a `w-20 h-20` thumbnail card rendered by `DragOverlay` in `AppLayout`) SHALL follow the cursor. The source card SHALL render at 40% opacity in its original position as a visual placeholder, via the `isDragging` state from `useSortable`.

#### Scenario: Ghost visible during drag

- **WHEN** the user picks up an image card
- **THEN** a small thumbnail overlay follows the cursor
- **AND** the source card is shown at reduced opacity as a placeholder
