## ADDED Requirements

### Requirement: DELETE /me Endpoint

The system SHALL expose a protected `DELETE /me` endpoint that permanently deletes the authenticated user's account: all owned app data, associated R2 objects, and their Kinde identity. The endpoint SHALL require a valid JWT.

On success, all of the user's owned app data (images, folders, tags, pending uploads) SHALL be removed before the response is returned. Removal of R2 objects and the Kinde identity SHALL proceed asynchronously after the response is returned.

#### Scenario: Authenticated user deletes their account

- **WHEN** an authenticated user sends `DELETE /me`
- **THEN** the response is `204 No Content`
- **AND** the user's `images`, `folders`, `tags`, and `pending_uploads` rows no longer exist in the database

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
6. Set `pending_kinde_deletion = true` on the user's row

The `users` row itself SHALL NOT be deleted as part of this transaction.

#### Scenario: All owned data is removed in one transaction

- **WHEN** the account data wipe transaction runs for a user with images, folders, tags, and pending uploads
- **THEN** none of the user's rows remain in `images`, `folders`, `tags`, `image_folders`, `image_tags`, or `pending_uploads`
- **AND** the user's row in `users` still exists with `pending_kinde_deletion = true`

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

### Requirement: Account Kinde Deletion Job

The system SHALL define an `AccountKindeDeletionArgs` job (kind `account_kinde_deletion`, `MaxAttempts: 5`) and a corresponding worker. After the account data wipe transaction commits, the system SHALL enqueue this job for the deleted user.

On each attempt, the worker SHALL call the Kinde Management API to delete the user (`DELETE /api/v1/user?id=<kinde_user_id>&is_delete_profile=true`). A response indicating the user no longer exists in Kinde SHALL be treated as success. On success, the worker SHALL hard-delete the user's row from the `users` table. On any other failure, the job SHALL retry with River's default backoff.

#### Scenario: Successful Kinde deletion removes the user row

- **WHEN** the `AccountKindeDeletionArgs` job successfully deletes the user from Kinde
- **THEN** the user's row in the `users` table is hard-deleted

#### Scenario: Kinde deletion failure retries

- **WHEN** the Kinde Management API call fails with an error other than "user not found"
- **THEN** the job is retried with River's default backoff
- **AND** the user's row in `users` is not deleted

#### Scenario: Already-deleted Kinde user is treated as success

- **WHEN** the Kinde Management API responds indicating the user no longer exists
- **THEN** the worker treats the call as successful
- **AND** the user's row in the `users` table is hard-deleted

### Requirement: Kinde Management API Client

The system SHALL provide a Kinde Management API client (`internal/platform/kinde`) that authenticates via M2M client-credentials and exposes a `DeleteUser(ctx, kindeUserID) error` method.

The client SHALL fetch an M2M access token using configured client credentials and cache it in memory, refetching when the cached token is at or near expiry. `DeleteUser` SHALL call `DELETE /api/v1/user?id=<kindeUserID>&is_delete_profile=true` using the cached token.

#### Scenario: M2M token is cached and reused

- **WHEN** `DeleteUser` is called multiple times before the cached M2M token expires
- **THEN** only one token-fetch request is made to Kinde across those calls

#### Scenario: Expired token is refetched

- **WHEN** `DeleteUser` is called after the cached M2M token has expired
- **THEN** the client fetches a new M2M token before calling the delete-user endpoint

### Requirement: Pending Account Deletion Lockout

The auth middleware SHALL reject requests from users whose `pending_kinde_deletion` flag is `true` with `401 Unauthorized`, regardless of whether their Kinde session/token is otherwise valid.

#### Scenario: Request from a tombstoned user is rejected

- **WHEN** an authenticated request is made by a user whose `users` row has `pending_kinde_deletion = true`
- **THEN** the response is `401 Unauthorized`
- **AND** the request does not reach the handler

### Requirement: Reconciliation Sweep for Stuck Tombstones

The system SHALL register a periodic River job (interval: 24 hours) that finds all users with `pending_kinde_deletion = true` and re-enqueues an `AccountKindeDeletionArgs` job for each.

#### Scenario: Stuck tombstoned users are re-enqueued

- **WHEN** the periodic reconciliation job runs
- **THEN** for every user row with `pending_kinde_deletion = true`, an `AccountKindeDeletionArgs` job is enqueued for that user's ID
