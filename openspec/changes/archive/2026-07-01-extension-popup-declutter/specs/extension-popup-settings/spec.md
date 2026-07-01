## ADDED Requirements

### Requirement: Log out row in Settings

The Settings view SHALL display a "Log out" row at the bottom, below all toggle and hotkey rows. The row SHALL render the text "Log out" in red, right-aligned. Clicking it SHALL call `clearAuth()` and transition the popup to the logged-out state. The Settings view SHALL accept an `onLogout` prop wired to the same `handleLogout` handler used previously by the main view's footer.

#### Scenario: Log out row is visible at the bottom of Settings

- **WHEN** the user opens the Settings view
- **THEN** a "Log out" row is displayed below the hotkey rows, with red right-aligned text

#### Scenario: Clicking Log out clears auth and transitions to logged-out state

- **WHEN** the user clicks "Log out" in the Settings view
- **THEN** `clearAuth()` is called and the popup transitions to the logged-out state
