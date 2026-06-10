## MODIFIED Requirements

### Requirement: Right panel opens when an image card is clicked

The system SHALL render a 320px right panel (`RightPanel` component) as a sibling to the main content area in `AppLayout`. The panel SHALL be hidden when no image is selected. When the user clicks an image card, the panel SHALL become visible and display that image's metadata. The panel SHALL NOT be rendered while focus mode is active, even if an image is selected.

#### Scenario: Clicking an image card opens the right panel

- **WHEN** the authenticated user clicks an image card in the gallery
- **THEN** the right panel becomes visible on the right side of the layout
- **AND** the panel displays the selected image's metadata

#### Scenario: Panel is hidden when no image is selected

- **WHEN** no image has been selected
- **THEN** the right panel is not rendered in the layout

#### Scenario: Panel stays hidden while focus mode is active

- **WHEN** focus mode is active
- **AND** the user clicks an image card
- **THEN** the right panel is not rendered, even though an image is now selected
