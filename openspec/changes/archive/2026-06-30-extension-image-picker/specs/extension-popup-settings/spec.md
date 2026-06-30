## ADDED Requirements

### Requirement: Browse images hotkey display

The Settings view SHALL display the `browse-images` command's currently configured keyboard shortcut in a dedicated row below the snip hotkey row. The row SHALL follow the same visual pattern as the snip hotkey row: a label on the left (`"Browse images hotkey"`) and a clickable button on the right showing the current shortcut value (or `"Not set"` if unassigned). Clicking the button SHALL invoke the same browser shortcut settings navigation as the snip hotkey control (Firefox: `browser.commands.openShortcutSettings()`; Chrome: open `chrome://extensions/shortcuts`). Both shortcuts SHALL be fetched in the same `browser.commands.getAll()` call that already retrieves the snip command.

#### Scenario: Browse images shortcut is shown in Settings

- **WHEN** the Settings view is opened
- **THEN** a row labeled `"Browse images hotkey"` displays the currently configured shortcut for the `browse-images` command (e.g. `"Alt+Shift+I"`)

#### Scenario: Clicking the browse images hotkey control opens shortcut settings

- **WHEN** the user clicks the browse images hotkey button
- **THEN** the browser's shortcut settings are opened (same behavior as the snip hotkey control)
