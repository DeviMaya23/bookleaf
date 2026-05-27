## Why

The extension is tested and working on Chrome but cannot run on Firefox despite having `webextension-polyfill` installed — `chrome.*` API calls are used directly in `background/index.ts` and `lib/auth.ts`, and the Firefox build (`npm run build:firefox`) produces a broken extension. Making Firefox a first-class supported target expands the extension's reach without any user-facing behavior changes.

## What Changes

- **`background/index.ts`**: Replace all `chrome.*` calls with `browser.*` via webextension-polyfill; convert `contextMenus.removeAll` from callback to async/await; guard `OffscreenCanvas` usage so thumbnail generation is skipped on Firefox (stores empty `dataUrl` instead of crashing).
- **`lib/auth.ts`**: Replace `chrome.identity.getRedirectURL()` and `chrome.identity.launchWebAuthFlow()` with their `browser.identity.*` equivalents via webextension-polyfill.
- **`popup/App.tsx`**: Render a themed placeholder box for recent-save entries where `dataUrl` is an empty string (Firefox saves until a proper OffscreenCanvas alternative is implemented).
- **`vite.config.ts`**: Inject `browser_specific_settings.gecko` into the manifest when building in `firefox` mode, providing a stable extension ID required by Firefox's identity API.

## Capabilities

### New Capabilities

- `extension-firefox-compat`: Firefox-specific build requirements — gecko extension ID, manifest differences, and OffscreenCanvas fallback behavior for the save flow.

### Modified Capabilities

- `extension-auth`: Auth flow switches from `chrome.identity.*` to `browser.identity.*`; the Firefox redirect URI format differs from Chrome's and both must be registered with the OAuth provider.
- `extension-recent-saves`: Recent saves entries may have an empty `dataUrl` on Firefox (OffscreenCanvas unavailable); the popup must render a placeholder rather than a broken image.

## Impact

- **`extensions/src/background/index.ts`**: All `chrome.*` calls replaced; thumbnail generation made conditional.
- **`extensions/src/lib/auth.ts`**: `chrome.identity.*` replaced with `browser.identity.*`.
- **`extensions/src/popup/App.tsx`**: Thumbnail render loop updated for empty `dataUrl` fallback.
- **`extensions/vite.config.ts`**: Manifest override injected for Firefox mode.
- **External**: Kinde OAuth app must have the Firefox redirect URI (`https://<gecko-id>.extensions.allizom.org/`) added to its allowed list alongside the existing Chrome redirect URI.
- **No changes** to `lib/storage.ts`, `lib/api.ts`, build scripts, or any backend code.
