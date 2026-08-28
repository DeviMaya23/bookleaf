## Purpose
Define the persistent user model keyed by Kinde user ID and the database migrations required for user state.

## Requirements
### Requirement: User GORM Struct

The system SHALL define a `User` GORM struct in `internal/domain/user.go` representing an authenticated user with an app-generated UUID primary key and a separate field for the IdP-issued subject.

Fields (all DB columns use snake_case):
- `ID` — app-generated UUID, `UUID` primary key (`id`); e.g. `550e8400-e29b-41d4-a716-446655440000`
- `IDPSubject` — IdP-issued subject string, `TEXT NOT NULL UNIQUE` (`idp_subject`); e.g. `kp_abc123`
- `VisionEnabled` — boolean flag indicating whether the user has opted into AI organising (`vision_enabled`); defaults to `false`
- `AICategorisationEnabled` — boolean flag indicating whether AI image categorisation is enabled for the user (`ai_categorisation_enabled`); defaults to `false`
- `AccountState` — account lifecycle state (`account_state`); defaults to `'active'`
- `PurgedAt` — nullable timestamp set when the account is purged (`purged_at`)
- `FolderIconsEnabled` — boolean flag indicating whether folder/system-entry icons are displayed in the sidebar (`folder_icons_enabled`); defaults to `true`
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

#### Scenario: User struct uses app-generated UUID as primary key

- **WHEN** the Go package is compiled
- **THEN** `User` has a `uuid.UUID` `ID` field tagged `gorm:"type:uuid;primaryKey"`
- **AND** there is no string field holding the Kinde subject as the primary key

#### Scenario: User struct has IDPSubject field

- **WHEN** the Go package is compiled
- **THEN** `User` has a `string` `IDPSubject` field tagged `gorm:"column:idp_subject;not null;uniqueIndex"`

### Requirement: Users DB Migration

The system SHALL include a `golang-migrate` SQL migration that creates the `users` table before `folders` and `images` (both depend on it).

#### Scenario: Migration creates users table

- **WHEN** migrations are applied to a fresh database
- **THEN** the `users` table exists with `id TEXT PRIMARY KEY`, `created_at`, `updated_at`, and `deleted_at` columns

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `users` table is dropped without error

### Requirement: vision_enabled DB Migration

The system SHALL include a `golang-migrate` SQL migration (`000004_add_vision_enabled_to_users`) that adds the `vision_enabled` column to the existing `users` table.

- Up: `ALTER TABLE users ADD COLUMN vision_enabled BOOLEAN NOT NULL DEFAULT false`
- Down: `ALTER TABLE users DROP COLUMN vision_enabled`

#### Scenario: Migration adds column with safe default

- **WHEN** the up migration is applied to a database with existing users
- **THEN** the `users` table has a `vision_enabled` column of type `BOOLEAN NOT NULL`
- **AND** all existing rows have `vision_enabled = false`

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `vision_enabled` column is dropped without error

### Requirement: ai_categorisation_enabled DB Migration

The system SHALL include a `golang-migrate` SQL migration (`000019_add_ai_categorisation_enabled_to_users`) that adds the `ai_categorisation_enabled` column to the existing `users` table.

- Up: `ALTER TABLE users ADD COLUMN ai_categorisation_enabled BOOLEAN NOT NULL DEFAULT false`
- Down: `ALTER TABLE users DROP COLUMN ai_categorisation_enabled`

#### Scenario: Migration adds column with safe default

- **WHEN** the up migration is applied to a database with existing users
- **THEN** the `users` table has an `ai_categorisation_enabled` column of type `BOOLEAN NOT NULL`
- **AND** all existing rows have `ai_categorisation_enabled = false`

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `ai_categorisation_enabled` column is dropped without error

### Requirement: pending_kinde_deletion DB Migration

The system SHALL include a `golang-migrate` SQL migration (`000014_add_pending_kinde_deletion_to_users`) that adds the `pending_kinde_deletion` column to the existing `users` table.

- Up: `ALTER TABLE users ADD COLUMN pending_kinde_deletion BOOLEAN NOT NULL DEFAULT false`
- Down: `ALTER TABLE users DROP COLUMN pending_kinde_deletion`

#### Scenario: Migration adds column with safe default

- **WHEN** the up migration is applied to a database with existing users
- **THEN** the `users` table has a `pending_kinde_deletion` column of type `BOOLEAN NOT NULL`
- **AND** all existing rows have `pending_kinde_deletion = false`

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `pending_kinde_deletion` column is dropped without error

### Requirement: folder_icons_enabled DB Migration

The system SHALL include a `golang-migrate` SQL migration (`000018_add_folder_icons_enabled_to_users`) that adds the `folder_icons_enabled` column to the existing `users` table.

- Up: `ALTER TABLE users ADD COLUMN folder_icons_enabled BOOLEAN NOT NULL DEFAULT true`
- Down: `ALTER TABLE users DROP COLUMN folder_icons_enabled`

#### Scenario: Migration adds column with safe default

- **WHEN** the up migration is applied to a database with existing users
- **THEN** the `users` table has a `folder_icons_enabled` column of type `BOOLEAN NOT NULL`
- **AND** all existing rows have `folder_icons_enabled = true`

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `folder_icons_enabled` column is dropped without error
