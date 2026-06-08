## Why

The "+ New folder" affordance currently lives only as a full-width text button in the sidebar footer, separated from the "FOLDERS" section it acts on. Users who already have folders expect a quick-add control right next to the section header (a common sidebar convention), while brand-new accounts benefit more from a prominent, explicit CTA when the folder list is empty. Splitting these two needs lets each affordance serve the moment it's best suited for.

## What Changes

- Add a small icon button (`+`) beside the "FOLDERS" section label in the sidebar, always visible, using the existing `ghost` / `icon-xs` (or `icon-sm`) button convention (as seen on the dialog close button). Clicking it opens the same new-folder dialog (`setNewFolderOpen(true)`) as today.
- Make the existing full-width "+ New folder" footer button conditional: it SHALL render only when the user's folder list is empty, acting as an onboarding prompt for brand-new accounts, and SHALL disappear once at least one folder exists.
- Both controls trigger the identical existing new-folder flow — no change to the dialog, the `POST /folders` call, or the refetch behavior.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `fe-sidebar-nav`: the "FOLDERS" section label now has an adjacent icon button that opens the new-folder dialog
- `app-shell`: the "+ New folder" footer affordance is no longer unconditionally visible — it now appears only when the folder list is empty
- `folder-management`: the new-folder creation flow can now be triggered from either the header icon button or the (conditionally shown) footer button, both opening the same dialog

## Impact

- Affected file: `frontend/src/components/FolderSidebar.tsx` (header row near the "FOLDERS" label, and the footer block containing the existing "+ New folder" button)
- No backend, API, or routing changes — purely a frontend layout/visibility change reusing the existing new-folder dialog and mutation
