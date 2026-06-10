## Why

The gallery "All" view can currently only be filtered by name and a single tag. Users need to narrow the gallery by multiple tags (match-any), multiple MIME types, and multiple folders at once. The current `GET /images` endpoint can't cleanly grow a folder filter, though: `folder_id` is already overloaded there as a mode-switch — its presence replaces the paginated/filtered gallery query with an entirely different "fetch this folder's contents in drag-and-drop order" query (full fetch, no pagination, ordered by `image_folders.position`, returns per-image `FolderPosition`). Those are two structurally different queries — one lists/filters images, the other traverses a folder's ordered membership — that happen to share a route. Before `folder_id` can become an ordinary filter, the "folder contents" query needs an honest home of its own.

## What Changes

- New endpoint serving a single folder's images in their custom drag-and-drop order (full fetch, ordered by `image_folders.position`, with per-image `FolderPosition`) — this is the "folder contents" query extracted out of `GET /images`. It stays in the image domain (image handler/usecase/repository), since `Image` is the resource being returned and `position` is owned by the `image_folders` join, not by `Folder` alone.
- **BREAKING**: `GET /images` no longer treats `folder_id` as a mode-switch. The folder-view branch (full fetch, position ordering, `FolderPosition`, no pagination) is removed from `ListImages`/`ImageRepository.List` entirely.
- `GET /images` gains multi-value "match any" filters:
  - `folder_ids` — replaces `folder_id`; matches images that belong to ANY of the given folders
  - `tag_ids` — replaces the existing single-value `tag_id`; matches images tagged with ANY of the given tags
  - `mime_types` — new; matches images whose MIME type is ANY of the given values
  - All filters compose with each other, with `name`, and with existing pagination/sorting
- Internal: multi-value filters are implemented via `WHERE EXISTS (...)` subqueries rather than `JOIN ... IN (...)`, so an image matching multiple values in a filter is not returned more than once — protecting the keyset cursor pagination's `limit + 1` probe and `(column, id)` comparison from row-duplication corruption.

## Capabilities

### New Capabilities
- `folder-image-listing`: dedicated endpoint and usecase/repository support for fetching a single folder's images in custom drag-and-drop order, with per-image folder position — the "board/reorder view" query extracted from the gallery list.

### Modified Capabilities
- `image-endpoints`: `GET /images` handler, `ListImages` usecase, and `ImageRepository.List` — `folder_id`/`tag_id` single-value params are replaced with `folder_ids`/`tag_ids`/`mime_types` multi-value match-any filters; the folder-view (mode-switch) branch is removed.

## Impact

- **Backend code**: `internal/handler/image.go` (`ListImages`), `internal/usecase/image_usecase.go` (`ListImages`), `internal/usecase/image_pagination.go` (`ListImagesParams`), `internal/repository/image_repository.go` (`List`), `internal/usecase/image_repository.go` (interface), `cmd/server/main.go` (new route registration), plus a new handler/usecase/repository method for the folder-contents query and its Bruno collection file.
- **Frontend**: out of scope for this proposal (BE-first; FE wiring is a follow-up phase), but the `folder_id` → `folder_ids` change is breaking — the existing folder-view fetch path in `ImageGrid.tsx` will need to be repointed to the new endpoint before/alongside deployment, or it will silently receive gallery-style paginated results instead of full ordered folder contents.
- **API contract**: `GET /images` query parameters change shape (`folder_id`/`tag_id` removed, `folder_ids`/`tag_ids`/`mime_types` added); a new route is added for folder contents.
