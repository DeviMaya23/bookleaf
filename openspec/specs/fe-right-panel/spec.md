### Requirement: Right panel opens when an image card is clicked

The system SHALL render a 320px right panel (`RightPanel` component) as a sibling to the main content area in `AppLayout`. The panel SHALL be hidden when no image is selected. When the user clicks an image card, the panel SHALL become visible and display that image's metadata.

#### Scenario: Clicking an image card opens the right panel

- **WHEN** the authenticated user clicks an image card in the gallery
- **THEN** the right panel becomes visible on the right side of the layout
- **AND** the panel displays the selected image's metadata

#### Scenario: Panel is hidden when no image is selected

- **WHEN** no image has been selected
- **THEN** the right panel is not rendered in the layout

---

### Requirement: Right panel displays a thumbnail at the top

The system SHALL display the image's `thumbnail_url` at the top of the right panel. The thumbnail SHALL be rendered at full panel width with natural aspect ratio (not a fixed height). A close button (✕) SHALL be overlaid on the thumbnail (top-right corner). Clicking the thumbnail SHALL open the lightbox.

#### Scenario: Thumbnail is shown at panel top

- **WHEN** the right panel is open for a selected image
- **THEN** the image thumbnail is displayed at the top of the panel at full panel width

#### Scenario: Clicking the thumbnail opens the lightbox

- **WHEN** the user clicks the thumbnail in the right panel
- **THEN** the lightbox opens and displays the full-resolution image

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

### Requirement: Right panel shows a details grid with image metadata

The system SHALL display a 2-column details grid below the source URL section, containing: file size (formatted), dimensions (width × height), folder name (or "Unsorted" if none), and upload date.

#### Scenario: Details grid shows correct metadata

- **WHEN** the right panel is open for an image
- **THEN** the details grid displays size, dimensions, folder, and added date

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
