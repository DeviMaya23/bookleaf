## MODIFIED Requirements

### Requirement: Images displayed in a masonry layout

The system SHALL render images using `MasonryLayout` with explicit round-robin column assignment. Each image card SHALL display the thumbnail at its natural aspect ratio derived from stored `width`/`height` fields. Image cards SHALL NOT have a visible border. The column count SHALL be derived from the gallery container's observed width divided by `TARGET_COL_WIDTH` (220px), updated reactively via `ResizeObserver`. Fixed CSS breakpoint column counts are removed.

#### Scenario: Image cards respect natural aspect ratio

- **WHEN** the image list contains images with varying dimensions
- **THEN** each card's thumbnail height reflects the image's natural aspect ratio

#### Scenario: Column count responds to container width

- **WHEN** the gallery container width changes (e.g. right panel opens or closes)
- **THEN** the column count updates to `Math.max(1, Math.floor(containerWidth / 220))`

#### Scenario: Cards have no border

- **WHEN** the image list contains images
- **THEN** no border is rendered around any image card
