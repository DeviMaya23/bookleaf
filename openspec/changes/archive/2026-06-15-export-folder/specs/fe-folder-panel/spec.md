## ADDED Requirements

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
