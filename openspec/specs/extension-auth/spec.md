# Spec: Extension Auth

## Purpose

Defines the authentication requirements for the Bookleaf browser extension, covering popup authentication state display, Kinde OAuth PKCE login flow, logout, and token storage schema.

## Requirements

### Requirement: Popup displays authentication state

The extension popup SHALL display one of two states based on whether a valid token exists in `chrome.storage.local`:
- **Unauthenticated**: A "Login with Bookleaf" button is shown
- **Authenticated**: A "Logged in" message and a "Logout" button are shown

The popup SHALL check `chrome.storage.local` on mount and render the appropriate state without a flash of incorrect content.

#### Scenario: Unauthenticated user sees login button

- **WHEN** the popup is opened and no token exists in `chrome.storage.local`
- **THEN** the popup renders a "Login with Bookleaf" button
- **AND** no logged-in content is shown

#### Scenario: Authenticated user sees logged-in state

- **WHEN** the popup is opened and a valid token exists in `chrome.storage.local`
- **THEN** the popup renders a "Logged in" indicator and a "Logout" button
- **AND** the login button is not shown

### Requirement: Kinde OAuth login via PKCE

When the user clicks "Login with Bookleaf", the popup SHALL send a `{ type: "start-login" }` message to the background script. The background script SHALL perform the full OAuth flow so that the auth window opening does not kill the popup's JavaScript context:

1. Generate a cryptographically random `code_verifier` (43–128 characters, URL-safe)
2. Derive `code_challenge` as the Base64url-encoded SHA-256 hash of the `code_verifier`
3. Generate a cryptographically random `state` value (at least 16 URL-safe characters) for CSRF protection
4. Construct the Kinde authorization URL with `response_type=code`, `client_id`, `redirect_uri`, `scope=openid profile email`, `code_challenge`, `code_challenge_method=S256`, `state`, and `audience` (from `VITE_KINDE_AUDIENCE`)
5. Call `browser.identity.launchWebAuthFlow({ url: authUrl, interactive: true })` (via `webextension-polyfill`) from the background script
6. Extract the `code` parameter from the redirect URI returned by the flow
7. Exchange the code for tokens via a `POST` to `{VITE_KINDE_ISSUER_URL}/oauth2/token` with the `code`, `code_verifier`, `client_id`, `redirect_uri`, and `grant_type=authorization_code`
8. Store the resulting `access_token`, optional `refresh_token`, and computed `expires_at` in extension storage under key `bookleaf_auth`
9. The popup SHALL listen for `browser.storage.onChanged` on `bookleaf_auth` and transition to the authenticated state when a valid token is written

The redirect URI SHALL be obtained via `browser.identity.getRedirectURL()` (via `webextension-polyfill`). The resulting URL format differs per browser:
- Chrome: `https://<extension-id>.chromiumapp.org/`
- Firefox: `https://bookleaf@evimay.me.extensions.allizom.org/`

Both redirect URI formats SHALL be registered in Kinde's allowed redirect URI list. This is a required deployment step — Firefox login will fail until the Firefox redirect URI is registered.

The `offline_access` scope SHALL NOT be requested. Kinde does not permit it on SPA clients by default. The `refresh_token` field in storage is therefore optional.

#### Scenario: Successful login stores token and updates UI

- **WHEN** the user clicks "Login with Bookleaf" and completes the Kinde login flow
- **THEN** the access token is stored in extension storage under `bookleaf_auth`
- **AND** the popup updates to show the authenticated state

#### Scenario: User cancels the login flow

- **WHEN** the user closes the Kinde login window without completing login
- **THEN** `browser.identity.launchWebAuthFlow` returns an error or empty URL
- **AND** the popup remains in the unauthenticated state
- **AND** no error is shown to the user (silent failure is acceptable)

#### Scenario: Token exchange fails

- **WHEN** the Kinde token endpoint returns an error during code exchange
- **THEN** the popup displays a brief error message (e.g., "Login failed. Please try again.")
- **AND** no token is stored in extension storage

### Requirement: Logout clears stored token and Kinde session

When the user clicks "Logout", the extension SHALL remove the `bookleaf_auth` key from `chrome.storage.local` and update the popup to the unauthenticated state immediately. It SHALL also make a best-effort non-interactive request to `{VITE_KINDE_ISSUER_URL}/logout` via `browser.identity.launchWebAuthFlow({ interactive: false })` to clear the Kinde server-side session, so that a subsequent login requires the user to re-authenticate rather than reusing the existing Kinde session silently.

#### Scenario: Logout removes token and shows login button

- **WHEN** the user clicks "Logout"
- **THEN** `bookleaf_auth` is removed from `chrome.storage.local`
- **AND** the popup renders the "Login with Bookleaf" button

#### Scenario: Logout clears Kinde session

- **WHEN** the user clicks "Logout"
- **THEN** a non-interactive request is made to the Kinde logout endpoint to clear the server-side session
- **AND** a subsequent login requires the user to authenticate again (no silent SSO reuse)

### Requirement: Auth token storage schema

Tokens SHALL be stored in `chrome.storage.local` under the key `bookleaf_auth` as a JSON object with the following shape:

```typescript
interface BookleafAuth {
  accessToken: string;
  refreshToken?: string; // optional — not issued by Kinde SPA clients by default
  expiresAt: number; // Unix timestamp in milliseconds
}
```

A typed helper module (`src/lib/storage.ts`) SHALL provide `getAuth()` and `setAuth()` / `clearAuth()` functions that read and write this schema.

#### Scenario: getAuth returns null when no token is stored

- **WHEN** `getAuth()` is called and `bookleaf_auth` is not in `chrome.storage.local`
- **THEN** `getAuth()` resolves to `null`

#### Scenario: setAuth persists the token object

- **WHEN** `setAuth({ accessToken, expiresAt })` is called
- **THEN** the object is stored in `chrome.storage.local` under `bookleaf_auth`
- **AND** a subsequent `getAuth()` call returns the same object
