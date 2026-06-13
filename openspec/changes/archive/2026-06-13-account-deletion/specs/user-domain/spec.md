## MODIFIED Requirements

### Requirement: User GORM Struct

The system SHALL define a `User` GORM struct in `internal/domain/user.go` representing an authenticated user managed by Kinde.

Fields (all DB columns use snake_case):
- `ID` — Kinde-generated user ID string, `TEXT` primary key (`id`); e.g. `kp_abc123`
- `VisionEnabled` — boolean flag indicating whether the user has opted into AI organising (`vision_enabled`); defaults to `false`
- `PendingKindeDeletion` — boolean flag indicating the user's app data has been wiped and their Kinde identity deletion is pending (`pending_kinde_deletion`); defaults to `false`
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

No UUID field. Kinde owns the identity layer; the DB stores only the Kinde user ID as the natural PK and app-specific state.

#### Scenario: User struct uses Kinde ID as primary key

- **WHEN** the Go package is compiled
- **THEN** `User` has a `string` `ID` field tagged `gorm:"primaryKey"`
- **AND** there is no separate UUID or `KindeID` field

#### Scenario: User struct includes vision_enabled field

- **WHEN** the Go package is compiled
- **THEN** `User` has a `bool` `VisionEnabled` field tagged with `gorm:"column:vision_enabled;default:false"`

#### Scenario: User struct includes pending_kinde_deletion field

- **WHEN** the Go package is compiled
- **THEN** `User` has a `bool` `PendingKindeDeletion` field tagged with `gorm:"column:pending_kinde_deletion;default:false"`

## ADDED Requirements

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
