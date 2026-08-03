## Purpose

Define the account tombstone lifecycle: how an account transitions through `active → pending_deletion → purged`, the domain model and repository methods that underpin those transitions, the `AccountWipeJob` that executes the full wipe, and the periodic workers that recover stuck tombstones and sweep expired purged rows.

## Requirements

### Requirement: Account State Domain Model

The system SHALL define an `AccountState` type (Go string constants) with three valid values: `AccountStateActive` (`"active"`), `AccountStatePendingDeletion` (`"pending_deletion"`), and `AccountStatePurged` (`"purged"`).

The `users` table SHALL have an `account_state TEXT NOT NULL DEFAULT 'active'` column with a CHECK constraint restricting values to the three valid states. The `users` table SHALL also have a nullable `purged_at TIMESTAMPTZ` column, which is non-null only when `account_state = 'purged'`.

The `domain.User` struct SHALL replace the `PendingKindeDeletion bool` field with `AccountState AccountState` and add `PurgedAt *time.Time`.

#### Scenario: Default account state is active

- **WHEN** a new user row is inserted without specifying `account_state`
- **THEN** the row has `account_state = 'active'`

#### Scenario: Invalid account state is rejected at the database level

- **WHEN** an `account_state` value other than `'active'`, `'pending_deletion'`, or `'purged'` is written to the `users` table
- **THEN** the database rejects the write with a constraint violation

### Requirement: SetAccountState Repository Method

The system SHALL expose `SetAccountState(ctx context.Context, id string, state domain.AccountState) error` on `UserRepository`. It SHALL update `account_state` for the given user ID. If no row matches, it SHALL return `ErrUserNotFound`.

#### Scenario: State is updated for an existing user

- **WHEN** `SetAccountState` is called with a valid user ID and state `'pending_deletion'`
- **THEN** the user's `account_state` column is `'pending_deletion'`
- **AND** the method returns `nil`

#### Scenario: Unknown user returns ErrUserNotFound

- **WHEN** `SetAccountState` is called with a user ID that does not exist
- **THEN** the method returns `ErrUserNotFound`

### Requirement: MarkPurged Repository Method

The system SHALL expose `MarkPurged(ctx context.Context, id string, purgedAt time.Time) error` on `UserRepository`. It SHALL set `account_state = 'purged'` and `purged_at = purgedAt` for the given user ID. If no row matches, it SHALL return `ErrUserNotFound`.

#### Scenario: User row is transitioned to purged

- **WHEN** `MarkPurged` is called with a valid user ID and a timestamp
- **THEN** the user's `account_state` is `'purged'`
- **AND** the user's `purged_at` matches the provided timestamp
- **AND** the method returns `nil`

### Requirement: ListByAccountState Repository Method

The system SHALL expose `ListByAccountState(ctx context.Context, state domain.AccountState) ([]*domain.User, error)` on `UserRepository`. It SHALL return all users whose `account_state` matches the given value.

#### Scenario: Returns only users matching the requested state

- **WHEN** `ListByAccountState` is called with `'pending_deletion'`
- **THEN** only user rows with `account_state = 'pending_deletion'` are returned
- **AND** users with other states are excluded

#### Scenario: Returns empty slice when no users match

- **WHEN** `ListByAccountState` is called and no users have the requested state
- **THEN** the method returns an empty slice and `nil` error

### Requirement: ListPurgedBefore Repository Method

The system SHALL expose `ListPurgedBefore(ctx context.Context, threshold time.Time) ([]*domain.User, error)` on `UserRepository`. It SHALL return all users with `account_state = 'purged'` and `purged_at < threshold`.

#### Scenario: Returns purged users whose purged_at predates the threshold

- **WHEN** `ListPurgedBefore` is called with a threshold timestamp
- **THEN** only users with `account_state = 'purged'` and `purged_at` strictly before the threshold are returned

#### Scenario: Users purged after the threshold are excluded

- **WHEN** `ListPurgedBefore` is called and a purged user has `purged_at >= threshold`
- **THEN** that user is not included in the result

### Requirement: MarkForDeletion Usecase

The system SHALL expose `MarkForDeletion(ctx context.Context, userID string) error` on `accountUsecase`. It SHALL:

1. Check `account_state` — if not `'active'`, return `nil` immediately (idempotent no-op)
2. Call `SetAccountState('pending_deletion')`
3. Best-effort enqueue `AccountWipeArgs{UserID}` — on enqueue failure, log a warning and return `nil` (do not propagate the error)

#### Scenario: Active user is marked for deletion and wipe job is enqueued

- **WHEN** `MarkForDeletion` is called for a user with `account_state = 'active'`
- **THEN** the user's `account_state` is `'pending_deletion'`
- **AND** an `AccountWipeArgs` job is enqueued for that user ID
- **AND** the method returns `nil`

#### Scenario: Already non-active user is a no-op

- **WHEN** `MarkForDeletion` is called for a user whose `account_state` is `'pending_deletion'` or `'purged'`
- **THEN** no state write occurs
- **AND** no job is enqueued
- **AND** the method returns `nil`

