## MODIFIED Requirements

### Requirement: Settings entry point

The popup header SHALL display a Settings icon (Lucide `Settings`) to the right of the existing "Open" button, visible whenever the popup is in the logged-in, main view. Clicking the Settings icon SHALL switch the popup to the Settings view.

#### Scenario: Settings icon switches to Settings

- **WHEN** the user clicks the Settings icon in the popup header while in the main view
- **THEN** the popup renders the Settings view instead of the main view
