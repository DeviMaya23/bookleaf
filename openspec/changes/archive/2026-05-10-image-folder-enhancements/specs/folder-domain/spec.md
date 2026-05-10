## MODIFIED Requirements

### Requirement: Folder GORM Struct

The system SHALL define a `Folder` GORM struct in `internal/domain/folder.go` representing a user-owned grouping of images that supports arbitrary nesting.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; Kinde user ID string (`user_id`), required
- `ParentID` — self-referencing FK to `folders(id)` (nullable; nil means top-level folder) (`parent_id`)
- `Name` — display name, required (`name`)
- `Description` — user-supplied annotation, nullable (`description`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)

`DeletedAt` is not present. Folders use hard delete only.

#### Scenario: Folder struct includes description field

- **WHEN** the Go package is compiled
- **THEN** `Folder` has a nullable `Description *string` field with a correct GORM column tag

#### Scenario: Top-level folder has no parent

- **WHEN** a `Folder` is created with `ParentID` set to nil
- **THEN** it is treated as a root-level folder owned by the user

#### Scenario: Folder struct has no soft-delete field

- **WHEN** the Go package is compiled
- **THEN** `Folder` does NOT have a `DeletedAt` field
- **AND** GORM does NOT append `deleted_at IS NULL` to queries on `Folder`

## ADDED Requirements

### Requirement: Folder Description Migration

The system SHALL include a `golang-migrate` SQL migration (000007, `add_folder_description`) that adds a `description text` nullable column to the `folders` table.

#### Scenario: Migration adds description column

- **WHEN** migration 000007 up is applied
- **THEN** the `folders` table gains a nullable `description` column of type `text`

#### Scenario: Migration is reversible

- **WHEN** migration 000007 down is applied
- **THEN** the `description` column is dropped from `folders` without error
