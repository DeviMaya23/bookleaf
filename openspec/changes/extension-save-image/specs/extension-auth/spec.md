## MODIFIED Requirements

### Requirement: Kinde OAuth login via PKCE

When the user clicks "Login with Bookleaf", the extension SHALL initiate a Kinde OAuth 2.0 Authorization Code flow with PKCE:

1. Generate a cryptographically random `code_verifier` (43–128 characters, URL-safe)
2. Derive `code_challenge` as the Base64url-encoded SHA-256 hash of the `code_verifier`
3. Generate a cryptographically random `state` value (at least 16 URL-safe characters) for CSRF protection
4. Construct the Kinde authorization URL with `response_type=code`, `client_id`, `redirect_uri`, `scope=openid profile email`, `code_challenge`, `code_challenge_method=S256`, `state`, and `audience` (from `VITE_KINDE_AUDIENCE`)
5. Call `chrome.identity.launchWebAuthFlow({ url: authUrl, interactive: true })`
6. Extract the `code` parameter from the redirect URI returned by the flow
7. Exchange the code for tokens via a `POST` to `{VITE_KINDE_ISSUER_URL}/oauth2/token` with the `code`, `code_verifier`, `client_id`, `redirect_uri`, and `grant_type=authorization_code`
8. Store the resulting `access_token`, optional `refresh_token`, and computed `expires_at` in `chrome.storage.local` under key `bookleaf_auth`
9. Update the popup UI to the authenticated state

The redirect URI SHALL be `https://<extension-id>.chromiumapp.org/` (obtained via `chrome.identity.getRedirectURL()`).

The `offline_access` scope SHALL NOT be requested. Kinde does not permit it on SPA clients by default. The `refresh_token` field in storage is therefore optional.

#### Scenario: Successful login stores token and updates UI

- **WHEN** the user clicks "Login with Bookleaf" and completes the Kinde login flow
- **THEN** the access token is stored in `chrome.storage.local` under `bookleaf_auth`
- **AND** the popup updates to show the authenticated state

#### Scenario: User cancels the login flow

- **WHEN** the user closes the Kinde login window without completing login
- **THEN** `chrome.identity.launchWebAuthFlow` returns an error or empty URL
- **AND** the popup remains in the unauthenticated state
- **AND** no error is shown to the user (silent failure is acceptable)

#### Scenario: Token exchange fails

- **WHEN** the Kinde token endpoint returns an error during code exchange
- **THEN** the popup displays a brief error message (e.g., "Login failed. Please try again.")
- **AND** no token is stored in `chrome.storage.local`

## MODIFIED Requirements

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
