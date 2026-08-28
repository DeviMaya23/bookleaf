## Purpose

Synchronise account deletions from Bookleaf to Booklet. When Booklet triggers a `DELETE /internal/accounts/:id` request, Bookleaf schedules full account deletion and, after the data wipe, notifies Booklet to remove the user on its side as well.

## Requirements

### Requirement: Booklet HTTP Client

The system SHALL provide a Booklet HTTP client (`internal/booklet/booklet.go`) that exposes a `DeleteUser(ctx context.Context, userID string) error` method. The client SHALL call `DELETE /internal/users/:id` on Booklet, sending the configured `BookletInternalSecret` as the `X-Booklet-Internal-Secret` request header. A `2xx` response SHALL be treated as success. A `404 Not Found` response SHALL also be treated as success — the user does not exist in Booklet, which is the desired end state. Any other non-`2xx` response SHALL return an error.

The `usecase` package SHALL define the `BookletClient` interface with the `DeleteUser` method. The concrete `booklet.Client` type SHALL implement this interface.

#### Scenario: Successful deletion call returns no error

- **WHEN** `DeleteUser` is called and Booklet responds with `2xx`
- **THEN** the method returns `nil`

#### Scenario: 404 response is treated as success

- **WHEN** `DeleteUser` is called and Booklet responds with `404 Not Found`
- **THEN** the method returns `nil`

#### Scenario: Non-2xx non-404 response returns error

- **WHEN** `DeleteUser` is called and Booklet responds with a non-`2xx` status other than `404`
- **THEN** the method returns a non-nil error describing the status code

### Requirement: BookletUserDeletion Job

The system SHALL define a `BookletUserDeletionArgs` job (kind: `booklet_user_deletion`, `MaxAttempts: 10`) carrying `IDPSubject string` — the Kinde subject identifying the user in Booklet. The worker SHALL call `ProcessBookletUserDeletion(ctx, idpSubject string)`, which calls `bookletClient.DeleteUser(ctx, idpSubject)`.

Booklet identifies users by Kinde subject. The internal UUID SHALL NOT be used here.

#### Scenario: Successful Booklet call completes the job

- **WHEN** the `BookletUserDeletionArgs` worker calls `DeleteUser` with the idp_subject and it returns nil
- **THEN** the job completes without error

#### Scenario: Failed Booklet call retries the job

- **WHEN** the `BookletUserDeletionArgs` worker calls `DeleteUser` and it returns an error
- **THEN** the job is retried with River's default backoff up to 10 attempts

### Requirement: DeleteAccount Job

The system SHALL define a `DeleteAccountArgs` job (kind: `delete_account`, `MaxAttempts: 5`) and a corresponding worker. The worker SHALL call `accountUsecase.DeleteAccount(ctx, userID)`. On success the job completes. On failure the job retries with River's default exponential backoff.

#### Scenario: Successful DeleteAccount call completes the job

- **WHEN** the `DeleteAccountArgs` worker calls `DeleteAccount` and it returns nil
- **THEN** the job completes without error

#### Scenario: DeleteAccount failure retries the job

- **WHEN** the `DeleteAccountArgs` worker calls `DeleteAccount` and it returns an error
- **THEN** the job is retried with River's default backoff up to 5 attempts

### Requirement: DELETE /internal/accounts/:id Endpoint

The system SHALL expose a `DELETE /internal/accounts/:id` endpoint on the existing `/internal` route group, protected by the existing `X-Bookleaf-Internal-Secret` middleware. The `:id` parameter is the Kinde subject (idp_subject).

The handler SHALL pass the Kinde subject directly to `MarkForDeletion(ctx, idpSubject string)` and return `202 Accepted`.

`MarkForDeletion` owns the full resolution and branching logic:
- It calls `GetByIDPSubject` internally.
- **User found, `account_state = 'active'`**: sets state to `'pending_deletion'` and enqueues `AccountWipeArgs{IDPSubject}`.
- **User found, `account_state != 'active'`**: no-op (already in the pipeline).
- **User not found** (unprovisioned): enqueues `AccountWipeArgs{IDPSubject}` directly for Kinde-only cleanup.

In all cases `MarkForDeletion` returns `nil` and the handler returns `202 Accepted`.

#### Scenario: Active account is marked for deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a Kinde subject that resolves to an active user
- **THEN** the user's `account_state` is set to `'pending_deletion'` before the response is returned
- **AND** an `AccountWipeArgs{IDPSubject: ...}` job is enqueued
- **AND** the response is `202 Accepted`

#### Scenario: Already-pending account returns 202 without re-enqueuing

- **WHEN** `DELETE /internal/accounts/:id` is called with a Kinde subject whose `account_state` is `'pending_deletion'`
- **THEN** the response is `202 Accepted`
- **AND** no additional job is enqueued

#### Scenario: Purged account returns 202 without re-enqueuing

- **WHEN** `DELETE /internal/accounts/:id` is called with a Kinde subject whose `account_state` is `'purged'`
- **THEN** the response is `202 Accepted`
- **AND** no additional job is enqueued

#### Scenario: Unprovisioned account is scheduled for Kinde-only deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a Kinde subject that has no row in `users`
- **THEN** an `AccountWipeArgs{IDPSubject: ...}` job is enqueued
- **AND** the response is `202 Accepted`

#### Scenario: Missing or invalid internal secret is rejected

- **WHEN** `DELETE /internal/accounts/:id` is called without a valid `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`
- **AND** no job is enqueued
