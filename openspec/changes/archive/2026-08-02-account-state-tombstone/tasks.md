## 1. Migration

- [x] 1.1 Write a migration that adds `account_state TEXT NOT NULL DEFAULT 'active' CHECK (account_state IN ('active', 'pending_deletion', 'purged'))` and `purged_at TIMESTAMPTZ` (nullable) to `users`; backfills `account_state = 'pending_deletion'` for rows where `pending_kinde_deletion = true`
- [x] 1.2 Write a second migration that drops `pending_kinde_deletion` (run after new code is deployed)

## 2. Domain

- [x] 2.1 In `internal/domain/user.go`, define `AccountState` as a named string type with constants `AccountStateActive`, `AccountStatePendingDeletion`, `AccountStatePurged`; replace `PendingKindeDeletion bool` with `AccountState AccountState` (GORM tag: `column:account_state;default:active`) and add `PurgedAt *time.Time` (GORM tag: `column:purged_at`)

## 3. UserRepository Interface

- [x] 3.1 In `internal/usecase/user_repository.go`, remove `MarkPendingKindeDeletion` and `ListPendingKindeDeletion`; add `SetAccountState(ctx context.Context, id string, state domain.AccountState) error`, `MarkPurged(ctx context.Context, id string, purgedAt time.Time) error`, `ListByAccountState(ctx context.Context, state domain.AccountState) ([]*domain.User, error)`, and `ListPurgedBefore(ctx context.Context, threshold time.Time) ([]*domain.User, error)`

## 4. Repository Implementation

- [x] 4.1 In `internal/repository/user_repository.go`, implement `SetAccountState` (UPDATE users SET account_state = ? WHERE id = ?; return `ErrUserNotFound` if no rows affected), `MarkPurged` (UPDATE account_state = 'purged', purged_at = ? WHERE id = ?), `ListByAccountState` (SELECT WHERE account_state = ?), and `ListPurgedBefore` (SELECT WHERE account_state = 'purged' AND purged_at < ?)
- [x] 4.2 Remove `MarkPendingKindeDeletion` and `ListPendingKindeDeletion` implementations

## 5. Kinde Client

- [x] 5.1 Add `DeleteUserSessions(ctx context.Context, kindeUserID string) error` to the `KindeClient` interface in `internal/usecase/kinde_client.go`
- [x] 5.2 Implement `DeleteUserSessions` in `internal/kinde/kinde.go` — call Kinde's session revocation endpoint using the cached M2M token; treat "not found" or "no sessions" responses as success; return an error for any other non-2xx

## 6. Job Args

- [x] 6.1 In `internal/usecase/job_args.go`, add `AccountWipeArgs` (kind: `account_wipe`, `MaxAttempts: 5`, field: `UserID string`) and `AccountWipeReconcileArgs` (kind: `account_wipe_reconcile`) and `PurgedAccountSweepArgs` (kind: `purged_account_sweep`)
- [x] 6.2 Remove `DeleteAccountArgs` and `AccountKindeDeletionArgs` from `job_args.go`

## 7. Usecase

- [x] 7.1 In `internal/usecase/account_usecase.go`, add `MarkForDeletion(ctx context.Context, userID string) error`: check `account_state != 'active'` → return nil; call `SetAccountState('pending_deletion')`; best-effort `enqueuer.Insert(AccountWipeArgs{UserID})` — log warn on failure, return nil regardless
- [x] 7.2 Add `WipeAccount(ctx context.Context, userID string) error`: (1) DB wipe transaction — clear folder parents, hard-delete images/folders/tags/pending-uploads, collect R2 paths; (2) enqueue `R2DeleteArgs` per path; (3) `kinde.DeleteUserSessions`; (4) `kinde.DeleteUser`; (5) `userRepo.MarkPurged(ctx, userID, time.Now())`; (6) enqueue `BookletUserDeletionArgs` if `bookletClient` non-nil
- [x] 7.3 Add `ReconcilePendingDeletions(ctx context.Context) error`: call `ListByAccountState('pending_deletion')`; for each user insert `AccountWipeArgs` with `UniqueOpts` (unique by kind + user ID while pending/running); log warn on individual insert failures
- [x] 7.4 Add `SweepPurgedAccounts(ctx context.Context) error`: compute cutoff `time.Now().Add(-25 * time.Hour)`; call `ListPurgedBefore(cutoff)`; call `HardDelete` per result; log warn on individual failures; return nil
- [x] 7.5 Remove `ScheduleAccountDeletion`, `DeleteAccount`, `ProcessAccountKindeDeletion`, and `ReconcilePendingKindeDeletions` from `account_usecase.go`

## 8. Workers

