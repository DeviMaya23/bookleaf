## Context

Account deletion currently runs synchronously: `DELETE /me` calls `DeleteAccount` directly, which wipes all DB rows in a transaction, then enqueues R2 and Kinde jobs. The `users` row is hard-deleted after the Kinde job succeeds. Two problems:

1. **JWT gap**: the row is hard-deleted immediately after Kinde identity deletion. A still-alive token (24h TTL) has nothing to check against — the auth middleware would fail or silently pass instead of returning a clean 401.
2. **Two divergent paths**: `DELETE /me` calls `DeleteAccount` directly (sync wipe); `DELETE /internal/accounts/:id` calls `ScheduleAccountDeletion` → enqueues `DeleteAccountArgs` → `DeleteAccount` (async). Any change to the deletion logic must be applied in both places.

The previous iteration of this design introduced `account_state` but still enqueued a job from the request handler — creating a crash window between the state write and the enqueue. The design here eliminates that window by removing the enqueue from the request path entirely.

## Goals / Non-Goals

**Goals:**
- `account_state` is the single authoritative signal for an account's lifecycle; once it is `pending_deletion`, the system eventually completes all cleanup without further input from the request handler
- No crash window: the handler's only DB write is the state flip, which is atomic
- Auth middleware can 401 any request from a non-active account through the full TTL window
- Both deletion entry points (`DELETE /me`, `DELETE /internal/accounts/:id`) converge on the same one-line usecase call
- Expired tombstones are cleaned up automatically

**Non-Goals:**
- Sub-second deletion latency — eventual cleanup within the reconcile interval is acceptable
- Token blacklisting or Kinde token introspection
- Per-user configurable TTL

## Decisions

### `account_state` as TEXT with CHECK constraint, default `'active'`

Three values: `active`, `pending_deletion`, `purged`. TEXT + CHECK instead of a Postgres enum type — adding enum values requires a DDL type alteration; dropping and re-adding a CHECK constraint is simpler. The Go domain layer defines `AccountState` as a named string type with constants.

`purged_at TIMESTAMPTZ` is nullable; non-null only when `account_state = 'purged'`.

### Handler calls `MarkForDeletion`: state flip + best-effort enqueue

`MarkForDeletion` does two things in sequence:
1. `SetAccountState('pending_deletion')` — a single UPDATE, atomic
2. `enqueuer.Insert(AccountWipeArgs{UserID})` — best-effort; logged as a warning on failure, does not cause `MarkForDeletion` to return an error

The idempotency check (`account_state != 'active'` → no-op) runs before the state write. If the server crashes between the state write and the enqueue, the row stays `pending_deletion` and the reconcile worker picks it up on its next tick.

The immediate enqueue is not just a latency optimisation — it also minimises the `pending_deletion` window. There is a narrow edge case where a user's Kinde account could be deleted by means outside Bookleaf's wipe job during this window (e.g. the user deletes it directly from Kinde), enabling re-registration with a new Kinde ID before Bookleaf's old data is wiped. Shortening this window via immediate enqueue reduces — but does not eliminate — that exposure.

### `AccountWipeReconcileWorker` is the recovery mechanism

A periodic River job (interval: 5 minutes) queries `ListByAccountState('pending_deletion')` and inserts an `AccountWipeArgs` job for each result using River's `UniqueOpts` (unique by kind + user ID while the job is pending or running). If a job already exists for that user, the insert is a no-op.

This recovers two failure cases: a crash between state write and enqueue in `MarkForDeletion`, and a `AccountWipeJob` that exhausted all retries and was discarded by River (UniqueOpts uniqueness only applies to pending/running jobs; a discarded job is re-enqueued on the next reconcile tick).

### `AccountWipeJob` combines DB wipe + Kinde into one job

Steps in order:
1. `kinde.DeleteUserSessions(ctx, userID)`
2. `kinde.DeleteUser(ctx, userID)`
3. Wipe DB (images, folders, tags, pending uploads) — idempotent DELETEs
4. Enqueue one `R2DeleteArgs` per R2 path collected during the wipe
5. `userRepo.MarkPurged(ctx, userID, now())`
6. Enqueue `BookletUserDeletionArgs` if `bookletClient` is configured

