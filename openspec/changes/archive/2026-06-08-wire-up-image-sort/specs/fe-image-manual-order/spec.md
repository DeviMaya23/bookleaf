## MODIFIED Requirements

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
