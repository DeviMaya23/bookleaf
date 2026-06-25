## Context

Relevant existing pieces:

- `AppLayout.tsx` already owns several mutually-exclusive-ish UI mode flags at the same altitude (`selectedImage`, `folderPanelOpen`, `focusMode`, `viewerImage`, `lightboxImage`) and already has a `useEffect` keyed on `viewKey` (folder/view change) that clears several of them at once (`AppLayout.tsx:63-68`). Selection state fits the same pattern.
- `ImageGrid.tsx` owns the only place that has the gallery's current display order (`orderedImages`, from `useManualReorder`, `ImageGrid.tsx:137`). `AppLayout` does not have access to this order — it only passes filter/sort params down and receives callbacks back (`onImageSelect`, `onImageDoubleClick`, etc., `AppLayout.tsx:298-312`).
- `ImageCard` (`ImageGrid.tsx:45`) already has a `ring-2 ring-primary` conditional class for `isDropTarget` (`ImageGrid.tsx:67`) and a per-card `ContextMenu`/`ContextMenuTrigger` wrapper (`ImageGrid.tsx:57-92`).
- `RightPanel.tsx:20-22` is a discriminated union (`mode: 'image' | 'folder'`) rendered from a single `props.mode === 'folder' ? ... : ...` ternary (`RightPanel.tsx:27-31`).
- `FolderInput`/`TokenInput` (`FolderInput.tsx`) is a multi-select, full-membership-sync control used to edit one image's complete folder list — its semantics (replace the full set on save) don't match "add this batch to one additional folder," so it isn't reused as-is for the bulk picker.
- The bulk endpoints already exist and return `{"succeeded_count": n}` with no per-item detail (see archived `images-bulk-folder-trash-endpoints`).

## Goals / Non-Goals

**Goals:**
- Selection bookkeeping (`selectedIds`, anchor) lives in `AppLayout`, since the right panel that consumes it is `AppLayout`'s sibling — consistent with how `selectedImage`/`folderPanelOpen` already work.
- Range computation (shift-click) happens where the order is known (`ImageGrid`), not duplicated in `AppLayout`.
- Reuse existing patterns (mode-flag clearing on view change, ring-based card indicators, immediate-action-on-pick like drag-drop) rather than introducing new ones.

**Non-Goals:**
- Marquee selection (deferred, per proposal).
- Any new global state mechanism — this stays within the existing prop-drilling/lifted-state style already used for the other mode flags.
- Granular cache patching after bulk actions — broad `invalidateQueries(['images'])`, same as existing single-image move/trash handlers.

## Decisions

**Selection state lives in `AppLayout`; range math lives in `ImageGrid`.** `AppLayout` holds `selectMode`, `selectedIds: Set<string>`, and `mainSelectedId: string | null`, and passes them down to `ImageGrid` along with one callback, `onSelectionChange(ids: Set<string>, anchorId: string | null)`. `ImageGrid` handles the click/shift-click branching internally (it has `orderedImages` to slice for the range) and calls `onSelectionChange` with the fully-computed result; `AppLayout` just stores whatever it's given. Alternative considered: lift `orderedImages` up to `AppLayout` — rejected, since the gallery fetch/order hooks (`useGalleryImages`, `useManualReorder`) are intentionally encapsulated inside `ImageGrid`, and duplicating that at a higher layer would mean two sources of truth for ordering.

**New single-select, searchable folder picker for bulk add, not a reuse of `FolderInput`.** `FolderInput`/`TokenInput` edit one image's complete folder membership (multi-select, full-replace-on-save). Bulk add-to-folder is additive and single-destination per action. A new lightweight picker (a search text input filtering the already-fetched `folders` query in `AppLayout`, single click on a filtered result = pick) avoids forcing the bulk flow through multi-select/sync semantics that don't apply to it. A flat unfiltered list was the first pass, but doesn't scale once a user has more than a handful of folders — the search input is the same "type to narrow, then click" pattern already used for the Tags/Folder filter sections in `GalleryToolbar`.

