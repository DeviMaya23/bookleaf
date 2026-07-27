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

The system SHALL define a `BookletUserDeletionArgs` job (kind: `booklet_user_deletion`, `MaxAttempts: 10`) and a corresponding worker. The worker SHALL call `bookletClient.DeleteUser(ctx, userID)`. On success the job completes. On failure the job retries with River's default exponential backoff.

#### Scenario: Successful Booklet call completes the job

- **WHEN** the `BookletUserDeletionArgs` worker calls `DeleteUser` and it returns nil
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

The system SHALL expose a `DELETE /internal/accounts/:id` endpoint on the existing `/internal` route group, protected by the existing `X-Bookleaf-Internal-Secret` middleware. The `:id` parameter is the Kinde user ID.

The handler SHALL delegate to `ScheduleAccountDeletion` and always return `202 Accepted` on success. The endpoint SHALL never return `404` — a user ID unknown to Bookleaf's database is still a valid Kinde identity, and Bookleaf SHALL schedule Kinde deletion regardless.

#### Scenario: Active account is scheduled for full deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID that exists in the `users` table with `pending_kinde_deletion = false`
- **THEN** a `DeleteAccountArgs` job is enqueued for that user ID
- **AND** the response is `202 Accepted`

#### Scenario: Already-pending account returns 202 without re-enqueuing

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID whose `pending_kinde_deletion` is `true`
- **THEN** the response is `202 Accepted`
- **AND** no additional job is enqueued

#### Scenario: Unprovisioned account is scheduled for Kinde-only deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID that does not exist in the `users` table
- **THEN** an `AccountKindeDeletionArgs` job is enqueued for that user ID
- **AND** the response is `202 Accepted`

#### Scenario: Missing or invalid internal secret is rejected

- **WHEN** `DELETE /internal/accounts/:id` is called without a valid `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`
- **AND** no job is enqueued
