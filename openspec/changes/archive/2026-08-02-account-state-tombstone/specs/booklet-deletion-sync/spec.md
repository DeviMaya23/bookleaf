## MODIFIED Requirements

### Requirement: DELETE /internal/accounts/:id Endpoint

The system SHALL expose a `DELETE /internal/accounts/:id` endpoint on the existing `/internal` route group, protected by the existing `X-Bookleaf-Internal-Secret` middleware. The `:id` parameter is the Kinde user ID.

The handler SHALL delegate to `MarkForDeletion` and always return `202 Accepted` on success. The endpoint SHALL never return `404` — a user ID unknown to Bookleaf's database is still a valid Kinde identity, and Bookleaf SHALL schedule Kinde deletion regardless.

`MarkForDeletion` SHALL treat any user with `account_state != 'active'` as already in the pipeline and return a no-op `202` without re-enqueueing. For an active user, it SHALL set `account_state = 'pending_deletion'` synchronously and best-effort enqueue `AccountWipeArgs` before returning.

For a user ID that does not exist in Bookleaf's `users` table, the handler SHALL enqueue `AccountWipeArgs` directly (Kinde-only cleanup — the wipe job handles missing user rows gracefully) and return `202 Accepted`.

#### Scenario: Active account is marked for deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID that exists with `account_state = 'active'`
- **THEN** the user's `account_state` is set to `'pending_deletion'` before the response is returned
- **AND** an `AccountWipeArgs` job is enqueued for that user ID
- **AND** the response is `202 Accepted`

#### Scenario: Already-pending account returns 202 without re-enqueuing

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID whose `account_state` is `'pending_deletion'`
- **THEN** the response is `202 Accepted`
- **AND** no additional job is enqueued

#### Scenario: Purged account returns 202 without re-enqueuing

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID whose `account_state` is `'purged'`
- **THEN** the response is `202 Accepted`
- **AND** no additional job is enqueued

#### Scenario: Unprovisioned account is scheduled for Kinde-only deletion

- **WHEN** `DELETE /internal/accounts/:id` is called with a user ID that does not exist in the `users` table
- **THEN** an `AccountWipeArgs` job is enqueued for that user ID
- **AND** the response is `202 Accepted`

#### Scenario: Missing or invalid internal secret is rejected

- **WHEN** `DELETE /internal/accounts/:id` is called without a valid `X-Bookleaf-Internal-Secret` header
- **THEN** the response is `401 Unauthorized`
- **AND** no job is enqueued
