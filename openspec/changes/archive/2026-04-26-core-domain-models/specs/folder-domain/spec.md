## ADDED Requirements

### Requirement: Folder GORM Struct

The system SHALL define a `Folder` GORM struct in `internal/domain/folder.go` representing a user-owned grouping of images that supports arbitrary nesting.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; will be Clerk's user ID string (`user_id`), required
- `ParentID` — self-referencing FK to `folders(id)` (nullable; nil means top-level folder) (`parent_id`)
- `Name` — display name, required (`name`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

#### Scenario: Folder struct supports nesting

- **WHEN** the Go package is compiled
- **THEN** `Folder` has a nullable `ParentID` UUID field referencing the same `folders` table

#### Scenario: Top-level folder has no parent

- **WHEN** a `Folder` is created with `ParentID` set to nil
- **THEN** it is treated as a root-level folder owned by the user

### Requirement: Folders DB Migration

The system SHALL include a `golang-migrate` SQL migration that creates the `folders` table before the `images` table (images depends on folders).

#### Scenario: Migration creates folders table

- **WHEN** migrations are applied to a fresh database
- **THEN** the `folders` table exists with columns matching the `Folder` struct
- **AND** `user_id` has a NOT NULL FK constraint referencing `users(id)`
- **AND** `parent_id` has a nullable self-referencing FK constraint on `folders(id)`
- **AND** `deleted_at` is a nullable timestamp column (soft delete)

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `folders` table is dropped without error
