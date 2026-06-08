## MODIFIED Requirements

### Requirement: Section label separating system entries from user folders
The sidebar SHALL render a visual section label (e.g., "FOLDERS") between the system entries and the user folder tree. A horizontal divider SHALL also appear between the Trash entry and the section label. An icon button SHALL be displayed adjacent to the section label that opens the new-folder dialog.

#### Scenario: Section label is present
- **WHEN** the sidebar is rendered
- **THEN** a "FOLDERS" label is visible between the Trash entry and the user folder list

#### Scenario: New-folder icon button is present beside the section label
- **WHEN** the sidebar is rendered
- **THEN** a small icon button is visible adjacent to the "FOLDERS" label

#### Scenario: New-folder icon button opens the new-folder dialog
- **WHEN** the user clicks the icon button beside the "FOLDERS" label
- **THEN** the new-folder dialog opens, identical to the dialog opened by the footer "+ New folder" button
