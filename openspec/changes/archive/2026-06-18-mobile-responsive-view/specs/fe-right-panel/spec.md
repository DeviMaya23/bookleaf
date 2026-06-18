## MODIFIED Requirements

### Requirement: Right panel opens when an image card is clicked

The system SHALL render a 320px right panel (`RightPanel` component) as a sibling to the main content area in `AppLayout`. The panel SHALL be hidden when no image is selected. When the user clicks an image card, the panel SHALL become visible and display that image's metadata. The panel SHALL NOT be rendered while focus mode is active, even if an image is selected. The panel SHALL NOT be rendered below the `sm` breakpoint, even if an image is selected.

#### Scenario: Clicking an image card opens the right panel

- **WHEN** the authenticated user clicks an image card in the gallery at or above the `sm` breakpoint
- **THEN** the right panel becomes visible on the right side of the layout
- **AND** the panel displays the selected image's metadata

#### Scenario: Panel is hidden when no image is selected

- **WHEN** no image has been selected
- **THEN** the right panel is not rendered in the layout

#### Scenario: Panel stays hidden while focus mode is active

- **WHEN** focus mode is active
- **AND** the user clicks an image card
- **THEN** the right panel is not rendered, even though an image is now selected

#### Scenario: Panel stays hidden below the breakpoint

- **WHEN** the viewport width is below the `sm` breakpoint
- **AND** the user taps an image card
- **THEN** the right panel is not rendered, even though an image is now selected

### Requirement: Right panel opens or updates when a folder is selected

The system SHALL render the right panel showing folder content (via `FolderPanelContent`) when the user selects a folder in the sidebar that differs from the currently active folder. Selecting the currently active folder again SHALL be a no-op — the panel's existing content, whatever it is currently displaying, SHALL remain unchanged. The panel SHALL NOT be rendered below the `sm` breakpoint, even if a folder is selected.

#### Scenario: Selecting a different folder opens or updates the panel with folder content

- **WHEN** the authenticated user clicks a sidebar folder that is not the currently active folder, at or above the `sm` breakpoint
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

#### Scenario: Panel stays hidden below the breakpoint when a folder is selected

- **WHEN** the viewport width is below the `sm` breakpoint
- **AND** the user selects a sidebar folder
- **THEN** the right panel is not rendered, even though that folder is now selected
