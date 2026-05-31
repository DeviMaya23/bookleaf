## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Right panel shows a details grid with image metadata

The system SHALL display a 2-column details grid below the Folders section, containing: file size (formatted), dimensions (width × height), and upload date. The folder name row SHALL be removed from the details grid (folder assignments are now shown in the editable Folders section above).

#### Scenario: Details grid shows correct metadata

- **WHEN** the right panel is open for an image
- **THEN** the details grid displays size, dimensions, and added date
- **AND** no folder row is present in the details grid
