## Why

`GET /images` already supports `folder_ids`/`tag_ids`/`mime_types` match-any filters (landed in the backend-only `advanced-gallery-filters` change), but the frontend has no UI to exercise them. Users can currently only narrow the gallery via the free-text name search. This phase wires up a filter control in the gallery toolbar — per Option C ("inline filter chip bar") of the "Filter & Sort" design handoff — so users can filter by tags, file type, and folder, with active filters visibly surfaced as removable chips.

## What Changes

- Add a "Filters" button to the gallery toolbar (alongside the existing search input and sort icon button, outside the current `max-w-xs` wrapper) that opens a panel for selecting tags, file types, and folders. The button shows a badge with the count of active filters.
- When one or more filters are active, a second row appears below the toolbar showing each active filter as a removable chip (label + `×`), plus a "Clear all" action. The row is absent entirely when no filters are active (no reserved space).
- Filter options shown in the panel differ by view:
  - **All**: Tags, File type, Folder — sent as `tag_ids`/`mime_types`/`folder_ids` query params on `GET /images`
  - **Unsorted**: Tags, File type only — no Folder section, since `unfiled=true` + any `folder_ids` would be a contradiction (always-empty result)
  - **Folder view**: Tags, File type only — no Folder section (the user is already scoped to one folder); matching is done **client-side** over the already-fetched full image list, the same pattern `fe-gallery-search` already uses for in-folder search
  - **Trash**: the Filters control is hidden entirely
- File type options are presented as friendly labels (e.g. "JPEG", "PNG") mapped to/from the underlying `mime_type` values (e.g. `image/jpeg`); tag and folder options use existing `getTags()`/`getFolders()` data and display `name` directly.
- `getImages`/`getAllImages` (`src/lib/images.ts`) gain optional `tagIds`/`mimeTypes`/`folderIds` params, sent as comma-separated query values, threaded through `ImageGrid`'s query key and fetcher so changing filters triggers a re-fetch (mirroring how `sort`/`direction` were threaded in `wire-up-image-sort`).
- Active filter selection resets to empty whenever the user switches views — mirroring the existing search-term and sort reset behavior.
- Active-filter chips reuse this codebase's existing tag-pill styling (`TagInput.tsx`'s `bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs` pattern) rather than introducing the design handoff's bespoke pill styling, for visual consistency with how tags are already displayed elsewhere.

**Out of scope for this phase:** a visible indicator/pill for the active *sort* (the sort icon button's existing active/inactive variant styling is left as-is); Options A/B layouts from the design handoff.

## Capabilities

### New Capabilities
- `fe-gallery-filters`: the filter control (toolbar button, panel, tag/file-type/folder selection, per-view option scoping and resets), the active-filter chip row, and wiring filter selections through to the image list query (server-side for All/Unsorted, client-side for Folder view).

## Impact

- **Frontend**: `src/components/AppLayout.tsx` (toolbar layout — Filters button placement, chip row, filter state, view-scoped reset alongside existing search/sort resets), `src/components/ImageGrid.tsx` (query key/fetcher threading `tagIds`/`mimeTypes`/`folderIds` for All/Unsorted; client-side tag/mime-type filtering for folder view, alongside the existing client-side title-search filter), `src/lib/images.ts` (`getImages`/`getAllImages` gain optional `tagIds`/`mimeTypes`/`folderIds` params), plus a new mime-type ↔ friendly-label mapping (location TBD in design).
- `src/lib/tags.ts`'s `getTags()` and `src/lib/folders.ts`'s `getFolders()` are consumed by the new filter panel (both already exist; `getFolders()` is already fetched in `AppLayout` for the sidebar).
- No backend changes — this phase consumes the existing `tag_ids`/`mime_types`/`folder_ids` contract on `GET /images` as-is.
- No database changes.
