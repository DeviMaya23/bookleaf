### Requirement: Trash view lists soft-deleted images
Navigating to `/trash` SHALL display the list of soft-deleted images fetched from `GET /images/trash`. The view SHALL use the same masonry grid layout as the regular image grid. Images SHALL be listed oldest-deleted-first (ascending `deleted_at`), as returned by the API.

#### Scenario: Trash view shows deleted images
- **WHEN** the user navigates to `/trash` and `GET /images/trash` returns images
- **THEN** the images are rendered in a masonry grid in oldest-deleted-first order

#### Scenario: Empty trash state
- **WHEN** the user navigates to `/trash` and `GET /images/trash` returns an empty list
- **THEN** an empty state message is shown

---

### Requirement: Restore image from trash via context menu
Each image in the trash view SHALL have a context menu with a **Restore** item (instead of Delete). Selecting Restore SHALL call `POST /images/:id/restore`. On success, the image SHALL be removed from the trash view and a success toast SHALL be shown.

#### Scenario: User restores an image successfully
- **WHEN** the user right-clicks an image in the trash view and selects "Restore"
- **THEN** `POST /images/:id/restore` is called
- **AND** the image disappears from the trash view
- **AND** a success toast is shown

#### Scenario: Restore fails with an error toast
- **WHEN** `POST /images/:id/restore` returns an error
- **THEN** the image remains in the trash view
- **AND** an error toast is shown

---

### Requirement: Trash view pagination
The trash view SHALL support cursor-based infinite scroll, loading additional pages of deleted images as the user scrolls, using the cursor returned by `GET /images/trash`.

#### Scenario: Additional pages load on scroll
- **WHEN** the user scrolls to the bottom of the trash view and more images exist
- **THEN** the next page is fetched and appended to the grid
