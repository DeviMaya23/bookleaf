## 1. Project Scaffold

- [ ] 1.1 Create `/extensions` directory with `package.json` (name, scripts: `build`, `build:firefox`, `dev`, `type-check`)
- [ ] 1.2 Install dependencies: `vite`, `vite-plugin-web-extension`, `webextension-polyfill`, `@types/webextension-polyfill`, `typescript`, `react`, `react-dom`, `@types/react`, `@types/react-dom`
- [ ] 1.3 Create `vite.config.ts` with `vite-plugin-web-extension` configured for Chrome (default) and Firefox (`browser` flag)
- [ ] 1.4 Create `tsconfig.json` with `target: ES2020`, `lib: ["DOM", "WebWorker"]`, and `strict: true`
- [ ] 1.5 Create `manifest.json` with MV3 fields: `manifest_version: 3`, `name`, `version`, `permissions` (`storage`, `identity`), `host_permissions`, `action.default_popup`, `background.service_worker`
- [ ] 1.6 Create `.env.example` with `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_API_BASE_URL`
- [ ] 1.7 Create `.gitignore` for `/extensions` (node_modules, dist, .env)

## 2. Directory and Entry Points

- [ ] 2.1 Create `src/background/index.ts` (empty service worker stub)
- [ ] 2.2 Create `src/popup/index.html` as the popup entry HTML file
- [ ] 2.3 Create `src/popup/main.tsx` — React entry point mounting `<App />` into the popup HTML
- [ ] 2.4 Create `src/popup/App.tsx` — placeholder component rendering "Bookleaf Extension"

## 3. Storage Helper

- [ ] 3.1 Create `src/lib/storage.ts` with `BookleafAuth` interface (`accessToken`, `refreshToken`, `expiresAt`)
- [ ] 3.2 Implement `getAuth(): Promise<BookleafAuth | null>` reading from `chrome.storage.local`
- [ ] 3.3 Implement `setAuth(auth: BookleafAuth): Promise<void>` writing to `chrome.storage.local`
- [ ] 3.4 Implement `clearAuth(): Promise<void>` removing `bookleaf_auth` from `chrome.storage.local`

## 4. Auth Library

- [ ] 4.1 Create `src/lib/auth.ts` with PKCE helpers: `generateCodeVerifier()` and `generateCodeChallenge(verifier)`
- [ ] 4.2 Implement `getRedirectUri(): string` using `chrome.identity.getRedirectURL()`
- [ ] 4.3 Implement `buildAuthUrl(codeChallenge: string): string` constructing the Kinde authorization URL with all required params
- [ ] 4.4 Implement `exchangeCodeForTokens(code: string, codeVerifier: string): Promise<BookleafAuth>` calling Kinde's token endpoint
- [ ] 4.5 Implement `login(): Promise<void>` — orchestrates PKCE generation, `launchWebAuthFlow`, code extraction, token exchange, and `setAuth`

## 5. Popup UI

- [ ] 5.1 Update `App.tsx` to read auth state from `getAuth()` on mount using `useEffect`
- [ ] 5.2 Render unauthenticated state: "Login with Bookleaf" button that calls `login()`
- [ ] 5.3 Render authenticated state: "Logged in" message and "Logout" button that calls `clearAuth()` then updates state
- [ ] 5.4 Handle loading state (while `getAuth()` is resolving) to prevent content flash
- [ ] 5.5 Display error message when `login()` throws (e.g., token exchange failure)

## 6. Verification

- [ ] 6.1 Run `npm run type-check` in `/extensions` and confirm zero TypeScript errors
- [ ] 6.2 Run `npm run build` and confirm `dist/` contains `manifest.json` and popup/background entries
- [ ] 6.3 Load the unpacked extension in Chrome (`dist/`) and verify the popup renders without console errors
- [ ] 6.4 Complete a full login flow in the loaded extension and confirm the token is stored in `chrome.storage.local`
- [ ] 6.5 Click "Logout" and confirm `bookleaf_auth` is cleared and the login button reappears
