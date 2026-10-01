## MODIFIED Requirements

### Requirement: Chip-style multi-folder picker with deferred Apply

The "Add to folder" section of the selection panel SHALL use `FolderInput` (the `TokenInput`-based chip component). The `FolderInput` SHALL render with all available folders as suggestions and no pre-selected folders.

A single "Apply" button SHALL appear below both the folder picker and the tag picker sections. The Apply button SHALL be disabled when both `pendingFolders` and `pendingTags` are empty. It SHALL be enabled when at least one folder chip or one tag chip is present.

Clicking Apply SHALL:
1. Call `onApply({ folderIds, tags })` with the pending folder IDs and pending `Tag` objects.
2. Clear both chip inputs.

`AppLayout`'s `onApply` handler calls `resolveOrCreateTags` on the received `tags` before firing the bulk tag request, keeping tag resolution out of the panel component.

`onApply` existing behavior (`bulkAddImagesToFolder` per folder, `bulkAddTags` for all tags) is defined at the `AppLayout` level.

`SelectionFolderPicker` SHALL be deleted.

#### Scenario: User adds two folders and clicks Apply

- **WHEN** the user adds two folder chips and clicks Apply
- **THEN** `onApply` is called with both folder IDs in `folderIds` and an empty `tags`

#### Scenario: User adds tags only and clicks Apply

- **WHEN** the user adds two tag chips and clicks Apply with no folder chips
- **THEN** `onApply` is called with an empty `folderIds` and both `Tag` objects in `tags`

#### Scenario: User adds both folders and tags and clicks Apply

- **WHEN** the user adds one folder chip and two tag chips, then clicks Apply
- **THEN** `onApply` is called with the folder ID in `folderIds` and both `Tag` objects in `tags`

#### Scenario: Apply button is disabled with no chips in either input

- **WHEN** neither folder chips nor tag chips have been added
- **THEN** the Apply button is disabled and non-interactive

#### Scenario: Apply button is enabled when only folder chips are present

- **WHEN** at least one folder chip is present and no tag chips are present
- **THEN** the Apply button is enabled

#### Scenario: Apply button is enabled when only tag chips are present

- **WHEN** at least one tag chip is present and no folder chips are present
- **THEN** the Apply button is enabled

#### Scenario: Removing all chips before Apply keeps button disabled

- **WHEN** the user adds chips and then removes all of them before clicking Apply
- **THEN** the Apply button is disabled

#### Scenario: Folder chip input shows available folders as suggestions

- **WHEN** the user focuses the folder chip input and begins typing a folder name
- **THEN** matching folders appear in the suggestion dropdown

#### Scenario: Tag chip input shows existing tags as suggestions

- **WHEN** the user focuses the tag chip input and begins typing a tag name
- **THEN** matching tags appear in the suggestion dropdown

#### Scenario: Free-text tag entry creates a new tag on Apply

- **WHEN** the user types a new tag name not in their existing tags and presses comma or blur to add it as a chip
- **AND** clicks Apply
- **THEN** the tag is created server-side via `resolveOrCreateTags` before the bulk tag request fires

## ADDED Requirements

### Requirement: Tag section in selection panel

The selection panel SHALL render an "Add tags" section below the "Add to folder" section. It SHALL use the existing `TagInput` component (chips + searchable text box with free-text creation). The `TagInput` SHALL render with all of the user's existing tags as suggestions and no pre-selected tags.

`SelectionPanelBody` SHALL accept `onApply: (params: { folderIds: string[]; tags: Tag[] }) => void` as a required prop, replacing the former `onAddToFolder: (folderId: string) => void` prop. It passes full `Tag` objects (not just IDs) so that `AppLayout` can call `resolveOrCreateTags` on them before firing the bulk tag request.

`RightPanel` selection mode content SHALL forward `panelContent.onApply` to `SelectionPanelBody`.

#### Scenario: Tag section is visible when images are selected

- **WHEN** the selection panel is visible with one or more images selected
- **THEN** an "Add tags" section is rendered below "Add to folder"

#### Scenario: RightPanel forwards onApply

- **WHEN** `RightPanel` renders selection mode content
- **THEN** `SelectionPanelBody` receives the `onApply` prop from `panelContent.onApply`

### Requirement: onApply handler in AppLayout

`AppLayout` SHALL provide a unified `onApply({ folderIds, tags })` handler to the selection panel.

For each folder ID in `folderIds`, the handler SHALL fire `POST /images/bulk/add-to-folder` (one request per folder, preserving existing behaviour). For `tags`, if non-empty, the handler SHALL first call `resolveOrCreateTags` to materialise any free-text entries into tag IDs, then fire `POST /images/bulk/tag` once with the resolved IDs. Both sets of requests SHALL be initiated concurrently via `Promise.allSettled`. On success of any sub-operation, the image list SHALL be invalidated. On complete success, select mode SHALL be exited and a success toast SHALL be shown. On partial or full failure, an error toast SHALL be shown per failed operation without exiting select mode.

#### Scenario: Both folder and tag requests succeed

- **WHEN** `onApply` is called with one folder ID and two tag objects and both requests succeed
- **THEN** the image list is invalidated, select mode is exited, and a success toast is shown

#### Scenario: Tag request fails, folder request succeeds

- **WHEN** `onApply` is called with one folder ID and two tag objects, the folder request succeeds, but the tag request fails
- **THEN** an error toast is shown for the tag failure
- **AND** select mode is not exited
