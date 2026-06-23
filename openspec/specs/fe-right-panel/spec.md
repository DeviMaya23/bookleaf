## Purpose

Detail panel for the selected image, rendered as a 320px sibling to the main content area in `AppLayout`. It shows the image's thumbnail, editable metadata (title, notes, source URL, folders, tags), a read-only details grid, and a download action — staying open and self-polling for thumbnail availability while an image is selected.

## Requirements

### Requirement: Right panel opens when an image card is clicked

The system SHALL render the right panel (`RightPanel` component) when an image is selected. On fine-pointer devices, the panel SHALL render as a 320px sidebar, a sibling to the main content area in `AppLayout`, opened by clicking an image card; this behavior is unchanged from before. On coarse-pointer devices (`useIsCoarsePointer()` is true), the panel SHALL render as a bottom drawer instead of a sidebar, and SHALL be opened via the "View details" item in the image card's context menu (per `fe-gallery-view`) rather than by tapping the card — tapping a card on a coarse-pointer device opens the lightbox instead, per `fe-image-lightbox`. The panel SHALL be hidden (not rendered in either shell) when no image is selected. The panel SHALL NOT be rendered while focus mode is active, even if an image is selected.

#### Scenario: Clicking an image card opens the right panel as a sidebar on a fine-pointer device

- **WHEN** a user on a fine-pointer device clicks an image card in the gallery
- **THEN** the right panel becomes visible as a sidebar on the right side of the layout
- **AND** the panel displays the selected image's metadata

#### Scenario: Selecting "View details" opens the right panel as a bottom drawer on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device selects "View details" from an image card's context menu
- **THEN** the right panel becomes visible as a bottom drawer
- **AND** the panel displays the selected image's metadata

#### Scenario: Tapping an image card does not open the right panel on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device taps an image card
- **THEN** the right panel is not opened
- **AND** the lightbox opens instead, per `fe-image-lightbox`

#### Scenario: Panel is hidden when no image is selected

- **WHEN** no image has been selected
- **THEN** the right panel is not rendered in either shell

#### Scenario: Panel stays hidden while focus mode is active

- **WHEN** focus mode is active
- **AND** an image becomes selected
- **THEN** the right panel is not rendered, even though an image is now selected

---

### Requirement: Right panel self-polls while the selected image has a pending thumbnail

The system SHALL poll `GET /images/:id` every 1000ms while the selected image's `thumbnail_url` is null. Polling SHALL stop automatically once `thumbnail_url` resolves to a non-null value. The panel SHALL display the resolved thumbnail as soon as it arrives, without requiring any user interaction.

#### Scenario: Panel polls and updates thumbnail when it resolves

- **WHEN** the right panel is open for an image with `thumbnail_url === null`
- **THEN** the panel polls `GET /images/:id` every 1000ms
- **AND** when the response contains a non-null `thumbnail_url`, the thumbnail is displayed and polling stops

#### Scenario: No polling when thumbnail is already present

- **WHEN** the right panel opens for an image that already has a non-null `thumbnail_url`
- **THEN** no periodic polling is performed

---

### Requirement: Right panel displays a thumbnail at the top

The system SHALL display the image's `thumbnail_url` at the top of the right panel. The thumbnail SHALL be rendered at full panel width with natural aspect ratio (not a fixed height). A close button (✕) SHALL be overlaid on the thumbnail (top-right corner). The thumbnail itself SHALL be a static display element with no click-to-open behavior.

#### Scenario: Thumbnail is shown at panel top

- **WHEN** the right panel is open for a selected image
- **THEN** the image thumbnail is displayed at the top of the panel at full panel width

#### Scenario: Clicking the thumbnail has no effect

- **WHEN** the user clicks the thumbnail in the right panel
- **THEN** no viewer or overlay opens
- **AND** the right panel remains as is

#### Scenario: Close button dismisses the panel

- **WHEN** the user clicks the ✕ close button overlaid on the thumbnail
- **THEN** the right panel closes and no image is selected

---

### Requirement: Right panel shows an editable title field

