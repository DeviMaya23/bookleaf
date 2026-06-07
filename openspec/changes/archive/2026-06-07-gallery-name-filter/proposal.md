## Why

The gallery sidebar and image grid have no way to find a folder or image by name — users with a large folder tree or many images have to scroll and scan visually. Both lists already carry enough data client-side or via the existing list endpoints to support name filtering with minimal backend surface area.

## What Changes

- Add a "Filter folders…" input to the folder sidebar that instantly filters the already-loaded folder tree client-side by case-insensitive substring match on folder name (no backend change — the full tree is already fetched in one shot).
- Add a "Search images by name…" input to the gallery toolbar (leftmost in the row that holds the "Image" upload button) that filters the currently active view (Eagle's pattern) by case-insensitive substring match on image title:
  - In folder views, where the full image list is already fetched unpaginated, filtering happens client-side with no backend round trip.
  - In All / Unsorted / Trash views, which are cursor-paginated, the search is debounced and sent to the backend via a new `name` query parameter on `GET /images`.
- The search term is local to the gallery view, resets when the user switches to a different view, and does not persist across navigation.
- Add an optional `name` query parameter to `GET /images` for case-insensitive substring filtering on image title. It composes with the existing `folder_id`, `unfiled`, `tag_id`, and cursor parameters and only applies to the cursor-paginated (non-folder) query path.

## Capabilities

### New Capabilities
- `fe-gallery-search`: Search-by-name behavior for the image gallery — an input scoped to the active view that filters client-side (folder views) or queries the backend with debouncing (All/Unsorted/Trash views), clearing on view switch.

### Modified Capabilities
- `image-endpoints`: `GET /images` gains an optional `name` query parameter that performs case-insensitive substring filtering on image title, applied only to the cursor-paginated (non-folder) branch of the list query.
- `fe-sidebar-nav`: Add a folder-list filter input performing instant client-side, case-insensitive substring filtering over the already-loaded folder tree.

## Impact

- Backend: `internal/handler/image.go` (query param parsing), `internal/usecase/image_usecase.go` + `image_pagination.go` (`ListImagesParams`), `internal/usecase/image_repository.go` interface, `internal/repository/image_repository.go` (SQL `WHERE title ILIKE` clause on the non-folder branch).
- Frontend: `frontend/src/components/FolderSidebar.tsx` (filter input + client-side tree filtering), `frontend/src/components/AppLayout.tsx` (search input rendered in the toolbar row, `searchTerm`/`debouncedSearchTerm` state, view-switch reset, passed down to `ImageGrid` as props), `frontend/src/components/ImageGrid.tsx` (consumes `searchTerm`/`debouncedSearchTerm` props, branching between client-side filter and debounced server query, `name` added to `queryKeyFor`/`fetcherFor`), `frontend/src/lib/images.ts` (`name` param threaded through `getAllImages`/`getImages`/`getTrashedImages`).
- No database schema changes — filtering uses `ILIKE` on the existing `title` column.
