### Requirement: Create folder via Dialog
The system SHALL allow the user to create a new folder by clicking either the icon button beside the "FOLDERS" section label or, when the folder list is empty, the "+ New folder" footer button — both of which open the same Dialog containing a text input for the folder name and a confirm button. On confirm, the system SHALL call `POST /folders` and refetch the folder list.

#### Scenario: User creates a folder successfully via the section label icon button
- **WHEN** the user clicks the icon button beside "FOLDERS", enters a name, and confirms
- **THEN** `POST /folders` is called with the entered name
- **AND** the Dialog closes
- **AND** the folder list is refetched and the new folder appears in the sidebar

#### Scenario: User creates a folder successfully via the footer button on an empty account
- **WHEN** the folder list is empty and the user clicks "+ New folder" in the footer, enters a name, and confirms
- **THEN** `POST /folders` is called with the entered name
- **AND** the Dialog closes
- **AND** the folder list is refetched and the new folder appears in the sidebar

#### Scenario: User cancels folder creation
- **WHEN** the user opens the new folder Dialog and dismisses it without confirming
- **THEN** no API call is made and the folder list is unchanged

### Requirement: Rename folder via ContextMenu and Dialog
The system SHALL allow the user to rename a folder by right-clicking it to open a ContextMenu, selecting "Rename", and entering a new name in a Dialog pre-populated with the current folder name. On confirm, the system SHALL call `updateFolder(getToken, id, { name })`, sending only the `name` field, and refetch the folder list.

#### Scenario: User renames a folder successfully
- **WHEN** the user right-clicks a folder, selects "Rename", changes the name, and confirms
- **THEN** `updateFolder` is called with `{ "name": "<new name>" }`, with no `parent_id` or `description` field in the body
- **AND** the Dialog closes
- **AND** the folder list is refetched and the updated name is shown in the sidebar
- **AND** the folder's `parent_id` and `description` are unchanged

#### Scenario: User cancels rename
- **WHEN** the user opens the rename Dialog and dismisses it without confirming
- **THEN** no API call is made and the folder name is unchanged

### Requirement: Delete folder via ContextMenu and confirmation Dialog
The system SHALL allow the user to delete a folder by right-clicking it to open a ContextMenu, selecting "Delete", and confirming in a confirmation Dialog. The folder SHALL only be deleted if the user explicitly confirms. On confirm, the system SHALL call `DELETE /folders/:id` and refetch the folder list.

#### Scenario: User deletes a folder after confirming
- **WHEN** the user right-clicks a folder, selects "Delete", and confirms in the confirmation Dialog
- **THEN** `DELETE /folders/:id` is called
- **AND** the Dialog closes
- **AND** the folder list is refetched and the deleted folder no longer appears in the sidebar

#### Scenario: User cancels deletion
- **WHEN** the user opens the delete confirmation Dialog and dismisses it without confirming
- **THEN** no API call is made and the folder list is unchanged

### Requirement: Create subfolder via context menu
The system SHALL allow the user to create a subfolder by right-clicking an existing folder and selecting "New subfolder". This SHALL open the folder name dialog. On confirm, the system SHALL call `POST /folders` with `name` and `parent_id` set to the right-clicked folder's id, then refetch the folder list.

#### Scenario: User creates a subfolder successfully
- **WHEN** the user right-clicks a folder, selects "New subfolder", enters a name, and confirms
- **THEN** `POST /folders` is called with the entered name and the parent folder's id as `parent_id`
- **AND** the dialog closes
- **AND** the folder list is refetched and the new child folder appears under the parent

#### Scenario: User cancels subfolder creation
- **WHEN** the user opens the new subfolder dialog and dismisses it without confirming
- **THEN** no API call is made and the folder list is unchanged

### Requirement: Folder list refetch after mutation
The system SHALL invalidate the folder list query after any successful create, rename, or delete mutation, causing the sidebar to refetch and display up-to-date data.

#### Scenario: Folder list updates after create
- **WHEN** a new folder is successfully created
- **THEN** the folder list query is invalidated and the sidebar re-renders with the new folder

#### Scenario: Folder list updates after rename
- **WHEN** a folder is successfully renamed
- **THEN** the folder list query is invalidated and the sidebar re-renders with the updated name

#### Scenario: Folder list updates after delete
- **WHEN** a folder is successfully deleted
- **THEN** the folder list query is invalidated and the sidebar re-renders without the deleted folder

### Requirement: View folder details via ContextMenu on coarse-pointer devices

The system SHALL show a "View details" item at the top of the folder `ContextMenu` (the same context menu used for "New subfolder", "Rename", "Change icon", and "Delete"), above "New subfolder", when the device is a coarse-pointer device (`useIsCoarsePointer()` is true). Selecting "View details" SHALL navigate to that folder (if it is not already the active folder, the same navigation a tap/click on the folder performs) and open the right panel for that folder as a bottom drawer, per `fe-right-panel`. This mirrors the fine-pointer behavior, where a single click both navigates to the folder and opens its panel — the right panel SHALL NEVER display metadata for a folder other than the currently active one, on either pointer type. The "View details" item SHALL NOT be rendered on fine-pointer devices, since selecting the folder already opens the right panel there.

#### Scenario: View details item shown on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device long-presses a folder in the sidebar to open its context menu
- **THEN** the context menu shows "View details" above "New subfolder"

#### Scenario: View details item not shown on a fine-pointer device

- **WHEN** a user on a fine-pointer device right-clicks a folder in the sidebar to open its context menu
- **THEN** the context menu does not show a "View details" item

#### Scenario: Selecting View details on a non-active folder navigates to it and opens the right panel

- **WHEN** a user on a coarse-pointer device selects "View details" from the context menu of a folder that is not the currently active folder
- **THEN** the system navigates to that folder
- **AND** the right panel opens as a bottom drawer showing that folder's metadata, per `fe-right-panel`

#### Scenario: Selecting View details on the active folder opens the right panel without navigating

- **WHEN** a user on a coarse-pointer device selects "View details" from the context menu of the currently active folder
- **THEN** no navigation occurs
- **AND** the right panel opens as a bottom drawer showing that folder's metadata, per `fe-right-panel`
