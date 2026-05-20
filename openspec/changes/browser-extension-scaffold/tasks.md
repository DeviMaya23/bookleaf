## 1. Project Scaffold

- [x] 1.1 Create `/extensions` directory with `package.json` (name, scripts: `build`, `build:firefox`, `dev`, `type-check`)
- [x] 1.2 Install dependencies: `vite`, `vite-plugin-web-extension`, `webextension-polyfill`, `@types/webextension-polyfill`, `typescript`, `react`, `react-dom`, `@types/react`, `@types/react-dom`
- [x] 1.3 Create `vite.config.ts` with `vite-plugin-web-extension` configured for Chrome (default) and Firefox (`browser` flag)
- [x] 1.4 Create `tsconfig.json` with `target: ES2020`, `lib: ["DOM", "WebWorker"]`, and `strict: true`
- [x] 1.5 Create `manifest.json` with MV3 fields: `manifest_version: 3`, `name`, `version`, `permissions` (`storage`, `identity`), `host_permissions`, `action.default_popup`, `background.service_worker`
- [x] 1.6 Create `.env.example` with `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_API_BASE_URL`
- [x] 1.7 Create `.gitignore` for `/extensions` (node_modules, dist, .env)

## 2. Directory and Entry Points

- [x] 2.1 Create `src/background/index.ts` (empty service worker stub)
- [x] 2.2 Create `src/popup/index.html` as the popup entry HTML file
- [x] 2.3 Create `src/popup/main.tsx` — React entry point mounting `<App />` into the popup HTML
- [x] 2.4 Create `src/popup/App.tsx` — placeholder component rendering "Bookleaf Extension"

## 3. Storage Helper

- [x] 3.1 Create `src/lib/storage.ts` with `BookleafAuth` interface (`accessToken`, `refreshToken`, `expiresAt`)
- [x] 3.2 Implement `getAuth(): Promise<BookleafAuth | null>` reading from `chrome.storage.local`
- [x] 3.3 Implement `setAuth(auth: BookleafAuth): Promise<void>` writing to `chrome.storage.local`
- [x] 3.4 Implement `clearAuth(): Promise<void>` removing `bookleaf_auth` from `chrome.storage.local`

## 4. Auth Library

- [x] 4.1 Create `src/lib/auth.ts` with PKCE helpers: `generateCodeVerifier()` and `generateCodeChallenge(verifier)`
- [x] 4.2 Implement `getRedirectUri(): string` using `chrome.identity.getRedirectURL()`
- [x] 4.3 Implement `buildAuthUrl(codeChallenge: string): string` constructing the Kinde authorization URL with all required params
- [x] 4.4 Implement `exchangeCodeForTokens(code: string, codeVerifier: string): Promise<BookleafAuth>` calling Kinde's token endpoint
- [x] 4.5 Implement `login(): Promise<void>` — orchestrates PKCE generation, `launchWebAuthFlow`, code extraction, token exchange, and `setAuth`

## 5. Popup UI

- [x] 5.1 Update `App.tsx` to read auth state from `getAuth()` on mount using `useEffect`
- [x] 5.2 Render unauthenticated state: "Login with Bookleaf" button that calls `login()`
- [x] 5.3 Render authenticated state: "Logged in" message and "Logout" button that calls `clearAuth()` then updates state
- [x] 5.4 Handle loading state (while `getAuth()` is resolving) to prevent content flash
- [x] 5.5 Display error message when `login()` throws (e.g., token exchange failure)

## 6. Verification

- [x] 6.1 Run `npm run type-check` in `/extensions` and confirm zero TypeScript errors
- [x] 6.2 Run `npm run build` and confirm `dist/` contains `manifest.json` and popup/background entries
- [ ] 6.3 Load the unpacked extension in Chrome (`dist/`) and verify the popup renders without console errors
- [ ] 6.4 Complete a full login flow in the loaded extension and confirm the token is stored in `chrome.storage.local`
- [ ] 6.5 Click "Logout" and confirm `bookleaf_auth` is cleared and the login button reappears

