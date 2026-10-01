# fe-selection-panel Specification

## Purpose
TBD - created by archiving change revamp-selection-panel. Update Purpose after archive.
## Requirements
### Requirement: Exit button in selection panel header

The selection panel header SHALL render a `×` icon button to the right of the "N selected" count text, in the same flex row. Clicking it SHALL call `onExitSelectMode`, which exits select mode and clears the selection. `SelectionPanelBody` SHALL accept `onExitSelectMode` as a required prop. `RightPanel.deriveContent` SHALL forward `panelContent.onExitSelectMode` to `SelectionPanelBody`.

#### Scenario: Clicking × exits select mode

- **WHEN** the selection panel is visible with one or more images selected
- **AND** the user clicks the × button in the header
- **THEN** select mode is exited
- **AND** the selection is cleared

#### Scenario: RightPanel forwards onExitSelectMode

- **WHEN** `RightPanel` renders selection mode content
- **THEN** `SelectionPanelBody` receives the `onExitSelectMode` prop from `panelContent.onExitSelectMode`

---

### Requirement: Chip-style multi-folder picker with deferred Apply

The "Add to folder" section of the selection panel SHALL use `FolderInput` (the `TokenInput`-based chip component) in place of the current `SelectionFolderPicker` flat list. The `FolderInput` SHALL render with all available folders as suggestions and no pre-selected folders. An "Apply" button SHALL appear below the chip input. The Apply button SHALL be disabled when no folders are selected in the chip input.

Clicking Apply SHALL call `onAddToFolder` once per selected folder chip, then clear the chip input. `onAddToFolder` existing behavior (fire `bulkAddImagesToFolder` and exit select mode on success) is unchanged.

`SelectionFolderPicker` SHALL be deleted.

#### Scenario: User adds two folders and clicks Apply

- **WHEN** the user adds two folder chips and clicks Apply
- **THEN** `onAddToFolder` is called once with the first folder's ID
- **AND** `onAddToFolder` is called once with the second folder's ID

#### Scenario: Apply button is disabled with no folders selected

- **WHEN** no folder chips have been added to the input
- **THEN** the Apply button is disabled and non-interactive

#### Scenario: Apply button is enabled when at least one folder chip is present

- **WHEN** at least one folder chip is present in the input
- **THEN** the Apply button is enabled

#### Scenario: Folder chip input shows available folders as suggestions

- **WHEN** the user focuses the folder chip input
- **AND** begins typing a folder name
- **THEN** matching folders appear in the suggestion dropdown

#### Scenario: Removing a chip before Apply fires no API call

- **WHEN** the user adds a folder chip then removes it via the × on the chip
- **AND** the user has not clicked Apply
- **THEN** `onAddToFolder` is not called

---

### Requirement: Download ZIP button in selection panel

The selection panel SHALL render a "Download as ZIP" button in the scrollable body, above the "Move to trash" button. Clicking it SHALL call `onDownloadZip`. While the download is in progress the button SHALL be disabled and show a loading indicator. On error, a toast SHALL be shown. The button SHALL be disabled when `selectedCount` is zero.

`SelectionPanelBody` SHALL accept `onDownloadZip: () => Promise<void>` as a required prop. The parent (`AppLayout` via `RightPanel`) SHALL provide the handler, which calls `POST /images/bulk/export` via the new `bulkExportImages` lib function, receives the blob, and triggers a browser download with filename `bookleaf-export.zip`.

#### Scenario: Clicking Download triggers a zip download

- **WHEN** the user clicks the "Download as ZIP" button
- **THEN** the browser initiates a file download named `bookleaf-export.zip`

#### Scenario: Download button is disabled while download is in progress

- **WHEN** a download is in progress
- **THEN** the "Download as ZIP" button is disabled

#### Scenario: Download error shows a toast

- **WHEN** the download request fails
- **THEN** a toast error message is shown
- **AND** the button returns to its enabled state

---

### Requirement: bulkExportImages frontend lib function

`src/lib/images.ts` SHALL export a function:

```ts
export async function bulkExportImages(getToken: GetToken, imageIds: string[]): Promise<Blob>
```

It SHALL POST to `/images/bulk/export` with `{ image_ids: imageIds }` and return the response as a `Blob`. It SHALL throw an `Error` if the response is not ok.

#### Scenario: Returns a Blob on success

- **WHEN** the server responds with `200 OK` and a zip body
- **THEN** `bulkExportImages` resolves with a `Blob`

#### Scenario: Throws on non-ok response

- **WHEN** the server responds with a non-2xx status
- **THEN** `bulkExportImages` throws an `Error`

