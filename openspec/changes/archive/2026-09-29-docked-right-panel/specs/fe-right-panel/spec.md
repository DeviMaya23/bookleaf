## MODIFIED Requirements

### Requirement: Right panel opens when an image card is clicked

On fine-pointer devices, the system SHALL switch the right panel's content to image mode when an image card is clicked. The panel itself SHALL always be present (per `fe-right-panel-docked`) — clicking an image changes what is shown, not whether the panel is mounted. On coarse-pointer devices, the behavior is unchanged: the bottom drawer opens via "View details" in the image card's context menu, and tapping a card opens the lightbox instead.

The panel SHALL NOT render while focus mode is active, per `focus-mode`.

#### Scenario: Clicking an image card switches the panel to image content on a fine-pointer device

- **WHEN** a user on a fine-pointer device clicks an image card in the gallery
- **THEN** the right panel displays that image's metadata
- **AND** the panel layout does not reflow

#### Scenario: Selecting "View details" opens the right panel as a bottom drawer on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device selects "View details" from an image card's context menu
- **THEN** the right panel becomes visible as a bottom drawer
- **AND** the panel displays the selected image's metadata

#### Scenario: Tapping an image card does not open the right panel on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device taps an image card
- **THEN** the right panel is not opened
- **AND** the lightbox opens instead, per `fe-image-lightbox`

#### Scenario: Panel stays hidden while focus mode is active

- **WHEN** focus mode is active
- **AND** the user clicks an image card
- **THEN** the right panel is not rendered, even though an image is now selected

---

### Requirement: Right panel displays a thumbnail at the top

The system SHALL display the image's `thumbnail_url` at the top of the right panel. The thumbnail SHALL be rendered at full panel width with natural aspect ratio. The thumbnail itself SHALL be a static display element with no click-to-open behavior. There is no close button overlaid on the thumbnail.

#### Scenario: Thumbnail is shown at panel top

- **WHEN** the right panel is open for a selected image
- **THEN** the image thumbnail is displayed at the top of the panel at full panel width

#### Scenario: Clicking the thumbnail has no effect

- **WHEN** the user clicks the thumbnail in the right panel
- **THEN** no viewer or overlay opens
- **AND** the right panel remains as is

---

### Requirement: Right panel opens or updates when a folder is selected

The system SHALL update the right panel's content to show folder details (via `FolderPanelContent`) when the user selects a folder in the sidebar, on a fine-pointer device. On a coarse-pointer device, selecting a different folder SHALL NOT open the panel; the panel is opened for that folder only via the "View details" item in the folder's context menu, per `folder-management`. If the panel is already open when the user selects a different folder, it SHALL update to show the newly selected folder's content, on either pointer type. Selecting the currently active folder again SHALL be a no-op.

#### Scenario: Selecting a different folder updates the panel with folder content on a fine-pointer device

- **WHEN** a user on a fine-pointer device selects a sidebar folder that is not the currently active folder
- **THEN** the right panel updates to display that folder's metadata via `FolderPanelContent`

#### Scenario: Selecting a different folder does not open the panel on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device selects a sidebar folder that is not the currently active folder
- **AND** the right panel is not currently open
- **THEN** the right panel remains closed

#### Scenario: An already-open panel updates to the newly selected folder on a coarse-pointer device

- **WHEN** the right panel is open on a coarse-pointer device showing a previously selected folder's content
- **AND** the user selects a different folder in the sidebar
- **THEN** the right panel updates to show the newly selected folder's metadata
- **AND** the panel does not close

#### Scenario: Re-selecting the active folder leaves the panel untouched

- **WHEN** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel's current content remains unchanged

#### Scenario: Re-selecting the active folder while image content is shown leaves it untouched

- **WHEN** the right panel is currently showing image content
- **AND** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel continues showing the same image content

---

### Requirement: Right panel shows a selection-actions mode when images are selected

The system SHALL render the right panel in a `selection` mode whenever `selectedIds` is non-empty, regardless of whether focus mode is active. This mode SHALL take priority over the `image`, `folder`, and `neutral` panel modes — while `selectedIds` is non-empty, the panel SHALL show selection content even if a `selectedImage` or active folder would otherwise apply.

The selection panel SHALL display:
- The current count of selected images.
- An "Add to folder" action that opens a single-select folder picker; choosing a folder immediately calls `POST /images/bulk/add-to-folder` with the current `selectedIds` and the chosen folder.
- A "Move to trash" action that immediately calls `POST /images/bulk/trash` with the current `selectedIds`, with no confirmation step.

The selection panel has no close control. The toolbar's Select mode toggle is the sole affordance for exiting select mode.

Unlike the `image`/`folder`/`neutral` modes, the selection panel SHALL remain visible while focus mode is active.

#### Scenario: Selection panel appears once at least one image is selected

- **WHEN** the user is in select mode and selects one image
- **THEN** the right panel renders in `selection` mode showing a count of 1

#### Scenario: Selection panel takes priority over the image panel

- **WHEN** `selectedIds` is non-empty
- **THEN** the right panel shows `selection` mode content, not `image` mode content, even if a `selectedImage` value is set

#### Scenario: Selection panel stays visible during focus mode

- **WHEN** focus mode is active
- **AND** at least one image is selected
- **THEN** the right panel remains visible in `selection` mode

#### Scenario: Choosing a folder from the Add to folder picker triggers the bulk request

- **WHEN** the user has 3 images selected and picks a folder from the "Add to folder" picker
- **THEN** the app calls `POST /images/bulk/add-to-folder` with `image_ids` containing the 3 selected IDs and the chosen `folder_id`

#### Scenario: Move to trash triggers the bulk request immediately

- **WHEN** the user has 3 images selected and clicks "Move to trash"
- **THEN** the app calls `POST /images/bulk/trash` with `image_ids` containing the 3 selected IDs, without any confirmation dialog

#### Scenario: A successful bulk action exits select mode entirely

- **WHEN** a bulk add-to-folder or bulk trash request completes successfully
- **THEN** `selectedIds` and the anchor are cleared, `selectMode` is turned off, and the selection panel is no longer shown

---

## REMOVED Requirements

### Requirement: Closing the selection panel exits select mode entirely

**Reason**: The selection panel no longer has a close control. The toolbar Select mode toggle is the sole affordance for exiting select mode, consistent with how all other mode exits work.

**Migration**: Users exit select mode via the Select mode toggle in the gallery toolbar.
