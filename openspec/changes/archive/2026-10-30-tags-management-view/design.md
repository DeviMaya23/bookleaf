# Design

## Context

See `proposal.md — Why` for motivation.

The tag CRUD endpoints (`POST`, `GET`, `PATCH /tags/:id`, `DELETE /tags/:id`) are fully implemented on the backend. The frontend has `getTags` and `createTag` in `lib/tags.ts` but no update/delete client functions. `useGalleryControls` resets all filter state (including `filterTagIds`) on every `viewKey` change via a `useEffect`. `AppLayout` is a single mounted component instance shared across all gallery routes — it does not unmount/remount between `/app`, `/app/unsorted`, `/app/folders/:id`, or the new `/app/tags`.

## Goals / Non-Goals

**Goals:**
- Tags view as a first-class nav destination with full CRUD
- Router-state pre-fill that is consumed once and does not persist across subsequent navigations

**Non-Goals:**
- Bulk-deleting or merging tags
- Drag-to-reorder tags
- Tag colour or icon customisation

## Decisions

### 1. Backend: optional count via a new repository method

The `GET /tags` handler reads an `include_count` query param. When present, it calls a new `ListByUserIDWithCount(ctx, userID)` method on `TagRepository` that runs a single LEFT JOIN + GROUP BY query and returns a `TagWithCount` struct (`Tag` + `ImageCount int`). When absent, it calls the existing `ListByUserID` as before.

The usecase layer gets a matching `ListWithCount` method that delegates to the new repo method. The handler maps both paths to the same response shape — the count field is omitted (`omitempty`) when zero or absent, so the existing response format is unaffected for callers that don't pass the param.

**Alternative considered**: preloading `Images` on the existing `ListByUserID` query. Rejected — loading full image rows to count them is wasteful.

### 2. Frontend feature directory

New `src/features/tags/` directory, consistent with the existing feature layout (`folder-sidebar/`, `gallery/`, `right-panel/`). Contains:

```
src/features/tags/
  components/
    TagsView.tsx          — full view component (list, search, groups)
    TagRow.tsx            — single tag row with view/rename/delete-confirm modes
  hooks/
    useTagMutations.ts    — rename and delete mutations (mirrors useFolderMutations)
```

### 3. Router-state pre-fill: consume-and-clear pattern

`AppLayout` watches `useLocation().state` via a `useEffect([location.state])`. When `location.state.filterTagIds` is present, it imperatively calls `gallery.setFilterTagIds(tagIds)` to seed the filter, then immediately clears the router state via `navigate(location.pathname, { replace: true, state: null })` so the pre-fill is not re-applied if the user navigates away and returns to All.

Effect ordering matters: the `viewKey` reset inside `useGalleryControls` (which clears `filterTagIds` to `[]`) fires before the `location.state` effect in `AppLayout` (registered after the hook call), so the clear always precedes the seed — the pre-fill wins.

**Alternative considered**: passing `initialFilterTagIds?: string[]` to `useGalleryControls` and seeding from `useState`. Rejected — `AppLayout` doesn't remount between routes; the `useState` initializer only runs on mount, which is once for the session. Re-navigation to `/app` from `/app/tags` never re-mounts the component, so the seed would never fire on the second click.

**Alternative considered**: reading router state only on mount. Rejected for the same reason.

**Alternative considered**: URL query params (`/app?tags=<id>`). Rejected — agreed during explore that router state is sufficient and avoids changing `useGalleryControls`' URL-read concerns.

### 4. Inline row state

Each `TagRow` manages its own mode locally: `'view' | 'rename' | 'delete-confirm'`. No modal; no lifted state to the parent list. The parent `TagsView` holds the fetched tag list; `TagRow` calls mutation callbacks (`onRename`, `onDelete`) which bubble up to `useTagMutations` and invalidate the `['tags']` cache on success.

### 5. `filterSectionsForViewType` and `defaultSortForViewType` handle `'tags'`

Both helpers in `useGalleryControls` switch on `view.type`. Adding `case 'tags': return []` to each prevents an unhandled-type warning and ensures the hook doesn't try to render filter sections or sort controls that don't apply.

## Risks / Trade-offs

- **Router-state race**: if the user clicks a tag and immediately hits back before the All view mounts, the state may already be cleared. Acceptable — back navigation takes them to Tags anyway.
- **Stale count**: `image_count` is fetched once on mount and not updated in real time. Renaming a tag or navigating away and back will re-fetch. This is consistent with how the rest of the app treats cached data.

## Migration Plan

No schema changes. The new `ListByUserIDWithCount` query is additive. No data migration required. Deploy is a standard BE + FE release.
