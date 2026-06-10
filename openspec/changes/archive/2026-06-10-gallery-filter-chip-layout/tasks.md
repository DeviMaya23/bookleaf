## 1. Shared UI primitives

- [x] 1.1 Pull `toggle` and `toggle-group` components into `frontend/src/components/ui/` via the shadcn CLI (`base-nova` style)
- [x] 1.2 Adjust `Toggle`/`ToggleGroup` styling (via `className`/variant overrides at the call site, not the shared component) to render as pill-shaped (`rounded-full`) chips matching the design handoff's `Chip` look, using existing `--secondary`/`--foreground`/`--border` tokens

## 2. Filter panel state

- [x] 2.1 Add `filterTagSearch` and `filterFolderSearch` string state to `AppLayout.tsx`, alongside the existing `filterTagIds`/`filterMimeTypes`/`filterFolderIds`
- [x] 2.2 Reset `filterTagSearch` and `filterFolderSearch` in the existing view-change `useEffect` that clears `filterTagIds`/`filterMimeTypes`/`filterFolderIds`
- [x] 2.3 Derive filtered tag and folder lists (case-insensitive substring match on `name` against `filterTagSearch`/`filterFolderSearch`); when the search string is non-empty and the filtered list is empty, the derived list SHALL be empty regardless of selection state

## 3. Filter panel layout

- [x] 3.1 Reorder and restructure the "Add filter" `DropdownMenuContent` into three `InlineDivider`-separated sections in order: File type, Tags, Folder (Folder section gated by `filterSections.includes('folders')` as today)
- [x] 3.2 Replace the File type `DropdownMenuCheckboxItem` list with a `ToggleGroup type="multiple"` bound to `filterMimeTypes` (`value`/`onValueChange`), rendering one `Toggle` per `MIME_TYPE_FILTER_OPTIONS` entry
- [x] 3.3 Add a search `<input>` above the Tags `DropdownMenuCheckboxItem` list, wired to `filterTagSearch`/`setFilterTagSearch`, with `onKeyDown` calling `event.stopPropagation()`
- [x] 3.4 Render the Tags `DropdownMenuCheckboxItem` list from the filtered tag list (task 2.3), wrapped in a `max-h-* overflow-y-auto` container
- [x] 3.5 Repeat 3.3–3.4 for the Folder section (search input + filtered, scroll-capped checkbox list), when shown
- [x] 3.6 Verify visually that the per-section `max-h` does not produce nested scrollbars against the panel's own `max-h-(--available-height) overflow-y-auto`, adjusting the value if needed

## 4. Tests

- [x] 4.1 Update `AppLayout.test.tsx` filter-panel tests for the new section order (File type, Tags, Folder) and divider headers
- [x] 4.2 Update file-type filter tests to interact via the toggle controls instead of `DropdownMenuCheckboxItem`
- [x] 4.3 Add tests for: searching tags filters the list, searching folders filters the list, a search with no matches hides all items (including a checked one) while leaving it active in the chip row, and typing in a search input does not close the panel

## 5. Build & verify

- [x] 5.1 Run `npm run build` in `frontend/` and fix any type or build errors
