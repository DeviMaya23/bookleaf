## MODIFIED Requirements

### Requirement: Folder GORM Struct

The system SHALL define a `Folder` GORM struct in `internal/domain/folder.go` representing a user-owned grouping of images that supports arbitrary nesting.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; Kinde user ID string (`user_id`), required
- `ParentID` — self-referencing FK to `folders(id)` (nullable; nil means top-level folder) (`parent_id`)
- `Name` — display name, required (`name`)
- `Description` — user-supplied annotation, nullable (`description`)
- `Icon` — user-selected icon key, nullable; `nil` means the default icon is used (`icon`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)

`DeletedAt` is not present. Folders use hard delete only.

#### Scenario: Folder struct supports nesting

- **WHEN** the Go package is compiled
- **THEN** `Folder` has a nullable `ParentID` UUID field referencing the same `folders` table

#### Scenario: Folder struct includes description field

- **WHEN** the Go package is compiled
- **THEN** `Folder` has a nullable `Description *string` field with a correct GORM column tag

#### Scenario: Folder struct includes icon field

- **WHEN** the Go package is compiled
- **THEN** `Folder` has a nullable `Icon *string` field with a correct GORM column tag (`gorm:"column:icon"`)

#### Scenario: Top-level folder has no parent

- **WHEN** a `Folder` is created with `ParentID` set to nil
- **THEN** it is treated as a root-level folder owned by the user

#### Scenario: Folder struct has no soft-delete field

- **WHEN** the Go package is compiled
- **THEN** `Folder` does NOT have a `DeletedAt` field
- **AND** GORM does NOT append `deleted_at IS NULL` to queries on `Folder`

## ADDED Requirements

### Requirement: Folder Icon Migration

The system SHALL include a `golang-migrate` SQL migration (`000017_add_folder_icon`) that adds an `icon text` nullable column to the `folders` table.

#### Scenario: Migration adds icon column

- **WHEN** migration 000017 up is applied
- **THEN** the `folders` table gains a nullable `icon` column of type `text`

#### Scenario: Migration is reversible

- **WHEN** migration 000017 down is applied
- **THEN** the `icon` column is dropped from `folders` without error

#### Scenario: Existing folders are unaffected by the migration

- **WHEN** migration 000017 up is applied to a database with existing folder rows
- **THEN** all existing rows have `icon = NULL`
