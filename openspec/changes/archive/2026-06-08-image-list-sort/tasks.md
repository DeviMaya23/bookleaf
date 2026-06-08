## 1. Usecase-layer types, sort dispatch, and interface signature

- [x] 1.1 In `internal/usecase/image_pagination.go`, add `Title *string` to `ImageCursor` and `cursorPayload`, and update `EncodeCursor`/`DecodeCursor` so a title-sort cursor round-trips correctly (populated only when the active sort orders by `title`)
- [x] 1.2 Add `Sort *string` and `Direction *string` fields to `ListImagesParams`
- [x] 1.3 Add a small sort-dispatch helper in `image_pagination.go` that maps an allow-listed `(field, direction)` pair to its comparison column, `ORDER BY`/keyset operator (`>` for asc, `<` for desc), and resolves the field's default direction (`created_at` → `desc`, `title` → `asc`) when direction is unspecified
- [x] 1.4 Update the `List` method signature in the `ImageRepository` interface (`internal/usecase/image_repository.go`) to add `sortField *string, direction *string`

## 2. Repository — sort-aware list query

- [x] 2.1 In `imageRepository.List` (folder branch, `internal/repository/image_repository.go`), honor an explicit `sortField` by swapping the `Order()` clause via the dispatch helper, falling back to today's `image_folders.position ASC` when `sortField` is nil; cursor/limit continue to be ignored
- [x] 2.2 In `imageRepository.List` (non-folder branch), replace the hardcoded `Order("images.created_at DESC, images.id DESC")` and keyset `Where(...)` with dispatch-driven construction that selects the column(s), `ORDER BY` direction, and comparison operator from the active sort/direction, defaulting to today's `created_at DESC` behaviour when `sortField` is nil; read the cursor field matching the active sort column (`cursor.CreatedAt` vs `cursor.Title`)

## 3. Usecase — thread sort params through

- [x] 3.1 In `imageUsecase.ListImages`, pass `params.Sort`/`params.Direction` straight through to `imageRepo.List` as `sortField`/`direction` with no additional validation or defaulting
- [x] 3.2 Update the `NextCursor` construction in `ListImages` to populate `ImageCursor.Title` (instead of/alongside `CreatedAt`) when the active sort orders by `title`, mirroring how `DeletedAt` is conditionally populated for trash cursors

## 4. Handler — query parameter parsing and validation

- [x] 4.1 In `ImageHandler.ListImages`, parse the `sort` query parameter and validate it against the allow-list (`created_at`, `title`); return `400 Bad Request` for any other non-empty value
- [x] 4.2 Parse the `direction` query parameter; validate it against `{asc, desc}` and return `400 Bad Request` on an invalid non-empty value — but only when `sort` is present and valid; when `sort` is absent or empty, accept and ignore `direction` without validating it
- [x] 4.3 When `sort` is present and valid but `direction` is absent or empty, resolve the field's default direction via the dispatch helper from task 1.3
- [x] 4.4 Pass the resolved `sort`/`direction` values (or `nil` when omitted) into `ListImagesParams`

## 5. Bruno

- [x] 5.1 Add commented example `sort`/`direction` query params to `bruno/images/list-images.bru`, mirroring the existing `~name`/`~tag_id` examples

## 6. Unit tests — usecase and handler layers

- [x] 6.1 Usecase: `ListImages` passes non-nil `Sort`/`Direction` through to `imageRepo.List` unchanged
- [x] 6.2 Usecase: `ListImages` passes nil `Sort`/`Direction` through as nil, preserving the view's default ordering
- [x] 6.3 Usecase: the sort-dispatch helper resolves the correct column, comparison operator, and default direction for each allow-listed `(field, direction)` combination, including when direction is unspecified
- [x] 6.4 Handler: `GET /images?sort=title&direction=desc` (and other valid combinations) are parsed and forwarded to the usecase correctly
- [x] 6.5 Handler: an invalid `sort` value returns `400`; an invalid `direction` value (with `sort` present) returns `400`, asserting the specific error response
- [x] 6.6 Handler: `direction` supplied without `sort` returns `200` with the view's default ordering unaffected
- [x] 6.7 Handler: omitting both `sort` and `direction` preserves existing behaviour (regression coverage for the "zero FE impact" guarantee)

## 7. Integration tests — repository

- [x] 7.1 Non-folder branch: keyset pagination returns correctly ordered, non-overlapping pages for each of `title asc`, `title desc`, `created_at asc`, `created_at desc`
- [x] 7.2 Non-folder branch: rows sharing the same value in the active sort column are tiebroken by `id` with no duplicates or gaps across a page boundary
- [x] 7.3 Folder branch: an explicit `sortField` overrides `image_folders.position ASC`; omitting it preserves position-based ordering and ignores cursor/limit as before

## 8. Verification

- [x] 8.1 Run `golangci-lint` and fix any issues it reports
