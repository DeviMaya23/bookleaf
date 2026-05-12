## Context

The frontend scaffold (`fe-project-setup`) is live with Vite + React + TypeScript + Tailwind + shadcn. `react-router-dom` v7, `@kinde-oss/kinde-auth-react` v5, and `@tanstack/react-query` are already installed. No routing or auth wiring exists yet — `App.tsx` is a single static component. The Kinde env vars are documented in `.env.example` but not populated in `.env`.

## Goals / Non-Goals

**Goals:**
- Wire `KindeProvider` so auth state is available app-wide
- Protect all non-auth routes behind `AuthGuard`
- Provide a `/login` page and a `/callback` route that completes the OAuth handshake
- Expose the Kinde access token on every backend API call via `Authorization: Bearer <token>`
- Provide a logout button in the top-right corner that ends the Kinde session

**Non-Goals:**
- Role-based access control or per-route permission checks
- Backend JWT validation (covered by the existing `kinde-auth` spec)
- Token refresh handling beyond what the Kinde SDK provides automatically

## Decisions

### 1. `KindeProvider` in `main.tsx`, not `App.tsx`

`KindeProvider` must wrap the router so that the `/callback` route (which completes the OAuth exchange) has access to Kinde context. Placing it in `main.tsx` above `<BrowserRouter>` guarantees this without prop drilling or context gymnastics.

Alternatives considered: wrapping inside `App.tsx` — rejected because the callback route would be inside the router but outside the provider, breaking the handshake.

### 2. `AuthGuard` as a layout component, not a HOC

`AuthGuard` is a React component rendered as a parent route element. Unauthenticated users are redirected to `/login`; authenticated users render `<Outlet />`. This keeps route definitions declarative in `App.tsx` and avoids wrapping each page individually.

Alternatives considered: per-page auth checks — rejected because it requires repetition and is easy to forget on new pages.

### 3. Callback page is a thin shell

`/callback` renders nothing meaningful — it just mounts and lets the Kinde SDK (via `KindeProvider`) automatically process the OAuth code exchange on mount. After the exchange completes the SDK updates its internal state; the page can then navigate to `/`. No custom token parsing needed.

### 4. API client in `src/lib/api.ts` using native `fetch`

A thin wrapper around `fetch` reads the Kinde access token at call time via `useKindeAuth().getToken()` (hook) or the SDK's imperative `getToken()`. Because the token is async, the wrapper is an async function that awaits the token before attaching it.

Alternatives considered: Axios interceptor — rejected to avoid adding a dependency; React Query's `queryFn` already accepts async functions cleanly. A module-level singleton interceptor was also considered but would require the Kinde client instance to be accessible outside React, which the SDK does not straightforwardly support in v5.

The API client exports a typed `apiFetch(path, options)` function. React Query `queryFn`s call this directly.

### 5. Logout button as a standalone component in the app shell

A `LogoutButton` component calls `useKindeAuth().logout()` on click, which ends the Kinde session and redirects to `VITE_KINDE_LOGOUT_REDIRECT_URL`. It is rendered top-right in the app layout, visible only when the user is authenticated (inside `AuthGuard`'s outlet). No separate logout page or confirmation dialog is needed.

### 6. No `.env` committed

Developers populate `frontend/.env` locally from `.env.example`. The Kinde vars are already documented there. CI/staging will inject them as environment secrets.

## Risks / Trade-offs

- **Token latency on first load** → `getToken()` may briefly be `null` while the SDK initialises. Mitigated: `AuthGuard` shows a loading state (or nothing) while `isLoading` is true, so `apiFetch` is never called before the token is ready.
- **Kinde SDK v5 API surface** → The SDK is at v5.11; minor API shape differences from v4 docs exist. Mitigation: use only `useKindeAuth()` hook surface (`isAuthenticated`, `isLoading`, `login`, `getToken`) which is stable across v5.
- **Env vars not set locally** → App will fail to init `KindeProvider`. Mitigation: add a clear error boundary or console warning when vars are missing; document in the change notes.
