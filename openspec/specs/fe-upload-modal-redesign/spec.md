## ADDED Requirements

### Requirement: Empty state preserves existing drop zone layout

When no file is staged, the modal SHALL display the existing dashed-border drop zone and an inactive title input field as before. The Upload button SHALL be disabled in the empty state.

#### Scenario: Modal shows drop zone when no file is selected

- **WHEN** the upload modal is open and no file has been selected
- **THEN** the dashed drop zone is visible with "Drop an image here or click to browse" text
- **AND** the Upload button is disabled

---

### Requirement: File selected state shows inline thumbnail preview

When a file is staged (via drop or file picker), the drop zone SHALL be replaced by an inline thumbnail preview row. The row SHALL display:
- A small image rendered via `URL.createObjectURL(file)` (approx 48px tall, cover crop)
- The filename
- A remove button (×) that clears the staged file and returns to the empty state

The object URL SHALL be created when the file is staged and revoked when the modal closes or the file is cleared.

#### Scenario: Drop zone is replaced by thumbnail preview after file selection

- **WHEN** the user selects or drops a valid file in the modal
- **THEN** the drop zone is no longer visible
- **AND** a thumbnail preview of the file is shown with the filename

#### Scenario: Remove button clears staged file

- **WHEN** the user clicks the × button on the thumbnail preview row
- **THEN** the staged file is cleared
- **AND** the drop zone reappears
- **AND** the object URL is revoked

#### Scenario: Object URL is revoked on modal close

- **WHEN** the modal is closed (via × or Escape)
- **THEN** the object URL for the staged file is revoked

---

### Requirement: Title field is pre-filled when file is staged

When a file is staged, the title input SHALL be automatically filled with `fileBaseName(file.name)` (the filename without extension). The user MAY edit the pre-filled value. If the user clears the field, it SHALL remain empty (no re-fill on blur).

#### Scenario: Title is pre-filled when file is staged

- **WHEN** the user selects or drops a valid file
- **THEN** the title input is filled with the filename without its extension

#### Scenario: User can override the pre-filled title

- **WHEN** the user edits the pre-filled title
- **THEN** the edited value is used on submit

---

### Requirement: Collapsible "Add details" section for notes and source URL

Below the title input, a collapsible section labelled "Add details ▸" SHALL be rendered. It SHALL be collapsed by default. Clicking the label SHALL toggle it open/closed. When open it SHALL reveal:
- A textarea labelled "Notes"
- A text input labelled "Source URL" with placeholder `https://…`

Both fields SHALL be empty by default and optional.

#### Scenario: "Add details" section is collapsed by default

- **WHEN** the upload modal is open
- **THEN** the "Add details" section is collapsed and notes/source fields are not visible

#### Scenario: Clicking "Add details" expands the section

- **WHEN** the user clicks the "Add details" toggle
- **THEN** the notes textarea and source URL input become visible

#### Scenario: Clicking again collapses the section

- **WHEN** the "Add details" section is open and the user clicks the toggle again
- **THEN** the section collapses and the fields are hidden

---

### Requirement: Notes and source URL are passed to the upload API

When the user submits the modal, the `description` and `source_url` values from the "Add details" section SHALL be included in the `POST /images` request body if non-empty. Empty strings SHALL be omitted (not sent as empty strings).

#### Scenario: Filled notes and source URL are sent on upload

- **WHEN** the user fills in notes and/or source URL and submits
- **THEN** `POST /images` includes `description` and/or `source_url` in the request body

#### Scenario: Empty notes and source URL fields are omitted

- **WHEN** the user leaves notes and source URL empty and submits
- **THEN** `POST /images` does not include `description` or `source_url` in the request body

---

### Requirement: Modal state is fully reset on close

When the modal closes (either via explicit close or after successful upload), all fields SHALL reset: staged file cleared, object URL revoked, title empty, notes empty, source URL empty, "Add details" section collapsed.

#### Scenario: Modal state resets after successful upload

- **WHEN** the upload completes successfully and the modal closes
- **THEN** all fields are cleared and "Add details" is collapsed

#### Scenario: Modal state resets on manual close

- **WHEN** the user closes the modal without uploading
- **THEN** all fields are cleared and "Add details" is collapsed
