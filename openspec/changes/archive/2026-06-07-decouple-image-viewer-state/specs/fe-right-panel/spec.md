## MODIFIED Requirements

### Requirement: Right panel displays a thumbnail at the top

The system SHALL display the image's `thumbnail_url` at the top of the right panel. The thumbnail SHALL be rendered at full panel width with natural aspect ratio (not a fixed height). A close button (✕) SHALL be overlaid on the thumbnail (top-right corner). The thumbnail itself SHALL be a static display element with no click-to-open behavior.

#### Scenario: Thumbnail is shown at panel top

- **WHEN** the right panel is open for a selected image
- **THEN** the image thumbnail is displayed at the top of the panel at full panel width

#### Scenario: Clicking the thumbnail has no effect

- **WHEN** the user clicks the thumbnail in the right panel
- **THEN** no viewer or overlay opens
- **AND** the right panel remains as is

#### Scenario: Close button dismisses the panel

- **WHEN** the user clicks the ✕ close button overlaid on the thumbnail
- **THEN** the right panel closes and no image is selected
