## 1. usePersistedBoolean hook

- [x] 1.1 Create `frontend/src/hooks/usePersistedBoolean.ts` — reads from `localStorage` on mount (defaulting to provided value if absent), writes on every change

## 2. RightPanel component refactor

- [x] 2.1 Replace the `RightPanelProps` union with the new `PanelContent` discriminated union (`image` | `folder` | `selection` | `neutral`); remove `onClose` from all modes; add `onExitSelectMode` for the selection branch
- [x] 2.2 Add `collapsed` state via `usePersistedBoolean('bookleaf-right-panel-collapsed', false)` inside `RightPanel`
- [x] 2.3 Add the panel shell header with chevron toggle button (`ChevronLeft`/`ChevronRight` from lucide); toggle calls `setCollapsed`
- [x] 2.4 Change the `<aside>` to always render; width class toggles between `w-80` and `w-8`; in collapsed state render only the chevron strip, not the content
- [x] 2.5 Add `neutral` mode rendering — shows the `viewLabel` string (e.g. "All", "Unsorted", "Trash") with no other fields
- [x] 2.6 Remove the ✕ close button from `ImagePanelBody` (currently overlaid on the thumbnail)

## 3. FolderPanelContent and SelectionPanelBody cleanup

- [x] 3.1 Remove `onClose` prop and ✕ button from `FolderPanelContent`
- [x] 3.2 Remove `onClose` prop and ✕ button from `SelectionPanelBody`

## 4. AppLayout wiring

- [x] 4.1 Derive `panelContent: PanelContent` from existing state using the priority order: selection → image → folder → neutral (view label from `view.type`)
- [x] 4.2 Replace the three-branch conditional `RightPanel` mount with a single always-rendered `<RightPanel panelContent={panelContent} onExitSelectMode={exitSelectMode} focusMode={focusMode} />`; keep the focus-mode gate (selection mode bypasses it)
- [x] 4.3 Remove all `onClose` callbacks passed to `RightPanel`; remove `folderPanelOpen` state and its setters if no longer needed (verify no other consumers)

## 5. Update AppLayout tests

- [x] 5.1 Update the `RightPanel` mock to accept the new `panelContent` prop shape (including `neutral` mode) and remove `onClose`
- [x] 5.2 Rewrite assertions that check `queryByTestId('right-panel')).not.toBeInTheDocument()` after non-focus-mode actions — replace with `data-mode="neutral"` or appropriate content assertions
- [x] 5.3 Update focus-mode tests that assert the panel is absent — these remain valid (focus mode still unmounts the panel) but verify they still pass with the new wiring
- [x] 5.4 Update the selection-panel close test ("closing the selection panel turns select mode off") — remove it; the toolbar toggle is now the only exit

## 6. Update RightPanel tests

- [x] 6.1 Update `renderPanel` helper and all test calls to use the new prop shape (no `onClose`); update the `hidden sm:flex` assertion to match the new always-rendered aside classes
- [x] 6.2 Add a test for the `neutral` mode rendering
- [x] 6.3 Add a test that the chevron toggle changes the collapsed state and persists it to `localStorage`

## 7. Lint and build

- [x] 7.1 Run `npm run lint` in `frontend/` and fix any issues
- [x] 7.2 Run `npm run build` in `frontend/` and fix any type errors
