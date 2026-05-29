## ADDED Requirements

### Requirement: Batch upload modal
The system SHALL provide a `BatchUploadModal` component, separate from the single-file `UploadModal`, for uploading multiple images in a single action.

#### Scenario: Modal opens from the toolbar
- **WHEN** the user selects "Upload multiple images" from the toolbar dropdown
- **THEN** the batch upload modal opens

#### Scenario: Modal opens from multi-file drag
- **WHEN** the user drops more than one file onto the main app surface
- **THEN** the batch upload modal opens with those files pre-loaded in the file list

---

### Requirement: Multi-file selection
The batch modal SHALL accept multiple files via a file picker (with `multiple` attribute) and via drag-and-drop onto the modal's drop zone. Both paths populate the same file list.

#### Scenario: File picker allows multi-select
- **WHEN** the user clicks the drop zone in the batch modal
- **THEN** the native file picker opens with multiple selection enabled

#### Scenario: Dragging files onto the modal drop zone adds them
- **WHEN** the user drags one or more files onto the batch modal drop zone
- **THEN** the dropped files are added to the file list

---

### Requirement: File count cap
The system SHALL reject a selection if more than 20 files are chosen. No files SHALL be accepted and an inline error SHALL be displayed. The user must re-select.

#### Scenario: Selecting more than 20 files shows an error
- **WHEN** the user selects more than 20 files
- **THEN** an inline error is shown: "You can upload a maximum of 20 files at a time"
- **AND** no files are added to the file list

#### Scenario: Selecting 20 or fewer files is accepted
- **WHEN** the user selects between 1 and 20 files
- **THEN** the files are added to the file list without an error

---

### Requirement: File type validation
The system SHALL accept only JPEG, PNG, GIF, and WEBP files. Unsupported files SHALL be rejected individually with a per-file "Unsupported type" status badge. Valid files in the same selection SHALL still be accepted.

#### Scenario: Unsupported file type is rejected individually
- **WHEN** the user selects a batch that includes a file with an unsupported type (e.g. PDF)
- **THEN** the unsupported file appears in the list with an "Unsupported type" badge
- **AND** the remaining valid files are accepted

---

### Requirement: File size validation
The system SHALL reject any individual file larger than 50 MB before uploading. Oversized files SHALL appear in the file list with a "Too large" status badge. The remaining valid files SHALL proceed to upload.

#### Scenario: Oversized file is shown with a warning badge
- **WHEN** the file list contains a file larger than 50 MB
- **THEN** that file is shown with a "Too large" status badge
- **AND** it is excluded from the upload queue

#### Scenario: Valid files in the batch upload despite one oversized file
- **WHEN** the file list contains a mix of valid and oversized files
- **THEN** the valid files upload normally
- **AND** the oversized files remain in the list with their badge

---

### Requirement: Upload queue with concurrency limit
The system SHALL upload at most 3 files concurrently. Additional files SHALL remain in a pending state until a slot becomes available.

#### Scenario: At most 3 files upload simultaneously
- **WHEN** the user starts uploading a batch of 6 files
- **THEN** 3 files begin uploading immediately
- **AND** the other 3 show a pending indicator until a slot frees up

---

### Requirement: Per-file upload sequence
For each file in the batch, the system SHALL execute the existing three-step upload sequence: (1) `POST /images` to initiate, (2) `PUT` to the presigned R2 URL, (3) `POST /images/:id/complete`. The `folder_id` SHALL be set to the current folder, or omitted when on a non-folder route. The title SHALL be the filename without extension.

#### Scenario: Each file follows the three-step upload sequence
- **WHEN** a file is uploaded as part of a batch
- **THEN** `POST /images` is called with the file's title (filename without extension), MIME type, and optional `folder_id`
- **AND** the file bytes are PUT to the returned presigned URL
- **AND** `POST /images/:id/complete` is called after the PUT succeeds

#### Scenario: folder_id is included when a folder is active
- **WHEN** the user is on `/folders/:folderId` and uploads a batch
- **THEN** each `POST /images` request includes `folder_id`

#### Scenario: folder_id is omitted on non-folder routes
- **WHEN** the user is not on a folder route and uploads a batch
- **THEN** `folder_id` is not included in any `POST /images` request

---

### Requirement: Per-file progress indicator
While a file is uploading, the system SHALL display an indeterminate spinner next to its filename. Pending files SHALL show a neutral waiting indicator.

#### Scenario: Uploading file shows a spinner
- **WHEN** a file is actively uploading
- **THEN** an indeterminate spinner is visible next to its filename in the file list

#### Scenario: Pending file shows a waiting indicator
- **WHEN** a file is queued but not yet uploading
- **THEN** a neutral pending indicator is shown

---

### Requirement: Per-file success state
When a file's upload sequence completes successfully, its row SHALL display a success indicator. The image SHALL be added to the gallery immediately via query invalidation. The folder suggestion returned by `complete` SHALL be ignored.

#### Scenario: Successful file shows a success indicator
- **WHEN** all three upload steps succeed for a file
- **THEN** the file row shows a success indicator
- **AND** the image appears in the gallery

---

### Requirement: Per-file failure and auto-retry
If any upload step fails, the system SHALL automatically retry the file once. If the retry also fails, the file SHALL be marked as permanently failed with a manual "Retry" button.

#### Scenario: First failure triggers auto-retry
- **WHEN** any upload step fails for a file
- **THEN** the system automatically retries the full three-step sequence once

#### Scenario: Second failure shows a manual retry button
- **WHEN** the auto-retry also fails
- **THEN** the file row shows a "Retry" button
- **AND** the file is marked as permanently failed

#### Scenario: Manual retry re-queues the file
- **WHEN** the user clicks the "Retry" button on a permanently failed file
- **THEN** the file re-enters the upload queue as pending
- **AND** follows the three-step upload sequence again

---

### Requirement: Batch upload modal stays open after completion
The batch modal SHALL NOT auto-close when uploads finish. The user SHALL close it manually.

#### Scenario: Modal remains open when all uploads complete
- **WHEN** all files in the batch reach a terminal state (success or permanent failure)
- **THEN** the modal remains open
- **AND** the user can close it manually

---

### Requirement: No per-file metadata in batch mode
The batch modal SHALL NOT present title, notes, or source URL input fields. The filename without extension SHALL be used as the title for each file automatically.

#### Scenario: No metadata fields are shown in the batch modal
- **WHEN** the batch upload modal is open
- **THEN** no title, notes, or source URL input fields are visible

---

### Requirement: No folder suggestion in batch mode
The system SHALL NOT display the folder suggestion UI for any file in a batch upload. The `suggested_folder_name` field from the `complete` response SHALL be ignored.

#### Scenario: Folder suggestion is not shown for batch uploads
- **WHEN** a file in a batch upload completes and the response includes a `suggested_folder_name`
- **THEN** no folder suggestion UI is shown
- **AND** the file is marked as successfully uploaded
