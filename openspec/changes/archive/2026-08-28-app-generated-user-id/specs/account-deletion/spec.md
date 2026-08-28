## MODIFIED Requirements

### Requirement: Kinde Management API Client

The system SHALL provide a Kinde Management API client (`internal/kinde`) that authenticates via M2M client-credentials and exposes `DeleteUser(ctx, idpSubject string) error` and `DeleteUserSessions(ctx, idpSubject string) error` methods. Both methods accept the IdP-issued subject (Kinde subject), NOT the internal user UUID.

The client SHALL fetch an M2M access token using configured client credentials and cache it in memory, refetching when the cached token is at or near expiry. `DeleteUser` SHALL call `DELETE /api/v1/user?id=<idpSubject>&is_delete_profile=true` using the cached token. `DeleteUserSessions` SHALL call the Kinde Management API session revocation endpoint for the given subject; a response indicating the user has no sessions or does not exist SHALL be treated as success.

#### Scenario: M2M token is cached and reused

- **WHEN** `DeleteUser` is called multiple times before the cached M2M token expires
- **THEN** only one token-fetch request is made to Kinde across those calls

#### Scenario: Expired token is refetched

- **WHEN** `DeleteUser` is called after the cached M2M token has expired
- **THEN** the client fetches a new M2M token before calling the delete-user endpoint

#### Scenario: DeleteUserSessions treats already-gone user as success

- **WHEN** `DeleteUserSessions` is called and Kinde responds indicating the user has no sessions or does not exist
- **THEN** the method returns `nil`

---

### Requirement: Account Wipe Job

The system SHALL define an `AccountWipeArgs` job (kind: `account_wipe`, `MaxAttempts: 5`) carrying `IDPSubject string` — the Kinde subject of the user to wipe. The worker SHALL call `WipeAccount(ctx, idpSubject string)`.

`WipeAccount` SHALL begin by calling `GetByIDPSubject(ctx, idpSubject)`:
- **User found**: perform the full wipe — DB transaction (clear folder parents, hard-delete images, folders, tags, pending uploads), enqueue R2 delete jobs, call `kinde.DeleteUserSessions(ctx, idpSubject)`, call `kinde.DeleteUser(ctx, idpSubject)`, mark the user purged, enqueue `BookletUserDeletionArgs` if Booklet is configured.
- **User not found** (unprovisioned): call `kinde.DeleteUserSessions(ctx, idpSubject)` and `kinde.DeleteUser(ctx, idpSubject)` only. No DB operations.

#### Scenario: Full wipe runs when user exists

- **WHEN** `WipeAccount` is called with an idp_subject that matches a user row
- **THEN** all of the user's images, folders, tags, and pending uploads are hard-deleted in a transaction
- **AND** `kinde.DeleteUser` is called with the idp_subject

#### Scenario: Kinde-only wipe runs when user row is absent

- **WHEN** `WipeAccount` is called with an idp_subject that has no matching row in `users`
- **THEN** `kinde.DeleteUser` is called with the idp_subject
- **AND** no DB operations are attempted

#### Scenario: MarkForDeletion enqueues AccountWipeArgs with idp_subject

- **WHEN** `MarkForDeletion` is called for an active provisioned user
- **THEN** an `AccountWipeArgs{IDPSubject: user.IDPSubject}` job is enqueued
