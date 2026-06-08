## 1. Header icon button

- [x] 1.1 Add an icon button (using `Button` with `variant="ghost"` and `size="icon-xs"` or `icon-sm`) adjacent to the "FOLDERS" section label in `FolderSidebar.tsx`, wired to `setNewFolderOpen(true)`
- [x] 1.2 Adjust the section label row's layout (flex/justify) so the label and icon button align cleanly

## 2. Conditional footer button

- [x] 2.1 Wrap the existing footer "+ New folder" button in a `folders.length === 0` condition so it renders only when the account has no folders

## 3. Tests

- [x] 3.1 Update `FolderSidebar.test.tsx`: assert the header icon button is always rendered and opens the new-folder dialog when clicked
- [x] 3.2 Update `FolderSidebar.test.tsx`: assert the footer "+ New folder" button is present when `folders` is empty and absent when `folders` is non-empty

## 4. Verification

- [x] 4.1 Run `npm run build` in `frontend/` and fix any issues that arise
