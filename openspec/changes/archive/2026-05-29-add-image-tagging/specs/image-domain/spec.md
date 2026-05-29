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

Associations:
- `User User` — belongs-to
- `Folder *Folder` — belongs-to (nullable)
- `Tags []Tag` — many-to-many via `image_tags` join table: `gorm:"many2many:image_tags;foreignKey:ID;joinForeignKey:ImageID;References:ID;joinReferences:TagID"`

#### Scenario: Image struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `gorm:"primaryKey"` UUID field and FK references to `users` and `folders`

#### Scenario: Image struct includes Tags association

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `Tags []Tag` field with a GORM many2many tag referencing the `image_tags` join table

#### Scenario: Image struct compiles with all metadata fields

- **WHEN** the Go package is compiled
- **THEN** `Image` has nullable pointer fields `Description *string`, `Width *int`, `Height *int`, and `FileSize *int64` with correct GORM column tags

#### Scenario: Image struct includes IsUploaded field

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `IsUploaded bool` field with GORM tag `column:is_uploaded;not null;default:false`

### Requirement: Image Repository Interface

The system SHALL define an `ImageRepository` interface in `internal/usecase/` that the SQL repository implements.

Methods:
- `Create(ctx, image *domain.Image) (*domain.Image, error)`
- `List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns non-deleted images ordered by `(created_at DESC, id DESC)`; fetches `limit + 1` rows so the caller can detect next-page existence; `folderID` nil means no folder filter; `tagID` nil means no tag filter; `unfiled` true limits to images with no folder; images are returned with their `Tags` preloaded
- `GetByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns non-deleted images only; result has `Tags` preloaded
- `GetDeletedByID(ctx, id uuid.UUID, userID string) (*domain.Image, error)` — returns soft-deleted images only
- `UpdateThumbnailPath(ctx, id uuid.UUID, thumbnailPath string) error` — updates `thumbnail_path`; no ownership check (called internally by goroutine)
- `UpdateAILabels(ctx, id uuid.UUID, labels json.RawMessage) error`
- `Update(ctx, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` — selectively updates the supplied fields for the image owned by `userID`; result has `Tags` preloaded
- `SoftDelete(ctx, id uuid.UUID, userID string) error`
- `Restore(ctx, id uuid.UUID, userID string) error`
- `ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)` — returns soft-deleted images ordered by `(deleted_at ASC, id ASC)`; fetches `limit + 1` rows; `cursor` nil means first page
- `CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)` — counts non-deleted images belonging to the given folder
- `ListStaleUploads(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)`
- `HardDelete(ctx context.Context, id uuid.UUID, userID string) error`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements `usecase.ImageRepository` without compilation errors

#### Scenario: List preloads tags for each image

- **WHEN** `List` is called and images have associated tags
- **THEN** each returned `domain.Image` has its `Tags` slice populated

#### Scenario: List filters by tagID when provided

- **WHEN** `List` is called with a non-nil `tagID`
- **THEN** only images associated with that tag are returned

#### Scenario: GetByID preloads tags

- **WHEN** `GetByID` is called for an image that has tags
- **THEN** the returned `domain.Image` has its `Tags` slice populated
