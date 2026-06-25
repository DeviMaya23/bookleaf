## 1. API wrappers

- [x] 1.1 Add `bulkAddImagesToFolder(getToken, imageIds: string[], folderId: string): Promise<{ succeeded_count: number }>` to `frontend/src/lib/images.ts`, calling `POST /images/bulk/add-to-folder`
- [x] 1.2 Add `bulkTrashImages(getToken, imageIds: string[]): Promise<{ succeeded_count: number }>` to `frontend/src/lib/images.ts`, calling `POST /images/bulk/trash`

## 2. Selection state and mechanics

- [x] 2.1 Add `selectMode`, `selectedIds` (`Set<string>`), and `mainSelectedId` state to `AppLayout.tsx`
- [x] 2.2 Extend the existing `viewKey`-keyed `useEffect` (`AppLayout.tsx:67-74`) to also turn `selectMode` off and clear `selectedIds`/`mainSelectedId` on view/folder change (navigation exits select mode entirely, not just the selection — distinct from the toggle-off case in 2.3, which only clears the selection)
- [x] 2.3 Add a handler to clear `selectedIds`/`mainSelectedId` when select mode is toggled off
- [x] 2.4 In `ImageGrid.tsx`, implement the click/shift-click selection logic against `orderedImages`: plain click toggles membership and moves the anchor; shift-click replaces `selectedIds` with the inclusive range between the anchor and the clicked image without moving the anchor; shift-click with no anchor behaves as a plain click. Expose this via a single `onSelectionChange(ids, anchorId)` callback up to `AppLayout`
- [x] 2.5 In `ImageCard` (`ImageGrid.tsx`), branch the click handler on `selectMode`: call the selection logic instead of `onSelect`/`onDoubleClick` when active
- [x] 2.6 Add an `isSelected` visual indicator class to `ImageCard`, distinct from the existing `isDropTarget` ring style
- [x] 2.7 Suppress the `ContextMenu`/`ContextMenuTrigger` wrapper in `ImageCard` when `selectMode` is active (render the inner content directly)
- [x] 2.8 Disable drag (`useSortable`) on `ImageCard` while `selectMode` is active, consistent with the existing `disabled: isTrash` pattern

## 3. Toolbar toggle

- [x] 3.1 Add a select-mode `Toggle` control in `AppLayout.tsx`, built the same way as the existing `focusToggle` (`AppLayout.tsx:258-269`), gated to fine-pointer devices (`!isCoarsePointer`) AND non-trash views (`view.type !== 'trash'`), same scoping as the existing Filters button (`GalleryToolbar.tsx:124`)
- [x] 3.2 Add a new prop to `GalleryToolbarProps` for this control (e.g. `selectModeToggle: ReactNode`) and render it beside the Filters button in `GalleryToolbar.tsx`
- [x] 3.3 Add a `controlsDisabled` boolean prop to `GalleryToolbar` that disables the sort dropdown trigger, the Filters dropdown trigger, and the name search input when true

## 4. Selection right panel

- [x] 4.1 Add a `mode: 'selection'` variant to the `RightPanelProps` union in `RightPanel.tsx` (`selectedCount`, `onAddToFolder`, `onMoveToTrash`, `onClose`)
- [x] 4.2 Build a new `SelectionPanelBody` component (in `frontend/src/features/right-panel/components/`) showing the selected count and the two action buttons
- [x] 4.3 Build a minimal single-select, searchable folder picker for "Add to folder" (a search input filtering folders from the already-fetched `['folders']` query; clicking a filtered folder immediately confirms the pick) — a new component distinct from `FolderInput`/`TokenInput`
- [x] 4.4 Wire "Add to folder" picker selection to call `bulkAddImagesToFolder`, then on success: invalidate `['images']`, show a toast (diffing `succeeded_count` against the selection size for the message), and exit select mode entirely (clear `selectedIds`/`mainSelectedId` AND turn `selectMode` off — same full reset as the toggle-off handler, not just a selection clear)
- [x] 4.5 Wire "Move to trash" to call `bulkTrashImages` directly (no confirmation dialog), then on success: invalidate `['images']`, show a toast, and exit select mode entirely (same full reset as 4.4)
- [x] 4.6 In `AppLayout.tsx`, add the `selection` mode as a new, highest-priority branch in the existing right-panel mode chain (`AppLayout.tsx:317-330`) — rendered whenever `selectedIds.size > 0`, ahead of the `image`/`folder` branches, and not gated by `!focusMode`

## 5. Tests

- [x] 5.1 Unit test `ImageGrid`'s selection logic: plain click toggles + moves anchor (added and removed cases); shift-click replaces selection with the correct range; repeated shift-click from the same anchor recomputes correctly; shift-click with no anchor behaves as a plain click
- [x] 5.2 Unit test `ImageCard` selection-mode behavior: context menu suppressed, click does not invoke `onSelect`/`onDoubleClick`, selection indicator class applied/removed based on `isSelected` — assert via `data-testid`/role, not copy text
- [x] 5.3 Unit test `GalleryToolbar`: select-mode toggle renders on fine-pointer only; sort/filter triggers disabled when `controlsDisabled` is true
- [x] 5.4 Unit test `RightPanel` selection mode: renders selected count; takes priority over `image` mode when `selectedIds` is non-empty; remains visible when focus mode is active
- [ ] 5.5 Unit test the bulk action wiring: successful add-to-folder/trash exits select mode entirely (selection, anchor, and `selectMode` all reset); toast reflects `succeeded_count` vs. selection size

## 6. Verification

- [x] 6.1 Run `npm run build` from `frontend/` and fix any issues
- [x] 6.2 Run `npm run lint` from `frontend/` and fix any issues