The system SHALL display the image `title` as an editable ghost input in the panel body, below the thumbnail. The field SHALL be pre-populated with the current title. Changes SHALL be auto-saved via `PATCH /images/:id` when the field loses focus, but only if the value has changed. The title SHALL NOT be saved as an empty string; if the user clears the field and blurs, the input SHALL revert to the previous value without making a PATCH call.

#### Scenario: Title is pre-populated and editable

- **WHEN** the right panel opens for a selected image
- **THEN** the title input is pre-filled with the image's current title
- **AND** the field is editable by the user

#### Scenario: Title is auto-saved on blur when changed

- **WHEN** the user edits the title input and then removes focus
- **AND** the new value is non-empty and differs from the original
- **THEN** the app calls `PATCH /images/<id>` with `{ "title": "<new value>" }`
- **AND** a success toast is shown

#### Scenario: Empty title reverts without saving

- **WHEN** the user clears the title input and removes focus
- **THEN** no PATCH request is made
- **AND** the title input reverts to the previous value

#### Scenario: Title is not saved on blur when unchanged

- **WHEN** the user focuses then blurs the title input without changing the value
- **THEN** no PATCH request is made

---

### Requirement: Right panel shows an editable notes field

The system SHALL display the image `description` as an editable textarea in the panel body, below the title. The field SHALL be pre-populated with the current description (or empty if null). Changes SHALL be auto-saved via `PATCH /images/:id` when the field loses focus, but only if the value has changed. A null description and an empty string SHALL be treated as equivalent for the purpose of change detection.

#### Scenario: Notes field is pre-populated and editable

- **WHEN** the right panel opens for a selected image with a description
- **THEN** the notes textarea is pre-filled with the current description

#### Scenario: Notes field is empty when description is null

- **WHEN** the right panel opens for a selected image with no description
- **THEN** the notes textarea is empty with a placeholder

#### Scenario: Notes are auto-saved on blur when changed

- **WHEN** the user edits the notes textarea and then removes focus
- **AND** the value differs from the original
- **THEN** the app calls `PATCH /images/<id>` with `{ "description": "<new value>" }`
- **AND** a success toast is shown

#### Scenario: Notes are not saved on blur when unchanged

- **WHEN** the user focuses then blurs the notes textarea without changing the value
- **THEN** no PATCH request is made

---

### Requirement: Right panel shows an editable source URL field with an Open button

The system SHALL display a source URL input field in the panel. The field SHALL be pre-populated with the image's `source_url`. Changes SHALL be auto-saved via `PATCH /images/:id` when the field loses focus, but only if the value has changed. An "Open ↗" button SHALL appear beside the field; it SHALL open the URL in a new browser tab when a URL is present, and be visually disabled when the field is empty.

#### Scenario: Source URL is pre-populated from image data

- **WHEN** the right panel opens for an image that has a `source_url`
- **THEN** the source URL input is pre-filled with that URL

#### Scenario: Source URL is auto-saved on blur when changed

- **WHEN** the user edits the source URL input and then removes focus from the field
- **AND** the value differs from the original
- **THEN** the app calls `PATCH /images/<id>` with `{ "source_url": "<new value>" }`
- **AND** a success toast is shown

#### Scenario: Source URL is not saved on blur when unchanged

- **WHEN** the user focuses then blurs the source URL input without changing the value
- **THEN** no PATCH request is made

#### Scenario: Open button opens URL in new tab

- **WHEN** the source URL field contains a non-empty URL
- **AND** the user clicks the Open ↗ button
- **THEN** the URL opens in a new browser tab

#### Scenario: Open button is disabled when source URL is empty

- **WHEN** the source URL field is empty
- **THEN** the Open ↗ button is visually disabled and clicking it has no effect

---

### Requirement: Right panel shows an editable Folders section

The system SHALL render a Folders section in `RightPanel` between the Source URL section and the Tags section. The section SHALL contain a `FolderInput` component pre-populated with the image's current folder assignments (resolved from `image.folder_ids` via the cached `['folders']` query).

Local folder state SHALL be a `{ id: string; name: string }[]` array initialised from `image.folder_ids` and reset whenever `image.id` changes.

