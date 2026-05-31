## MODIFIED Requirements

### Requirement: SetImageFolder Repository Method

The system SHALL add a `SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error` method to `ImageRepository` that is the single write path for assigning or removing a folder from an image.

Behaviour:
- When `folderID` is `nil`: DELETE the row from `image_folders` where `image_id = imageID` (unfile the image). If no row exists, this is a no-op (not an error).
- When `folderID` is non-nil: INSERT a row into `image_folders` with a computed position. If a row for `(imageID, folderID)` already exists, UPDATE its position. Position for a new row is computed using `fracdex.KeyBetween(maxPosition, "")` where `maxPosition` is the current maximum position value in that folder ordered lexicographically (empty string `""` if the folder is empty). The Go dependency `github.com/rocicorp/fracdex` is used for this computation.

#### Scenario: SetImageFolder assigns a folder to an image

- **WHEN** `SetImageFolder` is called with a non-nil `folderID`
- **THEN** a row is inserted into `image_folders` for the given `(imageID, folderID)` pair
- **AND** `position` is set to a valid fracdex key that sorts after all existing positions in that folder

#### Scenario: SetImageFolder assigns first image to an empty folder

- **WHEN** `SetImageFolder` is called with a non-nil `folderID` and the folder currently has no images
- **THEN** the new row receives the fracdex key produced by `fracdex.KeyBetween("", "")` (the initial key)

#### Scenario: SetImageFolder removes a folder from an image

- **WHEN** `SetImageFolder` is called with `folderID == nil`
- **THEN** the row for `imageID` in `image_folders` is deleted
- **AND** the image has no folder membership

#### Scenario: SetImageFolder is a no-op when unfiling an image with no folder

- **WHEN** `SetImageFolder` is called with `folderID == nil` for an image that has no row in `image_folders`
- **THEN** no error is returned

#### Scenario: SetImageFolder upserts if folder already assigned

- **WHEN** `SetImageFolder` is called with a `folderID` that already has a row for that image
- **THEN** the row is updated without error (no duplicate key violation)

---

## ADDED Requirements

### Requirement: UpdateImageFolderPosition Repository Method

The system SHALL add an `UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error` method to `ImageRepository`.

The implementation SHALL issue:
```sql
UPDATE image_folders SET position = ? WHERE image_id = ? AND folder_id = ?
```

- If no row exists for `(imageID, folderID)`, the method SHALL return `gorm.ErrRecordNotFound` (wrapped).
- No ownership check is performed at the repository layer; the usecase is responsible for verifying the image belongs to the authenticated user before calling this method.

#### Scenario: Position updated successfully

- **WHEN** `UpdateImageFolderPosition` is called with a valid `(imageID, folderID)` pair that exists in `image_folders`
- **THEN** `image_folders.position` is updated to the provided value and no error is returned

#### Scenario: Row not found returns error

- **WHEN** `UpdateImageFolderPosition` is called for an `(imageID, folderID)` pair with no row in `image_folders`
- **THEN** a wrapped `gorm.ErrRecordNotFound` error is returned

---

### Requirement: Position Rebalance Migration Script

The system SHALL include a standalone Go program at `cmd/migrate-positions/main.go` that rebalances existing integer placeholder positions to valid fracdex keys.

The script SHALL:
1. Connect to the database using the same `DATABASE_URL` environment variable as the main server
2. Select all `(folder_id, image_id)` pairs from `image_folders`, ordered by `folder_id ASC, position ASC` (lexicographic order; existing values are short integers that sort correctly as strings for this purpose)
3. For each folder group, generate a sequence of fracdex keys using `fracdex.KeyBetween` starting from `""`
4. Update each row's `position` in a single transaction per folder
5. Log progress (folder ID, number of images rebalanced) and exit with code 1 on any error

The script is a one-time operational tool and is not invoked by the main server or any migration pipeline.

#### Scenario: Script rebalances positions for all folders

- **WHEN** the script runs against a database with images in folders having integer positions
- **THEN** every row in `image_folders` has its `position` replaced with a valid fracdex key
- **AND** the relative order within each folder (by previous integer value) is preserved

#### Scenario: Script is idempotent on already-fracdex positions

- **WHEN** the script runs against a database where positions are already valid fracdex keys
- **THEN** the script completes without error (ordering is preserved; keys are regenerated but remain valid)
