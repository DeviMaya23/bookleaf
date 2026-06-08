## MODIFIED Requirements

### Requirement: FolderPanelContent component

The system SHALL provide a `FolderPanelContent` component at `frontend/src/components/FolderPanelContent.tsx`, rendered by `RightPanel` when a folder (rather than an image) is the active panel selection.

Props:
```ts
interface FolderPanelContentProps {
  folder: { id: string; name: string; description: string | null }
  onClose: () => void
}
```

Behaviour:
- Renders the folder's `name` as an editable title field and `description` as an editable notes field
- Both fields are pre-populated with the folder's current values and are independently auto-saved on blur via the consolidated `updateFolder(getToken, id, { name?, description? })` wrapper, sending only the field that changed — mirroring the image title/notes pattern in `RightPanel`
- A close button dismisses the panel via `onClose`

#### Scenario: Folder content renders with current values

- **WHEN** the right panel is showing folder content for a selected folder
- **THEN** the title field is pre-filled with the folder's `name`
- **AND** the description field is pre-filled with the folder's `description` (or empty if `null`)

---

### Requirement: Folder title is editable with auto-save on blur

The system SHALL display the folder `name` as an editable input. Changes SHALL be auto-saved via `updateFolder(getToken, id, { name })` when the field loses focus, but only if the value has changed and is non-empty. The title SHALL NOT be saved as an empty string; if the user clears the field and blurs, the input SHALL revert to the previous value without making an update call.

#### Scenario: Title is auto-saved on blur when changed

- **WHEN** the user edits the folder title input and removes focus
- **AND** the new value is non-empty and differs from the original
- **THEN** the app calls `updateFolder(getToken, id, { name: <new name> })`, sending only the `name` field
- **AND** a success toast is shown

#### Scenario: Unchanged title is not saved

- **WHEN** the user focuses and blurs the title input without changing its value
- **THEN** no `updateFolder` call is made

#### Scenario: Empty title reverts without saving

- **WHEN** the user clears the folder title input and removes focus
- **THEN** the input reverts to displaying the original title
- **AND** no `updateFolder` call is made

---

### Requirement: Folder description is editable with auto-save on blur

The system SHALL display the folder `description` as an editable notes field, pre-populated with the current value (empty if `null`). Changes SHALL be auto-saved via `updateFolder(getToken, id, { description })` when the field loses focus, but only if the value has changed. An empty description SHALL be saved as `null`, clearing it.

#### Scenario: Description is auto-saved on blur when changed

- **WHEN** the user edits the folder description and removes focus
- **AND** the new value differs from the original
- **THEN** the app calls `updateFolder(getToken, id, { description: <new text> })`, sending only the `description` field
- **AND** a success toast is shown

#### Scenario: Description can be cleared

- **WHEN** the user clears a previously non-empty description and removes focus
- **THEN** the app calls `updateFolder(getToken, id, { description: null })`, sending only the `description` field
- **AND** a success toast is shown

#### Scenario: Unchanged description is not saved

- **WHEN** the user focuses and blurs the description field without changing its value
- **THEN** no `updateFolder` call is made
