## ADDED Requirements

### Requirement: Root route displays unfoldered images
The system SHALL fetch images with `folder_id=null` when the user is on the root path (`/`) and render them in the gallery grid. The response envelope `{ images, next_cursor }` SHALL be handled for pagination.

#### Scenario: Navigating to root loads unfoldered images
- **WHEN** the authenticated user navigates to `/`
- **THEN** the app calls `GET /images?folder_id=null`
- **AND** the returned images are displayed in the gallery grid

### Requirement: Folder route displays folder images
The system SHALL fetch images for a specific folder when the user is on `/folders/:folder_id` and render them in the gallery grid. The response envelope `{ images, next_cursor }` SHALL be handled for pagination.

#### Scenario: Navigating to a folder route loads that folder's images
- **WHEN** the authenticated user navigates to `/folders/:folder_id`
- **THEN** the app calls `GET /images?folder_id=<folder_id>`
- **AND** the returned images for that folder are displayed in the gallery grid

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
- **AND** the gallery grid is not rendered

### Requirement: Images displayed as thumbnail cards in a responsive grid
The system SHALL render each image as a card showing its thumbnail and title. Long titles SHALL be truncated with an ellipsis. The grid SHALL use 6 columns on desktop and 2 columns on mobile.

#### Scenario: Image card displays thumbnail and title
- **WHEN** the image list contains images
- **THEN** each image is rendered as a card with its thumbnail and title visible

#### Scenario: Long image titles are truncated
- **WHEN** an image has a title longer than the card width
- **THEN** the title is truncated with a trailing ellipsis

#### Scenario: Grid is responsive
- **WHEN** the viewport is desktop width (≥ 1024px)
- **THEN** images are displayed in a 6-column grid

#### Scenario: Grid collapses on mobile
- **WHEN** the viewport is mobile width (< 768px)
- **THEN** images are displayed in a 2-column grid

### Requirement: Paginated image loading with "Load more"
The system SHALL use `useInfiniteQuery` to fetch images in pages. A "Load more" button SHALL be shown when a `next_cursor` is present in the last page's response. Clicking it SHALL fetch the next page by passing `cursor=<next_cursor>` to `GET /images`. When switching folders (or navigating to root), the accumulated pages SHALL be reset.

#### Scenario: Load more button shown when next page exists
- **WHEN** the image list response contains a non-null `next_cursor`
- **THEN** a "Load more" button is displayed below the image grid

#### Scenario: Load more button hidden on last page
- **WHEN** the image list response contains `next_cursor: null`
- **THEN** no "Load more" button is shown

#### Scenario: Clicking Load more fetches next page
- **WHEN** the user clicks the "Load more" button
- **THEN** the app calls `GET /images?folder_id=<current>&cursor=<next_cursor>`
- **AND** the newly fetched images are appended to the existing grid

#### Scenario: Changing folder resets pagination
- **WHEN** the user navigates to a different folder
- **THEN** the accumulated pages are discarded
- **AND** only the first page of the new folder's images is shown

### Requirement: Right-click context menu with delete option
The system SHALL show a context menu with a "Delete" option when the user right-clicks an image card.

#### Scenario: Right-click shows context menu
- **WHEN** the user right-clicks an image card
- **THEN** a context menu appears with a "Delete" option

### Requirement: Delete image with confirmation dialog
The system SHALL show a confirmation dialog before deleting an image. Upon confirmation, the system SHALL call `DELETE /images/:id` and refresh the image list.

#### Scenario: Confirming delete removes the image
- **WHEN** the user selects "Delete" from the image context menu
- **AND** a confirmation dialog appears
- **AND** the user confirms the deletion
- **THEN** the app calls `DELETE /images/<id>`
- **AND** the image list is refreshed and the deleted image no longer appears

#### Scenario: Cancelling delete keeps the image
- **WHEN** the user selects "Delete" from the image context menu
- **AND** a confirmation dialog appears
- **AND** the user cancels
- **THEN** no delete request is made
- **AND** the image remains in the gallery