- [x] 8.1 Create `internal/worker/account_wipe.go` — `AccountWipeWorker` calls `accountUsecase.WipeAccount(ctx, job.Args.UserID)`
- [x] 8.2 In `internal/worker/periodic.go`, replace `AccountKindeDeletionReconcileWorker` with `AccountWipeReconcileWorker` (calls `accountUsecase.ReconcilePendingDeletions`; interval: 5 minutes); add `PurgedAccountSweepWorker` (calls `accountUsecase.SweepPurgedAccounts`; interval: 24 hours)
- [x] 8.3 Remove `internal/worker/delete_account.go` and the old Kinde deletion worker file

## 9. Handlers

- [x] 9.1 In `internal/handler/me.go`, update `AccountUsecase` interface to replace `DeleteAccount` with `MarkForDeletion(ctx context.Context, userID string) error`; update `DeleteMe` to call `h.accountUsecase.MarkForDeletion(ctx, userID)` and return `c.NoContent(http.StatusAccepted)` (202)
- [x] 9.2 In `internal/handler/internal.go`, update `InternalAccountUsecase` interface to replace `ScheduleAccountDeletion` with `MarkForDeletion`; update the `DeleteAccount` handler to call `MarkForDeletion`
- [x] 9.3 In `internal/handler/middleware/auth.go`, replace the `user.PendingKindeDeletion` check with `user.AccountState != domain.AccountStateActive`

## 10. Wiring

- [x] 10.1 In `cmd/server/main.go`, register `worker.NewAccountWipeWorker`, `worker.NewAccountWipeReconcileWorker`, and `worker.NewPurgedAccountSweepWorker` with River; remove registration of old Kinde deletion and delete-account workers

## 11. Unit Tests — Usecase

- [x] 11.1 Update all stubs and fakes in `account_usecase_test.go` that reference `PendingKindeDeletion`, `MarkPendingKindeDeletion`, or `ListPendingKindeDeletion` to use the new `AccountState` field and methods; update the `KindeClient` stub to include `DeleteUserSessions`
- [x] 11.2 `TestAccountUsecase_MarkForDeletion_SetsStateAndEnqueuesJob`: assert `SetAccountState('pending_deletion')` is called and `AccountWipeArgs` is enqueued for an active user
- [x] 11.3 `TestAccountUsecase_MarkForDeletion_NoOpForNonActiveUser`: assert no state write and no enqueue when `account_state` is `'pending_deletion'` or `'purged'`
- [x] 11.4 `TestAccountUsecase_MarkForDeletion_EnqueueFailureReturnsNil`: assert that when the enqueue returns an error, `MarkForDeletion` still returns nil
- [x] 11.5 `TestAccountUsecase_WipeAccount_TransitionsToPurged`: assert `MarkPurged` is called after successful Kinde calls and the user is not hard-deleted
- [x] 11.6 `TestAccountUsecase_WipeAccount_DeleteUserSessionsCalledBeforeDeleteUser`: assert `DeleteUserSessions` is called before `DeleteUser`; assert that a `DeleteUserSessions` error causes `WipeAccount` to return an error without calling `DeleteUser`
- [x] 11.7 `TestAccountUsecase_ReconcilePendingDeletions_EnqueuesJobPerPendingUser`: assert `AccountWipeArgs` is inserted for each `pending_deletion` user
- [x] 11.8 `TestAccountUsecase_SweepPurgedAccounts_HardDeletesExpiredRows`: assert `HardDelete` is called for each user returned by `ListPurgedBefore`
- [x] 11.9 `TestAccountUsecase_SweepPurgedAccounts_ContinuesOnHardDeleteFailure`: assert a `HardDelete` error for one user does not prevent remaining users from being processed and `SweepPurgedAccounts` returns nil

## 12. Unit Tests — Handler

- [x] 12.1 In `handler/me_test.go`, update the `DeleteMe` test to assert the response is `202 Accepted` and `MarkForDeletion` is called with the correct user ID
- [x] 12.2 In `handler/middleware/auth_test.go`, rename and update the `TestAuthMiddleware_PendingKindeDeletion_Returns401` test to cover `account_state = 'pending_deletion'`; add `TestAuthMiddleware_PurgedUser_Returns401` asserting a `'purged'` user is also rejected
- [x] 12.3 In `handler/internal_test.go`, update the `DeleteAccount` handler tests to assert `MarkForDeletion` is called and scenarios for `pending_deletion`, `purged`, and unprovisioned users all return `202`

## 13. Integration Tests

- [x] 13.1 In `internal/repository/account_repository_integration_test.go`, replace `MarkPendingKindeDeletion` and `ListPendingKindeDeletion` tests with integration tests for `SetAccountState`, `MarkPurged`, `ListByAccountState`, and `ListPurgedBefore`

## 14. Lint

- [x] 14.1 Run `golangci-lint run ./...` from `backend/` and fix any issues
