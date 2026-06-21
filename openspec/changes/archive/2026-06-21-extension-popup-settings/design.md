## Context

The popup (`extensions/src/popup/App.tsx`) currently has a single conditional branch — `authState: "loading" | "unauthenticated" | "authenticated"` — that decides which of `LoggedOut`/`LoggedIn` to render. There's no existing notion of "multiple views" within the logged-in popup; everything lives in one component tree.

Two new preferences need a home: a drag-to-save on/off toggle (entirely new state, no prior storage key) and a snip-hotkey control. Neither browser's `commands` API supports a reliable in-popup remap: `browser.commands.update({ name, shortcut })` exists on Firefox, but during implementation it proved unsafe to drive from a captured `KeyboardEvent` — on macOS, `event.metaKey` maps to the Cmd key, but Firefox's shortcut grammar expects `MacCtrl`/`Command` rather than `Ctrl`, so a naively-formatted shortcut string was silently accepted and left the command unbound (no error, no working hotkey, old binding also cleared) until reset via the browser's own shortcut UI. `chrome.commands` has no update API at all. Given that, both platforms now deep-link to the browser's native shortcut settings instead. This mirrors the kind of platform divergence `extension-firefox-compat` already tracks elsewhere in this codebase.

## Goals / Non-Goals

**Goals:**
- Give standing preferences (dark mode, drag-to-save, hotkey) a dedicated Settings view, reachable and dismissable without leaving the popup.
- Make drag-to-save a real on/off switch — when off, the drop zone must not render at all, not just be inert.
- Let the user discover and get to the hotkey setting from the popup, even though the actual remap happens in the browser's own UI on both platforms.

**Non-Goals:**
- No new popup "page"/route — this stays a single-document view-state swap, matching the existing `authState` branching pattern.
- No in-popup key-combination capture on either platform — attempted on Firefox via `browser.commands.update()`, but dropped after it proved unsafe (see Context). Both platforms now redirect to native browser UI for the actual remap.
- No attempt to polyfill Chrome's missing `commands.update` (e.g. via a content script trick or workaround) — out of scope, and not something a content script can grant anyway since this is a privileged browser-UI action.
- No persistence of "settings view was last open" — the popup always opens to the main view; Settings is always entered fresh via the gear icon.

## Decisions

### 1. Settings is local view state in `App.tsx`, not a new component-level route

Add `const [view, setView] = useState<"main" | "settings">("main")` alongside the existing `authState`. `LoggedIn` renders either its current content or a new `Settings` component based on `view`, toggled by the gear icon (→ `"settings"`) and the Settings view's back arrow (→ `"main"`).

Alternative considered: a separate popup HTML page navigated via `window.location`. Rejected — WebExtension popups are a single fixed-size document; navigating away loses the loaded auth/recent-saves state for no benefit over a local toggle.

### 2. New storage key: `getDragEnabled` / `setDragEnabled`, default `true`

Added to `extensions/src/lib/storage.ts` following the exact shape of `getDarkMode`/`setDarkMode`. Default is `true` so existing users see no behavior change until they explicitly opt out via Settings.

### 3. Content script checks `getDragEnabled` at `dragstart`, not at content-script load time

The drop-zone-rendering logic in `extensions/src/content/index.ts` already gates on a resolved `srcUrl` at `dragstart`. The drag-enabled check is added as an additional gate at the same point (read fresh on each `dragstart` via `browser.storage.local`), rather than cached at script-load time — so a setting change takes effect on the very next drag gesture without requiring a page reload.

Alternative considered: cache the flag once on content-script injection. Rejected — would require either a page reload or a runtime message round-trip from the popup to invalidate the cache, adding complexity for no real benefit when reading storage at `dragstart` is cheap and already async-tolerant (the existing `dragstart` handler already does async resolution work for `srcUrl`/title).

### 4. Hotkey control deep-links to native browser settings on both platforms

Since `browser.commands` is available in any extension context (not background-worker-exclusive), the Settings view calls `browser.commands.openShortcutSettings()` directly on Firefox, and `browser.tabs.create({ url: "chrome://extensions/shortcuts" })` directly on Chrome — no background-worker round-trip needed. Browser detection reuses the build-time `import.meta.env.MODE` check already established in `extensions/src/background/index.ts` (`mode === "firefox" || mode === "firefox-production"`), which mirrors `extension-firefox-compat`'s existing pattern.

Originally this used `browser.commands.update({ name, shortcut })` with an in-popup key-combination capture on Firefox. That was implemented, tested manually, and dropped: capturing a `KeyboardEvent` and formatting it into a shortcut string (e.g. mapping `event.metaKey` to `"Command"`) produced strings Firefox's shortcut grammar didn't reliably accept — on macOS specifically, the update call appeared to succeed but left the command unbound, breaking both the new and the previously-working default shortcut, recoverable only by resetting it from the browser's own "Manage Extension Shortcuts" UI. Given the API's shortcut-string format isn't something this implementation can safely construct from raw key events, the control now does only navigation, not remapping.

### 5. Dark mode toggle is duplicated, not moved or extracted into a shared sub-component

The existing sun/moon toggle in `LoggedIn`'s user row stays exactly as-is. The Settings view's dark mode row is a second, independent toggle control wired to the same `getDarkMode`/`setDarkMode` calls and the same `isDark` state already lifted in `App.tsx` (passed down as a prop to both `LoggedIn` and the new `Settings` component) — not a new shared component, since it's two small toggle buttons reading the same boolean, not enough shared structure to justify extraction.

## Risks / Trade-offs

- [Risk] Users may expect clicking the hotkey control to let them type a new shortcut in-popup, since that's a common pattern in other extensions. → Mitigation: both platforms navigate to the same native browser shortcut settings on click, so behavior (and the lack of an in-popup capture affordance) is consistent across Firefox and Chrome.
- [Risk] Reading `getDragEnabled` from storage on every `dragstart` adds one more async storage read to an already-async resolution path. → Mitigation: negligible in practice — `chrome.storage.local` reads are fast and the existing `dragstart` handler already awaits multiple async resolvers before showing the drop zone.
- [Trade-off] Duplicating the dark-mode toggle means two pieces of UI can drift out of sync if one is updated without the other in a future change. → Mitigation: both call the same `setDarkMode`/read the same `isDark` prop from `App.tsx`'s single source of truth, so there's no separate state to drift — only the rendering markup is duplicated, not the logic.

## Open Questions

- Exact Firefox/Chrome detection mechanism to reuse for the hotkey control — to be confirmed against whatever `extension-firefox-compat` already uses, during implementation (not expected to require a new mechanism).
