## ADDED Requirements

### Requirement: Right panel shows a Tags section

The system SHALL render a Tags section in `RightPanel` between the Source URL section and the Details section. The section SHALL contain a `TagInput` component pre-populated with the image's current tags.

#### Scenario: Tags section is present in the right panel

- **WHEN** the right panel opens for any image
- **THEN** a Tags section is visible between Source and Details

#### Scenario: Tags section shows current image tags

- **WHEN** the right panel opens for an image that has associated tags
- **THEN** the TagInput renders each tag as a pill

#### Scenario: Tags section is empty for an untagged image

- **WHEN** the right panel opens for an image with no tags
- **THEN** the TagInput shows an empty input with placeholder text
