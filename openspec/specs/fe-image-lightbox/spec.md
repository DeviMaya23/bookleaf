### Requirement: Clicking the right panel thumbnail opens the lightbox overlay

The system SHALL open a full-screen lightbox Dialog when the user clicks the thumbnail image inside the right panel. The lightbox SHALL NOT open directly from an image card click in the gallery (that click opens the right panel instead).

#### Scenario: Thumbnail click in right panel opens lightbox

- **WHEN** the right panel is open for a selected image
- **AND** the authenticated user clicks the thumbnail at the top of the panel
- **THEN** a full-screen overlay Dialog opens
- **AND** the Dialog displays a loading spinner while the image URL is being fetched

### Requirement: Lightbox fetches the presigned URL on open

The system SHALL call `GET /images/:id` when the lightbox is opened from the panel thumbnail and use the returned `image_url` to display the full-resolution image.

#### Scenario: High-res image loads after fetch

- **WHEN** the lightbox opens from a thumbnail click
- **THEN** the app calls `GET /images/<id>`
- **AND** once the response resolves, the spinner is replaced by the full-resolution image rendered via `<img src={image_url} />`
- **AND** the image is constrained to `max-h-[90vh] max-w-[90vw]` with `object-contain`

#### Scenario: Spinner shown while fetching

- **WHEN** the lightbox is open and the `GET /images/:id` request is in flight
- **THEN** a centered loading spinner is displayed inside the Dialog

### Requirement: Lightbox is dismissable via multiple interactions

The system SHALL allow the user to close the lightbox by clicking the X button, clicking the backdrop, or pressing ESC. Closing the lightbox SHALL NOT close the right panel.

#### Scenario: X button closes lightbox but keeps panel open

- **WHEN** the lightbox is open
- **AND** the user clicks the X button
- **THEN** the lightbox closes
- **AND** the right panel remains visible with the same selected image

#### Scenario: Backdrop click closes lightbox

- **WHEN** the lightbox is open
- **AND** the user clicks outside the Dialog content area
- **THEN** the lightbox closes

#### Scenario: ESC key closes lightbox

- **WHEN** the lightbox is open
- **AND** the user presses the ESC key
- **THEN** the lightbox closes

### Requirement: Lightbox shows image only, no metadata

The system SHALL display only the full-resolution image in the lightbox. A visually hidden accessible title SHALL be present for screen readers.

#### Scenario: Lightbox content is image-only

- **WHEN** the lightbox is open and the image has loaded
- **THEN** only the image is visible inside the Dialog
- **AND** no title, description, or other metadata text is rendered visibly
