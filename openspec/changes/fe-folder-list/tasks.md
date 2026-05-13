## 1. Setup

- [ ] 1.1 Add `QueryClientProvider` wrapping the app in `main.tsx`
- [ ] 1.2 Add shadcn/ui `Dialog`, `ContextMenu` components via shadcn CLI

## 2. Folder API Layer

- [ ] 2.1 Create `src/lib/folders.ts` with typed API functions: `getFolders`, `createFolder`, `renameFolder`, `deleteFolder` — all using `apiFetch`

## 3. FolderSidebar Component

- [ ] 3.1 Create `src/components/FolderSidebar.tsx` — extract existing sidebar JSX from `AppLayout` into this component
- [ ] 3.2 Replace hardcoded folder list with `useQuery(['folders'], getFolders)` and render API folders below the "Unsorted" entry + divider
- [ ] 3.3 Update `AppLayout` to render `<FolderSidebar />` in place of the inlined sidebar JSX

## 4. Create Folder

- [ ] 4.1 Create `src/components/FolderNameDialog.tsx` — accepts `title`, `initialValue`, `onSubmit`, `open`, `onOpenChange` props; renders a Dialog with a text input and confirm button
- [ ] 4.2 Wire "+ New folder" button in `FolderSidebar` to open `FolderNameDialog` and call `createFolder` on submit
- [ ] 4.3 Add `useMutation` for create with `invalidateQueries(['folders'])` on success

## 5. Rename Folder

- [ ] 5.1 Wrap each API folder row in `FolderSidebar` with a shadcn `ContextMenu` containing "Rename" and "Delete" items
- [ ] 5.2 Wire "Rename" context menu item to open `FolderNameDialog` pre-populated with the folder's current name and call `renameFolder` on submit
- [ ] 5.3 Add `useMutation` for rename with `invalidateQueries(['folders'])` on success

## 6. Delete Folder

- [ ] 6.1 Create a delete confirmation Dialog in `FolderSidebar` (or inline): shows folder name, "Cancel" and "Delete" buttons
- [ ] 6.2 Wire "Delete" context menu item to open the confirmation Dialog; only call `deleteFolder` if user confirms
- [ ] 6.3 Add `useMutation` for delete with `invalidateQueries(['folders'])` on success
