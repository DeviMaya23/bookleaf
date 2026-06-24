## 1. Backend — repository layer

- [x] 1.1 In `backend/internal/usecase/trash_repository.go`, update the `TrashRepository.ListTrashed` interface signature to `ListTrashed(ctx context.Context, userID string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)`.
- [x] 1.2 In `backend/internal/repository/image_repository.go`, update `ListTrashed`'s implementation to accept `sortField`/`direction`, call the existing `usecase.ResolveSort(sortField, direction)` (no changes to `ResolveSort` itself), replace the hardcoded `Order("deleted_at ASC, id ASC")` with `Order(dispatch.OrderClause)`, and branch the keyset `WHERE` two ways on `dispatch.Column` (`title` vs `created_at`) exactly as `List` already does — no `deleted_at` branch.
- [x] 1.3 Update `backend/internal/repository/image_repository_integration_test.go`'s `ListTrashed` integration tests: replace any assertion of `deleted_at ASC` default ordering with `created_at DESC` default; add cases for explicit `sortField="title"` and explicit `sortField="created_at"` with both directions, asserting correct ordering and keyset pagination across pages (mirroring the existing `List` sort integration tests, if present, for structure).

## 2. Backend — usecase layer

- [x] 2.1 In `backend/internal/usecase/image_pagination.go`, add `Sort *string` and `Direction *string` fields to `ListTrashedParams`.
- [x] 2.2 In `backend/internal/usecase/trash_usecase.go`, update `ListTrashed` to pass `params.Sort`/`params.Direction` straight through to `imageRepo.ListTrashed` (or equivalent trash repo field), with no validation or defaulting in the usecase — mirroring `ImageUsecase.ListImages`.
- [x] 2.3 In `backend/internal/usecase/trash_usecase_test.go`, add `TestTrashUsecase_ListTrashed_PassesSortAndDirection` and `TestTrashUsecase_ListTrashed_PassesNilSortAndDirection`, mirroring `TestImageUsecase_ListImages_PassesSortAndDirection`/`PassesNilSortAndDirection` in `image_usecase_test.go`.

## 3. Backend — handler layer

- [x] 3.1 In `backend/internal/handler/trash.go`, update `ListTrashed` to parse and validate optional `sort` (`created_at`/`title`) and `direction` (`asc`/`desc`) query parameters, returning `400 Bad Request` on invalid values, defaulting `direction` via `usecase.ResolveSort(sortField, nil).DefaultDirection` when `sort` is present but `direction` is omitted — mirroring `ImageHandler.ListImages` in `handler/image.go` exactly. Pass the validated values into `usecase.ListTrashedParams.Sort`/`.Direction`.
- [x] 3.2 In `backend/internal/handler/trash_test.go`, add `TestTrashHandler_ListTrashed_SortAndDirection` (explicit sort+direction passed through) and `TestTrashHandler_ListTrashed_InvalidSortOrDirection` (400 on bad `sort`/`direction` values), mirroring `TestImageHandler_ListImages_SortAndDirection`/`TestImageHandler_ListFolderImages_InvalidSortOrDirection` in `handler/image_test.go`.
- [x] 3.3 Update existing `TestTrashHandler_ListTrashed_ReturnsList` (and any other existing `ListTrashed` handler test asserting ordering) if it has assumptions tied to the old `deleted_at ASC` default. (Verified: no existing test asserted the old default; no changes needed.)

## 4. Frontend — request wiring

- [x] 4.1 In `frontend/src/lib/images.ts`, add optional `sort?: 'created_at' | 'title'` and `direction?: 'asc' | 'desc'` parameters to `getTrashedImages`, and set them as `sort`/`direction` query params when present — mirroring `getImages`/`getAllImages`.
- [x] 4.2 In `frontend/src/features/gallery/hooks/useGalleryImages.ts`, update the `case 'trash':` branch in `fetcherFor` to compute `{ sort, direction }` via `sortParamsFor(sortBy, sortDir)` and pass them to `getTrashedImages`, same as the other view cases. (Also updated the existing `useGalleryImages.test.tsx` assertion that expected the old no-sort-params call shape.)
- [x] 4.3 Confirm (no code change expected, verify only) that `ImageGrid`'s `useInfiniteQuery` query key for the Trash view already includes `sortBy`/`sortDir` so that changing sort while viewing Trash triggers a re-fetch — add it to the query key if it's currently missing for the trash case. (Verified: `queryKeyFor`'s `case 'trash':` in `useGalleryImages.ts` already includes `sortKey`; no change needed.)

## 5. Bruno

- [x] 5.1 Update `bruno/images/list-trash.bru` to add `~sort` and `~direction` optional query params (commented/disabled by default, matching the `~cursor`/`~limit`/`~name` convention already used in that file), demonstrating valid usage.

## 6. Verification

- [x] 6.1 Run `golangci-lint run` from `backend/` and fix any issues introduced by this change. (0 issues.)
- [x] 6.2 Run `npm run build` and `npm run lint` from `frontend/` and fix any issues introduced by this change. (Both clean.)
- [x] 6.3 ~~Manually verify in a browser: open Trash view, confirm default order is newest-created-first; switch sort to `Name`/`Date added` with both directions and confirm the list re-orders and infinite-scroll pagination continues to work correctly across each sort mode.~~ Superseded by 7.12 — manual verification caught that "Date added" should have been "Date deleted"; see section 7.

## 7. Revision — correct "Date added" to "Date deleted" for Trash

Manual verification (task 6.3) found that Trash's date sort, wired in section 1-6 to `created_at` (reusing `List`'s default field unmodified), is semantically wrong: a trash listing should sort by when each item was *deleted*, not when it was *created*. This section reinstates `deleted_at` as a real, explicit sort field for `GET /images/trash` only (`GET /images` is unaffected).

