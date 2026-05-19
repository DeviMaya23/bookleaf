## MODIFIED Requirements

### Requirement: Upload API sequence

On submit, the system SHALL execute the 3-step upload sequence: (1) `POST /images` to initiate, (2) `PUT` to the presigned R2 URL with the file bytes, (3) `POST /images/:id/complete` to finalise. The `folder_id` SHALL be set to the current folder from the URL (`/folders/:folderId`), or omitted when on the root route (`/`). The `description` SHALL be set to the value from the "Add details" notes field if non-empty, or omitted. The `source_url` SHALL be set to the value from the "Add details" source URL field if non-empty, or omitted. The submit button SHALL be disabled and show a loading state while the sequence is in progress.

#### Scenario: Submit triggers the 3-step upload

- **WHEN** the user submits the upload form with a valid file
- **THEN** `POST /images` is called with the file's MIME type, title, and optional folder_id, description, and source_url
- **AND** the file bytes are PUT to the returned presigned URL
- **AND** `POST /images/:id/complete` is called after the PUT succeeds

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
