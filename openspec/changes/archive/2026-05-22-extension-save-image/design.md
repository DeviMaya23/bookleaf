## Context

The extension scaffold (`browser-extension-scaffold`) established the project structure, Kinde OAuth login, and token storage in `chrome.storage.local`. The background service worker (`src/background/index.ts`) is currently a stub. This change brings it to life with the save-image flow.

The existing Bookleaf backend upload sequence is:
1. `POST /images` — initiates upload, returns `{ upload_url, image_id }`
2. `PUT {upload_url}` — uploads file bytes directly to R2 via presigned URL
3. `POST /images/:id/complete` — finalises, triggers thumbnail generation and AI labelling

The extension will drive this same sequence from the background service worker, using the stored access token for authenticated calls to the Bookleaf API.

## Goals / Non-Goals

**Goals:**
- Register a "Save to Bookleaf" context menu item on `<img>` elements
- Fetch the image blob from the background SW (CORS bypassed via `host_permissions`)
- Drive the 3-step upload flow using the stored auth token
- Show a Chrome notification for success and failure
- Handle the unauthenticated case gracefully

**Non-Goals:**
- Folder selection — images always save to root
- Token refresh — expired token shows "Please log in first"
- Content script or page injection
- Firefox support (deferred, as per the scaffold design)
- Progress indication for large uploads

## Decisions

### D1: Fetch image in background SW, not via content script

The background SW with `host_permissions: ["<all_urls>"]` bypasses CORS for all cross-origin fetches — extension-level privilege, not web-page-level CORS enforcement. This covers the vast majority of publicly accessible images without needing a content script or `scripting` permission.

**Trade-off accepted**: Images behind Referer-based hotlink protection or login-gated cookies will fail to fetch. The failure surfaces as a "Save failed" notification. This is acceptable for the MVP.

**Alternative rejected**: `chrome.scripting.executeScript()` to fetch in page context — would handle hotlink protection but adds the `scripting` permission, tab injection complexity, and ArrayBuffer serialisation overhead. Deferred as a future enhancement.

### D2: Page title and URL sourced from context menu callback

`chrome.contextMenus.onClicked` provides `info.srcUrl` (the image src) and `info.pageUrl`. The page title is retrieved via `chrome.tabs.get(tab.id).title`. No messaging to the tab is needed.

### D3: Auth token read from `chrome.storage.local` in the background SW

`getAuth()` from `src/lib/storage.ts` is already usable in the service worker context (it uses `webextension-polyfill` which wraps `chrome.storage.local`). If `getAuth()` returns null or the token is expired, the flow aborts early with a "Please log in first" notification.

Token expiry is checked against `expiresAt` — a simple `Date.now() > auth.expiresAt` guard. No refresh is attempted.

### D4: API client in background SW

The existing Bookleaf backend validates Kinde JWTs. The SW will call the API directly via `fetch()` with `Authorization: Bearer <accessToken>`. A thin `apiFetch` helper (similar to the frontend's `src/lib/api.ts`) will be added to the extension's `src/lib/` so it can be shared across popup and background.

### D5: Upload flow error handling

Any failure in the 3-step sequence (network error, non-2xx response, blob fetch failure) shows a "Save failed" notification. No retry logic. The partial upload state on the backend (an initiated but never completed image) is cleaned up by the existing stale upload purge job (`stale-upload-cleanup` spec).

## Risks / Trade-offs

- **Hotlink-protected images fail silently** → User sees "Save failed" with no explanation of why. Acceptable for MVP; future versions can add the scripting fallback.
- **Large images are slow** → The SW fetches the full blob before uploading. No progress indicator. For typical web images (< 5 MB) this is imperceptible; for large originals it may feel slow. Acceptable for now.
- **Token expiry UX** → User gets "Please log in first" mid-session if the token expired. They must open the popup and log in again. No auto-refresh. Acceptable for the scaffold stage.
- **Service worker lifecycle** → MV3 service workers can be killed by the browser mid-flight on slow uploads. Chrome extends SW lifetime during active `fetch()` calls, so in practice the 3-step sequence should complete. Edge case for very slow networks.
