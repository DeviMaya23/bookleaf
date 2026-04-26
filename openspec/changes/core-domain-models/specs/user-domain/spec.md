## ADDED Requirements

### Requirement: User GORM Struct

The system SHALL define a `User` GORM struct in `internal/domain/user.go` representing an authenticated user managed by Clerk.

Fields (all DB columns use snake_case):
- `ID` — Clerk-generated user ID string, `TEXT` primary key (`id`); e.g. `user_abc123`
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

No UUID field. Clerk owns the identity layer; the DB stores only the Clerk ID as the natural PK.

#### Scenario: User struct uses Clerk ID as primary key

- **WHEN** the Go package is compiled
- **THEN** `User` has a `string` `ID` field tagged `gorm:"primaryKey"`
- **AND** there is no separate UUID or `ClerkID` field

### Requirement: Users DB Migration

The system SHALL include a `golang-migrate` SQL migration that creates the `users` table before `folders` and `images` (both depend on it).

#### Scenario: Migration creates users table

- **WHEN** migrations are applied to a fresh database
- **THEN** the `users` table exists with `id TEXT PRIMARY KEY`, `created_at`, `updated_at`, and `deleted_at` columns

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `users` table is dropped without error
