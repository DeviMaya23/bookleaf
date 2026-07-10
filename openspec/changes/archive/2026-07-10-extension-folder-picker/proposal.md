## Why

The default folder selector in extension Settings is a native `<select>` that renders folders as a flat list, hiding the parent–child hierarchy users have already organised their library into. A dedicated picker panel can expose the real tree and make the right folder easier to find at a glance.

## What Changes

- `GET /folders` response shape gains `parent_id: string | null` on each folder object (already available on the backend; the extension's `getFolders()` return type is updated to match).
- `extensions/src/lib/folderTree.ts` is added: pure `buildFolderTree` and `filterFolderTree` utilities, ported from the FE (`frontend/src/features/folder-sidebar/lib/folderTree.ts`).
- A new `FolderPicker` panel view is added to the extension popup — a third entry in the existing view stack (main → settings → folder-picker).
- The "Default folder" row in Settings changes from a native `<select>` to a button showing the current folder name; tapping it navigates to the FolderPicker panel.
- The FolderPicker panel renders the full folder tree with indentation, a filter input at the top, and confirm-on-click (selecting a row immediately calls `setDefaultFolder` and navigates back to Settings).
- Filter behaviour matches the FE: a parent is shown if it matches the query or has a matching descendant; non-matching children are hidden.

## Capabilities

### New Capabilities

- `extension-folder-picker`: FolderPicker panel view, folderTree utilities, and filter logic for the extension's default-folder selection flow.

### Modified Capabilities

- `extension-default-folder`: "Folder list fetch from API" gains `parent_id` in the return type; "Save destination section in Settings UI" changes from a `<select>` to a button that opens the FolderPicker panel.
- `extension-popup-settings`: "Settings view navigation" gains a FolderPicker as a third navigable view reachable from the "Default folder" row.

## Impact

- **`extensions/src/lib/api.ts`** — `getFolders()` return type updated to include `parent_id: string | null`.
- **`extensions/src/lib/folderTree.ts`** — new file with `buildFolderTree` and `filterFolderTree`.
- **`extensions/src/popup/FolderPicker.tsx`** — new component.
- **`extensions/src/popup/Settings.tsx`** — "Default folder" row replaced: `<select>` removed, button added.
- **`extensions/src/popup/App.tsx`** — view stack updated to include the `folder-picker` view; navigation wired between Settings and FolderPicker.
- **Backend** — no changes; `parent_id` is already returned by `GET /folders`.
