## 1. Setup

- [x] 1.1 Add `QueryClientProvider` wrapping the app in `main.tsx`
- [x] 1.2 Add shadcn/ui `Dialog`, `ContextMenu` components via shadcn CLI

## 2. Folder API Layer

- [x] 2.1 Create `src/lib/folders.ts` with typed API functions: `getFolders`, `createFolder`, `renameFolder`, `deleteFolder` — all using `apiFetch`

## 3. FolderSidebar Component

- [x] 3.1 Create `src/components/FolderSidebar.tsx` — extract existing sidebar JSX from `AppLayout` into this component
- [x] 3.2 Replace hardcoded folder list with `useQuery(['folders'], getFolders)` and render API folders below the "Unsorted" entry + divider
- [x] 3.3 Update `AppLayout` to render `<FolderSidebar />` in place of the inlined sidebar JSX

## 4. Create Folder

- [x] 4.1 Create `src/components/FolderNameDialog.tsx` — accepts `title`, `initialValue`, `onSubmit`, `open`, `onOpenChange` props; renders a Dialog with a text input and confirm button
- [x] 4.2 Wire "+ New folder" button in `FolderSidebar` to open `FolderNameDialog` and call `createFolder` on submit
- [x] 4.3 Add `useMutation` for create with `invalidateQueries(['folders'])` on success

## 5. Rename Folder

- [x] 5.1 Wrap each API folder row in `FolderSidebar` with a shadcn `ContextMenu` containing "Rename" and "Delete" items
- [x] 5.2 Wire "Rename" context menu item to open `FolderNameDialog` pre-populated with the folder's current name and call `renameFolder` on submit
- [x] 5.3 Add `useMutation` for rename with `invalidateQueries(['folders'])` on success

## 6. Delete Folder

- [x] 6.1 Create a delete confirmation Dialog in `FolderSidebar` (or inline): shows folder name, "Cancel" and "Delete" buttons
- [x] 6.2 Wire "Delete" context menu item to open the confirmation Dialog; only call `deleteFolder` if user confirms
- [x] 6.3 Add `useMutation` for delete with `invalidateQueries(['folders'])` on success
