## MODIFIED Requirements

### Requirement: Folder GORM Struct

The system SHALL define a `Folder` GORM struct in `internal/domain/folder.go` representing a user-owned grouping of images that supports arbitrary nesting.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; Kinde user ID string (`user_id`), required
- `ParentID` — self-referencing FK to `folders(id)` (nullable; nil means top-level folder) (`parent_id`)
- `Name` — display name, required (`name`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)

`DeletedAt` is removed. Folders use hard delete only.

#### Scenario: Folder struct supports nesting

- **WHEN** the Go package is compiled
- **THEN** `Folder` has a nullable `ParentID` UUID field referencing the same `folders` table

#### Scenario: Top-level folder has no parent

- **WHEN** a `Folder` is created with `ParentID` set to nil
- **THEN** it is treated as a root-level folder owned by the user

#### Scenario: Folder struct has no soft-delete field

- **WHEN** the Go package is compiled
- **THEN** `Folder` does NOT have a `DeletedAt` field
- **AND** GORM does NOT append `deleted_at IS NULL` to queries on `Folder`

---

## MODIFIED Requirements

### Requirement: Folders DB Migration

The system SHALL include a `golang-migrate` SQL migration that creates the `folders` table (migration 000002) and a subsequent migration (000005) that removes `deleted_at` from `folders` and changes the `images.folder_id` FK constraint from `ON DELETE SET NULL` to `ON DELETE RESTRICT`.

Migration 000005 (`remove_folders_soft_delete`) is the authoritative change for both the soft-delete removal and the FK tightening on images.

#### Scenario: Migration creates folders table

- **WHEN** all migrations are applied to a fresh database
- **THEN** the `folders` table exists with columns: `id`, `user_id`, `parent_id`, `name`, `created_at`, `updated_at`
- **AND** `user_id` has a NOT NULL FK constraint referencing `users(id)`
- **AND** `parent_id` has a nullable self-referencing FK constraint on `folders(id)` with `ON DELETE RESTRICT`
- **AND** the `deleted_at` column does NOT exist

#### Scenario: Migration 000005 removes deleted_at and tightens images FK

- **WHEN** migration 000005 up is applied
- **THEN** `deleted_at` column and `idx_folders_deleted_at` index are dropped from `folders`
- **AND** the `fk_images_folder` constraint on `images.folder_id` is changed to `ON DELETE RESTRICT`

#### Scenario: Migration 000005 is reversible

- **WHEN** migration 000005 down is applied
- **THEN** `deleted_at` column and `idx_folders_deleted_at` index are restored to `folders`
- **AND** the `fk_images_folder` constraint is restored to `ON DELETE SET NULL`