When `FolderInput` calls `onChange`:
1. Update local folder state
2. Call `PATCH /images/:id` with `{ folder_ids: <updated UUID array> }`
3. On success — show a success toast and invalidate `['images']`
4. On error — show an error toast

#### Scenario: Folders section is present in the right panel

- **WHEN** the right panel opens for any image
- **THEN** a Folders section is visible between Source URL and Tags

#### Scenario: Folders section shows current image folders

- **WHEN** the right panel opens for an image that belongs to one or more folders
- **THEN** the FolderInput renders each folder as a pill

#### Scenario: Folders section is empty for an unfiled image

- **WHEN** the right panel opens for an image with no folder memberships
- **THEN** the FolderInput shows an empty input with placeholder text

#### Scenario: Folder state resets when a different image is selected

- **WHEN** the user selects a different image while the panel is already open
- **THEN** the FolderInput reflects only the new image's folder assignments

#### Scenario: Adding a folder patches the image

- **WHEN** the user selects a folder from the FolderInput dropdown
- **THEN** `PATCH /images/:id` is called with the updated folder UUID array
- **AND** a success toast is shown on success

#### Scenario: Removing a folder patches the image

- **WHEN** the user clicks ✕ on a folder pill in the FolderInput
- **THEN** `PATCH /images/:id` is called with that folder's UUID excluded from the array

#### Scenario: Removing the last folder sends an empty array

- **WHEN** the user removes the only remaining folder assignment
- **THEN** `PATCH /images/:id` is called with `{ "folder_ids": [] }`

#### Scenario: Failed PATCH shows error toast

- **WHEN** `PATCH /images/:id` fails during a folder change
- **THEN** an error toast is shown

---

### Requirement: Right panel shows a details grid with image metadata

The system SHALL display a 2-column details grid below the Folders section, containing: file size (formatted), dimensions (width × height), and upload date. The folder name row SHALL be removed from the details grid (folder assignments are now shown in the editable Folders section above).

#### Scenario: Details grid shows correct metadata

- **WHEN** the right panel is open for an image
- **THEN** the details grid displays size, dimensions, and added date
- **AND** no folder row is present in the details grid

---

### Requirement: Right panel has a sticky Download image button

The system SHALL render a "Download image" button in a sticky footer at the bottom of the right panel. Clicking it SHALL call `GET /images/:id/download`, receive the presigned `download_url`, and trigger a browser file download. The button SHALL show a loading/disabled state while the download URL is being fetched.

#### Scenario: Clicking Download image triggers a file download

- **WHEN** the user clicks the "Download image" button
- **THEN** the app calls `GET /images/<id>/download`
- **AND** upon receiving the `download_url`, the browser begins downloading the file
- **AND** the filename matches the image title

#### Scenario: Button is disabled while fetching the download URL

- **WHEN** the download URL request is in flight
- **THEN** the Download image button is disabled and shows a loading indicator

---

### Requirement: `selectedImage` state is owned by `AppLayout`

The system SHALL lift the selected image state from `ImageGrid` to `AppLayout`. `ImageGrid` SHALL receive an `onImageSelect` callback prop and call it when a card is clicked. `AppLayout` SHALL pass `selectedImage` and `onClose` to `RightPanel`.

#### Scenario: Selecting an image via card click is handled at layout level

- **WHEN** the user clicks an image card
- **THEN** `AppLayout` receives the selected image and passes it to `RightPanel`
- **AND** the panel becomes visible

#### Scenario: Closing the panel clears the selected image

- **WHEN** the user closes the right panel
- **THEN** `AppLayout` sets the selected image to null
- **AND** the panel is no longer rendered

---

### Requirement: Right panel shows a Tags section

The system SHALL render a Tags section in `RightPanel` between the Folders section and the Details section. The section SHALL contain a `TagInput` component pre-populated with the image's current tags.

#### Scenario: Tags section is present in the right panel

- **WHEN** the right panel opens for any image
- **THEN** a Tags section is visible between Folders and Details

#### Scenario: Tags section shows current image tags

- **WHEN** the right panel opens for an image that has associated tags
- **THEN** the TagInput renders each tag as a pill

#### Scenario: Tags section is empty for an untagged image

