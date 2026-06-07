## 1. Backend — repository layer

- [x] 1.1 Add `name *string` parameter to `ImageRepository.List` in `internal/usecase/image_repository.go`
- [x] 1.2 In `internal/repository/image_repository.go`, add a case-insensitive substring filter (`images.title ILIKE '%<name>%'`) to the cursor-paginated branch (`folderID == nil`) of `List`, composing with the existing `unfiled`/`tagID`/cursor conditions; leave the folder-view branch (`folderID != nil`) untouched so `name` has no effect there

## 2. Backend — usecase layer

- [x] 2.1 Add a `Name *string` field to `ListImagesParams` in `internal/usecase/image_pagination.go`
- [x] 2.2 Thread `params.Name` through `imageUsecase.ListImages` (`internal/usecase/image_usecase.go`) into the repository `List` call, treating a nil or empty-string value as "no filter"

## 3. Backend — handler & routing

- [x] 3.1 Parse an optional `name` query parameter in `ListImages` (`internal/handler/image.go`), trimming whitespace and treating an empty string as absent, and include it in the constructed `usecase.ListImagesParams`
- [x] 3.2 Update `bruno/images/list-images.bru` to document the new `name` query parameter

## 4. Backend — unit tests

- [x] 4.1 Add usecase test scenarios for `ListImages` covering: `Name` is passed through to the repository when non-empty, and the filter is skipped when `Name` is nil or an empty string (usecase layer only, no SQL repo tests, per CONVENTIONS.md)
- [x] 4.2 Add handler test scenarios for `ListImages` covering: `name` query param is parsed and forwarded, and an empty/whitespace-only `name` is treated as absent

## 4a. Backend — trash search support (scope addition)

Discovered mid-implementation: tasks 6.4/6.5 require the Trash view to query the backend with `name`, the same as All/Unsorted, but the original proposal/spec only added `name` to `GET /images`, not `GET /images/trash`. Extended the same pattern to the trash listing path so Trash search isn't a silent no-op.

- [x] 4a.1 Add `name *string` parameter to `TrashRepository.ListTrashed` (`internal/usecase/trash_repository.go`) and the cursor-paginated query in `internal/repository/image_repository.go` (`ILIKE '%<name>%'` on `images.title`)
- [x] 4a.2 Add a `Name *string` field to `ListTrashedParams` (`internal/usecase/image_pagination.go`) and thread it through `trashUsecase.ListTrashed` into the repository call, treating nil/empty as "no filter"
- [x] 4a.3 Parse an optional `name` query parameter in `TrashHandler.ListTrashed` (`internal/handler/trash.go`), trimming whitespace and treating empty as absent
- [x] 4a.4 Update `bruno/images/list-trash.bru` to document the new `name` query parameter
- [x] 4a.5 Add usecase and handler test scenarios for `ListTrashed` mirroring 4.1/4.2 (Name passthrough, blank name skipped/treated as absent)

## 5. Frontend — API client

- [x] 5.1 Add an optional `name` parameter to `getAllImages` and `getImages` in `frontend/src/lib/images.ts`, appended to the request's `URLSearchParams` when non-empty

## 6. Frontend — image search (ImageGrid)

- [x] 6.1 Add a small local debounce hook (e.g. `useDebouncedValue`) under `frontend/src/hooks/` — a `useEffect` + `setTimeout` implementation, no new dependency
- [x] 6.2 Add a search input to the gallery toolbar in `AppLayout.tsx`, in the same row as and to the left of the "Image" upload button (`AppLayout.tsx:228-253`), wired to local state in `ImageGrid`
- [x] 6.3 Add a `useEffect` keyed on `view` that clears the search term when the active view changes
- [x] 6.4 Branch the filtering behavior on `view.type`: for `folder` views, filter the already-fetched image array client-side by case-insensitive substring match on `title`; for `all`/`unsorted`/`trash` views, pass the debounced search term as `name` to `getAllImages`/`getImages`/`getTrashedImages`
- [x] 6.5 Include the debounced search term in `queryKeyFor` for `all`/`unsorted`/`trash` views so changing it triggers a fresh paginated query from page one
- [x] 6.6 Add `placeholderData: keepPreviousData` to the `useInfiniteQuery` configuration so the grid retains previous results while a new debounced search request is in flight

## 7. Frontend — folder filter (FolderSidebar)

- [x] 7.1 Add a "Filter folders…" input to the bottom section of `FolderSidebar` (`FolderSidebar.tsx:396-404`), above `ProfileMenu`, wired to local component state
- [x] 7.2 Implement tree-aware client-side filtering: a folder matches if its own name contains the term (case-insensitive) or any descendant matches; when a descendant matches, its ancestor chain is retained in the filtered tree for hierarchy context
- [x] 7.3 Clearing the filter input restores the full, unfiltered tree

## 8. Frontend — build check

- [x] 8.1 Run `npm run build` in `frontend/` and fix any type or build errors

## 9. Backend — lint check

- [x] 9.1 Run `golangci-lint` on the backend and fix any issues

## 10. Frontend — fix UI placement (rework)

The first implementation pass placed both new inputs incorrectly relative to the design: the image search box was built as its own row above the grid (spec calls for it in the toolbar row beside the "Image" upload button), and the folder filter was built above the folder tree near the top of the sidebar (spec calls for it in the bottom section above `ProfileMenu`). Resolving the repositioning surfaced a real conflict between the spec (input must render in `AppLayout`'s toolbar) and `design.md` decision 3 (search state must live locally in `ImageGrid`, no lifting/prop-drilling) — the user chose to lift `searchTerm`/`debouncedSearchTerm` state to `AppLayout` and pass it down to `ImageGrid` as props, since `AppLayout` already lifts and passes down comparable state (`view`, `selectedImage`, etc.). Tasks 6.2 and 7.1 are completed directly above; this group's items are folded into that work.

- [x] 10.1 Move the image search input from its standalone row above `ImageGrid` into the toolbar row in `AppLayout.tsx` (the `flex justify-between` row containing the "Image" upload button at `AppLayout.tsx:228-253`), positioned to the button's left — implemented by lifting `searchTerm`/`debouncedSearchTerm` state to `AppLayout` and passing it down to `ImageGrid` as props
- [x] 10.2 Move the "Filter folders…" input from above the folder tree to the bottom section of `FolderSidebar` (`FolderSidebar.tsx:396-404`), directly above `ProfileMenu`
- [x] 10.3 Re-run `npm run build` in `frontend/` and fix any issues introduced by the repositioning
