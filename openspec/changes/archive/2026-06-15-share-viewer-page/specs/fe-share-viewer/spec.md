## ADDED Requirements

### Requirement: Public share viewer route

The app SHALL expose a public route `/share/:token`, registered outside `AuthGuard` and nested under `PublicThemeLock` (`data-theme="warm"`), rendering a `SharePage` component.

#### Scenario: Share link is accessible without authentication

- **WHEN** a user (with or without a session) navigates to `/share/:token`
- **THEN** `SharePage` renders without being redirected to login
- **AND** the page is rendered with `data-theme="warm"`

### Requirement: Share data loading state

`SharePage` SHALL call `GET /share/:token` on mount. While the request is in flight, it SHALL render a centered loading indicator with no topbar, grid, or side panel.

#### Scenario: Initial load shows a spinner only

- **WHEN** `SharePage` mounts and the `GET /share/:token` request has not yet resolved
- **THEN** a centered loading spinner is shown
- **AND** no topbar, image grid, or side panel is rendered

### Requirement: Invalid or expired share token state

If `GET /share/:token` responds with `404 Not Found` (or any other error), `SharePage` SHALL render a centered error state consisting of an error icon, a message indicating the link is invalid or has expired, and a "Shared via Bookleaf" branding link. No topbar, image grid, or side panel SHALL be rendered.

#### Scenario: Unknown token shows an error state

- **WHEN** `GET /share/:token` responds with `404 Not Found`
- **THEN** `SharePage` renders a centered error icon and a message that the link is invalid or has expired
- **AND** a "Shared via Bookleaf" branding link is shown
- **AND** no topbar, image grid, or side panel is rendered

#### Scenario: Revoked token shows the same error state

- **WHEN** `GET /share/:token` is called with a token that was previously valid but has since been revoked
- **THEN** `SharePage` renders the same error state as an unknown token

### Requirement: Populated share page layout

When `GET /share/:token` succeeds and the folder has one or more images, `SharePage` SHALL render: a topbar containing the "Bookleaf" wordmark and the folder's name, a masonry image grid of the folder's images, and a `SharedFolderPanel` side panel.

#### Scenario: Folder with images renders the full layout

- **WHEN** `GET /share/:token` succeeds for a folder containing images
- **THEN** the topbar shows "Bookleaf" and the folder's name
- **AND** the folder's images are rendered in a masonry grid
- **AND** `SharedFolderPanel` is rendered alongside the grid

### Requirement: Empty folder state

When `GET /share/:token` succeeds and the folder has zero images, `SharePage` SHALL render the full layout (topbar and `SharedFolderPanel`), with the image area showing an empty-state message ("No images here yet") instead of a grid, matching the empty-state pattern used by the authenticated gallery (`ImageGrid`).

#### Scenario: Empty shared folder shows an empty-state message

- **WHEN** `GET /share/:token` succeeds for a folder with zero images
- **THEN** the topbar and `SharedFolderPanel` are rendered
- **AND** the image area shows an icon and the message "No images here yet" instead of a grid

### Requirement: Masonry grid reuse for shared images

`SharePage` SHALL render the shared folder's images using the existing `MasonryLayout` component, mapping each share-response image (`title`, `thumbnail_url`, `width`, `height`) into the `Image` shape `MasonryLayout` expects via a local adapter, without modifying `MasonryLayout`'s props or `masonry.ts`. Card heights SHALL be derived from each image's `width`/`height` as `MasonryLayout` already does, falling back to a 1:1 ratio when either is `null`.

#### Scenario: Images render with aspect-ratio-derived heights

- **WHEN** the share response includes images with non-null `width` and `height`
- **THEN** each image card is rendered by `MasonryLayout` at a height derived from that image's aspect ratio

#### Scenario: Images without dimensions fall back to square

- **WHEN** a shared image has `width` or `height` equal to `null`
- **THEN** that image's card renders at a 1:1 aspect ratio, per `MasonryLayout`'s existing fallback behavior

