## 1. API client — thread filter params

- [x] 1.1 In `src/lib/images.ts`, add a `MIME_TYPE_FILTER_OPTIONS: { value: string; label: string }[]` constant (`image/jpeg`→`JPEG`, `image/png`→`PNG`, `image/gif`→`GIF`, `image/webp`→`WEBP`, `image/avif`→`AVIF`); add optional `tagIds?: string[]`, `mimeTypes?: string[]`, `folderIds?: string[]` params to `getImages` and `getAllImages`, each serialized as a comma-separated value on `tag_ids`/`mime_types`/`folder_ids` query params when non-empty (mirroring how `sort`/`direction` are conditionally appended)

## 2. Filter state & per-view scoping in `AppLayout`

- [x] 2.1 Add `filterTagIds: string[]`, `filterMimeTypes: string[]`, `filterFolderIds: string[]` state to `AppLayout`, alongside `searchTerm`/`sortBy`/`sortDir`
- [x] 2.2 Extend the `viewKey`-keyed reset `useEffect` (`AppLayout.tsx:126-134`) to also reset all three arrays to `[]`
- [x] 2.3 Add a per-view filter-section lookup: All → `['tags', 'mimeTypes', 'folders']`, Unsorted/Folder → `['tags', 'mimeTypes']`, Trash → `[]`
- [x] 2.4 Pass the filter state, setters, and per-view section list down to the filter panel (3); pass the read-only arrays down to `ImageGrid` as new props

## 3. Toolbar layout — Filters button + chip row

- [x] 3.1 Wrap the existing search+sort `max-w-xs` div in a new `flex items-center gap-2` div; add the "Filters" button as its sibling (outside `max-w-xs`), rendered only when `view.type !== 'trash'`
- [x] 3.2 Wrap the toolbar row in a `flex-col` container so a chip row can render below it; render the active-filter chip row (5) only when at least one filter is selected, with no reserved space otherwise

## 4. Filter panel UI

- [x] 4.1 Build the "Filters" `DropdownMenu` trigger using `buttonVariants`, switching to the active variant (matching the sort trigger's `default`/`outline` switch) and showing a count badge when `filterTagIds.length + filterMimeTypes.length + filterFolderIds.length > 0`
- [x] 4.2 Render a "Tags" section: `DropdownMenuLabel` + one `DropdownMenuCheckboxItem` per tag from `getTags()`, `checked`/`onCheckedChange` bound to `filterTagIds`
- [x] 4.3 Render a "File type" section (separated by `DropdownMenuSeparator`): one `DropdownMenuCheckboxItem` per `MIME_TYPE_FILTER_OPTIONS` entry, bound to `filterMimeTypes`
- [x] 4.4 Render a "Folder" section (separated by `DropdownMenuSeparator`, All view only per 2.3): one `DropdownMenuCheckboxItem` per folder from `getFolders()`, bound to `filterFolderIds`
- [x] 4.5 Verify whether `DropdownMenuCheckboxItem` (Base UI `Menu.CheckboxItem`) closes the menu on toggle; if so, apply `closeOnClick={false}` (or equivalent) to every checkbox item so multiple filters can be toggled without reopening the panel (design.md Decision #3 open question)

## 5. Active-filter chip row

- [x] 5.1 Build an `activeChips` array combining selected tags (name from `getTags()`), mime types (label from `MIME_TYPE_FILTER_OPTIONS`), and folders (name from `getFolders()`, All view only), each with a `key`/`label`/`onRemove`
- [x] 5.2 Render each chip using `TagInput.tsx`'s pill styling (`inline-flex items-center gap-1 bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs`) with an `X`-icon remove button, plus a "Clear all" text button that resets `filterTagIds`/`filterMimeTypes`/`filterFolderIds` to `[]`
- [x] 5.3 Wire chip remove / "Clear all" to update the corresponding state array(s) in `AppLayout`

## 6. Wire filters into `ImageGrid`'s query (All / Unsorted)

- [x] 6.1 Add `filterTagIds`/`filterMimeTypes`/`filterFolderIds` to `ImageGridProps`
- [x] 6.2 Extend `queryKeyFor` (`ImageGrid.tsx:111-119`) to include the three filter arrays for the `'all'` and `'unsorted'` cases
- [x] 6.3 Extend `fetcherFor` (`ImageGrid.tsx:121-131`) to pass `tagIds`/`mimeTypes`/`folderIds` to `getAllImages` for `'all'`, and `tagIds`/`mimeTypes` (never `folderIds`) to `getImages` for `'unsorted'`

## 7. Folder view — client-side tag/mime filtering

- [x] 7.1 Extend the existing client-side search filter (`ImageGrid.tsx:196-199`) so that, in folder view, an image is shown only if it matches the search term (if any) AND has at least one tag in `filterTagIds` (if any selected) AND its `mime_type` is in `filterMimeTypes` (if any selected) — per design.md Decision #6

## 8. Unit tests

- [x] 8.1 Test: the "Filters" button is hidden in Trash, and shows the correct sections (Tags+File type+Folder for All, Tags+File type only for Unsorted and Folder views)
- [x] 8.2 Test: the "Filters" button badge shows the total count across selected tags/file types/folders, and the button is inactive-styled when nothing is selected
- [x] 8.3 Test: selecting a tag and a file type while viewing All calls `getAllImages` (or issues `GET /images`) with `tag_ids=<id>` and `mime_types=image/jpeg`
- [x] 8.4 Test: selecting a folder while viewing All includes `folder_ids=<id>` in the request; viewing Unsorted never includes `folder_ids` (no Folder section is offered there)
- [x] 8.5 Test: in folder view, selecting a tag filters the displayed images client-side by `image.tags` with no additional network request, and combines correctly with an active search term
- [x] 8.6 Test: switching views clears all selected filters — the chip row disappears and the new view's request includes no filter parameters
- [x] 8.7 Test: removing an individual chip deselects only that filter, and "Clear all" resets all three filter arrays and removes the chip row

## 9. Build check

- [x] 9.1 Run `npm run build` from `frontend/` and fix any TypeScript or lint errors
