## 1. Background — new message handlers

- [x] 1.1 In `background/index.ts`, add a `trigger-snip` case to `runtime.onMessage`: query the active tab and run the same capture logic as `handleSnipCommand` (captureVisibleTab → send `snip-frame` to content script), failing silently if no content script is present
- [x] 1.2 In `background/index.ts`, add a `trigger-image-picker` case to `runtime.onMessage`: query the active tab and send `{ type: "open-image-picker" }` to the content script, failing silently if no content script is present

## 2. Main popup — footer

- [x] 2.1 In `App.tsx`, import `Scissors` and `LayoutGrid` from `lucide-react`
- [x] 2.2 In `App.tsx`, add `onSnip` and `onImagePicker` props to `LoggedIn`; implement both as handlers that call `browser.runtime.sendMessage` then `window.close()`
- [x] 2.3 In `App.tsx`, replace the footer's "Log out" button with two icon buttons: `Scissors` (title `"Snip"`) and `LayoutGrid` (title `"Pick images"`), wired to `onSnip` and `onImagePicker` respectively
- [x] 2.4 Remove the `onLogout` prop from `LoggedIn` and its wiring in `App` (logout is now handled from Settings)

## 3. Settings — logout row

- [x] 3.1 In `Settings.tsx`, add an `onLogout` prop
- [x] 3.2 In `Settings.tsx`, add a "Log out" row below the browse-images hotkey row: red text (`color: "#e53e3e"`), right-aligned, no border/background, wired to `onLogout`
- [x] 3.3 In `App.tsx`, pass `handleLogout` as `onLogout` to the `Settings` component

## 4. Lint and build

- [x] 4.1 Run `npm run lint` in `extensions/` and fix any issues
- [x] 4.2 Run `npm run build` in `extensions/` and fix any issues
