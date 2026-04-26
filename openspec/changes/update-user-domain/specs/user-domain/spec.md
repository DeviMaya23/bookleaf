## MODIFIED Requirements

### Requirement: User GORM Struct

The system SHALL define a `User` GORM struct in `internal/domain/user.go` representing an authenticated user managed by Clerk.

Fields (all DB columns use snake_case):
- `ID` — Clerk-generated user ID string, `TEXT` primary key (`id`); e.g. `user_abc123`
- `VisionEnabled` — boolean flag indicating whether the user has opted into AI organising (`vision_enabled`); defaults to `false`
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

No UUID field. Clerk owns the identity layer; the DB stores only the Clerk ID as the natural PK and app-specific state.

#### Scenario: User struct uses Clerk ID as primary key

- **WHEN** the Go package is compiled
- **THEN** `User` has a `string` `ID` field tagged `gorm:"primaryKey"`
- **AND** there is no separate UUID or `ClerkID` field

#### Scenario: User struct includes vision_enabled field

- **WHEN** the Go package is compiled
- **THEN** `User` has a `bool` `VisionEnabled` field tagged with `gorm:"column:vision_enabled;default:false"`

## ADDED Requirements

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
