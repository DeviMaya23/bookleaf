## Why

The folder side panel already fetches `image_count` (via `getFolder`) to gate the "Export folder" button, but never surfaces it to the user. Showing the count gives users quick context about a folder's contents without opening it.

## What Changes

- `FolderPanelContent` displays a small subtitle line under the folder name showing the image count (e.g. "12 images", "1 image", "0 images"), sourced from the existing `folderDetail.image_count` returned by `getFolder`.
- While the folder detail query is loading (`folderDetail` is `undefined`), the subtitle is omitted (no placeholder/skeleton).

## Capabilities

### Modified Capabilities
- `fe-folder-panel`: `FolderPanelContent` adds a header subtitle showing the folder's image count, derived from the already-fetched `image_count`.

## Impact

- **Frontend**: `frontend/src/features/right-panel/components/FolderPanelContent.tsx` (render the subtitle) and its test file. No new API calls, no backend changes.
