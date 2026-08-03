## ADDED Requirements

### Requirement: Booklet Notification Enqueued After Wipe

After the account data wipe transaction commits, the system SHALL enqueue a `BookletUserDeletionArgs` job for the deleted user, provided `bookletClient` is configured (non-nil). If `bookletClient` is not configured, the enqueue step SHALL be skipped silently and `DeleteAccount` SHALL still succeed.

#### Scenario: Booklet client configured — notification job is enqueued

- **WHEN** the account data wipe transaction commits and `bookletClient` is configured
- **THEN** a `BookletUserDeletionArgs` job is enqueued for the deleted user ID

#### Scenario: Booklet client not configured — enqueue is skipped

- **WHEN** the account data wipe transaction commits and `bookletClient` is nil
- **THEN** no `BookletUserDeletionArgs` job is enqueued
- **AND** `DeleteAccount` returns nil
