## ADDED Requirements

### Requirement: Upload modal accepts an externally staged file
The `UploadModal` SHALL accept an optional `initialFile` prop of type `File`. When provided and the modal opens, the file SHALL be staged immediately as if the user had selected it via the drop zone or file picker. The existing `handleFile` validation path SHALL be reused, meaning unsupported file types are still rejected with the inline error.

#### Scenario: Modal opens with file pre-staged when initialFile is provided
- **WHEN** the modal is opened with a valid `initialFile`
- **THEN** the file is staged immediately
- **AND** the file preview is shown in the drop zone area
- **AND** no user interaction with the drop zone is required

#### Scenario: Invalid initialFile type is rejected with inline error
- **WHEN** the modal is opened with an `initialFile` of an unsupported type
- **THEN** the inline type error is shown
- **AND** no file is staged

---

### Requirement: Title field is auto-focused when modal is opened via paste
When the modal is opened with an `initialFile` (i.e. via the clipboard paste path), the title `<input>` SHALL be focused automatically after the modal opens. The title field value SHALL remain blank; the placeholder SHALL show the filename without its extension as normal.

#### Scenario: Title input is focused when opened with initialFile
- **WHEN** the modal opens with a valid `initialFile`
- **THEN** the title input is focused
- **AND** the title field value is empty
- **AND** the placeholder shows the filename without extension

#### Scenario: Title input is not auto-focused on normal open
- **WHEN** the modal is opened via the upload button (no initialFile)
- **THEN** the title input is not focused automatically
