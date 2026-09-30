## ADDED Requirements

### Requirement: Right panel is always present as a docked column on desktop

On fine-pointer devices at or above the `sm` breakpoint, the system SHALL always render the right panel `<aside>` as a fixed-width column to the right of the main content area. The panel SHALL be mounted unconditionally — its presence SHALL NOT depend on whether an image is selected, a folder is active, or select mode is engaged. The grid width SHALL remain constant regardless of panel content changes.

The panel SHALL be omitted entirely while focus mode is active (per `focus-mode`), and on coarse-pointer devices the existing bottom-drawer behavior (per `fe-right-panel`) applies instead.

#### Scenario: Panel is present on initial load with no image selected

- **WHEN** the app loads on a fine-pointer device with no image selected
- **THEN** the right panel aside is present in the DOM
- **AND** the main content area width does not change when the user subsequently clicks an image

#### Scenario: Selecting then deselecting an image does not reflow the grid

- **WHEN** a user on a fine-pointer device selects an image and then navigates to a different folder
- **THEN** the right panel remains present throughout
- **AND** the main content area width stays constant

#### Scenario: Panel is absent while focus mode is active

- **WHEN** focus mode is active on a fine-pointer device
- **THEN** the right panel aside is not rendered

---

### Requirement: Right panel has a collapse/expand chevron toggle

The system SHALL render a chevron button in the right panel's shell header. Clicking it SHALL toggle the panel between expanded (`w-80`) and collapsed (`w-8`) states. In the collapsed state, only the chevron strip SHALL be visible — no panel content is shown. Only the user's explicit use of the toggle SHALL change the panel width.

#### Scenario: Chevron button collapses the panel

- **WHEN** the right panel is expanded and the user clicks the chevron button
- **THEN** the panel collapses to a narrow strip showing only the chevron
- **AND** the main content area expands to fill the vacated width

#### Scenario: Chevron button expands the panel

- **WHEN** the right panel is collapsed and the user clicks the chevron button
- **THEN** the panel expands to its full width
- **AND** panel content becomes visible again

#### Scenario: Selecting an image does not expand a collapsed panel

- **WHEN** the right panel is collapsed
- **AND** the user clicks an image card
- **THEN** the panel remains collapsed
- **AND** the panel width does not change

#### Scenario: Switching folders does not change panel width

- **WHEN** the right panel is expanded and the user navigates to a different folder
- **THEN** the panel remains expanded

---

### Requirement: Collapsed state persists in localStorage

The system SHALL persist the right panel's collapsed/expanded state in `localStorage` under the key `bookleaf-right-panel-collapsed`. The state SHALL be read on mount and written on every toggle. It SHALL survive image selection changes, folder navigation, and full page reloads.

#### Scenario: Collapsed state survives a page reload

- **WHEN** the user collapses the right panel and reloads the page
- **THEN** the right panel is collapsed on reload without any user action

#### Scenario: Expanded state survives a page reload

- **WHEN** the user expands the right panel and reloads the page
- **THEN** the right panel is expanded on reload

#### Scenario: Collapsed state survives folder navigation

- **WHEN** the right panel is collapsed and the user navigates to a different folder
- **THEN** the right panel remains collapsed after navigation

---

### Requirement: Panel shows neutral content for All, Unsorted, and Trash views

When no image is selected and the current view is All, Unsorted, or Trash (i.e. not a named folder), the system SHALL display the current view's label ("All", "Unsorted", or "Trash") as the panel content. No image count, no actions, and no editable fields SHALL be shown.

#### Scenario: Neutral content shows the view label for All

- **WHEN** the user is on the All view with no image selected
- **THEN** the panel shows "All" as its content with no other actions or metadata

#### Scenario: Neutral content shows the view label for Unsorted

- **WHEN** the user is on the Unsorted view with no image selected
- **THEN** the panel shows "Unsorted" as its content

#### Scenario: Neutral content shows the view label for Trash

- **WHEN** the user is on the Trash view with no image selected
- **THEN** the panel shows "Trash" as its content

#### Scenario: Neutral content is replaced when an image is selected

- **WHEN** the panel is showing neutral content
- **AND** the user clicks an image card
- **THEN** the panel switches to image content for the selected image
