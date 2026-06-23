## 1. Folder context menu "View details" item

- [x] 1.1 Add an `onViewDetails?: (folder: Folder) => void` prop to `FolderItem` (`FolderItem.tsx`).
- [x] 1.2 In `FolderItem`'s `ContextMenuContent` (`FolderItem.tsx:100-124`), add a "View details" item above "New subfolder", rendered only when `useIsCoarsePointer()` is true, calling `onViewDetails`. Use a `ContextMenuSeparator` below it, mirroring the existing image-card pattern (`ImageGrid.tsx:75-80`).
- [x] 1.3 Thread `onViewDetails` from `FolderItem` through `FolderSidebar` (`FolderSidebar.tsx`) as a new prop, alongside the existing `onRename`/`onDelete`/`onNewSubfolder`/`onChangeIcon` props, passed down to every `FolderItem` instance (including the recursive children render at `FolderSidebar.tsx:151-167`).

## 2. Gesture remap in AppLayout

- [x] 2.1 Add a new `handleFolderViewDetails` callback in `AppLayout.tsx`: `setFolderPanelOpen(true); setSelectedImage(null); setAutoFocusTitle(false)` — the same logic `onFolderSelect` runs today.
- [x] 2.2 Update the `onFolderSelect` callback passed to `FolderSidebar` (`AppLayout.tsx:206`): branch on `useIsCoarsePointer()` — coarse skips `setFolderPanelOpen(true)` (keeps `setSelectedImage(null)`/`setAutoFocusTitle(false)`); fine keeps the existing unconditional behavior unchanged.
- [x] 2.3 Wire the new `onFolderViewDetails` prop on `FolderSidebar` to `handleFolderViewDetails`.

## 3. Tests

- [x] 3.1 Update `FolderItem` tests (or add a new colocated test file if none exists) covering: "View details" item present when coarse pointer, absent when fine pointer, calls `onViewDetails` with the correct folder.
- [x] 3.2 Update `FolderSidebar` tests covering: `onViewDetails` threaded through to the rendered `FolderItem`s, including nested/child folders.
- [x] 3.3 Update or add `AppLayout`-level coverage (or the most relevant existing integration test) covering: selecting a different folder does not open the right panel on a coarse-pointer device; selecting a different folder still opens the panel on a fine-pointer device; an already-open panel updates to the newly selected folder's content on a coarse-pointer device without closing.

## 4. Verification

- [x] 4.1 Manually verify on a touch device/emulated coarse pointer: switching folders via tap does not open the bottom drawer; long-press on a folder shows the context menu with "View details" above "New subfolder"; selecting "View details" opens the bottom drawer with that folder's metadata.
- [x] 4.2 Manually verify on desktop (fine pointer): selecting a folder still auto-opens the right panel sidebar exactly as before; no "View details" item appears in the folder context menu.
- [x] 4.3 Run `npm run build` in `frontend/` and fix any errors.
- [x] 4.4 Run `npm run lint` in `frontend/` and fix any issues.
