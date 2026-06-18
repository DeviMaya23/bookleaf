## 1. Sidebar drawer

- [x] 1.1 Add `mobileOpen` and `onMobileClose` props to `FolderSidebar`; change its root `<aside>` className to translate off-canvas by default below `sm` (`-translate-x-full`), translate in when `mobileOpen` is true, and stay `translate-x-0` at `sm:` and up regardless of the prop
- [x] 1.2 Add a `mobileDrawerOpen` boolean (`useState`) to `AppLayout`, mirroring the existing `folderPanelOpen` pattern; pass it and a close handler into `FolderSidebar`
- [x] 1.3 Render a backdrop in `AppLayout` (`fixed inset-0 bg-black/35`, `sm:hidden`, `z-25`) only while `mobileDrawerOpen` is true; tapping it closes the drawer
- [x] 1.4 Close the drawer when a sidebar navigation entry is selected (existing `onFolderSelect`/view-change path) — only relevant below `sm`, no-op above it

## 2. Mobile top bar

- [x] 2.1 Create `MobileTopBar` component: fixed top, `sm:hidden`, hamburger button on the left, centered "Bookleaf" wordmark, `z-20`
- [x] 2.2 Wire the hamburger button to set `mobileDrawerOpen` to true in `AppLayout`
- [x] 2.3 Render `MobileTopBar` in `AppLayout`; add top padding/offset to the main content area so it isn't covered by the fixed bar below `sm`

## 3. Floating upload button

- [x] 3.1 Create `FloatingUploadButton` component: fixed bottom-right, `sm:hidden`, circular, `z-20`
- [x] 3.2 Wire it to the existing `setUploadOpen(true)` already in `AppLayout`
- [x] 3.3 Render `FloatingUploadButton` in `AppLayout`

## 4. Main content responsiveness

- [x] 4.1 Change `<main>`'s unconditional `ml-[240px]` (and the focus-mode `ml-0` branch) so the offset only applies at `sm:` and up; below `sm`, main is always full width regardless of focus mode

## 5. Right panel hidden below the breakpoint

- [x] 5.1 Change `RightPanel`'s root `<aside>` className from `flex` to `hidden sm:flex` so it never renders below `sm`, while `AppLayout`'s existing `selectedImage`/`folderPanelOpen` state and handlers stay untouched

## 6. Gallery toolbar mobile adjustments

- [x] 6.1 Wrap the sort `DropdownMenu` block in `GalleryToolbar` with `hidden sm:flex`
- [x] 6.2 Wrap the existing `<div className="flex">{uploadActions}</div>` in `GalleryToolbar` with `hidden sm:flex`

## 7. Focus-mode toggle hidden below the breakpoint

- [x] 7.1 Wrap the `Toggle` element passed into `GalleryToolbar`'s `focusToggle` slot (in `AppLayout`) with `hidden sm:flex`; leave `focusMode` state and its effect on `<main>` unchanged

## 8. Profile menu shows only Sign out below the breakpoint

- [x] 8.1 In `ProfileMenu.tsx`, hide the Settings `DropdownMenuItem` below `sm` (e.g. `hidden sm:flex` on that item, or conditional render gated by the same breakpoint convention used elsewhere) while keeping Sign out always rendered

## 9. Tests

- [x] 9.1 `FolderSidebar.test.tsx`: drawer opens when `mobileOpen` is true, closes via backdrop tap and via `onMobileClose`, off-canvas/visible classes are present as expected
- [x] 9.2 New `MobileTopBar.test.tsx`: renders hamburger + wordmark, hamburger click calls the open handler
- [x] 9.3 New `FloatingUploadButton.test.tsx`: renders, tapping it triggers the upload-open callback
- [x] 9.4 `RightPanel.test.tsx`: root element carries the `hidden sm:flex` classes
- [x] 9.5 `GalleryToolbar.test.tsx`: sort control and upload actions container carry `hidden sm:flex`
- [x] 9.6 `AppLayout.test.tsx`: focus toggle wrapper carries `hidden sm:flex`; main content container carries the responsive margin classes instead of the old unconditional one
- [x] 9.7 `ProfileMenu.test.tsx`: Settings item carries the mobile-hidden treatment; Sign out item does not

## 10. Verification

- [x] 10.1 Run `npm run build` and fix any issues that arise
- [x] 10.2 Run `npm run lint` and fix any issues that arise
