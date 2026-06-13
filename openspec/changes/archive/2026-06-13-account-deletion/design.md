## Context

Bookleaf currently has no account deletion path. User-owned data (`images`, `folders`, `tags`, `pending_uploads`) all have `ON DELETE RESTRICT` (or unspecified, which defaults to the same) foreign keys to `users.id`, so the `users` row cannot be deleted while any owned rows exist. Identity is managed entirely by Kinde — the `users.id` column *is* the Kinde user ID, with no local password/credentials to wipe.

The user has already created an M2M application in Kinde (dev environment) with `read:users`, `delete:user_session`, and `delete:users` scopes against the Management API, in preparation for this change.

## Goals / Non-Goals

**Goals:**
- Permanently and completely remove a user's app data (DB rows + R2 objects) on request.
- Permanently remove the user's identity from Kinde.
- Make the operation resumable/retry-safe if the process crashes or an external call (Kinde, R2) fails partway through.
- Ensure a deleted account cannot continue to be used, even if the user's current access token hasn't expired yet.

**Non-Goals:**
- Frontend trigger/confirmation UI (separate future change).
- A grace period / undo window — this is intentionally irreversible ("scorched earth"), not a soft delete.
- Email-based confirmation flows — no email infrastructure exists in this app.
- Admin-initiated deletion of other users (this is self-service `DELETE /me` only).

## Decisions

### 1. Two-phase deletion: synchronous app-data wipe, then async Kinde cleanup

`DELETE /me` performs the following **synchronously**, in one DB transaction:

1. Set `parent_id = NULL` on all of the user's folders (breaks the self-referential `fk_folders_parent` FK so all folders can be deleted regardless of nesting).
2. Read (unscoped, including soft-deleted/trashed) and then hard-delete all of the user's `images`, collecting `r2_path` + `thumbnail_path` for each.
3. Hard-delete all of the user's `folders`. (`image_folders` rows cascade automatically.)
4. Hard-delete all of the user's `tags`. (`image_tags` rows cascade automatically.)
5. Read and then hard-delete all of the user's `pending_uploads`, collecting `r2_path` for each.
6. Set `users.pending_kinde_deletion = true` (the user row itself is **not** deleted yet).

After the transaction commits, the usecase:
7. Enqueues one `R2DeleteArgs` job per collected image/pending-upload object (reusing the existing worker — no changes needed there).
8. Enqueues one `AccountKindeDeletionArgs{UserID}` job.

A new `AccountKindeDeletionWorker` then, **asynchronously**:
- Fetches an M2M access token from Kinde (cached).
- Calls `DELETE /api/v1/user?id=<kinde_user_id>&is_delete_profile=true` to fully remove the user's identity from Kinde (including across orgs/subscriber list — consistent with "scorched earth"). Documented responses are `200/400/403/429`; a response indicating the user no longer exists (exact code TBD, to be confirmed empirically) is treated as success for idempotency (covers retried/reconciled jobs hitting an already-deleted user).
- On success, hard-deletes the `users` row.
- On failure, retries with River's default backoff (`MaxAttempts: 5`, matching `R2DeleteArgs`).

