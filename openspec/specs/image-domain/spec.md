## ADDED Requirements

### Requirement: Image GORM Struct

The system SHALL define an `Image` GORM struct in `internal/domain/image.go` representing an uploaded asset owned by a user.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; will be Kinde's user ID string (`user_id`)
- `FolderID` — FK to folders table (nullable; nil means root) (`folder_id`)
- `Title` — display name, required (`title`)
- `Description` — user-supplied annotation, nullable (`description`)
- `SourceURL` — original source URL the image was saved from (nullable) (`source_url`)
- `R2Path` — path of the full-size image within the user's R2 bucket, required (`r2_path`)
- `ThumbnailPath` — path of the generated thumbnail within the user's R2 bucket, nullable (`thumbnail_path`)
- `MIMEType` — MIME type string (e.g. `image/jpeg`), required (`mime_type`)
- `Width` — image width in pixels, nullable; populated server-side at upload completion (`width`)
- `Height` — image height in pixels, nullable; populated server-side at upload completion (`height`)
- `FileSize` — file size in bytes, nullable; populated server-side at upload completion (`file_size`)
- `AILabels` — JSON-serialised array of AI-generated labels (nullable, BYOV only) (`ai_labels`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

#### Scenario: Image struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `gorm:"primaryKey"` UUID field and FK references to `users` and `folders`

#### Scenario: Image struct compiles with all metadata fields

- **WHEN** the Go package is compiled
- **THEN** `Image` has nullable pointer fields `Description *string`, `Width *int`, `Height *int`, and `FileSize *int64` with correct GORM column tags

### Requirement: Images DB Migration

The system SHALL include a `golang-migrate` SQL migration that creates the `images` table with all required columns and constraints.

#### Scenario: Migration creates images table

- **WHEN** migrations are applied to a fresh database
- **THEN** the `images` table exists with columns matching the `Image` struct
- **AND** `user_id` has a NOT NULL FK constraint referencing `users(id)`
- **AND** `folder_id` has a nullable FK constraint referencing `folders(id)`
- **AND** `deleted_at` is a nullable timestamp column (soft delete)

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `images` table is dropped without error

### Requirement: Image Metadata Migration

The system SHALL include a `golang-migrate` SQL migration (000006, `add_image_metadata_fields`) that adds the new metadata columns to the existing `images` table.

Columns added:
- `description text` — nullable
- `width integer` — nullable
- `height integer` — nullable
- `file_size bigint` — nullable

#### Scenario: Migration adds columns to images table

- **WHEN** migration 000006 up is applied
- **THEN** the `images` table gains columns `description`, `width`, `height`, and `file_size`, all nullable

#### Scenario: Migration is reversible

- **WHEN** migration 000006 down is applied
- **THEN** the four columns are dropped from `images` without error
