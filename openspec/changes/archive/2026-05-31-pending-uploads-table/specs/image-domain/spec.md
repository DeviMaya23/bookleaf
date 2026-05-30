## MODIFIED Requirements

### Requirement: Image GORM Struct

The system SHALL define an `Image` GORM struct in `internal/domain/image.go` representing an uploaded asset owned by a user.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; will be Kinde's user ID string (`user_id`)
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

The `IsUploaded` field is REMOVED. The `images` table no longer has an `is_uploaded` column. Every row in `images` is a committed, valid image — no filter is needed to exclude uncommitted records.

The `FolderID *uuid.UUID` column and `Folder *Folder` belongs-to association are REMOVED. The `images` table no longer has a `folder_id` column.

Associations:
- `User User` — belongs-to
- `ImageFolders []ImageFolder` — has-many via `image_folders` table: `gorm:"foreignKey:ImageID"`
- `Tags []Tag` — many-to-many via `image_tags` join table: `gorm:"many2many:image_tags;foreignKey:ID;joinForeignKey:ImageID;References:ID;joinReferences:TagID"`

#### Scenario: Image struct compiles without FolderID field

- **WHEN** the Go package is compiled
- **THEN** `Image` has no `FolderID` field and no `Folder *Folder` association
- **AND** there is no GORM column tag referencing `folder_id` on the `Image` struct

#### Scenario: Image struct includes ImageFolders association

- **WHEN** the Go package is compiled
- **THEN** `Image` has an `ImageFolders []ImageFolder` field with a GORM `foreignKey:ImageID` tag

#### Scenario: Image struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `gorm:"primaryKey"` UUID field and a FK reference to `users`

#### Scenario: Image struct includes Tags association

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `Tags []Tag` field with a GORM many2many tag referencing the `image_tags` join table

#### Scenario: Image struct compiles with all metadata fields

- **WHEN** the Go package is compiled
- **THEN** `Image` has nullable pointer fields `Description *string`, `Width *int`, `Height *int`, and `FileSize *int64` with correct GORM column tags

#### Scenario: Image struct has no IsUploaded field

- **WHEN** the Go package is compiled
- **THEN** `Image` has no `IsUploaded` field and no `is_uploaded` GORM column tag

## REMOVED Requirements

### Requirement: is_uploaded DB Migration

**Reason**: Replaced by migration 000011 in the `pending-uploads` capability, which drops the `is_uploaded` column as part of introducing the `pending_uploads` table.
**Migration**: See `pending-uploads` spec, Requirement: pending_uploads DB Migration.
