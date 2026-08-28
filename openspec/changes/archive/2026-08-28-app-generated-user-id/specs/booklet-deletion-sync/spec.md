## MODIFIED Requirements

### Requirement: BookletUserDeletion Job

The system SHALL define a `BookletUserDeletionArgs` job (kind: `booklet_user_deletion`, `MaxAttempts: 10`) carrying `IDPSubject string` — the Kinde subject identifying the user in Booklet. The worker SHALL call `ProcessBookletUserDeletion(ctx, idpSubject string)`, which calls `bookletClient.DeleteUser(ctx, idpSubject)`.

Booklet identifies users by Kinde subject. The internal UUID SHALL NOT be used here.

#### Scenario: Successful Booklet call completes the job

- **WHEN** the `BookletUserDeletionArgs` worker calls `DeleteUser` with the idp_subject and it returns nil
- **THEN** the job completes without error

#### Scenario: Failed Booklet call retries the job

- **WHEN** the `BookletUserDeletionArgs` worker calls `DeleteUser` and it returns an error
- **THEN** the job is retried with River's default backoff up to 10 attempts

---

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

#### Scenario: Unprovisioned account is scheduled for Kinde-only deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a Kinde subject that has no row in `users`
- **THEN** an `AccountWipeArgs{IDPSubject: ...}` job is enqueued
- **AND** the response is `202 Accepted`

#### Scenario: Missing or invalid internal secret is rejected

- **WHEN** `DELETE /internal/accounts/:id` is called without a valid `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`
- **AND** no job is enqueued
