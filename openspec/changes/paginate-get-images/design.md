## Context

`GET /images` and `GET /images/trash` both call unbounded repository methods (`List` and `ListTrashed`) ordered by `created_at DESC`. As a user's image library grows these queries become expensive and the payloads become large. The handler, usecase, and repository all need to change together; the response envelope is a breaking shape change for frontend consumers on both endpoints.

The sort order is already `created_at DESC`, so a keyset (cursor) strategy on `(created_at, id)` is a natural fit — it is stable, index-friendly, and avoids the page-drift problem of offset pagination. The same cursor type and encode/decode logic is shared between both endpoints.

## Goals / Non-Goals

**Goals:**
- Add cursor-based pagination to `GET /images` and `GET /images/trash` with a configurable `limit` (default 50)
- Keep the cursor opaque to clients (base64-encoded JSON internally)
- Return `next_cursor: null` when the caller has reached the last page
- Share cursor encoding/decoding logic between both endpoints
- Maintain the existing `folder_id` filter on `GET /images`

**Non-Goals:**
- Backwards-compatible plain-array response — the envelope change is accepted as breaking on both endpoints
- Server-side cursor validation beyond basic decode failure (invalid cursors return 400)

## Decisions

### Cursor encoding: base64(JSON{created_at, id})

The cursor encodes the `created_at` timestamp and `id` of the last item on the current page. On the next request the repository applies a keyset `WHERE` clause:

```sql
-- GET /images (List)
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (created_at, id) < ($cursor_created_at, $cursor_id)
ORDER BY created_at DESC, id DESC
LIMIT $limit + 1

-- GET /images/trash (ListTrashed)
WHERE user_id = $1
  AND deleted_at IS NOT NULL
  AND (created_at, id) < ($cursor_created_at, $cursor_id)
ORDER BY created_at DESC, id DESC
LIMIT $limit + 1
```

Fetching `limit + 1` rows lets us detect whether a next page exists without a separate `COUNT` query — if we get more than `limit` rows back, there is a next page and we use row `limit` as the cursor seed; otherwise `next_cursor` is `null`.

**Alternative considered — offset pagination**: Simple to implement but suffers from page drift (inserts shift rows between pages) and forces a full table scan for large offsets. Rejected.

**Alternative considered — integer sequence cursor**: A monotonic sequence column would be simpler, but images don't have one. Using `(created_at, id)` reuses existing indexed columns.

### Response envelope

Both endpoints change from a plain array to:

```json
{
  "images": [...],
  "next_cursor": "<opaque string | null>"
}
```

This is a **breaking change** on both endpoints. The frontend must be updated alongside this backend change.

### Limit capping

`limit` is accepted as a query parameter but capped at 200 server-side to prevent abuse. Requests with `limit > 200` are silently clamped (not rejected), keeping the API lenient. The same cap applies to both endpoints.

### Cursor decode failure

If the `cursor` query param cannot be base64-decoded or parsed as JSON, the handler returns `400 Bad Request` on both endpoints. This prevents silent wrong-page results.

### Interface changes

`ImageRepository.List`, `ImageRepository.ListTrashed`, `ImageUsecase.ListImages`, and `ImageUsecase.ListTrashed` signatures all change. The usecase gains shared param/result structs. `ListTrashed` reuses `ImageCursor` and gets its own `ListTrashedParams` / `ListTrashedResult` for symmetry and extensibility.

```go
type ImageCursor struct {
    CreatedAt time.Time
    ID        uuid.UUID
}

type ListImagesParams struct {
    FolderID *uuid.UUID
    Cursor   *ImageCursor  // nil = first page
    Limit    int
}

type ListImagesResult struct {
    Images     []*domain.Image
    NextCursor *ImageCursor  // nil = no more pages
}

type ListTrashedParams struct {
    Cursor *ImageCursor  // nil = first page
    Limit  int
}

type ListTrashedResult struct {
    Images     []*domain.Image
    NextCursor *ImageCursor  // nil = no more pages
}
```

## Risks / Trade-offs

- **Breaking response shape on two endpoints** → Frontend must ship its update for both simultaneously or the app breaks. Coordinate deployment.
- **Ties on created_at** → Two images with identical `created_at` could cause cursor instability. Including `id` in the keyset (secondary sort + cursor field) eliminates this.
- **No composite index on (user_id, created_at DESC, id DESC)** → The keyset query will do a filtered index scan. A composite index would improve performance at scale, but adding a migration is outside this change's scope. The existing index on `user_id` is sufficient for current data volumes.
- **Cursor expiry** → Cursors are stateless and never expire; a stale cursor from a long-lived session will simply return results relative to the encoded position, which is acceptable behaviour.
- **Trash cursor sorts by created_at not deleted_at** → Trash is sorted by `created_at DESC` (consistent with the active list), not `deleted_at`. This is intentional — it keeps the cursor type shared — but means trash is not ordered by when items were deleted.

## Migration Plan

1. Deploy backend with new paginated responses on `GET /images` and `GET /images/trash`
2. Update frontend to consume the new envelope on both endpoints
3. No database migration required for this change
4. Rollback: revert both frontend and backend together (breaking changes means they are coupled)
