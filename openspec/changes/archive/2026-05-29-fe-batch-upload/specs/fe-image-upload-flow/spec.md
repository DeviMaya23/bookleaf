## MODIFIED Requirements

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

## ADDED Requirements

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
