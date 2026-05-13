## Why

`GET /images` and `GET /images/trash` currently return all images in a single unbounded query, which will degrade in performance as a user's library grows. Cursor-based pagination on `created_at` gives the frontend a predictable, stable way to load images incrementally across both endpoints.

## What Changes

- `GET /images` and `GET /images/trash` each accept two new optional query parameters: `limit` (default 50) and `cursor` (opaque string encoding the `created_at` + `id` of the last seen item)
- Both response shapes change from a plain array to an object containing `images` and `next_cursor`
- `next_cursor` is `null` when no further pages exist
- The repository `List` and `ListTrashed` methods gain pagination parameters
- The usecase `ListImages` and `ListTrashed` methods gain pagination parameters and return a result struct with the cursor

## Capabilities

### New Capabilities

- `image-list-pagination`: Cursor-based pagination for `GET /images` and `GET /images/trash` — query params, response envelope, repository-level keyset query, and cursor encoding/decoding logic

### Modified Capabilities

- `image-endpoints`: `GET /images` and `GET /images/trash` response shapes, and the `List` and `ListTrashed` repository method signatures are changing

## Impact

- `internal/handler/image.go` — `ListImages` and `ListTrashed` handlers read new query params, return new response envelope
- `internal/usecase/image_usecase.go` — `ListImages` and `ListTrashed` signatures and return types change; cursor encode/decode logic shared between both
- `internal/repository/image_repository.go` — `List` and `ListTrashed` SQL queries gain keyset `WHERE` clause and `LIMIT`
- `internal/usecase/interfaces.go` (or wherever `ImageRepository` / `ImageUsecase` interfaces live) — all four method signatures update
- Frontend clients consuming `GET /images` and `GET /images/trash` must update to the new response envelope (**BREAKING** response shape change on both endpoints)
