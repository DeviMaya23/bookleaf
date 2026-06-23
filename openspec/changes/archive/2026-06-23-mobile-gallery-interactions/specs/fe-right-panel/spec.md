## MODIFIED Requirements

### Requirement: Right panel opens when an image card is clicked

The system SHALL render the right panel (`RightPanel` component) when an image is selected. On fine-pointer devices, the panel SHALL render as a 320px sidebar, a sibling to the main content area in `AppLayout`, opened by clicking an image card; this behavior is unchanged from before. On coarse-pointer devices (`useIsCoarsePointer()` is true), the panel SHALL render as a bottom drawer instead of a sidebar, and SHALL be opened via the "View details" item in the image card's context menu (per `fe-gallery-view`) rather than by tapping the card — tapping a card on a coarse-pointer device opens the lightbox instead, per `fe-image-lightbox`. The panel SHALL be hidden (not rendered in either shell) when no image is selected. The panel SHALL NOT be rendered while focus mode is active, even if an image is selected.

#### Scenario: Clicking an image card opens the right panel as a sidebar on a fine-pointer device

- **WHEN** a user on a fine-pointer device clicks an image card in the gallery
- **THEN** the right panel becomes visible as a sidebar on the right side of the layout
- **AND** the panel displays the selected image's metadata

#### Scenario: Selecting "View details" opens the right panel as a bottom drawer on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device selects "View details" from an image card's context menu
- **THEN** the right panel becomes visible as a bottom drawer
- **AND** the panel displays the selected image's metadata

#### Scenario: Tapping an image card does not open the right panel on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device taps an image card
- **THEN** the right panel is not opened
- **AND** the lightbox opens instead, per `fe-image-lightbox`

#### Scenario: Panel is hidden when no image is selected

- **WHEN** no image has been selected
- **THEN** the right panel is not rendered in either shell

#### Scenario: Panel stays hidden while focus mode is active

- **WHEN** focus mode is active
- **AND** an image becomes selected
- **THEN** the right panel is not rendered, even though an image is now selected

---

### Requirement: Right panel opens or updates when a folder is selected

The system SHALL render the right panel showing folder content (via `FolderPanelContent`) when the user selects a folder in the sidebar that differs from the currently active folder, in either shell (sidebar on fine-pointer devices, bottom drawer on coarse-pointer devices per `useIsCoarsePointer()`). Selecting the currently active folder again SHALL be a no-op — the panel's existing content, whatever it is currently displaying, SHALL remain unchanged.

#### Scenario: Selecting a different folder opens or updates the panel with folder content

- **WHEN** the authenticated user selects a sidebar folder that is not the currently active folder
- **THEN** the right panel becomes visible (or updates, if already visible), in the sidebar shell on fine-pointer devices or the bottom-drawer shell on coarse-pointer devices
- **AND** the panel displays that folder's metadata via `FolderPanelContent`

#### Scenario: Re-selecting the active folder leaves the panel untouched

- **WHEN** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel's current content remains unchanged
- **AND** no new panel state is set

#### Scenario: Re-selecting the active folder while image content is shown leaves it untouched

- **WHEN** the right panel is currently showing image content
- **AND** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel continues showing the same image content
- **AND** the panel does not switch to folder content

## ADDED Requirements

### Requirement: Right panel renders as a bottom drawer on coarse-pointer devices

On coarse-pointer devices, the system SHALL render the right panel's content (`ImagePanelBody` or `FolderPanelContent`, unchanged from the fine-pointer sidebar) inside a fixed bottom-anchored drawer with a backdrop, instead of the 320px sidebar. The drawer SHALL be binary (open/close only) — no swipe-to-dismiss or snap-point gestures. The drawer SHALL close when the user activates its close control or taps the backdrop, calling the same `onClose` callback used by the sidebar shell.

#### Scenario: Right panel renders as a bottom drawer on a coarse-pointer device

- **WHEN** the right panel is open on a coarse-pointer device
- **THEN** the panel renders as a fixed bottom-anchored drawer with a backdrop
- **AND** the panel does not render as a 320px sidebar

#### Scenario: Tapping the backdrop closes the drawer

- **WHEN** the bottom drawer is open
- **AND** the user taps the backdrop
- **THEN** the drawer closes via the same `onClose` callback the sidebar shell uses

#### Scenario: Drawer content is identical to sidebar content

- **WHEN** the right panel is open for a given image or folder
- **THEN** the content rendered inside the bottom drawer (on coarse-pointer devices) is the same `ImagePanelBody`/`FolderPanelContent` component used inside the sidebar (on fine-pointer devices)