**Picking a folder fires the bulk request immediately; no separate confirm step.** Matches the existing immediate-action pattern (drag-drop moves on drop, single-image trash has no confirm dialog — only *permanent* delete does, via `EmptyTrashDialog`/`DeleteImageDialog`). Bulk trash is a soft-delete, same reversibility as the single-image case, so it gets the same no-confirm treatment.

**Context menu suppression is a prop, not a structural change.** `ImageCard` conditionally skips wrapping in `ContextMenu` (renders the inner `div` directly) when `selectMode` is true, rather than restructuring how the menu is built.

**Toolbar control follows the existing `focusToggle` pattern.** `AppLayout` builds the select-mode toggle as a `ReactNode` (same as it already does for the focus-mode `Toggle`, `AppLayout.tsx:258-269`) and passes it into `GalleryToolbar` as a new prop placed beside the Filters button, rather than `GalleryToolbar` owning the boolean itself.

## Risks / Trade-offs

- **[Risk] Selected IDs can reference images no longer in the current view** after a successful bulk action moves/trashes them (the view's image list updates via `invalidateQueries`, but `selectedIds` itself isn't told to drop those IDs). → Mitigation: a successful bulk action exits select mode entirely (clearing `selectedIds`, the anchor, and turning `selectMode` off) — the same full reset as exiting via the toolbar toggle or navigating away.
- **[Risk] A stale `selectedImage`/`folderPanelOpen` from before entering select mode can resurface in the right panel** the moment `selectedIds` later empties — whether via closing the selection panel, a successful bulk action, or toggling select mode off — because none of those three reset paths touch `selectedImage`/`folderPanelOpen` themselves, only the selection state. The right-panel branch chain falls through to `selectedImage && !focusMode` once `selectedIds.size === 0`, so any leftover `selectedImage` reappears uninvited. → Mitigation: clear `selectedImage`/`folderPanelOpen`/`autoFocusTitle` at the moment select mode is *entered* (`handleSelectModeToggle(true)`), the same mutual-exclusivity pattern `handleImageSelect`/`handleFolderViewDetails` already use for each other — so there's nothing stale left to fall back to regardless of how selection later empties. Closing the selection panel's X was also changed to do the same full reset as a successful bulk action (clear selection + anchor, turn `selectMode` off), for consistency across all three "selection ends" paths.
- **[Risk] `succeeded_count` mismatches selection size silently** (per the already-shipped backend's log-only failure semantics) — if some images in the selection were concurrently deleted by another tab, the toast could show "10 of 12 moved" with no further detail available. → Mitigation: this is an accepted trade-off from the backend design (locked in the earlier proposal), not new here; the frontend just diffs `succeeded_count` against its own selection size for the toast copy.
- **[Risk] Disabling the context menu changes a discoverable affordance** (right-click) into nothing happening, with no visual cue why. → Mitigation: select mode itself is an explicit, visible toggle state (toolbar button), so the absence of context menu is scoped to an obviously-different mode rather than a silent regression.
- **[Risk] Click-dragging across cards can trigger the browser's native text/image selection**, independent of `selectedIds` — outside select mode this was incidentally masked by the `ContextMenuTrigger` wrapper and active `useSortable` listeners (both suppress native selection rendering as a side effect), but the underlying browser `Selection` was still being created and would surface visually once that wrapper/listeners were torn down (e.g. on entering select mode, which removes both). → Mitigation: `select-none` is applied directly to the card element in both modes, so the browser never creates the selection in the first place, rather than relying on incidental suppression that only held in one state.

## Migration Plan

No backend changes, no data migration. Purely additive frontend UI gated behind entering select mode; existing click/drag/context-menu behavior is unchanged when select mode is off. Ships as a normal frontend release.

## Open Questions

(none outstanding — resolved: a successful bulk action exits select mode entirely, same full reset as toggling off or navigating away.)
