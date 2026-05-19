## ADDED Requirements

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
