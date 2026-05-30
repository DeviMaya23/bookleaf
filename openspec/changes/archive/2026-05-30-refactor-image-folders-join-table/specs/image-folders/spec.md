## ADDED Requirements

### Requirement: image_folders DB Migration

The system SHALL include a `golang-migrate` SQL migration (`000010`) that creates the `image_folders` join table, backfills data from `images.folder_id`, adds indexes, and drops the `folder_id` column from `images`.

Migration steps (in order):
1. Create `image_folders` table:
```sql
CREATE TABLE image_folders (
  image_id  UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
  folder_id UUID NOT NULL REFERENCES folders(id) ON DELETE CASCADE,
  position  TEXT NOT NULL DEFAULT '',
  PRIMARY KEY (image_id, folder_id)
);
```
2. Create indexes:
```sql
CREATE INDEX idx_image_folders_folder_id ON image_folders (folder_id);
CREATE INDEX idx_image_folders_image_id  ON image_folders (image_id);
CREATE INDEX idx_image_folders_folder_position ON image_folders (folder_id, position);
```
3. Backfill from existing data:
```sql
INSERT INTO image_folders (image_id, folder_id, position)
SELECT id, folder_id,
  ROW_NUMBER() OVER (PARTITION BY folder_id ORDER BY created_at ASC)::TEXT
FROM images
WHERE folder_id IS NOT NULL;
```
4. Drop the old column:
```sql
ALTER TABLE images DROP COLUMN folder_id;
```

Down migration reverses these steps: adds `folder_id` back to `images`, backfills from `image_folders` (lowest position per image), then drops `image_folders`.

#### Scenario: Migration creates image_folders table

- **WHEN** migration 000010 up is applied
- **THEN** the `image_folders` table exists with columns `image_id`, `folder_id`, and `position`
- **AND** both `image_id` and `folder_id` have FK constraints with `ON DELETE CASCADE`
- **AND** `(image_id, folder_id)` is the primary key

#### Scenario: Migration backfills existing folder memberships

- **WHEN** migration 000010 up is applied against a database with images that have a non-null `folder_id`
- **THEN** each such image has a corresponding row in `image_folders`
- **AND** the `position` values within each folder are distinct and ordered by the image's `created_at`

#### Scenario: Migration drops folder_id from images

- **WHEN** migration 000010 up is applied
- **THEN** the `images` table no longer has a `folder_id` column

#### Scenario: Migration is reversible

- **WHEN** migration 000010 down is applied
- **THEN** `images` regains a `folder_id` column populated from `image_folders`
- **AND** the `image_folders` table is dropped

---

### Requirement: ImageFolder Domain Struct

The system SHALL define an `ImageFolder` GORM struct in `internal/domain/image.go` representing a row in the `image_folders` join table.

```go
type ImageFolder struct {
    ImageID  uuid.UUID `gorm:"primaryKey;column:image_id"`
    FolderID uuid.UUID `gorm:"primaryKey;column:folder_id"`
    Position string    `gorm:"column:position;not null;default:''"`
    Folder   Folder    `gorm:"foreignKey:FolderID;references:ID"`
}
```

#### Scenario: ImageFolder struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `ImageFolder` has composite primary key fields `ImageID` and `FolderID` with correct GORM column tags
- **AND** `Position` is a non-null string column

---

### Requirement: SetImageFolder Repository Method

The system SHALL add a `SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error` method to `ImageRepository` that is the single write path for assigning or removing a folder from an image.

Behaviour:
- When `folderID` is `nil`: DELETE the row from `image_folders` where `image_id = imageID` (unfile the image). If no row exists, this is a no-op (not an error).
- When `folderID` is non-nil: INSERT a row into `image_folders` with a computed position. If a row for `(imageID, folderID)` already exists, UPDATE its position. Position for a new row is computed as `(SELECT COALESCE(MAX(position::int), 0) + 1 FROM image_folders WHERE folder_id = folderID)::TEXT`.

#### Scenario: SetImageFolder assigns a folder to an image

- **WHEN** `SetImageFolder` is called with a non-nil `folderID`
- **THEN** a row is inserted into `image_folders` for the given `(imageID, folderID)` pair
- **AND** `position` is set to one greater than the current maximum position within that folder (or `"1"` if the folder is empty)

#### Scenario: SetImageFolder removes a folder from an image

- **WHEN** `SetImageFolder` is called with `folderID == nil`
- **THEN** the row for `imageID` in `image_folders` is deleted
- **AND** the image has no folder membership

#### Scenario: SetImageFolder is a no-op when unfilng an image with no folder

- **WHEN** `SetImageFolder` is called with `folderID == nil` for an image that has no row in `image_folders`
- **THEN** no error is returned

#### Scenario: SetImageFolder upserts if folder already assigned

- **WHEN** `SetImageFolder` is called with a `folderID` that already has a row for that image
- **THEN** the row is updated without error (no duplicate key violation)
