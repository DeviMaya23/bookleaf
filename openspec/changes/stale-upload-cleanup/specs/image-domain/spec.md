## MODIFIED Requirements

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
- `IsUploaded` — boolean flag, `false` when record is created by `InitiateUpload`, set to `true` by `CompleteUpload`; used to detect abandoned upload attempts (`is_uploaded`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

#### Scenario: Image struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `gorm:"primaryKey"` UUID field and FK references to `users` and `folders`

#### Scenario: Image struct compiles with all metadata fields

- **WHEN** the Go package is compiled
- **THEN** `Image` has nullable pointer fields `Description *string`, `Width *int`, `Height *int`, and `FileSize *int64` with correct GORM column tags

#### Scenario: Image struct includes IsUploaded field

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `IsUploaded bool` field with GORM tag `column:is_uploaded;not null;default:false`

## ADDED Requirements

### Requirement: is_uploaded DB Migration

The system SHALL include a `golang-migrate` SQL migration that adds the `is_uploaded` column to the existing `images` table.

Column:
- `is_uploaded boolean NOT NULL DEFAULT false`

#### Scenario: Migration adds is_uploaded column

- **WHEN** the up migration is applied
- **THEN** the `images` table gains an `is_uploaded` boolean column with `NOT NULL DEFAULT false`
- **AND** existing rows have `is_uploaded = false`

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `is_uploaded` column is dropped from `images` without error
