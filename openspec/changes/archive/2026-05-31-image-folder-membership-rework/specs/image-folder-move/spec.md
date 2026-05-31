## ADDED Requirements

### Requirement: POST /images/:id/move-folder — Move Image Between Folders

The system SHALL expose a `POST /images/:id/move-folder` endpoint on the protected route group that atomically removes an image from a source folder and adds it to a destination folder.

Request body:
```json
{
  "from_folder_id": "uuid | null",
  "to_folder_id": "uuid | null"
}
```

- Both fields are required in the request body. Neither may be absent.
- If `from_folder_id` is non-null: the row `(imageID, from_folder_id)` SHALL be deleted from `image_folders`. If no such row exists, this is a no-op (not an error).
- If `to_folder_id` is non-null: a row `(imageID, to_folder_id)` SHALL be inserted with a fracdex position appended to the end of that folder. If the row already exists, it SHALL be updated (upsert).
- If `from_folder_id == to_folder_id` (including both null): the operation is a no-op and SHALL return 200 without touching the database.
- No other `image_folders` rows for this image are affected.
- The image MUST be owned by the authenticated user.
- Returns `404 Not Found` if the image does not exist or belongs to another user.
- Returns `400 Bad Request` if the body is malformed.
- Returns `200 OK` with the updated image on success.

#### Scenario: Image is moved from one folder to another

- **WHEN** an authenticated `POST /images/:id/move-folder` is made with `{ "from_folder_id": "A", "to_folder_id": "B" }`
- **THEN** the response is `200 OK`
- **AND** the `(imageID, A)` row is deleted from `image_folders`
- **AND** a `(imageID, B)` row is inserted with a valid fracdex position
- **AND** any other `image_folders` rows for this image (e.g. folder C) are untouched

#### Scenario: Image is unfiled (moved to unsorted)

- **WHEN** an authenticated `POST /images/:id/move-folder` is made with `{ "from_folder_id": "A", "to_folder_id": null }`
- **THEN** the response is `200 OK`
- **AND** the `(imageID, A)` row is deleted from `image_folders`
- **AND** no new row is inserted

#### Scenario: Move to same folder is a no-op

- **WHEN** an authenticated `POST /images/:id/move-folder` is made with `from_folder_id` equal to `to_folder_id`
- **THEN** the response is `200 OK`
- **AND** no rows in `image_folders` are modified

#### Scenario: Image not found or not owned

- **WHEN** an authenticated `POST /images/:id/move-folder` is made for a non-existent or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `POST /images/:id/move-folder` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: MoveImageFolder Usecase Method

The `ImageUsecase` interface SHALL include a `MoveImageFolder(ctx context.Context, imageID uuid.UUID, userID string, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error` method.

The usecase SHALL:
1. Fetch the existing image by `imageID` and `userID`; return `gorm.ErrRecordNotFound` if not found
2. If `fromFolderID == toFolderID` (pointer equality or both nil): return nil immediately
3. Delegate to `ImageRepository.MoveImageFolder`

#### Scenario: Ownership check fails

- **WHEN** `MoveImageFolder` is called for an image that does not exist or is owned by a different user
- **THEN** the method returns `gorm.ErrRecordNotFound`

#### Scenario: Successful move delegates to repository

- **WHEN** `MoveImageFolder` is called with valid ownership and distinct from/to folders
- **THEN** `ImageRepository.MoveImageFolder` is called with the same arguments
- **AND** no error is returned on success

---

### Requirement: MoveImageFolder Repository Method

The `ImageRepository` interface SHALL include a `MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error` method.

The implementation SHALL execute within a single database transaction:
1. If `fromFolderID` is non-nil: `DELETE FROM image_folders WHERE image_id = ? AND folder_id = ?`. No error if row does not exist.
2. If `toFolderID` is non-nil: compute fracdex position via `fracdex.KeyBetween(maxPosition, "")` where `maxPosition` is the current max position in `toFolderID` (empty string if folder has no images), then upsert `(imageID, toFolderID, position)`.

#### Scenario: Move executes both delete and insert atomically

- **WHEN** `MoveImageFolder` is called with non-nil `fromFolderID` and non-nil `toFolderID`
- **THEN** the `(imageID, fromFolderID)` row is deleted
- **AND** a `(imageID, toFolderID)` row is inserted with a valid fracdex position
- **AND** both operations occur within the same transaction

#### Scenario: Unfile removes specific row only

- **WHEN** `MoveImageFolder` is called with non-nil `fromFolderID` and nil `toFolderID`
- **THEN** only the `(imageID, fromFolderID)` row is deleted
- **AND** no new row is inserted

---

### Requirement: MoveImageFolder Handler Unit Tests

The system SHALL have unit tests for the `MoveImageFolder` handler covering at minimum one success scenario and one failure scenario.

#### Scenario: Handler unit tests cover success and failure

- **WHEN** the `MoveImageFolder` handler is tested with a mock usecase
- **THEN** the success case asserts `200 OK`
- **AND** the failure case asserts `404 Not Found` when the usecase returns `gorm.ErrRecordNotFound`

---

### Requirement: MoveImageFolder Usecase Unit Tests

The system SHALL have unit tests for the `MoveImageFolder` usecase covering at minimum one success scenario and one failure scenario.

#### Scenario: Usecase unit tests cover success and failure

- **WHEN** the `MoveImageFolder` usecase is tested with a mocked repository
- **THEN** the success case asserts the repository method is called with the correct arguments
- **AND** the failure case (image not found) asserts the error is propagated without calling the repository

---

### Requirement: Bruno file for move-folder endpoint

The Bruno collection SHALL include a request file for `POST /images/:id/move-folder`.

#### Scenario: Bruno file exists for move-folder

- **WHEN** the Bruno collection is opened
- **THEN** a request named `Move Image to Folder` exists under the images folder
- **AND** it targets `POST /images/:id/move-folder` with a JSON body containing `from_folder_id` and `to_folder_id`
