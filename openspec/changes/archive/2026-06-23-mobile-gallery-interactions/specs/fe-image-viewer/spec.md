## MODIFIED Requirements

### Requirement: Double-clicking a gallery card opens the image viewer

The system SHALL open the image viewer when the user double-clicks an image card in the gallery on a fine-pointer device. `ImageGrid` SHALL expose an `onImageDoubleClick?: (image: Image) => void` callback prop. Double-clicking a card SHALL call both `onImageSelect` (selecting the image and opening the right panel if not already open) and `onImageDoubleClick`. `AppLayout` SHALL hold the viewer's displayed image as independent state (`viewerImage`), set to the double-clicked image in response to `onImageDoubleClick`. The viewer SHALL be considered open when `viewerImage` is non-null.

On coarse-pointer devices (`useIsCoarsePointer()` is true), `onImageDoubleClick` SHALL be a no-op — `viewerImage` SHALL never be set, even if a `dblclick` event is synthesized from a double-tap. Coarse-pointer devices open the lightbox (`fe-image-lightbox`) via single tap instead.

#### Scenario: Double-clicking a card opens the viewer on a fine-pointer device

- **WHEN** a user on a fine-pointer device double-clicks an image card in the gallery
- **THEN** `AppLayout` sets `viewerImage` to the double-clicked image
- **AND** the image viewer is rendered in place of the gallery grid
- **AND** the right panel is open showing the double-clicked image's metadata

#### Scenario: Double-clicking an already-selected image still opens the viewer on a fine-pointer device

- **WHEN** the right panel is already open for an image on a fine-pointer device
- **AND** the user double-clicks the same image card
- **THEN** the image viewer opens

#### Scenario: Single-click behavior is unchanged on a fine-pointer device

- **WHEN** a user on a fine-pointer device single-clicks an image card
- **THEN** only `onImageSelect` is called
- **AND** the viewer does NOT open

#### Scenario: A synthesized double-tap does not open the viewer on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device double-taps an image card, synthesizing `click` and `dblclick` events
- **THEN** `onImageDoubleClick` is a no-op
- **AND** `viewerImage` remains null
