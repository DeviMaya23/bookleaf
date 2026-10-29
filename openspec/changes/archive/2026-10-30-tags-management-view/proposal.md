# Proposal

## Why

Tags exist throughout the app — applied to images, surfaced in filters — but there is no dedicated place to manage them. Users cannot rename a misspelled tag or delete one that is no longer needed without going through the image right panel one tag at a time.

## What Changes

- Add a **Tags** entry to the sidebar navigation (below Trash, above the Folders divider)
- Add a `/app/tags` route and a `{ type: 'tags' }` view variant
- Introduce a **Tags view** in the main content area: alphabetically-grouped list, image count per tag, local search, and per-tag actions (rename, delete)
- Clicking a tag navigates to the All view with that tag pre-filled in the gallery filter (via router state)
- The gallery toolbar is hidden when the active view is `tags`; the Tags view renders its own local search input
- Inline rename: the row becomes a text input with Save/Cancel; duplicate-name error shown inline
- Inline delete confirm: the row transforms to "Remove tag? [Cancel] [Delete]" — no modal
- `GET /tags` gains an optional `?include_count=true` query parameter; when present the response includes an `image_count` field per tag (used only by the Tags view)
- `lib/tags.ts` gains `updateTag` and `deleteTag` client functions

## Capabilities

### New Capabilities

- `fe-tags-view`: Tags management page — sidebar entry, `/app/tags` route, view component with alphabetical groups, image count, local search, inline rename and delete confirm, click-through to gallery with pre-filled filter

### Modified Capabilities

- `tag-endpoints`: `GET /tags` optionally returns `image_count` per tag when `include_count=true` query param is passed
- `fe-image-tagging`: `lib/tags.ts` extended with `updateTag(getToken, id, name)` → calls `PATCH /tags/:id`; `deleteTag(getToken, id)` → calls `DELETE /tags/:id`
- `fe-gallery-filters`: Pre-filling the tag filter from router state is introduced as an exception to the "filter resets on navigation" rule — navigating from the Tags view to All with `{ state: { filterTagIds } }` seeds the initial filter

## Impact

- **Backend**: `GET /tags` handler and repository — new optional count join query; route spec (`tag-endpoints`)
- **Frontend**:
  - `src/lib/view.ts` — new `'tags'` view type
  - `src/lib/tags.ts` — two new client functions
  - `src/App.tsx` — new `/app/tags` route
  - `src/app-shell/useAppView.ts` — map `/app/tags` pathname
  - `src/app-shell/AppLayout.tsx` — conditional toolbar rendering; pass router-state seed to `useGalleryControls`
  - `src/features/folder-sidebar/components/FolderSidebar.tsx` — Tags nav entry
  - `src/features/gallery/hooks/useGalleryControls.ts` — accept router-state seed for initial `filterTagIds`
  - New: `src/features/tags/` feature directory with view component and mutations hook
