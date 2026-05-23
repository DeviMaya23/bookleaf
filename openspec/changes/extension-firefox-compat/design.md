## Context

The extension ships with `webextension-polyfill` declared as a dependency, and `storage.ts` already uses it correctly. However, `background/index.ts` and `lib/auth.ts` call `chrome.*` directly, bypassing the polyfill entirely. The `npm run build:firefox` script exists but produces a broken output because these raw `chrome.*` calls fail in Firefox's MV3 background.

The two highest-risk areas are the identity API (OAuth login) and thumbnail generation (`OffscreenCanvas`), which have no direct Firefox equivalent in a background service worker context.

## Goals / Non-Goals

**Goals:**
- Single codebase that produces working Chrome and Firefox extension builds
- All existing features work identically on both browsers: login/logout, dark/light mode, recent saves display, right-click save
- Firefox build declares a stable gecko extension ID so `browser.identity` redirect URIs are deterministic

**Non-Goals:**
- Proper thumbnail generation on Firefox (deferred to a follow-up — requires a content script canvas approach)
- Firefox Add-on signing or AMO publication
- Supporting any browser other than Chrome and Firefox

## Decisions

### 1. Use `browser.*` throughout via webextension-polyfill (no browser detection)

Replace every `chrome.*` call in `background/index.ts` and `lib/auth.ts` with the `browser.*` equivalent imported from `webextension-polyfill`. The polyfill normalises Chrome's callback-based APIs to Promises and is a no-op on Firefox (where `browser.*` is already Promise-based). This avoids any `if (chrome) / else if (browser)` branching and keeps the code uniform.

**Alternative considered**: Detect the browser at runtime (`typeof chrome !== "undefined"`) and branch. Rejected — adds dead code paths and defeats the purpose of having the polyfill.

### 2. `contextMenus.removeAll` — convert callback to async/await

`chrome.contextMenus.removeAll` is currently called with a callback. `browser.contextMenus.removeAll` returns a Promise. The `onInstalled` handler becomes an `async` function with `await browser.contextMenus.removeAll()` followed by `browser.contextMenus.create(...)`.

### 3. OffscreenCanvas — capability guard, placeholder `dataUrl`

Firefox MV3 background scripts do not support `OffscreenCanvas` or `canvas.convertToBlob`. Wrapping the thumbnail generation block in `if (typeof OffscreenCanvas !== "undefined")` confines it to Chrome. On Firefox, `handleSave` calls `addRecentSave` with `dataUrl: ""` so the save is still recorded; the popup renders a themed placeholder box instead of an `<img>`.

**Alternative considered**: Offload canvas drawing to a content script via `chrome.tabs.sendMessage`. Rejected for this iteration — adds a content script file, messaging complexity, and the tab may already be closed by the time the background processes the save.

### 4. Gecko ID — hardcoded in `vite.config.ts`, Firefox mode only

`browser.identity.launchWebAuthFlow` on Firefox requires a stable extension ID. Without it, Firefox generates a random ID on each install, making the OAuth redirect URI non-deterministic.

The ID `bookleaf@evimay.me` is injected into the manifest only during the Firefox build via a manifest override in `vite.config.ts`. The `manifest.json` source file is not changed. `vite-plugin-web-extension` accepts a `manifest` override option for this purpose.

**Alternative considered**: A separate `manifest.firefox.json`. Rejected — the plugin's override merging keeps the single-manifest contract and avoids divergence.

### 5. Kinde redirect URIs — manual registration step

Each browser produces a different redirect URI format:
- Chrome: `https://<extension-id>.chromiumapp.org/`
- Firefox: `https://bookleaf@evimay.me.extensions.allizom.org/`

Both must be added to Kinde's allowed redirect URI list. This is a one-time ops step, not a code change.

## Risks / Trade-offs

- **Firefox saves have no thumbnails** → The popup renders a placeholder box. Users won't see thumbnails for images saved on Firefox until the OffscreenCanvas follow-up is shipped. The data model is preserved (`dataUrl: ""`); no migration needed when thumbnails arrive.
- **Gecko ID is hardcoded, not tied to an AMO account** → Acceptable for development and side-loading. Must be replaced with the official ID before submitting to AMO; the AMO account associates the ID at submission time, not before.
- **Kinde redirect URI registration is a manual step** → Firefox login will silently fail until the Firefox redirect URI is added to Kinde. Documented as a required deploy step.
- **`browser.identity` on Firefox requires the extension to be signed or run in a permissive profile** → For development, `xpinstall.signatures.required` must be `false` in Firefox, or the extension must be temporarily installed via `about:debugging`. This is standard for extension development.
