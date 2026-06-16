### Requirement: Upload button in main content area
The system SHALL render an upload dropdown button in the top-right of the main content area. The dropdown SHALL offer two options: "Upload image" (opens the single-file upload modal) and "Upload multiple images" (opens the batch upload modal).

#### Scenario: Dropdown is visible in the content header
- **WHEN** an authenticated user views the main content area
- **THEN** an upload dropdown button is visible in the top-right corner

#### Scenario: Selecting "Upload image" opens the single-file modal
- **WHEN** the user opens the dropdown and selects "Upload image"
- **THEN** the single-file upload modal opens

#### Scenario: Selecting "Upload multiple images" opens the batch modal
- **WHEN** the user opens the dropdown and selects "Upload multiple images"
- **THEN** the batch upload modal opens

---

### Requirement: Upload modal structure
The upload modal SHALL contain a drop zone (which also acts as a file picker trigger), an optional title field, and a submit button. The modal SHALL be dismissible via the close button or pressing Escape.

#### Scenario: Modal displays all required elements
- **WHEN** the upload modal is open
- **THEN** a drop zone area, a title input field, and a submit button are visible

#### Scenario: Modal can be dismissed
- **WHEN** the user presses Escape or clicks the close button
- **THEN** the modal closes and no upload is triggered

---

### Requirement: File selection via drop zone and file picker
The drop zone SHALL accept a file by drag-and-drop or by clicking to open the native file picker. Only one file MAY be selected at a time. Upon selection, the filename SHALL be displayed inside the drop zone.

#### Scenario: Dragging a valid file onto the drop zone selects it
- **WHEN** the user drags a valid image file onto the drop zone and releases
- **THEN** the filename is displayed inside the drop zone
- **AND** the file is staged for upload

#### Scenario: Clicking the drop zone opens the file picker
- **WHEN** the user clicks the drop zone
- **THEN** the native file picker dialog opens

#### Scenario: Selecting a file via picker displays the filename
- **WHEN** the user selects a valid file through the file picker
- **THEN** the filename is displayed inside the drop zone

---

### Requirement: File type validation

The system SHALL accept JPEG, PNG, GIF, WebP, AVIF, and HEIC (Safari only) files. Validation SHALL run at selection time (both drop and file picker). Invalid files SHALL be rejected with an inline error message; the drop zone SHALL not update to show the rejected file.

HEIC files SHALL be accepted only when the browser is Safari; on other browsers, HEIC SHALL be treated as unsupported.

#### Scenario: Invalid file type is rejected on drop

- **WHEN** the user drops a file with an unsupported type (e.g. PDF, SVG)
- **THEN** an inline error is shown in the drop zone
- **AND** the file is not staged for upload

#### Scenario: Invalid file type is rejected from file picker

- **WHEN** the user selects a file with an unsupported type via the file picker
- **THEN** an inline error is shown in the drop zone
- **AND** the file is not staged for upload

#### Scenario: JPEG, PNG, GIF, WebP, and AVIF are accepted on all browsers

- **WHEN** the user selects or drops a JPEG, PNG, GIF, WebP, or AVIF file
- **THEN** no validation error is shown
- **AND** the file is staged for upload

#### Scenario: HEIC is accepted on Safari

- **WHEN** the user selects or drops a HEIC file on Safari
- **THEN** no validation error is shown
- **AND** the file is staged for upload

#### Scenario: HEIC is rejected on non-Safari browsers

- **WHEN** the user selects or drops a HEIC file on Chrome or Firefox
- **THEN** an inline error is shown in the drop zone
- **AND** the file is not staged for upload

---

### Requirement: Optional title field with filename placeholder
The title input field SHALL be optional. Its placeholder SHALL display the selected filename (without extension). The field SHALL NOT be auto-filled. On submit, if the field is blank, the filename (without extension) SHALL be used as the title.

