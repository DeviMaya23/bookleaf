## MODIFIED Requirements

### Requirement: Upload API sequence

On submit, the system SHALL execute a 4-step upload sequence:

1. `POST /images` — initiates the upload; returns `id`, `upload_url`, and `thumbnail_upload_url`
2. In parallel:
   a. `PUT` the file bytes (or converted JPEG bytes for HEIC) to `upload_url`
   b. Generate a JPEG thumbnail from the file and `PUT` it to `thumbnail_upload_url`
3. `POST /images/:id/complete` — finalises the upload, called only after both PUTs succeed, with a JSON body containing `width`, `height`, and `file_size`

The thumbnail-generation step SHALL decode the image (via `createImageBitmap`) to determine its pixel dimensions; the system SHALL reuse this decode to capture `width` and `height` rather than performing a second decode. `file_size` SHALL be the byte length of the uploaded blob (the converted JPEG blob for HEIC, the original blob otherwise). All three values SHALL be included in the `POST /images/:id/complete` request body.

The `folder_id` SHALL be set to the current folder from the URL (`/folders/:folderId`), or omitted when on the root route (`/`). The `description` SHALL be set to the value from the "Add details" notes field if non-empty, or omitted. The `source_url` SHALL be set to the value from the "Add details" source URL field if non-empty, or omitted. The submit button SHALL be disabled and show a loading state while the sequence is in progress.

#### Scenario: Submit triggers the 4-step upload

- **WHEN** the user submits the upload form with a valid file
- **THEN** `POST /images` is called with the file's MIME type, title, and optional fields
- **AND** the file bytes are PUT to `upload_url` in parallel with the thumbnail PUT to `thumbnail_upload_url`
- **AND** `POST /images/:id/complete` is called only after both PUTs succeed, with `width`, `height`, and `file_size` in the request body

#### Scenario: Dimensions and file size are derived from the decoded bitmap and uploaded blob

- **WHEN** the user submits a valid image file
- **THEN** the `width` and `height` sent to `POST /images/:id/complete` match the `ImageBitmap` dimensions produced while generating the thumbnail
- **AND** the `file_size` sent matches the byte length of the blob that was PUT to `upload_url`

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
