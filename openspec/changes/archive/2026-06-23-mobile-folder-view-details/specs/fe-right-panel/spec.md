## MODIFIED Requirements

### Requirement: Right panel opens or updates when a folder is selected

The system SHALL render the right panel showing folder content (via `FolderPanelContent`) when the user selects a folder in the sidebar that differs from the currently active folder, on a fine-pointer device — opening the sidebar shell. On a coarse-pointer device (`useIsCoarsePointer()` is true), selecting a different folder SHALL NOT open the panel; the panel is opened for that folder only via the "View details" item in the folder's context menu, per `folder-management`. If the panel happens to already be open (e.g. left open from a previous "View details" action) when the user selects a different folder, it SHALL update to show the newly selected folder's content rather than closing, on either pointer type. Selecting the currently active folder again SHALL be a no-op — the panel's existing content, whatever it is currently displaying, SHALL remain unchanged.

#### Scenario: Selecting a different folder opens or updates the panel with folder content on a fine-pointer device

- **WHEN** a user on a fine-pointer device selects a sidebar folder that is not the currently active folder
- **THEN** the right panel becomes visible (or updates, if already visible) in the sidebar shell
- **AND** the panel displays that folder's metadata via `FolderPanelContent`

#### Scenario: Selecting a different folder does not open the panel on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device selects a sidebar folder that is not the currently active folder
- **AND** the right panel is not currently open
- **THEN** the right panel remains closed
- **AND** the panel can be opened for that folder via "View details" in the folder's context menu, per `folder-management`

#### Scenario: An already-open panel updates to the newly selected folder on a coarse-pointer device

- **WHEN** the right panel is open on a coarse-pointer device, showing a previously selected folder's content
- **AND** the user selects a different folder in the sidebar
- **THEN** the right panel updates to show the newly selected folder's metadata
- **AND** the panel does not close

#### Scenario: Re-selecting the active folder leaves the panel untouched

- **WHEN** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel's current content remains unchanged
- **AND** no new panel state is set

#### Scenario: Re-selecting the active folder while image content is shown leaves it untouched

- **WHEN** the right panel is currently showing image content
- **AND** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel continues showing the same image content
- **AND** the panel does not switch to folder content
