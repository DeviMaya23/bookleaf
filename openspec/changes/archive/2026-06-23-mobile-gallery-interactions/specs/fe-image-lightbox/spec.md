## ADDED Requirements

### Requirement: Lightbox opens on single tap on coarse-pointer devices

The system SHALL open the mobile image lightbox (`ImageLightbox` component) when the user taps an image card on a coarse-pointer device. `AppLayout` SHALL hold the lightbox's displayed image as independent state (`lightboxImage`), distinct from `selectedImage` and `viewerImage`. The lightbox SHALL be considered open when `lightboxImage` is non-null. This requirement applies only when `useIsCoarsePointer()` is true; on fine-pointer devices, single tap/click continues to open the right panel per `fe-right-panel`, and the lightbox is never opened by tap/click.

#### Scenario: Tapping an image card opens the lightbox on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device taps an image card in the gallery
- **THEN** `AppLayout` sets `lightboxImage` to the tapped image
- **AND** the lightbox is rendered in place of the gallery grid

#### Scenario: Tapping an image card does not open the lightbox on a fine-pointer device

- **WHEN** a user on a fine-pointer device clicks an image card
- **THEN** `lightboxImage` remains null
- **AND** the right panel opens instead, per `fe-right-panel`

---

### Requirement: Lightbox displays the full-resolution image at fit-to-screen size with no zoom

The system SHALL fetch the full-resolution `image_url` via `GET /images/:id` when the lightbox mounts, displaying the thumbnail as a placeholder while it loads, consistent with the loading pattern used by the desktop viewer (`fe-image-viewer`). The image SHALL be displayed centered at fit-to-screen size (`object-fit: contain`) with its natural aspect ratio preserved. The lightbox SHALL NOT provide zoom (pinch, double-tap, or slider), rotate, flip, or drag-to-pan — this is best-effort static viewing, not a parity feature with the desktop viewer's transform toolbar.

#### Scenario: Thumbnail shown while full-res URL is loading

- **WHEN** the lightbox opens and `image_url` is not yet available
- **THEN** the thumbnail image is displayed centered and fit-to-screen

#### Scenario: Full-res image shown once URL resolves

- **WHEN** `GET /images/:id` returns a non-null `image_url`
- **THEN** the full-resolution image is displayed centered and fit-to-screen
- **AND** the thumbnail placeholder is replaced

#### Scenario: No zoom or transform controls are present

- **WHEN** the lightbox is open
- **THEN** no zoom slider, rotate, flip, or pan controls are rendered
- **AND** pinch or double-tap gestures on the image have no zoom effect

---

### Requirement: Lightbox is dismissed via a close control

The system SHALL close the lightbox (set `lightboxImage` to `null`) and return to the gallery when the user activates a visible close control. Closing the lightbox SHALL NOT open the right panel or change `selectedImage`.

#### Scenario: Close control closes the lightbox

- **WHEN** the lightbox is open
- **AND** the user activates the close control
- **THEN** `lightboxImage` is set to `null`
- **AND** the gallery grid is restored
- **AND** `selectedImage` is unchanged

---

### Requirement: Lightbox and desktop image viewer are mutually exclusive

The system SHALL ensure that, for a given device, only one of `viewerImage` (desktop viewer, `fe-image-viewer`) or `lightboxImage` (mobile lightbox) is ever set to a non-null value, determined entirely by `useIsCoarsePointer()` at the time an image card is tapped/clicked. Navigating to a different folder/view SHALL reset `lightboxImage` to `null`, mirroring the existing reset of `viewerImage` and `selectedImage` on view change.

#### Scenario: Coarse-pointer tap never sets viewerImage

- **WHEN** a user on a coarse-pointer device taps an image card
- **THEN** `lightboxImage` is set to that image
- **AND** `viewerImage` remains null

#### Scenario: Navigating to a different folder closes an open lightbox

- **WHEN** the lightbox is open showing an image from the current folder
- **AND** the user selects a different folder in the sidebar
- **THEN** the lightbox closes
- **AND** the gallery grid for the newly selected folder is shown