#### Scenario: Placeholder updates when a file is selected
- **WHEN** a file is selected
- **THEN** the title field placeholder shows the filename without its extension
- **AND** the title field value remains empty

#### Scenario: Blank title uses filename on submit
- **WHEN** the user submits with the title field left blank
- **THEN** the image is uploaded with the filename (without extension) as its title

#### Scenario: Filled title is used on submit
- **WHEN** the user types a title and submits
- **THEN** the image is uploaded with the typed title

---

### Requirement: Upload API sequence

On submit, the system SHALL execute a 4-step upload sequence:

1. `POST /images` — initiates the upload; returns `id`, `upload_url`, and `thumbnail_upload_url`
2. In parallel:
   a. `PUT` the file bytes (or converted JPEG bytes for HEIC) to `upload_url`
   b. Generate a JPEG thumbnail from the file and `PUT` it to `thumbnail_upload_url`; capture `width` and `height` from the `createImageBitmap` decode; compute a 64-character binary pHash string from the 32×32 greyscale thumbnail pixels using a DCT-based perceptual hash algorithm
3. `POST /images/:id/complete` — finalises the upload, called only after both PUTs succeed, with a JSON body containing `width`, `height`, `file_size`, and `phash`

The thumbnail-generation step SHALL decode the image (via `createImageBitmap`) to determine its pixel dimensions; the system SHALL reuse this decode to capture `width` and `height` rather than performing a second decode. `file_size` SHALL be the byte length of the uploaded blob (the converted JPEG blob for HEIC, the original blob otherwise). `phash` SHALL be computed on the thumbnail canvas after it is drawn to a 32×32 greyscale canvas, producing a 64-character binary string. All four values SHALL be included in the `POST /images/:id/complete` request body.

The `folder_id` SHALL be set to the current folder from the URL (`/folders/:folderId`), or omitted when on the root route (`/`). The `description` SHALL be set to the value from the "Add details" notes field if non-empty, or omitted. The `source_url` SHALL be set to the value from the "Add details" source URL field if non-empty, or omitted. The submit button SHALL be disabled and show a loading state while the sequence is in progress.

#### Scenario: Submit triggers the 4-step upload

- **WHEN** the user submits the upload form with a valid file
- **THEN** `POST /images` is called with the file's MIME type, title, and optional fields
- **AND** the file bytes are PUT to `upload_url` in parallel with the thumbnail PUT to `thumbnail_upload_url`
- **AND** `POST /images/:id/complete` is called only after both PUTs succeed, with `width`, `height`, `file_size`, and `phash` in the request body

#### Scenario: Dimensions and file size are derived from the decoded bitmap and uploaded blob

- **WHEN** the user submits a valid image file
- **THEN** the `width` and `height` sent to `POST /images/:id/complete` match the `ImageBitmap` dimensions produced while generating the thumbnail
- **AND** the `file_size` sent matches the byte length of the blob that was PUT to `upload_url`

#### Scenario: phash is computed from the thumbnail canvas

- **WHEN** the thumbnail is generated
- **THEN** a 64-character binary pHash string is computed from the 32×32 greyscale representation of the thumbnail
- **AND** the value is included in the `POST /images/:id/complete` request body as `phash`

#### Scenario: folder_id is sent when a folder is open

- **WHEN** the user is on `/folders/:folderId` and submits an upload
- **THEN** the `POST /images` request body includes `folder_id` set to the current folder's ID

#### Scenario: folder_id is omitted on the root route

- **WHEN** the user is on `/` and submits an upload
- **THEN** the `POST /images` request body does not include `folder_id`

#### Scenario: description is sent when notes field is filled

- **WHEN** the user fills in the notes field in "Add details" and submits
- **THEN** the `POST /images` request body includes `description` set to the entered value

#### Scenario: description is omitted when notes field is empty

- **WHEN** the user leaves the notes field empty and submits
- **THEN** the `POST /images` request body does not include `description`

#### Scenario: source_url is sent when source URL field is filled

