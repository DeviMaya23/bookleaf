## MODIFIED Requirements

### Requirement: Right-click context menu with delete option

The system SHALL show a context menu with a "Delete" option when the user right-clicks an image card on a fine-pointer device, or long-presses an image card on a coarse-pointer device (the existing touch trigger for this context menu). On coarse-pointer devices, the context menu SHALL additionally include a "View details" item above "Delete", which opens the right panel for that image as a bottom drawer (per `fe-right-panel`). The "View details" item SHALL NOT be rendered on fine-pointer devices, since clicking the image card already opens the right panel there.

#### Scenario: Right-click shows context menu on a fine-pointer device

- **WHEN** a user on a fine-pointer device right-clicks an image card
- **THEN** a context menu appears with a "Delete" option
- **AND** no "View details" item is present

#### Scenario: Long-press shows context menu with View details on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device long-presses an image card
- **THEN** a context menu appears with "View details" above "Delete"

#### Scenario: Selecting View details opens the right panel as a bottom drawer

- **WHEN** a user on a coarse-pointer device selects "View details" from an image card's context menu
- **THEN** the right panel opens as a bottom drawer showing that image's metadata, per `fe-right-panel`
