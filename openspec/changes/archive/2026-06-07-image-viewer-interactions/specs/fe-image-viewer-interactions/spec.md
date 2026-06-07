## ADDED Requirements

### Requirement: Viewer opens with image scaled to fit the canvas

The system SHALL calculate a fit zoom on mount such that the image fills ~90% of the canvas area while preserving aspect ratio. The fit zoom SHALL account for dimension swap when rotation is 90° or 270°. Pan SHALL be initialised to (0, 0).

#### Scenario: Image fits the canvas on open

- **WHEN** the viewer mounts for any image
- **THEN** the image is scaled to fit within the canvas area at approximately 90% of available space
- **AND** the image is centered with no pan offset

---

### Requirement: Mouse wheel zooms centered on the cursor position

The system SHALL zoom in or out when the user scrolls the mouse wheel over the canvas area. Zooming SHALL be anchored to the cursor position — the image point under the cursor SHALL remain stationary as zoom changes. Zoom SHALL be clamped to the range 5%–800%.

#### Scenario: Scroll up zooms in, anchored to cursor

- **WHEN** the user scrolls up over the canvas
- **THEN** the zoom level increases
- **AND** the image point under the cursor does not move

#### Scenario: Scroll down zooms out, anchored to cursor

- **WHEN** the user scrolls down over the canvas
- **THEN** the zoom level decreases
- **AND** the image point under the cursor does not move

#### Scenario: Zoom is clamped at minimum (5%)

- **WHEN** the user scrolls down while zoom is at 5%
- **THEN** zoom does not decrease below 5%

#### Scenario: Zoom is clamped at maximum (800%)

- **WHEN** the user scrolls up while zoom is at 800%
- **THEN** zoom does not increase above 800%

---

### Requirement: Drag pans the image

The system SHALL allow the user to drag the image within the canvas area by clicking and dragging. The cursor SHALL display as `grab` when idle and `grabbing` while dragging.

#### Scenario: Dragging moves the image

- **WHEN** the user clicks and drags within the canvas area
- **THEN** the image moves with the cursor
- **AND** the cursor displays as `grabbing`

#### Scenario: Releasing the mouse stops panning

- **WHEN** the user releases the mouse button
- **THEN** the image stays at its new position
- **AND** the cursor returns to `grab`

---

### Requirement: Rotate button rotates 90° clockwise and resets to fit

The system SHALL rotate the image 90° clockwise per click of the rotate button, cycling through 0°, 90°, 180°, and 270°. On each rotation, zoom SHALL reset to the fit zoom for the new orientation and pan SHALL reset to (0, 0).

#### Scenario: Clicking rotate advances rotation by 90°

- **WHEN** the user clicks the rotate button
- **THEN** the image rotates 90° clockwise
- **AND** zoom resets to fit the rotated image within the canvas
- **AND** pan resets to (0, 0)

#### Scenario: Rotation wraps at 270° back to 0°

- **WHEN** the image is at 270° and the user clicks rotate
- **THEN** the image returns to 0°

---

### Requirement: Flip button toggles horizontal flip

The system SHALL toggle a horizontal flip on the image when the user clicks the flip button. Flip SHALL compose correctly with any current rotation — it mirrors the image as the user currently sees it.

#### Scenario: Clicking flip mirrors the image horizontally

- **WHEN** the user clicks the flip button while flip is off
- **THEN** the image is mirrored horizontally

#### Scenario: Clicking flip again restores the original orientation

- **WHEN** the user clicks the flip button while flip is on
- **THEN** the horizontal mirror is removed

#### Scenario: Flip combines correctly with rotation

- **WHEN** the image is rotated and the user clicks flip
- **THEN** the image is mirrored as currently viewed (not as the raw unrotated image)

---

### Requirement: Zoom slider is two-way with wheel zoom

The system SHALL keep the zoom range slider and the mouse wheel zoom in sync. Moving the slider SHALL update zoom. Zooming via wheel SHALL update the slider's displayed value. The slider range SHALL be 5–800 (representing percent).

#### Scenario: Sliding updates zoom

- **WHEN** the user drags the zoom slider to a new position
- **THEN** the zoom level changes to the corresponding percentage
- **AND** the zoom percentage label updates

#### Scenario: Wheel zoom updates slider

- **WHEN** the user zooms via mouse wheel
- **THEN** the slider position moves to reflect the new zoom level

---

### Requirement: 1:1 button resets zoom to 100% and clears pan

The system SHALL set zoom to 1.0 (100%) and pan to (0, 0) when the user clicks the 1:1 button.

#### Scenario: 1:1 resets zoom and pan

- **WHEN** the user clicks the 1:1 button
- **THEN** zoom is set to 100%
- **AND** pan is reset to (0, 0)
- **AND** the zoom slider and percentage label reflect 100%

---

### Requirement: All transforms reset when the viewer's image changes

The system SHALL reset zoom to fit, pan to (0, 0), rotation to 0°, and flip to off whenever the `image` prop changes (i.e. a different image is opened in the viewer).

#### Scenario: Opening a different image resets all transforms

- **WHEN** the `image` prop changes to a different image
- **THEN** zoom resets to the fit zoom for the new image
- **AND** pan resets to (0, 0)
- **AND** rotation resets to 0°
- **AND** flip resets to off
