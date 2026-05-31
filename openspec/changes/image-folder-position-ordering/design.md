## Context

`image_folders.position` (TEXT) was introduced in migration 000010 as a placeholder using integer strings (`"1"`, `"2"`, …). It is never read — `GET /images` always orders by `created_at DESC`. The index `idx_image_folders_folder_position` already exists.

`SetImageFolder` generates positions using `MAX(position::int) + 1`, which breaks the moment any non-integer string is stored.

The goal is to activate this column with fracdex keys so folder views sort by user-defined order.

## Goals / Non-Goals

**Goals:**
- Folder views (`GET /images?folder_id=`) return images sorted by `image_folders.position ASC`
- New images appended to a folder receive a valid fracdex tail key
- A `PATCH /images/:id/position` endpoint lets callers write a new position (frontend computes the key)
- Existing integer positions are rebalanced to valid fracdex keys via a one-time migration script

**Non-Goals:**
- Frontend drag-and-drop UI (separate change)
- Position ordering for All, Unfiled, or Trash views
- Multi-folder membership (images have at most one folder; the data model allows more but the app enforces one)
- Server-side key computation on reorder

## Decisions

### 1. fracdex for key generation

**Chosen:** `github.com/rocicorp/fracdex` on the backend.

fracdex keys are ASCII strings that sort correctly under Postgres `TEXT` collation with no special configuration — `ORDER BY position ASC` just works. The library is byte-compatible with the JS `fractional-indexing` package, so frontend and backend produce identical key formats.

**Alternatives considered:**
- Integer positions with gap strategy (e.g. step 1000): simple but degrades and requires rebalancing after enough inserts between two items.
- ULID/timestamp-based ordering: not suitable for manual reordering since keys are immutable after creation.

### 2. Frontend computes reorder keys, backend stores them

**Chosen:** On drag-and-drop, the frontend uses the JS `fractional-indexing` library to compute `generateKeyBetween(beforeKey, afterKey)` from neighbour state it already holds, then sends `PATCH /images/:id/position { folder_id, position }`. The backend performs no computation — it validates the string is non-empty and writes it.

**Alternatives considered:**
- Backend computes from `{ before_id, after_id }`: requires the backend to fetch neighbour positions, adding a read per reorder. Unnecessary since the frontend already has the full sorted list.

### 3. Drop cursor pagination for folder views

**Chosen:** When `folder_id` is set, `GET /images` returns all images in one response (no `next_cursor`, no limit). The `ImageCursor` / limit parameters are ignored for this path.

The existing cursor encodes `(created_at, id)`, which is meaningless when sorting by `position`. Designing a position-based cursor adds complexity for a use case where collections are small (personal photo folders, not millions of items).

**Alternatives considered:**
- Position-based cursor (`position, id`): correct but adds a new cursor type, complicates the handler, and doesn't materially help given expected folder sizes.
- Offset pagination: simple but breaks with concurrent reorders.

### 4. One-time migration script (Go cmd) for rebalancing

**Chosen:** A standalone Go program at `cmd/migrate-positions/main.go` that connects to the database, iterates each folder's images ordered by their current `position::int`, generates fracdex keys sequentially using `fracdex.KeyBetween`, and updates each row. This runs once after the new code is deployed.

Fracdex key generation is Go code and cannot be expressed in a SQL migration. A separate script keeps the SQL migration files pure SQL and gives operators control over when the rebalance runs.

**Alternatives considered:**
- golang-migrate Go migration file: ties application-layer Go code into the migration pipeline; harder to reason about in production.
- Reset all positions to `''` in SQL then re-generate on first access: would require lazy generation logic in `SetImageFolder` and a fallback ordering strategy for the interim period.

### 5. `position` added to `GET /images` response for folder views only

**Chosen:** `imageResponse` gains a `Position *string` field (JSON `"position"`). It is populated from `ImageFolders[0].Position` when the image has a folder membership. For all/unfiled views the preloaded `ImageFolders` may be empty; in that case the field is `null` in the response.

The frontend needs position values to compute new keys around dropped items. Returning `null` for non-folder views avoids a separate response type while remaining correct.

### 6. New `UpdateImageFolderPosition` repository method

**Chosen:** A dedicated `UpdateImageFolderPosition(ctx, imageID uuid.UUID, folderID uuid.UUID, position string) error` is added to `ImageRepository`, separate from `SetImageFolder`. It issues a targeted `UPDATE image_folders SET position = ? WHERE image_id = ? AND folder_id = ?` and returns `ErrRecordNotFound` if the row does not exist (image not in that folder).

Reusing `SetImageFolder` for position updates would conflate two operations with different semantics and trigger unnecessary position re-computation.

## Risks / Trade-offs

- **Concurrent `SetImageFolder` tail collision** → Two simultaneous folder assignments may read the same `MAX(position)` and `fracdex.KeyBetween` will produce the same key for both. The upsert (`FirstOrCreate`) prevents a duplicate-key error, but one image's position silently overwrites the other's. Mitigation: accept the rare race (simultaneous folder assignments are unlikely in a single-user context); a future change can add advisory locking if needed.

- **fracdex key space growth** → Repeated inserts between two adjacent keys produce progressively longer strings. In practice this takes thousands of reorders between the same pair to become noticeable. Mitigation: the migration script can be re-run as a rebalance if key lengths grow problematic.

- **`position::int` cast in `SetImageFolder` breaks on fracdex strings** → The migration script must run before (or immediately after) any new images are added to folders post-deploy, to ensure no fracdex key is present when the old integer cast path could still run. Since the code change replaces the cast in the same deploy, this window is zero in practice.

- **No pagination for folder views** → A user with a very large folder (hundreds of images) receives the full payload in one response. Acceptable for the personal-library scale this product targets.

## Migration Plan

1. Deploy the new backend code (updated `SetImageFolder`, new list ordering, new endpoint).
2. Run `go run ./cmd/migrate-positions` against the production database. This script:
   - Selects all `(folder_id, image_id)` pairs ordered by `folder_id, position::int ASC`.
   - For each folder group, generates fracdex keys in sequence using `fracdex.KeyBetween`.
   - Batch-updates `image_folders.position`.
3. Verify: spot-check a folder via `GET /images?folder_id=<id>` and confirm images are ordered and `position` values are valid fracdex strings.

**Rollback:** The old code never read `position` for ordering, so rolling back to the previous binary leaves the data intact. Integer positions for any rows that existed before the migration are already gone, but the system was not using them anyway.

## Open Questions

- Should `GET /images/:id` (single image detail) also return `position`? Deferred — the detail view has no current need for it.
