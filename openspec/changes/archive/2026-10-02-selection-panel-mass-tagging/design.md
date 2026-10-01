## Context

The selection panel currently supports bulk folder assignment (`POST /images/bulk/add-to-folder`) and bulk trash (`POST /images/bulk/trash`). Both BE endpoints exist, but `BulkTrash` loops N individual `SoftDelete` calls at the DB level. Tags have no bulk path at all — users must open each image individually.

Two things change here:
1. The `BulkTrash` usecase replaces its per-image loop with a single atomic `UPDATE`.
2. A new `POST /images/bulk/tag` endpoint enables appending tags to many images at once, and the selection panel UI grows a tag section wired to it.

## Goals / Non-Goals

**Goals:**
- Make `BulkTrash` issue one DB statement instead of N
- New `POST /images/bulk/tag` that appends tags atomically across N images
- Selection panel Apply fires folder + tag bulk requests concurrently in one action

**Non-Goals:**
- `EmptyTrash` R2 job enqueueing is not changed (kept as N×`EnqueueR2Delete` loops)
- `BulkAddToFolder` fracdex loop is not changed
- No bulk tag removal — append-only for now
- No tag creation from the selection panel (tags must already exist or be created via the tag input's free-text path before Apply)

## Decisions

### 1. `BulkSoftDelete` replaces the loop in `BulkTrash`

`FilterOwnedImageIDs` still runs first to strip unowned IDs — this is unchanged. Then instead of looping `SoftDelete` per ID, a single `BulkSoftDelete(ctx, ownedIDs, userID)` call issues:

```sql
UPDATE images SET deleted_at = NOW()
WHERE id IN (?) AND user_id = ? AND deleted_at IS NULL
```

`BulkSoftDelete` is added to the `TrashRepository` interface. `succeeded_count` is derived from `RowsAffected`.

**Why not keep the loop:** the loop is not atomic and requires N round-trips. The ownership filter already runs in one query, so the subsequent update can safely use `IN (ownedIDs)`.

**The `deleted_at IS NULL` guard:** images already trashed are silently skipped (same best-effort semantics as before) without needing per-row error handling.

### 2. `AppendImageTagsBulk` on `TagRepository`

Adds to the `TagRepository` interface:

```go
AppendImageTagsBulk(ctx context.Context, imageIDs []uuid.UUID, tagIDs []uuid.UUID) error
```

Implementation: builds the cartesian product of `imageIDs × tagIDs` as `(image_id, tag_id)` rows, bulk-inserts in one transaction with `ON CONFLICT DO NOTHING`. This makes the operation idempotent — re-tagging already-tagged images is a no-op, not an error.

**Why append, not replace:** bulk replace would wipe existing tags on N images at once — destructive and almost certainly wrong. Append is the safe default for bulk operations where the full tag state of each image is unknown.

**Deduplication:** `TagInput` already prevents adding the same tag chip twice (checked by name in `createFromText`). The backend `ON CONFLICT DO NOTHING` handles any duplicates within the `image_ids` list itself.

### 3. Tag resolution happens before the bulk tag request

The existing single-image flow calls `resolveOrCreateTags` on the FE before sending tag IDs — this creates new tags server-side and returns their IDs. The same pattern applies here: when the user clicks Apply, the panel first calls `resolveOrCreateTags` to materialise any free-text tags, then fires `POST /images/bulk/tag` with the resolved IDs.

**Why not server-side tag creation in the bulk endpoint:** keeping the endpoint to ID-only input avoids mixed create+assign semantics and matches the existing `UpdateImage` contract.

### 4. `onApply({ folderIds, tags })` replaces `onAddToFolder`

The current `SelectionPanelBody` calls `onAddToFolder(folderId)` once per pending folder in a loop. This is replaced with a single `onApply({ folderIds: string[], tags: Tag[] })` prop. Full `Tag` objects are passed (not just IDs) so that `AppLayout` can call `resolveOrCreateTags` on free-text entries before firing the bulk request.

Apply fires all operations concurrently via `Promise.allSettled`:

```ts
const results = await Promise.allSettled(ops)
const allSucceeded = results.every((r) => r.status === 'fulfilled')
```

`Promise.allSettled` is used instead of `Promise.all` so that all operations complete before error handling — a failure in one does not suppress error toasts for others.

Note: `bulkAddImagesToFolder` takes one `folder_id` — if multiple folders are selected, they are fired as separate concurrent requests. Tags are a single request covering all resolved tag IDs.

**Why not a combined single endpoint:** a combined `POST /images/bulk/update` would couple two independent operations. Two concurrent requests keep the semantics clean, errors attributable per operation, and avoid a new endpoint with a broader surface area.

### 5. Apply button is disabled until either folders or tags are pending

Apply is disabled when `pendingFolders.length === 0 && pendingTags.length === 0`. Either alone is sufficient to enable it.

## Risks / Trade-offs

- **Partial Apply failure**: if the folder request succeeds but the tag request fails (or vice versa), the user sees an error toast for the failed operation but the successful one is already committed. This is acceptable — operations are independent, and a retry will be idempotent.
- **Large cartesian products**: `AppendImageTagsBulk` with 500 images × 20 tags = 10,000 rows in one insert. GORM's bulk create will chunk this into multiple statements under the hood, but no explicit limit is enforced at the handler layer for now. Acceptable given typical selection sizes.
- **`resolveOrCreateTags` latency on Apply**: if the user types several new tag names, Apply blocks briefly on N tag-create requests before firing the bulk tag request. This matches the existing single-image behaviour and is acceptable.

## Migration Plan

No schema changes. No migration needed — `image_tags` table and `ON CONFLICT DO NOTHING` work on the existing schema.

Deploy is a standard rolling deploy. The new endpoint is additive; no existing call sites change.
