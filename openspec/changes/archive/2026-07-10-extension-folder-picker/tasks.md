## 1. API Client

- [x] 1.1 Update `getFolders()` return type in `extensions/src/lib/api.ts` to include `parent_id: string | null` on each folder object

## 2. Folder Tree Utilities

- [x] 2.1 Create `extensions/src/lib/folderTree.ts` with `FolderNode` type, `buildFolderTree`, and `filterFolderTree`, ported from `frontend/src/features/folder-sidebar/lib/folderTree.ts` and adapted to use the extension's local folder shape (`{ id: string; name: string; parent_id: string | null }`)

## 3. FolderPicker Component

- [x] 3.1 Create `extensions/src/popup/FolderPicker.tsx`: header with back arrow labelled "Default folder", filter text input with placeholder `"Search folders…"`, a fixed "None (Unsorted)" first row, and a scrollable tree area for folder rows; each nesting level adds 12px left padding; accept `onBack: () => void` and `c: Colors` props
- [x] 3.2 On mount, call `getFolders()` and build the tree with `buildFolderTree`; disable the filter input while loading; on fetch error render only the "None (Unsorted)" row and re-enable the filter input
- [x] 3.3 Apply `filterFolderTree` to the tree when the filter input is non-empty; the "None (Unsorted)" row is always rendered regardless of the filter value
- [x] 3.4 On mount, call `getDefaultFolder()` and visually indicate the matching row (e.g. a checkmark or bold name)
- [x] 3.5 On any row click, call `setDefaultFolder({ id, name })` for a named folder or `clearDefaultFolder()` for "None (Unsorted)", then call `onBack()`

## 4. Settings Simplification

- [x] 4.1 Remove `folders`, `foldersLoading`, `defaultFolderId` state, the `getFolders` import, and the `handleFolderChange` handler from `Settings.tsx`
- [x] 4.2 Add `defaultFolderName: string | null` state initialised from `getDefaultFolder()` on mount; add `onOpenFolderPicker: () => void` prop; replace the `<select>` in the "Default folder" row with a button that displays `defaultFolderName ?? "None"` and calls `onOpenFolderPicker` on click

## 5. View Stack Wiring

- [x] 5.1 Extend the `View` type in `App.tsx` to `"main" | "settings" | "folder-picker"`; import `FolderPicker`; add a render branch for `view === "folder-picker"` that renders `<FolderPicker onBack={() => setView("settings")} c={c} />`
- [x] 5.2 Pass `onOpenFolderPicker={() => setView("folder-picker")}` to `<Settings>` in the settings render branch of `App.tsx`

## 6. Quality

- [x] 6.1 Run `npm run build` in `extensions/` and fix any type errors
- [x] 6.2 Run `npm run lint` in `extensions/` and fix any lint issues
