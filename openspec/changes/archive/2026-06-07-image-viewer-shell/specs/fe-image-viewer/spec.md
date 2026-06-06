## ADDED Requirements

### Requirement: Double-clicking a gallery card opens the image viewer

The system SHALL open the image viewer when the user double-clicks an image card in the gallery. `ImageGrid` SHALL expose an `onImageDoubleClick?: (image: Image) => void` callback prop. Double-clicking a card SHALL call both `onImageSelect` (selecting the image and opening the right panel if not already open) and `onImageDoubleClick`. `AppLayout` SHALL set `viewerOpen: true` in response to `onImageDoubleClick`.

#### Scenario: Double-clicking a card opens the viewer

- **WHEN** the authenticated user double-clicks an image card in the gallery
- **THEN** `AppLayout` sets `viewerOpen` to `true`
- **AND** the image viewer is rendered in place of the gallery grid
- **AND** the right panel is open showing the double-clicked image's metadata

#### Scenario: Double-clicking an already-selected image still opens the viewer

- **WHEN** the right panel is already open for an image
- **AND** the user double-clicks the same image card
- **THEN** the image viewer opens

#### Scenario: Single-click behavior is unchanged

- **WHEN** the user single-clicks an image card
- **THEN** only `onImageSelect` is called
- **AND** the viewer does NOT open

---

### Requirement: Image viewer replaces the gallery in the main content area

The system SHALL conditionally render either `ImageGrid` or `ImageViewer` in the `<main>` element of `AppLayout` based on the `viewerOpen` state. The right panel SHALL remain visible alongside the viewer.

#### Scenario: Viewer occupies the main content area

- **WHEN** the viewer is open
- **THEN** `ImageGrid` is not rendered
- **AND** `ImageViewer` fills the full main content area
- **AND** the right panel remains visible on the right

#### Scenario: Gallery is restored when viewer closes

- **WHEN** the viewer is closed
- **THEN** `ImageGrid` is rendered again in the main content area

---

### Requirement: Image viewer displays the full-resolution image

The system SHALL fetch the full-resolution `image_url` via `GET /images/:id` (using React Query key `['image', image.id]`) when the viewer mounts. While the URL is loading, the viewer SHALL display the thumbnail as a placeholder. Once the full-res URL resolves, the viewer SHALL display the full-resolution image centered in the canvas area. The canvas area SHALL NOT specify a background color.

#### Scenario: Thumbnail shown while full-res URL is loading

- **WHEN** the viewer opens and `image_url` is not yet available
- **THEN** the thumbnail image is displayed centered in the canvas area

#### Scenario: Full-res image shown once URL resolves

- **WHEN** `GET /images/:id` returns a non-null `image_url`
- **THEN** the full-resolution image is displayed centered in the canvas area
- **AND** the thumbnail placeholder is replaced

#### Scenario: Image is centered with natural aspect ratio preserved

- **WHEN** the viewer displays an image
- **THEN** the image is centered in the canvas area
- **AND** the image's natural aspect ratio is preserved

---

### Requirement: Image viewer has a toolbar with transform controls

The system SHALL render a 44px toolbar at the top of the viewer. The toolbar SHALL contain, in order from left: a back button, a zoom range slider (range 5–800), a zoom percentage label, a separator, a flip horizontal button, a rotate 90° CW button, a 1:1 button, a flex spacer. In this phase all controls except the back button SHALL be rendered but non-functional (no state changes on interaction).

#### Scenario: Toolbar is rendered at the top of the viewer

- **WHEN** the viewer is open
- **THEN** a toolbar is visible at the top of the viewer
- **AND** the toolbar contains the back button, zoom slider, zoom %, flip, rotate, and 1:1 controls

#### Scenario: Zoom slider and percentage label are present

- **WHEN** the viewer is open
- **THEN** a range input is visible in the toolbar
- **AND** a percentage label is visible next to it

---

### Requirement: Image viewer shows a filename and dimensions badge

The system SHALL render a badge overlaid at the bottom-center of the canvas area displaying the image's filename (title) and pixel dimensions (`width × height`).

#### Scenario: Badge displays filename and dimensions

- **WHEN** the viewer is open for an image with known dimensions
- **THEN** a badge is visible at the bottom-center of the canvas area
- **AND** the badge displays the image title and dimensions in the format `<title> · <width> × <height>`

---

### Requirement: Image viewer is dismissed via back button or Esc key

The system SHALL close the viewer and return to the gallery when the user clicks the back button in the toolbar or presses the Esc key. Closing the viewer SHALL NOT close the right panel or change the selected image.

#### Scenario: Back button closes the viewer

- **WHEN** the user clicks the back button in the toolbar
- **THEN** `viewerOpen` is set to `false`
- **AND** the gallery grid is restored
- **AND** the right panel remains open with the same image selected

#### Scenario: Esc key closes the viewer

- **WHEN** the viewer is open
- **AND** the user presses the Esc key
- **THEN** `viewerOpen` is set to `false`
- **AND** the gallery grid is restored
