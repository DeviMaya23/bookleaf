## 1. Background Script — Replace chrome.* with browser.*

- [x] 1.1 Add `import browser from "webextension-polyfill"` to `background/index.ts`
- [x] 1.2 Convert `onInstalled` listener to `async`; replace `chrome.contextMenus.removeAll(callback)` with `await browser.contextMenus.removeAll()`
- [x] 1.3 Replace `chrome.contextMenus.create(...)` with `browser.contextMenus.create(...)`
- [x] 1.4 Replace `chrome.contextMenus.onClicked.addListener(...)` with `browser.contextMenus.onClicked.addListener(...)`
- [x] 1.5 Replace `chrome.notifications.create(...)` with `browser.notifications.create(...)` in the `notify` function
- [x] 1.6 Replace `chrome.runtime.getURL(...)` with `browser.runtime.getURL(...)` inside `notify`

## 2. Auth — Replace chrome.identity.* with browser.identity.*

- [x] 2.1 Add `import browser from "webextension-polyfill"` to `lib/auth.ts`
- [x] 2.2 Replace `chrome.identity.getRedirectURL()` with `browser.identity.getRedirectURL()` in `getRedirectUri`
- [x] 2.3 Replace `await chrome.identity.launchWebAuthFlow(...)` with `await browser.identity.launchWebAuthFlow(...)` in `login`

## 3. Background Script — OffscreenCanvas Guard

- [x] 3.1 Wrap the `generateThumbnail` call in `handleSave` with `if (typeof OffscreenCanvas !== "undefined")`
- [x] 3.2 In the `else` branch, call `addRecentSave({ imageId, title, dataUrl: "", savedAt: Date.now() })` directly

## 4. Popup — Placeholder for Empty dataUrl

- [x] 4.1 Update the thumbnail strip render loop in `App.tsx`: if `save.dataUrl` is an empty string, render a `<div>` placeholder box instead of `<img>`
- [x] 4.2 Apply placeholder styles: `flex: 1`, `aspectRatio: "1"`, `borderRadius: 7`, `background: c.divider`, matching the `<img>` layout

## 5. Firefox Build — Gecko ID Manifest Override

- [x] 5.1 Update `vite.config.ts` to pass a `manifest` override to `vite-plugin-web-extension` when `mode === "firefox"`, merging `browser_specific_settings: { gecko: { id: "bookleaf@evimay.me" } }` into the base manifest
- [x] 5.2 Verify `npm run build` produces a Chrome manifest without `browser_specific_settings`
- [x] 5.3 Verify `npm run build:firefox` produces a Firefox manifest with `browser_specific_settings.gecko.id` set to `"bookleaf@evimay.me"`

## 6. Manual Verification

- [x] 6.1 Load the Chrome build as an unpacked extension; confirm login, dark mode toggle, right-click save, and recent saves all work
- [x] 6.2 Load the Firefox build via `about:debugging`; confirm login, dark mode toggle, right-click save, and recent saves (with placeholder) all work
- [x] 6.3 Add the Firefox redirect URI (`https://bookleaf@evimay.me.extensions.allizom.org/`) to Kinde's allowed redirect URI list before testing Firefox login
