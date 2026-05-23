## Context

The extension popup (`extensions/src/popup/App.tsx`) is currently a minimal auth gate with inline styles — no branding, no user context, no history. The background worker (`background/index.ts`) already fetches image blobs and uploads them via a 3-step flow, firing native Chrome notifications on success/failure.

Two new concerns are introduced: (1) a polished UI with dark mode, and (2) a "recently saved" strip that shows local thumbnails without any backend calls.

`storage.ts` uses `webextension-polyfill`; `background/index.ts` uses raw `chrome.*` APIs. Both patterns coexist and that is fine — we continue using `browser.*` in storage helpers and `chrome.*` in the background worker.

## Goals / Non-Goals

**Goals:**
- Replace the bare popup with a designed UI matching the Claude Design spec
- Store dark mode preference and username persistently in `chrome.storage.local`
- Generate thumbnails for recently saved images entirely in the background service worker with no backend calls from the popup
- Keep the popup fast to open: all data read from local storage, no async API calls on mount

**Non-Goals:**
- Toast / in-page save notification (separate proposal)
- Syncing recently saved list with the backend
- Pagination or full history view in the popup
- Custom SVG logo (use existing `icons/icon48.png`)

## Decisions

### D1: Username and avatar from id_token, not userinfo endpoint

**Decision:** The Kinde token endpoint returns an `id_token` alongside the `access_token` in an OIDC flow with `openid profile` scope. After token exchange, capture and base64url-decode the `id_token` payload to extract `given_name` (falling back to `name`, then `email`) and `picture`. Store both in `chrome.storage.local` (`bookleaf_username`, `bookleaf_avatar`).

**Rationale:** The `id_token` is the canonical OIDC source for user profile claims — `picture` is guaranteed to appear here when the `profile` scope is granted. The access token does not reliably carry `picture`. The FE uses `getUserProfile()` from the Kinde React SDK which reads equivalent data; the extension replicates the same outcome by decoding the `id_token` directly. No extra network call is needed.

**Alternative considered:** Decode the access token payload instead. Kinde access tokens do not reliably include `picture`; the `id_token` is the correct token for profile claims.

**Alternative considered:** Call `GET /userinfo` with the access token. Adds a network call and a failure mode at login time for purely cosmetic data.

### D2: Thumbnail generated in background worker via OffscreenCanvas

**Decision:** After `/complete` succeeds, generate a 60×60 JPEG thumbnail from the already-fetched blob using `createImageBitmap` + `OffscreenCanvas` in the service worker. Convert to base64 via chunked `btoa` (avoid stack overflow). Do not block the save result on thumbnail generation — wrap in try/catch so a thumbnail failure is silent.

**Rationale:** We already hold the blob in memory at save time. Reusing it costs zero extra network calls. `OffscreenCanvas` and `createImageBitmap` are both available in Chrome service workers (Chrome 69+).

**Alternative considered:** Fetch `thumbnail_url` from `GET /images/:id` after `/complete`. Requires an extra API call, and the presigned URL expires after 24h, making stored thumbnails stale overnight.

**Alternative considered:** Store `srcUrl` and render with `<img src>`. Unreliable — URLs may require auth on the source site, change, or be blocked by CORS when rendered in the popup.

### D3: `recentSaves` array capped at 5, FIFO in storage helper

**Decision:** `addRecentSave` prepends the new entry and slices to `MAX_RECENT_SAVES = 5` before writing. No secondary cleanup pass needed.

**Rationale:** Simple and atomic. The popup always reads the exact slice it renders — no filtering required.

### D4: Dark mode as a standalone storage key

**Decision:** Store dark mode preference as `bookleaf_dark_mode: boolean` separately from auth. The popup reads it on mount alongside `recentSaves`.

**Rationale:** Dark mode preference should survive logout/login. Coupling it to the auth object would clear it on `clearAuth()`.

### D5: Popup reads all local state on mount, no subscriptions

**Decision:** On mount, the popup reads `auth`, `username`, `recentSaves`, and `darkMode` in a single `Promise.all`. No `chrome.storage.onChanged` subscription.

**Rationale:** Extension popups are short-lived — they open, the user glances, they close. Real-time updates are not needed. A single read on open is simpler and sufficient.

## Risks / Trade-offs

- **`OffscreenCanvas` availability**: Available in Chrome 69+ service workers. Extensions already require a modern Chrome version; this is not a concern in practice. → No mitigation needed.
- **Chunked btoa for thumbnail blob**: `String.fromCharCode(...new Uint8Array(buf))` will stack overflow for large buffers. The 60×60 JPEG output is ~2–4KB, well within safe limits, but the implementation should chunk regardless for correctness. → Use a chunk size of 8192 bytes in the base64 conversion loop.
- **Thumbnail generation failure**: If `createImageBitmap` or `OffscreenCanvas` throws (malformed image, unsupported format), the save itself must not be affected. → Wrap thumbnail generation in its own try/catch; log and continue.
- **Storage quota**: `chrome.storage.local` has a 10MB default quota (unlimited with `unlimitedStorage` permission). Five 60×60 JPEGs at ~4KB each = ~20KB. Negligible. → No mitigation needed.

### D6: Use webextension-polyfill for all new extension code

**Decision:** All new code written in this change SHALL use `browser.*` from `webextension-polyfill` rather than raw `chrome.*` APIs, except for `chrome.identity` which has no Firefox equivalent (Firefox auth is a separate proposal).

**Rationale:** The scaffold intentionally included `webextension-polyfill` and a `build:firefox` script. Existing `background/index.ts` and `auth.ts` drifted to `chrome.*` — this proposal does not clean that up, but all new code should honour the original intent.

**In practice for this change:** The one new browser API call is `browser.tabs.create` (the "Open ↗" button). Storage helpers already use `browser.*` via the polyfill.

## Open Questions

_(none — all decisions resolved during exploration)_
