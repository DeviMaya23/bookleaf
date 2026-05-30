## ADDED Requirements

### Requirement: PendingUpload Domain Struct

The system SHALL define a `PendingUpload` GORM struct in `internal/domain/pending_upload.go` representing an upload-in-progress record that has not yet been committed to the `images` table.

Fields (all DB columns use snake_case):
- `ID` — UUID primary key; this same UUID becomes the committed `images.id` (`id`)
- `UserID` — FK to users table (`user_id`)
- `Title` — display name, required (`title`)
- `Description` — user-supplied annotation, nullable (`description`)
- `SourceURL` — original source URL the image was saved from, nullable (`source_url`)
- `R2Path` — path of the full-size image within the user's R2 bucket; computed at initiation (`r2_path`)
- `MIMEType` — MIME type string, required; used to set `Content-Type` on the presigned PUT URL (`mime_type`)
- `FolderID` — nullable FK to folders table with `ON DELETE SET NULL`; if non-nil, `SetImageFolder` will be called during `CompleteUpload` (`folder_id`)
- `CreatedAt` — GORM timestamp; used by the stale cleaner to identify abandoned records (`created_at`)

```go
type PendingUpload struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
    UserID      string     `gorm:"column:user_id;type:text;not null"`
    Title       string     `gorm:"column:title;not null"`
    Description *string    `gorm:"column:description"`
    SourceURL   *string    `gorm:"column:source_url"`
    R2Path      string     `gorm:"column:r2_path;not null"`
    MIMEType    string     `gorm:"column:mime_type;not null"`
    FolderID    *uuid.UUID `gorm:"column:folder_id"`
    CreatedAt   time.Time  `gorm:"column:created_at"`
}
```

#### Scenario: PendingUpload struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `PendingUpload` has a UUID primary key field and string `UserID`, `Title`, `R2Path`, `MIMEType` fields
- **AND** `Description`, `SourceURL`, and `FolderID` are nullable pointer fields

---

### Requirement: pending_uploads DB Migration

The system SHALL include a `golang-migrate` SQL migration (`000011`) that:
1. Hard-deletes any existing `images` rows where `is_uploaded = false` (purges limbo records before structural changes)
2. Drops the `is_uploaded` column from `images`
3. Creates the `pending_uploads` table

Migration SQL (in order):

```sql
-- 1. Purge in-flight limbo rows
DELETE FROM images WHERE is_uploaded = false;

-- 2. Drop the flag
ALTER TABLE images DROP COLUMN is_uploaded;

-- 3. Create pending_uploads
CREATE TABLE pending_uploads (
    id          UUID        NOT NULL PRIMARY KEY,
    user_id     TEXT        NOT NULL REFERENCES users(id),
    title       TEXT        NOT NULL,
    description TEXT,
    source_url  TEXT,
    r2_path     TEXT        NOT NULL,
    mime_type   TEXT        NOT NULL,
    folder_id   UUID        REFERENCES folders(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_pending_uploads_user_id   ON pending_uploads (user_id);
CREATE INDEX idx_pending_uploads_created_at ON pending_uploads (created_at);
```

Down migration reverses:
1. `DROP TABLE pending_uploads`
2. `ALTER TABLE images ADD COLUMN is_uploaded BOOLEAN NOT NULL DEFAULT true` — default `true` so existing committed rows are treated as uploaded

#### Scenario: Up migration purges limbo rows and drops is_uploaded

- **WHEN** migration 000011 up is applied
- **THEN** all `images` rows with `is_uploaded = false` are permanently deleted
- **AND** the `is_uploaded` column no longer exists on `images`

#### Scenario: Up migration creates pending_uploads table

- **WHEN** migration 000011 up is applied
- **THEN** `pending_uploads` exists with all required columns and a UUID primary key
- **AND** `folder_id` has a nullable FK to `folders(id)` with `ON DELETE SET NULL`

#### Scenario: Down migration is reversible

- **WHEN** migration 000011 down is applied
- **THEN** `pending_uploads` is dropped
- **AND** `images` regains an `is_uploaded` boolean column defaulting to `true`

---

### Requirement: PendingUploadRepository Interface

The system SHALL define a `PendingUploadRepository` interface in `internal/usecase/pending_upload_repository.go`.

Methods:
- `Create(ctx context.Context, p *domain.PendingUpload) (*domain.PendingUpload, error)` — inserts a row; calls `BeforeCreate` hook to assign UUID if nil
- `GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.PendingUpload, error)` — returns the row for the given id and userID; returns an error if not found
- `Delete(ctx context.Context, id uuid.UUID) error` — hard-deletes the row; no-op if not found
- `ListStale(ctx context.Context, olderThan time.Time) ([]*domain.PendingUpload, error)` — returns all rows where `created_at < olderThan`

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `pendingUploadRepository` in `internal/repository/` implements `usecase.PendingUploadRepository` without compilation errors

#### Scenario: GetByID returns error for unknown ID

- **WHEN** `GetByID` is called with an ID that does not exist in `pending_uploads`
- **THEN** an error is returned

#### Scenario: ListStale returns only old records

- **WHEN** `ListStale` is called with a threshold time
- **THEN** only rows whose `created_at` is before that threshold are returned