- **WHEN** the right panel opens for an image with no tags
- **THEN** the TagInput shows an empty input with placeholder text

---

### Requirement: Right panel opens or updates when a folder is selected

The system SHALL render the right panel showing folder content (via `FolderPanelContent`) when the user selects a folder in the sidebar that differs from the currently active folder, on a fine-pointer device — opening the sidebar shell. On a coarse-pointer device (`useIsCoarsePointer()` is true), selecting a different folder SHALL NOT open the panel; the panel is opened for that folder only via the "View details" item in the folder's context menu, per `folder-management`. If the panel happens to already be open (e.g. left open from a previous "View details" action) when the user selects a different folder, it SHALL update to show the newly selected folder's content rather than closing, on either pointer type. Selecting the currently active folder again SHALL be a no-op — the panel's existing content, whatever it is currently displaying, SHALL remain unchanged.

#### Scenario: Selecting a different folder opens or updates the panel with folder content on a fine-pointer device

- **WHEN** a user on a fine-pointer device selects a sidebar folder that is not the currently active folder
- **THEN** the right panel becomes visible (or updates, if already visible) in the sidebar shell
- **AND** the panel displays that folder's metadata via `FolderPanelContent`

#### Scenario: Selecting a different folder does not open the panel on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device selects a sidebar folder that is not the currently active folder
- **AND** the right panel is not currently open
- **THEN** the right panel remains closed
- **AND** the panel can be opened for that folder via "View details" in the folder's context menu, per `folder-management`

#### Scenario: An already-open panel updates to the newly selected folder on a coarse-pointer device

- **WHEN** the right panel is open on a coarse-pointer device, showing a previously selected folder's content
- **AND** the user selects a different folder in the sidebar
- **THEN** the right panel updates to show the newly selected folder's metadata
- **AND** the panel does not close

#### Scenario: Re-selecting the active folder leaves the panel untouched

- **WHEN** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel's current content remains unchanged
- **AND** no new panel state is set

#### Scenario: Re-selecting the active folder while image content is shown leaves it untouched

- **WHEN** the right panel is currently showing image content
- **AND** the authenticated user selects the sidebar folder that is already active
- **THEN** the right panel continues showing the same image content
- **AND** the panel does not switch to folder content

---

### Requirement: Right panel renders as a bottom drawer on coarse-pointer devices

On coarse-pointer devices, the system SHALL render the right panel's content (`ImagePanelBody` or `FolderPanelContent`, unchanged from the fine-pointer sidebar) inside a fixed bottom-anchored drawer with a backdrop, instead of the 320px sidebar. The drawer SHALL be binary (open/close only) — no swipe-to-dismiss or snap-point gestures. The drawer SHALL close when the user activates its close control or taps the backdrop, calling the same `onClose` callback used by the sidebar shell.

#### Scenario: Right panel renders as a bottom drawer on a coarse-pointer device

- **WHEN** the right panel is open on a coarse-pointer device
- **THEN** the panel renders as a fixed bottom-anchored drawer with a backdrop
- **AND** the panel does not render as a 320px sidebar

#### Scenario: Tapping the backdrop closes the drawer

- **WHEN** the bottom drawer is open
- **AND** the user taps the backdrop
- **THEN** the drawer closes via the same `onClose` callback the sidebar shell uses

#### Scenario: Drawer content is identical to sidebar content

- **WHEN** the right panel is open for a given image or folder
- **THEN** the content rendered inside the bottom drawer (on coarse-pointer devices) is the same `ImagePanelBody`/`FolderPanelContent` component used inside the sidebar (on fine-pointer devices)

---

### Requirement: Selecting an image replaces folder content in the right panel

The system SHALL ensure that selecting an image card always results in the right panel showing that image's content, replacing any folder content that was previously displayed. Image and folder content SHALL never be displayed in the panel simultaneously.

#### Scenario: Selecting an image while folder content is shown switches the panel to image content

- **WHEN** the right panel is currently showing folder content
- **AND** the authenticated user clicks an image card in the gallery
- **THEN** the right panel switches to displaying that image's metadata
- **AND** the folder content is no longer shown
