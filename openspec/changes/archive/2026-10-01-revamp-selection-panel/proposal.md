## Why

The current selection panel is a placeholder: it offers a flat list folder picker and a trash button, with no way to add images to multiple folders, no bulk download, and no exit affordance within the panel itself. This revamp completes the selection mode experience.

## What Changes

- Replace the flat folder search list with a chip-style multi-folder picker (matching the image detail panel's `FolderInput`), with a deferred Apply button so folder adds are committed explicitly
- Add a "Download as ZIP" button that streams selected images as a zip archive via a new backend endpoint
- Add an exit button (×) inline with the selected count in the panel header so users can leave select mode without reaching for the toolbar
- Wire up `onExitSelectMode` — already present in `PanelContent.selection` but not forwarded to `SelectionPanelBody`

## Capabilities

### New Capabilities

- `image-bulk-export`: New `POST /images/bulk/export` endpoint — accepts a list of image IDs, streams a ZIP archive of the originals. Mirrors the existing `folder-export` endpoint and reuses its zip-writing logic.
- `fe-selection-panel`: Revamped selection panel UI — chip-style multi-folder picker with Apply button, download ZIP button, and exit (×) button in the header.

### Modified Capabilities

## Impact

- **Backend**: New handler + usecase method `BulkExportImages` in image usecase; new repository method `GetManyByIDs` on `ImageRepository`; shared zip-writing helper extracted from `folderUsecase.ExportFolder`. New route registered in `main.go`. No changes to existing folder or bulk-folder-add endpoints.
- **Frontend**: `SelectionPanelBody` — fully replaces `SelectionFolderPicker` with `FolderInput` (TokenInput chip style) + Apply button; adds download button and exit button. `RightPanel` — passes `onExitSelectMode` through to `SelectionPanelBody`. New lib function `bulkExportImages` in `src/lib/images.ts`. No changes to `AppLayout` selection state or `GalleryToolbar`.
- **Extension**: Not affected.
- **Dependencies**: No new packages — `archive/zip` already used in `folder_usecase.go`; `FolderInput`/`TokenInput` already exist in the FE.
