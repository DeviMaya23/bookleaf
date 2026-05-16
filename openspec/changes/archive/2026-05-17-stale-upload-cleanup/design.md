## Context

The upload flow is a two-step process: `InitiateUpload` creates an `Image` record in the DB and returns a presigned PUT URL; the frontend uploads directly to R2 and then calls `CompleteUpload` to confirm. If the frontend never completes either step, the DB record persists indefinitely with no corresponding R2 object — a ghost entry that pollutes the image table.

The cleanup must be periodic (not request-triggered), must handle both the DB record and any orphaned R2 object, and must not introduce new infrastructure dependencies.

## Goals / Non-Goals

**Goals:**
- Soft-delete Image records where upload was never confirmed after 30 minutes
- Delete any corresponding R2 object for records being cleaned up
- Run cleanup automatically every 10 minutes in-process
- Mark uploads as confirmed when `CompleteUpload` is called

**Non-Goals:**
- Hard deletion of stale records (out of scope; handled by existing trash/purge flow if one exists)
- Distinguishing "FE never uploaded to R2" from "FE uploaded but never called /complete" — both are treated identically
- Graceful goroutine shutdown (main.go has no graceful shutdown today; this is consistent with existing behaviour)
- Retrying failed R2 deletes (a failed R2 delete is logged and the record is still soft-deleted)

## Decisions

### 1. `is_uploaded bool` over `upload_status varchar`

The cleanup logic only requires a binary signal: was this upload confirmed or not? `created_at` already provides the time dimension. A `varchar` status column would add expressiveness (e.g., `pending`, `failed`, `complete`) but no current requirement needs those extra states. A bool is simpler, smaller in the DB, and easier to index.

**Alternative considered:** `upload_status varchar` with values `pending` / `complete` — rejected for over-engineering given current scope.

### 2. Cleanup logic in `imageUsecase`, not in a standalone package

`CleanupStaleUploads` calls both the image repository (soft delete) and the storage service (R2 delete). Both are already dependencies of `imageUsecase`. Keeping it there avoids new wiring, stays consistent with the existing service layer pattern, and is directly unit-testable.

**Alternative considered:** A dedicated `CleanupService` — rejected as unnecessary abstraction for a single method.

### 3. Goroutine in `main.go` with `time.NewTicker`

A goroutine started after dependency wiring is the simplest approach: no new infrastructure (no cron daemon, no message queue), no new packages, and consistent with how Go services typically run background work.

**Alternative considered:** Using a cron library (e.g., `robfig/cron`) — rejected; a single ticker is sufficient and avoids an external dependency.

### 4. R2 delete is best-effort on cleanup

If the R2 delete fails during cleanup, the error is logged and the DB record is still soft-deleted. The `r2_path` remains on the soft-deleted record, so a future reconciliation pass could retry if ever needed. Blocking the soft-delete on a storage error would leave ghost DB entries, which is worse.

### 5. Cleanup threshold is 30 minutes

The presigned PUT URL TTL is also 30 minutes (`uploadURLTTL` in `image_usecase.go`). A stale record older than 30 minutes is guaranteed to have an expired URL — the frontend cannot complete the upload even if it tried.

## Risks / Trade-offs

- **Clock skew / race**: A record created just under 30 minutes ago but where the FE is still uploading could be cleaned up if the goroutine fires at exactly the wrong moment. Mitigation: the 30-minute threshold matches the presigned URL expiry, so any upload still in-flight at 30 minutes has an already-expired URL and cannot succeed anyway.
- **R2 delete failure**: Orphaned R2 objects remain if the delete call fails. These are small in number (only failed uploads) and can be reconciled manually or via a future cleanup pass. Logged at `warn` level.
- **No graceful shutdown**: The goroutine does not stop cleanly on SIGTERM today. Mid-cleanup, the process could be killed. Soft deletes are atomic per record, so partial runs leave no corrupt state — they just leave some stale records to be cleaned on next run.
- **DB query cost**: Every 10 minutes, a query scans for `is_uploaded = false AND created_at < now() - 30m`. With an index on `(is_uploaded, created_at)`, this is cheap. Without an index, it's a full table scan — acceptable at current scale but should be indexed.

## Migration Plan

1. Deploy DB migration adding `is_uploaded boolean NOT NULL DEFAULT false` to `images`
2. Deploy application code (goroutine + `CompleteUpload` update + `CleanupStaleUploads`)
3. No backfill needed: existing records without a confirmed upload are already stale and will be cleaned up by the goroutine on first run

**Rollback:** Drop the `is_uploaded` column (down migration). The goroutine and usecase method are inert without the column.

## Open Questions

- Should the cleanup goroutine emit a metric (e.g., count of records cleaned per run) for observability? Not scoped now but easy to add.
