## ADDED Requirements

### Requirement: Focus mode toggle is available in the gallery and image viewer toolbars

The system SHALL render a "Focus" toggle button as the leftmost element of
the top toolbar in the gallery view (to the left of the search input) and
as the leftmost element of the image viewer header (to the left of the Back
button). The toggle SHALL reflect its current on/off state visually (e.g.
`aria-pressed`).

#### Scenario: Focus toggle is visible in the gallery toolbar

- **WHEN** the gallery view is rendered
- **THEN** a Focus toggle button is visible to the left of the search input

#### Scenario: Focus toggle is visible in the image viewer header

- **WHEN** the image viewer is open
- **THEN** a Focus toggle button is visible to the left of the Back button

#### Scenario: Toggle reflects active state

- **WHEN** focus mode is active
- **THEN** the Focus toggle button is shown in a visually pressed/active state in whichever toolbar is currently rendered

---

### Requirement: Activating focus mode hides the left sidebar and expands the main content area

The system SHALL stop rendering `FolderSidebar` while focus mode is active,
and `<main>` SHALL expand to fill the full viewport width (no longer offset
by the sidebar's width).

#### Scenario: Enabling focus mode hides the sidebar

- **WHEN** the user activates the Focus toggle
- **THEN** `FolderSidebar` is no longer rendered
- **AND** the main content area expands to full viewport width

#### Scenario: Disabling focus mode restores the sidebar

- **WHEN** focus mode is active and the user deactivates the Focus toggle
- **THEN** `FolderSidebar` is rendered again at its fixed 240px width
- **AND** the main content area returns to its original width

---

### Requirement: Activating focus mode hides the right panel

The system SHALL stop rendering `RightPanel` while focus mode is active,
regardless of whether an image is selected or a folder panel would
otherwise be shown.

#### Scenario: Enabling focus mode hides an open right panel

- **WHEN** the right panel is open showing an image's or folder's details
- **AND** the user activates the Focus toggle
- **THEN** the right panel is no longer rendered

#### Scenario: Selecting an image while focus mode is active does not open the right panel

- **WHEN** focus mode is active
- **AND** the user clicks an image card
- **THEN** the right panel is not rendered

---

### Requirement: Underlying selection state continues to update while focus mode is active

The system SHALL continue to update `selectedImage` and `folderPanelOpen`
state in response to user interactions while focus mode is active, even
though `RightPanel` is not rendered. When focus mode is deactivated, the
right panel SHALL immediately reflect the most recently accumulated
selection state.

#### Scenario: Selecting a different image while focus mode is active updates state silently

- **WHEN** focus mode is active and the right panel is hidden
- **AND** the user clicks a different image card than the one previously selected
- **THEN** `selectedImage` updates to the newly clicked image
- **AND** the right panel remains hidden

#### Scenario: Disabling focus mode reveals the latest selection

- **WHEN** focus mode is active and the user has clicked image B (after previously having image A selected)
- **AND** the user deactivates the Focus toggle
- **THEN** the right panel becomes visible showing image B's details

---

### Requirement: Double-click opens the image viewer regardless of focus mode

The system SHALL continue to open `ImageViewer` on double-click of an image
card whether or not focus mode is active. While focus mode is active, the
viewer SHALL render at full width (no sidebar, no right panel).

#### Scenario: Double-clicking an image opens the viewer in focus mode

- **WHEN** focus mode is active
- **AND** the user double-clicks an image card
- **THEN** the image viewer opens
- **AND** it renders at full viewport width with no sidebar or right panel visible

---

### Requirement: Focus mode state is session-only

The system SHALL NOT persist focus mode state across page reloads or
navigation. Focus mode SHALL default to inactive on initial mount.

#### Scenario: Focus mode resets on reload

- **WHEN** focus mode is active
- **AND** the user reloads the page
- **THEN** focus mode is inactive after reload