- [x] 7.1 In `backend/internal/usecase/image_pagination.go`, add a `"deleted_at"` case to `ResolveSort` (default direction `desc`, newest-deleted-first), alongside the existing unchanged `"title"` and default `"created_at"` cases. (Also added `deleted_at` cases to `TestResolveSort`.)
- [x] 7.2 In `backend/internal/handler/trash.go`, change `ListTrashed`'s sort allow-list to `{deleted_at, created_at, title}`, and when the `sort` query param is absent, explicitly default the sort field to `"deleted_at"` (not left nil) before the existing direction-defaulting logic runs.
- [x] 7.3 In `backend/internal/repository/image_repository.go`, extend `ListTrashed`'s keyset `WHERE` branch from two-way (`title`/`created_at`) to three-way (`title`/`deleted_at`/`created_at`), using `cursor.DeletedAt` for the `deleted_at` branch.
- [x] 7.4 In `backend/internal/usecase/trash_usecase.go`, extend `ListTrashed`'s next-cursor construction to a three-way branch on `dispatch.Column`, populating `Title`, `DeletedAt`, or leaving the always-set `CreatedAt` as-is.
- [x] 7.5 Update/add backend tests: integration tests in `image_repository_integration_test.go` for `deleted_at` sort ordering/pagination (added to `TestImageRepository_ListTrashed_Pagination_SortAware` and `listAllTrashedPages`'s cursor branching) and the new default; handler tests in `trash_test.go` for the `deleted_at` default-substitution and updated allow-list (`created_at` still valid, new invalid values still 400).
- [x] 7.6 In `frontend/src/features/gallery/hooks/useGalleryControls.ts`: widen `SortBy` to include `'deleted_at'`; add a `deleted_at` entry to `FIELD_DEFAULT_DIRECTION` (`desc`); change Trash's default in `defaultSortForViewType` to `{ sortBy: 'deleted_at', sortDir: 'desc' }`; change Trash's `sortFieldOptions` to `['deleted_at', 'title']`. (`defaultSortForViewType` widened to take `AppView['type']` instead of `isFolder: boolean`, since it now needs to distinguish trash from all/unsorted too.)
- [x] 7.7 In `frontend/src/features/gallery/components/GalleryToolbar.tsx`: add `deleted_at: 'Date deleted'` to `SORT_FIELD_LABELS`; widen `DIR_LABELS` to include `deleted_at` direction labels (e.g. "Oldest deleted first"/"Newest deleted first").
- [x] 7.8 In `frontend/src/lib/images.ts`, change `getTrashedImages`'s `sort` parameter type from `'created_at' | 'title'` to `'deleted_at' | 'title'`. (Also widened `sortParamsFor`'s return type in `useGalleryImages.ts` and added narrowing type assertions at each `fetcherFor` call site, since the shared helper now returns a 3-way union no single caller fully accepts.)
- [x] 7.9 Update frontend tests in `useGalleryControls`/`GalleryToolbar`/`useGalleryImages` test files for the new default field/direction, options, and label. (Added new trash-specific test cases — none existed before, which is why the original mislabeling bug wasn't caught by tests.)
- [x] 7.10 Update `bruno/images/list-trash.bru`'s `~sort` example to a valid Trash value (`deleted_at` or `title`).
- [x] 7.11 Re-run `golangci-lint run` (backend) and `npm run build`/`npm run lint` (frontend); fix any issues. (Both clean; full BE `go test ./...` and FE `vitest run` — 351 tests — also passing.)
- [x] 7.12 Manually verify in a browser: open Trash view, confirm default order is newest-deleted-first; switch sort to `Name`/`Date deleted` with both directions and confirm the list re-orders and infinite-scroll pagination continues to work correctly across each sort mode.
