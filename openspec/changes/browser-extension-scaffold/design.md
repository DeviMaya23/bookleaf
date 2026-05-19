## Context

Bookleaf's web app uses Kinde for auth via the `@kinde-oss/kinde-auth-react` SDK, which relies on browser redirects and React Router. Browser extensions cannot use the same flow — they have no `window.location` redirect mechanism in the popup or service worker, and the React SDK is not suitable for an extension context.

The extension is a new codebase living in `/extensions`, fully independent of the existing `frontend/` and `backend/` directories. It targets Manifest V3 (MV3), which requires a background **service worker** (not a persistent background page) and restricts certain APIs.

The existing backend validates Kinde-issued JWTs via a JWKS endpoint. The extension just needs to obtain a valid Kinde access token — the backend doesn't care how it was issued.

## Goals / Non-Goals

**Goals:**
- Scaffold a Vite + TypeScript MV3 extension project under `/extensions`
- Cross-browser support for Chrome and Firefox via `webextension-polyfill`
- Implement Kinde OAuth login using `chrome.identity.launchWebAuthFlow` with PKCE
- Persist the access token (and refresh token) in `chrome.storage.local`
- Show login/logged-in state in the extension popup

**Non-Goals:**
- Image saving or any feature beyond authentication
- Firefox-specific `identity` API differences (treat as stretch goal)
- Token refresh automation (out of scope for scaffold; logged-out state on expiry is acceptable)
- Content scripts for interacting with web pages

## Decisions

### D1: Build tooling — `vite-plugin-web-extension`

Use [`vite-plugin-web-extension`](https://vite-plugin-web-extension.aklinker1.io/) as the Vite plugin for the extension build. It handles multi-entry builds (popup, background service worker, content scripts) from a single manifest, generates the final `manifest.json`, and supports both Chrome and Firefox targets via a `browser` flag in the build command.

**Alternative considered**: `@crxjs/vite-plugin` — more popular but less actively maintained and has had MV3 compatibility issues with Firefox. `vite-plugin-web-extension` is the current recommended choice for cross-browser MV3 projects.

### D2: Kinde OAuth via PKCE + `chrome.identity.launchWebAuthFlow`

Browser extensions cannot use the Kinde React SDK's redirect-based flow. Instead:

1. The popup generates a `code_verifier` and `code_challenge` (S256 PKCE)
2. It calls `chrome.identity.launchWebAuthFlow({ url: kindeAuthUrl, interactive: true })` — this opens Kinde's hosted login in a browser-managed window
3. Kinde redirects to the extension's redirect URI (`https://<ext-id>.chromiumapp.org/callback`) with an authorization code
4. The popup exchanges the code + `code_verifier` for tokens via Kinde's token endpoint (direct `fetch` to `VITE_KINDE_ISSUER_URL/oauth2/token`)
5. Access token and refresh token are stored in `chrome.storage.local`

The Kinde application in the Kinde dashboard must be configured as a **SPA** client (no client secret) and the extension's redirect URI must be added to the allowed list.

**Why PKCE**: Extensions are public clients — there's no safe place to store a client secret. PKCE is the correct OAuth 2.0 flow for public clients.

### D3: Token storage in `chrome.storage.local`

`chrome.storage.local` is the correct storage mechanism for extensions:
- Available to both the popup and the service worker (unlike `localStorage`, which is not available in service workers)
- Persists across popup open/close cycles
- Automatically cleared when the extension is uninstalled

The token will be stored under a single key, e.g., `bookleaf_auth`, as a JSON object `{ accessToken, refreshToken, expiresAt }`.

### D4: Popup as the primary UI entry point

The popup (`popup.html` + `popup.ts` with a minimal React component tree) handles:
- Checking `chrome.storage.local` for an existing token
- Showing a **Login** button if no token exists, triggering the OAuth flow
- Showing a **Logged in** state with a logout button if a token is present

No content scripts or options page are included in this scaffold.

### D5: Project structure

```
/extensions
├── manifest.json          # Source manifest (vite-plugin-web-extension reads this)
├── package.json
├── vite.config.ts
├── tsconfig.json
├── .env.example           # VITE_KINDE_CLIENT_ID, VITE_KINDE_ISSUER_URL
└── src/
    ├── popup/
    │   ├── index.html
    │   ├── main.tsx
    │   └── App.tsx        # Login / logged-in UI
    ├── background/
    │   └── index.ts       # Service worker (stub for now)
    └── lib/
        ├── auth.ts        # PKCE helpers + launchWebAuthFlow + token exchange
        └── storage.ts     # chrome.storage.local typed wrappers
```

## Risks / Trade-offs

- **Firefox `identity` API**: Firefox supports `browser.identity.launchWebAuthFlow` but requires a different redirect URI format. This is deferred — the scaffold targets Chrome first, Firefox support is wired via polyfill but not explicitly tested in this proposal.
- **PKCE code verifier in popup**: The code verifier is generated in the popup context. If the popup closes before the OAuth flow completes, the verifier is lost and the code exchange will fail. This is acceptable for the scaffold; a production implementation would persist the verifier temporarily in `sessionStorage` or `chrome.storage.session`.
- **Token expiry**: No auto-refresh is implemented. When the access token expires, the user will see an unauthenticated state and must log in again. Acceptable for the scaffold.
- **Kinde redirect URI registration**: The extension ID changes in dev vs. production (and between browsers). The Kinde application must have the correct redirect URI added. This is a manual setup step documented in the project README.

## Open Questions

- None. Decision made: the extension will use a dedicated Kinde application (separate Client ID) to keep redirect URI config and auth config isolated from the web frontend.
