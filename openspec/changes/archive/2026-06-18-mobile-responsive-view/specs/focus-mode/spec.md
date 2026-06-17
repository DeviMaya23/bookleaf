## MODIFIED Requirements

### Requirement: Focus mode toggle is available in the gallery and image viewer toolbars

The system SHALL render a "Focus" toggle button as the leftmost element of
the top toolbar in the gallery view (to the left of the search input) and
as the leftmost element of the image viewer header (to the left of the Back
button). The toggle SHALL reflect its current on/off state visually (e.g.
`aria-pressed`). Below the `sm` breakpoint, the Focus toggle SHALL NOT be
rendered in either toolbar.

#### Scenario: Focus toggle is visible in the gallery toolbar

- **WHEN** the gallery view is rendered at or above the `sm` breakpoint
- **THEN** a Focus toggle button is visible to the left of the search input

#### Scenario: Focus toggle is visible in the image viewer header

- **WHEN** the image viewer is open at or above the `sm` breakpoint
- **THEN** a Focus toggle button is visible to the left of the Back button

#### Scenario: Toggle reflects active state

- **WHEN** focus mode is active at or above the `sm` breakpoint
- **THEN** the Focus toggle button is shown in a visually pressed/active state in whichever toolbar is currently rendered

#### Scenario: Focus toggle is hidden below the breakpoint

- **WHEN** the viewport width is below the `sm` breakpoint
- **THEN** the Focus toggle button is not rendered in either the gallery toolbar or the image viewer header
