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
- `IsUploaded` — boolean flag, `false` when record is created by `InitiateUpload`, set to `true` by `CompleteUpload` (`is_uploaded`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

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

#### Scenario: Image struct includes IsUploaded field

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `IsUploaded bool` field with GORM tag `column:is_uploaded;not null;default:false`