- **WHEN** the user fills in the source URL field in "Add details" and submits
- **THEN** the `POST /images` request body includes `source_url` set to the entered value

#### Scenario: source_url is omitted when source URL field is empty

- **WHEN** the user leaves the source URL field empty and submits
- **THEN** the `POST /images` request body does not include `source_url`

#### Scenario: Submit button shows loading state during upload

- **WHEN** the upload sequence is in progress
- **THEN** the submit button is disabled and displays a loading indicator

---

### Requirement: Upload success flow (modal)

On successful upload via the upload modal, the system SHALL close the modal, show a success toast, and refresh the image list. The modal SHALL close immediately on success regardless of any async job state.

#### Scenario: Success closes modal and shows toast

- **WHEN** all 4 upload steps succeed via the upload modal
- **THEN** the modal closes
- **AND** a success toast is shown
- **AND** the image list is refreshed

---

### Requirement: Duplicate warning toast after upload

After a successful single-file upload, if the `POST /images/:id/complete` response includes a non-empty `duplicates` array, the system SHALL show a warning toast in addition to the standard success toast. The warning toast SHALL name the title of the first duplicate match. The upload is considered successful regardless of whether duplicates are detected.

#### Scenario: Warning toast shown when duplicate detected

- **WHEN** all 4 upload steps succeed and the `completeUpload` response contains one or more duplicates
- **THEN** a warning toast is displayed naming the first duplicate's title
- **AND** the modal closes
- **AND** the image list is refreshed

#### Scenario: No warning toast when no duplicates

- **WHEN** all 4 upload steps succeed and the `completeUpload` response has an empty `duplicates` array
- **THEN** only the standard success toast is shown
- **AND** no warning toast appears

---

### Requirement: Upload success flow (drag-and-drop)

The drag-and-drop file upload path (dropping a single file onto the main content area) uses the same 4-step upload sequence (`POST /images`, parallel PUTs to R2 and thumbnail URL, `POST /images/:id/complete`). On success it SHALL refresh the image list and trigger the same post-upload feedback hook as the modal path. There is no modal to close.

#### Scenario: Single-file drop succeeds

- **WHEN** a single supported image file is dropped onto the main content area
- **AND** all 4 upload steps succeed
- **THEN** the image list is refreshed
- **AND** the post-upload feedback hook is triggered with the new image ID

---

### Requirement: Upload failure flow
If any step in the upload sequence fails, the system SHALL show an error toast and keep the modal open.

#### Scenario: Upload failure shows error toast and keeps modal open
- **WHEN** any step of the upload sequence returns an error
- **THEN** an error toast is shown
- **AND** the modal remains open

---

### Requirement: Toast notifications
The system SHALL use `sonner` for toast notifications. A `<Toaster />` SHALL be mounted at the app root. Toasts SHALL be shown for upload success, upload failure, and after accepting or ignoring a folder suggestion.

#### Scenario: Toaster is mounted at app root
- **WHEN** the app renders
- **THEN** the `<Toaster />` component from `sonner` is present in the DOM

#### Scenario: Success toast is shown after upload
- **WHEN** an upload completes successfully
- **THEN** a success toast message is briefly displayed

#### Scenario: Error toast is shown after upload failure
- **WHEN** an upload fails
- **THEN** an error toast message is briefly displayed

---

### Requirement: Multi-file drag onto the app surface opens the batch modal
When more than one file is dragged and dropped onto the main app surface, the system SHALL open the batch upload modal with those files pre-loaded instead of auto-uploading.

#### Scenario: Dropping multiple files opens the batch modal
- **WHEN** the user drops more than one file onto the main app surface
- **THEN** the batch upload modal opens
- **AND** the dropped files are pre-loaded into the file list

#### Scenario: Dropping a single file continues the existing auto-upload behaviour
- **WHEN** the user drops exactly one file onto the main app surface
- **THEN** the existing single-file auto-upload sequence runs
- **AND** the batch modal does not open
