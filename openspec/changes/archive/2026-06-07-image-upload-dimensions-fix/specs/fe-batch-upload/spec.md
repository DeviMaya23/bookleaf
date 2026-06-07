## MODIFIED Requirements

### Requirement: Per-file upload sequence

For each file in the batch, the system SHALL execute the existing three-step upload sequence: (1) `POST /images` to initiate, (2) `PUT` to the presigned R2 URL, (3) `POST /images/:id/complete` with a JSON body containing `width`, `height`, and `file_size`. The `folder_id` SHALL be set to the current folder, or omitted when on a non-folder route. The title SHALL be the filename without extension.

The thumbnail-generation step SHALL decode the image (via `createImageBitmap`, the same decode used to build the thumbnail PUT to R2) to determine its pixel dimensions; the system SHALL reuse this decode to capture `width` and `height` rather than performing a second decode. `file_size` SHALL be the byte length of the uploaded blob (the converted JPEG blob for HEIC, the original blob otherwise).

#### Scenario: Each file follows the three-step upload sequence
- **WHEN** a file is uploaded as part of a batch
- **THEN** `POST /images` is called with the file's title (filename without extension), MIME type, and optional `folder_id`
- **AND** the file bytes are PUT to the returned presigned URL
- **AND** `POST /images/:id/complete` is called after the PUT succeeds, with `width`, `height`, and `file_size` in the request body

#### Scenario: Dimensions and file size are derived from the decoded bitmap and uploaded blob
- **WHEN** a file in a batch is uploaded
- **THEN** the `width` and `height` sent to `POST /images/:id/complete` match the `ImageBitmap` dimensions produced while generating that file's thumbnail
- **AND** the `file_size` sent matches the byte length of the blob that was PUT to the presigned upload URL

#### Scenario: folder_id is included when a folder is active
- **WHEN** the user is on `/folders/:folderId` and uploads a batch
- **THEN** each `POST /images` request includes `folder_id`

#### Scenario: folder_id is omitted on non-folder routes
- **WHEN** the user is not on a folder route and uploads a batch
- **THEN** `folder_id` is not included in any `POST /images` request
