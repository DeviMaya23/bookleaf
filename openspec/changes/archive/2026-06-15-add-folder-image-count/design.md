## Context

`FolderPanelContent` (`frontend/src/features/right-panel/components/FolderPanelContent.tsx`) already fetches `folderDetail` via `getFolder(getToken, folder.id)`, which includes `image_count`. This value is currently only used to gate the "Export folder" button's disabled state — it isn't shown to the user.

## Goals / Non-Goals

**Goals:**
- Show the folder's image count to the user as a small subtitle under the folder name in the panel header.
- Use the existing `folderDetail.image_count` value — no new fetch.

**Non-Goals:**
- No new metadata fields (subfolder count, dates, etc.) — out of scope for this change.
- No backend changes.

## Decisions

- **Placement**: render the subtitle directly below the folder name input, inside the existing header `<div>` (`FolderPanelContent.tsx` lines 68-82).
- **Text format**: `"{n} image"` / `"{n} images"` based on `image_count` (singular for `1`, plural otherwise, including `0`).
- **Loading state**: while `folderDetail` is `undefined` (initial load), render nothing — no skeleton/placeholder, consistent with avoiding layout flicker for a single short line that resolves quickly.
- **Styling**: small muted text (e.g. `text-xs text-muted-foreground`), matching the visual weight of other secondary labels in the right panel (e.g. the "Notes" section label).

## Risks / Trade-offs

- [Count becomes stale if images are added/removed elsewhere while the panel is open] → Already accepted by the existing export-button gating logic (same `folderDetail` query, `['folder', folder.id]`); no new staleness introduced.
