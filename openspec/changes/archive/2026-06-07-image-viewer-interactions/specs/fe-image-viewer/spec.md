## MODIFIED Requirements

### Requirement: Image viewer has a toolbar with transform controls

The system SHALL render a 44px toolbar at the top of the viewer. The toolbar SHALL contain, in order from left: a back button, a zoom range slider (range 5–800), a zoom percentage label, a separator, a flip horizontal button, a rotate 90° CW button, a 1:1 button, a flex spacer. All controls SHALL be fully functional.

#### Scenario: Toolbar is rendered at the top of the viewer

- **WHEN** the viewer is open
- **THEN** a toolbar is visible at the top of the viewer
- **AND** the toolbar contains the back button, zoom slider, zoom %, flip, rotate, and 1:1 controls

#### Scenario: Zoom slider and percentage label are present and reflect current zoom

- **WHEN** the viewer is open
- **THEN** a range input is visible in the toolbar
- **AND** the percentage label displays the current zoom level
- **AND** the slider position reflects the current zoom level
