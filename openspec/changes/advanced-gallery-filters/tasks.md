## 1. Scope the shared-contract change

- [ ] 1.1 Grep all call sites of `ImageRepository.List`, `ImageUsecase.ListImages`, and `ListImagesParams` across `internal/handler/`, `internal/usecase/`, `internal/repository/`, and their test files (including hand-written mocks like `mockImageRepository` in `image_usecase_test.go`) — produce the concrete list of files this change touches before editing, since the `List` signature and `ListImagesParams` struct are shared contracts with multiple consumers

## 2. Repository layer — relocate folder-view query, rewrite List filters

- [ ] 2.1 Add `ListByFolder(ctx, userID, folderID, sortField, direction) ([]*domain.Image, error)` to the `ImageRepository` interface (`internal/usecase/image_repository.go`) and implement it in `internal/repository/image_repository.go`, relocating the existing folder-view branch (`image_repository.go:38-58`) essentially unchanged — full fetch, `image_folders.position ASC` default ordering with explicit-sort override, `Preload("Tags")`/`Preload("ImageFolders")`, scoped to one folder and `userID`
- [ ] 2.2 Rewrite `List` in `internal/repository/image_repository.go`: remove the `folderID != nil` branch entirely; change signature to `List(ctx, userID, unfiled, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name, sortField, direction, cursor, limit)`; implement `folderIDs`/`tagIDs` as correlated `WHERE EXISTS (SELECT 1 FROM image_folders/image_tags WHERE image_id = images.id AND folder_id/tag_id IN (...))` subqueries (not JOINs, to avoid row duplication); implement `mimeTypes` as `WHERE images.mime_type IN (...)`; keep `unfiled`/`name`/sort/cursor logic as-is, composing via `AND`
- [ ] 2.3 Update the `ImageRepository` interface definition in `internal/usecase/image_repository.go` to reflect the new `List` signature and the addition of `ListByFolder`
- [ ] 2.4 Update `internal/repository/image_repository_integration_test.go`: adjust existing `List` test cases for the new signature (single-tag/folder cases become single-element multi-value cases); add cases for match-any with multiple values (tags, folders, mime types), the at-most-once-per-image guarantee when an image matches multiple supplied values, AND-composition across filter dimensions, and the `unfiled`+`folder_ids` contradiction yielding an empty result; add a new `ListByFolder` test covering position ordering, explicit-sort override, and folder scoping — per CONVENTIONS.md, integration tests only (no unit tests for the SQL repository)

## 3. Usecase layer — drop folder-view branch, add ListFolderImages

- [ ] 3.1 Update `ListImagesParams` in `internal/usecase/image_pagination.go`: remove `FolderID *uuid.UUID` and `TagID *uuid.UUID`; add `FolderIDs []uuid.UUID`, `TagIDs []uuid.UUID`, `MIMETypes []string`
- [ ] 3.2 Rewrite `ListImages` in `internal/usecase/image_usecase.go`: remove the `params.FolderID != nil` branch and its `FolderPosition` population loop; pass `FolderIDs`/`TagIDs`/`MIMETypes` straight through to `imageRepo.List`
- [ ] 3.3 Add `ListFolderImages(ctx, userID, folderID, sort, direction) ([]ImageItem, error)` to `ImageUsecase` (interface + implementation in `internal/usecase/image_usecase.go`): verify folder ownership (return not-found if missing/unowned), delegate to `imageRepo.ListByFolder`, populate `ImageItem.FolderPosition` from each image's matching `ImageFolders` entry (the logic relocated from the removed `ListImages` branch), and populate `ImageItem.ThumbnailURL` via the existing `thumbnailURL` helper
- [ ] 3.4 In `internal/usecase/image_usecase_test.go`: update `mockImageRepository` to the new `List` signature and add a `ListByFolder` method; update/replace `ListImages` test scenarios that referenced `FolderID`/`TagID` with scenarios for `FolderIDs`/`TagIDs`/`MIMETypes` passthrough and composition; add `ListFolderImages` scenarios — folder found and owned (asserts `FolderPosition`/`ThumbnailURL` on the returned items), and folder not found/not owned (asserts the specific not-found error, no repository list call made)

## 4. Handler layer — new query params, new endpoint

- [ ] 4.1 Rewrite the `ListImages` handler's query parsing in `internal/handler/image.go`: remove `folder_id`/`tag_id` parsing; add CSV parsing + per-element validation for `folder_ids`/`tag_ids` (UUIDs, `400` on any invalid element, empty segments ignored) and `mime_types` (non-empty strings); pass the resulting slices into `ListImagesParams`
- [ ] 4.2 Add a new handler method (e.g. `ListFolderImages`) for `GET /images/in-folder/:id`: parse/validate `:id` as UUID (`400` if invalid), parse/validate optional `sort`/`direction` against the existing allow-list, delegate to `imageUsecase.ListFolderImages`, map `404` for not-found, and return `200` with a plain JSON array of `imageResponse` items (no pagination envelope)
- [ ] 4.3 Register `GET /images/in-folder/:id` → the new handler method on the protected route group in `cmd/server/main.go`, alongside the other `/images/*` routes
- [ ] 4.4 In `internal/handler/image_test.go`: update `ListImages` test scenarios for the new param parsing/validation (valid CSV multi-value params, invalid-element-returns-400, empty-segment-ignored); add scenarios for the new folder-images handler — success with position-ordered items, explicit sort override, invalid folder UUID (`400`), folder not found/unowned (`404`), invalid sort/direction (`400`)

## 5. Bruno collection

- [ ] 5.1 Update `bruno/images/list-images.bru`: replace `folder_id`/`tag_id` query param entries with `folder_ids`/`tag_ids`/`mime_types` examples (comma-separated values)
- [ ] 5.2 Create `bruno/images/list-folder-images.bru` for `GET /images/in-folder/:id`, with example `:id` and optional `sort`/`direction` query params

## 6. Verification

- [ ] 6.1 Run `golangci-lint run` from the backend module and fix any issues raised
- [ ] 6.2 Run the full backend unit and integration test suites and confirm everything passes
