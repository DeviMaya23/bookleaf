## Why

Account deletion currently hard-deletes the `users` row immediately after Kinde identity deletion, leaving a window where a still-alive JWT (24h TTL) could pass local validation. Additionally, the DB wipe runs synchronously in the request handler, and the user-triggered and Booklet-triggered deletion paths are separate code paths doing the same thing.

This change introduces an `account_state` enum as the single authoritative signal for an account's lifecycle — active, pending deletion, purged — and makes the entire deletion pipeline async. The handler flips the state and best-effort enqueues the wipe job immediately; a periodic reconcile worker serves as the recovery mechanism for cases where the enqueue did not complete.

## What Changes

- **Replace `pending_kinde_deletion bool` with `account_state` enum** (`active` / `pending_deletion` / `purged`) plus a `purged_at` timestamp on the `users` row
- **Auth middleware checks `account_state != 'active'`** — blocks both `pending_deletion` and `purged` users
- **Both `DELETE /me` and `DELETE /internal/accounts/:id` converge** — both call a single `MarkForDeletion` usecase that flips `account_state` to `pending_deletion` and best-effort enqueues `AccountWipeJob` immediately; a crash between the two operations is recovered by the reconcile worker
- **New periodic `AccountWipeReconcileWorker`** (recovery mechanism) — finds all `pending_deletion` rows and enqueues an `AccountWipeJob` for each not already in flight, via River deduplication; minimises the re-registration gap by ensuring the `pending_deletion` window is short even if the initial enqueue was lost
- **New `AccountWipeJob`** combines what were previously three separate jobs (DB wipe, Kinde wipe, session kill) into one: kill Kinde sessions → delete Kinde user → wipe DB data → enqueue R2 delete jobs per object → `MarkPurged`; also enqueues `BookletUserDeletionArgs` if configured; Kinde runs first so identity deletion is not delayed by DB failures
- **New periodic `PurgedAccountSweepWorker`** — hard-deletes `purged` rows where `purged_at + 25h < now()` (24h token TTL + 1h buffer)
- **`DELETE /me` returns `202 Accepted`** instead of `204 No Content`

## Capabilities

### New Capabilities

- `account-tombstone`: `account_state` enum and full lifecycle (`active → pending_deletion → purged → [hard deleted]`), `purged_at` field, `MarkForDeletion` usecase (state flip + best-effort enqueue), `AccountWipeReconcileWorker` as recovery driver, `AccountWipeJob`, and `PurgedAccountSweepWorker`

### Modified Capabilities

- `account-deletion`: `DELETE /me` now returns `202`; wipe is fully async; auth middleware checks `account_state`; `AccountKindeDeletionArgs` and `DeleteAccountArgs` replaced by `AccountWipeJob`; `ReconcilePendingKindeDeletions` replaced by `AccountWipeReconcileWorker`
- `booklet-deletion-sync`: `DELETE /internal/accounts/:id` now calls `MarkForDeletion` (state flip + best-effort enqueue) instead of `ScheduleAccountDeletion`; idempotency gate broadened to `account_state != 'active'`

## Impact

- `backend/internal/domain/user.go` — replace `PendingKindeDeletion bool` with `AccountState AccountState`; add `PurgedAt *time.Time`
- DB migration — add `account_state` column (default `'active'`), add `purged_at`, backfill `pending_kinde_deletion = true` → `'pending_deletion'`, drop old column
- `backend/internal/usecase/user_repository.go` — replace `MarkPendingKindeDeletion` / `ListPendingKindeDeletion` with `SetAccountState`, `MarkPurged`, `ListByAccountState`, `ListPurgedBefore`
- `backend/internal/repository/user_repository.go` — implement updated interface
- `backend/internal/usecase/account_usecase.go` — new `MarkForDeletion`; new `WipeAccount` (combined DB + Kinde); new `SweepPurgedAccounts`; remove `ScheduleAccountDeletion`, `DeleteAccount`, `ProcessAccountKindeDeletion`, `ReconcilePendingKindeDeletions`
- `backend/internal/usecase/job_args.go` — new `AccountWipeArgs`; remove `DeleteAccountArgs`, `AccountKindeDeletionArgs` (replaced)
- `backend/internal/handler/me.go` — `DeleteMe` calls `MarkForDeletion`, returns `202`
- `backend/internal/handler/internal.go` — `DeleteAccount` handler calls `MarkForDeletion`
- `backend/internal/handler/middleware/auth.go` — check `account_state != 'active'`
- `backend/internal/platform/kinde/kinde.go` — new `DeleteUserSessions(ctx, userID) error`
- `backend/internal/worker/` — new `AccountWipeWorker`; updated `AccountWipeReconcileWorker`; new `PurgedAccountSweepWorker`; remove old Kinde deletion and delete-account workers
- No frontend or extension changes
