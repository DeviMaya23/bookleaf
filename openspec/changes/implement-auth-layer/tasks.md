## 1. Provider Setup

- [x] 1.1 Wrap `<BrowserRouter>` and `<App />` in `KindeProvider` in `main.tsx`, reading all four `VITE_KINDE_*` env vars
- [x] 1.2 Add a console warning in `main.tsx` for any missing `VITE_KINDE_*` env var at startup

## 2. Routing Skeleton

- [x] 2.1 Replace the current `App.tsx` content with a `<Routes>` tree: `/login`, `/callback`, and a catch-all protected layout using `AuthGuard`

## 3. AuthGuard Component

- [x] 3.1 Create `src/components/AuthGuard.tsx` — renders nothing while `isLoading`, redirects to `/login` when unauthenticated, renders `<Outlet />` when authenticated

## 4. Login Page

- [x] 4.1 Create `src/pages/LoginPage.tsx` with a single sign-in button that calls `login()`
- [x] 4.2 Redirect authenticated users from `/login` to `/`
- [x] 4.3 Read React Router location state and display an error message above the sign-in button when present

## 5. Callback Route

- [x] 5.1 Create `src/pages/CallbackPage.tsx` that shows a loading state while the SDK processes the OAuth exchange
- [x] 5.2 Navigate to `/` once `isAuthenticated` is true after the exchange completes
- [x] 5.3 Detect Kinde error params in the callback URL (`error`, `error_description`) and redirect to `/login` with the error message in location state

## 6. Logout Button

- [x] 6.1 Create `src/components/LogoutButton.tsx` that calls `logout()` on click
- [x] 6.2 Render `LogoutButton` in the top-right corner of the authenticated app layout (inside `AuthGuard`'s outlet)

## 7. API Client

- [x] 7.1 Create `src/lib/api.ts` exporting `apiFetch(path, options?)` — awaits `getToken()`, attaches `Authorization: Bearer <token>`, prefixes `VITE_API_BASE_URL`

## 8. Unit Tests

- [x] 8.1 `AuthGuard` — test unauthenticated redirect to `/login` and authenticated render of `<Outlet />`
- [x] 8.2 `CallbackPage` — test successful navigation to `/` and error redirect to `/login` with location state
- [x] 8.3 `LoginPage` — test sign-in button renders, error message displays when location state contains an error
- [x] 8.4 `apiFetch` — test bearer token is attached to the request header
