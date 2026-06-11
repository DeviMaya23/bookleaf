## ADDED Requirements

### Requirement: Folder sidebar confirmation dialogs are separate presentational components
The "Delete folder" and "Empty trash" confirmation dialogs SHALL each be their own presentational component under `frontend/src/features/folder-sidebar/components/`, mirroring `features/gallery/components/DeleteImageDialog.tsx`'s pattern, rather than inline JSX within `FolderSidebar.tsx`.

#### Scenario: Delete folder dialog component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/folder-sidebar/components/DeleteFolderDialog.tsx`
  exists

#### Scenario: Empty trash dialog component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/folder-sidebar/components/EmptyTrashDialog.tsx`
  exists

### Requirement: Trash nav entry is a separate component
The "Trash" nav row, its context menu, and the empty-trash confirmation flow SHALL be extracted into their own `TrashEntry` component under `frontend/src/features/folder-sidebar/components/`, following the same one-component-per-file convention as `UnsortedEntry.tsx`, rather than left inline within `FolderSidebar.tsx`.

#### Scenario: Trash entry component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/folder-sidebar/components/TrashEntry.tsx`
  exists
