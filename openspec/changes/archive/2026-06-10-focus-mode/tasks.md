## 1. Focus mode state and layout in AppLayout

- [x] 1.1 Add `focusMode` state (`useState<boolean>(false)`) to `AppLayout`
- [x] 1.2 Conditionally render `FolderSidebar` only when `!focusMode`
- [x] 1.3 Make `<main>`'s left margin conditional: `ml-[240px]` when `!focusMode`, `ml-0` when `focusMode`
- [x] 1.4 Append `&& !focusMode` to the existing `RightPanel` render condition (covers both `selectedImage` and `folderPanelOpen` branches)

## 2. Focus toggle button — gallery toolbar

- [x] 2.1 Add a `Toggle` button (from `ui/toggle.tsx`) using the `Focus` icon (lucide-react) as the leftmost element of the gallery toolbar, left of the search input
- [x] 2.2 Wire `aria-pressed={focusMode}` and a click handler that flips `focusMode`
- [x] 2.3 Style the active state consistent with existing `aria-pressed:bg-secondary aria-pressed:text-secondary-foreground` pattern used by filter chips

## 3. Focus toggle button — image viewer header

- [x] 3.1 Add `focusMode: boolean` and `onToggleFocusMode: () => void` props to `ImageViewer`
- [x] 3.2 Pass `focusMode` and a toggle callback from `AppLayout` to `ImageViewer`
- [x] 3.3 Render the same `Toggle` + `Focus` icon button as the leftmost element of the viewer header, left of the Back button, using the new props

## 4. Tests

- [x] 4.1 `AppLayout.test.tsx`: enabling focus mode hides `FolderSidebar` and removes `<main>`'s `ml-[240px]` class
- [x] 4.2 `AppLayout.test.tsx`: enabling focus mode hides an open `RightPanel` (image mode and folder mode)
- [x] 4.3 `AppLayout.test.tsx`: clicking an image card while focus mode is active updates selection state without rendering `RightPanel`, and disabling focus mode afterward reveals that selection in `RightPanel`
- [x] 4.4 `AppLayout.test.tsx`: double-clicking an image card while focus mode is active opens `ImageViewer` at full width with no `RightPanel`
- [x] 4.5 `ImageViewer.test.tsx`: Focus toggle renders left of the Back button and reflects `focusMode`/`onToggleFocusMode` props

## 5. Verification

- [x] 5.1 Run `npm run build` and fix any type or build errors
