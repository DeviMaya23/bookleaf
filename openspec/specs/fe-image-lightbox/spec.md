## ADDED Requirements

### Requirement: Clicking an image card opens a lightbox overlay
The system SHALL open a full-screen lightbox Dialog when the user left-clicks an image card in the gallery grid.

#### Scenario: Left-click opens lightbox
- **WHEN** the authenticated user left-clicks an image card
- **THEN** a full-screen overlay Dialog opens
- **AND** the Dialog displays a loading spinner while the image URL is being fetched

### Requirement: Lightbox fetches the presigned URL on open
The system SHALL call `GET /images/:id` when a lightbox is opened and use the returned `image_url` to display the full-resolution image.

#### Scenario: High-res image loads after fetch
- **WHEN** the lightbox opens for an image
- **THEN** the app calls `GET /images/<id>`
- **AND** once the response resolves, the spinner is replaced by the full-resolution image rendered via `<img src={image_url} />`
- **AND** the image is constrained to `max-h-[90vh] max-w-[90vw]` with `object-contain`

#### Scenario: Spinner shown while fetching
- **WHEN** the lightbox is open and the `GET /images/:id` request is in flight
- **THEN** a centered loading spinner (`Loader2`) is displayed inside the Dialog

### Requirement: Lightbox is dismissable via multiple interactions
The system SHALL allow the user to close the lightbox by clicking the X button, clicking the backdrop, or pressing ESC.

#### Scenario: X button closes lightbox
- **WHEN** the lightbox is open
- **AND** the user clicks the X button in the top-right corner
- **THEN** the lightbox closes and the gallery is shown

#### Scenario: Backdrop click closes lightbox
- **WHEN** the lightbox is open
- **AND** the user clicks outside the Dialog content area
- **THEN** the lightbox closes

#### Scenario: ESC key closes lightbox
- **WHEN** the lightbox is open
- **AND** the user presses the ESC key
- **THEN** the lightbox closes

### Requirement: Lightbox shows image only, no metadata
The system SHALL display only the full-resolution image in the lightbox. Title, description, and other metadata SHALL NOT be rendered visibly. A visually hidden accessible title SHALL be present for screen readers.

#### Scenario: Lightbox content is image-only
- **WHEN** the lightbox is open and the image has loaded
- **THEN** only the image is visible inside the Dialog
- **AND** no title, description, or other metadata text is rendered visibly
