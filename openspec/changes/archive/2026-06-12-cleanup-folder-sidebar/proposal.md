## Why

`frontend/src/features/folder-sidebar/components/FolderSidebar.tsx` (270
lines) already has good extraction precedent in the same feature
(`FolderItem`, `FolderNameDialog`, `RootDropZone`, `UnsortedEntry`,
`useFolderMutations`), but two pieces of UI were left inline that don't fit
that precedent: two confirmation dialogs that mirror an existing
gallery-feature pattern, and a "Trash" nav entry that is structurally the
same kind of thing as the already-extracted `UnsortedEntry`, just left
inline alongside its own state and mutation.

## What Changes

- Extract the inline "Delete folder" confirmation dialog into
  `features/folder-sidebar/components/DeleteFolderDialog.tsx`, mirroring
  `features/gallery/components/DeleteImageDialog.tsx`'s shape: `{ folder:
  Folder | null, onCancel: () => void, onConfirm: () => void }`,
  presentational only.
- Extract the "Trash" nav row — including its context menu, the "Empty
  trash" confirmation dialog, and `emptyTrashMutation` — into a new
  self-contained `features/folder-sidebar/components/TrashEntry.tsx`,
  following the same one-component-per-file convention as the existing
  `UnsortedEntry.tsx` (active-state styling + `onClick`, but owning its own
  dialog and mutation rather than a `useDroppable`).
- The "Empty trash" confirmation dialog itself becomes
  `features/folder-sidebar/components/EmptyTrashDialog.tsx`, owned by and
  rendered from `TrashEntry`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

`frontend-structure` — adds requirements describing the new component file
locations (`DeleteFolderDialog.tsx`, `TrashEntry.tsx`,
`EmptyTrashDialog.tsx`) under `features/folder-sidebar/components/`. This is
a zero-functional-change structural cleanup; no behavior changes.

## Impact

- `frontend/src/features/folder-sidebar/components/FolderSidebar.tsx`
  shrinks — loses the inline delete-folder dialog, the inline "Trash" nav
  row + context menu, the inline empty-trash dialog, `confirmEmptyTrash`
  state, and `emptyTrashMutation`.
- New files: `features/folder-sidebar/components/DeleteFolderDialog.tsx`,
  `features/folder-sidebar/components/TrashEntry.tsx`,
  `features/folder-sidebar/components/EmptyTrashDialog.tsx`, plus test files
  for each.
- No change to `FolderSidebar`'s external props, `AppLayout`'s usage of
  `FolderSidebar`, or any user-visible behavior.
- No backend, API, database, or browser-extension changes. No new
  dependencies.
