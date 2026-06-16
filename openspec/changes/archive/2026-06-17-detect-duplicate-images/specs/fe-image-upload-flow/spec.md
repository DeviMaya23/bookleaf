## MODIFIED Requirements

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

## ADDED Requirements

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
