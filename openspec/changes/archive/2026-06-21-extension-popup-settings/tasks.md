## 1. Storage

- [x] 1.1 Add `getDragEnabled`/`setDragEnabled` to `extensions/src/lib/storage.ts`, mirroring the existing `getDarkMode`/`setDarkMode` pattern, defaulting to `true` when no value has ever been stored
- [x] 1.2 Unit test `getDragEnabled`/`setDragEnabled`: returns `true` when unset, returns the last-set value after `setDragEnabled(false)`/`setDragEnabled(true)`

## 2. Content script: gate drop zone on drag-enabled setting

- [x] 2.1 In `extensions/src/content/index.ts`, read `getDragEnabled()` at `dragstart` (fresh read, not cached) and skip drop-zone rendering entirely when it resolves `false`, regardless of whether `srcUrl` would otherwise resolve
- [x] 2.2 Unit test the `dragstart` handler: drop zone is not rendered when drag-enabled is `false` even with a resolvable `srcUrl`; existing resolvable/unresolvable `srcUrl` behavior is unchanged when drag-enabled is `true`

## 3. Popup: Settings view scaffolding

- [x] 3.1 In `extensions/src/popup/App.tsx`, add `view: "main" | "settings"` state, defaulting to `"main"` on every popup open
- [x] 3.2 Add a gear icon to the popup header (next to the existing "Open" button), wired to set `view` to `"settings"`
- [x] 3.3 Create a `Settings` component rendered when `view === "settings"`, with a back arrow that sets `view` back to `"main"`
- [x] 3.4 Verify entering/leaving Settings does not refetch auth, username, avatar, or recent saves (no re-run of the existing `useEffect` data-loading call)

## 4. Popup: dark mode toggle in Settings

- [x] 4.1 Add a dark mode toggle row to `Settings`, reusing the same `isDark`/`onToggleDark` state and handler already lifted in `App.tsx` (passed as props to both `LoggedIn` and `Settings`)
- [x] 4.2 Manually verify toggling from either location (main view user row, Settings) updates the other's visual state without reopening the popup

## 5. Popup: drag-to-save toggle in Settings

- [x] 5.1 Add a drag-to-save on/off toggle row to `Settings`, reading `getDragEnabled` on mount and calling `setDragEnabled` on change
- [x] 5.2 Manually verify: toggling off in Settings, then dragging an image on a regular webpage, does not show the drop zone; toggling back on restores it

## 6. Popup: snip hotkey display and remap

- [x] 6.1 Add a row to `Settings` displaying the snip command's current shortcut, read via `browser.commands.getAll()`
- [x] 6.2 Identify and reuse the existing Firefox-vs-Chrome detection mechanism already used by `extension-firefox-compat` (do not introduce a new detection mechanism without checking for an existing one first)
- [x] 6.3 On Firefox: implement a "Change" control that calls `browser.commands.openShortcutSettings()` (in-popup capture via `browser.commands.update()` was attempted and dropped — see design.md Decision 4 for why)
- [x] 6.4 On Chrome: implement a "Change" control labeled to indicate it opens browser settings, calling `browser.tabs.create({ url: "chrome://extensions/shortcuts" })`
- [x] 6.5 Unit test the Firefox path: clicking "Change" calls `browser.commands.openShortcutSettings()` and never calls `browser.commands.update`
- [x] 6.6 Unit test the Chrome path: clicking "Change" calls `browser.tabs.create` with `chrome://extensions/shortcuts` and does not attempt `browser.commands.update`

## 7. Verification

- [x] 7.1 Run `npm run build` in `extensions/` and fix any errors
- [ ] 7.2 Run `npm run lint` in `extensions/` and fix any issues (skipped as per usual with extensions)
- [x] 7.3 Manually verify the full Settings flow on Firefox: open Settings, toggle dark mode, toggle drag-to-save, click "Change" on the hotkey row and confirm it opens the browser's shortcut settings
- [x] 7.4 Manually verify the full Settings flow on Chrome: open Settings, toggle dark mode, toggle drag-to-save, click "Change" on the hotkey row and confirm it opens `chrome://extensions/shortcuts`
- [x] 7.5 Manually verify existing right-click, drag-drop (when enabled), and snip save flows still work unchanged after this change
