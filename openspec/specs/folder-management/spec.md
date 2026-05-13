### Requirement: Create folder via Dialog
The system SHALL allow the user to create a new folder by clicking the "+ New folder" button, which opens a Dialog containing a text input for the folder name and a confirm button. On confirm, the system SHALL call `POST /folders` and refetch the folder list.

#### Scenario: User creates a folder successfully
- **WHEN** the user clicks "+ New folder", enters a name, and confirms
- **THEN** `POST /folders` is called with the entered name
- **AND** the Dialog closes
- **AND** the folder list is refetched and the new folder appears in the sidebar

#### Scenario: User cancels folder creation
- **WHEN** the user opens the new folder Dialog and dismisses it without confirming
- **THEN** no API call is made and the folder list is unchanged

### Requirement: Rename folder via ContextMenu and Dialog
The system SHALL allow the user to rename a folder by right-clicking it to open a ContextMenu, selecting "Rename", and entering a new name in a Dialog pre-populated with the current folder name. On confirm, the system SHALL call `PUT /folders/:id` and refetch the folder list.

#### Scenario: User renames a folder successfully
- **WHEN** the user right-clicks a folder, selects "Rename", changes the name, and confirms
- **THEN** `PUT /folders/:id` is called with the new name
- **AND** the Dialog closes
- **AND** the folder list is refetched and the updated name is shown in the sidebar

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
