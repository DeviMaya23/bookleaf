# extension-popup-settings

## Purpose

Defines the extension popup's Settings view: how it is entered and exited, and the dark mode, drag-to-save, and snip hotkey controls it exposes.

## Requirements

### Requirement: Settings entry point

The popup header SHALL display a gear icon to the right of the existing "Open" button, visible whenever the popup is in the logged-in, main view. Clicking the gear icon SHALL switch the popup to the Settings view.

#### Scenario: Gear icon switches to Settings

- **WHEN** the user clicks the gear icon in the popup header while in the main view
- **THEN** the popup renders the Settings view instead of the main view

### Requirement: Settings view navigation

The Settings view SHALL display a back arrow. Clicking it SHALL switch the popup back to the main view, leaving all main-view state (recent saves, user info, auth state) unchanged. The Settings view SHALL be local UI state within the popup component tree — opening or closing it SHALL NOT reload the popup or refetch recent saves/auth/user data.

#### Scenario: Back arrow returns to the main view

- **WHEN** the user clicks the back arrow while in the Settings view
- **THEN** the popup renders the main view exactly as it was before entering Settings, with no data refetch

#### Scenario: Popup always opens to the main view

- **WHEN** the popup is reopened after having been closed while in the Settings view
- **THEN** the popup opens to the main view, not the Settings view

### Requirement: Dark mode toggle in Settings

The Settings view SHALL display a dark mode toggle that reads and writes the same `getDarkMode`/`setDarkMode` storage as the existing toggle in the main view's user row. Toggling dark mode from either location SHALL update both locations' visual state without requiring the popup to be reopened.

#### Scenario: Toggling dark mode in Settings updates the main view

- **WHEN** the user toggles dark mode from the Settings view, then returns to the main view
- **THEN** the main view reflects the new dark/light theme and its own toggle shows the updated state

#### Scenario: Toggling dark mode in the main view updates Settings

- **WHEN** the user toggles dark mode from the main view's user row, then opens Settings
- **THEN** the Settings view's dark mode toggle reflects the updated state

### Requirement: Drag-to-save toggle

The Settings view SHALL display an on/off toggle for drag-to-save, backed by a new `getDragEnabled`/`setDragEnabled` storage pair. The default value, when no value has ever been stored, SHALL be `true`.

#### Scenario: Default is enabled for existing and new users

- **WHEN** no `dragEnabled` value has ever been stored
- **THEN** the Settings toggle shows drag-to-save as on
- **AND** the drop zone behaves exactly as it did before this change

#### Scenario: Disabling drag-to-save persists immediately

- **WHEN** the user turns the drag-to-save toggle off in Settings
- **THEN** `setDragEnabled(false)` is called
- **AND** subsequent `dragstart` gestures on any page do not render the drop zone (per the corresponding change to `extension-drag-drop-save`)

#### Scenario: Re-enabling drag-to-save persists immediately

- **WHEN** the user turns the drag-to-save toggle back on in Settings
- **THEN** `setDragEnabled(true)` is called
- **AND** subsequent `dragstart` gestures resume showing the drop zone under the same conditions as before this change

### Requirement: Snip hotkey display

The Settings view SHALL display the snip command's currently configured keyboard shortcut, read via the `commands` API, as a clickable control that navigates to the browser's shortcut settings.

#### Scenario: Current shortcut is shown

- **WHEN** the Settings view is opened
- **THEN** the currently configured shortcut for the snip command is displayed

### Requirement: Snip hotkey remap on Firefox

On Firefox, clicking the hotkey control SHALL open the browser's native extension-shortcut settings via `browser.commands.openShortcutSettings()`, rather than offering in-popup key-combination capture.

#### Scenario: Clicking the hotkey control on Firefox opens the browser's shortcut settings

- **WHEN** the user clicks the hotkey control on Firefox
- **THEN** `browser.commands.openShortcutSettings()` is called
- **AND** no in-popup key-capture UI is shown
- **AND** `browser.commands.update` is never called

### Requirement: Snip hotkey remap on Chrome (best-effort)

On Chrome, where the `commands` API provides no way for an extension to set its own shortcut, clicking the hotkey control SHALL open `chrome://extensions/shortcuts` in a new tab.

#### Scenario: Clicking the hotkey control on Chrome opens the browser's shortcut settings

- **WHEN** the user clicks the hotkey control on Chrome
- **THEN** a new tab opens to `chrome://extensions/shortcuts`
- **AND** no in-popup key-capture UI is shown
- **AND** `browser.commands.update` is never called
