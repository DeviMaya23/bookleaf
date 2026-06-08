## Why

`GET /images` already supports `sort`/`direction` query parameters (`created_at`/`title`, `asc`/`desc` — landed in the backend-only `image-list-sort` change), but the frontend has no UI to exercise it. Users currently get a fixed ordering per view — `created_at DESC` for All/Unsorted/Trash, drag-reorder `position ASC` for folders — with no way to switch to alphabetical or oldest-first. This phase wires up a sort control in the gallery toolbar (per the "Filter & Sort" design handoff, Option A) so users can choose how their images are ordered, scoped to exactly the two fields the backend already supports.

## What Changes

- Add a sort icon button to the gallery toolbar (next to the search input, per Option A of the design handoff) that opens a panel with radio options for sort field and a direction toggle.
- Sort field options differ by view:
  - **Folder views**: `Manual` (default), `Date added`, `Name` — `Manual` sits on top and represents "no sort param sent," reproducing today's `position ASC` drag-order behavior exactly.
  - **All / Unsorted / Trash**: `Date added` (default), `Name` — no `Manual` option, since these views have no `position` concept.
- The direction toggle (`↑`/`↓` with field-specific labels — "Oldest/Newest first" for dates, "A→Z"/"Z→A" for names) is hidden when `Manual` is selected, since direction has no meaning for manual order.
- Sort selection resets to the view's default whenever the user switches views (mirrors the existing search-term reset behavior), so sort state is per-view rather than persisted globally.
- Drag-to-*reorder* in folder views is enabled only when `Manual` is the active sort — selecting `Date added` or `Name` turns off the position-reorder mechanics (the card stops being a sortable drop target and no `position` updates are persisted), since reordering would be meaningless (and immediately overridden) under an explicit sort. Dragging an image onto a folder to move it stays enabled regardless of sort choice — exactly like it already works in All/Unsorted — so a non-manually-sorted folder behaves identically to those views for drag-and-drop purposes.
- `getImages`/`getAllImages` gain optional `sort`/`direction` params, threaded through `ImageGrid`'s query key and fetcher so changing the sort triggers a re-fetch from the chosen ordering.

**Out of scope for this phase:** `file_size` and `dimensions` sort fields shown in the design handoff — the backend doesn't support them yet, and adding them would require new backend work (and, for `dimensions`, a derived-data strategy). They're deferred to a future phase once backend support exists. The "Sort & filter"/"chip bar" alternative layouts (Options B/C) from the design handoff are also out of scope — only Option A's two-separate-icon-buttons layout is being built.

## Capabilities

### New Capabilities
- `fe-gallery-sort`: the sort control (toolbar button, panel, field/direction selection, per-view defaults and resets) and its wiring into the image-list query.

### Modified Capabilities
- `fe-image-manual-order`: drag-to-reorder gains an additional gating condition — enabled only in folder views **and** only when the active sort is `Manual` (previously gated on folder view alone).

## Impact

- Frontend: `src/components/AppLayout.tsx` (toolbar layout, sort state, view-change reset), `src/components/ImageGrid.tsx` (query key/fetcher threading `sort`/`direction`, drag-disable condition extension across the three existing `isFolderView`/`view.type === 'folder'` gates), `src/lib/images.ts` (`getImages`/`getAllImages` gain optional `sort`/`direction` params).
- No backend changes — this phase consumes the existing `sort`/`direction` contract on `GET /images` as-is.
- No database changes.
