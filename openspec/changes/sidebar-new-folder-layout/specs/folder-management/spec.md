## MODIFIED Requirements

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
