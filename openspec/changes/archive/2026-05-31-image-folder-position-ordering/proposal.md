## Why

Images assigned to a folder have a `position` column on `image_folders`, but it is never used for ordering — the list endpoint always sorts by `created_at DESC`. This change activates position-based ordering for folder views using fractional indexing (fracdex), enabling drag-and-drop reordering in a future frontend change.

## What Changes

- Add `fracdex` as a Go dependency for generating fractional index keys
- Add a one-time DB migration that rebalances existing integer placeholder positions to valid fracdex keys
- `SetImageFolder`: replace `MAX(position::int)` integer arithmetic with `fracdex.KeyBetween(lastPosition, "")` to append new images at the tail of their folder
- `GET /images?folder_id=<id>`: sort by `image_folders.position ASC`, drop cursor/limit (return all images in the folder), include `position` in each image response
- `GET /images` (all/unfiled views): no change to ordering or pagination
- **New endpoint** `PATCH /images/:id/position`: accepts `{ folder_id, position }`, writes the fracdex key to `image_folders` — no server-side computation, frontend supplies the key

## Capabilities

### New Capabilities

- `image-position-reorder`: PATCH /images/:id/position endpoint — stores a caller-supplied fracdex position string for an image within a specific folder

### Modified Capabilities

- `image-folders`: `SetImageFolder` position assignment changes from integer increment to fracdex tail key; existing spec covers the method contract
- `image-endpoints`: GET /images folder-view response shape changes (adds `position` field, drops pagination); new PATCH endpoint added
- `image-list-pagination`: folder views are excluded from cursor-based pagination; all/unfiled/trash pagination unchanged

## Impact

- **Backend**: `internal/repository/image_repository.go`, `internal/usecase/image_usecase.go`, `internal/handler/image.go`, `cmd/server/main.go`
- **DB migrations**: new `golang-migrate` migration file for fracdex rebalance
- **New Go dependency**: `github.com/rocicorp/fracdex`
- **API contract**: `GET /images?folder_id=` response adds `position` field and removes `next_cursor`; new `PATCH /images/:id/position` route
