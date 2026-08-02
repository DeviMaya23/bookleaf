## MODIFIED Requirements

### Requirement: DELETE /me Endpoint

The system SHALL expose a protected `DELETE /me` endpoint that permanently deletes the authenticated user's account: all owned app data, associated R2 objects, and their Kinde identity. The endpoint SHALL require a valid JWT.

The endpoint SHALL delegate to `MarkForDeletion`. All deletion work proceeds asynchronously. The response SHALL be `202 Accepted`. Login is blocked synchronously before the response is returned.

#### Scenario: Authenticated user deletes their account

- **WHEN** an authenticated user sends `DELETE /me`
- **THEN** the response is `202 Accepted`
- **AND** the user's `account_state` is `'pending_deletion'` before the response is returned

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `DELETE /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no deletion occurs

### Requirement: Account Data Wipe Transaction

The system SHALL delete all of a user's owned data in a single database transaction, in an order that satisfies existing foreign key constraints:

1. Clear `parent_id` on all of the user's folders (so the self-referential folder FK does not block deletion regardless of nesting depth)
2. Hard-delete all of the user's images, including soft-deleted (trashed) images, recording each image's `r2_path` and `thumbnail_path`
3. Hard-delete all of the user's folders
4. Hard-delete all of the user's tags
5. Hard-delete all of the user's pending uploads, recording each one's `r2_path`

The `users` row SHALL NOT be modified by this transaction. The `account_state` transition happens before this transaction, via `MarkForDeletion`.

#### Scenario: All owned data is removed in one transaction

- **WHEN** the account data wipe transaction runs for a user with images, folders, tags, and pending uploads
- **THEN** none of the user's rows remain in `images`, `folders`, `tags`, `image_folders`, `image_tags`, or `pending_uploads`
- **AND** the user's row in `users` still exists with `account_state = 'pending_deletion'`

#### Scenario: Nested folders are fully removed

- **WHEN** the account data wipe transaction runs for a user with nested folders (a folder with a non-null `parent_id`)
- **THEN** all of the user's folders are deleted regardless of their nesting depth
- **AND** no foreign key violation occurs

#### Scenario: Trashed images are included in the wipe

- **WHEN** the account data wipe transaction runs for a user who has soft-deleted (trashed) images
- **THEN** those trashed images are hard-deleted along with non-trashed images
- **AND** their `r2_path` and `thumbnail_path` are recorded for R2 cleanup

### Requirement: R2 Cleanup Enqueued After Wipe

After the account data wipe transaction commits, the system SHALL enqueue one `R2DeleteArgs` job (using the existing R2 delete worker) for each image's `r2_path`/`thumbnail_path` and each pending upload's `r2_path` recorded during the wipe.

#### Scenario: R2 delete jobs are enqueued for deleted images

- **WHEN** the account data wipe transaction commits for a user who owned images
- **THEN** an `R2DeleteArgs` job is enqueued for each image's `r2_path` and `thumbnail_path` (if present)

#### Scenario: R2 delete jobs are enqueued for pending uploads

- **WHEN** the account data wipe transaction commits for a user who had pending uploads
- **THEN** an `R2DeleteArgs` job is enqueued for each pending upload's `r2_path`

### Requirement: Kinde Management API Client

The system SHALL provide a Kinde Management API client (`internal/kinde`) that authenticates via M2M client-credentials and exposes `DeleteUser(ctx, kindeUserID) error` and `DeleteUserSessions(ctx, kindeUserID) error` methods.

The client SHALL fetch an M2M access token using configured client credentials and cache it in memory, refetching when the cached token is at or near expiry. `DeleteUser` SHALL call `DELETE /api/v1/user?id=<kindeUserID>&is_delete_profile=true` using the cached token. `DeleteUserSessions` SHALL call the Kinde Management API session revocation endpoint for the given user; a response indicating the user has no sessions or does not exist SHALL be treated as success.

#### Scenario: M2M token is cached and reused

- **WHEN** `DeleteUser` is called multiple times before the cached M2M token expires
- **THEN** only one token-fetch request is made to Kinde across those calls

#### Scenario: Expired token is refetched

- **WHEN** `DeleteUser` is called after the cached M2M token has expired
- **THEN** the client fetches a new M2M token before calling the delete-user endpoint

#### Scenario: DeleteUserSessions treats already-gone user as success

- **WHEN** `DeleteUserSessions` is called and Kinde responds indicating the user has no sessions or does not exist
- **THEN** the method returns `nil`

### Requirement: Pending Account Deletion Lockout

The auth middleware SHALL reject requests from users whose `account_state` is not `'active'` with `401 Unauthorized`, regardless of whether their Kinde session/token is otherwise valid. This covers both `'pending_deletion'` and `'purged'` states.

#### Scenario: Request from a pending-deletion user is rejected

- **WHEN** an authenticated request is made by a user whose `account_state` is `'pending_deletion'`
- **THEN** the response is `401 Unauthorized`
- **AND** the request does not reach the handler

#### Scenario: Request from a purged user is rejected

- **WHEN** an authenticated request is made by a user whose `account_state` is `'purged'`
- **THEN** the response is `401 Unauthorized`
- **AND** the request does not reach the handler

## REMOVED Requirements

### Requirement: Account Kinde Deletion Job

**Reason**: Replaced by `AccountWipeJob` (defined in `account-tombstone`), which combines DB wipe and Kinde deletion into a single idempotent job. The separate `AccountKindeDeletionArgs` job no longer exists.

**Migration**: Any `account_kinde_deletion` jobs remaining in River's queue at deploy time should be drained or cancelled before the new code is deployed.

### Requirement: Reconciliation Sweep for Stuck Tombstones

**Reason**: Replaced by `AccountWipeReconcileWorker` (defined in `account-tombstone`), which queries `account_state = 'pending_deletion'` and enqueues `AccountWipeArgs` — covering both the crash-recovery case and the case where the initial enqueue was lost.

**Migration**: No data migration needed; the new reconcile worker reads from the same `users` table using the new `account_state` column.

### Requirement: Booklet Notification Enqueued After Wipe

**Reason**: Moved into `AccountWipeJob` as step 6 (enqueue `BookletUserDeletionArgs` if `bookletClient` is configured). The requirement is now part of the `AccountWipeJob` spec in `account-tombstone`.

**Migration**: No change to `BookletUserDeletionArgs` itself; only the point of enqueue changes.
