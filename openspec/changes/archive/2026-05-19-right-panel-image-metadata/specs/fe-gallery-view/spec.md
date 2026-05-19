## MODIFIED Requirements

### Requirement: Images displayed in a Pinterest-style masonry layout

The system SHALL render images in a CSS `column-count` masonry layout. Each image card SHALL display the thumbnail at its natural aspect ratio (no fixed height). Image cards SHALL NOT have a visible border. The column count SHALL be responsive: 2 columns on mobile, 3 on medium viewports, 4 on large viewports (≥ 1024px). Cards SHALL use `break-inside: avoid` to prevent column breaks within a card.

#### Scenario: Image cards respect natural aspect ratio

- **WHEN** the image list contains images with varying dimensions
- **THEN** each card's thumbnail height reflects the image's natural aspect ratio

#### Scenario: Masonry layout uses correct column counts

- **WHEN** the viewport is desktop width (≥ 1024px)
- **THEN** images are displayed in 4 columns

#### Scenario: Masonry collapses on mobile

- **WHEN** the viewport is mobile width (< 768px)
- **THEN** images are displayed in 2 columns

#### Scenario: Cards have no border

- **WHEN** the image list contains images
- **THEN** no border is rendered around any image card

---

### Requirement: Clicking an image card opens the right panel

The system SHALL call the `onImageSelect` callback prop with the clicked image when the user left-clicks an image card. The gallery SHALL NOT directly manage the lightbox; the right panel is responsible for lightbox triggering.

#### Scenario: Left-click on image card selects the image

- **WHEN** the authenticated user left-clicks an image card
- **THEN** the `onImageSelect` callback is called with that image
- **AND** the right panel opens showing the selected image's metadata

---

### Requirement: Root route displays unfoldered images

The system SHALL fetch images with `unfiled=true` when the user is on the root path (`/`) and render them in the masonry gallery. The response envelope `{ images, next_cursor }` SHALL be handled for pagination.

#### Scenario: Navigating to root loads unfoldered images

- **WHEN** the authenticated user navigates to `/`
- **THEN** the app calls `GET /images?unfiled=true`
- **AND** the returned images are displayed in the masonry gallery

### Requirement: Folder route displays folder images

The system SHALL fetch images for a specific folder when the user is on `/folders/:folder_id` and render them in the masonry gallery.

#### Scenario: Navigating to a folder route loads that folder's images

- **WHEN** the authenticated user navigates to `/folders/:folder_id`
- **THEN** the app calls `GET /images?folder_id=<folder_id>`
- **AND** the returned images for that folder are displayed in the masonry gallery

### Requirement: Loading state shown during fetch

The system SHALL display a loading spinner while the image list is being fetched.

#### Scenario: Spinner shown while fetching

- **WHEN** the image list request is in flight
- **THEN** a spinner is displayed in the main content area

### Requirement: Empty state shown when no images exist

The system SHALL display an empty state message when the image list response is empty.

#### Scenario: Empty state shown with no images

- **WHEN** the image list response returns zero images
- **THEN** the message "No images here yet" is displayed with an image icon

### Requirement: Paginated image loading with "Load more"

The system SHALL use `useInfiniteQuery` to fetch images in pages. A "Load more" button SHALL be shown when a `next_cursor` is present in the last page's response. Clicking it SHALL fetch the next page. When switching folders, the accumulated pages SHALL be reset.

#### Scenario: Load more button shown when next page exists

- **WHEN** the image list response contains a non-null `next_cursor`
- **THEN** a "Load more" button is displayed below the masonry gallery

#### Scenario: Clicking Load more fetches next page

- **WHEN** the user clicks the "Load more" button
- **THEN** the app calls `GET /images?folder_id=<current>&cursor=<next_cursor>`
- **AND** the newly fetched images are appended to the existing gallery

#### Scenario: Changing folder resets pagination

- **WHEN** the user navigates to a different folder
- **THEN** the accumulated pages are discarded and only the first page is shown

### Requirement: Right-click context menu with delete option

The system SHALL show a context menu with a "Delete" option when the user right-clicks an image card.

#### Scenario: Right-click shows context menu

- **WHEN** the user right-clicks an image card
- **THEN** a context menu appears with a "Delete" option

### Requirement: Delete image with confirmation dialog

The system SHALL show a confirmation dialog before deleting an image. Upon confirmation, the system SHALL call `DELETE /images/:id` and refresh the image list.

#### Scenario: Confirming delete removes the image

- **WHEN** the user confirms deletion in the dialog
- **THEN** the app calls `DELETE /images/<id>`
- **AND** the image list is refreshed and the deleted image no longer appears

#### Scenario: Cancelling delete keeps the image

- **WHEN** the user cancels the delete dialog
- **THEN** no delete request is made
- **AND** the image remains in the gallery