### Requirement: Hover download button on image card

Each image card in the masonry grid SHALL show a circular download button, hidden by default and revealed on hover, that downloads the image's full-resolution file directly via its `download_url`. The share API response SHALL include a `download_url` per image, distinct from `full_res_url`, generated with a `Content-Disposition: attachment` header so the browser downloads the file (rather than navigating to or opening it) regardless of origin.

#### Scenario: Hovering a card reveals a download button

- **WHEN** a user hovers over an image card
- **THEN** a circular download button becomes visible on the card

#### Scenario: Clicking the download button downloads the full-resolution image

- **WHEN** a user clicks an image card's download button
- **THEN** the browser downloads the image from that image's `download_url` as a file, rather than opening or navigating to it

### Requirement: Image card selection outline on click

Clicking an image card (on the card itself, not its download button) SHALL toggle a thin outline around that card as visual feedback. This is a purely visual toggle: it does not open the lightbox, navigate, or trigger any other action.

#### Scenario: Clicking a card shows a selection outline

- **WHEN** a user clicks an image card
- **THEN** that card renders with a thin outline
- **AND** no navigation or other action occurs

#### Scenario: Clicking a selected card again removes the outline

- **WHEN** a user clicks an image card that already shows the selection outline
- **THEN** the outline is removed

### Requirement: Lightbox on double-click

Double-clicking an image card SHALL open a fullscreen lightbox overlay showing that image's full-resolution version (`full_res_url`). The lightbox SHALL support: closing via a close button, the Escape key, or a click on the backdrop (but not on the image or controls); navigating to the previous/next image via on-screen buttons and the Left/Right arrow keys, disabled at the first/last image; and an image position counter.

#### Scenario: Double-click opens the lightbox at that image

- **WHEN** a user double-clicks an image card
- **THEN** a fullscreen lightbox opens showing that image's full-resolution version

#### Scenario: Arrow keys navigate within bounds

- **WHEN** the lightbox is open and the user presses the Right arrow key on an image that is not the last
- **THEN** the lightbox shows the next image
- **AND WHEN** the lightbox is showing the last image
- **THEN** pressing the Right arrow key has no effect

#### Scenario: Escape and backdrop click close the lightbox

- **WHEN** the lightbox is open and the user presses Escape, or clicks the backdrop outside the image
- **THEN** the lightbox closes

#### Scenario: Clicking the image does not close the lightbox

- **WHEN** the lightbox is open and the user clicks on the displayed image
- **THEN** the lightbox remains open

### Requirement: SharedFolderPanel content

`SharedFolderPanel` SHALL display, in a read-only form: the folder's name, the image count, the folder's notes (or "No notes added" if the folder has no description), an export button, and a "Shared via Bookleaf" branding line. None of these fields SHALL be editable.

#### Scenario: Panel shows folder metadata and notes

- **WHEN** `SharedFolderPanel` is rendered for a folder with a description and N images
- **THEN** it shows the folder's name, "N images", and the folder's notes text
- **AND** it shows a "Shared via Bookleaf" branding line

#### Scenario: Folder without notes shows empty-notes message

- **WHEN** `SharedFolderPanel` is rendered for a folder with no description
- **THEN** it shows "No notes added" in place of notes text

### Requirement: Export folder from share panel

`SharedFolderPanel`'s export button SHALL trigger a download of `GET /share/:token/export` (a zip of the shared folder's images) when clicked, and SHALL be disabled when the shared folder has zero images.

#### Scenario: Clicking export downloads the folder as a zip

- **WHEN** a user clicks the export button on a shared folder with at least one image
- **THEN** the browser downloads the response of `GET /share/:token/export` as a zip file

#### Scenario: Export is disabled for an empty folder

- **WHEN** `SharedFolderPanel` is rendered for a folder with zero images
- **THEN** the export button is disabled
