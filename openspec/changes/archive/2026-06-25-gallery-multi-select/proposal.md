## Why

The gallery has no way to act on more than one image at a time — every action (move to folder, trash) is a single-image drag or context-menu action. With the backend bulk endpoints (`POST /images/bulk/add-to-folder`, `POST /images/bulk/trash`) already shipped, the gallery can offer a "select multiple" mode so users can move or trash a batch of images in one action instead of repeating the same gesture per image.

## What Changes

- Add a "select mode" toggle to `GalleryToolbar`, placed next to the existing Filters button. While active:
  - Filter and sort controls remain visible but are disabled (greyed out).
  - Per-card right-click context menu is disabled.
  - Clicking an image card toggles it in/out of the selection instead of opening the viewer or right panel.
  - Shift-clicking an image card replaces the current selection with the contiguous range between a fixed anchor (the most recently plain-clicked image) and the shift-clicked image. The anchor only moves on a plain click. If no anchor exists yet, shift-click behaves like a plain click.
  - No ctrl/cmd-click support in this iteration.
  - Selected cards render a distinct visual indicator (border/ring), independent of and visually distinct from the existing drag drop-target ring.
- Add a new "selection" mode to the right panel, shown whenever at least one image is selected (hidden otherwise). It shows the selected count and two actions:
  - "Add to folder" — opens a folder picker, then calls `POST /images/bulk/add-to-folder` with the current selection.
  - "Move to trash" — calls `POST /images/bulk/trash` with the current selection directly.
  - The right panel's existing focus-mode-hides-panel behavior does not apply to this mode — the selection panel stays visible during focus mode; the user can manually dismiss it.
  - A successful bulk action (either one) exits select mode entirely — same full reset (selection, anchor, and select mode itself) as toggling off or navigating away. This avoids the right panel falling back to a stale `image`/`folder` selection once the selection empties out.
- Selection state (selected IDs and the shift-click anchor) is cleared whenever select mode is turned off. Navigating to a different view/folder goes further and turns select mode itself off (same trigger as the existing viewer/right-panel-clearing effect on view change) — the user does not land in an empty select mode in the new view, and must re-enable it explicitly.
- Marquee (click-and-drag rectangle) selection is explicitly out of scope for this change — deferred to a future iteration.
- Not available on mobile/coarse-pointer devices — select mode is fine-pointer only, consistent with how drag-and-drop is already scoped.
- Not available in the trash view — neither bulk action (add-to-folder, move-to-trash) is meaningful for already-trashed images. Bulk trash operations (e.g. mass restore) are deferred to a future proposal, same as the other deferred items above.

## Capabilities

### New Capabilities
- `fe-gallery-multi-select`: Select-mode toggle, click/shift-click selection mechanics over the gallery grid, per-card selection indicator, and the bulk "add to folder" / "move to trash" actions wired to the existing bulk endpoints.

### Modified Capabilities
- `fe-right-panel`: Adds a third panel mode (`selection`) alongside the existing `image` and `folder` modes, with its own visibility rule (shown when selection is non-empty, regardless of focus mode) distinct from the existing image/folder panel's focus-mode-hides-panel rule.

## Impact

- **`frontend/src/app-shell/AppLayout.tsx`**: New state (`selectMode`, `selectedIds`, anchor ID), branching in the existing view-change effect to clear it, a new branch in the right-panel mode chain, and branching in the existing image-click handler.
- **`frontend/src/features/gallery/components/GalleryToolbar.tsx`**: New toggle control next to Filters; new prop to disable sort/filter controls while select mode is active.
- **`frontend/src/features/gallery/components/ImageGrid.tsx`**: `ImageCard` gains select-mode-aware click handling, a selection indicator class, and conditional suppression of its context-menu wrapper.
- **`frontend/src/features/right-panel/components/RightPanel.tsx`**: New `mode: 'selection'` variant in the existing prop union, with its own panel body component for the bulk actions and folder picker.
- **`frontend/src/lib/images.ts`**: New wrapper functions for `POST /images/bulk/add-to-folder` and `POST /images/bulk/trash` (backend already shipped; no frontend wrapper exists yet).
- No backend changes — this proposal only wires the frontend to the already-shipped bulk endpoints.
