## ADDED Requirements

### Requirement: Right panel opens or updates when a folder is selected

The system SHALL render the right panel showing folder content (via `FolderPanelContent`) when the user selects a folder in the sidebar that differs from the currently active folder. Selecting the currently active folder again SHALL be a no-op — the panel's existing content, whatever it is currently displaying, SHALL remain unchanged.

#### Scenario: Selecting a different folder opens or updates the panel with folder content

- **WHEN** the authenticated user clicks a sidebar folder that is not the currently active folder
- **THEN** the right panel becomes visible (or updates, if already visible)
- **AND** the panel displays that folder's metadata via `FolderPanelContent`

#### Scenario: Re-selecting the active folder leaves the panel untouched

- **WHEN** the authenticated user clicks the sidebar folder that is already active
- **THEN** the right panel's current content remains unchanged
- **AND** no new panel state is set

#### Scenario: Re-selecting the active folder while image content is shown leaves it untouched

- **WHEN** the right panel is currently showing image content
- **AND** the authenticated user clicks the sidebar folder that is already active
- **THEN** the right panel continues showing the same image content
- **AND** the panel does not switch to folder content

---

### Requirement: Selecting an image replaces folder content in the right panel

The system SHALL ensure that selecting an image card always results in the right panel showing that image's content, replacing any folder content that was previously displayed. Image and folder content SHALL never be displayed in the panel simultaneously.

#### Scenario: Selecting an image while folder content is shown switches the panel to image content

- **WHEN** the right panel is currently showing folder content
- **AND** the authenticated user clicks an image card in the gallery
- **THEN** the right panel switches to displaying that image's metadata
- **AND** the folder content is no longer shown
