## Purpose

Folder content for the right panel, rendered via the `FolderPanelContent` component when a folder (rather than an image) is the active panel selection. It shows the folder's editable metadata (title, description) with auto-save on blur.

## Requirements

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

---

### Requirement: Folder detail query for image count

`FolderPanelContent` SHALL fetch the folder's detail (including `image_count`) via a new `getFolder(getToken, id)` helper in `frontend/src/lib/folders.ts`, calling `GET /folders/:id`. The query SHALL be keyed on the folder's `id` (e.g. `['folder', folder.id]`) so it refetches when the active folder changes.

```ts
export interface FolderDetail extends Folder {
  image_count: number
}

export async function getFolder(getToken: GetToken, id: string): Promise<FolderDetail>
```

#### Scenario: Panel fetches folder detail on mount

- **WHEN** `FolderPanelContent` renders for a given folder
- **THEN** it calls `getFolder(getToken, folder.id)` and uses the returned `image_count` to drive the export button's disabled state

#### Scenario: Panel refetches when the active folder changes

- **WHEN** the user selects a different folder while the panel is open
- **THEN** the folder detail query re-runs for the newly selected folder's `id`

---

### Requirement: Export folder download helper

`frontend/src/lib/folders.ts` SHALL export an `exportFolder` helper that calls `GET /folders/:id/export` and returns the response body as a `Blob`.

```ts
export async function exportFolder(getToken: GetToken, id: string): Promise<Blob>
```

- The request SHALL include the `Authorization` header (via the existing `apiFetch` wrapper)
- If the response is not `ok`, the helper SHALL throw an error

#### Scenario: Successful export returns a Blob

- **WHEN** `exportFolder` is called and the backend responds `200 OK` with an `application/zip` body
- **THEN** the helper resolves to a `Blob` containing the response body

#### Scenario: Failed export throws

- **WHEN** `exportFolder` is called and the backend responds with a non-`ok` status
- **THEN** the helper throws an error

---

### Requirement: Export folder button

`FolderPanelContent` SHALL render an "Export folder" button (mirroring the placement/style of `DownloadButton` for images — a sticky footer button).

Behavior:
- The button is disabled when the folder's `image_count` is `0` (including while the folder detail query is loading, since `image_count` is not yet known)
- On click, the button enters a "Preparing export..." state (spinner + label change, mirroring `DownloadButton`'s "Downloading…" state) and calls `exportFolder(getToken, folder.id)`
- On success, the returned `Blob` is converted to an object URL, and a download is triggered via a temporary `<a>` element with `download="<folder name>.zip"`
- On failure, an error toast is shown (`"Failed to export folder"`) and the button returns to its normal state
- The object URL is revoked after the download is triggered

#### Scenario: Button disabled for an empty folder

- **WHEN** the folder detail query resolves with `image_count: 0`
- **THEN** the "Export folder" button is disabled

#### Scenario: Button enabled for a non-empty folder

- **WHEN** the folder detail query resolves with `image_count` greater than `0`
- **THEN** the "Export folder" button is enabled

#### Scenario: Clicking the button downloads a zip

- **WHEN** the user clicks "Export folder" on a non-empty folder
- **THEN** the button shows a "Preparing export..." state until `exportFolder` resolves
- **AND** on success, a download of `<folder name>.zip` is triggered

#### Scenario: Export failure shows an error toast

- **WHEN** `exportFolder` rejects
- **THEN** an error toast `"Failed to export folder"` is shown
- **AND** the button returns to its normal (non-loading) state
