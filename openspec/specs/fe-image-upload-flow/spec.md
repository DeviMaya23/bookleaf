### Requirement: Upload button in main content area
The system SHALL render a **+ Image** button in the top-right of the main content area. Clicking it SHALL open the upload modal.

#### Scenario: Button is visible in the content header
- **WHEN** an authenticated user views the main content area
- **THEN** a "+ Image" button is visible in the top-right corner

#### Scenario: Clicking the button opens the upload modal
- **WHEN** the user clicks the "+ Image" button
- **THEN** the upload modal opens

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
The system SHALL accept only JPEG, PNG, GIF, and WEBP files. Validation SHALL run at selection time (both drop and file picker). Invalid files SHALL be rejected with an inline error message; the drop zone SHALL not update to show the rejected file.

#### Scenario: Invalid file type is rejected on drop
- **WHEN** the user drops a file with an unsupported type (e.g. PDF, SVG)
- **THEN** an inline error is shown in the drop zone
- **AND** the file is not staged for upload

#### Scenario: Invalid file type is rejected from file picker
- **WHEN** the user selects a file with an unsupported type via the file picker
- **THEN** an inline error is shown in the drop zone
- **AND** the file is not staged for upload

#### Scenario: Valid file types are accepted
- **WHEN** the user selects or drops a JPEG, PNG, GIF, or WEBP file
- **THEN** no validation error is shown
- **AND** the file is staged for upload

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
On submit, the system SHALL execute the 3-step upload sequence: (1) `POST /images` to initiate, (2) `PUT` to the presigned R2 URL with the file bytes, (3) `POST /images/:id/complete` to finalise. The `folder_id` SHALL be set to the current folder from the URL (`/folders/:folderId`), or omitted when on the root route (`/`). The submit button SHALL be disabled and show a loading state while the sequence is in progress.

#### Scenario: Submit triggers the 3-step upload
- **WHEN** the user submits the upload form with a valid file
- **THEN** `POST /images` is called with the file's MIME type, title, and optional folder_id
- **AND** the file bytes are PUT to the returned presigned URL
- **AND** `POST /images/:id/complete` is called after the PUT succeeds

#### Scenario: folder_id is sent when a folder is open
- **WHEN** the user is on `/folders/:folderId` and submits an upload
- **THEN** the `POST /images` request body includes `folder_id` set to the current folder's ID

#### Scenario: folder_id is omitted on the root route
- **WHEN** the user is on `/` and submits an upload
- **THEN** the `POST /images` request body does not include `folder_id`

#### Scenario: Submit button shows loading state during upload
- **WHEN** the upload sequence is in progress
- **THEN** the submit button is disabled and displays a loading indicator

---

### Requirement: Upload success flow
On successful upload (with no folder suggestion), the system SHALL close the modal, show a success toast, and refresh the image list.

#### Scenario: Success closes modal and shows toast
- **WHEN** all 3 upload steps succeed and `complete` returns no `suggested_folder_name`
- **THEN** the modal closes
- **AND** a success toast is shown
- **AND** the image list is refreshed

---

### Requirement: Upload failure flow
If any step in the upload sequence fails, the system SHALL show an error toast and keep the modal open.

#### Scenario: Upload failure shows error toast and keeps modal open
- **WHEN** any step of the upload sequence returns an error
- **THEN** an error toast is shown
- **AND** the modal remains open

---

### Requirement: Folder suggestion view
If `POST /images/:id/complete` returns a non-null `suggested_folder_name`, the modal body SHALL replace the upload form with a suggestion view. The suggestion view SHALL display the suggested folder name and offer two actions: **Accept** and **Ignore**.

#### Scenario: Suggestion view shown when complete returns a suggestion
- **WHEN** `POST /images/:id/complete` returns a non-null `suggested_folder_name`
- **THEN** the upload form is replaced by the suggestion view inside the modal
- **AND** the suggested folder name is displayed
- **AND** Accept and Ignore buttons are present

#### Scenario: Accepting the suggestion calls accept-suggestion API
- **WHEN** the user clicks Accept in the suggestion view
- **THEN** `POST /images/:id/accept-suggestion` is called with the suggested folder name
- **AND** the modal closes
- **AND** a success toast is shown

#### Scenario: Ignoring the suggestion closes the modal
- **WHEN** the user clicks Ignore in the suggestion view
- **THEN** the modal closes without calling the accept-suggestion API
- **AND** a success toast is shown

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
