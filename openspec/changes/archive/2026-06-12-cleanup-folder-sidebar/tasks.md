## 1. Extract DeleteFolderDialog

- [x] 1.1 Create `frontend/src/features/folder-sidebar/components/DeleteFolderDialog.tsx` with props `{ folder: Folder | null, onCancel: () => void, onConfirm: () => void }`, mirroring `DeleteImageDialog.tsx`'s structure, containing the "Delete folder" dialog markup moved from `FolderSidebar.tsx`.
- [x] 1.2 Replace the inline delete-folder `Dialog` in `FolderSidebar.tsx` with `<DeleteFolderDialog folder={deleteTarget} onCancel={() => setDeleteTarget(null)} onConfirm={handleDelete} />`; remove the now-unused inline markup.
- [x] 1.3 Add `DeleteFolderDialog.test.tsx` covering: dialog renders with the folder's name when `folder` is non-null, stays closed when `folder` is null, `onCancel` fires on Cancel/close, `onConfirm` fires on Delete.

## 2. Extract TrashEntry and EmptyTrashDialog

- [x] 2.1 Create `frontend/src/features/folder-sidebar/components/EmptyTrashDialog.tsx` with props `{ open: boolean, onCancel: () => void, onConfirm: () => void }`, containing the "Empty trash" dialog markup moved from `FolderSidebar.tsx`.
- [x] 2.2 Create `frontend/src/features/folder-sidebar/components/TrashEntry.tsx` with props `{ active: boolean, onClick: () => void }`, containing the "Trash" nav row, its `ContextMenu`/"Empty trash" item, local open state, `emptyTrashMutation` (moved verbatim from `FolderSidebar.tsx`), and rendering `<EmptyTrashDialog />`.
- [x] 2.3 Replace the inline "Trash" nav row, context menu, empty-trash dialog, `confirmEmptyTrash` state, and `emptyTrashMutation` in `FolderSidebar.tsx` with `<TrashEntry active={view.type === 'trash'} onClick={() => navigate('/trash')} />`.
- [x] 2.4 Add `EmptyTrashDialog.test.tsx` covering: renders open/closed per `open` prop, `onCancel`/`onConfirm` fire on the respective buttons.
- [x] 2.5 Add `TrashEntry.test.tsx` covering: active-state styling, context menu shows "Empty trash", confirming the dialog calls `emptyTrash` and invalidates the trash query — adapting the scenarios currently in `FolderSidebar.test.tsx`'s "FolderSidebar trash context menu" describe block.

## 3. Final checks

- [x] 3.1 Remove the "FolderSidebar trash context menu" describe block from `FolderSidebar.test.tsx` (now covered by `TrashEntry.test.tsx`).
- [x] 3.2 Run `npm run build` and fix any resulting issues.
