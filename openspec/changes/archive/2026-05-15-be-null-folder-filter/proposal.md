## Why

`GET /images` currently has no way to return only unfoldered images — omitting `folder_id` returns all images regardless of folder, and there is no supported filter for `WHERE folder_id IS NULL`. The frontend gallery view needs this to correctly populate the "Unsorted" view.

## What Changes

- `GET /images` gains a new `unfiled` boolean query parameter: `unfiled=true` filters for images where `folder_id IS NULL`; absent or `false` preserves existing behaviour
- `ListImagesParams` gains an `Unfiled bool` field; no changes to the existing `FolderID *uuid.UUID` field
- Repository `List` logic updated to emit `WHERE folder_id IS NULL` when `Unfiled = true`

## Capabilities

### New Capabilities

<!-- None -->

### Modified Capabilities

- `image-endpoints`: `GET /images` gains an `unfiled` query parameter for filtering unfoldered images

## Impact

- `internal/handler/image.go` — `ListImages` handler
- `internal/usecase/image_usecase.go` — `ListImagesParams` type
- `internal/repository/image_repository.go` — `List` method
- Handler and usecase unit tests
