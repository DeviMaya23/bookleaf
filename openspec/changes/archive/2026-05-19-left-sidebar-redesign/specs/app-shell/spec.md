## MODIFIED Requirements

### Requirement: Folder list in sidebar
The system SHALL fetch the folder list from `GET /folders` and display it in the sidebar as a nested tree below a "FOLDERS" section label. The sidebar SHALL show three pinned system entries above the section label: **All**, **Unsorted**, and **Trash** (de-emphasized). A horizontal divider and section label SHALL separate the system entries from the user folder tree. The "+ New folder" affordance SHALL remain visible below the folder tree in the footer area.

#### Scenario: System entries are always visible
- **WHEN** the application shell is rendered
- **THEN** "All", "Unsorted", and "Trash" entries are visible above the folder tree

#### Scenario: Folder tree is populated from API
- **WHEN** the application shell is rendered and `GET /folders` returns a non-empty list
- **THEN** each root-level folder is displayed in the sidebar below the FOLDERS section label, with children nested underneath their parents

#### Scenario: Section label separates system entries from folder tree
- **WHEN** the application shell is rendered
- **THEN** a "FOLDERS" section label is visible between the Trash entry and the user folder tree

#### Scenario: New folder affordance is visible
- **WHEN** the application shell is rendered
- **THEN** a "+ New folder" button is visible in the sidebar footer

#### Scenario: Empty API folder list
- **WHEN** the application shell is rendered and `GET /folders` returns an empty list
- **THEN** the three system entries and section label are shown, with an empty user folder tree and "+ New folder" still present