#### Scenario: Enqueue failure does not fail MarkForDeletion

- **WHEN** `MarkForDeletion` is called and the job enqueue returns an error
- **THEN** the method logs a warning
- **AND** returns `nil` (the state flip is preserved; the reconcile recovers the enqueue)

### Requirement: AccountWipeJob

The system SHALL define an `AccountWipeArgs` job (kind: `account_wipe`, `MaxAttempts: 5`) and a corresponding worker. On each attempt, the worker SHALL call `accountUsecase.WipeAccount(ctx, userID)`.

`WipeAccount` SHALL execute the following steps in order:
1. Call `kinde.DeleteUserSessions(ctx, userID)`
2. Call `kinde.DeleteUser(ctx, userID)`
3. Delete all of the user's owned DB rows (images, folders, tags, pending uploads) in a single transaction, collecting R2 paths; the transaction SHALL NOT modify the `users` row
4. Enqueue one `R2DeleteArgs` job per collected R2 path
5. Call `userRepo.MarkPurged(ctx, userID, now())`
6. Enqueue `BookletUserDeletionArgs{UserID}` if `bookletClient` is configured

Kinde runs before the DB wipe so that identity deletion is not delayed by DB failures. All steps are idempotent: Kinde calls treat HTTP 200 and 400 as success (400 indicates the user no longer exists); DB DELETEs are no-ops on already-deleted rows; `MarkPurged` on an already-purged row is a safe UPDATE. On any step failure, the job returns an error and River retries the entire job.

#### Scenario: Full wipe transitions user to purged

- **WHEN** `AccountWipeJob` runs successfully for a user
- **THEN** the user's images, folders, tags, and pending uploads no longer exist in the database
- **AND** the user's `account_state` is `'purged'`
- **AND** `purged_at` is set to approximately the time the job ran

#### Scenario: Kinde call failure retries the job

- **WHEN** `kinde.DeleteUser` returns an error
- **THEN** the job is retried with River's default backoff
- **AND** the user's `account_state` remains `'pending_deletion'`

#### Scenario: Already-wiped user is handled idempotently

- **WHEN** `AccountWipeJob` runs a second time for a user whose Kinde identity and DB data were already deleted
- **THEN** the Kinde calls return success (400 treated as success)
- **AND** the DB wipe step completes without error (no-op DELETEs)
- **AND** the job proceeds to `MarkPurged`

### Requirement: AccountWipeReconcileWorker

The system SHALL register a periodic River job (`AccountWipeReconcileArgs`, kind: `account_wipe_reconcile`, interval: 5 minutes). On each run, the worker SHALL call `accountUsecase.ReconcilePendingDeletions(ctx)`.

`ReconcilePendingDeletions` SHALL query `ListByAccountState('pending_deletion')` and attempt to insert an `AccountWipeArgs` job for each result using River's `UniqueOpts` (unique by kind + user ID while the job is active — available, scheduled, pending, running, or retryable — or discarded). An existing active or discarded job for a given user ID SHALL be treated as a no-op insert; only users with no job record at all receive a new enqueue.

#### Scenario: Pending deletion users without an active wipe job are enqueued

- **WHEN** the reconcile worker runs and a user has `account_state = 'pending_deletion'` with no active `AccountWipeJob`
- **THEN** an `AccountWipeArgs` job is enqueued for that user

#### Scenario: Users with an already-active wipe job are skipped

- **WHEN** the reconcile worker runs and a user has `account_state = 'pending_deletion'` with an active `AccountWipeJob` already in River
- **THEN** no duplicate job is enqueued

### Requirement: PurgedAccountSweepWorker

The system SHALL register a periodic River job (`PurgedAccountSweepArgs`, kind: `purged_account_sweep`, interval: 24 hours). On each run, the worker SHALL call `accountUsecase.SweepPurgedAccounts(ctx)`.

`SweepPurgedAccounts` SHALL compute a cutoff of `now() - purgedAccountTTL` (constant: `purgedAccountTTL = 25 * time.Hour`, representing the 24h Kinde access token TTL plus a 1h buffer), query `ListPurgedBefore(cutoff)`, and call `HardDelete` for each result. A failure to hard-delete a single row SHALL be logged as a warning and SHALL NOT abort the sweep for remaining rows. `SweepPurgedAccounts` SHALL return `nil` even if individual hard-deletes fail.

#### Scenario: Expired tombstones are hard-deleted

- **WHEN** the sweep job runs and a user has `account_state = 'purged'` with `purged_at` more than 25 hours ago
- **THEN** that user's row is hard-deleted from `users`

#### Scenario: Recent tombstones are not deleted

- **WHEN** the sweep job runs and a purged user's `purged_at` is fewer than 25 hours ago
- **THEN** that user's row is not deleted

#### Scenario: Hard-delete failure is logged and sweep continues

- **WHEN** the sweep job runs and `HardDelete` returns an error for one user
- **THEN** the error is logged as a warning
- **AND** the sweep continues processing remaining users
- **AND** `SweepPurgedAccounts` returns `nil`
