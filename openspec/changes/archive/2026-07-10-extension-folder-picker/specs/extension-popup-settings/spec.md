## MODIFIED Requirements

### Requirement: Settings view navigation

The Settings view SHALL display a back arrow. Clicking it SHALL switch the popup back to the main view, leaving all main-view state (recent saves, user info, auth state) unchanged. The Settings view SHALL be local UI state within the popup component tree — opening or closing it SHALL NOT reload the popup or refetch recent saves/auth/user data.

From the Settings view, clicking the "Default folder" button SHALL navigate to the FolderPicker panel view. The FolderPicker panel SHALL display a back arrow that returns to the Settings view. The navigation stack is linear: main → settings → folder-picker, with each back arrow moving one step back.

#### Scenario: Back arrow in Settings returns to the main view

- **WHEN** the user clicks the back arrow while in the Settings view
- **THEN** the popup renders the main view exactly as it was before entering Settings, with no data refetch

#### Scenario: Back arrow in FolderPicker returns to the Settings view

- **WHEN** the user clicks the back arrow while in the FolderPicker view
- **THEN** the popup renders the Settings view exactly as it was before entering FolderPicker

#### Scenario: Popup always opens to the main view

- **WHEN** the popup is reopened after having been closed while in the Settings or FolderPicker view
- **THEN** the popup opens to the main view, not the Settings or FolderPicker view
