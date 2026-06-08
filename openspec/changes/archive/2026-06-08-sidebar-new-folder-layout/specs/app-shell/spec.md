## MODIFIED Requirements

### Requirement: Folder list in sidebar
The system SHALL fetch the folder list from `GET /folders` and display it in the sidebar as a nested tree below a "FOLDERS" section label. The sidebar SHALL show three pinned system entries above the section label: **All**, **Unsorted**, and **Trash** (de-emphasized). A horizontal divider and section label SHALL separate the system entries from the user folder tree. An icon button adjacent to the "FOLDERS" section label SHALL always be available to create a new folder. The full-width "+ New folder" affordance in the footer area SHALL be displayed only when the user's folder list is empty.

#### Scenario: System entries are always visible
- **WHEN** the application shell is rendered
- **THEN** "All", "Unsorted", and "Trash" entries are visible above the folder tree

#### Scenario: Folder tree is populated from API
- **WHEN** the application shell is rendered and `GET /folders` returns a non-empty list
- **THEN** each root-level folder is displayed in the sidebar below the FOLDERS section label, with children nested underneath their parents

#### Scenario: Section label separates system entries from folder tree
- **WHEN** the application shell is rendered
- **THEN** a "FOLDERS" section label is visible between the Trash entry and the user folder tree

#### Scenario: New-folder icon button is always visible
- **WHEN** the application shell is rendered, regardless of whether the folder list is empty or populated
- **THEN** the icon button beside the "FOLDERS" section label is visible

#### Scenario: Footer new-folder affordance is shown for an empty account
- **WHEN** the application shell is rendered and `GET /folders` returns an empty list
- **THEN** the three system entries and section label are shown, with an empty user folder tree
- **AND** the full-width "+ New folder" button is visible in the sidebar footer

#### Scenario: Footer new-folder affordance is hidden once folders exist
- **WHEN** the application shell is rendered and `GET /folders` returns a non-empty list
- **THEN** the full-width "+ New folder" button is not present in the sidebar footer
