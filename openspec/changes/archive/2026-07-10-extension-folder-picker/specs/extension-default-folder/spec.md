## MODIFIED Requirements

### Requirement: Folder list fetch from API

`api.ts` SHALL expose a `getFolders()` function that calls `GET /folders` via `apiFetch` and returns a `Promise<Array<{ id: string; name: string; parent_id: string | null }>>`. On a non-OK response it SHALL throw an error.

#### Scenario: Successful fetch returns folder list with parent_id

- **WHEN** `GET /folders` responds with a 200 and a JSON array of folder objects
- **THEN** `getFolders()` resolves with an array of `{ id, name, parent_id }` objects

#### Scenario: Non-OK response throws

- **WHEN** `GET /folders` responds with a non-2xx status
- **THEN** `getFolders()` rejects with an error

### Requirement: Save destination section in Settings UI

The Settings panel (`Settings.tsx`) SHALL render a **"Save destination"** section between the drag-to-save row and the Hotkeys section header. The section SHALL contain:

- A section label styled identically to the existing "Hotkeys" label (10px, uppercase, `c.textSec`).
- A row with a text label "Default folder" and a button flush-right that displays the current default folder name, or `"None"` when no default is stored.
- Clicking the button SHALL navigate to the FolderPicker panel view.
- The displayed name SHALL be read from `getDefaultFolder()` on Settings mount and updated when the user returns from the FolderPicker panel.

#### Scenario: Settings shows stored folder name on the button

- **WHEN** Settings mounts and a default folder `{ id: "xyz", name: "Mood" }` is stored
- **THEN** the "Default folder" button reads "Mood"

#### Scenario: Settings shows "None" when no default is stored

- **WHEN** Settings mounts and no default folder is stored
- **THEN** the "Default folder" button reads "None"

#### Scenario: Clicking the button navigates to FolderPicker

- **WHEN** the user clicks the "Default folder" button in Settings
- **THEN** the popup navigates to the FolderPicker panel view

#### Scenario: Returning from FolderPicker reflects the updated selection

- **WHEN** the user selects a folder in FolderPicker and the popup navigates back to Settings
- **THEN** the "Default folder" button displays the name of the newly selected folder