**Why this order (DB/R2 first, Kinde last):** explored at length in the proposal discussion — if the Kinde deletion happened first and the app then crashed/failed before wiping app data, the result is an unrecoverable state: the user can never authenticate again, yet their PII remains in the DB/R2 with no path to clean it up. Doing app-data cleanup first means any failure leaves, at worst, an "empty shell" account — recoverable either by the user retrying (if their session is still valid) or by the reconciliation sweep below (which needs no user interaction, since Kinde Management API calls use M2M credentials, not the user's session).

**Alternative considered:** delete the `users` row in the same transaction as steps 1–5, and handle Kinde deletion via a job that just takes the Kinde user ID (no DB row needed). Rejected because without a tombstone row, there's nothing to drive the reconciliation sweep if the initial job enqueue is lost (e.g., crash between transaction commit and job insert) — River job inserts can't trivially be made part of the GORM transaction here given the existing `riverEnqueuer` wiring.

### 2. Tombstone field on `users`: `pending_kinde_deletion`

A new boolean column, `pending_kinde_deletion BOOLEAN NOT NULL DEFAULT false`, added via migration. Drives both the async worker and the reconciliation sweep (decision 4).

### 3. Immediate lockout via auth middleware check

The auth middleware already calls `userUsecase.GetOrProvision` (which does a `GetByID`) on every request. It will additionally check `user.PendingKindeDeletion` and return `401 Unauthorized` if true — independent of whether Kinde has actually revoked the session yet.

**Why:** Kinde's deletion call is async (job-based, can be delayed/retried). Without this check, a user could continue using the app with a still-valid access token for some period after "deleting" their account, even though their data is already gone (resulting in confusing empty-state UI). Checking the local tombstone flag gives immediate, synchronous lockout regardless of Kinde API timing — and is the reason a separate Kinde session-revocation call isn't needed (see decision 1).

### 4. Periodic reconciliation sweep for stuck tombstones

A new periodic River job (interval: 24h, alongside the existing `TrashPurgeArgs`/`CleanupStaleUploadsArgs` jobs) queries `users` where `pending_kinde_deletion = true` and re-enqueues `AccountKindeDeletionArgs` for each. This covers:
- The job insert being lost after the DB transaction committed (process crash between steps 6 and 8).
- A job that was discarded after exhausting `MaxAttempts` due to a prolonged Kinde outage.

### 5. New Kinde Management API client package

A new `internal/platform/kinde` package (mirroring the layering of `internal/storage` for R2) provides:
- M2M client-credentials token fetch with in-memory caching (refetch when near expiry).
- `DeleteUser(ctx, kindeUserID) error` — calls `DELETE /api/v1/user?id=<kindeUserID>&is_delete_profile=true`; treats a "user not found"-equivalent response as success (exact status code to be confirmed during implementation, since Kinde's documented responses are `200/400/403/429`).

New required config (`KindeConfig` additions, following `requireEnv` pattern):
- `KINDE_M2M_CLIENT_ID`
- `KINDE_M2M_CLIENT_SECRET`
- `KINDE_M2M_TOKEN_URL` (or derive from `KINDE_ISSUER_URL` + `/oauth2/token`)
- Management API audience (likely `<issuer>/api`, to be confirmed against the M2M app's configured API in the Kinde dashboard)

Only the `delete:users` scope is required on the M2M application.

This is a new external dependency, called out explicitly per the proposal's Impact section.

### 6. `DELETE /me` response

Returns `204 No Content` on successful completion of the synchronous app-data wipe (steps 1–6 above). The Kinde deletion and R2 cleanup are best-effort background work from the client's perspective — by the time `204` is returned, the user's data is already gone and (per decision 3) their session is already locked out on their very next request.

## Risks / Trade-offs

- **[Large accounts → long transaction]** A user with thousands of images means step 2 reads+deletes thousands of rows in one transaction, holding locks longer. → Acceptable for v1 (self-service deletion is rare and not latency-sensitive); revisit batching if it becomes a problem.
- **[Kinde API call failures]** Kinde Management API could be down or scopes misconfigured. → Mitigated by job retries (5 attempts with backoff) + 24h reconciliation sweep; app data is already gone either way, so the user-facing harm is limited to "Kinde account technically still exists."
- **[M2M token cache races]** Concurrent requests to the worker could both see an expired cached token and refetch simultaneously. → Acceptable; an extra token fetch is cheap and Kinde won't reject concurrent valid client-credentials requests.
- **[In-flight uploads at deletion time]** A `pending_upload` or an upload mid-`CompleteUpload` could race with the deletion transaction. → Accepted as out of scope for v1; the transaction simply deletes whatever `pending_uploads`/`images` rows exist at the time it runs. No additional guard.
- **[Management API audience]** The exact Management API audience value for the M2M token hasn't been verified against this Kinde tenant's dashboard config. → Implementer must confirm during implementation of the `internal/platform/kinde` client.

## Migration Plan

1. Add migration `000014_add_pending_kinde_deletion_to_users` — `ALTER TABLE users ADD COLUMN pending_kinde_deletion BOOLEAN NOT NULL DEFAULT false` (down: drop column).
2. Add new required env vars for the Kinde M2M client; `config.Load()` fails fast if missing (consistent with existing `requireEnv` usage).
3. No backfill needed — `pending_kinde_deletion` defaults to `false` for all existing users.
4. Rollback: dropping the column is safe as long as no deletions are mid-flight (in practice, self-service deletion volume is low; if a rollback is needed mid-flight, any in-progress `AccountKindeDeletionArgs` jobs would fail on their next attempt referencing the dropped column — acceptable for a dev-stage rollback).

## Open Questions

- Exact Management API base URL / audience for the M2M token — to be confirmed from the Kinde dashboard during implementation.
