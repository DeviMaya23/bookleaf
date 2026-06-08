## MODIFIED Requirements

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
