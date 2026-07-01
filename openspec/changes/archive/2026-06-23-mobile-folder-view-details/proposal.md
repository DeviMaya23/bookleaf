## Why

On coarse-pointer (mobile) devices, switching to a different folder in the sidebar currently auto-opens the right panel as a bottom drawer to show that folder's metadata — every folder switch interrupts the user with an unrequested drawer. This was carried over unchanged from the desktop sidebar behavior when `mobile-gallery-interactions` introduced the bottom-drawer shell; that change made the equivalent image-card flow opt-in (tap opens a lightbox, drawer access is opt-in via a "View details" context-menu item) but left folder-switch behavior on auto-open. This change brings folder behavior in line with that same opt-in pattern.

## What Changes

- On coarse-pointer devices, selecting a folder in the sidebar no longer auto-opens the right panel. The panel still updates if it is already open and showing the active folder's content (no spec change to that case), but switching folders SHALL NOT force it open.
- `FolderItem`'s existing long-press-triggered context menu (already used for "New subfolder"/"Rename"/"Change icon"/"Delete") gains a "View details" item, rendered only on coarse-pointer devices, that opens the right panel for that folder as a bottom drawer — mirroring the existing image-card "View details" item added by `mobile-gallery-interactions`.
- Fine-pointer (desktop) behavior is unchanged: selecting a folder continues to auto-open the right panel as a sidebar exactly as today.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `fe-right-panel`: "Right panel opens or updates when a folder is selected" changes from unconditional auto-open (either shell) to fine-pointer-only auto-open; on coarse-pointer devices selecting a folder no longer opens the panel.
- `folder-management`: Adds a "View details" item to the folder `ContextMenu` (the one already used for Rename/Delete/New subfolder/Change icon), rendered only on coarse-pointer devices, opening the right panel as a bottom drawer for that folder.

## Impact

- `frontend/src/app-shell/AppLayout.tsx` — the `onFolderSelect` callback passed to `FolderSidebar` (`AppLayout.tsx:206`) stops unconditionally calling `setFolderPanelOpen(true)`; a new callback (`onFolderViewDetails` or similar) wired to the same `setFolderPanelOpen(true)` logic is added for the context-menu item.
- `frontend/src/features/folder-sidebar/components/FolderItem.tsx` — new "View details" `ContextMenuItem`, gated on `useIsCoarsePointer()` (same hook added by `mobile-gallery-interactions`), above the existing items.
- `frontend/src/features/folder-sidebar/components/FolderSidebar.tsx` — threads the new callback from `FolderItem` up to `AppLayout`, alongside the existing `onRename`/`onDelete`/`onNewSubfolder`/`onChangeIcon` props.
- No new dependencies, no new components — reuses `useIsCoarsePointer` and the existing `RightPanel` bottom-drawer shell, both already in the codebase.