Kinde runs before the DB wipe so that identity deletion is not delayed by DB failures. If the job fails on the DB step and retries, the Kinde calls are no-ops (200 or 400 on an already-deleted user — both treated as success). The account stays `pending_deletion` while DB wipe retries are in flight; the reconcile worker re-enqueues if all retries are exhausted.

R2 and Booklet remain separate jobs because they have independent retry needs (R2: per-object granularity; Booklet: 10 attempts, different failure domain). Kinde is combined with the DB wipe because the separation only existed to keep Kinde off the synchronous request — that constraint is gone.

If any step fails, River retries the entire job. All steps are idempotent: DB DELETEs are no-ops on already-deleted rows; `DeleteUserSessions` and `DeleteUser` treat "not found" as success; `MarkPurged` is an UPDATE on an already-purged row (idempotent); R2/Booklet enqueues are safe to repeat.

### `MarkPurged` happens at the end of `AccountWipeJob`

A user is `purged` once their Kinde identity is deleted — that is when their tokens become semantically invalid. `purged_at` drives the sweep TTL, so it should reflect when the identity was actually destroyed, not when the DB wipe ran.

### `PurgedAccountSweepWorker` runs every 24h, TTL = 25h

Queries `ListPurgedBefore(now() - 25h)` and calls `HardDelete` per result. 25h = 24h Kinde access token TTL + 1h buffer. Hard-delete failures are logged as warnings; the sweep continues for remaining rows and returns nil (partial success is acceptable). The constant `purgedAccountTTL = 25 * time.Hour` lives in `account_usecase.go`.

### `DELETE /me` returns `202 Accepted`

Previously `204 No Content` (sync wipe completed before response). 202 is correct when work is accepted and continues asynchronously.

### Old job types removed

`DeleteAccountArgs` (kind: `delete_account`) and `AccountKindeDeletionArgs` (kind: `account_kinde_deletion`) are replaced by `AccountWipeArgs` (kind: `account_wipe`). The old workers are removed. Any jobs of the old kinds still in River's queue at deploy time will fail to find a worker — they should be drained or cancelled before deploying.

### `KindeClient` interface gains `DeleteUserSessions`

Added alongside the existing `DeleteUser`. The concrete `kinde.Client` calls the Kinde Management API session revocation endpoint.

Both `DeleteUserSessions` and `DeleteUser` treat HTTP `200` and `400` as success. Kinde returns `200` on a successful operation and `400` (with an `INVALID_USER` error code) when the user ID is not recognised — which covers both the "already deleted" and "never existed" cases. `403`, `429`, and any other non-200/400 status are returned as errors and trigger a River retry. No response body parsing is used; the status code alone is the signal.

## Risks / Trade-offs

**Deletion latency up to reconcile interval** → login is blocked immediately (state flip is synchronous). The wipe job starts within 5 minutes. Acceptable for account deletion.

**Old River job kinds at deploy time** → `delete_account` and `account_kinde_deletion` jobs already in the queue will have no worker. Mitigation: drain or cancel those jobs before deploying, or keep the old workers registered for one release cycle.

**`AccountWipeJob` retry re-runs Kinde calls** → if the DB wipe fails, the job retries and re-runs the Kinde calls first (as no-ops — 200 or 400 on an already-deleted user) before hitting the DB again. Harmless but slightly wasteful. Acceptable given deletion is rare.

**`purged` rows visible to normal GORM queries** → `MarkPurged` does not set `deleted_at`, so GORM's soft-delete filter does not exclude them. Intentional — the auth middleware needs to find them by Kinde ID to return 401.

## Migration Plan

1. Add `account_state TEXT NOT NULL DEFAULT 'active' CHECK (account_state IN ('active', 'pending_deletion', 'purged'))` column
2. Add `purged_at TIMESTAMPTZ` (nullable)
3. `UPDATE users SET account_state = 'pending_deletion' WHERE pending_kinde_deletion = true`
4. Deploy new code
5. Drop `pending_kinde_deletion` column (second migration, after deploy)
