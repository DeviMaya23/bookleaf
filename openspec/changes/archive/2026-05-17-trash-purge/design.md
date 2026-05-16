## Context

Soft-deleted images are stored with `deleted_at` set to the deletion timestamp. The existing `ListTrashed` and `Restore` flows work off this column. Currently nothing ever removes these records or their R2 objects.

The `stale-upload-cleanup` change (same branch) already established:
- The goroutine + ticker pattern in `main.go`
- `DeleteObject` on `StorageService`
- The `CleanupStaleUploads` usecase method as a reference implementation

Trash purge follows the same pattern with two key differences: hard delete instead of soft delete, and needing to clean up both `r2_path` and `thumbnail_path`.

## Goals / Non-Goals

**Goals:**
- Permanently remove image DB records trashed more than 30 days ago
- Delete corresponding R2 objects (`r2_path` and `thumbnail_path` if present) before hard-deleting the record
- Run automatically every 24 hours in-process

**Non-Goals:**
- Notifying users before purge (out of scope; no email/notification infrastructure)
- Surfacing expiry dates in the trash API or FE (deferred)
- Configurable retention window (hardcoded to 30 days for now)
- Purging folders (folders have no soft delete)
- Recovering from partial purge runs (idempotency handled naturally — next run picks up any records missed)

## Decisions

### 1. Hard delete via `Unscoped()`, scoped to the record's own `UserID`

GORM's soft-delete means a plain `Delete` just sets `deleted_at`. To permanently remove a record we need `.Unscoped().Delete(...)`. Scoping by `UserID` is retained as a safety guard — the purge queries by `deleted_at` age, but the hard-delete step still filters by `id AND user_id`, matching the existing `SoftDelete` pattern.

### 2. R2 cleanup is best-effort; hard delete proceeds regardless

If the `DeleteObject` call for `r2_path` or `thumbnail_path` fails, the error is logged at `warn` level and the hard delete still proceeds. Blocking the hard delete on a storage failure would leave ghost records accumulating indefinitely, which is worse than an orphaned R2 object. Orphaned objects can be reconciled separately if needed.

**Alternative considered:** Atomic cleanup — only hard-delete if R2 delete succeeds. Rejected: a transient R2 error would prevent permanent removal of records, defeating the purpose of the purge.

### 3. 24-hour ticker interval

Trash purge doesn't need sub-hour precision. A 24-hour interval keeps DB query load minimal. The 30-day threshold means a record expiring at T+30d will be caught within 24 hours of crossing the threshold.

**Alternative considered:** Running at the same 10-minute interval as stale cleanup. Rejected: unnecessary frequency for data with a 30-day retention window.

### 4. Separate goroutine from stale-upload cleanup

Two distinct goroutines with different tick intervals (10 min vs 24 hr) are cleaner than a shared goroutine trying to schedule both. Each has an independent concern and lifecycle.

### 5. `thumbnail_path` deletion is conditional

`thumbnail_path` is nullable — not all images have a thumbnail (e.g., if thumbnail generation failed at upload time). The purge skips the R2 delete for `thumbnail_path` when it is nil, avoiding a spurious delete call.

## Risks / Trade-offs

- **Irreversibility**: Hard deletes cannot be undone. Mitigation: the 30-day window is generous; records are only purged well past the point where restoration would be expected.
- **R2 orphans on hard-delete failure**: If the DB hard-delete succeeds but R2 delete failed, the R2 object remains but the record is gone — no way to retry. Mitigation: log at warn, accept small leak; reconcile manually if R2 costs become significant.
- **No graceful shutdown**: Like stale cleanup, this goroutine doesn't stop on SIGTERM. Mid-purge kill leaves some records un-purged; they'll be caught on next run. No corrupt state risk since hard deletes are atomic per record.
- **Clock skew**: Negligible — 30-day threshold with 24-hour check interval means worst-case slip is under 24 hours.

## Migration Plan

No schema changes required. Deploy the application update; the goroutine starts automatically and begins purging on the first 24-hour tick.

**Rollback**: Remove the goroutine and `PurgeExpiredTrash` call. Records that were already hard-deleted cannot be recovered, but no new deletions will occur.
