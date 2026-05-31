### Requirement: MasonryLayout renders images in explicit round-robin columns

The system SHALL render images via a `MasonryLayout` component that assigns `item[i]` to `column[i % numCols]`. Column count SHALL be derived as `Math.max(1, Math.floor(containerWidth / TARGET_COL_WIDTH))` where `TARGET_COL_WIDTH = 220`. Container width SHALL be tracked reactively via `ResizeObserver` on the gallery container element.

#### Scenario: Column count adapts to container width

- **WHEN** the gallery container is 700px wide
- **THEN** images are distributed across 3 columns (`floor(700 / 220) = 3`)

#### Scenario: Column count adapts when right panel opens

- **WHEN** the right panel opens and the gallery container narrows to 440px
- **THEN** images are redistributed across 2 columns (`floor(440 / 220) = 2`)

#### Scenario: Column count never falls below 1

- **WHEN** the gallery container is narrower than `TARGET_COL_WIDTH`
- **THEN** images are displayed in a single column

### Requirement: Image height derived from stored aspect ratio

Each image card's height SHALL be computed as `colWidth / (image.width / image.height)`, so the image fills the column width while preserving its natural aspect ratio. If `image.width` or `image.height` is null, the image SHALL render with a 1:1 fallback aspect ratio.

#### Scenario: Portrait image fills column width at correct height

- **WHEN** an image has dimensions 790×1184 and the column width is 220px
- **THEN** the image renders at 220px wide and approximately 330px tall

#### Scenario: Null dimensions fall back to square

- **WHEN** an image has null width or height
- **THEN** the image renders as a square at the column width

### Requirement: Image card includes title below the thumbnail

Each image card SHALL display the image title in a text element below the thumbnail. The title SHALL be truncated to one line with an ellipsis if it overflows.

#### Scenario: Title rendered below image

- **WHEN** the gallery renders an image with a title
- **THEN** the title text appears below the thumbnail within the same card

### Requirement: MasonryLayout accepts a layoutMode prop seam

`ImageGrid` SHALL accept a `layoutMode` prop typed as `'masonry'` (expandable to `'justified' | 'grid'` in future). Only `'masonry'` is implemented; passing any other value SHALL render nothing and log a warning.

#### Scenario: Default layoutMode renders masonry

- **WHEN** `ImageGrid` is rendered without a `layoutMode` prop
- **THEN** the masonry layout is used
