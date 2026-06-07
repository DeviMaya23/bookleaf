## MODIFIED Requirements

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
