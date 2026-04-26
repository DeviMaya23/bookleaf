## ADDED Requirements

### Requirement: Image GORM Struct

The system SHALL define an `Image` GORM struct in `internal/domain/image.go` representing an uploaded asset owned by a user.

Fields:
- `ID` — UUID primary key
- `UserID` — FK to users table (owner)
- `FolderID` — FK to folders table (nullable; nil means root)
- `Title` — display name, required
- `SourceURL` — original source URL the image was saved from (nullable)
- `R2Path` — path of the full-size image within the user's R2 bucket, required
- `ThumbnailPath` — path of the generated thumbnail within the user's R2 bucket, nullable
- `MIMEType` — MIME type string (e.g. `image/jpeg`), required
- `VisionLabels` — JSON-serialised array of labels from Google Vision (nullable, BYOV only)
- `CreatedAt`, `UpdatedAt` — GORM timestamps

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

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `images` table is dropped without error
