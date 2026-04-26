## ADDED Requirements

### Requirement: Image GORM Struct

The system SHALL define an `Image` GORM struct in `internal/domain/image.go` representing an uploaded asset owned by a user.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; will be Clerk's user ID string (`user_id`)
- `FolderID` — FK to folders table (nullable; nil means root) (`folder_id`)
- `Title` — display name, required (`title`)
- `SourceURL` — original source URL the image was saved from (nullable) (`source_url`)
- `R2Path` — path of the full-size image within the user's R2 bucket, required (`r2_path`)
- `ThumbnailPath` — path of the generated thumbnail within the user's R2 bucket, nullable (`thumbnail_path`)
- `MIMEType` — MIME type string (e.g. `image/jpeg`), required (`mime_type`)
- `AILabels` — JSON-serialised array of AI-generated labels (nullable, BYOV only) (`ai_labels`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

#### Scenario: Image struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `gorm:"primaryKey"` UUID field and FK references to `users` and `folders`

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
