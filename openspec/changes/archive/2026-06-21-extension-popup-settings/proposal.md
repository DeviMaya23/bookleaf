## Why

The extension now has two save-behavior features users may want to configure — drag-to-save (toggle on/off) and the snip hotkey (remap from its default) — plus the existing dark mode preference. None of these are "things you check at a glance" like recent saves; they're standing preferences, and bolting them into the main popup (which is built around recent-saves browsing) would clutter it. A dedicated Settings view gives preferences a home without disrupting the main popup's existing layout.

## What Changes

- Add a gear icon to the popup header (top-right, alongside the existing "Open" button) that swaps the popup into a Settings view. A back arrow in the Settings view returns to the main popup.
- The Settings view is a local view-state toggle (`view: "main" | "settings"`) in `extensions/src/popup/App.tsx`, not a new route — WebExtension popups are a single fixed document, consistent with how the existing `authState` branching already works.
- Settings contains:
  - **Dark mode toggle** — duplicates the existing sun/moon toggle already in the main view's user row; both read/write the same `getDarkMode`/`setDarkMode` storage.
  - **Drag-to-save toggle** — new on/off preference, backed by a new `getDragEnabled`/`setDragEnabled` storage pair (mirroring the existing dark-mode storage pattern). When off, the content script does not render the drag-drop zone at all.
  - **Snip hotkey control** — shows the current shortcut. Clicking "Change" deep-links to the browser's native shortcut settings rather than capturing a key combination in-popup: `browser.commands.openShortcutSettings()` on Firefox, `browser.tabs.create({ url: "chrome://extensions/shortcuts" })` on Chrome. (In-popup capture via `browser.commands.update()` was attempted and dropped — on Firefox/macOS it silently mis-mapped the Meta key to `Ctrl` instead of `MacCtrl`/`Command`, leaving the command unbound until fixed manually in the browser's own shortcut UI.)
- **Modified**: the drag-and-drop save flow's drop-zone-visibility requirement gains a precondition — the drop zone SHALL NOT render when the drag-enabled setting is off, regardless of whether `dragstart` would otherwise resolve a `srcUrl`.

## Capabilities

### New Capabilities
- `extension-popup-settings`: the Settings view itself — entry point (gear icon), view-swap behavior, dark mode toggle (duplicate), drag-to-save toggle, and the platform-split snip hotkey remap control.

### Modified Capabilities
- `extension-drag-drop-save`: the "Drop zone visibility and positioning" requirement gains a new precondition (drag-enabled setting must be on) alongside the existing `srcUrl`-resolved condition.

## Impact

- `extensions/src/popup/App.tsx`: add `view` state, gear icon, Settings view component, dark mode toggle (reused), drag-to-save toggle, hotkey control.
- `extensions/src/lib/storage.ts`: add `getDragEnabled`/`setDragEnabled`.
- `extensions/src/content/index.ts`: drop-zone rendering checks `getDragEnabled` before rendering on `dragstart`.
- `extensions/src/background/index.ts` (or popup, if `commands.update` is called directly from there): Firefox-only hotkey remap call.
- `extensions/manifest.json`: no change to the `commands` entry itself — only how its shortcut is set/read.
- No backend changes. Extension-only.
