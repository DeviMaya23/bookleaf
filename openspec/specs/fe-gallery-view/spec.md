### Requirement: Root route displays unfoldered images
The system SHALL fetch images with `unfiled=true` when the user is on the root path (`/`) and render them in the masonry gallery. The response envelope `{ images, next_cursor }` SHALL be handled for pagination.

#### Scenario: Navigating to root loads unfoldered images
- **WHEN** the authenticated user navigates to `/`
- **THEN** the app calls `GET /images?unfiled=true`
- **AND** the returned images are displayed in the masonry gallery

### Requirement: Folder route displays folder images
The system SHALL fetch images for a specific folder when the user is on `/folders/:folder_id` and render them in the masonry gallery. The response envelope `{ images, next_cursor }` SHALL be handled for pagination.

#### Scenario: Navigating to a folder route loads that folder's images
- **WHEN** the authenticated user navigates to `/folders/:folder_id`
- **THEN** the app calls `GET /images?folder_id=<folder_id>`
- **AND** the returned images for that folder are displayed in the masonry gallery

### Requirement: Folder sidebar navigates via URL
The system SHALL navigate to `/folders/:folder_id` when the user clicks a folder in the sidebar, and navigate to `/` when the user clicks "Unsorted".

#### Scenario: Clicking a folder updates the URL and loads folder images
- **WHEN** the user clicks a folder item in the sidebar
- **THEN** the URL changes to `/folders/<folder_id>`
- **AND** the gallery fetches and displays images for that folder

#### Scenario: Clicking Unsorted navigates to root
- **WHEN** the user clicks the "Unsorted" item in the sidebar
- **THEN** the URL changes to `/`
- **AND** the gallery fetches and displays unfoldered images

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

### Requirement: Clicking an image card opens the right panel
The system SHALL call the `onImageSelect` callback prop with the clicked image when the user left-clicks an image card. The gallery SHALL NOT directly manage the lightbox; the right panel is responsible for lightbox triggering.

#### Scenario: Left-click on image card selects the image
- **WHEN** the authenticated user left-clicks an image card
- **THEN** the `onImageSelect` callback is called with that image
- **AND** the right panel opens showing the selected image's metadata

### Requirement: Paginated image loading with "Load more"
The system SHALL use `useInfiniteQuery` to fetch images in pages. A "Load more" button SHALL be shown when a `next_cursor` is present in the last page's response. Clicking it SHALL fetch the next page by passing `cursor=<next_cursor>` to `GET /images`. When switching folders (or navigating to root), the accumulated pages SHALL be reset.

#### Scenario: Load more button shown when next page exists
- **WHEN** the image list response contains a non-null `next_cursor`
- **THEN** a "Load more" button is displayed below the masonry gallery

#### Scenario: Load more button hidden on last page
- **WHEN** the image list response contains `next_cursor: null`
- **THEN** no "Load more" button is shown

#### Scenario: Clicking Load more fetches next page
- **WHEN** the user clicks the "Load more" button
- **THEN** the app calls `GET /images?folder_id=<current>&cursor=<next_cursor>`
- **AND** the newly fetched images are appended to the existing gallery

#### Scenario: Changing folder resets pagination
- **WHEN** the user navigates to a different folder
- **THEN** the accumulated pages are discarded
- **AND** only the first page of the new folder's images is shown

### Requirement: Gallery self-polls while any image has a pending thumbnail
The system SHALL set `refetchInterval` on the gallery's `useInfiniteQuery` to 1000ms while any loaded image has `thumbnail_url === null`. Polling SHALL stop automatically (interval returns `false`) once all loaded images have a non-null `thumbnail_url`. This covers all upload paths without any upload-specific wiring.

#### Scenario: Polling active while pending thumbnails exist
- **WHEN** the gallery's loaded image list contains at least one image with `thumbnail_url === null`
- **THEN** the gallery refetches `GET /images` every 1 second

#### Scenario: Polling stops once all thumbnails resolve
- **WHEN** all loaded images have a non-null `thumbnail_url`
- **THEN** the gallery stops polling and makes no further periodic refetch requests

### Requirement: Right-click context menu with delete option
The system SHALL show a context menu with a "Delete" option when the user right-clicks an image card.

#### Scenario: Right-click shows context menu
- **WHEN** the user right-clicks an image card
- **THEN** a context menu appears with a "Delete" option

### Requirement: Delete image moves it to trash
Selecting "Delete" from the context menu SHALL call `DELETE /images/:id` (no confirmation dialog). On success, the image SHALL be removed from the local gallery array without triggering a full gallery refetch. The right panel SHALL close if the deleted image was selected. A success toast reading "Image moved to trash" SHALL be shown on success.

#### Scenario: Delete moves the image to trash
- **WHEN** the user selects "Delete" from the image context menu
- **THEN** the app calls `DELETE /images/<id>`
- **AND** the deleted image is removed from the gallery without a full reload
- **AND** a success toast "Image moved to trash" is shown

#### Scenario: Right panel closes when the selected image is deleted
- **WHEN** the user deletes an image that is currently open in the right panel
- **THEN** the right panel closes on successful deletion

#### Scenario: Delete fails with an error toast
- **WHEN** the user selects "Delete" from the image context menu
- **AND** the `DELETE /images/<id>` request fails
- **THEN** the image remains in the gallery
- **AND** an error toast is shown
