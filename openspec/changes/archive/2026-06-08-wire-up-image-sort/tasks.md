## 1. API client — thread `sort`/`direction`

- [x] 1.1 In `src/lib/images.ts`, add optional `sort?: 'created_at' | 'title'` and `direction?: 'asc' | 'desc'` parameters to `getImages` and `getAllImages`; when present, set them as `sort`/`direction` query params on the `GET /images` request (mirroring how `cursor`/`name` are conditionally appended); when absent (i.e. `Manual`), the request shape is unchanged from today

## 2. Sort state & per-view defaults in `AppLayout`

- [x] 2.1 Add `sortBy: 'manual' | 'created_at' | 'title'` and `sortDir: 'asc' | 'desc'` state to `AppLayout`, alongside the existing `searchTerm`/`debouncedSearchTerm` (`AppLayout.tsx:80-81`)
- [x] 2.2 Add a small per-view-default lookup: folder views → `{ sortBy: 'manual', sortDir: undefined }`; All/Unsorted/Trash → `{ sortBy: 'created_at', sortDir: 'desc' }`
- [x] 2.3 Extend (or add a sibling to) the `viewKey`-keyed `useEffect` that resets `searchTerm` (`AppLayout.tsx:92-94`) to also reset `sortBy`/`sortDir` to the new view's default from 2.2
- [x] 2.4 Pass `sortBy`/`sortDir` (and their setters, for the control in 3) down to wherever the sort control renders, and `sortBy`/`sortDir` (read-only) down to `ImageGrid` as new props

## 3. Sort control UI (toolbar button + panel)

- [x] 3.1 Add a sort icon button to the toolbar row in `AppLayout.tsx` (around the search input / upload button area, `AppLayout.tsx:234-269`), built from `DropdownMenu`/`DropdownMenuTrigger` with a `buttonVariants`-styled icon-only trigger (matching the visual placement in Option A of the design handoff)
- [x] 3.2 Inside `DropdownMenuContent`, render a `DropdownMenuRadioGroup`/`DropdownMenuRadioItem` list of sort field options, sourced from the per-view option lists: `['manual', 'created_at', 'title']` for folder views (with `Manual` first), `['created_at', 'title']` for All/Unsorted/Trash
- [x] 3.3 Render a direction toggle row below the radio group (showing field-specific labels — "Oldest first"/"Newest first" for `created_at`, "A → Z"/"Z → A" for `title` — and an `↑`/`↓` indicator) only when `sortBy !== 'manual'`; clicking it flips `sortDir`
- [x] 3.4 Wire field selection to update `sortBy`, and reset `sortDir` to that field's own default direction (`created_at` → `desc`/"Newest first", `title` → `asc`/"A → Z") whenever the selected field changes — whether switching from `Manual` to an orderable field, or between `created_at`/`title` directly. This matches the per-field default convention the backend already establishes (`image-list-sort`)

## 4. Active indicator on sort trigger

- [x] 4.1 Compute `sortActive = sortBy !== viewDefault.sortBy || (sortBy !== 'manual' && sortDir !== fieldDefaultDirection[sortBy])`
- [x] 4.2 Apply a distinguishing `buttonVariants` style/variant to the sort trigger when `sortActive` is true (no new badge/indicator pattern — reuse the variant mechanism the trigger is already built from)

## 5. Wire sort into `ImageGrid`'s query

- [x] 5.1 Add `sortBy`/`sortDir` to `ImageGridProps` (`ImageGrid.tsx:123-132`)
- [x] 5.2 Extend `queryKeyFor` (`ImageGrid.tsx:103-110`) to include `sortBy`/`sortDir` (e.g. `undefined` when `manual`, so the key matches today's shape in that case)
- [x] 5.3 Extend `fetcherFor` (`ImageGrid.tsx:112-121`) to pass `sort`/`direction` through to `getImages`/`getAllImages` — translating `sortBy === 'manual'` to "omit both params" and otherwise passing `sortBy` as `sort` and `sortDir` as `direction`
- [x] 5.4 Verify `placeholderData: keepPreviousData` (`ImageGrid.tsx:180`) keeps prior results visible across a sort-triggered re-fetch, same as it does for `debouncedSearchTerm` changes

## 6. Drag gating: reorder mechanics active only for folder + Manual

- [x] 6.1 Leave `useSortable({ disabled: isTrash, ... })` (`ImageGrid.tsx:53`) unchanged — pickup must remain available outside trash so folder-restructure drag keeps working everywhere (per the corrected Decision #5 in `design.md`)
- [x] 6.2 Change the `SortableContext` wrap condition (`ImageGrid.tsx:327-330`) from `isFolderView` to `isFolderView && sortBy === 'manual'`, so a non-manually-sorted folder renders the plain `grid` exactly like All/Unsorted
- [x] 6.3 Extend the `onDragOver` reorder-preview guard (`ImageGrid.tsx:152`) from `if (!isFolderView || ...) return` to also bail when `sortBy !== 'manual'`
- [x] 6.4 Extend the `sortEndTrigger` persistence effect guard (`ImageGrid.tsx:237`) from `view.type !== 'folder'` to also bail when `sortBy !== 'manual'`, so `position` is only ever written in Manual mode

## 7. Unit tests

- [x] 7.1 Test: selecting `Name` in a folder view calls `getImages` with `sort=title` and the field's default direction (`asc`)
- [x] 7.2 Test: selecting `Manual` in a folder view (after a prior explicit sort) calls `getImages` with no `sort`/`direction` params
- [x] 7.3 Test: switching from a folder (with an explicit sort active) to the "All" view resets the sort control to `Date added` / "Newest first"
- [x] 7.4 Test: switching between two folders resets the sort control to `Manual`
- [x] 7.5 Test: the direction toggle is not rendered when `Manual` is selected, and is rendered (with field-appropriate label) when `Date added`/`Name` is selected
- [x] 7.6 Test: in a folder view with an explicit sort active, dragging an image card onto another image card does not trigger a `position` update (no `PATCH /images/:id/position` call), while dragging onto a folder still calls the move-to-folder handler
- [x] 7.7 Test: the sort trigger renders in its active visual state when a non-default sort is selected, and inactive when the view's default is selected

## 8. Build check

- [x] 8.1 Run `npm run build` from `frontend/` and fix any TypeScript or lint errors
