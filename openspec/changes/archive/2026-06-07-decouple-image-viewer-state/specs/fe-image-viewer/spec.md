## MODIFIED Requirements

### Requirement: Double-clicking a gallery card opens the image viewer

The system SHALL open the image viewer when the user double-clicks an image card in the gallery. `ImageGrid` SHALL expose an `onImageDoubleClick?: (image: Image) => void` callback prop. Double-clicking a card SHALL call both `onImageSelect` (selecting the image and opening the right panel if not already open) and `onImageDoubleClick`. `AppLayout` SHALL hold the viewer's displayed image as independent state (`viewerImage`), set to the double-clicked image in response to `onImageDoubleClick`. The viewer SHALL be considered open when `viewerImage` is non-null.

#### Scenario: Double-clicking a card opens the viewer

- **WHEN** the authenticated user double-clicks an image card in the gallery
- **THEN** `AppLayout` sets `viewerImage` to the double-clicked image
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

The system SHALL conditionally render either `ImageGrid` or `ImageViewer` in the `<main>` element of `AppLayout` based on whether `viewerImage` is non-null. The right panel SHALL remain visible alongside the viewer whenever the right panel has a selected image, independent of whether the viewer is open.

#### Scenario: Viewer occupies the main content area

- **WHEN** `viewerImage` is non-null
- **THEN** `ImageGrid` is not rendered
- **AND** `ImageViewer` fills the full main content area
- **AND** the right panel remains visible on the right if it has a selected image

#### Scenario: Gallery is restored when viewer closes

- **WHEN** `viewerImage` becomes null
- **THEN** `ImageGrid` is rendered again in the main content area

---

### Requirement: Image viewer is dismissed via back button or Esc key

The system SHALL close the viewer (set `viewerImage` to `null`) and return to the gallery when the user clicks the back button in the toolbar or presses the Esc key. Closing the viewer this way SHALL NOT close the right panel or change `selectedImage`.

#### Scenario: Back button closes the viewer

- **WHEN** the user clicks the back button in the toolbar
- **THEN** `viewerImage` is set to `null`
- **AND** the gallery grid is restored
- **AND** the right panel remains open with the same image selected

#### Scenario: Esc key closes the viewer

- **WHEN** the viewer is open
- **AND** the user presses the Esc key
- **THEN** `viewerImage` is set to `null`
- **AND** the gallery grid is restored

## ADDED Requirements

### Requirement: Closing the right panel does not close the image viewer

The system SHALL keep the image viewer open when the right panel is closed while the viewer is displaying an image. Closing the right panel SHALL only clear `selectedImage` (and the right panel itself); it SHALL NOT clear `viewerImage`. The viewer's containing area SHALL widen to fill the space vacated by the right panel via the existing flex layout, with no recalculation of the image's zoom or fit.

#### Scenario: Viewer remains open and widens when the right panel is closed

- **WHEN** the image viewer is open
- **AND** the user closes the right panel
- **THEN** the image viewer remains open showing the same image
- **AND** the viewer's containing area widens to fill the space the right panel occupied
- **AND** the image's current zoom level is unchanged

---

### Requirement: Navigating to a different folder dismisses the viewer and right panel

The system SHALL close the image viewer and the right panel when the active folder/view changes (e.g. the user selects a different folder in the sidebar). `AppLayout` SHALL reset `viewerImage` and `selectedImage` to `null` in response to a change in the active view, returning the main content area to the gallery grid for the newly selected folder.

#### Scenario: Selecting a different folder closes an open viewer

- **WHEN** the image viewer is open showing an image from the current folder
- **AND** the user selects a different folder in the sidebar
- **THEN** the image viewer closes
- **AND** the gallery grid for the newly selected folder is shown

#### Scenario: Selecting a different folder closes an open right panel

- **WHEN** the right panel is open showing an image's metadata
- **AND** the user selects a different folder in the sidebar
- **THEN** the right panel closes

---

### Requirement: Deleting the viewed image closes the viewer

The system SHALL close the image viewer when the image it is currently displaying is deleted, regardless of whether the right panel is showing the same image or a different one. `AppLayout`'s image-deletion handler SHALL check `viewerImage?.id` independently of `selectedImage?.id` and reset `viewerImage` to `null` when they match.

#### Scenario: Deleting the image open in the viewer closes the viewer

- **WHEN** the image viewer is open showing a given image
- **AND** that image is deleted
- **THEN** the image viewer closes
- **AND** the gallery grid is shown

#### Scenario: Deleting the viewed image closes the viewer even if the right panel shows a different image

- **WHEN** the image viewer is open showing image A
- **AND** the right panel is showing image B's metadata
- **AND** image A is deleted
- **THEN** the image viewer closes
- **AND** the right panel remains open showing image B's metadata
