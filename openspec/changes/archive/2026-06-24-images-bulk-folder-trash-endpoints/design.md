## Context

Two existing single-image code paths are relevant:

- `SetImageFolder(ctx, imageID, folderID *uuid.UUID)` (`image_repository.go:105`) — inserts a row into `image_folders` with a fracdex position appended after the current max for that folder. It uses a plain `Create`, so it errors on a primary-key conflict (`(image_id, folder_id)` already exists). It's only ever called today from `MoveImageFolder`, where the destination folder differs from the source, so a conflict in practice doesn't occur.
- `SoftDelete(ctx, id, userID)` (`image_repository.go:325`) — deletes-with-`deleted_at` via GORM's default soft-delete scope (which already excludes rows where `deleted_at IS NOT NULL`), returning `gorm.ErrRecordNotFound` when `RowsAffected == 0`. This naturally covers "not found," "wrong owner," and "already trashed" as the same not-found outcome.

Neither path is bulk-aware, and `SetImageFolder`'s conflict-on-duplicate behavior is wrong for bulk add-to-folder, where re-adding an already-present image must succeed (per the locked product decision: idempotent membership counts as success).

## Goals / Non-Goals

**Goals:**
- One request adds N images to one folder, or trashes N images, with the backend owning iteration and per-item validation.
- A failure on one image (bad ID, wrong owner, already trashed) does not abort the rest of the batch.
- Response carries only a count; no per-item detail crosses the API boundary.

**Non-Goals:**
- Frontend consumption of these endpoints (separate proposal).
- Bulk remove-from-folder or bulk restore-from-trash (explicitly deferred to a future proposal).
- Batch size limits or pagination of the request body.
- A single round-trip bulk SQL statement (e.g. one `INSERT ... ON CONFLICT DO NOTHING` covering all rows) — deferred; see Decisions.

## Decisions

**New repository method instead of reusing `SetImageFolder`.** Add `AddImageToFolder(ctx, imageID, folderID uuid.UUID) error` using `clause.OnConflict{DoNothing: true}` on `(image_id, folder_id)`, distinct from `SetImageFolder` (errors on conflict) and `SyncImageFolders` (full diff/replace, wrong shape for "add without disturbing other memberships"). Alternative considered: catch the unique-constraint error from `SetImageFolder` and treat it as success in the usecase — rejected, since it relies on string/error-code matching against a driver-level constraint violation to mean "this was fine," which is more fragile than declaring the intent (`DO NOTHING`) at the query level.

**Per-image-ID loop, no single enclosing transaction across the whole batch.** Partial success requires that one image's failure not rollback another's success, which rules out wrapping the entire batch in one DB transaction. Each image's insert (add-to-folder) or soft-delete (trash) is its own independent statement. Alternative considered: one bulk `INSERT ... ON CONFLICT DO NOTHING ... VALUES (...), (...), ...` for add-to-folder — would need N fracdex positions precomputed in Go before the insert (each `KeyBetween` call depends on the previous one) and a separate pre-query to filter unowned/invalid IDs anyway, so it doesn't save a round trip over looping; deferred as a later optimization if batch sizes turn out large in practice.

**Ownership pre-filtering via one query, not N.** Before the per-item loop, run one query to get the subset of requested `image_ids` that exist, are owned by the caller, and (for trash) aren't already trashed: `SELECT id FROM images WHERE id IN (?) AND user_id = ?`. IDs not in the result are skipped + logged without ever reaching the mutation loop. This keeps the request at O(1) ownership queries instead of O(N).

**Folder validation is a precondition, not a per-item outcome.** `folder_id` is checked once up front via the existing folder `GetByID` (ownership + existence) before touching any image. An invalid/unowned folder fails the entire request with 404 — unlike per-image failures, there's no per-item interpretation of "wrong folder," so it isn't folded into the partial-success count.

**Trash reuses `SoftDelete` as-is**, looped per validated ID; no new repository method needed there.

**Response is `{"succeeded_count": n}`** where `n` counts every image_id that made it through validation and was processed — including idempotent no-op adds (image already in the folder), per the locked decision that idempotent membership counts as success.

## Risks / Trade-offs

- **N+1 queries for large selections** (one ownership pre-filter query, then up to N insert/delete statements) → could be slow for very large batches. Mitigation: none for v1 (no batch cap, per product decision) — acceptable since selections originate from manual UI multi-select, not programmatic bulk import. Revisit with a true bulk SQL statement if real usage shows large batches.
- **Silent total failure is indistinguishable from total success of zero items** — if every `image_id` in a request is invalid, the response is `{"succeeded_count": 0}`, identical in shape to a request that validly processed zero items. Accepted per the product decision that failures are log-only; the frontend already knows its own selection size and isn't relying on this response to detect that case.
- **Fracdex position races under concurrent bulk requests to the same folder** — `AddImageToFolder` recomputes `MAX(position)` per row in the loop, same pattern as the existing `SetImageFolder`. Not a new risk introduced by this change, just inherited at larger scale (more rows per request touching the same folder's position sequence).

## Migration Plan

No schema changes. Two new routes and usecase methods behind the existing auth middleware; deploy as a normal backend release. No rollback considerations beyond reverting the deploy — nothing reads these endpoints yet.
