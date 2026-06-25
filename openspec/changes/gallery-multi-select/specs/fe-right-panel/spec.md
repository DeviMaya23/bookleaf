## ADDED Requirements

### Requirement: Right panel shows a selection-actions mode when images are selected

The system SHALL render the right panel in a `selection` mode whenever `selectedIds` is non-empty, regardless of whether focus mode is active. This mode SHALL take priority over the `image` and `folder` panel modes — while `selectedIds` is non-empty, the panel SHALL show selection content even if a `selectedImage` or active folder would otherwise apply. The panel SHALL be hidden when `selectedIds` is empty, the same as the existing "hidden when no image is selected" rule for the `image` mode.

The selection panel SHALL display:
- The current count of selected images.
- An "Add to folder" action that opens a single-select folder picker; choosing a folder immediately calls `POST /images/bulk/add-to-folder` with the current `selectedIds` and the chosen folder.
- A "Move to trash" action that immediately calls `POST /images/bulk/trash` with the current `selectedIds`, with no confirmation step.

Unlike the `image`/`folder` modes, the selection panel SHALL remain visible while focus mode is active; the user dismisses it manually (e.g. via a close control) rather than it being hidden automatically.

#### Scenario: Selection panel appears once at least one image is selected

- **WHEN** the user is in select mode and selects one image
- **THEN** the right panel renders in `selection` mode showing a count of 1

#### Scenario: Selection panel is hidden when selection is empty

- **WHEN** select mode is active but no images are currently selected
- **THEN** the right panel is not rendered

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
- **THEN** `selectedIds` and the anchor are cleared, `selectMode` is turned off, and the selection panel is hidden
- **AND** the right panel does not fall back to showing a previously-selected image or folder as a result

### Requirement: Entering select mode clears any open image or folder panel

The system SHALL clear `selectedImage` and `folderPanelOpen` (the same way `handleImageSelect`/`handleFolderViewDetails` already clear each other) at the moment select mode is turned on. This prevents a previously-open `image`/`folder` panel from resurfacing once the selection later empties, by whatever means (closing the selection panel, or a successful bulk action).

#### Scenario: Entering select mode while the image panel is open closes it

- **WHEN** the `image` mode right panel is open for a selected image
- **AND** the user turns select mode on
- **THEN** the right panel closes (no `image` mode content is shown)
- **AND** it does not reopen once the user later empties their multi-select

#### Scenario: Entering select mode while the folder panel is open closes it

- **WHEN** the `folder` mode right panel is open
- **AND** the user turns select mode on
- **THEN** the right panel closes (no `folder` mode content is shown)
- **AND** it does not reopen once the user later empties their multi-select

### Requirement: Closing the selection panel exits select mode entirely

The system SHALL, when the user dismisses the selection panel via its close control, perform the same full reset as a successful bulk action or the toolbar toggle-off: clear `selectedIds` and the anchor, and turn `selectMode` off. The user must re-enter select mode via the toolbar toggle to select again.

#### Scenario: Closing the selection panel turns select mode off

- **WHEN** select mode is active with one or more images selected
- **AND** the user clicks the close control on the selection panel
- **THEN** `selectedIds` and the anchor are cleared, and `selectMode` is turned off
- **AND** the select-mode toolbar toggle reflects the off state
