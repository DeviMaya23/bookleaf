## ADDED Requirements

### Requirement: Main content area detects OS file drags

The `<main>` element in `AppLayout` SHALL listen for native `dragover` and `dragleave`/`dragend` browser events. It SHALL activate the overlay only when `event.dataTransfer.types` includes `'Files'`. This prevents interference with dnd-kit's internal drag events, which never carry file types.

#### Scenario: Dragging a file from the OS over the main area activates the overlay

- **WHEN** the user drags a file from their OS over the main content area
- **THEN** `event.dataTransfer.types` includes `'Files'`
- **AND** the drop overlay state is set to active

#### Scenario: dnd-kit internal drags do not activate the file overlay

- **WHEN** the user drags an image card or folder item (dnd-kit drag)
- **THEN** `event.dataTransfer.types` does not include `'Files'`
- **AND** the file drop overlay is not shown

---

### Requirement: Full-page drop overlay is shown during OS file drag

While a valid OS file drag is over the main area, a full-page overlay SHALL be rendered over the main content with the text "Drop to upload" and a visual indicator. The overlay SHALL disappear when the drag leaves the main area or after the drop event fires.

#### Scenario: Overlay is visible while file is dragged over main area

- **WHEN** the user drags a file from the OS over the main content area
- **THEN** a full-screen overlay with "Drop to upload" text is visible

#### Scenario: Overlay disappears when drag leaves the area

- **WHEN** the user drags the file away from the main content area without dropping
- **THEN** the overlay disappears

---

### Requirement: File type validation on drop

On drop, the system SHALL read the first file from `event.dataTransfer.files`. It SHALL validate the MIME type against the accepted list (image/jpeg, image/png, image/gif, image/webp). Invalid types SHALL show an error toast and perform no upload.

#### Scenario: Valid image file type proceeds to upload

- **WHEN** the user drops a JPEG, PNG, GIF, or WEBP file onto the main area
- **THEN** the upload sequence begins

#### Scenario: Invalid file type shows error toast

- **WHEN** the user drops a file with an unsupported MIME type (e.g. PDF, SVG)
- **THEN** an error toast is shown with a message indicating the unsupported type
- **AND** no upload is initiated

---

### Requirement: Auto-upload to current folder on valid drop

On a valid file drop, the system SHALL run the 3-step upload sequence (initiateUpload → putToR2 → completeUpload) using the `folderId` derived from the current route (`/folders/:folderId` or null for root/unsorted). The `completeUpload` call SHALL include a JSON body containing `width`, `height`, and `file_size`. No modal SHALL be shown. `description` and `source_url` SHALL NOT be set (they can be edited in the right panel after upload). A loading state SHALL be indicated by the overlay while the upload is in progress.

The thumbnail-generation step SHALL decode the image (via `createImageBitmap`, the same decode used to build the thumbnail PUT to R2) to determine its pixel dimensions; the system SHALL reuse this decode to capture `width` and `height` rather than performing a second decode. `file_size` SHALL be the byte length of the uploaded blob (the converted JPEG blob for HEIC, the original blob otherwise).

#### Scenario: File dropped on a folder view uploads to that folder

- **WHEN** the user is on `/folders/:folderId` and drops a valid file
- **THEN** `POST /images` is called with the file's MIME type, filename as title, and the current `folder_id`
- **AND** the file bytes are PUT to the presigned URL
- **AND** `POST /images/:id/complete` is called with `width`, `height`, and `file_size` in the request body

#### Scenario: Dimensions and file size are derived from the decoded bitmap and uploaded blob

- **WHEN** a dropped file is auto-uploaded
- **THEN** the `width` and `height` sent to `POST /images/:id/complete` match the `ImageBitmap` dimensions produced while generating the thumbnail
- **AND** the `file_size` sent matches the byte length of the blob that was PUT to the presigned upload URL

#### Scenario: File dropped on root view uploads without folder assignment

- **WHEN** the user is on `/` or `/unsorted` and drops a valid file
- **THEN** `POST /images` is called without a `folder_id`

#### Scenario: Upload failure shows error toast

- **WHEN** any step of the upload sequence returns an error
- **THEN** an error toast is shown
- **AND** the overlay is dismissed

---

### Requirement: Right panel opens after successful auto-upload with title focused

On successful auto-upload (regardless of folder suggestion), the system SHALL call `GET /images/:id` to fetch the full image object, set it as the selected image to open the right panel, and auto-focus the title input. Any folder suggestion from `completeUpload` SHALL be silently ignored for the auto-upload flow (the user can edit folder assignment via drag-and-drop).

#### Scenario: Right panel opens after successful auto-upload

- **WHEN** the auto-upload sequence completes successfully
- **THEN** the image list is refreshed
- **AND** `GET /images/:id` is called for the newly uploaded image
- **AND** the right panel opens showing that image
- **AND** the title input in the right panel is focused

#### Scenario: Folder suggestion is ignored in auto-upload flow

- **WHEN** `completeUpload` returns a non-null `suggested_folder_name`
- **THEN** no folder suggestion prompt is shown
- **AND** the right panel opens as normal
