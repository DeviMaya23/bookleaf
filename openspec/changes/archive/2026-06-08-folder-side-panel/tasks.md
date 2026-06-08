## 1. Folder update API

- [x] 1.1 Add `updateFolderDetails(getToken, id, { name?, description? })` to `frontend/src/lib/folders.ts`, wrapping `PUT /folders/:id` (mirrors `renameFolder`/`moveFolder`)

## 2. FolderPanelContent component

- [x] 2.1 Create `frontend/src/components/FolderPanelContent.tsx` accepting `{ folder, onClose }` and rendering editable title and description fields pre-populated from `folder`
- [x] 2.2 Implement title field: `useState` + `useRef(original)` reset on `folder.id` change, `onBlur` diff-and-save via `updateFolderDetails` (only if changed and non-empty), revert to original if cleared
- [x] 2.3 Implement description field: same pattern, but allow saving an empty value as `null`
- [x] 2.4 Wire both fields through a single `useMutation` that invalidates `['folders']` and shows success/error toasts (mirrors `RightPanel`'s `saveMutation`)
- [x] 2.5 Render close button wired to `onClose`

## 3. RightPanel mode switch

- [x] 3.1 Update `RightPanel` to accept either an image-mode or folder-mode selection and conditionally render the existing image body or `<FolderPanelContent>` within the shared chrome (width, close, footer as applicable)

## 4. AppLayout wiring

- [x] 4.1 Add a thin `folderPanelOpen: boolean` trigger to `AppLayout` (not a stored `Folder` copy), set by the folder-select callback and cleared by the panel's close button / image selection — do *not* reset it on the `viewKey`-change effect (that would clobber the same-action `true` set by the folder click, reproducing the staleness bug); panel visibility is naturally gated by the derived folder being non-null, so navigating to non-folder views hides it without needing an explicit reset
- [x] 4.2 Derive the active folder's data for the panel from `view` (when `view.type === 'folder'`) plus the existing `['folders']` query — e.g. `view.type === 'folder' ? folders.find(f => f.id === view.id) ?? null : null` — gated by `folderPanelOpen`; ensure selecting an image clears `folderPanelOpen`, and selecting a different folder clears `selectedImage`, so the two stay mutually exclusive
- [x] 4.3 Pass the derived folder content into `RightPanel` and update the render guard to show the panel when either `selectedImage` or the derived folder content is active

## 5. FolderSidebar trigger

- [x] 5.1 Add an `onFolderSelect` callback prop to `FolderSidebar`, invoked from the folder click handler only when the clicked folder differs from the currently active folder (reusing the existing `isActive` comparison)
- [x] 5.2 Wire `AppLayout` to pass a handler that sets `folderPanelOpen` (and clears `selectedImage`) from this callback

## 6. Verification

- [x] 6.1 Run `npm run build` in `frontend/` and fix any type or build errors
